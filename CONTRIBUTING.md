# Contributing to cloudshell-fog

Thank you for your interest in contributing! This document explains how to get started.

## Code of Conduct

Please be respectful and constructive in all interactions. We follow the [Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/) code of conduct.

## Getting Started

1. **Fork** the repository and clone your fork.
2. Install dependencies:
   - Go 1.22+ (`go.dev/dl`)
   - Node.js 20+ (for the web UI)
   - Docker (optional)
3. Build and run locally:
   ```bash
   make build        # compile the gateway binary
   make frontend     # build the web UI bundle
   make run-dev      # run with stub connector + dev auth shim
   ```
4. Run the test suite before and after your change:
   ```bash
   make test
   make vet
   ```

## How to Contribute

### Reporting Bugs

Open an issue using the **Bug Report** template. Include:
- Steps to reproduce
- Expected vs. actual behaviour
- Go and OS versions

### Requesting Features

Open an issue using the **Feature Request** template. Describe the problem you want to solve and your proposed solution.

### Submitting Pull Requests

1. Open an issue first to discuss the change (skip for trivial fixes).
2. Create a branch off `main` with a descriptive name, e.g. `feat/fog-node-heartbeat` or `fix/session-ttl`.
3. Make your changes following the conventions below.
4. Ensure `make test` and `make vet` pass.
5. Open a pull request against `main`. Fill in the PR template.

## Coding Conventions

- **Go**: follow standard `gofmt` / `goimports` formatting. Run `go vet ./...` before submitting.
- **Comments**: every exported symbol must have a Go doc comment. Package-level doc comments are required.
- **Errors**: wrap errors with `fmt.Errorf("context: %w", err)` so callers can use `errors.Is`/`errors.As`.
- **Logging**: use the structured `slog.Logger` passed through the call chain; never call `log.Print*` directly.
- **TypeScript**: follow the existing style in `web/src/`. Run `npm run build` to verify.
- **Commits**: use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, `chore:`, etc.).

## Project Layout

```
cmd/gateway/        # main binary entrypoint
internal/
  api/              # HTTP session management endpoints
  audit/            # structured audit-event emission
  auth/             # OIDC validation and session-token minting
  connector/        # runtime connector interface + k8s and stub impls
  otel/             # OpenTelemetry setup
  placement/        # fog-aware placement engine
  policy/           # YAML-driven admission policy
  pty/              # WebSocket PTY handler
  session/          # session store and TTL sweeper
web/                # TypeScript + xterm.js browser UI
config/             # default policy configuration
deploy/             # Kubernetes, Argo CD, and Tekton manifests
docs/spec/          # interface contracts and architecture specs
```

## License

By contributing you agree that your contributions will be licensed under the [Apache 2.0 License](LICENSE).
