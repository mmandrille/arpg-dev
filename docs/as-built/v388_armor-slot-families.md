# v388 As-built — Armor Slot Families

## What shipped

- Eight new armor templates (`cave_leather_cap`, `cave_tiara`, `cave_leather_vest`, `cave_full_plate`,
  `cave_cloth_wraps`, `cave_gauntlets`, `cave_soft_boots`, `cave_plate_boots`) alongside unchanged
  medium `cave_*` templates for head, chest, gloves, and boots.
- Family tradeoffs are data-driven: armor tier in `base_stats`, penalties/bonuses in weighted
  `rollable_stats` (e.g. full plate 2× mail armor with rolled `movement_speed_percent` −25..−10;
  tiara `magic` 8 requirement with skill-affix roll bias).
- Depth-3+ treasure class includes new families; `armor_slot_families_lab` world + extended bot
  scenario `armor_slot_families_lab` prove tiara requirement gate, skill roll, and plate move penalty.
- Placeholder visuals borrow existing slot presentation families.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'ArmorSlotFamil' -count=1
make bot scenario=armor_slot_families_lab
```

## Deferred

- Per-family client icons, stash family filters, belt/ring/amulet families, full depth-band TC rebalance.
