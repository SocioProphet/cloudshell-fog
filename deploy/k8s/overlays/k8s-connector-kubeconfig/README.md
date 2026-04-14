# Kubeconfig-based Kubernetes connector overlay

Use this overlay when the gateway runs **outside** the target cluster or should use an explicitly mounted kubeconfig instead of in-cluster credentials.

## What this overlay does

- sets `CONNECTOR_MODE=k8s`
- sets `KUBECONFIG=/var/run/cloudshell/kubeconfig/config`
- mounts a secret named `cloudshell-kubeconfig`

## Secret expectation

The target namespace must contain a secret:

```text
cloudshell-kubeconfig
```

with key:

```text
config
```

containing a valid kubeconfig file.
