# v433 As Built — Ranger Tier-3 Mesh Swap

Date: 2026-07-04

## Shipped

- Swapped `assets/characters/ranger/green_hood.glb` to CC0 Poly Pizza **Adventurer** (Polygonal Mind); archived `green_hood_legacy.glb`.
- `HERO_TARGET_HEIGHTS["ranger"]` = 1.82 m; ranger rest-pose arm fold preserved.
- Extended client scenario `96_ranger_tier3_visual`.

## Verification

```bash
make validate-assets
.venv/bin/pytest tools/assets/test_rig_hero_glbs.py::test_ranger_runtime_glb_has_required_bones -q
SCENARIO=96_ranger_tier3_visual HEADLESS=1 ./scripts/bot_client_local.sh
python3 skills/showme/scripts/render_focus.py --focus classes
```

All five hero classes now use Tier-3 CC0 external bodies with distinct target heights (barbarian 1.97m → rogue 1.70m).
