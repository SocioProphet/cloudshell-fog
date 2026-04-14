# Runtime Isolation v0 — cloudshell-fog

## 0. Purpose

Define the current and target runtime isolation posture for session execution.

## 1. Current implementation baseline

Reviewed implementation includes:
- per-session namespace intent
- hardened pod security context in k8s connector
- non-root execution
- `allowPrivilegeEscalation: false`
- dropped Linux capabilities
- RuntimeDefault seccomp

This is a strong MVP baseline for Kubernetes-backed shell sessions.

## 2. Isolation goals

The runtime boundary should prevent:
- one subject's shell from interacting with another subject's session
- unnecessary privilege inside session runtime pods
- unrestricted east/west movement in cluster
- casual host-level breakout paths

## 3. Isolation layers

### Layer A — session identity binding
- session ID → namespace / runtime ref
- subject → session ownership checks

### Layer B — namespace separation
- each session gets its own namespace (target posture)
- namespace receives standard policy pack

### Layer C — pod security posture
- non-root
- no privilege escalation
- drop all capabilities by default
- RuntimeDefault seccomp
- read-only root filesystem where feasible

### Layer D — network isolation
- default deny
- explicitly allow DNS / HTTPS egress only unless policy expands it

### Layer E — image trust
- digest-based runtime image selection for production
- signed image verification

## 4. Current trade-offs

### Containers first
Pros:
- operationally simple
- fast startup
- fits Kubernetes connector already implemented

Cons:
- weaker boundary than microVM-based isolation
- depends heavily on cluster hardening and admission controls

### Future stronger isolation
Potential future options:
- Kata / VM-backed containers
- microVM-backed runtime connector
- additional per-tenant node isolation

These are not required for the current MVP but should remain possible within the connector abstraction.

## 5. Recommendation

For the current repository stage:
- keep Kubernetes hardened-container posture as the active default
- complete per-session namespace bootstrap
- complete policy enforcement and signed-image trust
- evaluate stronger runtime classes only after the current seam fixes are merged

## 6. Open questions

- do we require read-only root FS for the session runtime image?
- do we want dedicated runtime classes by profile?
- when does the project warrant microVM-backed execution?
