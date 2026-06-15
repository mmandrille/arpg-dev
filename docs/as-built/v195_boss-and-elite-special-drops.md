# v195 As-built: Boss And Elite Special Drops

Date: 2026-06-15
Status: Complete - `make ci` green

## What shipped

- Extended loot and treasure-class entries with authored `unique_item_id` and `set_item_id`
  references.
- Added server validation for authored drop references against enabled ready unique and set catalogs.
- Added fixed payload spawning for authored unique/set drops, reusing the same builders as the debug
  unique chest.
- Added `boss_special_tc_tier_1`, making `boss_drop_tier_1` drop `Conduit Staff`, `Stormrunner
  Covenant Bow`, and one rolled `cave_amulet`.
- Added `elite_objective_special_tc_1`, making the elite objective reward chest drop `Verdant
  Vanguard Gloves` and one rolled `cave_ring`.
- Added protocol bot scenario `84_boss_special_drops.json`, proving Cave Warden drops the authored
  unique and set rewards.

## Verification

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'TreasureClass|SpecialDrop|Boss|EliteObjective|DungeonMonsterLootRate' -count=1`
- `make bot scenario=84_boss_special_drops.json`
- `make ci`

## Deferred

- Final drop rates, weighted special pools, random set rarity drops, and personal loot remain future
  itemization/economy work.
