# Spec ↔ Code ↔ Ops Conformance Matrix v0 — cloudshell-fog

Last reviewed against canonical upstream: 2026-04-11

This document records what currently matches, what drifts, and what still appears absent.
It is intentionally narrower than the broader agent work trackers.

## 1. High-level verdict

The repository is **directionally consistent**:
- product README
- interface docs
- architecture docs
- Go implementation
- deployment manifests
- Kyverno policy baseline

The main problem is **not conceptual drift**.
The main problem is **maturity drift**: docs have evolved faster than some parts of implementation and ops.

## 2. Conformance table

| Area | Spec / Docs state | Code / Ops state | Current verdict |
|---|---|---|---|
| Product identity | Fog-first cloud shell gateway with trusted cloud fallback | README and architecture reflect this; placement code uses fog-first/cloud-fallback model | aligned |
| Control API | `POST/GET/DELETE /v1/sessions` documented | implemented in `internal/api/handler.go` | aligned with some response-shape drift |
| PTY attach | WSS attach with JSON frames | implemented in `internal/pty/pty.go` | aligned |
| OIDC + attach token split | documented and architected | implemented in `internal/auth/auth.go` and gateway wiring | aligned |
| Runtime connector abstraction | documented | implemented via `connector.Connector`, stub, and k8s backends | aligned |
| Placement semantics | newer docs describe more formal filters/tie-break/fallback semantics | code implements a simpler MVP scorer | partial alignment |
| Session state machine | lifecycle described across docs, but no dedicated formal state-machine doc found during review | code uses simpler states (`pending`, `running`, `terminated`, `expired`) | under-specified / drift |
| Audit events | documented | implemented in `internal/audit` and emitted by API / PTY / sweeper | aligned |
| OpenTelemetry | documented | gateway sets up stdout OTEL providers | aligned for MVP |
| Kubernetes deployment | documented in `deploy/README.md` | `deploy/k8s/*.yaml` exist | aligned with production-hardening gaps |
| GitOps / Argo CD | documented | `deploy/argocd/application.yaml` exists | aligned |
| Tekton / Chains | documented | `deploy/tekton/pipeline.yaml` exists | partially aligned (pipeline exists; enforcement depth still needs verification) |
| Supply-chain admission | policy baseline docs present | Kyverno policies for digest + signed images + runtime baseline exist | partially aligned |

## 3. Confirmed strengths

### 3.1 Real implementation exists
Canonical repo contains real Go implementation surface:
- gateway entrypoint
- auth
- API
- PTY bridge
- placement
- policy
- session store
- connector abstraction
- OTEL setup

### 3.2 Real policy baseline exists
Canonical repo contains actual Kyverno ClusterPolicy resources for:
- runtime hardening baseline
- require-image-digest
- verify signed images

### 3.3 Real deploy surface exists
Canonical repo contains:
- deploy/k8s manifests
- Argo CD application
- Tekton pipeline docs/manifests

## 4. Confirmed drift / gaps

### 4.1 Connector selection drift
`cmd/gateway/main.go` currently always wires the stub connector and logs that `USE_K8S=1` should enable k8s, but the reviewed entrypoint does not branch on that env var.

### 4.2 API response-shape drift
The written interface shape is richer than the current `CreateSession` response.
Current implementation returns `placement` as a region string rather than a structured placement object with tier and reasons.

### 4.3 Session lifecycle depth drift
No dedicated `docs/spec/state-machine-v0.md` was found during review.
Implementation state surface appears thinner than the broader lifecycle semantics discussed in docs/conversation.

### 4.4 Placement maturity drift
`docs/spec/fog-placement-v0.md` defines a more formal placement model than the current code scorer.
Current code is a pragmatic MVP, not the full documented model.

### 4.5 Image reference bug / typo to verify and fix
The implementation commit shows default image ref as `ghcr.io/socioproph/cloudshell-runtime:latest` in API handler code, which appears suspicious and should be verified against intended registry naming.

### 4.6 Missing machine-readable contracts (spot-check)
During this review, these paths were not found when directly fetched from canonical main:
- `docs/spec/state-machine-v0.md`
- `docs/spec/openapi/control-api.v1.yaml`

These should be added if still absent.

## 5. Priority backlog (implementation-facing)

### Priority 0 — correctness / trust
1. Fix connector selection so k8s backend can actually be enabled.
2. Verify and correct default runtime image reference.
3. Align API response schema with documented placement richness, or relax docs intentionally.

### Priority 1 — state rigor
4. Add formal state-machine doc.
5. Encode `disconnected` / resume semantics explicitly if desired, or narrow docs to current behavior.

### Priority 2 — contract enforcement
6. Add machine-readable OpenAPI for control API.
7. Add JSON Schema for WSS PTY frames.
8. Add JSON Schema for audit events.

### Priority 3 — ops hardening
9. Verify Tekton + Chains + Kyverno path end-to-end against actual production invariants.
10. Ensure deployment manifests stop advertising `latest` in places where policy requires digest pinning for production.

## 6. Working rule for future agents

Do not claim the repo is “just docs.”
Do not claim the repo is “done.”
The truth is in the middle:
- a real MVP implementation exists
- policy and docs have advanced quickly
- the next best work is conformance hardening, not conceptual restart
