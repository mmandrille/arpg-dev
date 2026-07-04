# v422 As Built — Hero Rig Rollout

Date: 2026-07-03

## Shipped

- Extended `HERO_TARGET_HEIGHTS` to barbarian, paladin, rogue, ranger, sorcerer (~1.85m).
- Regenerated all hero runtime GLBs; updated manifest hashes.
- Fixed `/showme` `visual_capture.gd` parse error on `item-icons` focus.

## Proof

```bash
make gen-assets && make validate-assets
godot --headless --path client --script res://tests/test_animation.gd
```
