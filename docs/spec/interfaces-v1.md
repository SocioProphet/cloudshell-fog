# Interface Specification v1 — Fog-Optimised Cloud Shell

This document defines the standards-aligned interface contracts for cloudshell-fog.

## 0. Standards alignment

| Concern | Standard |
|---|---|
| Identity | OpenID Connect (OIDC) / OAuth 2.0 (Auth Code + PKCE) |
| Tokens | JWT (short-lived access tokens) — optional DPoP (sender-constrained) in a future revision |
| Terminal transport | WebSocket (WSS) with explicit JSON message schema |
| Observability | OpenTelemetry (OTEL) |
| Runtime packaging | OCI images + OCI registry |
| Supply chain | SBOM (SPDX or CycloneDX) + signatures (Sigstore/cosign) + provenance attestations (in-toto / Tekton Chains) |

---

## 1. Planes and responsibilities

### 1.1 UI / Edge plane

- Browser terminal UI (xterm.js) embedded in a console page.
- OIDC login is handled at the console edge; the gateway receives already-validated access tokens.
- Terminal attach uses WSS; file transfer uses HTTPS endpoints (optional, not yet implemented).

### 1.2 Control plane (HTTP API)

All endpoints require a valid OIDC access token in the `Authorization: Bearer` header.

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/sessions` | Create a session. Body: `{ profile, ttl_seconds, placement_hint?, image_ref? }`. Response: `{ session_id, attach: { ws_url, token, expires_at }, placement: { region, node_id, tier, reasons[] } }`. |
| `GET` | `/v1/sessions/{id}` | Get session status. Response: `{ status, placement, created_at, expires_at, image_ref }`. |
| `DELETE` | `/v1/sessions/{id}` | Terminate a session. Response: `{ terminated: true }`. |

See [`docs/reference/api.md`](../reference/api.md) for the complete reference with examples.

### 1.3 Data plane (WebSocket PTY)

```
wss://<gateway>/v1/sessions/{id}/pty?token=<short-lived-JWT>
```

JSON message frames (binary payloads are base64-encoded to avoid encoding pitfalls):

| Direction | Frame |
|---|---|
| Client → Server | `{"type":"resize","cols":int,"rows":int}` |
| Client → Server | `{"type":"stdin","data_b64":"<base64>"}` |
| Server → Client | `{"type":"stdout","data_b64":"<base64>"}` |
| Server → Client | `{"type":"exit","code":int}` |

### 1.4 Runtime connector (internal contract)

The connector interface is internal to the gateway. Implementations may use the Kubernetes API, gRPC, or HTTP, but must satisfy:

| Method | Signature |
|---|---|
| `Allocate` | `(session_id, profile, placement, image_ref) → runtime_ref` |
| `AttachPTY` | `(runtime_ref) → (reader, writer)` |
| `Terminate` | `(runtime_ref) → error` |

Current implementations: `StubConnector` (dev) and `K8sConnector` (production).

### 1.5 Audit and observability

Minimum audit events emitted per operation:

| Event | Trigger |
|---|---|
| `session.created` | Successful `POST /v1/sessions` |
| `session.attached` | PTY WebSocket connected |
| `session.terminated` | `DELETE /v1/sessions/{id}` or TTL expiry |
| `placement.decided` | Placement engine selected a node (includes reason codes and tier) |
| `runtime.allocated` | Connector provisioned a runtime (includes image digest) |
| `policy.denied` | Policy engine rejected a request (includes rule and reason) |

OpenTelemetry spans wrap session create, PTY attach, and termination flows. See [Observability guide](../guides/observability.md).

---

## 2. Security invariants

| Invariant | Description |
|---|---|
| No long-lived secrets in the browser or gateway | Session tokens are short-lived (15 min) and session-scoped. |
| Session token is session-bound | A PTY session token cannot be used to create new sessions. |
| Pinned image digests in production | Mutable image tags (`latest`, etc.) are rejected by the admission policy. |
| Supply-chain attestation required | Production admission requires a cosign signature, a Tekton Chains provenance attestation, and an SPDX/CycloneDX SBOM. |
| Network isolation | Each session pod runs in its own namespace with default-deny NetworkPolicies; only DNS and HTTPS egress are permitted. |
| Least-privilege RBAC | The gateway service account holds the minimum permissions required to manage session namespaces and pods. |

---

## 3. Fog placement semantics

Placement decision inputs:

- Latency estimate / region preference (`placement_hint`)
- Node health and free capacity
- Data locality constraints
- Trust tier: `fog` (attested edge node) vs `cloud` (managed cloud region)

Degraded operation:

- If no fog node is reachable or healthy, the placement engine falls back to `CLOUD_FALLBACK_REGION`.
- Session resume after fog node failure is best-effort; the current status is surfaced to the user.
