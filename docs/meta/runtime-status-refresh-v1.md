# Runtime Status Refresh v1 — cloudshell-fog

Purpose: record the current runtime baseline after the first merged hook-up wave so future Fog work starts from what is actually live on `main`.

## Current runtime baseline on `main`

Observed current state:
- the gateway entrypoint now calls `buildConnector(logger)`
- connector mode resolution lives in `cmd/gateway/connector_mode.go`
- `CONNECTOR_MODE=stub` and `CONNECTOR_MODE=k8s` are the explicit runtime modes
- Kubernetes REST config resolution prefers `KUBECONFIG` and otherwise attempts in-cluster config
- recent merged repo work has also added operator runbooks, spec/schema validation workflow, and refreshed repo control docs

## Why this note exists

Older conformance/backlog docs were written before the runtime hook-up merged. Those docs remain useful, but without a refresh note they can be misread as if connector selection were still entirely hypothetical.

## What is now clearly true

1. `cloudshell-fog` is the correct runtime home for Fog gateway behavior.
2. Shared Fog contracts live upstream in `SocioProphet/api-spec` under `fog/`.
3. Shared Fog deployment/profile composition lives upstream in `SocioProphet/manifests` under `fog/`.
4. Release-proof and trust-graph work belongs in `SocioProphet/prophet-platform`.
5. Shared policy contracts belong in `SocioProphet/policy-fabric`.

## Immediate consequences for next work

- runtime connector, placement, session, PTY, and runtime policy enforcement work should land here
- new shared schema/protocol fields should be pushed to `api-spec/fog/`
- new shared deployment/profile semantics should be pushed to `manifests/fog/`
- this repo should not become a second canonical source of shared Fog contracts
