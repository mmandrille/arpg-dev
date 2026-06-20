# v296 As Built: Hazard Holes / Chasms

Date: 2026-06-19
Spec: [`docs/specs/v296_spec-hazard-holes-chasms.md`](../specs/v296_spec-hazard-holes-chasms.md)
Plan: [`docs/plans/v296_2026-06-19-hazard-holes-chasms.md`](../plans/v296_2026-06-19-hazard-holes-chasms.md)

## What shipped

- Added schema-backed `obstacle_generation.holes` tuning to dungeon generation rules.
- Extended layout wall protocol objects with optional `kind: "hole"` while omitted kind remains a normal wall.
- Added deterministic non-boss dungeon hole placement after water and before final reachability validation accepts the floor.
- Reused obstacle-kind helpers so holes block grounded movement/pathfinding, travel arrival, loot/corpse placement, monsters, companions, and dashes.
- Kept holes from blocking projectiles and fog line of sight.
- Updated generated obstacle goldens with hole kind/count proof.
- Rendered hole layout entries as flat code-native dark chasm surfaces while retaining existing stone wall and water rendering.
- Kept all non-wall floor features out of fog occluder layout.

## Proof

Focused verification:

```bash
make validate-shared
cd server && go test ./internal/game -run 'Path|DungeonObstacle|Hole|Water|GeneratedDungeon|LevelTransition|Collision'
cd server && go test ./internal/game -run TestDungeonObstaclesGolden
ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=reachable_dungeon_obstacles ./scripts/bot_local.sh
make client-unit
HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=dungeon_wall_rendering ./scripts/bot_visual.sh
make maintainability
```

Result: green on 2026-06-19.

Note: the direct bot script path was used because the local Docker wrapper conflicts with an
already-running fixed-name `arpg-postgres` container in this multi-worktree setup.

## Deferred

- Flying navigation exceptions, barbarian leap over water/holes, obstacle variety, tall LOS blockers, and floor/wall shader polish remain separate selected World Detail/Navigation queue slices.
- Holes do not add falling, damage, knockback, bridges, recovery, combat, projectile, fog occlusion, economy, or persistence behavior in this slice.
- Final autoloop batch `make ci` remains pending until the selected World Detail/Navigation queue is complete.
