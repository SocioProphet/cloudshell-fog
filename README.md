# cloudshell-fog

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CI](https://github.com/SocioProphet/cloudshell-fog/actions/workflows/ci.yml/badge.svg)](https://github.com/SocioProphet/cloudshell-fog/actions/workflows/ci.yml)

**cloudshell-fog** is an open-source, fog-optimised cloud shell gateway. It gives users a browser-accessible terminal that runs as close to their data as possible — on a fog/edge node when available, falling back gracefully to a trusted cloud region.

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

## Features

- **OIDC authentication** — validates short-lived access tokens; dev shim for local work.
- **Fog placement** — selects the nearest healthy node (fog tier first, cloud fallback).
- **Policy engine** — YAML-configured profiles with CPU/RAM/storage quotas and group-based access.
- **Session lifecycle** — TTL-based expiry with automatic runtime cleanup.
- **PTY over WebSocket** — resize, stdin/stdout, exit frames (JSON + base64).
- **Audit trail** — structured log events for every session, placement, and policy decision.
- **OpenTelemetry** — traces, metrics, and structured logs via stdout exporters.
- **GitOps-ready** — Kubernetes manifests, Argo CD Application, and Tekton Chains pipeline included.

## Table of Contents

- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Architecture](#architecture)
- [Deployment](#deployment)
- [Contributing](#contributing)
- [Security](#security)
- [License](#license)

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 20+ (for the web UI)
- Docker (optional, for container builds)

### Run locally (stub connector, no OIDC)

```bash
# 1. Build the Go binary
make build

# 2. Build the web UI
make frontend

# 3. Run in dev mode (stub connector, dev auth shim)
make run-dev
```

Open <http://localhost:8080> in your browser.

### Build the Docker image

```bash
make docker-build
docker run -p 8080:8080 cloudshell-fog:dev
```

### Run tests

```bash
make test
make vet
```

## Configuration

The gateway is configured via environment variables:

| Variable | Default | Description |
|---|---|---|
| `LISTEN_ADDR` | `:8080` | TCP address the HTTP server listens on |
| `GATEWAY_URL` | `ws://localhost:8080` | Public base URL used to construct WebSocket attach URLs |
| `POLICY_CONFIG` | `config/policy.yaml` | Path to the policy YAML file |
| `CLOUD_FALLBACK_REGION` | `us-east-1` | Region used when no fog node is available |
| `OIDC_ISSUER_URL` | _(unset — dev shim)_ | OIDC provider issuer URL |
| `OIDC_CLIENT_ID` | _(unset — dev shim)_ | OIDC client ID for token validation |
| `SESSION_TOKEN_SIGNING_KEY` | _(random 32 bytes)_ | HMAC key for minting short-lived session tokens |
| `USE_STUB_CONNECTOR` | _(unset)_ | Set to `1` to force the no-op stub connector |
| `USE_K8S` | _(unset)_ | Set to `1` to use the Kubernetes runtime connector |

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

## API Reference

All session endpoints require an OIDC `Bearer` token in the `Authorization` header.

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

See [`docs/architecture.md`](docs/architecture.md) for a detailed description of all components and their interactions.

Specification documents:

- [`docs/spec/minimum-spec-v0.md`](docs/spec/minimum-spec-v0.md) — original planes-and-components spec.
- [`docs/spec/interfaces-v0.md`](docs/spec/interfaces-v0.md) — v0 interface contracts.
- [`docs/spec/interfaces-v1.md`](docs/spec/interfaces-v1.md) — v1 standards-aligned interface contracts.

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

See [`deploy/README.md`](deploy/README.md) for full deployment instructions.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to contribute to this project.

## Security

See [SECURITY.md](SECURITY.md) for the security policy and how to report vulnerabilities.

## License

Apache License 2.0 — see [LICENSE](LICENSE).
