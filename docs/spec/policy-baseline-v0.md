# Policy Baseline v0 — cloudshell-fog

This document defines the baseline policy posture for cloudshell-fog across the gateway, the cluster, and the deployment pipeline.

## 0. Policy layers

We intentionally separate policy into three layers.

### 0.1 Product-local gateway policy

Enforced inside the gateway:

- subject authentication
- allowed profile selection
- TTL bounds
- active-session quotas
- fog/cloud placement constraints
- session ownership checks for PTY attach

This policy is product-specific and latency-sensitive.

### 0.2 Cluster admission policy

Enforced by Kubernetes admission controls:

- pinned digest requirement
- signature verification
- provenance verification
- SBOM requirement
- namespace and pod-security constraints
- NetworkPolicy defaults

This policy is deployment-specific and non-negotiable in production.

### 0.3 Organisational / cross-service policy

Reserved for broader SocioProphet environments where cloudshell-fog participates in a larger governed platform.

Examples:

- environment-level access policy
- inter-service trust assertions
- fleet-wide policy exceptions
- export control / data-sovereignty constraints

## 1. Default recommendation

### 1.1 Kubernetes admission

Default recommendation: **Kyverno**.

Reason:

- Kubernetes-native resource matching
- strong ergonomics for image, namespace, and manifest policy
- good fit for digest / signature / provenance enforcement
- lower friction for teams reading and maintaining cluster policy

### 1.2 Broader policy language

Keep **OPA/Rego** available for cases where cloudshell-fog becomes part of a larger platform policy fabric.

Reason:

- richer logical composition
- reusable across non-Kubernetes contexts
- better fit for organisation-wide evaluation beyond admission hooks

So the baseline is:

- **Kyverno** for cluster admission defaults
- **OPA/Rego** for broader, cross-plane policy where needed

## 2. Minimum production rules

A production deployment must enforce at least:

1. images referenced by digest only
2. approved signatures only
3. provenance attestation required
4. SBOM presence required
5. session runtime pods isolated by namespace and NetworkPolicy
6. no privileged runtime pods
7. least-privilege service accounts
8. explicit egress policy for runtime namespaces

## 3. Gateway policy invariants

The gateway must not allow:

- PTY attach without a valid session-scoped token
- session creation outside configured profile bounds
- TTL above the policy profile maximum
- subject access to profiles they are not entitled to use
- attach to another subject's session

## 4. Runtime policy invariants

Session runtimes should default to:

- non-root where practical
- read-only root filesystem where practical
- bounded CPU / memory / storage from the resolved profile
- default-deny network policy
- explicit DNS and HTTPS egress only unless expanded by policy

## 5. Fog-specific policy concerns

Fog nodes are not all equivalent.

Policy should be able to discriminate on:

- attestation posture
- secure boot / TPM presence
- connectivity health
- jurisdiction / data-sovereignty attributes
- maintenance / quarantine state

These attributes should feed placement decisions and, in stricter environments, admission gating.

## 6. Exceptions model

Exceptions must be explicit, time-bounded, and auditable.

We should not permit silent or indefinite policy bypasses for:

- mutable tags
- unsigned images
- missing provenance
- privileged runtime pods
- unrestricted egress

## 7. Reserved follow-on work

- codify Kyverno baseline rules in `policy/`
- document Rego examples for cross-service governance
- bind fog trust-tier claims into placement and admission
- define exception workflow and evidence requirements
