# In-cluster Kubernetes connector overlay

Use this overlay when the gateway runs **inside** the target cluster and should use in-cluster Kubernetes credentials.

## What this overlay does

- sets `CONNECTOR_MODE=k8s`
- enables ServiceAccount token mounting for the gateway ServiceAccount

## Why this exists

The base manifests are intentionally conservative and currently disable automatic ServiceAccount token mounting. That is safer by default, but it blocks the standard `rest.InClusterConfig()` path used by the Kubernetes connector.

This overlay makes that credential path explicit instead of implicit.

## Apply with Argo CD or Kustomize

Point the deployment tooling at:

```text
deploy/k8s/overlays/k8s-connector-incluster/
```
