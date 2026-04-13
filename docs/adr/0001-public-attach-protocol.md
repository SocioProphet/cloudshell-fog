# ADR 0001 — Public attach protocol uses WSS frames, not SSH

Status: accepted

## Context

The browser-facing shell attach path must be usable from web clients and auditable at the message level.

## Decision

Use WebSocket-based PTY attach with an explicit JSON frame model for the public attach surface.

SSH may still exist behind the scenes as a runtime detail or debugging path, but it is not the primary public protocol.

## Rationale

- aligns with browser-native clients
- explicit message model supports clearer auditing and schema documentation
- keeps public attach semantics separate from low-level runtime details

## Consequences

- PTY message schema becomes part of the stable contract surface
- WebSocket security and token-binding behavior must remain explicit
