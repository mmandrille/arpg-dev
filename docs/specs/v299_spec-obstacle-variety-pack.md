# v299 Spec - Obstacle Variety Pack

Status: Complete
Date: 2026-06-19
Codename: obstacle-variety-pack

## Purpose

Add visible dungeon obstacle variety without changing the server's authoritative collision shape.
Generated and preset wall layout entries can now be marked as `rock`, `column`, or `rubble`, the
server continues to treat them as solid blockers, and the Godot client renders each kind with a
distinct non-rectangular code-native visual backed by the same rectangular collision/layout data.

This follows v295-v298's obstacle-kind foundation. It improves dungeon readability and variety
while preserving the current AABB movement/pathing model so the later LOS blocker slice can decide
which tall obstacle kinds affect fog and visibility.

## Non-goals

- No true polygon, rotated, per-rock, or destructible collision; server collision remains the
  existing layout rectangle for every obstacle variant.
- No fog-of-war, visibility, minimap, lighting, projectile VFX, combat balance, loot placement,
  boss-floor generation, bridge/swim/fall/recovery, or terrain-cost changes.
- No imported assets, shaders, Godot plugins, Blender pipeline work, or production art.
- No new protocol message shape; only latest-schema wall-kind enums are extended for the existing
  optional `kind` field.
- No changes to water, hole, flying navigation, or Barbarian Leap exceptions beyond ensuring the
  new solid kinds remain hard blockers.

## Acceptance Criteria

- Shared schemas allow wall layout kind values `rock`, `column`, and `rubble` anywhere preset or
  authoritative wall layout objects already allow `wall`, `water`, and `hole`.
- Dungeon generation rules own solid obstacle variety through schema-backed weights, including
  validation that weights are non-negative and at least one solid kind is enabled.
- Generated non-boss dungeon solid obstacle groups deterministically choose from `wall`, `rock`,
  `column`, and `rubble`; water and hole generation remains separate.
- The existing generated-obstacle reachability proof still treats rock/column/rubble as blocked
  geometry and keeps required targets reachable.
- Rock, column, and rubble block normal walking, auto-navigation, grounded monster movement,
  companions, and solid projectile collision like existing solid walls.
- Flying monsters and Barbarian Leap do not ignore rock, column, or rubble.
- Rock, column, and rubble do not become fog/LOS occluders in this slice; the dedicated
  line-of-sight blocker slice owns that gameplay perception change.
- A compact preset lab includes rock, column, and rubble blockers in the wall layout and routes the
  player around them to pick up loot.
- The Godot wall renderer produces distinct node names/metadata and non-rectangular code-native
  visuals for rock, column, and rubble while preserving water/hole rendering.
- A protocol bot scenario observes all three new kinds and proves an entity action can path around
  them.
- A headless visual bot replay for the same lab shows all three obstacle visuals.

## Scope and Likely Files

- Shared contracts/rules:
  - `shared/rules/dungeon_generation.v0.json`
  - `shared/rules/dungeon_generation.v0.schema.json`
  - `shared/rules/worlds.v0.json`
  - `shared/rules/worlds.v0.schema.json`
  - `shared/protocol/session_snapshot.v8.schema.json`
  - `shared/protocol/state_delta.v8.schema.json`
  - `shared/golden/dungeon_obstacles.json`
  - `shared/golden/dungeon_obstacles.v0.schema.json`
- Server generation/collision:
  - `server/internal/game/rules.go`
  - `server/internal/game/obstacle_blocking.go`
  - `server/internal/game/dungeon_obstacle_variety.go`
  - `server/internal/game/dungeon_doors.go`
  - `server/internal/game/dungeon_gen.go`
  - `server/internal/game/dungeon_obstacles_golden_test.go`
  - `server/internal/game/dungeon_obstacle_variety_test.go`
  - `server/internal/game/monster_navigation_traits_test.go`
  - `server/internal/game/mobility_obstacle_crossing_test.go`
- Client rendering/tests:
  - `client/scripts/wall_renderer.gd`
  - `client/tests/test_factories.gd`
- Bot proof:
  - `tools/bot/scenarios/101_obstacle_variety_pack.json`
- Docs:
  - `docs/plans/v299_2026-06-19-obstacle-variety-pack.md`
  - `docs/as-built/v299_obstacle-variety-pack.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external assets, shader packages, Godot addons, and imported model
assets. Borrow the existing wall renderer, generated wall layout path, water/hole kind metadata
pattern, and code-native mesh/material approach.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'ObstacleVariety|DungeonObstacle|GeneratedObstacleCollisionPaths|FlyingNavigationTrait|LeapObstacle|MobilityObstacle'`
- `cd server && go test ./internal/game -run TestDungeonObstaclesGolden`
- `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=obstacle_variety_pack ./scripts/bot_local.sh`
- `make client-unit`
- `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=obstacle_variety_pack ./scripts/bot_visual.sh`
- `make maintainability`

Manual visual proof, if desired:

```bash
make bot-visual scenario=obstacle_variety_pack
```

## Open Questions and Risks

- No required questions for this run. Defaults: the variants are visual/solid-wall kinds, generation
  weights are data-owned, and fog/LOS behavior remains unchanged.
- Risk: changing generated obstacle kind selection can perturb dungeon goldens. Update the focused
  obstacle golden and keep assertions semantic where possible.
- Risk: `column` visually implies height, but LOS/fog blocking is explicitly deferred to avoid
  partially implementing the later line-of-sight blocker slice.
