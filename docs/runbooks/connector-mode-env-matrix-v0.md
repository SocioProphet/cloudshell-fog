# Connector Mode Environment Matrix v0

This runbook summarizes the environment variables and deployment expectations for the supported gateway connector modes.

## 1. Stub mode

Purpose:
- local development
- API / PTY contract testing without Kubernetes dependency

Required env:
- `CONNECTOR_MODE=stub` (or unset, if default remains stub)
- `GATEWAY_URL` as appropriate for local UI attach
- `SESSION_TOKEN_SIGNING_KEY` optional in dev, strongly recommended otherwise

Not required:
- `KUBECONFIG`
- in-cluster ServiceAccount credentials

Expected behavior:
- gateway starts without cluster access
- session allocation uses stub connector
- PTY behaves like in-process echo shell

## 2. Kubernetes mode — in-cluster

Purpose:
- gateway runs inside target cluster
- connector uses `rest.InClusterConfig()`

Required env:
- `CONNECTOR_MODE=k8s`
- `GATEWAY_URL`
- `POLICY_CONFIG`

Required deployment posture:
- ServiceAccount token projection/mount enabled
- gateway has RBAC for namespaces / pods / exec
- network and policy artifacts reconciled in cluster

Not required:
- `KUBECONFIG`

Expected behavior:
- gateway selects k8s connector
- runtime sessions are allocated in Kubernetes namespaces

## 3. Kubernetes mode — kubeconfig

Purpose:
- gateway runs outside target cluster or uses explicit kubeconfig

Required env:
- `CONNECTOR_MODE=k8s`
- `KUBECONFIG=/path/to/config`
- `GATEWAY_URL`

Required deployment/runtime posture:
- kubeconfig file mounted or present on host
- cluster credentials valid for namespace/pod/exec operations

Expected behavior:
- gateway selects k8s connector
- kubeconfig path is used before in-cluster config

## 4. Fail-fast expectations

The gateway should fail fast when:
- `CONNECTOR_MODE` is unsupported
- `CONNECTOR_MODE=k8s` but neither a valid `KUBECONFIG` nor valid in-cluster config is available

## 5. Notes

This matrix is an operator-facing restatement of the implementation/helper logic introduced by the connector-mode and credential-model work.
