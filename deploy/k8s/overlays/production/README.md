# Production overlay guidance

This overlay makes the production image posture explicit without mutating the dev/demo base manifests.

## Goals

- Do not advertise `:latest` as production-safe.
- Make immutable image references first-class.
- Keep the base manifest usable for local/demo flows while giving production a concrete overlay path.

## Files

- `kustomization.yaml` — composes the base deployment and production image patch.
- `patch-deployment-image.yaml` — replaces mutable gateway/runtime image refs with digest-form placeholders.

## Required production changes

Before use, replace all `REPLACE_WITH_REAL_DIGEST` placeholders with real image digests:

```yaml
image: ghcr.io/socioprophet/cloudshell-fog/gateway@sha256:<real-digest>
```

and:

```yaml
- name: RUNTIME_IMAGE_REF
  value: ghcr.io/socioprophet/cloudshell-fog/runtime@sha256:<real-digest>
```

## Recommendation

Treat `deploy/k8s/` as base/dev-oriented.
Treat `deploy/k8s/overlays/production/` as the production starting point for image pinning and stricter environment-specific settings.

Production deployments should also compose the appropriate connector overlay:

- `deploy/k8s/overlays/k8s-connector-incluster/`
- `deploy/k8s/overlays/k8s-connector-kubeconfig/`

## Why placeholders are used

The repository cannot know the digest of a future built image at authoring time. The important production invariant is not the literal placeholder; it is that production deployment paths require digest-form image references instead of mutable tags.
