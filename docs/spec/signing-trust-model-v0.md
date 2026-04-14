# Signing Trust Model v0 — cloudshell-fog

## Purpose

Define how cloudshell-fog should verify signed images in a way that is production-usable rather than placeholder-only.

## Current situation

The repository already contains a Kyverno policy for signed image verification, but the checked-in trust material is still placeholder data.

## Supported trust models

### 1. Public-key based verification

Use an explicitly managed cosign public key.

Pros:
- simple mental model
- deterministic trust root

Cons:
- key rotation burden
- secret/material distribution burden

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

## Required constraints

Whatever verification mode is used, policy should constrain at least:
- allowed image registry prefixes
- expected signer identity or trusted key material
- expected repository / workflow / builder context where possible
- immutable digest usage

## Repo follow-on

- keep the existing placeholder-based policy as an example only
- add a keyless-oriented policy variant
- document which policy is intended for production
