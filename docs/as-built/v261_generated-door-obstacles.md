# v261 As-built - Generated Door Obstacles

Date: 2026-06-18

## What shipped

- Added schema-backed generated-door tuning under `obstacle_generation.doors`.
- Extracted generated dungeon output types from `dungeon_gen.go` to keep the hotspot under its
  maintainability ratchet.
- Added deterministic generated-door placement for eligible horizontal generated wall segments.
- Door placement now uses candidate-level commits so failed obstacle attempts cannot leak stale
  doors into the successful floor.
- Eligible walls are split into left/right wall pieces with a gap occupied by a generated closed
  `wooden_door`.
- Generated doors populate as normal `interactable` entities with the existing closed-door barrier
  and activation behavior.
- Extended the dungeon obstacle golden and bot scenario so the pinned generated dungeon opens a
  generated door through the live protocol path.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game
make bot scenario=reachable_dungeon_obstacles
make maintainability
```

## Manual visual check

```bash
make bot-visual scenario=14_dungeon_wall_rendering
```

## Scope limits

- No full room/corridor PCG, secret doors, locked/keyed doors, destructible doors, rotated door
  visuals, door-specific minimap icons, fog/LOS doorway behavior, or boss-floor door generation
  shipped.
- No reconnect/resume, database, replay-format, loot, monster, or combat tuning change shipped.
