# Policy Baseline Assets

This directory contains implementation-facing policy assets that correspond to the clean-room specification set in `docs/spec/`.

## Purpose

The goal is to move from policy discussion to enforceable controls.

Current scope:

- Kubernetes admission policy examples for image and runtime enforcement
- documentation for the chosen split between gateway-local policy and cluster policy
- a root Kustomize bundle for GitOps reconciliation
- a place to grow future organisation-level policy assets

## Structure

- `kustomization.yaml` — root policy bundle intended for Argo CD reconciliation.
- `kyverno/` — Kubernetes-native admission policy examples and baseline manifests.

## Current posture

The current baseline follows the policy split described in `docs/spec/policy-baseline-v0.md`:

- gateway-local policy remains in product code
- cluster admission policy is implemented here
- broader organisation/cross-service policy remains a later integration layer

The root `policy/kustomization.yaml` points at the Kyverno bundle, and `policy/kyverno/kustomization.yaml` intentionally selects the keyless image-verification policy variant:

- `kyverno/require-image-digest.yaml`
- `kyverno/verify-signed-images-keyless.yaml`
- `kyverno/runtime-baseline.yaml`

The public-key policy remains available as a compatibility/example path, but it contains placeholder public key material and should not be treated as the default production bundle until real trust material is supplied.

## Default engine choice

The default recommendation is:

- **Kyverno** for Kubernetes-native admission and runtime policy
- **OPA/Rego** retained for wider, cross-plane governance work where needed

## GitOps posture

The Argo CD policy application points at this `policy/` subtree. The root Kustomize file therefore defines the policy bundle reconciled by GitOps.

This closes the old ambiguity where policy assets existed but the policy root did not declare which policy set should be applied.

## Important notes

These policies are intentionally conservative and are meant to be reviewed and tuned before production rollout.

The recommended production baseline is keyless Sigstore-style image verification plus immutable image digest enforcement and runtime hardening.
