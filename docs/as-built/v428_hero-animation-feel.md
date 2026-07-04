# v428 As Built — Hero Animation Feel Pass

Date: 2026-07-04  
Spec-gate: exempt (client presentation + shared feel JSON only)

## Shipped

- Tuned `shared/assets/combat_feel_presentation.v0.json` melee lunge (distance/recovery) for snappier attack feedback on normalized hero rigs.
- Adjusted `item_visuals.v0.json` mount offsets for `long_sword`, generic `shield` (kite GLB), and `helm` on ~1.85m heroes.

## Verification

```bash
make validate-shared
godot --headless --path client --script res://tests/test_melee_lunge_presentation.gd
godot --headless --path client --script res://tests/test_attack_animation_scaling.gd
godot --headless --path client --script res://tests/test_item_visuals.gd
SCENARIO=82_melee_lunge_micro_step HEADLESS=1 ./scripts/bot_client_local.sh
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin
```

Visual: `make bot-visual scenario=82_melee_lunge_micro_step`
