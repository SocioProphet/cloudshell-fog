# cloudshell-fog

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![CI](https://github.com/SocioProphet/cloudshell-fog/actions/workflows/ci.yml/badge.svg)](https://github.com/SocioProphet/cloudshell-fog/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/SocioProphet/cloudshell-fog)](https://goreportcard.com/report/github.com/SocioProphet/cloudshell-fog)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/placeholder/badge)](https://www.bestpractices.dev/projects/placeholder)

**cloudshell-fog** is an open-source, fog-optimised cloud shell gateway. It delivers a secure, browser-accessible terminal that runs _as close to the user's data as possible_ — on a fog/edge node when one is available, and gracefully falls back to a trusted cloud region when not.

> Think of it as a self-hosted, privacy-first alternative to browser-based cloud shells (Google Cloud Shell, AWS CloudShell) — designed for edge computing, data-sovereignty use cases, and organisations that cannot route developer traffic through a hyperscaler.

```
Browser (xterm.js)
      │  OIDC access-token (Authorization header)
      ▼
┌─────────────────────────────────────────────────────┐
│               cloudshell-fog  gateway               │
│  ┌──────────┐  ┌───────────┐  ┌──────────────────┐  │
│  │  Auth    │  │ Placement │  │  Policy Engine   │  │
│  │  (OIDC)  │  │  Engine   │  │  (YAML rules)    │  │
│  └──────────┘  └───────────┘  └──────────────────┘  │
│  ┌──────────────────────────────────────────────┐   │
│  │            Session Store (in-memory)         │   │
│  └──────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────┐   │
│  │         Runtime Connector (k8s / stub)       │   │
│  └──────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────┐   │
│  │            Audit + OpenTelemetry             │   │
│  └──────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
      │  WebSocket PTY  (wss://.../v1/sessions/{id}/pty)
      ▼
 Runtime pod / shell process
```

## Why cloudshell-fog?

| Concern | cloudshell-fog | Hosted cloud shells |
|---|---|---|
| **Data stays on-premises / at the edge** | ✅ Fog-first placement | ❌ Data traverses hyperscaler backbone |
| **Self-hosted, no vendor lock-in** | ✅ MIT licensed, any k8s cluster | ❌ Tied to one cloud provider |
| **OIDC with your own identity provider** | ✅ Works with Keycloak, Dex, Okta, etc. | ⚠️ Provider-specific IAM |
| **Supply-chain integrity** | ✅ Tekton Chains, cosign, SBOM | ❌ Opaque |
| **OpenTelemetry observability** | ✅ Traces, metrics, structured logs | ⚠️ Proprietary metrics only |
| **GitOps deployment** | ✅ Argo CD + Tekton manifests included | ❌ Console-driven |

## Features

- **OIDC authentication** — validates short-lived access tokens against any OIDC provider; dev shim for local work without a provider.
- **Fog placement** — selects the nearest healthy node (fog tier first, cloud fallback) using latency, capacity, and trust-tier signals.
- **Policy engine** — YAML-configured resource profiles with CPU/RAM/storage quotas and group-based access control.
- **Session lifecycle** — TTL-based expiry with automatic runtime cleanup; background sweeper reclaims stale sessions.
- **PTY over WebSocket** — resize, stdin/stdout, exit frames (JSON + base64) over a secure WebSocket connection.
- **Audit trail** — structured log events for every session, placement, and policy decision.
- **OpenTelemetry** — traces, metrics, and structured logs exported via stdout (drop-in for any OTel collector).
- **GitOps-ready** — Kubernetes manifests, Argo CD Application, and Tekton Chains supply-chain pipeline included out of the box.

## Table of Contents

- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Architecture](#architecture)
- [Deployment](#deployment)
- [Guides](#guides)
- [Contributing](#contributing)
- [Security](#security)
- [Changelog](#changelog)
- [License](#license)

## Quick Start

### Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| [Go](https://go.dev/dl) | 1.22+ | Build the gateway binary |
| [Node.js](https://nodejs.org) | 20+ | Build the web UI |
| [Docker](https://docs.docker.com/get-docker/) | any recent | Container builds (optional) |
| [kubectl](https://kubernetes.io/docs/tasks/tools/) | 1.28+ | Kubernetes deployment (optional) |

### 1 — Run locally (no OIDC, no Kubernetes)

This is the fastest way to see cloudshell-fog working. It uses a stub connector that opens a real `/bin/sh` on your machine and a dev auth shim that accepts any token.

```bash
# Clone the repository
git clone https://github.com/SocioProphet/cloudshell-fog.git
cd cloudshell-fog

# Build the gateway binary
make build

# Build the web UI
make frontend

# Start in dev mode
make run-dev
```

Open <http://localhost:8080> — you should see a terminal connected to a shell.

> **What just happened?** `make run-dev` sets `USE_STUB_CONNECTOR=1`, so sessions are backed by a local `/bin/sh` process instead of a Kubernetes pod. The dev auth shim skips OIDC validation. This is safe for local experimentation only.

### 2 — Build and run as a Docker container

```bash
make docker-build
docker run -p 8080:8080 \
  -e USE_STUB_CONNECTOR=1 \
  cloudshell-fog:dev
```

### 3 — Run tests

```bash
make test   # Go unit tests
make vet    # go vet static analysis
```

For production deployment on Kubernetes see the [Deployment](#deployment) section and [`deploy/README.md`](deploy/README.md).

## Configuration

The gateway is configured via environment variables:

| Variable | Default | Description |
|---|---|---|
| `LISTEN_ADDR` | `:8080` | TCP address the HTTP server listens on |
| `GATEWAY_URL` | `ws://localhost:8080` | Public base URL used to construct WebSocket attach URLs |
| `POLICY_CONFIG` | `config/policy.yaml` | Path to the policy YAML file |
| `CLOUD_FALLBACK_REGION` | `us-east-1` | Region used when no fog node is available |
| `OIDC_ISSUER_URL` | _(unset — dev shim)_ | OIDC provider issuer URL (e.g. `https://accounts.example.com`) |
| `OIDC_CLIENT_ID` | _(unset — dev shim)_ | OIDC client ID for token validation |
| `SESSION_TOKEN_SIGNING_KEY` | _(random 32 bytes)_ | HMAC key for minting short-lived session tokens |
| `USE_STUB_CONNECTOR` | _(unset)_ | Set to `1` to force the no-op stub connector (dev/test only) |
| `USE_K8S` | _(unset)_ | Set to `1` to use the Kubernetes runtime connector |

> When neither `USE_STUB_CONNECTOR` nor `USE_K8S` is set, the gateway defaults to the stub connector and emits a warning. Always set one explicitly in production.

See [`docs/reference/configuration.md`](docs/reference/configuration.md) for the full reference, including Kubernetes Secret injection and recommended production values.

### Policy file

`config/policy.yaml` defines named resource profiles:

```yaml
profiles:
  - name: default
    cpu: "500m"
    memory: "512Mi"
    storage: "1Gi"
    max_ttl_seconds: 3600
    max_sessions: 3
    allowed_groups: []   # empty = all authenticated users

  - name: large
    cpu: "2"
    memory: "2Gi"
    storage: "10Gi"
    max_ttl_seconds: 7200
    max_sessions: 1
    allowed_groups: [power-users, admins]
```

Profiles are matched by name in the `POST /v1/sessions` request. `allowed_groups` is matched against the `groups` claim of the OIDC token; an empty list permits any authenticated user.

## API Reference

All session endpoints require an OIDC `Bearer` token in the `Authorization` header.

See [`docs/reference/api.md`](docs/reference/api.md) for the full reference including error codes, curl examples, and the WebSocket frame schema.

### `POST /v1/sessions`

Create a new shell session.

**Request body**

```json
{
  "profile": "default",
  "ttl_seconds": 3600,
  "placement_hint": "eu-west-1",
  "image_ref": "ghcr.io/socioprophet/cloudshell-runtime:latest"
}
```

**Response `201 Created`**

```json
{
  "session_id": "550e8400-...",
  "attach": {
    "ws_url": "wss://shell.example.com/v1/sessions/550e8400-.../pty",
    "token": "<short-lived JWT>",
    "expires_at": "2026-01-01T00:15:00Z"
  },
  "placement": "eu-west-1"
}
```

**Quick example**

```bash
TOKEN=$(curl -s https://accounts.example.com/token ...)
curl -s -X POST https://shell.example.com/v1/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"profile":"default","ttl_seconds":3600}' | jq .
```

### `GET /v1/sessions/{id}`

Get session status.

**Response `200 OK`**

```json
{
  "session_id": "550e8400-...",
  "status": "running",
  "placement": "eu-west-1",
  "created_at": "2026-01-01T00:00:00Z",
  "expires_at": "2026-01-01T01:00:00Z"
}
```

### `DELETE /v1/sessions/{id}`

Terminate a session.

**Response `200 OK`**

```json
{ "terminated": true }
```

### `GET /v1/sessions/{id}/pty?token=<JWT>` (WebSocket)

Attach a PTY. Authentication is via the short-lived `token` query parameter minted by `POST /v1/sessions`.

**Frame schema (JSON over WebSocket)**

| Direction | Frame |
|---|---|
| Client → Server | `{"type":"stdin","data_b64":"<base64>"}` |
| Client → Server | `{"type":"resize","cols":220,"rows":50}` |
| Server → Client | `{"type":"stdout","data_b64":"<base64>"}` |
| Server → Client | `{"type":"exit","code":0}` |

## Architecture

See [`docs/architecture.md`](docs/architecture.md) for a detailed description of all components and their interactions, including data-flow diagrams for session creation and PTY attach.

**Component summary:**

| Package | Responsibility |
|---|---|
| `cmd/gateway` | Binary entrypoint; wires components, registers routes, graceful shutdown |
| `internal/auth` | OIDC token validation; short-lived session token minting |
| `internal/api` | HTTP session management (`POST`, `GET`, `DELETE /v1/sessions`) |
| `internal/pty` | WebSocket PTY handler; frame pumping; resize forwarding |
| `internal/session` | In-memory session store; TTL sweeper |
| `internal/placement` | Fog-aware placement engine; fog-first, cloud-fallback |
| `internal/policy` | YAML-driven admission policy; quota and group checks |
| `internal/connector` | Runtime connector interface; Kubernetes and stub implementations |
| `internal/audit` | Structured audit-event emission via `slog` |
| `internal/otel` | OpenTelemetry provider setup (traces + metrics) |

Specification documents:

- [`docs/spec/interfaces-v1.md`](docs/spec/interfaces-v1.md) — standards-aligned interface contracts (OIDC, WebSocket frame schema, security invariants).
- [`docs/spec/interfaces-v0.md`](docs/spec/interfaces-v0.md) — v0 interface contracts (historical reference).
- [`docs/spec/minimum-spec-v0.md`](docs/spec/minimum-spec-v0.md) — original planes-and-components design notes.

## Deployment

### Kubernetes

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/rbac.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
kubectl apply -f deploy/k8s/networkpolicy.yaml
```

Set the `SESSION_TOKEN_SIGNING_KEY` secret before deploying:

```bash
kubectl -n cloudshell-system create secret generic cloudshell-secrets \
  --from-literal=session-token-signing-key=$(openssl rand -hex 32)
```

### Argo CD

```bash
kubectl apply -f deploy/argocd/appproject.yaml
kubectl apply -f deploy/argocd/application.yaml
```

### CI/CD with Tekton Chains

See [`deploy/tekton/`](deploy/tekton/) for the build pipeline and Tekton Chains supply-chain configuration.

See [`deploy/README.md`](deploy/README.md) for the full deployment guide including ingress configuration, OIDC setup, and the production hardening checklist.

## Guides

Step-by-step how-to guides for common tasks:

| Guide | Description |
|---|---|
| [Getting Started](docs/guides/getting-started.md) | Full walkthrough from clone to working shell in Kubernetes |
| [OIDC Configuration](docs/guides/oidc-configuration.md) | Connect cloudshell-fog to Keycloak, Dex, Okta, or any OIDC provider |
| [Observability](docs/guides/observability.md) | Collect and query OTel traces and metrics with Prometheus and Jaeger |
| [Troubleshooting](docs/guides/troubleshooting.md) | Diagnose common issues with sessions, placement, and connectivity |

Reference documentation:

| Reference | Description |
|---|---|
| [API Reference](docs/reference/api.md) | Complete HTTP and WebSocket API reference with curl examples |
| [Configuration Reference](docs/reference/configuration.md) | All environment variables, policy file schema, and recommended production settings |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to contribute to this project.

## Security

See [SECURITY.md](SECURITY.md) for the security policy and how to report vulnerabilities.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a history of notable changes.

## License

MIT License — see [LICENSE](LICENSE).
