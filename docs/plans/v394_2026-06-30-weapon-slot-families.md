# v394 Plan — Weapon Slot Families

Date: 2026-06-30
Spec: `docs/specs/v394_spec-weapon-slot-families.md`

## Tasks

- [x] Shared: four weapon templates (`cave_rapier`, `cave_heavy_blade`, `cave_hunting_bow`, `cave_war_bow`) with family tradeoffs
- [x] Shared: depth-3+ treasure classes include new weapons; `weapon_slot_families_lab` world preset
- [x] Server: Go tests for family roll keys, requirement gates, and heavy negative attack-speed rolls
- [x] Bot: extended `weapon_slot_families_lab` scenario for rapier gate + heavy blade / hunting bow roll keys
- [x] Goldens: regenerate `shop_offers.json` after treasure-class catalog expansion

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'WeaponSlotFamil' -count=1
make bot scenario=weapon_slot_families_lab
```

## Deferred

- Dagger/staff/axe/greatsword families, shield families, per-family stash filters, full depth rebalance
