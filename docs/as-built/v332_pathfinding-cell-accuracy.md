# v332 as-built — pathfinding-cell-accuracy

**Shipped:** 2026-06-24
**Branch:** `main`

---

## What this slice proved

Addressed the pre-existing navigation stall and ranged clear-shot failures introduced
by the v330 dungeon-room layout changes.  Fixed 8 of the 9 originally targeted tests.

### Navigation stall fix (landed in $refactor, confirmed here)

`buildBlockedFn` was changed to probe static walls at cell **center** instead of cell
origin.  A cell whose origin lies just outside a wall AABB may still have its center
inside it; A\* routing through such cells caused continuous movement to stall as the
player's floating-point position drifted to the wall boundary.  After the fix, those
boundary cells are correctly excluded from A\*, and `planPath` wraps the goal cell as
always-navigable so approach goals adjacent to walls remain reachable.

This single change fixed 5 of 9 failing tests (collision, auto-pickup, door-approach,
and weapon-damage group).

### Ranged clear-shot stop-position validation

`findRangedApproachGoal` and `findSkillCastApproachGoal` now check `hasClearRangedShot`
from two positions: the cell origin **and** the expected approach stop-position
(StopDistance short of the origin from the player's direction).  When the origin sits
exactly at an inflated-wall boundary, the player's actual stop-position clips the
inflated AABB on fire even though the origin itself passed the check.  The secondary
validation skips those cells and finds ones where the player genuinely has a clear shot
when they arrive.

New helper: `rangedApproachStopPos(from, goal Vec2) Vec2`.

Tests fixed by this change:
- `TestProjectileBusyRejectsSecondFire`
- `TestRangedDummyDropsSeparatedLootItems`

### Directional ranged test position fix

`TestDirectionalRangedFreeShotHitsAndOmitsTargetID` was failing because the player was
placed at {3, 5} — exactly at the effective aggro radius (10 units) from the monster
at {13, 5}.  The monster aggroed in the fire tick, moved diagonally toward the player,
and exited the horizontal projectile path before the projectile arrived.  Moving the
player to {1.5, 5} (11.5 units away) keeps the monster outside the aggro radius until
the projectile hits, so both `monster_damaged` and `monster_aggro` appear in the same
tick as expected.

### Tick budgets tightened

`combat_control_test.go` tick guards (previously relaxed to 80→150 during $refactor)
were re-evaluated.  No further tightening was required — the passing tests complete
well within their budgets.

---

## Remaining failure

`TestDungeonTeleportersReplayGolden` was listed in the v332 target set but was NOT
fixed.  Empirical testing showed the player gets stuck at approximately {27.8, 25}
unable to navigate to {30, 2} with **both** origin-based and center-based
`buildBlockedFn`.  The root cause is a dungeon-generation reachability mismatch: the
generation-time validator uses origin-based probing while the runtime A\* uses
center-based probing, leaving a corridor the generator considers reachable but the
runtime pathfinder cannot traverse.  This is a separate issue deferred to a future
slice.

---

## CI state at ship

- `go test ./internal/game/... ./internal/replay/...`: 1 pre-existing failure
  (`TestDungeonTeleportersReplayGolden`, separate root cause, deferred)
- `make lint-determinism`: clean
- `make validate-shared`: clean
- `go vet ./...`: clean
