# v303 Spec - Hazard Holes / Chasms

Status: Complete
Date: 2026-06-19
Codename: hazard-holes-chasms

## Purpose

Add deterministic holes/chasms to generated dungeon floors as non-walkable floor hazards. Holes
should be owned by shared dungeon-generation rules, appear in authoritative layout snapshots and
level-transition deltas with a distinct obstacle kind, block normal grounded navigation, render
clearly as holes in the Godot client, and keep required generated dungeon targets reachable.

This slice builds directly on v302's obstacle-kind foundation and keeps holes as terrain blockers
only so later slices can add barbarian leap and flying navigation exceptions.

## Non-goals

- No falling, damage, death, knockback, slow, bridge, ladder, rescue, teleport, or recovery logic.
- No barbarian leap, flying-navigation exception, water interaction, obstacle variety art pack, or
  line-of-sight blocker behavior in this slice.
- No boss-floor hole generation, room/corridor PCG, rotated/polygon holes, destructible holes,
  production hole art, imported assets, Godot addons, or shader plugins.
- No combat, projectile, fog/visibility, loot/economy, persistence, replay-format, monster stat,
  skill, or item tuning changes.

## Acceptance Criteria

- `shared/rules/dungeon_generation.v0.json` owns schema-backed hole/chasm generation tuning under
  the existing obstacle-generation area.
- Shared protocol schemas allow layout entries with kind `hole`; omitted kind remains a normal
  wall.
- Generated non-boss dungeon floors can include deterministic rectangular hole/chasm layout entries
  with stable ordering and IDs in snapshots and wall-layout deltas.
- Player movement, auto-navigation, monster/companion pathfinding, travel arrival, corpse placement,
  and loot placement treat holes as blocked for normal grounded movement.
- Dungeon reachability validation treats holes as blocked and proves stairs, generated doors,
  chests, monsters, and loot remain reachable.
- Projectiles and fog/LOS continue to ignore holes unless a future blocker kind opts into those
  behaviors.
- The Godot client renders holes/chasms as flat, dark, readable floor hazards rather than wall boxes
  or water.
- Protocol bot proof descends into a pinned generated floor, observes at least one `hole` layout
  entry, and still completes the reachable generated-door route.
- Client unit proof verifies hole layout entries normalize/render as hole surfaces and do not become
  fog occluders.
- Existing water obstacle, generated wall/door, pathfinding, fog, and collision tests remain green.

## Scope and Likely Files

- Shared rules/schema:
  - `shared/rules/dungeon_generation.v0.json`
  - `shared/rules/dungeon_generation.v0.schema.json`
- Shared protocol/schema/examples:
  - `shared/protocol/session_snapshot.v8.schema.json`
  - `shared/protocol/state_delta.v8.schema.json`
  - `shared/protocol/examples/session_snapshot.json`
  - `shared/protocol/examples/state_delta.json`
  - `shared/protocol/examples/state_delta_level_transition.json`
- Server generation/navigation:
  - `server/internal/game/dungeon_gen.go`
  - `server/internal/game/dungeon_water.go` if shared floor-feature validation is factored
  - `server/internal/game/dungeon_holes.go`
  - `server/internal/game/obstacle_blocking.go`
  - `server/internal/game/rules.go`
  - `server/internal/game/sim.go`
  - `server/internal/game/types.go`
- Client presentation:
  - `client/scripts/wall_renderer.gd`
  - `client/scripts/ground_wall_factory.gd`
  - `client/scripts/fog_of_war_overlay.gd`
- Tests and bot:
  - `server/internal/game/pathfind_test.go`
  - `server/internal/game/dungeon_obstacles_golden_test.go`
  - `server/internal/game/fog_of_war_test.go`
  - `shared/golden/dungeon_obstacles.json`
  - `shared/golden/dungeon_obstacles.v0.schema.json`
  - `tools/bot/runtime_assertions.py`
  - `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`
  - `client/tests/test_factories.gd`
  - `client/tests/test_fog_of_war_overlay.gd`
- Docs:
  - `docs/plans/v303_2026-06-19-hazard-holes-chasms.md`
  - `docs/as-built/v303_hazard-holes-chasms.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external assets, imported hole/chasm art, shader plugins, and Godot
addons. Borrow the v302 obstacle-kind layout, deterministic generator validation, and code-native
texture/material approach.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'Path|DungeonObstacle|Hole|Water|GeneratedDungeon|LevelTransition|Collision'`
- `cd server && go test ./internal/game -run TestDungeonObstaclesGolden`
- `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=reachable_dungeon_obstacles ./scripts/bot_local.sh`
- `make client-unit`
- `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=dungeon_wall_rendering ./scripts/bot_visual.sh`
- `make maintainability`

Manual visual proof, if desired:

```bash
make bot-visual scenario=dungeon_wall_rendering
```

## Open Questions and Risks

- No required questions for this run. Defaults: holes generate only on non-boss dungeon floors,
  block normal grounded movement, do not block LOS/projectiles, and use code-native client
  material/geometry.
- Risk: water and hole placement together can isolate targets. The generator must validate final
  reachability with both feature kinds included as blocked geometry.
- Risk: `server/internal/game/sim.go` and client renderers are shared surfaces. Keep edits narrow
  and use focused helper files where possible.
