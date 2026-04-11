# Agent Work Tracker v0 — cloudshell-fog

Status: active planning / architecture capture  
Scope: canonical backlog and de-duplication tracker for cloudshell-fog  
Policy: clean-room only; no proprietary diagrams, screenshots, OCR dumps, or copied vendor artifacts

## Why this file exists

This repository currently has no open GitHub issues visible to the connector session used for this review. The current connector session also does not expose issue creation. To avoid duplicate effort and preserve agent context, this file acts as a repository-native work tracker until issue creation is available or a human creates equivalent issues.

All future agents SHOULD read this file together with the current canonical docs before proposing major changes:

- `README.md`
- `docs/spec/minimum-spec-v0.md`
- `docs/spec/interfaces-v1.md`
- `docs/architecture.md`

## Repository review summary (what already exists)

The canonical repository already contains a strong baseline:

1. Product framing in `README.md`
   - Go-based fog-optimised cloud shell gateway
   - OIDC authentication
   - fog-first placement with trusted cloud fallback
   - YAML policy engine
   - in-memory session store
   - runtime connector abstraction (`k8s` / `stub`)
   - structured audit events
   - OpenTelemetry
   - Tekton Chains / cosign / SBOM / Argo CD language

2. Initial clean-room minimum specification in `docs/spec/minimum-spec-v0.md`
   - planes
   - threat model seed
   - Wetty integration boundary
   - high-level Tekton / Argo posture

3. Standards-aligned interface specification in `docs/spec/interfaces-v1.md`
   - OIDC / OAuth2
   - JWT attach tokens
   - WSS PTY frame schema
   - runtime connector contract
   - production supply-chain invariants
   - fog placement and fallback semantics

4. Package-level architecture in `docs/architecture.md`
   - `cmd/gateway`
   - `internal/auth`
   - `internal/api`
   - `internal/pty`
   - `internal/session`
   - `internal/placement`
   - `internal/policy`
   - `internal/connector`
   - `internal/audit`
   - `internal/otel`

## Re-established goals and direction

### Primary product goal

Provide a browser-accessible terminal that executes inside a secure, auditable runtime placed as close as possible to the user and/or data (fog / edge first, trusted cloud fallback second).

### Distinguishing design goals

- clean-room design and implementation
- self-hostable and open-source
- standards-aligned identity, transport, observability, and supply chain
- explicit placement logic and truthful degraded behavior under partition/failure
- runtime abstraction that can evolve from hardened containers to stronger isolation tiers
- GitOps and supply-chain verification as first-class design constraints, not bolt-ons

## What has been accomplished

### Architecture / design
- Product intent is clear and consistent across README + spec + architecture docs.
- Control-plane responsibilities are defined: session lifecycle, placement, policy, attach token minting.
- Data-plane responsibilities are defined: runtime allocation + PTY attach.
- Audit + OTEL responsibilities are defined.
- Fog semantics are present: prefer fog, fall back to trusted cloud.

### Interfaces
- HTTP session lifecycle endpoints are already documented.
- WSS PTY attach schema is already documented.
- Runtime connector abstraction exists conceptually and in architecture docs.

### Security and supply chain
- Production posture already references digest pinning, signatures, provenance, and SBOM.
- Session-bound short-lived attach tokens are already part of the design.

### Delivery posture
- Repo already speaks GitOps and Tekton/Argo language.
- Canonical repo exists in `SocioProphet`; mirror exists in `SociOS-Linux`.

## What remains incomplete / under-specified

The repo is strong, but still needs deeper formalisation in several areas.

### 1. State truthfulness / lifecycle rigor
Missing artifact:
- `docs/spec/state-machine-v0.md`

Needed content:
- explicit session states
- transition rules
- idempotency guarantees
- race handling (`terminate` vs `allocate`, multi-attach policy)
- TTL and idle-timeout semantics

