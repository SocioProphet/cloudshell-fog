# Compliance Profile v0 — FIPS and FedRAMP Compatibility

This document defines the federal-compatible compliance posture for `cloudshell-fog`.

## 0. Scope and posture

`cloudshell-fog` can be designed for:

- **FIPS-aligned cryptographic operation**
- **FedRAMP-compatible deployment and control mapping**

It does **not** become FedRAMP-authorized merely by implementing this profile. Authorization remains a system- and environment-level outcome involving assessed controls, inherited controls, documentation, and continuous monitoring.

## 1. Cryptographic baseline

### 1.1 FIPS posture

The federal profile should use cryptographic services backed by modules or update streams with active validation under the NIST Cryptographic Module Validation Program (CMVP) wherever required by the deployment tier and risk posture.

Practical consequences:

- choose TLS libraries and runtime images with a documented FIPS/CMVP story
- document the exact cryptographic modules used by the gateway and runtime images
- treat the cryptographic module inventory as part of release evidence

### 1.2 TLS posture

The federal profile should support:

- TLS 1.2 with FIPS-appropriate cipher suites
- TLS 1.3 support in current deployments

Weak or deprecated protocol/cipher configurations must not be part of the federal profile.

## 2. FedRAMP compatibility posture

### 2.1 Rev. 5 alignment

The design should remain compatible with FedRAMP Rev. 5 / NIST SP 800-53 Rev. 5 style control expectations.

This means the repo and deployment posture should make it straightforward to demonstrate:

- access control
- audit and accountability
- configuration management
- identification and authentication
- system and communications protection
- supply-chain and software integrity controls
- continuous monitoring hooks

### 2.2 Cryptographic module policy

For federal deployment profiles, the safest default is:

- use validated cryptographic modules or update streams where available
- document whether every module is CMVP-validated or an update stream of a validated module
- make the cryptographic inventory visible to assessors and operators

For higher-assurance federal profiles, this should be treated as required rather than optional.

## 3. Repo-level design consequences

### 3.1 Runtime images

Runtime and service images should have a documented cryptographic baseline:

- base image lineage
- crypto library selection
- whether the image is intended for a federal / FIPS build lane
- SBOM evidence for the crypto-relevant packages

### 3.2 Build and provenance

The CI/CD profile should preserve evidence needed for federal review:

- pinned digests
- SBOM
- signature
- provenance attestation
- cryptographic module inventory or reference

### 3.3 Configuration

The federal profile should be explicit rather than implicit.

Recommended future knobs:

- `federal_profile: true|false`
- `minimum_trust_tier`
- `require_fips_validated_crypto`
- `strict_egress`

## 4. Network and boundary posture

A FedRAMP-compatible deployment should assume:

- TLS termination uses the federal crypto profile
- outbound traffic is constrained and auditable
- session runtime namespaces are default-deny by policy
- audit and telemetry export paths are explicit and reviewable

## 5. Assessment-ready evidence

The design should make it easy to produce evidence for:

- what crypto modules are used
- how images were built
- what policy gates were enforced
- which node / trust tier ran a session
- what events were audited for each session lifecycle step

## 6. Non-goals

This document does not claim:

- an Authority to Operate
- a completed SSP/control inheritance package
- that all deployment environments automatically satisfy federal requirements

It only defines the product-side and deployment-side design choices that make those outcomes achievable.

## 7. Immediate follow-on work

- add a federal/fips build lane profile in deployment docs
- document approved cryptographic module choices for gateway/runtime images
- add platform-level deployment guidance for federal tenants in `prophet-platform`
