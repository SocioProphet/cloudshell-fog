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
- Argo CD policy and Tekton application surfaces exist
- policy root Kustomize selects the keyless Kyverno bundle
- runtime image resolution uses the canonical resolver once this PR lands

That means EF-0001, EF-0002, EF-0004, EF-0005, EF-0006, EF-0007, EF-0008, and the keyless baseline portion of EF-0010 should now be read as completed baseline rather than outstanding work.

## Priority 0 — correctness / trust / deployability

### EF-0001 — Wire actual connector selection in gateway
Status:
- **completed on `main`**

Observed current state:
- `cmd/gateway/main.go` calls `buildConnector(logger)`
- `cmd/gateway/connector_mode.go` resolves `CONNECTOR_MODE`
- supported modes include `stub` and `k8s`
- unsupported or misconfigured modes fail fast

### EF-0002 — Verify and correct default runtime image reference
Status:
- **completed once this PR lands**

Observed current state:
- `internal/api/runtime_image.go` defines the canonical resolver
- `RUNTIME_IMAGE_REF` is the environment override
- explicit API `image_ref` still wins
- canonical dev/demo fallback is `ghcr.io/socioprophet/cloudshell-fog/runtime:dev`
- `CreateSession` uses `resolveRuntimeImageRef(...)` once this PR lands
- configuration docs describe `RUNTIME_IMAGE_REF` and production digest expectations

Residual follow-on:
- production deployments should set `RUNTIME_IMAGE_REF` to a pinned digest
- EF-0003 still owns final dev/demo versus production-safe image posture cleanup

### EF-0003 — Stop advertising mutable tags as production-safe defaults
Status:
- **partially mitigated; still open for final production posture**

Observed current state:
- production overlay guidance exists
- pinned Tekton task variant exists
- base deployment remains dev/demo-oriented

Remaining follow-on:
- continue separating dev/demo base manifests from production-safe overlays
- keep production hardening docs and manifests aligned

### EF-0004 — Resolve k8s connector credential model
Status:
- **implemented baseline and documented on `main`; continue tightening operational posture**

Observed current state:
- explicit in-cluster connector overlay exists
- explicit kubeconfig-based connector overlay exists
- deployment guide documents both modes

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
Status:
- **completed on `main`**

Observed current state:
- `deploy/argocd/application-policy.yaml` points Argo CD at `policy/`
- `policy/kustomization.yaml` selects the Kyverno baseline bundle
- `deploy/argocd/application-tekton.yaml` covers Tekton scope

### EF-0009 — Turn per-session network policy from template to actual mechanism
Status:
- **still open**

Required outcome:
- define how per-session namespaces get their policies applied
- likely via runtime provisioning code or namespace bootstrap controller/job

### EF-0010 — Complete signed-image verification trust material
Status:
- **keyless baseline selected on `main`; public-key fallback still requires real trust material if used**

Observed current state:
- keyless Kyverno verification policy exists
- signing trust model recommends keyless baseline
- root policy bundle selects keyless verification
- placeholder public-key policy remains as compatibility/example path, not bundled default

## Suggested implementation order
1. EF-0003
2. EF-0009
3. production hardening follow-through for EF-0010 where non-keyless trust material is required

## Working rule
Completed items may remain in this backlog when they explain how the repository moved from reviewed drift to live runtime behavior. The backlog should not silently pretend merged fixes are still missing.
