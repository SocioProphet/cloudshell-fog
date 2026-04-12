# Production overlay guidance

This overlay exists to make the production posture explicit without mutating the current dev-oriented base manifests.

## Goals

- Do not advertise `:latest` as production-safe
- Make immutable image references first-class
- Keep the base manifest usable for local/demo flows while giving production a clear path

## Required production changes

1. Replace mutable image tags with immutable digest references.
2. Keep runtime hardening and policy enforcement enabled.
3. Ensure Argo CD points at this overlay (or an equivalent production overlay) rather than the dev/demo base.

## Example image form

```yaml
image: ghcr.io/socioprophet/cloudshell-fog/gateway@sha256:REPLACE_WITH_REAL_DIGEST
```

## Recommendation

Treat `deploy/k8s/` as base/dev-oriented.
Treat `deploy/k8s/overlays/production/` as the place where production-safe image pinning and stricter environment-specific settings live.
