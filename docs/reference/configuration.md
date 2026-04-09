# Configuration Reference

This document is the complete reference for all cloudshell-fog configuration options.

## Environment variables

### Server

| Variable | Default | Description |
|---|---|---|
| `LISTEN_ADDR` | `:8080` | TCP address the HTTP server binds to. Use `0.0.0.0:8080` to bind all interfaces explicitly. |
| `GATEWAY_URL` | `ws://localhost:8080` | Public base URL used to construct WebSocket attach URLs returned in API responses. In production set to the public `wss://` URL (e.g. `wss://shell.example.com`). |

### Authentication

| Variable | Default | Description |
|---|---|---|
| `OIDC_ISSUER_URL` | _(unset)_ | OIDC provider issuer URL. When set alongside `OIDC_CLIENT_ID`, real OIDC validation is active. When unset, the dev auth shim accepts any token — **never leave unset in production**. |
| `OIDC_CLIENT_ID` | _(unset)_ | OAuth 2.0 client ID. Must match the `aud` claim in access tokens. |
| `SESSION_TOKEN_SIGNING_KEY` | _(random 32 bytes per restart)_ | Hex-encoded HMAC key (minimum 32 bytes) used to mint and verify short-lived session tokens. If left unset, a random key is generated at startup — this invalidates all session tokens on pod restart. **Always set this in production via a Kubernetes Secret.** |

### Connectors

| Variable | Default | Description |
|---|---|---|
| `USE_STUB_CONNECTOR` | _(unset)_ | Set to `1` to use the no-op stub connector. The stub opens a `/bin/sh` process on the gateway host for each session. **Development and testing only.** |
| `USE_K8S` | _(unset)_ | Set to `1` to use the Kubernetes runtime connector. The connector creates a per-session namespace and pod. Requires the gateway to run with the RBAC permissions in `deploy/k8s/rbac.yaml`. |

If neither variable is set, the gateway defaults to the stub connector and logs a warning.

### Placement

| Variable | Default | Description |
|---|---|---|
| `CLOUD_FALLBACK_REGION` | `us-east-1` | Region identifier used for placement when no fog nodes are available or healthy. This value is returned as the `placement` field in session responses. |

### Policy

| Variable | Default | Description |
|---|---|---|
| `POLICY_CONFIG` | `config/policy.yaml` | Filesystem path to the policy YAML file. Relative paths are resolved from the gateway's working directory. |

---

## Policy file schema

`config/policy.yaml` defines the resource profiles available to users.

### Full schema

```yaml
profiles:
  - name: string              # Required. Profile identifier used in POST /v1/sessions.
    cpu: string               # Required. Kubernetes CPU resource request/limit (e.g. "500m", "2").
    memory: string            # Required. Kubernetes memory request/limit (e.g. "512Mi", "2Gi").
    storage: string           # Required. Ephemeral storage limit (e.g. "1Gi", "10Gi").
    max_ttl_seconds: integer  # Required. Maximum allowed TTL for a session using this profile.
    max_sessions: integer     # Required. Maximum concurrent sessions per user for this profile.
    allowed_groups:           # Optional. List of OIDC group names allowed to use this profile.
      - string                #   Empty list means any authenticated user may use this profile.
```

### Example

```yaml
profiles:
  - name: default
    cpu: "500m"
    memory: "512Mi"
    storage: "1Gi"
    max_ttl_seconds: 3600
    max_sessions: 3
    allowed_groups: []

  - name: large
    cpu: "2"
    memory: "2Gi"
    storage: "10Gi"
    max_ttl_seconds: 7200
    max_sessions: 1
    allowed_groups: [power-users, admins]

  - name: ci
    cpu: "1"
    memory: "1Gi"
    storage: "5Gi"
    max_ttl_seconds: 1800
    max_sessions: 5
    allowed_groups: [ci-bots]
```

### Profile selection

The `profile` field in `POST /v1/sessions` must exactly match a profile `name`. If no match is found, the request is rejected with `403 Forbidden`.

### Group matching

`allowed_groups` is matched against the `groups` claim of the OIDC access token. Matching is case-sensitive and exact (no wildcard or prefix support). An empty `allowed_groups` list permits any authenticated user.

---

## Kubernetes Secret injection (recommended for production)

Rather than setting environment variables directly in the Deployment manifest, inject sensitive values from Kubernetes Secrets:

```bash
# Session token signing key
kubectl -n cloudshell-system create secret generic cloudshell-secrets \
  --from-literal=session-token-signing-key=$(openssl rand -hex 32)

# OIDC credentials (optional — can also use configmap for non-secret values)
kubectl -n cloudshell-system create secret generic cloudshell-oidc \
  --from-literal=oidc-issuer-url=https://accounts.example.com \
  --from-literal=oidc-client-id=cloudshell-gateway
```

Reference in `deploy/k8s/deployment.yaml`:

```yaml
env:
  - name: SESSION_TOKEN_SIGNING_KEY
    valueFrom:
      secretKeyRef:
        name: cloudshell-secrets
        key: session-token-signing-key
  - name: OIDC_ISSUER_URL
    valueFrom:
      secretKeyRef:
        name: cloudshell-oidc
        key: oidc-issuer-url
  - name: OIDC_CLIENT_ID
    valueFrom:
      secretKeyRef:
        name: cloudshell-oidc
        key: oidc-client-id
```

---

## Recommended production settings

```bash
LISTEN_ADDR=:8080
GATEWAY_URL=wss://shell.example.com
OIDC_ISSUER_URL=https://accounts.example.com
OIDC_CLIENT_ID=cloudshell-gateway
SESSION_TOKEN_SIGNING_KEY=<from secret manager>
USE_K8S=1
CLOUD_FALLBACK_REGION=us-east-1
POLICY_CONFIG=/etc/cloudshell/policy.yaml
```

Mount `policy.yaml` as a ConfigMap to allow hot-reload without a redeployment:

```bash
kubectl -n cloudshell-system create configmap cloudshell-policy \
  --from-file=policy.yaml=config/policy.yaml

# Reference in deployment.yaml
volumes:
  - name: policy
    configMap:
      name: cloudshell-policy
containers:
  - name: gateway
    volumeMounts:
      - name: policy
        mountPath: /etc/cloudshell
    env:
      - name: POLICY_CONFIG
        value: /etc/cloudshell/policy.yaml
```

---

## Related

- [Getting Started](../guides/getting-started.md) — step-by-step deployment walkthrough
- [OIDC Configuration](../guides/oidc-configuration.md) — provider-specific OIDC setup
- [API Reference](api.md) — HTTP and WebSocket API
