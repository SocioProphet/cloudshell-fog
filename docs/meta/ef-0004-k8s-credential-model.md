# EF-0004 — Kubernetes credential model

## Problem

The repository currently contains:
- a Kubernetes connector implementation (client-go)
- a hardened ServiceAccount configuration (`automountServiceAccountToken: false`)

This creates ambiguity about how the gateway authenticates to the cluster.

## Supported credential modes

### 1. In-cluster ServiceAccount (recommended for cluster deployment)

Requirements:
- ServiceAccount token projection must be enabled (default in Kubernetes unless explicitly disabled)
- `automountServiceAccountToken` must be `true` OR a projected token volume must be mounted

Resolution path:
- `rest.InClusterConfig()`

### 2. External kubeconfig (recommended for local/dev or out-of-cluster gateway)

Requirements:
- `KUBECONFIG` environment variable points to valid kubeconfig file

Resolution path:
- `clientcmd.BuildConfigFromFlags()`

## Required repo alignment

- If in-cluster mode is desired:
  - update `deploy/k8s/rbac.yaml` to allow token mounting OR explicitly mount projected token

- If kubeconfig mode is desired:
  - update deployment to mount kubeconfig secret/configmap

## Recommendation

Default behavior:
- try `KUBECONFIG`
- fallback to in-cluster config

Explicit production posture should choose ONE path and document it clearly.
