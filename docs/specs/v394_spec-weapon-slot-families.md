# v394 Spec — Weapon Slot Families

Status: Draft
Date: 2026-06-30
Codename: weapon-slot-families

## Purpose

Extend v388 armor-family tradeoffs to **sword and bow** main-hand weapons: light / medium / heavy
families with requirement and roll-pool identity.

| Weapon | Light | Medium (existing) | Heavy |
|--------|-------|-------------------|-------|
| Sword | Cave Rapier — dex, fast, crit rolls | Cave Blade | Cave Heavy Blade — str, high damage, negative attack-speed rolls |
| Bow | Hunting Bow — dex, fast rolls | Cave Bow | War Bow — str+dex, high damage, negative attack-speed rolls |

## Non-goals

- Shield, dagger, staff, axe, greatsword families; belt/ring/amulet.
- New stats or protocol fields; production art.

## Acceptance criteria

- [ ] Four new templates; existing `cave_blade` and `cave_bow` unchanged as medium tier.
- [ ] Heavy sword `damage_max` ≥ 2× medium sword; heavy blade rolls include negative `attack_speed_percent`.
- [ ] Visual/presentation mappings borrow existing sword/bow assets.
- [ ] Depth-3+ treasure class includes new templates; `weapon_slot_families_lab` world preset.
- [ ] Extended bot scenario: rapier requirement gate + heavy blade / hunting bow roll keys.
- [ ] `make validate-shared` passes.

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'WeaponSlotFamil' -count=1
make bot scenario=weapon_slot_families_lab
```
