# Changelog

All notable changes to cloudshell-fog are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
cloudshell-fog uses [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

### Added
- Comprehensive documentation overhaul: getting-started guide, OIDC configuration guide, observability guide, troubleshooting guide, full API reference, and configuration reference.
- `docs/guides/` directory with end-to-end how-to guides.
- `docs/reference/` directory with API and configuration references.
- Go Report Card badge and OpenSSF Best Practices badge in README.
- "Why cloudshell-fog?" comparison table in README.
- Architecture component summary table in README.

### Changed
- License updated to MIT.
- README restructured with improved hero section, quick-start table, and links to new guide docs.

---

## [0.1.0] — 2026-01-01

### Added
- Initial release of cloudshell-fog.
- OIDC authentication middleware with dev shim for local development.
- Fog-aware placement engine: fog-tier-first, cloud-fallback.
- YAML-driven policy engine with CPU/RAM/storage quotas and group-based access control.
- In-memory session store with TTL-based sweeper.
- PTY over WebSocket with JSON frame schema (stdin, stdout, resize, exit).
- Kubernetes runtime connector: per-session namespace and pod.
- Stub runtime connector for local development.
- Structured audit-event emission via `slog` (session.created, session.attached, session.terminated, placement.decided, runtime.allocated, policy.denied).
- OpenTelemetry provider setup with stdout exporters.
- Kubernetes deployment manifests (`deploy/k8s/`).
- Argo CD Application and AppProject manifests (`deploy/argocd/`).
- Tekton build pipeline and Chains supply-chain configuration (`deploy/tekton/`).
- xterm.js browser UI served as static files from `web/public/`.
- Makefile targets: `build`, `test`, `vet`, `lint`, `frontend`, `docker-build`, `run-dev`.
- Multi-stage Dockerfile.
- GitHub Actions CI workflow.
- `CONTRIBUTING.md`, `SECURITY.md`, and issue/PR templates.

[Unreleased]: https://github.com/SocioProphet/cloudshell-fog/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/SocioProphet/cloudshell-fog/releases/tag/v0.1.0
