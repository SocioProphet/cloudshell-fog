# Mesh Integration v0 — cloudshell-fog

This document defines the initial service-mesh stance for cloudshell-fog.

## 0. Goal

Use service-mesh primitives only where they close real gaps in identity, ingress/egress control, and multicluster traffic policy.

Do **not** impose mesh complexity on the browser-facing product surface or on short-lived session runtimes unless the operational benefit is clear.

## 1. Current default stance

### 1.1 Browser and operator surface

The browser-facing product contract remains:

- OIDC / OAuth for identity
- HTTPS / JSON for control-plane APIs
- WSS for PTY attach

This surface is not replaced by service-mesh protocols.

### 1.2 Session runtimes

Short-lived shell session pods are **not** mesh members by default.

Reasons:

- sidecar or ambient enrollment adds operational weight to highly ephemeral workloads
- shell sessions already have their own namespace and NetworkPolicy posture
- the current primary gaps are admission policy, supply-chain verification, and deployment wiring, not east-west service graph complexity

## 2. Where Istio is justified

Istio becomes useful at the stable platform boundary rather than inside every shell session.

### 2.1 Dedicated ingress / gateway

Use case:

- long-lived WebSocket PTY traffic
- explicit ownership of ingress policy for shell traffic
- TLS termination and route isolation from unrelated platform edges

### 2.2 Egress gateway (optional)

Use case:

- auditable outbound traffic for session runtimes
- stricter outbound control than raw NetworkPolicy alone
- environments that require central egress choke points

### 2.3 Service-to-service identity for stable control-plane services

Use case:

- gateway to supporting services
- stronger mTLS posture for non-ephemeral platform components
- multicluster service communication once cloudshell-fog is deployed as a platform service

## 3. Where Istio is not the first answer

Istio does not replace:

- Tekton / Tekton Chains
- Argo CD Autopilot-compatible GitOps
- Kyverno admission policy
- gateway-local profile / TTL / placement policy

Those remain first-order controls.

## 4. Admiral stance

Admiral is **deferred**.

It becomes interesting only when all of the following are true:

- we already operate a genuine multicluster Istio estate
- cloudshell-fog is one service among many in that mesh
- cross-cluster service discovery and traffic policy automation become operationally painful without an additional control layer

Until then, Admiral is additional complexity without solving the current highest-value problems.

## 5. Recommended phased model

### Phase 0 — current baseline

- no mandatory service mesh for session runtimes
- Kubernetes NetworkPolicy for isolation
- OIDC + HTTPS + WSS at the product boundary
- Kyverno for cluster admission and runtime hardening

### Phase 1 — platform-edge mesh

- optional dedicated Istio ingress / gateway for cloudshell-fog
- optional egress gateway for tightly controlled environments
- no default mesh enrollment for per-session runtimes

### Phase 2 — stable control-plane mesh

- mTLS and service identity for stable, long-lived cloudshell-fog services
- multicluster connectivity where prophet-platform deployment topology requires it

### Phase 3 — multicluster automation

- consider Admiral only if a real multicluster Istio control burden emerges

## 6. Relationship to prophet-platform

The natural home for mesh and deployment wiring is `prophet-platform`, not `cloudshell-fog`.

cloudshell-fog should define:

- the product boundary
- the runtime assumptions
- the mesh recommendation

prophet-platform should decide:

- whether an Istio ingress/gateway is instantiated
- whether an egress gateway is required in a given environment
- how cloudshell-fog coexists with the rest of the platform traffic topology

## 7. Resulting decision

Current decision:

- Istio: optional, targeted, and downstream/platform-owned
- Admiral: not part of the current minimum architecture
- Session runtimes: stay out of the mesh by default
