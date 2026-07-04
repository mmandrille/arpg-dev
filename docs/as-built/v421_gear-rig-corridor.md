# v421 As Built — Gear Rig Corridor

Date: 2026-07-03
Spec: [`docs/specs/v421_spec-gear-rig-corridor.md`](../specs/v421_spec-gear-rig-corridor.md)
Plan: [`docs/plans/v421_2026-07-03-gear-rig-corridor.md`](../plans/v421_2026-07-03-gear-rig-corridor.md)

## Shipped

- Added eight deterministic armor slot GLBs via `gen_glb.py` (helm, chest, gloves, boots, belt, amulet, ring, shield).
- Manifest fallback equipment entries now point at dedicated armor GLBs instead of `rusty_sword.glb`.
- `EquipmentVisualResolver` loads GLB equipment when manifest path is not the sword placeholder.
- Paladin mesh height normalized to ~1.85m in `rig_hero_glbs.py`; class presentation scale reduced from 10.0 to 1.0.

## Proof

```bash
make gen-assets
make validate-assets
make validate-shared
godot --headless --path client --script res://tests/test_animation.gd
godot --headless --path client --script res://tests/test_item_visuals.gd
.venv/bin/python -m pytest tools/assets/test_rig_hero_glbs.py -q
```

## Deferred

- `/showme` gear capture blocked by pre-existing parse error in `visual_capture.gd:92`.
- Other hero height normalization (v422).
- Weapon variety and monster rig passes (v423–v425).
