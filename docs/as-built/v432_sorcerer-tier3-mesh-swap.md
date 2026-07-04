# v432 As Built — Sorcerer Tier-3 Mesh Swap

Date: 2026-07-04

## Shipped

- Swapped `assets/characters/sorcerer/mage.glb` to CC0 Poly Pizza **Wizard** (Polygonal Mind); archived `mage_legacy.glb`.
- `HERO_TARGET_HEIGHTS["sorcerer"]` = 1.80 m.
- Extended client scenario `95_sorcerer_tier3_visual`.

## Verification

```bash
make validate-assets
.venv/bin/pytest tools/assets/test_rig_hero_glbs.py::test_sorcerer_runtime_glb_has_required_bones -q
SCENARIO=95_sorcerer_tier3_visual HEADLESS=1 ./scripts/bot_client_local.sh
```
