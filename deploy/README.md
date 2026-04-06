# Deployment Guide

This directory contains deployment manifests for cloudshell-fog.

## Directory Layout

```
deploy/
  k8s/            Kubernetes manifests (apply directly or via Argo CD)
  argocd/         Argo CD Application and AppProject definitions
  tekton/         Tekton build pipeline and Tekton Chains supply-chain config
```

## Kubernetes

### Prerequisites

- Kubernetes 1.28+
- `kubectl` configured against your target cluster
- An OIDC provider accessible to the cluster nodes

### Steps

```bash
# 1. Create the namespace
kubectl apply -f k8s/namespace.yaml

# 2. Create RBAC (ServiceAccount, ClusterRole, ClusterRoleBinding, ConfigMap, Secret)
kubectl apply -f k8s/rbac.yaml

# 3. Generate and store the session-token signing key
kubectl -n cloudshell-system create secret generic cloudshell-secrets \
  --from-literal=session-token-signing-key=$(openssl rand -hex 32)

# 4. Deploy the gateway
kubectl apply -f k8s/deployment.yaml

# 5. Create the Service
kubectl apply -f k8s/service.yaml

# 6. Apply NetworkPolicies
kubectl apply -f k8s/networkpolicy.yaml
```

### Configuring OIDC

Edit `k8s/deployment.yaml` and uncomment the `OIDC_ISSUER_URL` and `OIDC_CLIENT_ID` environment variables:

```yaml
- name: OIDC_ISSUER_URL
  value: "https://accounts.example.com"
- name: OIDC_CLIENT_ID
  value: "cloudshell"
```

### Exposing the Gateway

The Service is `ClusterIP` by default. Expose it externally via an Ingress (requires a WebSocket-capable ingress controller):

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

Set `GATEWAY_URL` in the Deployment to `wss://shell.example.com`.

## Argo CD

### Prerequisites

- Argo CD installed in your cluster (v2.10+)
- The application namespace `cloudshell-system` already exists

### Steps

```bash
# 1. Create the AppProject (restricts source repos and destination namespaces)
kubectl apply -f argocd/appproject.yaml

# 2. Create the Application (points at this repository)
kubectl apply -f argocd/application.yaml
```

Argo CD will sync the `deploy/k8s/` manifests and keep them reconciled.

## Tekton CI/CD with Supply-Chain Attestations

### Prerequisites

- Tekton Pipelines v0.59+
- Tekton Chains v0.20+ (for provenance attestation and signing)
- A container registry accessible from the cluster

### Steps

```bash
# 1. Install the build task
kubectl apply -f tekton/task-build.yaml

# 2. Install the pipeline
kubectl apply -f tekton/pipeline.yaml

# 3. Apply the Tekton Chains configuration
kubectl apply -f tekton/chains-config.yaml
```

Tekton Chains automatically signs the resulting OCI image and produces an in-toto provenance attestation. The admission policy (see `config/policy.yaml` and `deploy/k8s/rbac.yaml`) can be extended to require the attestation before a session pod is scheduled.

## Production Hardening Checklist

- [ ] Replace `latest` image tags with pinned digests (`image@sha256:...`).
- [ ] Configure real OIDC (`OIDC_ISSUER_URL`, `OIDC_CLIENT_ID`).
- [ ] Generate and store a strong `SESSION_TOKEN_SIGNING_KEY` in a secret manager.
- [ ] Set `GATEWAY_URL` to the public `wss://` URL.
- [ ] Add TLS termination at the ingress layer.
- [ ] Enable Tekton Chains signing and verify attestations at admission time.
- [ ] Review and tighten NetworkPolicies for session namespaces.
- [ ] Set up monitoring: scrape OpenTelemetry metrics from the gateway.
