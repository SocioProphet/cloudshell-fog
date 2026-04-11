# Fog Placement Specification v0 — cloudshell-fog

This document defines the first formal placement model for cloudshell-fog.

## 0. Goal

Choose the best runtime location for an interactive shell session while respecting:

- latency and user experience
- resource capacity
- data locality
- trust posture
- graceful degradation to cloud fallback

## 1. Placement entities

Each candidate node or cluster advertises at least:

- `node_id`
- `region`
- `tier` (`fog` or `cloud`)
- `healthy` (boolean)
- `free_cpu`
- `free_memory`
- `free_storage`
- `latency_ms` (estimated to requester)
- `trust_tier` (for example: `attested`, `managed`, `unknown`)
- optional locality / sovereignty labels

## 2. Hard filters

Before scoring, reject candidates that fail any hard requirement:

- unhealthy
- insufficient free capacity for the requested profile
- outside required locality / sovereignty bounds
- below minimum trust tier for the request
- administratively quarantined

A filtered-out node is not eligible, regardless of latency.

## 3. Soft scoring

Among eligible candidates, score using weighted factors.

Illustrative factors:

- latency
- remaining headroom after placement
- trust tier preference
- locality preference match
- current session-stickiness preference (resume near prior placement)

Representative scoring intuition:

- prefer attested fog over generic cloud when both satisfy constraints
- prefer healthier and roomier nodes over nearly saturated nodes
- prefer closer nodes when trust and capacity are comparable

## 4. Tie-break strategy

If scores are effectively tied:

1. prefer `fog` over `cloud`
2. prefer higher `trust_tier`
3. prefer lower latency
4. prefer more remaining capacity
5. stable deterministic tie-break on `node_id`

## 5. Fallback semantics

If no eligible fog candidate exists:

- fall back to the configured trusted cloud region
- emit explicit placement reasons (`fog unavailable`, `capacity exhausted`, `trust insufficient`, etc.)

This fallback is not a silent implementation detail. It is part of the product contract.

## 6. Resume semantics

Session resume is best-effort.

Cases:

- if the original node is healthy and session state is alive, resume there
- if the original node is gone but a resumable backing state exists, restore in the nearest eligible node
- if neither is possible, surface non-resumable status explicitly to the user and create a new session only through an explicit request

We should not pretend a migrated session is identical if runtime state was lost.

## 7. Audit requirements

Every placement decision should emit:

- candidate set considered
- filters applied
- chosen node
- reason codes
- fallback indicator

This is required for operator trust and for debugging fog-placement pathologies.

## 8. Trust-tier guidance

Suggested trust tiers:

- `attested_fog` — fog node with attestation evidence and acceptable posture
- `managed_cloud` — known cloud fallback with managed controls
- `unverified` — available but below production preference

Requests may specify minimum trust constraints. High-assurance tenants may reject `managed_cloud` fallback entirely.

## 9. Reserved future work

- explicit placement score formula and weights
- session-affinity vs anti-affinity rules
- carbon / power / cost signals
- jurisdictional routing profiles
- integration of hardware attestation evidence into scheduling
