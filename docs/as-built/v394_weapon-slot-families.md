# v394 As-built — Weapon Slot Families

## What shipped

- Four weapon templates: `cave_rapier`, `cave_heavy_blade`, `cave_hunting_bow`, `cave_war_bow` alongside
  medium `cave_blade` and `cave_bow`.
- Family tradeoffs: light dex-biased speed/crit rolls; heavy high base damage with rolled negative
  `attack_speed_percent`.
- Depth-3+ treasure classes include new weapons; `weapon_slot_families_lab` + extended bot scenario.
- Regenerated `shop_offers.json` golden after treasure-class catalog expansion.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'WeaponSlotFamil' -count=1
make bot scenario=weapon_slot_families_lab
```

## Deferred

- Dagger/staff/axe/greatsword families, shield families, per-family stash filters, full depth rebalance.
