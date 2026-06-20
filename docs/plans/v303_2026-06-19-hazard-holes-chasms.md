# v303 Plan - Hazard Holes / Chasms

Status: Complete
Goal: Add deterministic non-walkable hole/chasm blockers that generated dungeon navigation routes around and the client renders distinctly from water and stone walls.
Architecture: Reuse the authoritative layout obstacle stream from v302 with a new additive `hole` kind. Server movement/pathfinding treats holes as grounded blockers, while projectile and fog/LOS checks remain wall-only. Client rendering branches by kind to draw holes as flat dark chasms.
Tech stack: shared JSON rules/schemas, Go deterministic dungeon generation/navigation, Python protocol bot, Godot GDScript renderer/tests, SDD docs.

## Baseline and Decisions

Baseline: v302 `water-obstacles-foundation` is committed on `codex/world-detail-navigation` as `d8a69cb8`.

Autoloop note: final batch `make ci` and the due review/refactor handoff remain deferred until the selected World Detail/Navigation queue completes.

Asset/plugin decision: reject external assets, imported chasm art, shader plugins, and Godot addons. Borrow the v302 obstacle-kind foundation, wall renderer branch pattern, deterministic texture factory, and reachable obstacle bot scenario.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Add data-owned hole generation tuning. |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Validate hole tuning. |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Allow additive layout obstacle kind `hole`. |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Allow additive layout obstacle kind `hole` in wall updates. |
| Modify | `shared/protocol/examples/session_snapshot.json` | Keep protocol example representative. |
| Modify | `shared/protocol/examples/state_delta.json` | Keep state-delta example representative. |
| Modify | `shared/protocol/examples/state_delta_level_transition.json` | Keep level-transition example representative. |
| Create | `server/internal/game/dungeon_holes.go` | Focused hole generation helper and validation support. |
| Modify | `server/internal/game/dungeon_water.go` | Share any reusable floor-feature validation/helpers only if it keeps code smaller. |
| Modify | `server/internal/game/obstacle_blocking.go` | Add `hole` kind behavior. |
| Modify | `server/internal/game/dungeon_gen.go` | Wire holes into non-boss generated floors after water and before final reachability validation. |
| Modify | `server/internal/game/rules.go` | Load and validate hole rules. |
| Modify | `server/internal/game/pathfind_test.go` | Prove pathfinding detours around hole cells. |
| Modify | `server/internal/game/fog_of_war_test.go` | Prove holes do not occlude fog line of sight. |
| Modify | `server/internal/game/dungeon_obstacles_golden_test.go` | Include hole counts/kinds in deterministic generated obstacle golden. |
| Modify | `shared/golden/dungeon_obstacles.json` | Update pinned generated obstacle proof. |
| Modify | `shared/golden/dungeon_obstacles.v0.schema.json` | Validate hole kind/count in the golden. |
| Modify | `tools/bot/scenarios/28_reachable_dungeon_obstacles.json` | Assert at least one hole while preserving reachable door route. |
| Modify | `client/scripts/wall_renderer.gd` | Render kind `hole` as flat chasm surface. |
| Modify | `client/scripts/ground_wall_factory.gd` | Add deterministic hole/chasm texture helper. |
| Modify | `client/scripts/fog_of_war_overlay.gd` | Keep non-wall floor features out of fog occluders. |
| Modify | `client/tests/test_factories.gd` | Unit-proof hole rendering/material metadata. |
| Modify | `client/tests/test_fog_of_war_overlay.gd` | Unit-proof holes do not generate fog shadows. |
| Add | `docs/as-built/v303_hazard-holes-chasms.md` | Record shipped behavior and proof. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v303 lifecycle row. |
| Modify | `PROGRESS.md` | Advance current status and keep review/refactor handoff due after selected autoloop batch. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files expected:
- [x] `server/internal/game/dungeon_gen.go` must stay within the grandfathered +25 allowance.
- [x] `server/internal/game/sim.go` should not need changes unless a layout view/schema detail is missed.
- [x] `client/scripts/main.gd` should not be changed.
- [x] `tools/bot/runtime_assertions.py` should remain under 600 lines; current kind filtering should already cover holes.
- [x] Other touched files must stay under their ratchet targets.

Decision:
- [x] Use focused helper file `dungeon_holes.go`; avoid growing large coordinators beyond narrow wiring.

Verification:

```bash
make maintainability
```

## Task 1 - Shared Rules and Protocol Shape

