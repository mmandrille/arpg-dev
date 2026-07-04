# Ranger Character (Tier 3)

Imported hero body rigged for Godot animation clips.

- **Source mesh:** `assets/characters/ranger/green_hood.glb` (Tier-3 external GLB; CC0 Warrior from Poly Pizza)
- **Legacy mesh:** `assets/characters/ranger/green_hood_legacy.glb` (pre-v433 Sketchfab green-hood; for before/after comparison)
- **Runtime model:** `client/assets/characters/ranger/ranger.glb`
- **Rig tool:** `python3 tools/assets/rig_hero_glbs.py` (target height ~1.82m; culls mastjie T-pose shoulder caps + shoulder-only arm weights for shared `character_anims`)
- **Manifest:** `character_ranger_v0` in `assets/manifests/assets.v0.json`
- **Class binding:** `shared/assets/class_presentations.v0.json` → `ranger.model.asset_id` (`scale: 1.0` since v433 fix)
- **License:** CC0-1.0 (mastjie / Poly Pizza — see manifest provenance)

## Verify

```bash
make validate-assets
.venv/bin/pytest tools/assets/test_rig_hero_glbs.py::test_ranger_runtime_glb_has_required_bones -q
python3 skills/showme/scripts/render_focus.py --focus gear --class-id ranger --items starter_ranger_bow,helm,mail,boots
make bot-client SCENARIO=96_ranger_tier3_visual HEADLESS=1
```

Visual replay: `make bot-visual scenario=96_ranger_tier3_visual`
