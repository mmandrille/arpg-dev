# v78 Spec — Main config drop profiles

Status: Complete

## Goal

Make `main_config.v0.json` the operational source for dungeon monster drop chance. Changing `gameplay.base_drop_rate_percent` should update dungeon monster loot rates at server rules load time without hand-editing every depth treasure class.

## Scope

- Apply `main_config.gameplay.base_drop_rate_percent` to the primary attempts used by dungeon monster loot tables and depth loot bands.
- Preserve the existing item-entry weights and drop ranking inside each treasure class.
- Replace validation that expected authored 20/80 treasure class mirrors with checks against the loaded/derived rules.
- Add focused Go coverage proving a main-config-only drop-rate edit changes all dungeon monster loot tables.

## Out of Scope

- Chest drop profiles.
- Rebalancing individual gold/consumable/gear/weapon/jewelry weights.
- A generalized profile DSL for all treasure classes.

## Acceptance

- `make validate-shared` passes.
- Focused Go tests prove main-config-only drop-rate edits affect dungeon monster treasure classes.
- `make ci` passes before commit.
