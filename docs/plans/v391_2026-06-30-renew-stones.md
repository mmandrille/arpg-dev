# v391 Plan — Renew Stones

Date: 2026-06-30
Spec: `docs/specs/v391_spec-renew-stones.md`

## Tasks

- [x] Shared: `renew_stone` item + presentation; `resource_loot_drops` in `main_config` + schema
- [x] Server game: `resource_loot_drops.go`, `renew_stone_items.go`, `item_reroll.go`, bishop debug renew drop
- [x] Server game: monster/chest/boss hooks; remove `upgrade_shard_drops.go`
- [x] HTTP: `item_renew` recipe, `POST /v0/account-stash/items/renew`, bag merge endpoint
- [x] Store: `RenewInventoryItem`, `MergeLeveledConsumablesFromBag`, generalized leveled consumable spend
- [x] Client: upgrade + renew recipes only; bag-only staging; merge by `item_instance_id`
- [x] Bot: `resource_loot_pool_drop`, `client/blacksmith_renew_item`; updated bag merge; removed hone scenarios
- [x] Tests: resource loot, reroll, renew HTTP, bishop renew debug, blacksmith panel unit test

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'ResourceLoot|RenewStone|ItemReroll|BishopDebug|UpgradeShard' -count=1
cd server && go test ./internal/store/... -count=1
cd server && go test ./internal/http/... -run 'Renew|Blacksmith|Merge' -count=1
make client-unit
make bot scenario=resource_loot_pool_drop
make bot-client SCENARIO=blacksmith_renew_item HEADLESS=1
```
