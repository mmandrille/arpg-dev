# v251 Plan - Direct Auto Navigation

Status: Implemented
Goal: Make authoritative auto-navigation pick direct shortest paths with cleaner corner behavior.
Architecture: Keep navigation server-owned. Improve A* tie-breaking in `PlanPath`, then store exact
auto-nav goals so player click-to-move can complete the final in-cell offset.
Tech stack: Go simulation, focused Go tests, docs.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/pathfind.go` | Prefer diagonal shortest routes, then fewer turns among equal diagonal paths |
| Modify | `server/internal/game/handlers.go` | Store planned exact/approach goal on queued auto-nav |
| Add | `server/internal/game/auto_nav.go` | Continue player auto-nav toward exact goal before finishing |
| Modify | `server/internal/game/sim.go` | Add exact-goal state and finish hook |
| Modify | `server/internal/game/pathfind_test.go` | Prove diagonal and corner path shape |
| Add | `server/internal/game/auto_nav_test.go` | Prove click-to-move reaches an in-cell exact target |

## Maintenance Ratchet

Touched hotspot:
- `server/internal/game/sim.go` is grandfathered at 6572 lines and must not grow beyond the
  ratchet allowance. Keep changes surgical.

Verification:
```bash
make maintainability
```

## Task 1 - Path Cost Tie-Breaking

- [x] Extend path node cost to preserve shortest path length first, then prefer diagonals, then
  fewer turns among equal diagonal routes.
- [x] Keep deterministic cell ordering for final ties.
- [x] Add focused path shape tests.

```bash
cd server && go test ./internal/game -run TestPlanPath
```

## Task 2 - Exact Move-To Completion

- [x] Store auto-nav goal positions for move/action/skill queued navigation.
- [x] When grid steps are exhausted, move the player toward the exact goal until within stop
  distance, then finish pending action/skill dispatch if any.
- [x] Add a focused move-to test for a target with an in-cell Y offset.

Additional guard:
- [x] Monster chase-goal selection skips a candidate when the first planned movement resolves to no
  movement, preserving boss preferred-target behavior around nearby dynamic blockers.

```bash
cd server && go test ./internal/game -run TestMoveTo
```

## Final Verification

- [x] `cd server && go test ./internal/game -run 'TestPlanPath|TestMoveTo'`
- [x] `cd server && go test ./internal/game -run 'TestBossAggroPreferredTargetWinsOverNearestPlayer|TestPlanPath|TestMoveTo'`
- [x] `cd server && go test ./internal/game`
- [ ] `make maintainability` - blocked by pre-existing oversized-file ratchet failures in
  `server/internal/game/game_test.go`, `server/internal/realtime/runner.go`,
  `server/internal/realtime/session_loop.go`, `server/internal/store/repos.go`, and
  `server/internal/store/store_test.go`.
