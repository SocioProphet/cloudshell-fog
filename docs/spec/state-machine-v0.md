# Session State Machine v0 — cloudshell-fog

This document formalises the session lifecycle reflected by the current implementation and clarifies which states are implemented now versus reserved for future maturity.

## 0. Purpose

The session state machine exists to make the following explicit:
- legal states
- legal transitions
- idempotency expectations
- timeout / expiry behavior
- future extension points for richer fog-degraded semantics

## 1. Current implemented states

The currently reviewed implementation uses these states:

- `pending`
- `running`
- `terminated`
- `expired`

These states are represented in the in-memory session store and used by the API / sweeper / PTY logic.

## 2. State meanings

### `pending`
Session has been accepted by the control plane but runtime allocation is not yet fully completed.

Current implementation behavior:
- session record is created with `pending`
- runtime allocation is attempted immediately after record creation

### `running`
Runtime allocation succeeded and the session is considered attachable.

Current implementation behavior:
- PTY attach is allowed only when session is `running`

### `terminated`
Session has been explicitly terminated or allocation failed in a way that caused teardown.

Current implementation behavior:
- API `DELETE /v1/sessions/{id}` converges here
- failed runtime allocation path also forces a terminal state

### `expired`
Reserved terminal state for TTL-driven expiry semantics.

Current implementation note:
- current sweeper removes expired sessions and emits termination audit
- implementation may not persist `expired` long enough for it to be externally visible

## 3. Current legal transitions

### create flow
`(none)` → `pending`
- session record created

`pending` → `running`
- runtime allocation succeeds

`pending` → `terminated`
- runtime allocation fails after record creation

### delete flow
`pending` → `terminated`
`running` → `terminated`
- explicit delete attempts runtime teardown and converges to terminal state

### expiry flow
`pending|running` → `expired` or effective termination
- current sweeper semantics are implementation-driven and may remove the session record after termination callback

## 4. PTY attach gate

PTY attach is currently valid only when:
- session exists
- subject matches token subject
- session token is valid and session-scoped
- session status is `running`

All other states must reject attach.

## 5. Idempotency expectations

### DELETE
Delete should be treated as idempotent from the API contract perspective.
Repeated delete requests for the same session should converge to the same terminal truth.

### GET
GET must not mutate state.

### POST
POST creates a new session ID each time in the current implementation.
Idempotency-key support is future work.

## 6. Race conditions to keep in mind

### terminate vs allocate
If termination arrives while allocation is still in flight, implementation must converge to a terminal truth and avoid orphaned runtimes.

### attach vs terminate
If attach races with terminate, terminate wins. Attach should fail or disconnect.

### sweeper vs attach
If the TTL sweeper expires a session while attach is being attempted, attach should fail or disconnect cleanly.

## 7. Reserved future states (not yet implemented as first-class states)

The following states are useful for a more mature fog-aware lifecycle but are not yet clearly represented as first-class status values in reviewed code:

- `attach_ready`
- `disconnected`
- `failed`
- `resuming`

These should be added only when implementation is ready to expose them truthfully.

## 8. Recommended next evolution

Short-term:
- keep current state names but document behavior precisely
- make expiry semantics explicit (`expired` vs direct removal)

Later:
- introduce `disconnected` when fog partition / attach drop semantics are surfaced as user-visible truth
- introduce `failed` for clearer operational debugging
