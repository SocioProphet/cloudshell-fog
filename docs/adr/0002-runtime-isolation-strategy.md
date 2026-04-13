# ADR 0002 — Containers-first runtime isolation, stronger path later

Status: accepted

## Context

The repository already contains a Kubernetes connector and hardened pod security posture. We need a practical default while preserving a path to stronger isolation later.

## Decision

Use hardened Kubernetes-backed containers as the default runtime isolation model for the current phase.

Preserve the connector abstraction so stronger backends (for example VM-backed or microVM-backed approaches) can be introduced later without rewriting the entire control plane.

## Rationale

- fits existing implementation
- minimizes complexity while closing the highest-value seams first
- allows the project to move from MVP to production-track posture before adding another execution substrate

## Consequences

- current security depends heavily on namespace, pod, and policy hardening
- stronger runtime backends remain future work, not blocked design work
