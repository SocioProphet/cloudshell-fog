# Minimum Spec v0 — Fog-Optimized Cloud Shell (Clean-Room Notes)

## 0) Goal
Browser-accessible terminal for executing commands inside a secure, auditable runtime.
Placement must support **fog**: run close to the user/data when possible, degrade gracefully when not.

## 1) Planes
### 1.1 UI / Edge plane
- Web terminal UI (Wetty-class) over websocket.
- Console embedding surface: launch, resume, terminate session.
- Optional: “open port preview” (HTTP) separate from PTY channel.

### 1.2 Control plane
- Session lifecycle: create/resume/terminate; TTL; idle timeout.
- Placement: choose runtime location (cloud region / fog node / local cluster) based on policy + capacity + latency.
- Policy/quotas: max sessions, CPU/RAM profiles, storage quota, egress rules.
- Identity binding: OIDC token → subject → session; short-lived session credentials.

### 1.3 Data plane
- Runtime sandbox per user/session: container or microVM.
- Runtime connector: create/attach/exec/resize PTY; file ingress/egress (optional).
- Network: per-session network policy; optional service mesh sidecar if needed.

### 1.4 Audit / Observability plane
- Audit events: session created, attached, command streams (configurable), file transfer, port exposure.
- Metrics/tracing: gateway latency, attach failures, resource consumption, placement decisions.

## 2) Threat model (minimum)
- Prevent cross-session escape and privilege escalation.
- Prevent stolen browser tokens from becoming long-lived runtime access.
- Make execution provenance auditable (who/when/where/what runtime image).

## 3) Wetty integration contract (what Wetty does vs what we wrap)
Wetty provides:
- Browser terminal + websocket-to-PTY bridge.

We must provide:
- Auth/admission, session routing, runtime allocation/connector, quotas/policy, audit/obs.

## 4) 2026 CI/CD contract (high-level)
- GitOps: Argo CD Autopilot for cluster/app bootstrapping.
- Pipelines: Tekton for build/test; Tekton Chains for provenance/signing.
- Admission: enforce pinned digests + required attestations.
