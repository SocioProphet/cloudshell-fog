# SCOPE-D Delivery Envelopes

CloudShell Fog participates in the SCOPE-D platform as an authorized edge operator bastion and mesh assurance node.

This document defines the first delivery boundary between SCOPE-D and CloudShell Fog.

## Role

CloudShell Fog is not a scanner and not an autonomous execution agent in this flow. It accepts reviewable, policy-gated SCOPE-D delivery envelopes for edge assurance workflows.

The initial envelope validator accepts only non-executing review packages.

## Accepted source

`sourceSystem` must be:

```text
scope-d
```

## Accepted purposes

- `edge_assurance_review`
- `policy_gated_delivery_review`

## Required fields

A delivery envelope must include:

- schema version;
- envelope id;
- source system;
- purpose;
- artifact references;
- required policy references;
- operator approval requirement;
- non-execution flags;
- receipt hash.

## Prohibited capabilities in v0.1

The validator rejects any envelope that requests:

- execution;
- prior execution;
- network access;
- mutation;
- credential access;
- payload delivery.

## Expected SCOPE-D artifact references

SCOPE-D should reference artifacts such as:

- intelligence enrichment export;
- detection candidate export;
- cyber graph export;
- operator case bundle;
- client assurance report;
- PolicyFabric approval record.

## Validation path

The Go package is:

```text
internal/delivery
```

The validator is:

```go
Validate(envelope DeliveryEnvelope) error
```

The receipt helper is:

```go
ComputeReceiptHash(envelope DeliveryEnvelope) string
```

## Test path

```bash
go test ./internal/delivery
```

## Next slices

1. Add API endpoint for staging delivery envelopes.
2. Persist staged envelopes in the existing session/audit subsystem.
3. Add PolicyFabric signature verification.
4. Add SCOPE-D receipt verification.
5. Add edge sync status for Noetica.
6. Add read-only operator review endpoint.
7. Add deployment receipt generation after approved future delivery modes.
