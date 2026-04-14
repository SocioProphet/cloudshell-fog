# Manual Smoke Tests v0 — cloudshell-fog

Use this checklist after deploying or wiring a new connector mode.

## A. Gateway starts

- [ ] gateway process starts without panic
- [ ] startup logs show selected connector mode
- [ ] startup logs show auth mode (OIDC or dev shim)

## B. Session create API

- [ ] `POST /v1/sessions` succeeds with valid auth
- [ ] invalid `profile` fails cleanly
- [ ] policy-denied request fails with non-200 response

## C. Session status API

- [ ] `GET /v1/sessions/{id}` returns status for owned session
- [ ] cross-subject lookup is rejected

## D. Session delete API

- [ ] `DELETE /v1/sessions/{id}` terminates session
- [ ] repeated delete converges cleanly or returns terminal truth

## E. PTY attach path

- [ ] valid session token attaches successfully
- [ ] wrong-session token is rejected
- [ ] invalid/expired token is rejected
- [ ] stdin/stdout flow works
- [ ] resize events are accepted without breaking attach

## F. Stub mode specific

- [ ] no cluster access required
- [ ] shell prompt / echo behavior visible

## G. Kubernetes mode specific

- [ ] session namespace is created
- [ ] session pod is created and reaches running state
- [ ] connector can attach PTY to pod
- [ ] namespace teardown happens on delete or expiry

## H. Audit / observability

- [ ] session.created emitted
- [ ] placement.decided emitted
- [ ] runtime.allocated emitted
- [ ] session.attached emitted
- [ ] session.terminated emitted

## I. Policy / supply chain sanity

- [ ] runtime baseline policy present
- [ ] digest policy present
- [ ] signed-image policy or keyless variant present
