# v332 plan — pathfinding-cell-accuracy

**Date:** 2026-06-24
**Spec:** [`v332_spec-pathfinding-cell-accuracy.md`](../specs/v332_spec-pathfinding-cell-accuracy.md)
**Branch:** `main`

---

## File map

| File | Change |
|------|--------|
| `server/internal/game/sim.go` | Replace `buildBlockedFn`; extract `playerDynamicBlocked` |
| `shared/golden/dungeon_teleporters.json` | Regenerate if test output differs |
| `server/internal/game/game_test.go` | Optionally tighten tick guards |
| `server/internal/game/combat_control_test.go` | Optionally tighten tick guards |
| `docs/progress/slice-lifecycle.md` | Add v332 row |
| `docs/as-built/v332_pathfinding-cell-accuracy.md` | Write after green |
| `PROGRESS.md` | Update status + CI gate |

---

## Step 1 — Implement `playerDynamicBlocked` helper

In `server/internal/game/sim.go`, extract the dynamic-entity loop (monsters + closed-door
barriers) from `playerPositionBlocked` into a new unexported helper. `playerPositionBlocked`
calls both the wall loop and the new helper to preserve existing movement-resolution behavior.

```go
// playerDynamicBlocked reports whether pos is blocked by a dynamic entity
// (live monster or closed-door barrier).  Walls are NOT checked here.
// Used by buildBlockedFn which needs separate probes for walls vs. entities.
func (s *Sim) playerDynamicBlocked(pos Vec2) bool {
    for _, id := range sortedEntityIDs(s.activeLevel().entities) {
        e := s.activeLevel().entities[id]
        if e.kind == monsterEntity && e.hp > 0 {
            if circlesOverlap(pos, playerRadius, e.pos, monsterRadius) {
                return true
            }
            continue
        }
        if e.kind == interactableEntity && e.state == interactableClosed {
            if def, ok := s.rules.Interactables[e.interactableDefID]; ok && def.BarrierWhenClosed != nil {
                if circleIntersectsAABB(pos, playerRadius, e.pos, def.BarrierWhenClosed.Size) {
                    return true
                }
            }
        }
    }
    return false
}
```

Update `playerPositionBlocked` to call it:

```go
func (s *Sim) playerPositionBlocked(pos Vec2) bool {
    for _, wall := range s.activeWalls() {
        if obstacleBlocksMovement(wall) && circleIntersectsAABB(pos, playerRadius, wall.pos, wall.size) {
            return true
        }
    }
    return s.playerDynamicBlocked(pos)
}
```

Verify: `go build ./...` clean, `go vet ./...` clean.

---

## Step 2 — Replace `buildBlockedFn` with split probe

Replace the existing `buildBlockedFn` with the two-probe version. The key invariant: wall
AABBs are checked at cell **center** (prevents routing through boundary cells); dynamic
entities are checked at cell **origin** (preserves closed-door approach from the player's
side).

```go
func (s *Sim) buildBlockedFn() func(gx, gy int) bool {
    nav := s.activeNav()
    return func(gx, gy int) bool {
        origin := gridToWorld(nav, gridCell{x: gx, y: gy})
        // Cell center: cells adjacent to a wall AABB have their center inside
        // the AABB even though their origin is outside.  Routing through them
        // stalls continuous navigation when position drifts to the boundary.
        center := Vec2{X: origin.X + nav.CellSize/2, Y: origin.Y + nav.CellSize/2}
        for _, wall := range s.activeWalls() {
            if obstacleBlocksMovement(wall) && circleIntersectsAABB(center, playerRadius, wall.pos, wall.size) {
                return true
            }
        }
        // Cell origin: closed-door barriers span the doorway; using center
        // would falsely block approach cells on the player's side.
        return s.playerDynamicBlocked(origin)
    }
}
```

Verify: `go build ./...` and `go vet ./...` still clean.

---

## Step 3 — Run the 9 target tests

```bash
cd server
go test ./internal/game/... ./internal/replay/... \
  -run "TestRangedProjectileGolden|TestProjectileBusyRejectsSecondFire|TestDirectionalRangedFreeShotHitsAndOmitsTargetID|TestRangedAutoApproachThenFire|TestRangedDummyDropsSeparatedLootItems|TestRangedBowLootRequiresMeleeReach|TestRangedBlockedLineAutoMovesUntilClearThenFires|TestActionIntentAutoApproachAndAttack|TestDungeonTeleportersReplayGolden" \
  -count=1 -v 2>&1 | grep -E "^--- (PASS|FAIL)"
```

**Expected:** all 9 PASS.

If `TestDungeonTeleportersReplayGolden` fails with a golden-mismatch (player position
differs but teleporter discovery is still correct):

```bash
go test ./internal/game/... -run TestDungeonTeleportersReplayGolden \
  -update -count=1
```

Verify the updated `shared/golden/dungeon_teleporters.json` has:
- `expected_level: -3` (player ends on level -3 after round-trip teleport)
- `discovered_teleporters` shows level -3 and town both `discovered: true`

