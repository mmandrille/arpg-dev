# v332 — pathfinding-cell-accuracy

**Status:** Draft
**Date:** 2026-06-24
**Codename:** `pathfinding-cell-accuracy`

---

## Purpose

Fix 9 pre-existing test failures introduced by the v330 room-layout + v332 movement-overhaul
batch. All 9 failures share one root cause: `buildBlockedFn` probes cell **origins**
`(gx*cellSize, gy*cellSize)` when checking whether a pathfinding cell is navigable. A cell
origin can sit just outside a wall AABB while the cell **center** sits inside it. A* routes
the player through such cells; continuous movement then stalls when the player's floating-point
position drifts to the wall boundary.

The fix splits `buildBlockedFn` into two checks:

1. **Static walls** → probed at cell **center** `(gx*cellSize + cellSize/2, gy*cellSize + cellSize/2)`,
   which correctly marks wall-adjacent cells as impassable.
2. **Dynamic entities** (live monsters, closed-door barriers) → probed at cell **origin** (current
   behavior), preserving the closed-door approach-from-player's-side behavior that the door test relies on.

No gameplay values, movement speed, or dungeon-generation reachability validator are changed.

---

## Non-goals

- Changing monster AI pathfinding (separate approach pipeline).
- Updating `generatedTargetReachableFrom` in `dungeon_gen.go` (dungeon generation reachability
  uses origin-check intentionally and is separately validated).
- Any balance tuning, new features, or client changes.
- Fixing the remaining maintainability debt (`main.gd`, `rules.go` type clusters).

---

## Acceptance criteria

All checks are server-side Go tests and existing CI gates:

1. `go test ./internal/game/... ./internal/replay/... -count=1` exits 0 — all 9 currently-failing
   tests pass:
   - `TestRangedProjectileGolden` (both subtests)
   - `TestProjectileBusyRejectsSecondFire`
   - `TestDirectionalRangedFreeShotHitsAndOmitsTargetID`
   - `TestRangedAutoApproachThenFire`
   - `TestRangedDummyDropsSeparatedLootItems`
   - `TestRangedBowLootRequiresMeleeReach`
   - `TestRangedBlockedLineAutoMovesUntilClearThenFires`
   - `TestActionIntentAutoApproachAndAttack`
   - `TestDungeonTeleportersReplayGolden`
2. `TestActionAutoApproachQueuesWhenOutOfRange/door` stays green (door approach from player's side
   still finds a valid goal cell).
3. `TestCollisionIgnoresDeadMonster`, `TestCollisionBlocksWallAndAllowsRoute` stay green (wall
   routing through collision_lab unchanged or improved).
4. `make lint-determinism` passes (no bare map ranges or time.Now introduced).
5. `make validate-shared` passes (no schema/rules changes needed).
6. `go vet ./...` clean.

---

## Scope and files likely touched

### Production

| File | Change |
|------|--------|
| `server/internal/game/sim.go` | Replace `buildBlockedFn` single-probe implementation with a two-part check: (a) wall check at cell center, (b) dynamic-entity check via new `playerDynamicBlocked(pos Vec2) bool` helper at cell origin. Extract `playerDynamicBlocked` from the existing inline loop in `playerPositionBlocked`. |

`buildBlockedFn` change sketch:

```go
func (s *Sim) buildBlockedFn() func(gx, gy int) bool {
    nav := s.activeNav()
    return func(gx, gy int) bool {
        origin := gridToWorld(nav, gridCell{x: gx, y: gy})
        // Cell center: correct probe for wall AABBs.  Cells adjacent to a wall
        // AABB have their center inside the AABB even though their origin is
        // outside — routing through them stalls continuous navigation.
        center := Vec2{X: origin.X + nav.CellSize/2, Y: origin.Y + nav.CellSize/2}
        for _, wall := range s.activeWalls() {
            if obstacleBlocksMovement(wall) && circleIntersectsAABB(center, playerRadius, wall.pos, wall.size) {
                return true
            }
        }
        // Cell origin: dynamic entities use origin so closed-door barriers don't
        // block approach cells on the player's side of the door.
        return s.playerDynamicBlocked(origin)
    }
}
```

No changes to `approach.go`, `pathfind.go`, `auto_nav.go`, or `dungeon_gen.go`.

### Tests

| File | Change |
|------|--------|
| `server/internal/game/combat_control_test.go` | With shorter/cleaner paths (wall-adjacent cells now avoided), tick budgets raised in the first $refactor pass (80→150) may be reduceable back toward 80–120. **Only adjust if tests are flaky or dramatically over-budget.** |
| `server/internal/game/game_test.go` | `TestActionIntentAutoApproachAndAttack` tick guard (currently 200) may be reduceable once approach works correctly. |
| `shared/golden/dungeon_teleporters.json` | May need regeneration if the path through the generated dungeon now reaches the teleporter at a different final position. Run `go test -update -run TestDungeonTeleportersReplayGolden` to check; update golden if the new path is valid and the teleporter discovery outcome is correct. |

### Docs

| File | Change |
|------|--------|
| `docs/progress/slice-lifecycle.md` | Add v332 row on completion. |
| `docs/as-built/v332_pathfinding-cell-accuracy.md` | Write after slice ships. |
| `PROGRESS.md` | Update current status, CI gate, open gaps. |

---

## Test and bot proof

- **Unit tests:** The 9 failing tests listed in acceptance criteria are the primary proof. All must
  pass without relaxing their assertions.
- **Regression guard:** All currently-passing game and replay tests must stay green.
- **Smoke (client-unit / client-smoke):** Not changed by this slice — skip client-unit if CI is
  too slow; client-smoke requires a running server. Run `make client-unit` if time allows.
- **Bot / replay:** `TestDungeonTeleportersReplayGolden` is a replay determinism test; if the
  golden needs regeneration, do so explicitly with `-update` and document the new expected position
  in the as-built.
- **No visual proof required** — this is a pure server navigation fix with no client changes.

---

## Open questions and risks

1. **Golden regeneration scope.** `dungeon_teleporters.json` may need updating. The test builds
   inputs dynamically, so the inputs themselves might change (different path, different tick count
   to reach the teleporter). The plan must include a step to run `-update` and verify the new
   golden represents correct behavior (player still discovers and uses the level -3 teleporter).

2. **Corridor accessibility.** If any generated dungeon corridor is exactly 1 cell wide (cell
   width 1.0 = wallThickness-clearance-1.0), and the corridor cell's center sits inside the wall
   AABB, the fix would mark that corridor as impassable. This would cause `GenerateDungeonLevel`
   to fail its reachability validation and generate an alternative layout — or the generation
   would fail under the max-attempts limit. The plan must include a run of
   `TestGeneratedDungeonTargetsReachable` across multiple seeds to confirm no regressions.

3. **Tick budget adjustments.** Some tests might now reach targets faster (cleaner paths avoid
   dead-end wall cells) or slower (longer detour paths). The plan should run all 9 failing tests
   first at current budgets; if any newly-passing test is flaky near its budget, tighten the
   guard by removing unnecessary headroom.

4. **`findApproachGoalMatching` goal position.** Approach goals use `goal = gridToWorld(nav, cell)` 
   (cell origin). Once wall-adjacent cells are excluded from the candidate set, goals will naturally
   be at cell origins that are not blocked by walls at either origin OR center. The inRange check
   may need the goal position to be within `playerMeleeReach + targetRadius + epsilon` of the target 
   — verify this still holds for doors and interactables now that the candidate pool is smaller.
