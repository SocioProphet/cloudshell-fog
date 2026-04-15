# EF-0001 — `main.go` hookup patch

This note exists because the active GitHub connector session can safely create new files and branches, but does not expose a straightforward SHA-aware update path for existing files through the high-level contents API.

The implementation branch already contains:
- `cmd/gateway/connector_mode.go`
- `cmd/gateway/connector_mode_test.go`

The remaining code change to complete EF-0001 is a surgical edit in `cmd/gateway/main.go`.

## Required change

Replace this block:

```go
// ── Runtime connector ─────────────────────────────────────────────────────
var conn connector.Connector
conn = connector.NewStubConnector()
logger.Info("using stub connector (set USE_K8S=1 and provide kubeconfig for k8s)")
```

With this block:

```go
// ── Runtime connector ─────────────────────────────────────────────────────
conn, _, err := buildConnector(logger)
if err != nil {
	return fmt.Errorf("build connector: %w", err)
}
```

## Why this is safe

- no API shape changes
- no policy changes
- no placement changes
- no session-store changes
- only swaps connector instantiation from hardcoded stub to explicit mode resolution

## Expected behavior after hookup

### `CONNECTOR_MODE=stub`
- gateway uses stub connector
- startup logs selected mode as `stub`

### `CONNECTOR_MODE=k8s`
- gateway loads k8s config from:
  1. `KUBECONFIG`, if set
  2. in-cluster config otherwise
- invalid k8s config fails fast at startup
- startup logs selected mode and config source

## Follow-on tasks (not part of this exact edit)
- verify k8s credential model against deployment manifests / ServiceAccount settings
- add integration/startup tests around `buildConnector`
- clean up deployment examples that still imply `:latest` is production-safe
