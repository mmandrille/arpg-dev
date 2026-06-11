# v82 Spec - Realtime Fanout Level Snapshot

Status: Complete
Date: 2026-06-11
Codename: `realtime-fanout-level-snapshot`

## Purpose

Realtime fanout should use the same per-client level view that was observed while the session loop
mutex was held for the tick. This closes the v70/v80 review finding where `fanoutResult` queried
`Sim.PlayerCurrentLevel` after unlocking, which could make fanout observe level state that drifted
from the tick result being persisted.

## Non-goals

- No protocol/schema changes.
- No gameplay, movement, transition, persistence, or replay behavior changes.
- No broad realtime loop refactor beyond the fanout level snapshot boundary.
- No client presentation changes.

## Acceptance Criteria

- `doTick` snapshots every connected client's current level while holding `sessionLoop.mu`.
- `fanoutResult` no longer calls `Sim.PlayerCurrentLevel`; it consumes the tick-time level snapshot.
- Existing same-level, actor-only, and cross-level fanout behavior is preserved.
- Focused realtime tests prove fanout can be driven from an explicit client-level snapshot.
- `make maintainability`, focused realtime Go tests, and `make ci` pass before commit.

## Scope and Likely Files

- `server/internal/realtime/session_loop.go`
- `server/internal/realtime/session_loop_test.go`
- `docs/plans/v82_2026-06-11-realtime-fanout-level-snapshot.md`
- `docs/as-built/v82_realtime-fanout-level-snapshot.md`
- `PROGRESS.md`

## Test and Bot Proof

- `cd server && go test ./internal/realtime/...`
- `make maintainability`
- `make ci`

No bot scenario is required because this slice changes realtime fanout internals without adding
gameplay, protocol, or player-facing behavior.

## Open Questions and Risks

- No blocking questions.
- Risk: tests that call `fanoutResult` directly need to pass the new client-level snapshot. Keep
  their setup explicit so the concurrency boundary remains visible.
