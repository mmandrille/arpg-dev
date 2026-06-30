# v387 Plan — Class Weapon Affinities

Status: Complete
Goal: Rollable class-tied weapon bonuses with server authority and green/red client tooltips.

## Task 1 — Shared contracts
- [x] `item_templates` schema: `class_affinities`, `dagger`/`war_hammer` types, affinity stat keys
- [x] Five exemplar templates + presentation families
- [x] Lab loot in `skill_progression_lab`

## Task 2 — Server
- [x] `class_item_affinities.go`: roll, status, apply stats
- [x] Extend `ItemRollPayload`, item views, derived stats, equip preview
- [x] Go tests

## Task 3 — Client
- [x] Tooltip green/red affinity lines (inventory, shop, market)
- [x] Unit test for formatting

## Task 4 — Bot
- [x] `109_class_weapon_affinities_lab.json` extended

## Final verification
```bash
make validate-shared
cd server && go test ./internal/game/... -run ClassAffinity -count=1
make client-unit
make bot scenario=109_class_weapon_affinities_lab
```
