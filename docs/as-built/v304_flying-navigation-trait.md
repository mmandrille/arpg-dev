# v304 As-Built - Flying Navigation Trait

Date: 2026-06-19
Status: Complete

## Shipped

- Added optional `navigation_trait` to monster rules with schema/server validation. Omitted values
  remain grounded; `dungeon_bat` is now marked `flying`.
- Added a narrow server helper for monster obstacle semantics. Flying monsters ignore `water` and
  `hole` wallObstacle kinds for pathfinding and final monster movement collision, but normal walls,
  closed interactable barriers, players, and living entities remain blockers.
- Added optional preset world wall `kind` support and a compact `flying_navigation_lab` with water
  and hole blockers.
- Preserved preset wall kinds through server loading and the Godot `render_world_walls` path so
  lab water/hole tiles render with the existing obstacle visuals.
- Added `tools/bot/scenarios/99_flying_navigation_trait.json` as the protocol and visual proof.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'FlyingNavigationTrait|FlyingNavigationLab|Path'
ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=flying_navigation_trait ./scripts/bot_local.sh
make client-unit
HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=flying_navigation_trait ./scripts/bot_visual.sh
make maintainability
```

## Deferred

- Barbarian/player mobility exceptions over water and holes.
- Obstacle variety, tall LOS blockers, shader polish, falling/damage, bridges, recovery, or terrain
  costs.
- Projectile, fog/LOS, loot placement, player auto-navigation, companion navigation, and boss-floor
  generation changes.
- Full autoloop batch `make ci`, repo review, and refactor handoff remain deferred until the selected
  World Detail/Navigation queue completes.
