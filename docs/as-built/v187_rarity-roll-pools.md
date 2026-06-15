# v187 As Built: Rarity Roll Pools

Date: 2026-06-15
Status: Complete - `make ci` green

## What Shipped

- Item rarity rows now use `stat_rolls_min` / `stat_rolls_max` ranges instead of a single fixed
  roll count.
- Configured roll counts are `common` 1, `magic` 1-2, `rare` 2-4, and `unique`/`set` 3-5.
- Roll candidates can declare `min_rarity`; omitted values default to common, and higher rarities
  inherit lower-rarity pools.
- `set` rarity is declared for the roll-count contract but marked `random_rollable: false`, so
  fixed set items do not become random shop/dungeon drops in this slice.
- The Go item roller now chooses roll count from the configured range using the existing seeded RNG.
- Shared validator and shop/item goldens were updated for the new deterministic roll stream.
- Protocol bot scenario `78_rarity_roll_pools.json` proves a deterministic unique item exposes
  high-rarity roll count bounds and inherited stat-pool payloads.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'ItemRoll|ItemTemplate|Rarity'`
- `make client-unit`
- `make bot scenario=78_rarity_roll_pools.json`
- `make ci`

## Deferred

- Prefix/suffix affix names and procedural item labels.
- Blacksmith/crafting routes that add or improve rolls.
- Final rarity weights and stat range balance.
- Live rare combat-affix behavior beyond existing stat aggregation.
- Skill cooldown and mana-cost affix behavior.
