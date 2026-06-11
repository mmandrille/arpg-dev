# Spec: `runtime-sim-error-construction`

Status: Implemented
Date: 2026-06-11
Branch: `main`
Codename: `runtime-sim-error-construction`
Slice: v63 - runtime sim error construction
Baseline: v62 `monster-depth-stat-scaling`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../reviews/20260610_v60-overview.md`](../reviews/20260610_v60-overview.md) - recommends removing runtime `NewSim` panic paths
- [`../reviews/backend/20260610_v60-backend.md`](../reviews/backend/20260610_v60-backend.md) - identifies default sim construction as a high-risk crash path
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - deterministic authoritative sim and replay

## 1. Purpose

The v60 backend review found that `game.NewSim` still panics when default sim construction fails.
Runtime code mostly uses `NewSimWithWorldProgression`, but the exported panic wrapper remains easy
to call from future server paths. This slice removes that footgun from production code by making
default-world construction return an error and keeping panic behavior only as an explicit test
helper.

The slice proves:

- Runtime callers have an error-returning default-world constructor.
- Panic-on-construction is no longer the exported default API.
- Existing tests that intentionally want terse setup use a clearly named `MustNewSim` helper.
- Invalid world/rule construction errors can be asserted without crashing the process.

## 2. Non-goals

- No gameplay, protocol, persistence, replay, or client behavior changes.
- No broad rewrite of every test to handle constructor errors inline.
- No changes to rule validation semantics.
- No panic recovery wrapper around production runtime paths; callers should receive normal errors.

## 3. Acceptance Criteria

1. `server/internal/game.NewSim` returns `(*Sim, error)` for the default world.
2. A clearly named `server/internal/game.MustNewSim` helper exists for tests and panics only when
   called explicitly.
3. Production/runtime packages do not call `MustNewSim`.
4. Existing game and realtime tests are updated to use either `NewSim` with error handling or
   `MustNewSim` where the rule set is already known valid.
5. Focused tests cover invalid default-world construction returning an error without a panic.
6. `make test-go` and `make ci` pass.

## 4. Scope And Likely Files

```text
server/internal/game/sim.go - change NewSim signature and add MustNewSim
server/internal/game/game_test.go - update terse test setup calls
server/internal/game/shop_test.go - update terse test setup calls
server/internal/realtime/session_loop_test.go - update test setup call
docs/plans/v63_2026-06-11-runtime-sim-error-construction.md - implementation plan
PROGRESS.md - lifecycle update when v63 ships
docs/as-built/v63_runtime-sim-error-construction.md - as-built summary
```

## 5. Test And Bot Proof

- Go tests cover constructor error behavior and all existing sim tests.
- No bot scenario is required because this is an internal backend safety refactor with no gameplay
  or protocol behavior change.
- `make ci` remains the landing gate.

## 6. Open Questions And Risks

- The main risk is mechanical churn in Go tests. Keep the production API change small and use
  `MustNewSim` only where tests already rely on valid shared fixtures.
