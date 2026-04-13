# ADR 0004 — Single canonical repo, layered additive PRs

Status: accepted

## Context

The project now contains implementation code, specs, deployment artifacts, policy, and CI/CD resources. The question is whether to split this surface prematurely into multiple repositories.

## Decision

Keep `SocioProphet/cloudshell-fog` as the canonical implementation repo and `SociOS-Linux/cloudshell-fog` as the mirror.

Use layered, additive PRs for major seams instead of early repository splitting.

## Rationale

- minimizes coordination cost
- keeps code/spec/ops changes reviewable together
- matches the current execution pattern already used in the project

## Consequences

- merge discipline matters because more of the platform surface is co-located
- future repo splitting remains possible if the operational burden becomes real rather than hypothetical