### 2. Fog placement depth
Missing artifact:
- `docs/spec/fog-placement-v0.md`

Needed content:
- scoring model inputs and weights
- reason-code catalog
- tie-breaking rules
- failure and resume semantics
- trust-tier interpretation

### 3. Policy baseline and admission enforcement
Missing artifact:
- `docs/spec/policy-baseline-v0.md`

Needed content:
- production invariants
- admission controller baseline choice (Kyverno vs Gatekeeper / OPA)
- supply-chain verification rule inventory
- runtime guardrail inventory (namespace, network, privilege, resource limits)

### 4. Runtime isolation doctrine
Missing artifact:
- `docs/spec/runtime-isolation-v0.md`

Needed content:
- containers vs microVM trade-offs
- tiered isolation strategy
- profile model (`cpu`, `mem`, `storage`, `isolation_class`, `network_class`)

### 5. Threat model and privacy posture
Missing artifact:
- `docs/spec/security-threat-model-v0.md`

Needed content:
- assets
- adversaries
- trust assumptions
- threat → mitigation mapping
- privacy defaults (no keystroke capture by default)

### 6. Machine-readable interface artifacts
Missing artifacts:
- `docs/spec/openapi/control-api.v1.yaml`
- `docs/spec/jsonschema/ws-pty.v1.json`
- `docs/spec/jsonschema/audit-event.v1.json`

Needed content:
- enforceable API schema
- enforceable WSS frame schema
- enforceable audit-event envelope schema

### 7. Decision memory
Missing artifacts:
- ADRs capturing key trade-offs

Suggested ADRs:
- attach protocol uses WSS publicly, SSH only as backend option
- containers-first MVP with stronger isolation path later
- policy invariants first, engine selection second
- single repo per org initially, split later only if lifecycle/access control justifies it

## Recommended work breakdown (non-duplicative)

### Workstream A — Spec enrichment
Add:
- state machine doc
- fog placement doc
- policy baseline doc
- runtime isolation doc
- threat model doc
- OpenAPI + JSON Schemas
- ADRs

### Workstream B — Reference implementation hardening
Review current Go implementation against spec and check for gaps in:
- token scoping
- PTY attach validation
- session ownership checks
- race handling
- timeout cleanup
- OTEL attribute naming and coverage

### Workstream C — Ops hardening
Add / verify:
- Tekton pipeline definitions
- Chains attestation path
- Argo CD Autopilot structure
- admission policies
- namespace/network policy manifests

## De-duplication protocol for future agents

Before starting new work, an agent SHOULD:

1. Read this tracker.
2. Read the current `README.md`, `docs/spec/interfaces-v1.md`, and `docs/architecture.md`.
3. Search issues / PRs / branches for overlapping work.
4. Update this file or replace it with formal GitHub issues once issue creation is available.
5. Avoid creating parallel “complete rewrite” specs unless there is an explicit decision to supersede current docs.

## Proposed next deliverables

1. `docs/spec/state-machine-v0.md`
2. `docs/spec/fog-placement-v0.md`
3. `docs/spec/policy-baseline-v0.md`
4. `docs/spec/runtime-isolation-v0.md`
5. `docs/spec/security-threat-model-v0.md`
6. `docs/spec/openapi/control-api.v1.yaml`
7. `docs/spec/jsonschema/ws-pty.v1.json`
8. `docs/spec/jsonschema/audit-event.v1.json`
9. `docs/adr/0001-attach-protocol-wss.md`
10. `docs/adr/0002-runtime-isolation-tiering.md`
11. `docs/adr/0003-policy-engine-baseline.md`
12. `docs/adr/0004-repo-splitting.md`

## Notes on source of truth

Canonical upstream for planning and implementation remains:
- `SocioProphet/cloudshell-fog`

Mirror:
- `SociOS-Linux/cloudshell-fog`

Until issue creation is available in the active connector session, this file is the working backlog / anti-duplication registry.
