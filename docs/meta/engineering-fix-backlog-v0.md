# Engineering Fix Backlog v0 — cloudshell-fog

Purpose: convert the conformance findings into concrete, implementation-facing tasks.

## Priority 0 — make the real path real

### EF-0001 — Wire actual connector selection in gateway
Problem:
- `cmd/gateway/main.go` currently instantiates the stub connector unconditionally.
- `internal/connector/k8s.go` exists but is not selected by reviewed gateway wiring.

Required outcome:
- gateway chooses backend based on explicit config
- supported modes should include at least:
  - `stub`
  - `k8s`
- unsupported / misconfigured mode should fail fast at startup

Acceptance criteria:
- `CONNECTOR_MODE=stub` uses stub connector
- `CONNECTOR_MODE=k8s` uses k8s connector
- startup logs selected connector mode clearly
- missing k8s config causes startup error, not silent fallback

### EF-0002 — Verify and correct default runtime image reference
Problem:
- earlier implementation review showed a suspicious default image ref path

Required outcome:
- choose canonical runtime image name/path
- make it configurable
- avoid stale or typo-prone hardcoded defaults

Acceptance criteria:
- code, docs, and deploy manifests all reference the same canonical runtime image naming scheme

### EF-0003 — Stop advertising mutable tags as production-safe defaults
Problem:
- deployment and pipeline artifacts still use `:latest` in places while policy requires digests in production

Required outcome:
- clearly separate dev examples from production examples
- production paths should use digest language or explicit placeholders

Acceptance criteria:
- `deploy/k8s/deployment.yaml` no longer looks production-ready while using `:latest`
- Tekton task tool images are pinned to versions/digests where practical
- production hardening docs and manifests agree

### EF-0004 — Resolve k8s connector credential model
Problem:
- gateway deployment hardening and ServiceAccount token settings may block in-cluster client-go auth

Required outcome:
- document and implement one of:
  - in-cluster config using ServiceAccount token
  - external kubeconfig mount / secret / projected config
- remove ambiguity

Acceptance criteria:
- `CONNECTOR_MODE=k8s` startup path can actually construct a valid `*rest.Config`
- deploy docs explain credential model explicitly

## Priority 1 — align external contracts to internal truth

### EF-0005 — Align session create response with richer placement model
Problem:
- docs imply richer placement metadata than the current flatter API response

Required outcome:
- either return structured placement object (preferred)
- or intentionally narrow docs to current implementation

Acceptance criteria:
- `interfaces-v1.md` and API output match exactly

### EF-0006 — Add formal state-machine spec
Missing artifact:
- `docs/spec/state-machine-v0.md`

Required content:
- states
- transitions
- idempotency
- race handling
- disconnect / terminate semantics

### EF-0007 — Add machine-readable interface contracts
Missing artifacts:
- `docs/spec/openapi/control-api.v1.yaml`
- `docs/spec/jsonschema/ws-pty.v1.json`
- `docs/spec/jsonschema/audit-event.v1.json`

Acceptance criteria:
- schemas match current implementation or documented desired response if implementation is updated

## Priority 2 — make GitOps/security less ceremonial

### EF-0008 — Reconcile policy deployment path under Argo CD
Problem:
- current Argo Application points only at `deploy/k8s`
- policy under `policy/` is not obviously reconciled by Argo

Required outcome:
- either expand Argo scope
- or document a second Application / app-of-apps structure

### EF-0009 — Turn per-session network policy from template to actual mechanism
Problem:
- `deploy/k8s/networkpolicy.yaml` reads as template/example, not full dynamic implementation

Required outcome:
- define how per-session namespaces get their policies applied
- likely via runtime provisioning code or namespace bootstrap controller/job

### EF-0010 — Complete signed-image verification trust material
Problem:
- current Kyverno signed-image policy contains placeholder public key material

Required outcome:
- choose and document real verification mode:
  - pinned public key(s)
  - or keyless verification path

## Suggested implementation order
1. EF-0001
2. EF-0004
3. EF-0002
4. EF-0003
5. EF-0005
6. EF-0006 / EF-0007
7. EF-0008 / EF-0009 / EF-0010
