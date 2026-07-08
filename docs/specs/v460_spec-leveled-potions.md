# v460 Spec — Leveled Potions

Status: Approved  
Date: 2026-07-08  
Codename: `leveled-potions`  
Baseline: v459 `skill-ground-aim`

## Purpose

Floor-scaled consumable potions with a new rejuv type, depth-tier vendor stock, and presentation that hides level on ground loot but shows level on bag/hotbar icons.

- Health (`red_potion`) and mana (`blue_potion`) potions looted on floor `L` restore `3 × L` of the matching resource.
- Rejuv (`rejuv_potion`, purple) drops at ~20% of potion rolls and restores `max(33%, L%)` of both HP and mana.
- Ground labels: generic “Health Potion” / “Mana Potion” / “Rejuv Potion” with red/blue/purple — no level.
- Inventory bag and consumable hotbar icons show the numeric potion level on the icon.
- Town vendor sells health, mana, and rejuv at the character’s `deepest_dungeon_depth` tier with level-scaled buy price.

## Non-goals

- Potion crafting, alchemy, auto-pickup filters, or market restrictions beyond existing sell/buyback.
- Renaming `red_potion` / `blue_potion` item_def_ids.
- Potion stack merge UX beyond existing consumable behavior.
- Production potion art (code-native purple family).

## Acceptance criteria

1. Potions dropped on floor `L` carry `item_level = max(1, L)` in roll payload / `rolled_stats`.
2. Health potion restores `3 × level` HP; mana potion restores `3 × level` mana (capped at max).
3. Rejuv restores `max(33%, level%)` of both resources; rejects only when both HP and mana are full.
4. Ground loot labels are generic type names without level; colors red/blue/purple.
5. Bag and hotbar potion icons display the level number.
6. Vendor offers leveled health, mana, and rejuv at `max(1, deepest_dungeon_depth)` with incremental buy price from shared rules.
7. Starter potions are level 1 (restore 3 HP/mana).
8. Bot lab scenario proves depth-10 loot restore and vendor buy at max depth.

## Scope

| Area | Files |
|------|-------|
| Rules | `shared/rules/main_config.v0.json`, `items.v0.json`, `shops.v0.json`, schemas |
| Presentation | `shared/assets/item_presentations.v0.json` |
| Server | `server/internal/game/potion_items.go`, `sim.go`, `shop.go`, `handlers.go`, `starter_loadout.go` |
| Client | `loot_node_factory.gd`, `inventory_panel.gd`, `consumable_bar.gd`, helper for icon label |
| Bot | `tools/bot/scenarios/*potion*`, update `08_heal_lab.json` |
| Tests | Go unit tests, GDScript unit tests |

## Asset and plugin decision

- Adopt: existing `ItemIconDrawer`, `item_presentations` families, leveled-consumable `item_level` payload pattern from upgrade shards.
- Borrow: `BlacksmithUpgradePreviewScript.item_level()` for reading level from items.
- Reject: external assets/plugins.

## Test proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'Potion|Consumable|Shop' -count=1
make bot scenario=potion_depth_lab
make client-unit
```

## Open questions

None — defaults from `/next` brief apply (`deepest_dungeon_depth` for shop tier, `base × level` pricing, 20/40/40 potion split).
