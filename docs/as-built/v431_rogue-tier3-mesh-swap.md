# v431 As Built — Rogue Tier-3 Mesh Swap

Date: 2026-07-04

## Shipped

- Swapped `assets/characters/rogue/assasine.glb` to CC0 Poly Pizza **Thief Icon** (Quaternius); archived `assasine_legacy.glb`.
- `HERO_TARGET_HEIGHTS["rogue"]` = 1.70 m (shortest hero).
- Re-rigged runtime `client/assets/characters/rogue/rogue.glb`; updated manifest provenance.
- Extended client scenario `94_rogue_tier3_visual`.

## Verification

```bash
make validate-assets
.venv/bin/pytest tools/assets/test_rig_hero_glbs.py::test_rogue_runtime_glb_has_required_bones -q
SCENARIO=94_rogue_tier3_visual HEADLESS=1 ./scripts/bot_client_local.sh
python3 skills/showme/scripts/render_focus.py --focus gear --class-id rogue
```
