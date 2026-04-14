# Live Status v1 — cloudshell-fog

Last refreshed against canonical upstream after the first merge wave.

## 1. Current open PRs

At the time of this refresh, the canonical repo still has these draft PRs open:
- PR #6 — connector mode helpers for gateway
- PR #15 — operator runbooks and spec validation workflow

## 2. Recently merged PRs from the hardening train

The following PRs from the additive hardening wave are now merged:
- PR #7 — runtime image resolution and k8s credential model
- PR #8 — production overlay + pinned Tekton + Argo scope alignment
- PR #9 — session namespace policy pack
- PR #10 — keyless signing trust model
- PR #11 — threat model + runtime isolation + ADRs
- PR #12 — merge order + validation checklist
- PR #13 — structured placement response helper
- PR #14 — k8s connector overlays

## 3. What this means

The repository has materially advanced. The stack is no longer a hypothetical train; most of it is already on `main`.

Remaining work is now much narrower and concentrated in a few hook-up seams.

## 4. Residual high-value tasks

### RH-0001 — finish PR #6
Apply the `main.go` hook-up so the gateway actually calls `buildConnector(logger)`.

### RH-0002 — merge PR #15
Land the runbooks and spec-contract validation workflow.

### RH-0003 — complete the placement response hook-up
PR #13 merged the helper and tests, but the handler hook-up note still needs to be applied in `internal/api/handler.go` if the richer placement object is desired in the live API.

## 5. Working guidance

From this point onward, agents should not assume the earlier merge-order docs are current.
They should first read this file, then verify current open PR state before proposing more layers.
