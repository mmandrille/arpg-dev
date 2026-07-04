# v427 As Built — Paladin Tier-3 Hero Body Proof

Date: 2026-07-04

## Shipped

- Updated `assets/characters/paladin/README.md` for `knight.glb` Tier-3 pipeline.
- Added `test_paladin_runtime_glb_has_required_bones` contract test.
- Extended client scenario `paladin_tier3_visual` (paladin class + full gear equip + `visual_model: character`).

## Verification

```bash
.venv/bin/pytest tools/assets/test_rig_hero_glbs.py::test_paladin_runtime_glb_has_required_bones -q
SCENARIO=92_paladin_tier3_visual HEADLESS=1 ./scripts/bot_client_local.sh
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin
```

Visual check: `make bot-visual scenario=92_paladin_tier3_visual`
