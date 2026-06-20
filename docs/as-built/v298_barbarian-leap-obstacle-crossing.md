# v298 As-Built - Barbarian Leap Obstacle Crossing

Date: 2026-06-19
Status: Complete

## Shipped

- Added schema-backed `mobility.ignore_obstacle_kinds` support for skills and assigned `water` and
  `hole` only to Barbarian `leap`.
- Added server validation that rejects unsupported mobility ignore kinds, including normal walls.
- Reworked player mobility endpoint resolution so Leap can sweep across ignored water/hole
  obstacles, but the final landing must remain valid floor; hard blockers like walls and closed
  interactable barriers still stop all player mobility.
- Kept Dash and Charge on the default hard-blocking path so water/hole exceptions do not leak into
  other mobility skills or ordinary walking.
- Added `barbarian_leap_obstacle_lab` and
  `tools/bot/scenarios/100_barbarian_leap_obstacle_crossing.json` to prove Leap crosses a rendered
  water/hole strip.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'LeapObstacle|MobilityObstacle|RogueDash|GeneratedObstacleCollisionPaths'
ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=barbarian_leap_obstacle_crossing ./scripts/bot_local.sh
make client-unit
HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=barbarian_leap_obstacle_crossing ./scripts/bot_visual.sh
make maintainability
```

## Deferred

- Normal player walking, click-to-move, auto-navigation, swimming, bridges, falling, recovery,
  terrain costs, and persistent traversal abilities over water or holes.
- Dash, Charge, Teleport, monsters, companions, projectiles, fog/LOS, loot placement, and dungeon
  generation changes beyond the compact preset lab.
- Obstacle variety, tall LOS blockers, and wall/floor shader polish remain later selected
  World Detail/Navigation slices.
- Full autoloop batch `make ci`, repo review, and refactor handoff remain deferred until the
  selected World Detail/Navigation queue completes.
