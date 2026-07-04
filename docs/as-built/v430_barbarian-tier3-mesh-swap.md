# v430 As Built — Barbarian Tier-3 Mesh Swap

Date: 2026-07-04

## Shipped

- Swapped `assets/characters/barbarian/goliath_barbarian.glb` to CC0 Poly Pizza **Male Fighter** (mastjie); archived prior mesh as `goliath_barbarian_legacy.glb`.
- Set `HERO_TARGET_HEIGHTS["barbarian"]` to 1.97 m (tallest hero).
- Re-rigged runtime `client/assets/characters/barbarian/barbarian.glb`.
- Updated `character_barbarian_v0` manifest provenance (CC0-1.0, Poly Pizza origin URL, new `sha256`).
- Extended client scenario `93_barbarian_tier3_visual`.

## Verification

```bash
make validate-assets
.venv/bin/pytest tools/assets/test_rig_hero_glbs.py::test_barbarian_runtime_glb_has_required_bones -q
SCENARIO=93_barbarian_tier3_visual HEADLESS=1 ./scripts/bot_client_local.sh
python3 skills/showme/scripts/render_focus.py --focus gear --class-id barbarian
```

Visual replay: `make bot-visual scenario=93_barbarian_tier3_visual`