The specific `expected_player_position` coordinates may change — that is acceptable as long
as the level and discovery state are correct.

---

## Step 4 — Run the critical regression tests

```bash
cd server
go test ./internal/game/... \
  -run "TestActionAutoApproachQueuesWhenOutOfRange|TestCollisionIgnoresDeadMonster|TestCollisionBlocksWallAndAllowsRoute|TestGeneratedDungeonTargetsReachable|TestDungeonObstacleGeneration" \
  -count=1 -v 2>&1 | grep -E "^=== RUN|^--- (PASS|FAIL)"
```

**Expected:** all PASS. If `TestActionAutoApproachQueuesWhenOutOfRange/door` fails, the
door-barrier carve-out is broken — revisit step 2 to ensure the dynamic entity check uses
`origin`, not `center`.

---

## Step 5 — Run full game+replay test suite

```bash
cd server
go test ./internal/game/... ./internal/replay/... -count=1 2>&1 | grep -E "^--- FAIL|^FAIL|^ok"
```

**Expected:** no failures. If any new failures appear (regressions), diagnose before
proceeding. If the previously-failing tests now pass but a formerly-passing test fails, that
is a regression introduced by this change and must be fixed before committing.

---

## Step 6 — Optionally tighten over-budget tick guards

With cleaner paths (wall-adjacent cells avoided), approach is typically faster. If the full
test suite passes and any test loop completes well within its budget (e.g., 5 ticks out of
150), tighten the guard to remove false safety margin. Only adjust if:
- The test passes consistently at the tighter budget (run 3 times to confirm).
- The change does not make a formerly-fast-path test flaky.

Files that may benefit: `combat_control_test.go` (loops at 150/300),
`game_test.go` (TestActionIntentAutoApproachAndAttack at 200).

Do NOT reduce if the test is using the full budget — leave headroom for variable dungeon
generation seeds.

---

## Step 7 — CI gates

```bash
make lint-determinism   # must pass
make validate-shared    # must pass
go vet ./...            # clean
```

No client changes → skip `make client-unit` and `make client-smoke` unless time allows.

---

## Step 8 — Commit

One commit per logical group:

1. **Production fix:**
   ```
   fix: buildBlockedFn uses cell-center probe for walls, origin for dynamic entities
   ```
   Files: `server/internal/game/sim.go`

2. **Golden update** (only if dungeon_teleporters.json changed):
   ```
   fix: regenerate dungeon_teleporters golden after pathfinding-cell-accuracy fix
   ```
   Files: `shared/golden/dungeon_teleporters.json`

3. **Test budget tightening** (only if step 6 found headroom):
   ```
   test: tighten approach tick budgets after cell-center pathfinding fix
   ```
   Files: `combat_control_test.go` and/or `game_test.go`

---

## Step 9 — Docs

1. Add v332 row to `docs/progress/slice-lifecycle.md`:
   ```
   | **v332** | `pathfinding-cell-accuracy` | Complete | spec | plan | as-built |
   ```
2. Write `docs/as-built/v332_pathfinding-cell-accuracy.md` — what the fix proved:
   - Which 9 tests went from FAIL to PASS.
   - Whether the dungeon_teleporters golden needed regeneration and why.
   - Whether tick budgets were tightened and by how much.
   - Confirm door approach behavior unaffected.
3. Update `PROGRESS.md`:
   - `Latest completed slice` → v332 pathfinding-cell-accuracy
   - `CI gate` → v332 `go test ./...` green on 2026-06-24
   - Remove navigation-stall open gap entry.

---

## Verification commands summary

```bash
# Step 3 — 9 target tests
cd server && go test ./internal/game/... ./internal/replay/... \
  -run "TestRangedProjectileGolden|TestProjectileBusyRejectsSecondFire|TestDirectionalRangedFreeShotHitsAndOmitsTargetID|TestRangedAutoApproachThenFire|TestRangedDummyDropsSeparatedLootItems|TestRangedBowLootRequiresMeleeReach|TestRangedBlockedLineAutoMovesUntilClearThenFires|TestActionIntentAutoApproachAndAttack|TestDungeonTeleportersReplayGolden" \
  -count=1 -v 2>&1 | grep -E "PASS|FAIL"

# Step 4 — critical regressions
cd server && go test ./internal/game/... \
  -run "TestActionAutoApproachQueuesWhenOutOfRange|TestCollisionIgnoresDeadMonster|TestCollisionBlocksWallAndAllowsRoute|TestGeneratedDungeonTargetsReachable|TestDungeonObstacleGeneration" \
  -count=1 -v 2>&1 | grep -E "PASS|FAIL"

# Step 5 — full suite
cd server && go test ./internal/game/... ./internal/replay/... -count=1 2>&1 | grep -E "FAIL|^ok"

# Step 7 — CI gates
make lint-determinism && make validate-shared && go vet ./...
```
