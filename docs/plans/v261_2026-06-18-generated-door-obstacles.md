# v261 Plan - Generated Door Obstacles

Status: Complete
Goal: Generate deterministic closed wooden doors in eligible generated dungeon wall openings.
Architecture: Store door tuning inside `obstacle_generation`, select eligible horizontal generated
wall segments after reachable obstacles are found, split those wall segments into two pieces, and
emit normal generated `wooden_door` interactables. Keep live barrier behavior owned by existing
interactable rules; generated reachability checks remain wall-only so closed doors do not invalidate
the floor at generation time.
Tech stack: Shared JSON/schema, Go dungeon generation/simulation, existing bot scenario.

## Baseline and Shortcut Decision

Builds on v40/v252/v254 generated dungeon obstacles and existing `wooden_door` interactables. The
slice does not introduce a door art pipeline or a generalized oriented-barrier system; it restricts
door openings to horizontal wall segments because the current door barrier is horizontal.

Asset/plugin decision: reject external art, plugins, and imported door assets. Borrow the existing
`wooden_door` interactable, barrier, action, and client presentation.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Add generated-door tuning |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Validate generated-door tuning |
| Modify | `server/internal/game/dungeon_gen.go` | Minimal call-site/type edits for generated doors |
| Add | `server/internal/game/dungeon_doors.go` | Door placement and wall splitting helpers |
| Add | `server/internal/game/dungeon_generated_types.go` | Move generated type declarations out of hotspot file |
| Add | `server/internal/game/dungeon_door_rules.go` | Door-generation rule validation |
| Modify | `server/internal/game/dungeon_population.go` | Populate generated door interactables |
| Modify | `server/internal/game/rules.go` | Add door rules field and one validation call |
| Modify | `server/internal/game/game_test.go` | Semantic generated-door tests |
| Modify | `server/internal/game/dungeon_obstacles_golden_test.go` | Include generated doors in obstacle golden |
| Modify | `shared/golden/dungeon_obstacles.json` | Update pinned obstacle proof |
| Modify | `tools/bot/scenarios/28_reachable_dungeon_obstacles.json` | Open a generated door in the obstacle bot proof |
| Modify | `docs/specs/v261_spec-generated-door-obstacles.md` | Mark complete during close-out |
| Modify | `docs/progress/slice-lifecycle.md` | Add v261 lifecycle row |
| Add | `docs/as-built/v261_generated-door-obstacles.md` | Record shipped behavior and proof |
| Modify | `PROGRESS.md` | Update current status and next selected autoloop item |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines, and grandfathered files stay under their
ratchet allowance.

Hotspot / over-limit files touched:
- [x] `server/internal/game/dungeon_gen.go`
- [x] `server/internal/game/rules.go`
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Move generated-type declarations out of `dungeon_gen.go` before adding door logic.
- [x] Put door placement and validation helpers in new focused files.
- [x] Keep `rules.go` to one field and one validation call.

Verification:
```bash
make maintainability
```

## Task 1 - Shared Door Tuning

Files:
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Modify: `server/internal/game/rules.go`
- Add: `server/internal/game/dungeon_door_rules.go`

- [x] Step 1.1: Add `obstacle_generation.doors` with `enabled`, `interactable_def_id`,
  `max_count`, `min_wall_length`, and `gap_width`.
- [x] Step 1.2: Validate that enabled doors use `wooden_door`, positive dimensions, and known
  interactable data.
- [x] Step 1.3: Keep all tuning values in shared data, not code constants.

```bash
make validate-shared
cd server && go test ./internal/game -run TestLoadRules
```

## Task 2 - Door Generation and Population

Files:
- Modify: `server/internal/game/dungeon_gen.go`
- Add: `server/internal/game/dungeon_generated_types.go`
- Add: `server/internal/game/dungeon_doors.go`
- Modify: `server/internal/game/dungeon_population.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Move generated type declarations out of `dungeon_gen.go`.
- [x] Step 2.2: Add generated-door output to `generatedDungeonLevel`.
- [x] Step 2.3: Split eligible horizontal generated walls into left/right wall pieces with a door
  gap.
- [x] Step 2.4: Populate generated doors as normal closed interactables.
- [x] Step 2.5: Add semantic tests for determinism, closed door state, wall split geometry, and
  boss-floor exclusion.

```bash
cd server && go test ./internal/game
```

## Task 3 - Bot and Lifecycle Proof

Files:
- Modify: `server/internal/game/dungeon_obstacles_golden_test.go`
- Modify: `shared/golden/dungeon_obstacles.json`
- Modify: `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`
- Modify: `docs/specs/v261_spec-generated-door-obstacles.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v261_generated-door-obstacles.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Extend the obstacle golden with generated door expectations.
- [x] Step 3.2: Extend the reachable dungeon obstacles bot scenario to open a generated door.
- [x] Step 3.3: Close spec/as-built/progress/lifecycle docs and leave doorway LOS plus quest path
  marker as remaining selected autoloop scope.

```bash
make bot scenario=reachable_dungeon_obstacles
make maintainability
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game`
- [x] `make bot scenario=reachable_dungeon_obstacles`
- [x] `make maintainability`
- [ ] Autoloop final batch gate: `make ci`

Manual visual proof, if desired:

```bash
make bot-visual scenario=14_dungeon_wall_rendering
```
