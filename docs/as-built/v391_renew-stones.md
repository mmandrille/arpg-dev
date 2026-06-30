# v391 As-built — Renew Stones

Codename: renew-stones
Date: 2026-06-30

## Proved

- `renew_stone` leveled consumable with light-blue badge presentation (`#5ec8ff` / `#b8ecff`, label `Rn`)
- Unified `resource_loot_drops` pool (`upgrade_shard`, `renew_stone`) with tiered monster/chest/boss chances
- Blacksmith **Upgrade Item** + **Renew Item** only; hone/reinforce removed server + client
- Renew spends renew stone with `level >= item_level` + sell-price gold; rerolls affixes via `RerollItemRollPayload`
- Bag-only blacksmith staging for upgrade, renew, and merge (3 same-level consumables → +1)
- Bishop debug drops renew stone when `ARPG_GAMEPLAY_DEBUG=1`
- Extended bot scenarios: `resource_loot_pool_drop`, `client/blacksmith_renew_item`, updated `client/blacksmith_shard_merge`

## Key files

- `shared/rules/main_config.v0.json`, `items.v0.json`, `shared/assets/item_presentations.v0.json`
- `server/internal/game/resource_loot_drops.go`, `renew_stone_items.go`, `item_reroll.go`, `bishop_debug.go`
- `server/internal/http/account_stash_renew.go`, `account_stash_merge.go`
- `server/internal/store/leveled_consumable_store.go`
- `client/scripts/blacksmith_recipes.gd`, `blacksmith_panel.gd`, `blacksmith_merge_panel.gd`

## Verification run

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'ResourceLoot|RenewStone|ItemReroll|BishopDebug|UpgradeShard' -count=1
cd server && go test ./internal/store/... -count=1
cd server && go test ./internal/http/... -run 'Renew|Blacksmith|Merge' -count=1
make bot scenario=resource_loot_pool_drop
make bot-client SCENARIO=blacksmith_renew_item HEADLESS=1
make bot-visual scenario=blacksmith_renew_item
