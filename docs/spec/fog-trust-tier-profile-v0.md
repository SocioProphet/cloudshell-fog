# Fog Trust-Tier Profile v0 — cloudshell-fog

This document defines the first explicit trust-tier model for fog and cloud placement candidates.

## 0. Goal

Separate simple reachability from actual runtime trustworthiness.

A node that is reachable and has free capacity is not automatically acceptable for a session. Placement and, in stricter environments, admission must consider trust posture explicitly.

## 1. Baseline trust tiers

### 1.1 `attested_fog`

Use when the candidate is a fog/edge node and can present acceptable trust evidence.

Representative signals:

- verified node identity
- measured or attested boot/runtime state
- acceptable secure-boot / TPM / hardware-root posture where available
- not quarantined
- healthy and schedulable

### 1.2 `managed_cloud`

Use for a known cloud fallback environment with managed controls but without the same edge-local attestation semantics.

Representative signals:

- cluster is known and operator-controlled
- baseline supply-chain and runtime policy are enforced
- environment is within accepted jurisdictional constraints

### 1.3 `unverified`

Use when the node may be technically reachable and schedulable but lacks sufficient trust evidence.

This tier may be acceptable for lower-assurance or development profiles, but it should not silently substitute for stronger tiers in production.

### 1.4 `quarantined`

Use when a node is administratively or automatically excluded.

Examples:

- failed attestation checks
- maintenance hold
- incident response isolation
- policy non-compliance

A quarantined node is ineligible for placement.

## 2. Hard vs soft trust use

### 2.1 Hard filtering

Placement must reject a node outright when:

- the request requires a minimum trust tier above the node's tier
- the node is quarantined
- the node fails mandatory sovereignty or trust constraints

### 2.2 Soft preference

Where multiple nodes satisfy the minimum, trust tier should still influence scoring.

Example preference order:

1. `attested_fog`
2. `managed_cloud`
3. `unverified`

## 3. Relationship to placement

Trust tier is an input to the placement engine, not merely an audit tag.

It should influence:

- candidate eligibility
- fallback behavior
- reason codes emitted in `placement.decided`
- resume behavior after node loss

## 4. Relationship to policy

The trust tier profile interacts with policy in two places.

### 4.1 Gateway policy

Requests may later specify a minimum acceptable trust tier.

Examples:

- default user profile: `managed_cloud` or better
- high-assurance profile: `attested_fog` only
- development profile: may allow `unverified`

### 4.2 Cluster / platform policy

Cluster admission may also reject scheduling into environments that do not meet the declared trust profile for the workload or tenant.

## 5. Evidence model (reserved)

This repo does not yet define the full evidence envelope, but likely sources include:

- substrate / node-attestation systems
- secure boot posture
- TPM-backed or equivalent node identity
- operator quarantine state
- cluster compliance labels

This is where future alignment with broader SourceOS and platform attestation work belongs.

## 6. Product consequences

The UI and API should never pretend that all placements are equivalent.

At minimum, placement results should be able to surface:

- chosen tier
- fallback status
- whether the placement met or only minimally satisfied the request

## 7. Reserved follow-on work

- map concrete node evidence into tier assignment
- bind tier constraints into session profile definitions
- decide whether trust tier should become part of audit receipts or contracts exported to prophet-platform
