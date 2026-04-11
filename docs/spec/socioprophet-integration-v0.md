# SocioProphet Integration v0 — cloudshell-fog

This document explains how cloudshell-fog fits into the broader SocioProphet platform canon.

## 0. Product role

cloudshell-fog is the interactive execution surface for a governed, fog-aware operator shell.

It is not the entire platform. It is the browser-accessible terminal and session gateway layer that allows a human operator (and later, selected automated agents) to enter an execution environment with:

- strong identity
- bounded runtime policy
- fog-aware placement
- auditable session lifecycle
- verifiable build and deployment provenance

## 1. Relationship to the broader stack

### 1.1 SocioProphet

Within SocioProphet, cloudshell-fog occupies the operator/runtime edge of the platform.

Its immediate concerns are:

- interactive session creation
- execution placement
- session auditability
- runtime isolation
- standards-native browser access

It should remain small enough to be composable into a larger governed platform.

### 1.2 SociOS / SourceOS

Where SourceOS and related Linux work define the host and operating substrate, cloudshell-fog defines the network-accessible shell/session control surface.

Relationship:

- SourceOS / Linux substrate: host trust, node posture, runtime base images, local execution semantics
- cloudshell-fog: remote/browser shell access, fog placement, session policy, attach semantics

This means node attestation and trust posture from the substrate can later become scheduling and admission inputs for cloudshell-fog.

### 1.3 Socios commons / automation lane

In a Socios-style community or automation environment, cloudshell-fog can become the governed shell interface for:

- operators
- maintainers
- approved automation subjects
- support and debugging workflows

But the user-facing and browser-facing contract remains standards-native; we do not collapse that into an agent mesh prematurely.

## 2. Protocol layering

### 2.1 Current product boundary

The current user-facing product boundary stays on:

- HTTPS/JSON for session control
- WSS for PTY attach
- OIDC/OAuth for identity

This is the right boundary for browser-native adoption and operational clarity.

### 2.2 Future internal / agent boundary

As the broader SocioProphet protocol work matures, cloudshell-fog may expose selected control-plane or event-plane interactions through:

- CloudEvents for interoperable asynchronous events
- TriRPC for richer internal or agent-to-agent coordination

Current rule:

- do not replace the product surface with TriRPC
- reserve TriRPC for future internal / mesh-level integration where it adds clear value

## 3. Governance alignment

cloudshell-fog already aligns with core governance expectations:

- identity-bound sessions
- policy-bound runtime allocation
- audit emission for operator actions
- supply-chain-verifiable images and deployments

To align more deeply with the broader platform, it should next absorb:

- stronger node trust evidence in placement
- formal exception handling for policy bypasses
- broader organisation-level policy hooks beyond local gateway checks

## 4. Stack alignment

### 4.1 Deployment and promotion

- Argo CD Autopilot-compatible GitOps model
- Tekton build pipelines
- Tekton Chains provenance and signing
- digest-explicit promotion

### 4.2 Policy

- gateway-local policy for profile / TTL / session-count bounds
- Kyverno-first cluster admission policy for digest/signature/provenance/runtime isolation
- OPA/Rego retained for wider cross-service and organisation-level policy logic

### 4.3 Observability

- OpenTelemetry as the normative telemetry envelope
- structured audit events for session, placement, and policy outcomes
- future CloudEvents profile where asynchronous external event publication is needed

## 5. Immediate integration backlog

1. codify admission policies under a `policy/` tree
2. bind fog trust-tier claims to actual node evidence and scheduling filters
3. connect deployment docs to the new standards / policy / CI-CD profile
4. define which internal events, if any, should later gain a CloudEvents or TriRPC envelope

## 6. Result

cloudshell-fog should be understood as:

- a governed shell edge for the SocioProphet platform
- a browser-native operator surface
- a fog-aware session scheduler
- a standards-first component that can later plug into richer protocol and governance layers without losing operational simplicity
