# v295 Spec - Water Obstacles Foundation

Status: Complete after focused autoloop verification; final batch `make ci` pending.
Date: 2026-06-19
Codename: water-obstacles-foundation

## Purpose

Add deterministic water obstacles to generated dungeon floors. Water should be authored by shared
dungeon-generation rules, block normal walking and auto-navigation, render as water in the Godot
client, and keep generated floor targets reachable by routing around it.

This slice establishes an obstacle-kind foundation so later slices can add holes, flying movement,
barbarian leap exceptions, richer obstacle art, and tall line-of-sight blockers without treating
every blocker as a stone wall.

## Non-goals

- No swimming, bridge building, water combat effects, resource costs, damage-over-time, slowdown,
  knockback, or loot interactions.
- No barbarian leap, flying-navigation exception, chasm/hole, obstacle variety pack, or
  line-of-sight blocker behavior in this slice.
- No boss-floor water generation, room/corridor PCG, rotated/polygon water shapes, animated water,
  production water art, imported assets, Godot addons, or shader plugins.
- No persistence, database, replay-format, loot, monster, skill, economy, or combat tuning change.
- No water-specific fog/visibility occlusion; water blocks walking but does not become a tall wall
  or LOS blocker.

## Acceptance Criteria

- `shared/rules/dungeon_generation.v0.json` owns schema-backed water generation tuning under the
  existing obstacle-generation area.
- Generated non-boss dungeon floors can include deterministic rectangular water obstacle tiles or
  strips with stable IDs/order in snapshots and wall/layout deltas.
- The server distinguishes obstacle kind from source so water can block normal movement and
  pathfinding without being treated as a stone wall for wall-only presentation or LOS semantics.
- Player auto-navigation and monster pathfinding treat water as blocked for normal grounded
  movement and route around it when a reachable path exists.
- Dungeon reachability validation treats water as blocked and proves stairs, generated doors,
  chests, monsters, and other generated targets remain reachable.
- The Godot client renders water blockers as flat, readable water surfaces instead of wall boxes,
  while existing perimeter/generated wall rendering remains unchanged.
- Protocol bot proof descends into a pinned generated floor, observes at least one water obstacle,
  and moves through a route that must go around the water.
- Client bot or unit proof verifies water layout entries render as water and appear in bot debug
  state separately from stone wall counts.
- Existing generated-wall, generated-door, fog, collision, and pathfinding tests remain green.

## Scope and Likely Files

- Shared rules/schema:
  - `shared/rules/dungeon_generation.v0.json`
  - `shared/rules/dungeon_generation.v0.schema.json`
- Shared protocol/schema/examples:
  - `shared/protocol/session_snapshot.v8.schema.json`
  - `shared/protocol/state_delta.v8.schema.json`
  - `shared/protocol/examples/session_snapshot.json`
  - `shared/protocol/examples/state_delta_level_transition.json`
- Server generation/navigation:
  - `server/internal/game/dungeon_gen.go`
  - `server/internal/game/dungeon_generated_types.go`
  - `server/internal/game/dungeon_population.go`
  - `server/internal/game/rules.go`
  - `server/internal/game/sim.go`
  - focused helper files if needed to keep hotspots within ratchets
- Client presentation:
  - `client/scripts/wall_renderer.gd` or a focused obstacle renderer/helper
  - `client/scripts/ground_wall_factory.gd`
  - `client/scripts/main.gd` only for narrow layout/debug synchronization if required
- Tests and bot:
  - `server/internal/game/pathfind_test.go`
  - `server/internal/game/dungeon_obstacles_golden_test.go`
  - `server/internal/game/game_test.go` or focused new generated-water tests
  - `shared/golden/dungeon_obstacles.json`
  - `tools/bot/runtime_assertions.py`
  - `tools/bot/scenarios/28_reachable_dungeon_obstacles.json` or a new focused water scenario
  - `client/tests/test_factories.gd` or a focused renderer test
  - `tools/bot/scenarios/client/14_dungeon_wall_rendering.json` or a new focused water scenario
- Docs:
  - `docs/plans/v295_2026-06-19-water-obstacles-foundation.md`
  - `docs/as-built/v295_water-obstacles-foundation.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external assets, imported water art, shader plugins, and Godot
addons. Borrow the existing server-generated obstacle layout, wall rendering pipeline, deterministic
ground/wall texture factory, and bot wall-layout assertion conventions; add code-native water
material/geometry only.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'Path|DungeonObstacle|Water|GeneratedDungeon|LevelTransition|Collision'`
- `make bot scenario=reachable_dungeon_obstacles` or a new `make bot scenario=water_obstacles_foundation`
- `make client-unit`
- `HEADLESS=1 make bot-visual scenario=dungeon_wall_rendering` or a new focused water client
  scenario
- `make maintainability`

Manual visual proof, if desired after implementation:

```bash
make bot-visual scenario=dungeon_wall_rendering
```

## Open Questions and Risks

- No required questions for this run. Defaults: water is generated only on non-boss dungeon floors,
  blocks normal grounded movement, does not block LOS/projectiles, and uses code-native client
  material/geometry.
- Risk: adding obstacle kind to the current `walls` layout is a protocol/schema contract change.
  Keep it additive and update schemas/examples, or introduce a focused layout helper if the plan
  finds a cleaner existing path.
- Risk: `client/scripts/main.gd` and `server/internal/game/sim.go` are large coordinators. Keep
  edits narrow and extract focused helpers if the implementation would otherwise grow hotspots.
- Risk: generated water can accidentally isolate required targets. The generator must validate
  reachability with water included as blocked geometry before accepting a layout.
