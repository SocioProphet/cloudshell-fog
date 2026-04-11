# Kyverno Baseline

This directory contains Kyverno policy examples for the `cloudshell-fog` production posture.

## Why Kyverno here?

We use Kyverno for the cluster-local enforcement layer because it is Kubernetes-native and maps well to:

- image digest enforcement
- signature and provenance verification
- namespace and runtime hardening
- NetworkPolicy and pod policy adjacency

## Why not only OPA/Rego?

OPA/Rego remains valuable for broader platform policy, but these cluster admission controls are more readable and maintainable in Kyverno for the current repo scope.

## Files

- `require-image-digest.yaml` — requires pinned image digests for Pod images.
- `verify-signed-images.yaml` — verifies image signatures using Kyverno `verifyImages`; intended to be adapted with the real public key or keyless trust configuration.
- `runtime-baseline.yaml` — enforces conservative runtime controls for managed session pods.
- `kustomization.yaml` — bundles the baseline policies.

## Compatibility note

Kyverno currently supports image verification via `verifyImages` on `ClusterPolicy`, and newer `ImageValidatingPolicy` types also exist in modern Kyverno releases. These baseline examples intentionally use `ClusterPolicy` for a conservative compatibility path and to align with the repo's current structure.
