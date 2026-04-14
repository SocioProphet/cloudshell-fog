# Residual Hook-Ups v1 — cloudshell-fog

This file captures the remaining small existing-file edits after the first merge wave.

## 1. Gateway connector-mode hook-up

Status:
- helper + tests exist in PR #6
- the remaining edit is in `cmd/gateway/main.go`

Required change:

Replace:

```go
var conn connector.Connector
conn = connector.NewStubConnector()
logger.Info("using stub connector (set USE_K8S=1 and provide kubeconfig for k8s)")
```

With:

```go
conn, _, err := buildConnector(logger)
if err != nil {
    return fmt.Errorf("build connector: %w", err)
}
```

## 2. Placement response hook-up

Status:
- helper + tests were merged via PR #13
- the remaining edit is in `internal/api/handler.go`

Required change:

Replace the current flat response block with:

```go
writeJSON(w, http.StatusCreated, BuildCreateSessionResponseVNext(
    sessionID,
    h.gatewayURL,
    token,
    tokenExpiresAt,
    decision,
))
```

## 3. Follow-on after hook-up

After the placement hook-up lands:
- align `docs/spec/openapi/control-api.v1.yaml` to the richer placement object
- consider whether a v2 response schema is warranted for compatibility signaling

## 4. Why this file exists

These are the smallest remaining seams between the now-merged additive work and the live behavior of the repo.
