# v460 Plan — Leveled Potions

Status: Complete  
Goal: Floor-scaled health/mana/rejuv potions with depth-tier vendor stock and bag/hotbar level labels.  
Architecture: Data-driven `potion_rules` in `main_config`; `potion_items.go` owns restore math, shop tier/pricing, drop kind resolution, and summaries; server stores `item_level` in roll payload; client shows generic ground labels and numeric level on bag/hotbar icons.

## Baseline and shortcut decision

- Reuses leveled-consumable payload pattern from upgrade shards / renew stones.
- Adopt in-repo `ItemIconDrawer` + `item_presentations`; reject external assets.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/main_config.v0.json` | `potion_rules` tuning |
| Modify | `shared/rules/items.v0.json`, schema | Health/mana base 3, `rejuv_potion` |
| Modify | `shared/rules/shops.v0.json`, schema | Leveled fixed offers + rejuv |
| Modify | `shared/rules/worlds.v0.json`, schema | `potion_depth_lab`, `item_level` on loot |
| Create | `server/internal/game/potion_items.go` | Restore, shop, drop, summary helpers |
| Create | `server/internal/game/potion_items_test.go` | Unit tests |
| Modify | `server/internal/game/sim.go` | Loot payloads, consume, annotate views |
| Modify | `server/internal/game/shop.go`, `handlers.go` | Leveled buy/sell/appraisal |
| Create | `client/scripts/potion_icon_label.gd` | Level overlay on icons |
| Modify | `client/scripts/inventory_panel.gd`, `consumable_bar.gd`, `loot_node_factory.gd` | Wire presentation |
| Create | `tools/bot/scenarios/potion_depth_lab.json`, `potion_shop_lab.json` | Extended proofs |
| Modify | goldens, `08_heal_lab.json`, `33_shop_stock_lifecycle.json` | Contract updates |

## Task 1 — Shared rules and schemas

- [x] Step 1.1: `potion_rules`, items, shops, worlds, schemas
- [x] Step 1.2: `make validate-shared`

## Task 2 — Server authority

- [x] Step 2.1: `potion_items.go` + loot/consume/shop integration
- [x] Step 2.2: `go test ./internal/game/... -run 'Potion|Shop'`

## Task 3 — Client presentation

- [x] Step 3.1: Ground generic labels; bag/hotbar level overlay
- [x] Step 3.2: `make client-unit` (existing suites)

## Task 4 — Bot proof

- [x] Step 4.1: `potion_depth_lab`, `potion_shop_lab` (extended)
- [x] Step 4.2: Update `08_heal_lab`, shop goldens

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'Potion|Consumable|Shop' -count=1
make bot scenario=potion_depth_lab
make bot scenario=potion_shop_lab
make client-unit
```
