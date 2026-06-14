# v162 As-built — Objective chest presentation

Date: 2026-06-14
Status: Complete

## What shipped

- Added optional `elite_objective` metadata to v8 entity views and schemas.
- Server snapshots/deltas now mark generated elite-objective reward chests without changing
  interaction, lock, or loot authority.
- Extracted chest mesh part creation into `client/scripts/chest_presentation.gd` and added a
  display-only objective marker for marked treasure chests.
- Client entity records retain the objective flag, keep the marker visible after open state sync,
  and expose `elite_objective` / `has_objective_marker` through bot debug rows.
- Added client bot presentation matching for objective marker fields and scenario
  `tools/bot/scenarios/client/41_objective_chest_presentation.json`.

## Verification

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestEliteObjectiveChestRequiresLeaderKill|TestPopulateDungeonLevelTracksEliteObjectiveChestIDs' -count=1`
- `make client-unit`
- `make bot scenario=68_dungeon_elite_side_objective.json`
- `make bot-client scenario=41_objective_chest_presentation.json`
- `make ci`

Visual verification command:

```bash
make bot-visual scenario=68_dungeon_elite_side_objective.json
```
