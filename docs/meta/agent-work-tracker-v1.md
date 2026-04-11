# Agent Work Tracker v1 — cloudshell-fog

Status: active planning / anti-duplication tracker  
Supersedes: `docs/meta/agent-work-tracker-v0.md` where they conflict  
Last reviewed against upstream: 2026-04-11

## Why v1 exists

Upstream moved after the v0 tracker assumptions. Several artifacts previously listed as missing are now present on `main`. This file records the current state so future agents do not duplicate already-landed work.

## Current upstream baseline (confirmed present on canonical `SocioProphet/cloudshell-fog`)

### Core docs already present
- `README.md`
- `docs/spec/minimum-spec-v0.md`
- `docs/spec/interfaces-v1.md`
- `docs/architecture.md`

### Additional docs now present
- `docs/spec/fog-placement-v0.md`
- `docs/spec/policy-baseline-v0.md`
- `docs/spec/cicd-contract-v0.md`
- standards / compliance / trust-tier / integration profile docs

### Policy work now present
- baseline policy README(s)
- signed-image verification baseline
- require-image-digest Kyverno baseline
- runtime baseline Kyverno policy
- Kyverno baseline kustomization

### Code / implementation baseline already landed
The repository history shows a Go backend implementation commit covering:
- session API
- auth
- PTY bridge
- placement
- policy
- connector

## What this means

The repository has NOT changed direction radically.
It HAS advanced materially in documentation and policy hardening.

So the right move is:
- do not rewrite the repo conceptually from scratch
- do not duplicate docs that already exist
- enrich only the areas that remain under-specified or need implementation review

## Remaining high-value gaps (current)

### Still likely needed / should be verified
1. `docs/spec/state-machine-v0.md`
2. `docs/spec/runtime-isolation-v0.md`
3. `docs/spec/security-threat-model-v0.md`
4. machine-readable artifacts if still absent:
   - `docs/spec/openapi/control-api.v1.yaml`
   - `docs/spec/jsonschema/ws-pty.v1.json`
   - `docs/spec/jsonschema/audit-event.v1.json`
5. ADR set for key decisions
6. code-vs-spec conformance review against actual Go implementation

## Required de-duplication protocol for future agents

Before adding work:
1. Read `README.md`, `docs/architecture.md`, and `docs/spec/interfaces-v1.md`.
2. Read both `agent-work-tracker-v0.md` and this v1 file; prefer v1.
3. Search recent commits before claiming a spec artifact is missing.
4. Prefer surgical enrichment over complete rewrites.

## Recommended next workstreams

### Workstream A — Spec conformance review
Compare current Go implementation to:
- `interfaces-v1.md`
- `fog-placement-v0.md`
- `policy-baseline-v0.md`

### Workstream B — Remaining spec depth
Add only if absent:
- state machine doc
- runtime isolation doc
- threat model doc
- machine-readable schemas
- ADRs

### Workstream C — Ops verification
Verify that Tekton / Chains / Argo / Kyverno artifacts in-tree actually match the current written invariants.

## Source of truth
Canonical upstream:
- `SocioProphet/cloudshell-fog`

Mirror:
- `SociOS-Linux/cloudshell-fog`
