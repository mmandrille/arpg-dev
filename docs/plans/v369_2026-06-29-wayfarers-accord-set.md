# v369 Plan — Wayfarer's Accord Set

Status: Ready for implementation
Date: 2026-06-29
Spec: [`docs/specs/v369_spec-wayfarers-accord-set.md`](../specs/v369_spec-wayfarers-accord-set.md)

## Architecture

Set items remain fixed rolled payloads sourced from `set_items.v0.json`. The new package reuses
existing rule validation, unique chest exposure, elite special drops, item payload construction, and
equipped set bonus aggregation. Pieces avoid weapon slots so all four classes can equip five pieces
alongside their class weapon.

## File map

| Action | Path | Notes |
|--------|------|-------|
| Modify | `shared/rules/set_items.v0.json` | Add `wayfarers_accord` five-piece catalog. |
| Modify | `shared/rules/treasure_classes.v0.json` | Add new piece ids to `elite_objective_special_tc_1`. |
| Modify | `server/internal/game/unique_chest_test.go` | Add `TestWayfarersAccordSetPayloadsAndBonuses`. |
| Add | `tools/bot/scenarios/84_wayfarers_accord_set.json` | Extended chest take proof. |
| Add | `docs/as-built/v369_wayfarers-accord-set.md` | Slice summary. |
| Modify | `docs/progress/slice-lifecycle.md`, `PROGRESS.md` | Close-out metadata. |

## Tasks

### 1. Shared catalog

- [x] Step 1.1: Add `wayfarers_accord` with head/chest/gloves/boots/amulet pieces and universal bonuses.
- [x] Step 1.2: Register all five piece ids in `elite_objective_special_tc_1`.
- [x] Verify: `make validate-shared`

### 2. Server proof

- [x] Step 2.1: Add focused Go test for payload, 2/3/4/full bonuses, and `all_skills` rank lift.
- [x] Verify: `cd server && go test ./internal/game -run 'TestWayfarersAccord|TestUniqueTestChest|TestSetItem' -count=1`

### 3. Bot proof

- [x] Step 3.1: Add extended scenario taking `Wayfarer's Accord Pendant` from the unique chest.
- [x] Verify: `make bot scenario=84_wayfarers_accord_set.json`

### 4. Client sanity

- [x] Step 4.1: Confirm set collection panel still loads enabled sets from rules without code changes.
- [x] Verify: `make client-unit`

### 5. Close-out

- [x] Step 5.1: Write as-built, update lifecycle and `PROGRESS.md`.
- [x] Verify: focused commands above (batch `make ci` deferred to autoloop post-loop handoff).
