# v461 Plan — Dungeon Surface Kit

Status: Ready for implementation
Goal: Refresh dungeon base surfaces and remove rubble/debris clutter so floors, walls, and water
read more clearly.
Architecture: This slice keeps the existing authoritative split intact. Shared dungeon-generation
rules own depth-band palette data and the generated obstacle mix; the server only stops selecting
`rubble` during obstacle generation, while the Godot client refreshes procedural materials for
ground, walls, water, and holes from the same depth-band palettes. Existing torch, fog, and wall
rendering systems remain in place and are validated against the updated surfaces rather than
replaced.
Tech stack: shared JSON rules/schema, Go server rule loading/tests, Godot GDScript client, client
bot scenario, docs.

## Baseline and shortcut decision

Baseline: v460 `leveled-potions`

Reuse:
- Adopt: `GroundWallFactory`, `WallRenderer`, biome palette loading, existing wall/floor/water
  procedural texture generation.
- Borrow: `test_factories.gd`, obstacle-variety tests, existing dungeon client-bot scenario
  patterns.
- Reject: external texture packs, shader packs, trim/debris decorator sets, and Godot addons.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Add a middle biome band and remove rubble from generated obstacle weights |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Keep schema aligned with updated palette and obstacle data |
| Modify | `server/internal/game/dungeon_obstacle_variety.go` | Remove rubble from generated solid-kind selection |
| Modify | `server/internal/game/dungeon_obstacle_variety_test.go` | Prove generated solid-kind selection excludes rubble |
| Modify | `client/scripts/ground_wall_factory.gd` | Refresh palette-driven wall/floor/water/hole material generation |
| Modify | `client/scripts/wall_renderer.gd` | Remove rubble rendering path and strengthen water/hole surface presentation |
| Modify | `client/tests/test_factories.gd` | Assert biome differentiation and no rubble visual expectation |
| Add | `tools/bot/scenarios/client/99_dungeon_surface_kit.json` | Focused visual proof for refreshed dungeon surfaces |
| Modify | `tools/bot/scenarios/101_obstacle_variety_pack.json` | Remove rubble expectation from gameplay-side obstacle proof |
| Modify | `PROGRESS.md` | Advance current status and note deferred follow-ups |
| Modify | `docs/progress/slice-lifecycle.md` | Add v461 completion row |
| Add | `docs/as-built/v461_dungeon-surface-kit.md` | Record what shipped and what remains deferred |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd`
- [ ] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [ ] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] No over-limit hotspot extraction is required for this slice because the planned files are
  focused presentation/rules modules below the ratchet ceiling.

Verification:
```bash
make maintainability
```

## Task 1 — Shared dungeon rules cleanup

Files:
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`

- [ ] Step 1.1: Add a distinct mid-depth biome palette between shallow and deep.
```bash
make validate-shared
```

- [ ] Step 1.2: Remove `rubble` from generated obstacle kind weights while keeping the rules valid.
```bash
make validate-shared
```

## Task 2 — Server obstacle variety alignment

Files:
- Modify: `server/internal/game/dungeon_obstacle_variety.go`
- Modify: `server/internal/game/dungeon_obstacle_variety_test.go`

- [ ] Step 2.1: Remove rubble from generated solid-kind selection logic.
```bash
cd server && go test ./internal/game -run TestSolidObstacleKindWeights -count=1
```

- [ ] Step 2.2: Keep validation/rule-loading coverage green for the updated weight shape.
```bash
cd server && go test ./internal/game -run TestLoadRules -count=1
```

## Task 3 — Client dungeon surface refresh

Files:
- Modify: `client/scripts/ground_wall_factory.gd`
- Modify: `client/scripts/wall_renderer.gd`
- Modify: `client/tests/test_factories.gd`

- [ ] Step 3.1: Refresh floor and wall material generation so shallow, mid, and deep bands read as
  clearly different stone spaces.
```bash
godot --headless --path client --script res://tests/test_factories.gd
```

- [ ] Step 3.2: Improve water and hole contrast against the floor using the existing plane and
  overlay approach.
```bash
godot --headless --path client --script res://tests/test_factories.gd
```

- [ ] Step 3.3: Remove the rubble rendering path and replace rubble expectations in focused client
  tests with the new simpler obstacle set.
```bash
make client-unit
```

## Task 4 — Bot scenarios

Files:
- Add: `tools/bot/scenarios/client/99_dungeon_surface_kit.json`
- Modify: `tools/bot/scenarios/101_obstacle_variety_pack.json`

- [ ] Step 4.1: Add a focused client-bot dungeon surface scenario that descends one floor and
  asserts refreshed dungeon surface debug state.
```bash
HEADLESS=1 make bot-client scenario=99_dungeon_surface_kit
```

- [ ] Step 4.2: Update obstacle-variety proof to stop expecting rubble generation.
```bash
make bot scenario=101_obstacle_variety_pack
```

## Task 5 — Lifecycle docs and close-out

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v461_dungeon-surface-kit.md`
- Modify: `docs/specs/v461_spec-dungeon-surface-kit.md`

- [ ] Step 5.1: Mark the spec/plan/as-built/progress trail complete for v461.
```bash
make ci
```

## Final verification

- [ ] `make validate-shared`
- [ ] `cd server && go test ./internal/game -run TestSolidObstacleKindWeights -count=1`
- [ ] `godot --headless --path client --script res://tests/test_factories.gd`
- [ ] `make client-unit`
- [ ] `HEADLESS=1 make bot-client scenario=99_dungeon_surface_kit`
- [ ] `make bot scenario=101_obstacle_variety_pack`
- [ ] `make maintainability`
- [ ] `make ci`

## Deferred scope

- Dungeon props/architectural set dressing remain deferred.
- Boss-floor-specific surface art remains deferred.
- Audio and post-processing mood passes remain deferred.
