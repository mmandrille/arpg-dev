# Paladin Character (Tier 3)

Imported hero body rigged for Godot animation clips.

- **Source mesh:** `assets/characters/paladin/knight.glb` (user-provided static GLB)
- **Runtime model:** `client/assets/characters/paladin/paladin.glb`
- **Rig tool:** `python3 tools/assets/rig_hero_glbs.py` (target height ~1.85m via `HERO_TARGET_HEIGHTS`)
- **Manifest:** `character_paladin_v0` in `assets/manifests/assets.v0.json`
- **Class binding:** `shared/assets/class_presentations.v0.json` → `paladin.model.asset_id`
- **License:** user-provided-unverified (see manifest provenance)

## Verify

```bash
python3 tools/assets/rig_hero_glbs.py
make validate-assets
.venv/bin/pytest tools/assets/test_rig_hero_glbs.py -q
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin --items starter_paladin_sword,starter_paladin_shield,helm,mail,boots
make bot-client SCENARIO=92_paladin_tier3_visual HEADLESS=1
```
