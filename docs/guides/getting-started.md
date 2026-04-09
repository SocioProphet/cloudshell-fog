# Getting Started with cloudshell-fog

This guide walks you through every step from cloning the repository to having a working cloudshell-fog deployment on Kubernetes with real OIDC authentication.

If you only want to experiment locally without Kubernetes or OIDC, jump to the [Local Development](#local-development) section.

## Prerequisites

| Tool | Minimum version | Purpose |
|---|---|---|
| [Go](https://go.dev/dl) | 1.22 | Build the gateway binary |
| [Node.js](https://nodejs.org) | 20 | Build the web UI |
| [Docker](https://docs.docker.com/get-docker/) | any recent | Build the container image |
| [kubectl](https://kubernetes.io/docs/tasks/tools/) | 1.28 | Deploy to Kubernetes |
| [make](https://www.gnu.org/software/make/) | 3.8+ | Run build targets |

---

## Local Development

Local development uses a **stub connector** (a real `/bin/sh` on your machine) and a **dev auth shim** that bypasses OIDC. This is useful for iterating on the gateway logic without any external dependencies.

### 1. Clone and build

```bash
git clone https://github.com/SocioProphet/cloudshell-fog.git
cd cloudshell-fog

# Compile the gateway binary to ./bin/gateway
make build

# Install npm dependencies and build the xterm.js web bundle
make frontend
```

### 2. Start the gateway

```bash
make run-dev
```

This sets `USE_STUB_CONNECTOR=1` and starts the gateway on <http://localhost:8080>.

### 3. Open the shell

Navigate to <http://localhost:8080> in your browser. The web UI will automatically create a session and attach a terminal.

To confirm the gateway is healthy:

```bash
curl -s http://localhost:8080/healthz
# → {"status":"ok"}
```

### 4. Create a session via the API

The dev auth shim accepts any `Bearer` token:

```bash
curl -s -X POST http://localhost:8080/v1/sessions \
  -H "Authorization: Bearer dev-token" \
  -H "Content-Type: application/json" \
  -d '{"profile":"default","ttl_seconds":600}' | jq .
```

You will receive a response with a `ws_url` and a short-lived `token` for PTY attach.

---

## Kubernetes Deployment

### Step 1: Build and push the container image

```bash
# Build the image
make docker-build

# Tag and push to your registry (replace with your registry)
docker tag cloudshell-fog:dev registry.example.com/cloudshell-fog:latest
docker push registry.example.com/cloudshell-fog:latest
```

Update the `image:` field in `deploy/k8s/deployment.yaml` to point to your registry.

> **Production tip:** Use a pinned digest (`image@sha256:...`) instead of a mutable tag to satisfy the supply-chain policy and prevent image substitution attacks.

### Step 2: Create the namespace and RBAC

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/rbac.yaml
```

This creates the `cloudshell-system` namespace and a least-privilege service account that the gateway uses to manage session pods.

### Step 3: Generate the session-token signing key

The gateway mints short-lived session tokens using an HMAC key. Generate a cryptographically strong key and store it as a Kubernetes Secret:

```bash
kubectl -n cloudshell-system create secret generic cloudshell-secrets \
  --from-literal=session-token-signing-key=$(openssl rand -hex 32)
```

> Store this key in your secret manager (e.g. Vault, AWS Secrets Manager). Rotating it invalidates all active session tokens.

### Step 4: Configure OIDC

Edit `deploy/k8s/deployment.yaml` and uncomment the OIDC environment variables:

```yaml
env:
  - name: OIDC_ISSUER_URL
    value: "https://accounts.example.com"          # your provider's issuer URL
  - name: OIDC_CLIENT_ID
    value: "cloudshell"                             # your client ID
  - name: USE_K8S
    value: "1"
  - name: GATEWAY_URL
    value: "wss://shell.example.com"
```

For a step-by-step walkthrough of configuring specific OIDC providers (Keycloak, Dex, Okta) see [OIDC Configuration](oidc-configuration.md).

### Step 5: Deploy

```bash
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
kubectl apply -f deploy/k8s/networkpolicy.yaml
```

Wait for the deployment to become ready:

```bash
kubectl -n cloudshell-system rollout status deployment/cloudshell-gateway
```

### Step 6: Expose the gateway

The default Service is `ClusterIP`. Expose it via an Ingress with WebSocket support:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cloudshell-gateway
  namespace: cloudshell-system
  annotations:
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
spec:
  tls:
    - hosts: [shell.example.com]
      secretName: cloudshell-tls
  rules:
    - host: shell.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: cloudshell-gateway
                port:
                  number: 80
```

Apply with `kubectl apply -f ingress.yaml`.

### Step 7: Verify

```bash
# Health check
curl -s https://shell.example.com/healthz

# Create a session (replace TOKEN with a real OIDC access token)
curl -s -X POST https://shell.example.com/v1/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"profile":"default","ttl_seconds":3600}' | jq .
```

Open `https://shell.example.com` in your browser and log in with your OIDC provider.

---

## Production Hardening Checklist

Before going to production, work through this checklist:

- [ ] Replace `latest` image tag with a pinned digest (`image@sha256:...`).
- [ ] Configure real OIDC (`OIDC_ISSUER_URL`, `OIDC_CLIENT_ID`).
- [ ] Generate and store a strong `SESSION_TOKEN_SIGNING_KEY` in a secret manager.
- [ ] Set `GATEWAY_URL` to the public `wss://` URL.
- [ ] Add TLS termination at the ingress layer.
- [ ] Enable Tekton Chains signing and verify attestations at admission time.
- [ ] Review and tighten NetworkPolicies for session namespaces.
- [ ] Set up monitoring — see [Observability](observability.md).

---

## Next Steps

- [OIDC Configuration](oidc-configuration.md) — connect to Keycloak, Dex, or Okta
- [Observability](observability.md) — collect traces and metrics
- [Troubleshooting](troubleshooting.md) — diagnose common issues
- [API Reference](../reference/api.md) — full HTTP and WebSocket API
