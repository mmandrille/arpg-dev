# v295 Plan - Water Obstacles Foundation

Status: Complete after focused autoloop verification; final batch `make ci` pending.
Goal: Add deterministic generated water blockers that route normal navigation around water and render as water in the client.
Architecture: Extend the existing generated obstacle layout with an additive `kind` field so walls and water share deterministic ordering, snapshots, and level-transition updates while preserving different behavior flags. Server movement/pathfinding treats water as blocked, but wall-only LOS/projectile checks keep using stone walls. Client rendering branches by kind to draw water as a flat surface and keep stone walls unchanged.
Tech stack: shared JSON rules/schemas, Go deterministic dungeon generation and navigation, Python protocol bot, Godot GDScript renderer/tests, SDD docs.

## Baseline and Shortcut Decision

Baseline: v294 full-CI residual stabilization is complete and green on `main`; this worktree was created from `eb6b8c09`.

Autoloop note: `PROGRESS.md` says a repo-wide review/refactor handoff is due after v294. This selected feature autoloop proceeds under the autoloop rule that review/refactor are recorded for the post-loop handoff, not inserted into the selected feature queue.

Asset/plugin decision: reject external assets, imported water art, shader plugins, and Godot addons. Borrow the existing generated obstacle pipeline, wall layout deltas, `WallRenderer`, `GroundWallFactory`, and bot wall-layout assertions. Use code-native water material/geometry.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Add data-owned water generation tuning. |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Validate water tuning. |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Allow additive layout obstacle kind. |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Allow additive layout obstacle kind in wall updates. |
| Modify | `shared/protocol/examples/session_snapshot.json` | Keep protocol examples valid with kind defaults/fields. |
| Modify | `shared/protocol/examples/state_delta.json` | Keep state-delta example valid with a water kind. |
| Modify | `shared/protocol/examples/state_delta_level_transition.json` | Keep level-transition layout example valid. |
| Modify | `server/internal/game/sim.go` | Add obstacle kind/behavior fields to generated blockers and layout views. |
| Modify/Create | `server/internal/game/dungeon_water.go` | Focused water generation helper and validation support. |
| Modify | `server/internal/game/dungeon_gen.go` | Wire water generation into non-boss generated floors without growing coordinator logic. |
| Modify/Create | `server/internal/game/obstacle_blocking.go` | Centralize obstacle-kind behavior for movement/projectile/LOS checks. |
| Modify | `server/internal/game/types.go` | Add optional `kind` to `WallView`. |
| Modify | `server/internal/game/pathfind_test.go` | Prove pathfinding detours around water blocker cells. |
| Modify | `server/internal/game/dungeon_obstacles_golden_test.go` | Include water in deterministic generated obstacle golden. |
| Modify | `shared/golden/dungeon_obstacles.json` | Update pinned generated obstacle proof. |
| Modify | `tools/bot/runtime_assertions.py` | Add obstacle-kind count assertions for protocol bot. |
| Modify | `tools/bot/scenarios/28_reachable_dungeon_obstacles.json` | Prove water generated and reachable route remains valid. |
| Modify | `client/scripts/wall_renderer.gd` | Render kind `water` as flat water surface, keep stone walls unchanged. |
| Modify | `client/scripts/ground_wall_factory.gd` | Add deterministic water material/texture helper. |
| Modify | `client/scripts/fog_of_war_overlay.gd` | Keep water out of tall wall fog occluders. |
| Modify | `client/tests/test_factories.gd` | Unit-proof water node rendering/material metadata. |
| Modify | `client/tests/test_fog_of_war_overlay.gd` | Unit-proof water does not generate fog shadows. |
| Add | `docs/as-built/v295_water-obstacles-foundation.md` | Record shipped behavior and proof. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v295 lifecycle row. |
| Modify | `PROGRESS.md` | Advance current status and keep review/refactor handoff due after selected autoloop batch. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd` was not changed.
- [x] `server/internal/game/dungeon_gen.go` stayed within the grandfathered +25 allowance.
- [x] `server/internal/game/sim.go` stayed within the grandfathered +25 allowance.
- [x] `tools/bot/runtime_assertions.py` remained under the 600-line target.
- [x] Other touched files stayed under their ratchet targets.
- [x] Did every touched grandfathered file stay at or below its baseline allowance or receive a documented net-non-positive extraction?

Decision:
- [x] Extract focused helper files (`dungeon_water.go`, `obstacle_blocking.go`) and avoid `main.gd` edits.
- Deferred extraction rationale: not used; focused helper extraction shipped.

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

- [x] Step 1.1: Add schema-backed `obstacle_generation.water` tuning with enabled flag, max attempts, target count, and rectangular size parameters.
- [x] Step 1.2: Add additive optional `kind` to layout wall objects with allowed values `wall` and `water`; default/omitted kind remains wall-compatible for existing data.
- [x] Step 1.3: Keep protocol examples valid and representative.

```bash
make validate-shared
```

## Task 2 - Server Water Generation and Blocking Semantics

Files:
- Create: `server/internal/game/dungeon_water.go`
- Create/Modify: `server/internal/game/obstacle_blocking.go`
- Modify: `server/internal/game/dungeon_gen.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`

- [x] Step 2.1: Extend internal obstacle records with kind/behavior metadata while preserving existing wall ordering and IDs for non-water blockers.
- [x] Step 2.2: Generate deterministic non-boss water obstacles after required targets are placed and before reachability validation accepts the floor.
- [x] Step 2.3: Make movement, player auto-navigation, monster pathfinding, travel arrival, corpse/loot placement blockers include water.
- [x] Step 2.4: Keep projectile and fog/LOS checks wall-only unless a blocker explicitly opts into those behaviors.
- [x] Step 2.5: Include `kind` in `WallView` only when useful for client/server proofs.

```bash
cd server && go test ./internal/game -run 'Path|DungeonObstacle|Water|GeneratedDungeon|LevelTransition|Collision'
```

## Task 3 - Golden and Protocol Bot Proof

Files:
- Modify: `server/internal/game/dungeon_obstacles_golden_test.go`
- Modify: `shared/golden/dungeon_obstacles.json`
- Modify: `tools/bot/runtime_assertions.py`
- Modify: `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`

- [x] Step 3.1: Update the generated-obstacle golden to include water counts/kinds and route-reachable targets.
- [x] Step 3.2: Add protocol bot assertions for water obstacle count/kind.
- [x] Step 3.3: Update the pinned reachable dungeon obstacle scenario to prove water appears and the player can still use the generated door route.

```bash
cd server && go test ./internal/game -run TestDungeonObstaclesGolden
ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=reachable_dungeon_obstacles ./scripts/bot_local.sh
```

## Task 4 - Godot Water Rendering and Client Proof

Files:
- Modify: `client/scripts/wall_renderer.gd`
- Modify: `client/scripts/ground_wall_factory.gd`
- Modify: `client/scripts/fog_of_war_overlay.gd`
- Modify: `client/tests/test_factories.gd`
- Modify: `client/tests/test_fog_of_war_overlay.gd`

- [x] Step 4.1: Render layout entries with kind `water` as low flat blue water surfaces rather than one-meter wall boxes.
- [x] Step 4.2: Preserve existing generated/perimeter/preset wall rendering, metadata, fog layout, and wall counts.
- [x] Step 4.3: Unit-proof normalized water layout entries, flat mesh rendering, and non-occluding fog behavior.

```bash
make client-unit
HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=dungeon_wall_rendering ./scripts/bot_visual.sh
```

## Task 5 - Lifecycle Docs and Focused Gates

Files:
- Modify: `docs/specs/v295_spec-water-obstacles-foundation.md`
- Modify: `docs/plans/v295_2026-06-19-water-obstacles-foundation.md`
- Add: `docs/as-built/v295_water-obstacles-foundation.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark spec and plan complete after proof.
- [x] Step 5.2: Write as-built summary with water behavior, proof commands, and deferred scope.
- [x] Step 5.3: Update progress/current status and lifecycle row. Keep review/refactor handoff due after the selected feature queue unless a hard stop occurs.

```bash
make maintainability
```

## Final Verification

Focused autoloop slice gates:

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'Path|DungeonObstacle|Water|GeneratedDungeon|LevelTransition|Collision'`
- [x] `cd server && go test ./internal/game -run TestDungeonObstaclesGolden`
- [x] `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=reachable_dungeon_obstacles ./scripts/bot_local.sh`
- [x] `make client-unit`
- [x] `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=dungeon_wall_rendering ./scripts/bot_visual.sh`
- [x] `make maintainability`

Autoloop batch gate:

- [ ] Final `make ci` is deferred to the selected batch after all requested world-detail/navigation slices are complete and committed.

## Deferred Scope

- Hazard holes/chasms, flying movement exceptions, barbarian leap over obstacles, non-rectangular obstacle visuals, tall LOS blockers, and wall/floor shader polish remain separate selected queue items.
- Water does not change combat, projectiles, LOS, loot, persistence, or economy in this slice.
