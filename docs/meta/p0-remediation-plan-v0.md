# P0 Remediation Patch Plan v0 — cloudshell-fog

Purpose: translate the engineering backlog into the first concrete implementation wave.

This plan deliberately focuses on changes that reduce the highest trust and operability risk with the smallest architectural blast radius.

## Patch Set A — Make connector mode real

### Objective
Turn connector selection from implied behavior into explicit behavior.

### Files likely touched
- `cmd/gateway/main.go`
- possibly a new helper file under `internal/connector/` or `cmd/gateway/`
- deployment docs if env vars change

### Proposed behavior
Support explicit connector mode selection:
- `CONNECTOR_MODE=stub`
- `CONNECTOR_MODE=k8s`

Optional later:
- `CONNECTOR_MODE=auto`

### Required startup behavior
- if mode is `stub`: use `connector.NewStubConnector()`
- if mode is `k8s`: construct a Kubernetes connector from an explicit config path or in-cluster config
- if mode is unknown: fail fast
- if mode is `k8s` but kube config is invalid/unavailable: fail fast

### Why this is first
The current gateway still behaves like a stub-first demo even though the repo now presents a real k8s path.

### Acceptance checks
1. Start with `CONNECTOR_MODE=stub` and confirm stub path.
2. Start with `CONNECTOR_MODE=k8s` and missing config; startup must error.
3. Start with `CONNECTOR_MODE=k8s` and valid config; allocation path must use `KubernetesConnector`.

## Patch Set B — Resolve k8s credential model explicitly

### Objective
Document and implement how the gateway authenticates to Kubernetes when using the k8s connector.

### Problem
The current deploy hardening may block in-cluster auth because ServiceAccount token automount is disabled.

### Decision candidates
A. In-cluster config via ServiceAccount token
- requires enabling token mount for the gateway pod or projecting a token explicitly

B. Explicit kubeconfig mount
- mount a kubeconfig secret/config into the gateway pod
- avoids implicit SA token usage, but adds secret distribution burden

### Recommended v0 choice
Prefer **in-cluster config** for cluster-local deployments, but make it explicit in docs/manifests.

### Files likely touched
- `cmd/gateway/main.go`
- `deploy/k8s/rbac.yaml`
- `deploy/k8s/deployment.yaml`
- `deploy/README.md`

### Acceptance checks
- k8s mode starts inside cluster without ad-hoc manual kubeconfig hacks
- least-privilege RBAC still holds

## Patch Set C — Canonicalize runtime image reference

### Objective
Remove ambiguity and probable typos in runtime image naming.

### Files likely touched
- `internal/api/handler.go`
- docs that mention runtime image naming

### Required behavior
- choose one canonical runtime image reference pattern
- make it configurable via env or request override with policy controls
- avoid typo-prone hardcoded fallback strings

### Acceptance checks
- one canonical image naming convention appears across code and docs
- no stale / suspicious fallback strings remain

## Patch Set D — Separate dev examples from production-safe image policy

### Objective
Stop mixing `:latest` examples with production policy claims.

### Files likely touched
- `deploy/k8s/deployment.yaml`
- `deploy/README.md`
- `deploy/tekton/task-build.yaml`

### Required behavior
- deployment manifest should either:
  - use digests in production examples
  - or clearly mark mutable tags as dev/demo only
- Tekton task tool images should be pinned to stable versions/digests where practical

### Acceptance checks
- no production-looking manifest contradicts immutable-image policy
- pipeline toolchain is less mutable

## Patch Set E — Reconcile policy and Tekton under GitOps scope

### Objective
Make GitOps coverage match the security claims.

### Problem
Current Argo Application points only at `deploy/k8s`.
Policy under `policy/` and Tekton under `deploy/tekton/` are outside that scope.

### Options
A. expand one Application to a higher-level path
B. app-of-apps / multiple Applications

### Recommended v0 choice
Document and implement **multiple Applications or app-of-apps**, so:
- gateway deploy surface
- policy surface
- CI/CD surface
can evolve without one giant path assumption.

### Acceptance checks
- Argo-managed scope is explicit
- policy artifacts are no longer “present in repo but not reconciled"

## Suggested implementation order
1. Patch Set A
2. Patch Set B
3. Patch Set C
4. Patch Set D
5. Patch Set E

## Definition of done for wave 1
Wave 1 is complete when:
- k8s mode is actually selectable and functional
- credential model is explicit and deployable
- runtime image naming is canonicalized
- deploy and pipeline artifacts no longer undercut immutable-image policy
- GitOps scope covers the security-relevant surfaces
