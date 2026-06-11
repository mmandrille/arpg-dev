# v83 As-built - Defensive Client Envelope Payloads

## What shipped

The Godot client's `_handle_message` path now normalizes envelope payloads through
`_envelope_payload`, returning an empty dictionary for missing, null, or non-dictionary payloads.
Snapshot, delta, accepted, rejected, and error message handling all use the guarded payload.

## What it proves

- Malformed accepted/rejected/error/delta envelopes no longer crash central client message handling.
- Valid accepted messages still clear matching pending action targets.
- Malformed rejected messages still clear pending interactable and waypoint action state.
- The change is client-only and does not alter protocol or server behavior.

## Verification

```bash
make client-unit
make maintainability
make ci
```

## Deferred

- Combat event presenter extraction.
- Broader `main.gd` decomposition.
