# v447 As-built — Class Gear Retune

Date: 2026-07-06

## Shipped

- Retuned `item_visuals.v0.json` for showme/starter set: sword grip via `z` offset + `rotation x:90`, shield vertical via `rotation y:-90`, helm/mail/boots scale `1.0` (Tier-3 imports already normalized).
- Per-class `class_transforms` on `long_sword`, `helm`, `mail`, `boots` for rogue/barbarian/sorcerer/ranger/paladin height deltas.
- Added `boots_right_socket` on `foot_r` in `gear_sockets.v0.json`; `equipment_visuals.gd` mounts mirrored right boot.
- Fixed off-hand weapon mirror to negate `z` grip offset when main-hand item equips off-hand.
- Tightened `equipped_gear_fit_probe` with min global-scale bands + boots mirror assertion.
- Updated `item_visual_scale_probe` shield scale ceiling for normalized kite GLB.

## Limits

- Representative + starter + showme items only; not all 67 `3d_model` defs.
- Boots use mirrored single-boot GLB, not per-foot art variants.
- Rogue helm may need further art pass tuning on smallest body.

## Verification

```bash
.venv/bin/python tools/validate_shared.py
make validate-assets
godot --headless --path client --script res://tests/test_item_visuals.gd
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin
python3 skills/showme/scripts/render_focus.py --focus gear --class-id rogue
```
