# TurtleTerm Integration Profile v0 — CloudShell FOG

## Purpose

TurtleTerm is the SourceOS policy-aware, agent-addressable terminal workbench for trusted command execution, terminal receipts, agent delegation, and reproducible operator workflows.

CloudShell FOG is the browser/fog/cloud shell execution plane for session lifecycle, placement, PTY attach, runtime allocation, and audit.

This profile defines how the two systems should interoperate without creating duplicate terminal truth systems.

## Ownership boundary

### CloudShell FOG owns

- browser / fog / cloud shell session lifecycle
- session placement and runtime allocation
- WSS PTY attach contract
- Kubernetes/fog runtime connector behavior
- CloudShell audit events
- placement metadata (`region`, `node_id`, `tier`, `reasons`)

### TurtleTerm / SourceOS terminal contracts own

- local/operator terminal command lifecycle receipts
- command stdout/stderr digests
- SourceOS terminal session/event/receipt schemas
- local agent terminal workflow metadata
- reproducible operator command receipts

## Integration principle

CloudShell FOG should reference TurtleTerm receipt contracts where command-level receipts are needed.
It should not invent a parallel receipt schema for local/operator command execution.

## Correlation model

A CloudShell session MAY propagate SourceOS terminal context into commands or operator workflows:

- `SOURCEOS_TERMINAL_SESSION_ID` = CloudShell session ID or derived stable session identifier
- `SOURCEOS_WORKSPACE` = CloudShell workspace / project identifier, if known
- `SOURCEOS_ACTOR_ID` = authenticated subject or mapped actor identity
- `SOURCEOS_POLICY_BUNDLE_ID` = CloudShell policy/profile identifier, if known
- `SOURCEOS_EXECUTION_DOMAIN` = `cloudshell-fog`, `k8s`, `fog`, or more specific runtime domain

## Event mapping

| CloudShell FOG concept | TurtleTerm / SourceOS terminal concept |
|---|---|
| `session.created` | `sourceos.terminal.session.v0` |
| `session.attached` | terminal frontend attach context |
| `runtime.allocated` | execution domain / runtime metadata |
| `placement.decided` | placement metadata attached to receipt context |
| command execution inside terminal | `command.started` / `command.completed` receipt events |
| `session.terminated` | terminal session completion / teardown context |

## Placement metadata

When CloudShell FOG launches or coordinates a TurtleTerm-backed command workflow, it SHOULD preserve:

- selected region
- selected node ID
- trust tier
- placement reasons
- runtime image or runtime profile

These fields should be attached as receipt metadata rather than replacing TurtleTerm's receipt schema.

## Non-goals

- CloudShell FOG does not replace TurtleTerm's local receipt schema.
- TurtleTerm does not own CloudShell FOG's placement engine.
- This profile does not require all CloudShell PTY streams to emit per-command receipts by default.

## Open questions

1. Should SourceOS terminal schemas move into a shared terminal-contracts repository later?
2. Should CloudShell FOG expose a receipt-export endpoint for completed sessions?
3. Should AgentPlane be the canonical bridge for launching TurtleTerm workflows from CloudShell FOG?

## Tracking

- CloudShell FOG: issue #35
- TurtleTerm: issue #1
