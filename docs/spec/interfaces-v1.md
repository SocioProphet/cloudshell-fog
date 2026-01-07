# Interfaces v1 — Fog-Optimized Cloud Shell (Standards-Aligned)

Clean-room notes. No vendor diagrams/artifacts.

## 0) Standards we align to (minimum set)
- Identity: OpenID Connect (OIDC) / OAuth 2.0 (Auth Code + PKCE)
- Tokens: JWT (short-lived). Optional sender-constrained tokens (DPoP) later.
- Terminal transport: WebSocket (WSS) with explicit message schema
- Observability: OpenTelemetry (OTEL)
- Runtime packaging: OCI images + OCI registry
- Supply chain: SBOM (SPDX or CycloneDX) + signatures (Sigstore/cosign) + provenance attestations (in-toto style; Tekton Chains)

## 1) Planes & responsibilities
### 1.1 UI / Edge plane
- Browser terminal UI (Wetty-class) embedded in a console.
- OIDC login handled at the console edge; gateway receives validated access tokens.
- Terminal attach uses WSS; file transfer uses HTTPS endpoints (optional).

### 1.2 Control plane (HTTP API)
All endpoints require OIDC access token validation.

- POST /v1/sessions
  - req: { profile, ttl_seconds, placement_hint?, image_ref? }
  - resp: { session_id, attach: { ws_url, token, expires_at }, placement }

- GET /v1/sessions/{id}
  - resp: { status, placement, created_at, expires_at, image_ref }

- DELETE /v1/sessions/{id}
  - resp: { terminated: true }

### 1.3 Data plane (WSS attach)
- wss://<gateway>/v1/sessions/{id}/pty?token=...

Message frames (JSON; payloads base64 to avoid encoding pitfalls):
- {type:"resize", cols:int, rows:int}
- {type:"stdin",  data_b64:string}
- {type:"stdout", data_b64:string}
- {type:"exit",   code:int}

### 1.4 Runtime connector (internal contract)
Implementation can be gRPC/HTTP/k8s API, but the contract is:
- allocate(session_id, profile, placement, image_ref) -> runtime_ref
- attach_pty(runtime_ref) -> streams
- enforce_limits(runtime_ref) -> ok/fail
- terminate(runtime_ref)

### 1.5 Audit + Observability (OTEL)
Emit:
- traces for session create/attach/terminate
- metrics: latency, attach failures, session durations, placement outcomes
- structured logs
Minimum audit events:
- session.created, session.attached, session.terminated
- placement.decided (include reason codes)
- runtime.allocated (include image digest)
- policy.denied (include rule id)

## 2) Security invariants (minimum)
- No long-lived secrets in browser or gateway.
- Session token is short-lived and session-bound; cannot mint new sessions.
- Production runs pinned image digests only (no mutable tags).
- Admission policy requires:
  - cosign signature verification
  - provenance attestation (Tekton Chains)
  - SBOM present (SPDX or CycloneDX)

## 3) Fog behavior (minimum semantics)
Placement inputs:
- latency estimate / region
- node health + capacity
- data locality constraints
- trust tier (attested fog node vs cloud fallback)

Degraded operation:
- if fog node unreachable, fail over to nearest trusted cloud region
- session resume is best-effort; explicit status surfaced to the user
