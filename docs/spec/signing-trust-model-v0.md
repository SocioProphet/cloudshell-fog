# Signing Trust Model v0 — cloudshell-fog

## Purpose

Define how cloudshell-fog should verify signed images in a way that is production-usable rather than placeholder-only.

## Current situation

The repository contains both:
- a public-key based Kyverno policy variant, which remains useful as a compatibility/example path but still requires real trust material before production use
- a keyless Kyverno policy variant, which is the recommended baseline for CI-produced images

The root policy bundle now selects the keyless variant by default for GitOps reconciliation.

## Supported trust models

### 1. Public-key based verification

Use an explicitly managed cosign public key.

Pros:
- simple mental model
- deterministic trust root

Cons:
- key rotation burden
- secret/material distribution burden

Use this mode only after replacing placeholder public key material with a real trusted key and documenting key rotation.

### 2. Keyless verification (recommended baseline)

Use Sigstore keyless signing plus identity constraints tied to the CI system and repository.

Pros:
- avoids distributing long-lived private signing keys
- aligns with Tekton Chains + OIDC-based provenance flow
- easier to scale across repos and builders

Cons:
- requires careful identity constraint definition
- requires Rekor/Fulcio/TUF availability assumptions

## Recommended baseline

For cloudshell-fog v0/v1:
- prefer **keyless verification** for CI-produced images
- keep public-key verification as a fallback or compatibility mode
- keep immutable digest enforcement enabled alongside signature verification

## Required constraints

Whatever verification mode is used, policy should constrain at least:
- allowed image registry prefixes
- expected signer identity or trusted key material
- expected repository / workflow / builder context where possible
- immutable digest usage

## Repo status

The current GitOps policy bundle selects:
- `policy/kyverno/require-image-digest.yaml`
- `policy/kyverno/verify-signed-images-keyless.yaml`
- `policy/kyverno/runtime-baseline.yaml`

The placeholder public-key policy remains checked in for compatibility and examples, but is not the bundled default.
