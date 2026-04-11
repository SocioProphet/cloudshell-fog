# CI/CD Contract v0 — cloudshell-fog

This document defines the build, attestation, promotion, and deployment contract for cloudshell-fog.

## 0. Goal

We want a deployment pipeline that is:

- reproducible
- auditable
- GitOps-driven
- enforceable at admission time
- portable across fog and cloud clusters

## 1. Control model

### 1.1 Source of truth

Git is the source of truth for:

- application manifests
- deployment overlays
- policy definitions
- promotion state

### 1.2 Reconciliation model

Argo CD is the reconciler.

We align to an **Argo CD Autopilot-compatible** layout so that environments and applications are represented declaratively and promoted by Git change rather than console mutation.

## 2. Pipeline model

Tekton is the build and attestation engine.

We split pipelines into two lanes.

### 2.1 Service lane

For the gateway and control-plane services:

1. lint / static analysis
2. unit tests
3. build OCI image
4. generate SBOM
5. sign image digest
6. emit provenance attestation
7. publish digest to registry

### 2.2 Runtime lane

For runtime shell images:

1. build from pinned base image
2. generate SBOM
3. sign image digest
4. emit provenance attestation
5. publish digest to registry

The runtime lane must be treated as security-sensitive because these images are the execution substrate for user sessions.

## 3. Required artifacts

Every production image must have:

- a pinned OCI digest
- an SBOM
- a signature
- a provenance attestation

Recommended defaults:

- SBOM: SPDX
- signature: cosign
- provenance: Tekton Chains / in-toto-style statement

## 4. Admission invariants

A production cluster must reject runtime or gateway images unless all of the following are true:

- image is referenced by digest, not mutable tag
- signature verifies against approved trust roots
- provenance attestation exists and matches the expected repo / pipeline identity
- SBOM exists for the digest

This applies equally to fog-tier and cloud-tier deployments.

## 5. Promotion model

Promotion is a Git change that updates image digests in deployment manifests or environment overlays.

Suggested promotion sequence:

1. dev / local
2. trusted cloud staging
3. attested fog staging
4. production cloud fallback
5. production fog tiers

We prefer to prove images and manifests in a trusted cloud lane before promoting into heterogeneous fog nodes.

## 6. Argo CD layout guidance

Suggested logical layout:

- `apps/` for application definitions
- `infra/` for namespaces, ingress, cert-manager, and shared controllers
- `policy/` for admission and runtime-policy resources
- `environments/` for promotion overlays

The exact directory structure may vary, but the invariant is that promotion remains Git-driven and digest-explicit.

## 7. Tekton Chains role

Tekton Chains is not optional in the production profile.

Its role is to:

- sign pipeline outputs
- emit provenance tied to pipeline execution
- provide a machine-verifiable link between source, pipeline, and digest

Without Chains (or equivalent provenance), the repo may still build, but it does not satisfy the production supply-chain profile.

## 8. Local-development profile

For local development, we may permit reduced guarantees:

- stub connector
- unsigned local images
- no Argo reconciliation
- no admission gating

But local shortcuts must remain clearly separated from the production profile.

## 9. Open questions reserved for later docs

- exact admission controller default (Kyverno vs Gatekeeper)
- keyless vs key-backed signing defaults
- fog-tier promotion gates based on node attestation posture
- rollback semantics for failed fog deployments
