# v83 Spec - Defensive Client Envelope Payloads

Status: Complete
Date: 2026-06-11
Codename: `defensive-client-envelope-payloads`

## Purpose

The Godot client should tolerate malformed or partial server envelopes without crashing in
`_handle_message`. The v80 review identified direct `env["payload"]` indexing in the central
message handler even though delta handling already prefers defensive `.get()` access.

## Non-goals

- No protocol/schema changes.
- No server behavior changes.
- No client UI or presentation feature changes.
- No broad `main.gd` extraction; this slice only hardens the message payload boundary.

## Acceptance Criteria

- `_handle_message` reads envelope payloads through a defensive dictionary helper or equivalent
  `.get("payload", {})` path.
- Missing, null, or non-dictionary payloads do not crash for `session_snapshot`, `state_delta`,
  `intent_accepted`, `intent_rejected`, or `error` envelopes.
- Existing valid snapshot, delta, accept/reject, and error handling behavior is preserved.
- Headless client unit coverage pins malformed envelope behavior.
- `make client-unit`, `make maintainability`, and `make ci` pass before commit.

## Scope and Likely Files

- `client/scripts/main.gd`
- `client/tests/test_delta_apply.gd`
- `docs/plans/v83_2026-06-11-defensive-client-envelope-payloads.md`
- `docs/as-built/v83_defensive-client-envelope-payloads.md`
- `PROGRESS.md`

## Test and Bot Proof

- `make client-unit`
- `make maintainability`
- `make ci`

No bot scenario is required because this is malformed-envelope hardening in the client unit layer,
not new gameplay or presentation behavior.

## Open Questions and Risks

- No blocking questions.
- Risk: `_apply_snapshot({})` may touch scene-tree paths on a bare `MainScript.new()` test object.
  Unit coverage should choose malformed cases that exercise `_handle_message` without requiring
  scene nodes, or initialize only the state needed for the pure path.
