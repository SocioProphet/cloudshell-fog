# Policy Baseline Assets

This directory contains implementation-facing policy assets that correspond to the clean-room specification set in `docs/spec/`.

## Purpose

The goal is to move from policy discussion to enforceable controls.

Current scope:

- Kubernetes admission policy examples for image and runtime enforcement
- documentation for the chosen split between gateway-local policy and cluster policy
- a place to grow future organisation-level policy assets

## Structure

- `kyverno/` — Kubernetes-native admission policy examples and baseline manifests

## Current posture

The current baseline follows the policy split described in `docs/spec/policy-baseline-v0.md`:

- gateway-local policy remains in product code
- cluster admission policy is implemented here
- broader organisation/cross-service policy remains a later integration layer

## Default engine choice

The default recommendation is:

- **Kyverno** for Kubernetes-native admission and runtime policy
- **OPA/Rego** retained for wider, cross-plane governance work where needed

## Important notes

These policies are intentionally conservative and are meant to be reviewed and tuned before production rollout.

They are not yet wired into the deployment docs or Argo application surface automatically. That is a follow-on integration task.
