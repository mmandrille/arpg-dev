# v304 Plan - Flying Navigation Trait

Status: Complete
Goal: Add a rules-owned flying monster navigation trait so bats can cross water and holes while grounded monsters still route around them.
Architecture: Keep the trait server-authoritative and data-owned by monster rules. Add a narrow helper that decides whether a monster definition treats an obstacle kind as blocking, then use it from monster pathfinding, cached-goal validation, and final movement resolution. Add optional preset wall `kind` support to create a compact bot lab without changing protocol schemas.
Tech stack: shared JSON rules/schemas, Go deterministic sim navigation, Python protocol bot scenario, small Godot wall-renderer adapter, SDD docs.

## Baseline and Decisions

Baseline: v303 `hazard-holes-chasms` is committed on `codex/world-detail-navigation` as `0bfebc71`.

Autoloop note: final batch `make ci` and the due review/refactor handoff remain deferred until the selected World Detail/Navigation queue completes.

Asset/plugin decision: reject external assets, imported bat/flying VFX, shaders, and Godot addons. Borrow existing bat presentation, the v302/v303 obstacle-kind layout/rendering path, and existing compact world/bot lab patterns.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/monsters.v0.json` | Mark `dungeon_bat` as flying. |
| Modify | `shared/rules/monsters.v0.schema.json` | Validate optional navigation trait. |
| Modify | `shared/rules/worlds.v0.json` | Add compact flying navigation lab with water/hole preset walls. |
| Modify | `shared/rules/worlds.v0.schema.json` | Allow optional wall `kind` for preset labs. |
| Modify | `server/internal/game/rules.go` | Load/validate monster navigation trait and preset wall kinds. |
| Add | `server/internal/game/monster_navigation_traits.go` | Trait constants and obstacle-kind blocking helper. |
| Add | `server/internal/game/monster_navigation_traits_test.go` | Focused flying-vs-grounded path/collision proof. |
| Modify | `server/internal/game/sim.go` | Use trait-aware monster blocking and preserve preset wall kinds. |
| Modify | `client/scripts/wall_renderer.gd` | Propagate preset wall `kind` into existing render path. |
| Modify | `client/tests/test_factories.gd` | Unit-proof preset world water/hole wall kinds render through `render_world_walls` if practical. |
| Add | `tools/bot/scenarios/99_flying_navigation_trait.json` | End-to-end bat movement proof. |
| Modify | `docs/specs/v304_spec-flying-navigation-trait.md` | Mark complete after proof. |
| Modify | `docs/plans/v304_2026-06-19-flying-navigation-trait.md` | Track execution checkboxes. |
| Add | `docs/as-built/v304_flying-navigation-trait.md` | Record shipped behavior and proof. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v304 lifecycle row. |
| Modify | `PROGRESS.md` | Advance current status and keep autoloop review/refactor handoff due. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files expected:
- [x] `server/internal/game/sim.go` is grandfathered; only narrow call-site changes are allowed.
- [x] `server/internal/game/rules.go` is grandfathered; schema validation additions must stay minimal.
- [x] `client/scripts/main.gd` should not be changed.
- [x] `tools/bot/runtime_assertions.py` should not need changes.
- [x] Other touched files must stay under their ratchet targets.

Decision:
- [x] Add focused helper/test files for navigation traits instead of growing coordinators.

Verification:

```bash
make maintainability
```

## Task 1 - Shared Rules and Preset Lab Shape

Files:
- Modify: `shared/rules/monsters.v0.json`
- Modify: `shared/rules/monsters.v0.schema.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `shared/rules/worlds.v0.schema.json`

- [x] Step 1.1: Add optional monster `navigation_trait` enum with default grounded semantics.
- [x] Step 1.2: Mark `dungeon_bat` as `flying`.
- [x] Step 1.3: Allow preset wall entities to carry optional `kind` values `wall`, `water`, or `hole`.
- [x] Step 1.4: Add a compact `flying_navigation_lab` with bat, grounded monster, and water/hole obstacle strips.

```bash
make validate-shared
```

## Task 2 - Server Trait Semantics

Files:
- Modify: `server/internal/game/rules.go`
- Add: `server/internal/game/monster_navigation_traits.go`
- Modify: `server/internal/game/sim.go`

- [x] Step 2.1: Add server constants/methods for grounded and flying navigation traits.
- [x] Step 2.2: Validate unknown navigation traits and preset wall kinds during rules load.
- [x] Step 2.3: Preserve preset wall kinds in `wallObstacle` values.
- [x] Step 2.4: Make monster path planning, cached-goal validation, and final movement resolution use trait-aware obstacle blocking through the shared monster blocked-cell function.
- [x] Step 2.5: Keep normal walls, closed doors, players, and living entities blocking for every monster.

```bash
cd server && go test ./internal/game -run 'FlyingNavigationTrait|FlyingNavigationLab|Path'
```

## Task 3 - Focused Go Proof

Files:
- Add: `server/internal/game/monster_navigation_traits_test.go`
- Modify as needed: `server/internal/game/game_test.go` only if existing helpers are reused directly.

- [x] Step 3.1: Prove flying and grounded monsters see different blocked cells for water and holes.
- [x] Step 3.2: Prove flying pathing can cross water/hole while grounded pathing detours.
- [x] Step 3.3: Prove flying monsters still treat normal walls as blocking.

```bash
cd server && go test ./internal/game -run 'FlyingNavigation|MonsterChase|Path'
```

## Task 4 - Bot and Client Lab Proof

Files:
- Add: `tools/bot/scenarios/99_flying_navigation_trait.json`
- Modify: `client/scripts/wall_renderer.gd`
- Modify: `client/tests/test_factories.gd`

- [x] Step 4.1: Add a protocol bot scenario that waits for the bat to move across the lab and remain near the player.
- [x] Step 4.2: Preserve existing wall rendering while allowing preset water/hole kinds to flow through `render_world_walls`.
- [x] Step 4.3: Unit-proof preset water/hole rendering if the client factory can load the lab deterministically.

```bash
ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=flying_navigation_trait ./scripts/bot_local.sh
make client-unit
HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=flying_navigation_trait ./scripts/bot_visual.sh
```

## Task 5 - Lifecycle Docs and Focused Gates

Files:
- Modify: `docs/specs/v304_spec-flying-navigation-trait.md`
- Modify: `docs/plans/v304_2026-06-19-flying-navigation-trait.md`
- Add: `docs/as-built/v304_flying-navigation-trait.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark spec and plan complete after proof.
- [x] Step 5.2: Write as-built summary with trait behavior, proof commands, and deferred scope.
- [x] Step 5.3: Update progress/current status and lifecycle row. Keep review/refactor handoff due after the selected queue unless a hard stop occurs.

```bash
make maintainability
```

## Final Verification

Focused autoloop slice gates:

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'FlyingNavigationTrait|FlyingNavigationLab|Path'`
- [x] `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=flying_navigation_trait ./scripts/bot_local.sh`
- [x] `make client-unit`
- [x] `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=flying_navigation_trait ./scripts/bot_visual.sh`
- [x] `make maintainability`

Autoloop batch gate:

- [x] Final `make ci` is deferred to the selected batch after all requested world-detail/navigation slices are complete and committed.

## Deferred Scope

- Barbarian leap over water/holes, player movement exceptions, obstacle variety, tall LOS blockers, wall/floor shader polish, falling/damage, and bridge/recovery mechanics remain separate or future slices.
- Flying trait does not change projectiles, LOS, fog, loot/corpse placement, player navigation, companion navigation, persistence, or economy in this slice.
