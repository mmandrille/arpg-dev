# v423 As Built — Weapon Variety GLBs

Date: 2026-07-03

## Shipped

- Added `weapon_long_sword_v0`, `weapon_rapier_v0`, `equipment_shield_kite_v0`, `equipment_shield_tower_v0` GLBs.
- Mapped `long_sword`, `rapier`, `starter_paladin_sword`, and `starter_paladin_shield` to distinct assets.

## Proof

```bash
make gen-assets && make validate-assets && make validate-shared
godot --headless --path client --script res://tests/test_item_visuals.gd
```
