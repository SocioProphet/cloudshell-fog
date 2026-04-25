# FogStack Canonical Upstream Map v0

Purpose: record the current canonical upstream split so runtime work in `cloudshell-fog` stays aligned with the merged Fog foundations and does not recreate parallel ownership.

## Canonical upstream homes

### 1. Shared contract home
Repository: `SocioProphet/api-spec`
Subtree: `fog/`

Owns:
- OpenAPI for Fog topology and planner APIs
- AsyncAPI for telemetry and alerts buses
- JSON Schemas for Fog identity, posture, deployment profile, and claim objects
- CUE profile bundle seeds
- ADRs and example profile documents

Rule:
- shared protocol / schema changes should land here first or in lockstep with runtime updates

### 2. Deployment and profile home
Repository: `SocioProphet/manifests`
Subtree: `fog/`

Owns:
- base kustomize packaging for Fog deployments
- overlays for `single-home`, `multi-home`, and `regional-multimesh`
- site seed demos and deployment/profile notes

Rule:
- deployment composition should land here, not in the runtime repo, unless a repo-local example is purely illustrative

### 3. Runtime gateway home
Repository: `SocioProphet/cloudshell-fog`

Owns:
- gateway runtime behavior
- connector selection and runtime connector implementations
- placement engine behavior
- policy enforcement behavior at runtime
- session lifecycle and PTY handling
- operator runbooks and repo-local runtime hardening docs

Rule:
- runtime changes belong here, but they should reference the canonical `api-spec` and `manifests` surfaces instead of redefining them

### 4. Release-proof and trust graph lane
Repository: `SocioProphet/prophet-platform`

Owns:
- Fog release sealing and release-proof execution
- linked trust/evidence graph around releases
- CI-backed proof execution for release artifacts

Rule:
- release trust / proof work belongs there, not here

### 5. Policy contract lane
Repository: `SocioProphet/policy-fabric`

Owns:
- policy decision contract surfaces that multiple runtimes may consume

Rule:
- shared policy contracts should land there first; runtime-specific enforcement adapters can live here

## Immediate alignment implications for `cloudshell-fog`

1. Keep runtime connector, placement, policy, session, and attach-path work in this repo.
2. Do not create duplicate OpenAPI / AsyncAPI / JSON Schema authority here for shared Fog objects.
3. When runtime needs new shared contract fields, update `api-spec/fog/` in lockstep.
4. When runtime needs new deployment/profile semantics, update `manifests/fog/` in lockstep.
5. Use this repo for runtime implementation and operator documentation, not as the sole canonical source of all Fog doctrine.
