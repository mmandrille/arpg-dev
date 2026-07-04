# Rogue Character (Tier 3)

Imported hero body rigged for Godot animation clips.

- **Source mesh:** `assets/characters/rogue/assasine.glb` (Tier-3 external GLB; CC0 Thief Icon from Poly Pizza)
- **Legacy mesh:** `assets/characters/rogue/assasine_legacy.glb` (pre-v431 user-provided assassin; for before/after comparison)
- **Runtime model:** `client/assets/characters/rogue/rogue.glb`
- **Rig tool:** `python3 tools/assets/rig_hero_glbs.py` (target height ~1.70m via `HERO_TARGET_HEIGHTS`)
- **Manifest:** `character_rogue_v0` in `assets/manifests/assets.v0.json`
- **Class binding:** `shared/assets/class_presentations.v0.json` → `rogue.model.asset_id` + `idle_stance` lean
- **License:** CC0-1.0 (Quaternius / Poly Pizza — see manifest provenance)

## Verify

```bash
make validate-assets
.venv/bin/pytest tools/assets/test_rig_hero_glbs.py::test_rogue_runtime_glb_has_required_bones -q
python3 skills/showme/scripts/render_focus.py --focus gear --class-id rogue --items starter_rogue_sword,leather_cap,leather_vest,soft_boots
make bot-client SCENARIO=94_rogue_tier3_visual HEADLESS=1
```

Visual replay: `make bot-visual scenario=94_rogue_tier3_visual`
