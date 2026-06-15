# v175 As-built — Elite Objective HUD

Date: 2026-06-14

## What shipped

- Added a compact elite objective HUD tracker that appears on floors with an `elite_objective` reward chest.
- The tracker derives state from current entity metadata: alive `monster_pack_leader` count and objective chest open state.
- HUD states cover active leader-clear, claim unlocked chest, complete, and hidden.
- Added bot wait/assert support for `elite_objective_tracker` debug state.
- Added client bot scenario `44_elite_objective_hud.json` for the pinned elite objective floor.

## Proof

- `make maintainability`
- `make client-unit`
- `make bot-client scenario=44_elite_objective_hud.json`

## Scope limits

- No server protocol or objective rule changes shipped.
- Minimap pins, compass routing, and richer quest/objective models remain deferred.