Files:
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Modify: `shared/protocol/session_snapshot.v8.schema.json`
- Modify: `shared/protocol/state_delta.v8.schema.json`
- Modify: `shared/protocol/examples/session_snapshot.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Modify: `shared/protocol/examples/state_delta_level_transition.json`

- [x] Step 1.1: Add schema-backed `obstacle_generation.holes` tuning with enabled flag, max attempts, target count, and rectangular size parameters.
- [x] Step 1.2: Extend optional layout `kind` enums to include `hole`; omitted kind remains wall-compatible.
- [x] Step 1.3: Keep protocol examples valid and representative.

```bash
make validate-shared
```

## Task 2 - Server Hole Generation and Blocking Semantics

Files:
- Create: `server/internal/game/dungeon_holes.go`
- Modify: `server/internal/game/dungeon_water.go`
- Modify: `server/internal/game/obstacle_blocking.go`
- Modify: `server/internal/game/dungeon_gen.go`
- Modify: `server/internal/game/rules.go`

- [x] Step 2.1: Add hole generation rules and validation.
- [x] Step 2.2: Generate deterministic non-boss holes after water and before final reachability validation.
- [x] Step 2.3: Make hole kind block movement through existing obstacle behavior helpers.
- [x] Step 2.4: Keep holes from blocking projectiles and fog/LOS.
- [x] Step 2.5: Preserve stable wall layout ordering and IDs for all generated obstacle kinds.

```bash
cd server && go test ./internal/game -run 'Path|DungeonObstacle|Hole|Water|GeneratedDungeon|LevelTransition|Collision'
```

## Task 3 - Golden and Protocol Bot Proof

Files:
- Modify: `server/internal/game/dungeon_obstacles_golden_test.go`
- Modify: `shared/golden/dungeon_obstacles.json`
- Modify: `shared/golden/dungeon_obstacles.v0.schema.json`
- Modify: `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`

- [x] Step 3.1: Update generated-obstacle golden to include hole counts/kinds and reachable targets.
- [x] Step 3.2: Assert hole count/kind in the pinned reachable dungeon obstacle protocol scenario.
- [x] Step 3.3: Keep the generated-door route proof green with water and holes both present.

```bash
cd server && go test ./internal/game -run TestDungeonObstaclesGolden
ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=reachable_dungeon_obstacles ./scripts/bot_local.sh
```

## Task 4 - Godot Hole Rendering and Client Proof

Files:
- Modify: `client/scripts/wall_renderer.gd`
- Modify: `client/scripts/ground_wall_factory.gd`
- Modify: `client/scripts/fog_of_war_overlay.gd`
- Modify: `client/tests/test_factories.gd`
- Modify: `client/tests/test_fog_of_war_overlay.gd`

- [x] Step 4.1: Render kind `hole` as a low flat dark chasm surface, distinct from water and walls.
- [x] Step 4.2: Preserve existing wall and water rendering metadata.
- [x] Step 4.3: Unit-proof hole rendering and non-occluding fog behavior.

```bash
make client-unit
HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=dungeon_wall_rendering ./scripts/bot_visual.sh
```

## Task 5 - Lifecycle Docs and Focused Gates

Files:
- Modify: `docs/specs/v303_spec-hazard-holes-chasms.md`
- Modify: `docs/plans/v303_2026-06-19-hazard-holes-chasms.md`
- Add: `docs/as-built/v303_hazard-holes-chasms.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark spec and plan complete after proof.
- [x] Step 5.2: Write as-built summary with hole behavior, proof commands, and deferred scope.
- [x] Step 5.3: Update progress/current status and lifecycle row. Keep review/refactor handoff due after the selected queue unless a hard stop occurs.

```bash
make maintainability
```

## Final Verification

Focused autoloop slice gates:

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'Path|DungeonObstacle|Hole|Water|GeneratedDungeon|LevelTransition|Collision'`
- [x] `cd server && go test ./internal/game -run TestDungeonObstaclesGolden`
- [x] `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=reachable_dungeon_obstacles ./scripts/bot_local.sh`
- [x] `make client-unit`
- [x] `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=dungeon_wall_rendering ./scripts/bot_visual.sh`
- [x] `make maintainability`

Autoloop batch gate:

- [ ] Final `make ci` is deferred to the selected batch after all requested world-detail/navigation slices are complete and committed.

## Deferred Scope

- Barbarian leap, flying movement exceptions, obstacle variety, tall LOS blockers, wall/floor shader polish, falling/damage, and bridge/recovery mechanics remain separate or future slices.
- Holes do not change combat, projectiles, LOS, loot, persistence, or economy in this slice.
