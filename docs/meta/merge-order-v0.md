# Merge Order v0 — cloudshell-fog stacked PR train

This document exists to make the current PR train mergeable without guesswork.

## Current draft PR stack

1. PR #6 — connector mode helpers for gateway
2. PR #7 — runtime image resolution + k8s credential model
3. PR #8 — production overlay + pinned Tekton + Argo scope alignment
4. PR #9 — session namespace policy pack
5. PR #10 — keyless signing trust model
6. PR #11 — threat model + runtime isolation + ADRs

## Recommended merge sequence

### Step 1 — finish and merge PR #6 first
Reason:
- it unlocks the real execution path (`stub` vs `k8s`)
- all other operational work assumes that seam is closed

Required before merge:
- apply the documented `main.go` hookup patch from:
  - `docs/meta/ef-0001-main-go-hookup.md`

### Step 2 — merge PR #7
Reason:
- clarifies runtime image resolution
- clarifies k8s credential model

### Step 3 — merge PR #8
Reason:
- makes deployment/GitOps guidance consistent with production posture

### Step 4 — merge PR #9
Reason:
- makes per-session namespace isolation operationally explicit

### Step 5 — merge PR #10
Reason:
- hardens signed-image trust posture

### Step 6 — merge PR #11
Reason:
- finalizes decision memory and security/isolation prose around the merged implementation train

## Alternative

PRs #7 through #11 can be batch-merged after PR #6 if desired because they are mostly additive and non-conflicting. PR #6 remains the only one with the unresolved existing-file seam.
