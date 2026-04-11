# Connector Mode + Kubernetes Credential Design v0 — cloudshell-fog

Purpose: define the exact behavioral contract for the first implementation patch wave around connector selection and Kubernetes access.

## 1. Problem statement

Current repository state:
- docs and config reference `USE_STUB_CONNECTOR` / `USE_K8S`
- gateway entrypoint still wires the stub connector unconditionally (per reviewed main.go)
- `internal/connector/k8s.go` exists but is not clearly reachable from the current entrypoint
- deployment hardening disables ServiceAccount token automount, which may block in-cluster client-go auth if not handled explicitly

This document defines the intended fix without changing the overall architecture.

## 2. Design goals

1. make connector selection explicit, deterministic, and observable at startup
2. avoid silent fallback from `k8s` to `stub`
3. keep local-dev easy
4. make the k8s credential model explicit enough that deployment manifests and code stop disagreeing
5. minimize blast radius: patch the seam, do not redesign the whole service

## 3. Connector mode contract

### 3.1 Canonical env var
Introduce:
- `CONNECTOR_MODE`

Allowed values:
- `stub`
- `k8s`

Optional future value:
- `auto`

### 3.2 Backward compatibility
For one migration window, preserve legacy env vars:
- `USE_STUB_CONNECTOR=1`
- `USE_K8S=1`

Resolution order (deterministic):
1. if `CONNECTOR_MODE` is set, it wins
2. else if `USE_K8S=1`, mode = `k8s`
3. else if `USE_STUB_CONNECTOR=1`, mode = `stub`
4. else default = `stub` for local developer convenience, but emit a clear warning

### 3.3 Startup rules
- `stub` mode: start with `connector.NewStubConnector()`
- `k8s` mode: attempt to build a Kubernetes connector; failure is fatal
- unknown mode: fatal startup error

### 3.4 Logging rules
Startup MUST log:
- selected connector mode
- credential source for `k8s` mode
- namespace prefix for k8s runtime allocation

## 4. Kubernetes credential model

### 4.1 Supported credential sources
Support exactly two credential modes for v0:

A. `in-cluster`
- uses `rest.InClusterConfig()`
- intended for normal cluster-local gateway deployment

B. `kubeconfig`
- uses an explicit kubeconfig file path mounted into the gateway container
- intended for local/dev or unusual control-plane topologies

### 4.2 Canonical env vars
Introduce:
- `K8S_AUTH_MODE` = `in-cluster` | `kubeconfig`
- `KUBECONFIG_PATH` = path to kubeconfig when `K8S_AUTH_MODE=kubeconfig`
- `K8S_NAMESPACE_PREFIX` = default `cloudshell-`

### 4.3 Resolution rules
When `CONNECTOR_MODE=k8s`:
- if `K8S_AUTH_MODE=in-cluster` or unset, use `rest.InClusterConfig()`
- if `K8S_AUTH_MODE=kubeconfig`, require `KUBECONFIG_PATH`
- if credential construction fails, startup fails

### 4.4 Deployment implications
If using `in-cluster` auth:
- the gateway pod must have a projected or mounted ServiceAccount token available to client-go
- current manifest hardening around token automount must be revisited and made explicit

If using `kubeconfig` auth:
- deploy docs must explain secret/config mount
- this mode should be described as non-default for production in-cluster deployments

## 5. Recommended v0 operational choice

Recommended default for real cluster deployments:
- `CONNECTOR_MODE=k8s`
- `K8S_AUTH_MODE=in-cluster`

Recommended default for local development:
- `CONNECTOR_MODE=stub`

## 6. Minimal code delta shape

### 6.1 Gateway entrypoint
Add a small resolver function in `cmd/gateway/main.go` (or nearby helper) that:
- reads env vars
- resolves connector mode
- resolves k8s credential source if needed
- returns `connector.Connector`

Pseudo-shape:
- `resolveConnector(ctx, logger) (connector.Connector, error)`

### 6.2 No architectural changes required
This should not require changes to:
- API handler contract
- PTY handler contract
- session store contract
- placement engine contract

The patch is a seam fix, not a redesign.

## 7. Documentation changes required

Update at least:
- `README.md`
- `docs/reference/configuration.md`
- `deploy/README.md`
- `deploy/k8s/deployment.yaml`

Docs should clearly distinguish:
- dev stub mode
- real k8s mode
- in-cluster auth vs kubeconfig auth

## 8. Acceptance checklist

1. `CONNECTOR_MODE=stub` starts successfully and uses stub connector.
2. `CONNECTOR_MODE=k8s` with invalid credentials fails at startup.
3. `CONNECTOR_MODE=k8s` with valid in-cluster config starts successfully.
4. startup logs show selected connector mode and auth mode.
5. docs and manifests match actual code behavior.

## 9. Out-of-scope for this patch

Not part of the first connector patch:
- changing the session API response shape
- rewriting placement logic
- changing the PTY frame schema
- replacing the in-memory session store
