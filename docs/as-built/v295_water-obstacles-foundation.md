# v295 As Built: Water Obstacles Foundation

Date: 2026-06-19
Spec: [`docs/specs/v295_spec-water-obstacles-foundation.md`](../specs/v295_spec-water-obstacles-foundation.md)
Plan: [`docs/plans/v295_2026-06-19-water-obstacles-foundation.md`](../plans/v295_2026-06-19-water-obstacles-foundation.md)

## What shipped

- Added schema-backed `obstacle_generation.water` tuning to dungeon generation rules.
- Extended layout wall protocol objects with optional `kind: "water"` while default omitted kind remains a normal wall.
- Added deterministic non-boss dungeon water placement after generated targets are placed and before reachability validation accepts the floor.
- Added obstacle-kind helpers so water blocks movement/pathfinding, travel arrival, loot/corpse placement, monsters, companions, and dashes, but does not block projectiles or fog line of sight.
- Updated generated obstacle goldens with water kinds and minimum water counts.
- Rendered water layout entries as flat code-native Godot water surfaces while retaining existing stone wall rendering.
- Kept water out of fog occluder layout so water is readable floor terrain, not tall cover.

## Proof

Focused verification:

```bash
make validate-shared
cd server && go test ./internal/game -run 'Path|DungeonObstacle|Water|GeneratedDungeon|LevelTransition|Collision'
cd server && go test ./internal/game -run TestDungeonObstaclesGolden
ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=reachable_dungeon_obstacles ./scripts/bot_local.sh
make client-unit
HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=dungeon_wall_rendering ./scripts/bot_visual.sh
make maintainability
```

Result: green on 2026-06-19.

Note: `make bot scenario=reachable_dungeon_obstacles` hit a local Docker wrapper issue because the fixed-name `arpg-postgres` container was already running from another worktree. The same protocol bot script was run directly against the healthy existing Postgres container and a temporary server on `:18081`.

## Deferred

- Barbarian leap, flying navigation exceptions, hazard holes/chasms, obstacle variety, tall LOS blockers, and floor/wall shader polish remain separate selected World Detail/Navigation queue slices.
- Water does not affect combat, projectiles, fog/visibility occlusion, economy, persistence, swimming, bridges, slowdown, or damage effects.
- Final autoloop batch `make ci` remains pending until the selected World Detail/Navigation queue is complete.
