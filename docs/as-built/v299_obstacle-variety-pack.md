# v299 As-Built - Obstacle Variety Pack

Date: 2026-06-19
Spec: [`docs/specs/v299_spec-obstacle-variety-pack.md`](../specs/v299_spec-obstacle-variety-pack.md)
Plan: [`docs/plans/v299_2026-06-19-obstacle-variety-pack.md`](../plans/v299_2026-06-19-obstacle-variety-pack.md)

## Shipped

- Extended shared wall-kind enums to accept `rock`, `column`, and `rubble` in preset worlds, latest v8 protocol schemas, and the generated dungeon obstacle golden.
- Added schema-backed `obstacle_generation.solid_kind_weights` so generated non-boss solid obstacle groups deterministically choose `wall`, `rock`, `column`, or `rubble`.
- Kept rock, column, and rubble on the existing rectangular wall collision model: they block normal walking, pathfinding, grounded monsters, companions, and solid projectiles.
- Kept rock, column, and rubble out of flying and Barbarian Leap mobility exceptions, which remain limited to water/hole-style floor obstacles.
- Left fog and line-of-sight behavior unchanged for this slice; new solid variants are not visibility occluders beyond existing wall/door behavior.
- Allowed generated doors to split any solid line variant so door reachability proofs do not depend on rolling only default `wall` blockers.
- Added `obstacle_variety_lab` with rock, column, and rubble blockers plus a reachable quest leaf, and added the `obstacle_variety_pack` bot scenario for protocol and visual proof.
- Added Godot code-native non-rectangular visuals for rock, column, and rubble using the existing wall renderer and stable node metadata.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'ObstacleVariety|DungeonObstacle|GeneratedObstacleCollisionPaths|FlyingNavigationTrait|LeapObstacle|MobilityObstacle'`
- `cd server && go test ./internal/game -run TestDungeonObstaclesGolden`
- `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=obstacle_variety_pack ./scripts/bot_local.sh`
- `make client-unit`
- `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=obstacle_variety_pack ./scripts/bot_visual.sh`
- `make maintainability`

## Deferred

- True non-rectangular server collision, rotated collision, destructible blockers, terrain costs, and boss-floor obstacle generation.
- Fog, visibility, minimap occlusion, and high-obstacle line-of-sight behavior; those remain the dedicated line-of-sight blocker slice.
- Production obstacle assets, imported models, shader/material polish, and biome-specific art treatment.
- Final selected-batch `make ci` and the due review/refactor handoff after the World Detail/Navigation queue completes.
