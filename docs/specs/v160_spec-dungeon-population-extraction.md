# v160 Spec — Dungeon population extraction

Date: 2026-06-14
Status: Complete
Codename: dungeon-population-extraction

## Purpose

Move generated dungeon runtime population out of the broad `Sim` coordinator into a focused server
module while preserving existing generated floor behavior. The slice keeps stairs, teleporters,
chests, elite-objective chest IDs, loose loot, generated monsters, boss visual selection, rarity
scaling, party HP scaling, and corpse spawning behavior unchanged.

## Non-goals

- No gameplay tuning, monster placement changes, loot table changes, or objective rule changes.
- No protocol/schema change and no client presentation change.
- No broad rewrite of dungeon generation, pathfinding, or `game_test.go`.

## Acceptance criteria

- `ensureDungeonLevel` still generates missing dungeon levels deterministically for the same seed
  and rules.
- Runtime population of generated stairs, teleporters, chests, loot, monsters, bosses, and corpses
  remains equivalent to the pre-extraction behavior.
- Generated elite-objective chest runtime IDs are still tracked on the owning `LevelState`.
- Unknown generated loot, monster, boss template, boss base monster, and monster loot table errors
  remain clear and level-scoped.
- `sim.go` shrinks and the maintainability baseline is lowered if the ratchet allows it.

## Scope and likely files

- `server/internal/game/dungeon_population.go`: focused runtime population helpers.
- `server/internal/game/sim.go`: delegate generated level population to the focused helper.
- `server/internal/game/dungeon_population_test.go`: focused behavior-preserving coverage.
- `.maintainability/file-size-baseline.tsv`: lower `sim.go` baseline when verified.
- `docs/as-built/v160_dungeon-population-extraction.md` and `PROGRESS.md`: lifecycle closeout.

## Test and bot proof

- `cd server && go test ./internal/game -run 'TestPopulateDungeonLevelTracksEliteObjectiveChestIDs|TestPopulateDungeonLevelPreservesBossAndRarityRuntimeState' -count=1`
- `cd server && go test ./internal/game -run 'TestGeneratedDungeonSourcesUseDepthLootTables|TestGeneratedDungeonMonsterRarityGolden|TestDungeonEliteObjectiveChestRequiresEliteLeader|TestEliteObjectiveChestRequiresLeaderKill' -count=1`
- `make maintainability`
- `make ci`

## Open questions and risks

- No blocking product questions. This is a server-internal extraction with no player-facing behavior
  change.
- Determinism risk is limited to preserving allocation and iteration order from the generated
  slices; the implementation must keep the current stairs -> teleporters -> chests -> loot ->
  monsters -> corpses order.
