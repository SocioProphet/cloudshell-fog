# Interfaces v0 — Cloud Shell (Fog-Optimized)

This document defines the minimum interface contracts to wrap a Wetty-class terminal.

## 1) Auth / Admission
### 1.1 Identity
- OIDC provider issues ID token + access token.
- Shell Gateway accepts only short-lived access tokens.

### 1.2 Admission decision
Inputs:
- subject (sub), groups/roles, requested profile (cpu/mem), requested placement hint (optional)
Outputs:
- session_id
- websocket_url (or route token)
- session_token (short-lived, bound to session_id)

## 2) Session API (HTTP)
- POST /v1/sessions
  - body: { profile, ttl, placement_hint? }
  - returns: { session_id, ws_url, session_token, expires_at }
- GET /v1/sessions/{id}
  - returns: { status, placement, created_at, expires_at }
- DELETE /v1/sessions/{id}
  - returns: { terminated: true }

## 3) Terminal attach (WebSocket)
- ws(s)://.../v1/sessions/{id}/pty?token=...
Messages:
- resize: {type:"resize", cols, rows}
- stdin:  {type:"stdin", data: base64}
- stdout: {type:"stdout", data: base64}
- exit:   {type:"exit", code}

## 4) Runtime Connector contract (internal)
Responsibilities:
- allocate_runtime(session_id, profile, placement)
- attach_pty(session_id) -> stream handle
- enforce_limits(session_id) (cpu/mem/fs/egress)
- emit_audit(event)

## 5) Audit event schema (minimum)
Event: { ts, session_id, subject, type, placement, details }
Types:
- session.created, session.attached, pty.resize, pty.stdin, pty.stdout, session.terminated
