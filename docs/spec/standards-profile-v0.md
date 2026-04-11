# Standards Profile v0 — cloudshell-fog

This document states the standards and protocol boundaries that cloudshell-fog treats as normative in its current design.

## 0. Design posture

We keep the browser-facing and operator-facing surfaces simple and standards-native first. We do **not** force a custom wire protocol where ordinary HTTPS, OIDC, WebSocket, OCI, and OpenTelemetry already solve the problem cleanly.

We reserve richer inter-agent protocol work for the internal and future integration layers.

## 1. Normative standards (current baseline)

| Concern | Default standard / mechanism | Status |
|---|---|---|
| End-user identity | OpenID Connect (OIDC) over OAuth 2.0 (Authorization Code + PKCE) | Normative |
| API transport | HTTPS + JSON | Normative |
| Interactive terminal transport | WebSocket Secure (WSS) | Normative |
| Runtime packaging | OCI image + OCI registry | Normative |
| Telemetry | OpenTelemetry traces, metrics, structured logs | Normative |
| Supply chain | SBOM + signature + provenance attestation | Normative |
| GitOps | Argo CD Autopilot-compatible reconciliation model | Normative |
| CI/CD | Tekton Pipelines + Tekton Chains | Normative |
| Kubernetes runtime policy | NetworkPolicy + admission controls | Normative |

## 2. Supply-chain profile

The minimum production profile is:

- OCI image referenced by pinned digest only.
- SBOM present for that digest.
- Signature present for that digest.
- Provenance attestation present for that digest.
- Admission enforcement validates all of the above before scheduling the runtime.

Recommended encodings:

- SBOM: SPDX by default; CycloneDX acceptable where needed.
- Provenance: in-toto-style statement produced through Tekton Chains.
- Signature verification: Sigstore/cosign.

## 3. Policy profile

We distinguish two policy layers:

### 3.1 In-gateway policy

The gateway owns immediate admission checks tied to the interactive shell product itself:

- profile selection
- TTL bounds
- per-subject session count
- group/role gating
- fog/cloud placement constraints

This remains intentionally simple and product-local.

### 3.2 Cluster / platform policy

The cluster enforces runtime and supply-chain invariants:

- pinned digests only
- signature verification
- provenance verification
- SBOM presence requirement
- namespace / pod security constraints
- NetworkPolicy defaults

Default recommendation: use **Kyverno** for Kubernetes-native admission policy. Keep **OPA/Rego** available for broader organisational and cross-service policy where richer composition is needed.

## 4. Eventing and protocol boundaries

### 4.1 Current external protocol boundary

The browser and operator interfaces stay on:

- HTTPS/JSON for control plane
- WSS for PTY attach

This is the stable, lowest-friction surface.

### 4.2 Current internal event profile

OpenTelemetry is the current observability envelope.

Where asynchronous control-plane events are introduced later, **CloudEvents** is the preferred envelope for interoperable event publication.

### 4.3 Future inter-agent profile

`cloudshell-fog` may later expose selected internal control operations through **TriRPC** or other agent-to-agent transports, but only after the HTTP/WSS product surface is stable.

Current rule:

- Browser/user surface: HTTP + WSS
- Internal observability/event surface: OpenTelemetry, optional CloudEvents
- Future agent mesh surface: candidate for TriRPC alignment

## 5. Cryptographic and identity posture

- Short-lived access tokens only.
- Session attach tokens are session-bound and cannot mint or mutate sessions.
- No long-lived browser secrets.
- Runtime trust tiers may incorporate node attestation and secure-boot claims where available.

## 6. Resulting architectural stance

cloudshell-fog is therefore:

- browser-native, not protocol-exotic
- GitOps-native, not click-ops-driven
- supply-chain-verifiable, not opaque
- fog-aware, but standards-grounded
- compatible with broader SocioProphet protocol work without prematurely binding the product to it
