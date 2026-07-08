# v460 — Leveled Potions

Spec: [`docs/specs/v460_spec-leveled-potions.md`](../specs/v460_spec-leveled-potions.md)  
Plan: [`docs/plans/v460_2026-07-08-leveled-potions.md`](../plans/v460_2026-07-08-leveled-potions.md)

## What it proved

- Health and mana potions loot with `item_level = floor depth` and restore `3 × level` HP or mana.
- Rejuv potion (~20% of treasure-class potion rolls) restores `max(33%, level%)` of both resources.
- Ground loot shows generic type names (no level); bag and hotbar icons show the numeric level.
- Town vendor sells health, mana, and rejuv at `max(1, deepest_dungeon_depth)` with `base_price × level` buy pricing.
- Starter potions are level 1 (3 HP/mana restore).

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'Potion|Consumable|Shop' -count=1
make bot scenario=potion_depth_lab
make bot scenario=potion_shop_lab
make client-unit
```

Visual check (optional): `make bot-visual scenario=potion_depth_lab` — preset level-10 red potion on ground, level on hotbar after pickup.

## Deferred

- Potion stack merge UX
- CI pack promotion (`potion_*_lab` scenarios are extended-only)
- Dedicated GDScript unit test for `potion_icon_label.gd`
