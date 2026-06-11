# v78 As-built — Main Config Drop Profiles

Status: Complete

## Shipped

- Dungeon monster treasure classes now receive their primary success/no-drop split from `main_config.gameplay.base_drop_rate_percent` during rules loading.
- Existing treasure class entry weights remain authored in `treasure_classes.v0.json`, preserving the current drop ranking.
- Shared validation now reports the global drop-rate owner in `main_config`.
- Focused Go tests prove changing only `main_config.v0.json` changes all dungeon monster loot table drop rates.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestMainConfig|TestDungeonMonsterLootRate'`
- `make ci`
