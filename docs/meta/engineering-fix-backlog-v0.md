# Engineering Fix Backlog v0 — cloudshell-fog

Purpose: convert conformance findings into concrete, implementation-facing tasks while preserving the distinction between work that is already live on `main` and work that still remains open.

## Current status checkpoint

This backlog originally captured the post-review gaps before the first runtime hook-up wave landed.

As of the current `main` line:
- connector selection is **wired** through the gateway entrypoint
- `CONNECTOR_MODE` resolution exists
- `stub` and `k8s` modes exist as explicit runtime choices
- unsupported modes fail fast at startup
- session create and session status now emit structured placement objects
- session state now persists node / tier / reasons placement metadata
- repo-local state-machine and machine-readable interface contract artifacts exist under `docs/spec/`

That means EF-0001, EF-0005, EF-0006, and EF-0007 should now be read as completed baseline rather than outstanding work.

## Priority 0 — correctness / trust / deployability

### EF-0001 — Wire actual connector selection in gateway
Status:
- **completed on `main`**

Observed current state:
- `cmd/gateway/main.go` calls `buildConnector(logger)`
- `cmd/gateway/connector_mode.go` resolves `CONNECTOR_MODE`
- supported modes include `stub` and `k8s`
- unsupported or misconfigured modes fail fast

Residual follow-on:
- keep deployment docs/manifests aligned with the now-live connector path
- keep credential-model documentation explicit as EF-0004 lands

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
Status:
- **implemented baseline and documented on `main`; continue tightening operational posture**

Observed current state:
- explicit in-cluster connector overlay exists
- explicit kubeconfig-based connector overlay exists
- deployment guide documents both modes

Remaining follow-on:
- keep overlays, docs, and deployment guidance aligned as operator posture evolves

## Priority 1 — align external contracts to internal truth

### EF-0005 — Align session create response with richer placement model
Status:
- **completed on `main`**

Observed current state:
- `POST /v1/sessions` returns a structured placement object
- `GET /v1/sessions/{id}` now also returns a structured placement object
- session state persists node / tier / reasons metadata

Residual follow-on:
- keep shared Fog contract surfaces in lockstep if this runtime shape is promoted upward

### EF-0006 — Add formal state-machine spec
Status:
- **completed on `main`**

Observed current state:
- `docs/spec/state-machine-v0.md` exists and documents current legal states, transitions, idempotency, race handling, and future extension points

### EF-0007 — Add machine-readable interface contracts
Status:
- **completed on `main`**

Observed current state:
- `docs/spec/openapi/control-api.v1.yaml` exists
- `docs/spec/jsonschema/ws-pty.v1.json` exists
- `docs/spec/jsonschema/audit-event.v1.json` exists

Residual follow-on:
- keep the OpenAPI and JSON Schema surfaces aligned with runtime response-shape evolution

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
1. EF-0002
2. EF-0003
3. EF-0008
4. EF-0009
5. EF-0010

## Working rule
Completed items may remain in this backlog when they explain how the repository moved from reviewed drift to live runtime behavior. The backlog should not silently pretend merged fixes are still missing.