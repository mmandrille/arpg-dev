# v160 As-built — Dungeon population extraction

Date: 2026-06-14
Status: Complete

## What shipped

- Moved generated dungeon runtime population from `server/internal/game/sim.go` into
  `server/internal/game/dungeon_population.go`.
- Preserved the existing runtime allocation order for generated stairs, teleporters, chests,
  loose loot, monsters, and corpse spawning.
- Kept generated elite-objective chest runtime IDs on `LevelState`, including the v159 kill gate.
- Kept generated monster rarity scaling, boss template substitution, boss visuals, boss pattern
  deck initialization, loot table validation, and party HP scaling unchanged.

## Proof

- `cd server && go test ./internal/game -run 'TestPopulateDungeonLevelTracksEliteObjectiveChestIDs|TestPopulateDungeonLevelPreservesBossAndRarityRuntimeState' -count=1`
- `cd server && go test ./internal/game -run 'TestGeneratedDungeonSourcesUseDepthLootTables|TestGeneratedDungeonMonsterRarityGolden|TestDungeonEliteObjectiveChestRequiresEliteLeader|TestEliteObjectiveChestRequiresLeaderKill' -count=1`
- `make maintainability`
- `make ci`

## Deferred

- No gameplay tuning or objective rule changes were included. The next selected autoloop slice
  owns the product behavior change from "kill one elite leader" to "clear every generated elite
  leader."
