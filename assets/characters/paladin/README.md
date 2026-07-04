# Paladin Character (Tier 3)

Imported hero body rigged for Godot animation clips.

- **Source mesh:** `assets/characters/paladin/knight.glb` (Tier-3 external GLB; CC0 evaluation mesh from Poly Pizza)
- **Legacy mesh:** `assets/characters/paladin/knight_legacy.glb` (pre-v429 user-provided knight; for before/after comparison)
- **Runtime model:** `client/assets/characters/paladin/paladin.glb`
- **Rig tool:** `python3 tools/assets/rig_hero_glbs.py` (target height ~1.85m via `HERO_TARGET_HEIGHTS`)
- **Manifest:** `character_paladin_v0` in `assets/manifests/assets.v0.json`
- **Class binding:** `shared/assets/class_presentations.v0.json` → `paladin.model.asset_id`
- **License:** CC0-1.0 (mastjie / Poly Pizza — see manifest provenance)

## Swap in an AI-generated mesh

1. Export a **static** GLB from Meshy, Tripo, or similar (no embedded skin or animations).
2. Replace `knight.glb` (optionally rename current file to `knight_<label>.glb` for comparison).
3. Probe orientation: `python3 skills/3dmodel/scripts/create_model_probe.py --model assets/characters/paladin/knight.glb --key paladin_ai`
4. Re-rig: `python3 tools/assets/rig_hero_glbs.py`
5. Update manifest `sha256`, `origin`, and `license` in `assets/manifests/assets.v0.json`.
6. `cd client && godot --headless --import`
7. Run verification commands below.

## Verify

```bash
python3 tools/assets/rig_hero_glbs.py
make validate-assets
.venv/bin/pytest tools/assets/test_rig_hero_glbs.py -q
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin --items starter_paladin_sword,starter_paladin_shield,helm,mail,boots
make bot-client SCENARIO=92_paladin_tier3_visual HEADLESS=1
```

Visual replay: `make bot-visual scenario=92_paladin_tier3_visual`
