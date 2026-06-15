# v176 As-built — Elite Objective Minimap Pin

Date: 2026-06-15

## What shipped

- Added a compact top-right elite objective minimap widget with a player dot and objective pin.
- Added `EliteObjectiveMinimapState` to derive normalized pin coordinates from existing entity records and local player position.
- The minimap is visible only for a closed `elite_objective` chest and hides when no objective exists or the chest is open.
- Bot debug state now exposes `elite_objective_minimap` visibility, pin presence, status, and normalized coordinates.
- Added client bot scenario `45_elite_objective_minimap_pin.json` for the pinned elite objective floor.

## Proof

- `make client-unit`
- `make maintainability`
- `make bot-client scenario=45_elite_objective_minimap_pin.json`

## Scope limits

- No server, protocol, objective rule, loot, or dungeon generation changes shipped.
- The widget is a compact presentation aid, not a full minimap, route line, compass, or persistent floor map.
