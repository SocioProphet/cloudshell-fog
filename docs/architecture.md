# Architecture

This document describes the internal components of the cloudshell-fog gateway and how they interact.

## Component Overview

```
Browser
  │  HTTP/WebSocket  (OIDC Bearer token or short-lived session token)
  ▼
┌──────────────────────────────────────────────────────────────────┐
│                      gateway  (cmd/gateway)                      │
│                                                                  │
│  ┌─────────┐   ┌──────────┐   ┌─────────┐   ┌───────────────┐   │
│  │  auth   │   │   api    │   │   pty   │   │  static files │   │
│  │ (OIDC)  │──▶│ handler  │   │ handler │   │  web/public/  │   │
│  └─────────┘   └────┬─────┘   └────┬────┘   └───────────────┘   │
│                     │              │                              │
│        ┌────────────▼──────────────▼──────────────┐              │
│        │              session.Store               │              │
│        └──────────────────────────────────────────┘              │
│                     │              │                              │
│        ┌────────────▼──┐  ┌────────▼───────┐                     │
│        │   placement   │  │   connector    │                     │
│        │   Engine      │  │  (k8s/stub)    │                     │
│        └───────────────┘  └────────────────┘                     │
│        ┌───────────────┐  ┌────────────────┐                     │
│        │   policy      │  │     audit      │                     │
│        │   Engine      │  │   (slog emit)  │                     │
│        └───────────────┘  └────────────────┘                     │
│        ┌──────────────────────────────────┐                      │
│        │       otel (traces + metrics)    │                      │
│        └──────────────────────────────────┘                      │
└──────────────────────────────────────────────────────────────────┘
  │  k8s API / exec stream
  ▼
Runtime pod  (per-session namespace)
```

## Packages

### `cmd/gateway`

Binary entrypoint. Wires all components together, configures HTTP routes, and manages graceful shutdown.

Key responsibilities:
- Parse environment variables and load policy config.
- Initialise OpenTelemetry providers.
- Build the auth middleware (real OIDC or dev shim).
- Register HTTP routes on a `gorilla/mux` router.
- Run the session TTL sweeper as a background goroutine.
- Perform graceful HTTP server shutdown on `SIGINT`/`SIGTERM`.

### `internal/auth`

**OIDC token validation** and **short-lived session token minting**.

- `OIDCValidator` — validates Bearer tokens against an OIDC provider's JWKS endpoint using `coreos/go-oidc`.
- `Middleware` — HTTP middleware that calls `OIDCValidator` and injects `subject` + `groups` into the request context.
- `SessionTokenMinter` — mints short-lived HMAC-signed JWTs (15 min default) bound to a `session_id`. Used for the PTY WebSocket endpoint so the short-lived token cannot create new sessions.

### `internal/api`

HTTP session management endpoints (`POST /v1/sessions`, `GET /v1/sessions/{id}`, `DELETE /v1/sessions/{id}`).

On `POST /v1/sessions`:
1. Decode and validate request body.
2. List active sessions for the subject; check per-subject quota via the policy engine.
3. Run admission policy (`policy.Engine.CheckAdmission`) to resolve resource limits.
4. Run placement (`placement.Engine.Decide`) to select a node.
5. Create a session record in the store.
6. Allocate a runtime via the connector.
7. Emit audit events: `session.created`, `placement.decided`, `runtime.allocated`.
8. Mint a session token and return attach info.

### `internal/pty`

WebSocket PTY handler (`GET /v1/sessions/{id}/pty?token=...`).

- Validates the short-lived session token.
- Looks up the session; verifies ownership.
- Calls `connector.AttachPTY` to get bidirectional streams.
- Pumps frames between the WebSocket and the PTY streams.
- Emits `session.attached` audit event.
- Handles `resize` frames by forwarding terminal dimensions to the PTY.

### `internal/session`

In-memory session store and TTL sweeper.

- `InMemoryStore` — thread-safe map of `Session` structs.
- `Sweeper` — polls every 30 s; terminates sessions whose `ExpiresAt` has passed by calling the connector and emitting a `session.terminated` audit event.

### `internal/placement`

Fog-aware placement engine.

- `Registry` — thread-safe node registry (upsert/list).
- `Node` — holds region, trust tier (`fog` or `cloud`), health, free capacity, and latency.
- `Engine.Decide` — selects the best available node: prefers fog-tier nodes with sufficient capacity and good health; falls back to the cloud-fallback region when no fog node qualifies.
- Returns a `Decision` with the chosen node ID, region, trust tier, and a slice of human-readable reason strings for auditing.

### `internal/policy`

YAML-driven admission policy engine.

- Loads `config/policy.yaml` at startup.
- `Engine.CheckAdmission(profile, groups, ttlSeconds, activeSessions)` — finds the matching profile by name, checks group membership (if `allowed_groups` is non-empty), validates TTL against `max_ttl_seconds`, and checks the active session count against `max_sessions`.
- Returns the resolved `Profile` (CPU/memory/storage limits) or an error describing the denial reason.

### `internal/connector`

Runtime connector interface and implementations.

- `Connector` interface: `Allocate`, `AttachPTY`, `Terminate`.
- `StubConnector` — no-op implementation for local development; returns fake refs and `/bin/sh` streams.
- `K8sConnector` — Kubernetes implementation; creates a per-session namespace and pod, then uses `kubectl exec`-style streaming for PTY attach.

### `internal/audit`

Structured audit-event emission.

- `Emitter` interface with `Emit(ctx, Event)`.
- `LogEmitter` — writes events as structured JSON via `slog`.
- `Event` — `{ ts, session_id, subject, type, placement, details }`.
- Event types: `session.created`, `session.attached`, `session.terminated`, `placement.decided`, `runtime.allocated`, `policy.denied`.

### `internal/otel`

OpenTelemetry initialisation.

- Sets up a `TracerProvider` and `MeterProvider` with stdout exporters.
- Returns a `Providers` struct with a `Shutdown` method for clean teardown.

## Data Flow: Session Create

```
POST /v1/sessions
  → auth.Middleware  (validate OIDC token → subject, groups)
  → api.Handler.CreateSession
      → policy.Engine.CheckAdmission   (quota + group check)
      → placement.Engine.Decide        (pick best node)
      → session.Store.Create           (persist pending session)
      → connector.Allocate             (start runtime pod)
      → session.Store.Update           (mark running)
      → audit.Emitter.Emit × 3        (created, placement, runtime)
      → auth.SessionTokenMinter.Mint   (short-lived attach token)
  ← 201 { session_id, attach: { ws_url, token, expires_at }, placement }
```

## Data Flow: PTY Attach

```
GET /v1/sessions/{id}/pty?token=...  (WebSocket upgrade)
  → pty.Handler.ServeHTTP
      → auth.SessionTokenMinter.Verify  (validate token)
      → session.Store.Get               (ownership check)
      → connector.AttachPTY             (open PTY streams)
      → audit.Emitter.Emit              (session.attached)
      → goroutine: ws → stdin pump
      → goroutine: stdout → ws pump
      → resize frames forwarded to PTY
  ↔ WebSocket frames (stdin/stdout/resize/exit)
```

## Security Model

See [`docs/spec/interfaces-v1.md`](spec/interfaces-v1.md) §2 for the full security invariants. Key points:

- OIDC tokens are validated on every session-management request; they are never stored.
- Session tokens are short-lived (15 min), session-scoped JWTs that can only be used for PTY attach.
- Each session pod runs in an isolated Kubernetes namespace with default-deny NetworkPolicies.
- The gateway service account uses least-privilege RBAC.
- Production deployments must use pinned OCI image digests (enforced by admission policy).
