# ADR 0003 — Kyverno baseline for admission, broader policy via OPA/Rego if needed

Status: accepted

## Context

The repository already contains Kyverno policy artifacts and a separate product-local policy engine. We need a clear baseline statement to reduce future ambiguity.

## Decision

Use Kyverno as the current Kubernetes admission-policy baseline.

Keep OPA/Rego available as a broader policy language when cross-service or non-Kubernetes policy composition becomes necessary.

## Rationale

- matches the current repo artifacts
- reduces mismatch between docs and live policy files
- allows cluster admission controls to remain concrete and approachable

## Consequences

- Kyverno policy remains the first-class cluster policy expression in the current phase
- broader platform governance can still adopt OPA/Rego where that becomes valuable
