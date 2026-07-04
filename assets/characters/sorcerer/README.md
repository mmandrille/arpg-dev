# Sorcerer Character (Tier 3)

Imported hero body rigged for Godot animation clips.

- **Source mesh:** `assets/characters/sorcerer/mage.glb` (Tier-3 external GLB; CC0 Wizard from Poly Pizza)
- **Legacy mesh:** `assets/characters/sorcerer/mage_legacy.glb` (pre-v432 user-provided mage; for before/after comparison)
- **Runtime model:** `client/assets/characters/sorcerer/sorcerer.glb`
- **Rig tool:** `python3 tools/assets/rig_hero_glbs.py` (target height ~1.80m via `HERO_TARGET_HEIGHTS`)
- **Manifest:** `character_sorcerer_v0` in `assets/manifests/assets.v0.json`
- **License:** CC0-1.0 (Polygonal Mind / Poly Pizza — see manifest provenance)

## Verify

```bash
make validate-assets
.venv/bin/pytest tools/assets/test_rig_hero_glbs.py::test_sorcerer_runtime_glb_has_required_bones -q
python3 skills/showme/scripts/render_focus.py --focus gear --class-id sorcerer --items starter_sorcerer_staff,helm,mail,boots
make bot-client SCENARIO=95_sorcerer_tier3_visual HEADLESS=1
```

Visual replay: `make bot-visual scenario=95_sorcerer_tier3_visual`
