# Spec ↔ Code ↔ Ops Conformance Matrix v1 — cloudshell-fog

Last reviewed against canonical upstream: 2026-04-11  
Supersedes: `docs/meta/spec-code-conformance-v0.md` where they conflict.

## 1. High-level verdict

`cloudshell-fog` is a real MVP implementation with real docs and real policy artifacts.
The strongest remaining problems are **implementation/ops fidelity**, not conceptual incoherence.

In plain language:
- the architecture is coherent
- the Go service is real
- Kyverno policy exists
- Argo/Tekton artifacts exist
- but several critical paths are still partly ceremonial or inconsistent

## 2. Sharpened conformance findings

### 2.1 What is genuinely solid

#### A. Real Go implementation exists
The canonical repo contains actual code for:
- gateway entrypoint
- OIDC validation and session-token minting
- session API
- PTY WebSocket bridge
- in-memory session store
- placement engine
- YAML policy engine
- stub + k8s connector backends
- OTEL setup

#### B. Real policy artifacts exist
The repo contains actual Kyverno ClusterPolicy resources for:
- runtime hardening baseline
- require-image-digest
- verify signed images

#### C. Real deploy artifacts exist
The repo contains:
- Kubernetes manifests
- Argo CD Application
- Tekton pipeline + task
- Tekton Chains config

## 3. Confirmed drift / weaknesses (more precise than v0)

### 3.1 Connector selection is still not wired through the gateway
The reviewed `cmd/gateway/main.go` unconditionally instantiates the stub connector and merely logs that `USE_K8S=1` should be used.

Implication:
- the existence of `internal/connector/k8s.go` does not currently mean the deployed gateway will actually use it.

Priority: P0

### 3.2 Deployment manifest still advertises mutable image tags
`deploy/k8s/deployment.yaml` still uses:
- `ghcr.io/socioprophet/cloudshell-fog/gateway:latest`

This conflicts with the production policy posture that requires immutable digests.

Priority: P0

### 3.3 Tekton task also uses mutable tool image tags
`deploy/tekton/task-build.yaml` uses:
- `gcr.io/kaniko-project/executor:latest`
- `anchore/syft:latest`

This weakens the supply-chain posture of the build pipeline itself.

Priority: P1

### 3.4 Argo CD currently syncs only `deploy/k8s`, not policy or Tekton
`deploy/argocd/application.yaml` points Argo to:
- `path: deploy/k8s`

Implication:
- Kyverno policy under `policy/` is not automatically reconciled by this Application
- Tekton resources under `deploy/tekton/` are not automatically reconciled either

So the repo is “GitOps-ready” only for the gateway deployment surface, not yet for the full security/control surface.

Priority: P0

### 3.5 NetworkPolicy manifest is template-like, not truly per-session-applied
`deploy/k8s/networkpolicy.yaml` states it should apply to every `cloudshell-<sessionID>` namespace, but the actual manifest namespace shown is `cloudshell-system`.

Implication:
- as committed, this is more of a template/example than a complete implementation of per-session network isolation.

Priority: P0

### 3.6 In-cluster auth for k8s connector may be blocked by current deployment security settings
`deploy/k8s/rbac.yaml` sets the gateway ServiceAccount with:
- `automountServiceAccountToken: false`

and `deploy/k8s/deployment.yaml` also uses a hardened pod spec.

Implication:
- if the gateway is intended to use in-cluster Kubernetes credentials via the ServiceAccount token, current settings may prevent that.
- this is compatible with the current stub-only wiring, but may block the eventual real k8s path unless credentials are provided another way.

Priority: P0/P1 depending on intended deployment model.

### 3.7 API response remains flatter than richer placement semantics
`internal/api/handler.go` returns `placement` as a region string.
The richer interface/placement docs imply a more expressive placement object (tier / reasons / placement_id semantics).

Priority: P1

### 3.8 Session lifecycle remains thinner than fog-grade degraded semantics
Implementation state surface appears to be limited to simpler statuses (`pending`, `running`, `terminated`, `expired`).
A dedicated state-machine doc was not found during direct fetch review.

Priority: P1

### 3.9 Signed-image verification policy is not production-complete yet
`policy/kyverno/verify-signed-images.yaml` still contains a placeholder public key block.

Implication:
- the policy shape exists, but production trust material is not fully wired.

Priority: P0 before any serious production claim.

### 3.10 API/default image reference should be verified
The earlier implementation review showed a suspicious default runtime image ref (`ghcr.io/socioproph/...`) that may be a typo or stale string.
This should be verified against current mainline source before code changes are planned.

Priority: P0 verification task.

## 4. Missing / not directly confirmed in this review

These were not found by direct fetch on canonical `main` during review and remain good candidates if still absent:
- `docs/spec/state-machine-v0.md`
- `docs/spec/openapi/control-api.v1.yaml`
- `docs/spec/jsonschema/ws-pty.v1.json`
- `docs/spec/jsonschema/audit-event.v1.json`
- ADR set for key decisions

## 5. Priority fix backlog

### P0 — correctness / trust / deployability
1. Wire actual connector selection in gateway (`stub` vs `k8s`).
2. Verify/fix default runtime image reference.
3. Replace mutable image tags in deployment manifests with digest strategy or clearly mark as non-production examples.
4. Decide how k8s connector authenticates in-cluster and align ServiceAccount token mount / kubeconfig approach.
5. Reconcile policy deployment path under GitOps (Argo currently ignores `policy/`).
6. Replace placeholder cosign public key / complete keyless verification path.

### P1 — maturity / conformance
7. Add formal state-machine doc.
8. Add machine-readable interface contracts (OpenAPI / JSON Schema).
9. Align API response schema with richer placement model or intentionally narrow docs.
10. Turn per-session NetworkPolicy from template semantics into actual provisioned behavior.
11. Pin Tekton task tool images by digest/version.

## 6. Working rule for future agents

Do not stop at README/spec inspection.
Always check:
- gateway entrypoint wiring
- deployment manifests
- GitOps application scope
- policy manifests
- build pipeline image references

This repo’s main risks are now in the seams between docs, code, and ops.
