# Barbarian Character (Tier 3)

Imported hero body rigged for Godot animation clips.

- **Source mesh:** `assets/characters/barbarian/goliath_barbarian.glb` (Tier-3 external GLB; CC0 Male Fighter from Poly Pizza)
- **Legacy mesh:** `assets/characters/barbarian/goliath_barbarian_legacy.glb` (pre-v430 user-provided goliath; for before/after comparison)
- **Runtime model:** `client/assets/characters/barbarian/barbarian.glb`
- **Rig pipeline:** `make gen-assets` → `rig_hero_glbs.py` → `rig_canonical_hero.py` (barbarian only)
- **Manifest:** `character_barbarian_v0` in `assets/manifests/assets.v0.json`
- **Class binding:** `shared/assets/class_presentations.v0.json` → `barbarian.model.asset_id`
- **License:** CC0-1.0 (mastjie / Poly Pizza — see manifest provenance)

## Canonical re-skin (v443)

Barbarian uses a **hybrid rig** that keeps the tier-3 art mesh while fixing the skeleton/weight
problems that forced the interim procedural body.

| Piece | File | Role |
|-------|------|------|
| Frozen bind pose | `tools/assets/canonical_skeleton.py` | 17-bone hierarchy + locals (same as `gen_glb._full_humanoid_glb`) |
| A-pose landmarks | `joint_globals_from_mesh()` | Shoulder/elbow/hand/knee/foot joints from mesh vertex clusters after height normalize |
| Segment weights | `vertex_weights()` | Up to 4 influences per vertex via inverse-distance bone segments |
| Rig entrypoint | `tools/assets/rig_canonical_hero.py` | Skins static GLB → runtime `barbarian.glb` |
| Hero orchestration | `tools/assets/rig_hero_glbs.py` | `CANONICAL_RIG_IDS = {"barbarian"}`; other classes use heuristic single-bone rig |

**Why landmarks?** The goliath mesh is **A-pose** (arms horizontal). Frozen canonical hand joints
assume **arms-down** bind pose — correct for procedural mesh + shared clips, but wrong for this art.
`joint_globals_from_mesh()` keeps canonical torso/head/legs and places arm/foot joints on detected
mesh landmarks so bones sit inside the mesh in the skeleton viewer and deform cleanly.

**Procedural fallback (dev only):** `tools/assets/gen_glb.py::barbarian_glb()` + `barbarian_mesh.py`
— not written by `make gen-assets`; useful for weight/skeleton experiments.

## Regenerate

```bash
make gen-assets   # gen_glb.py (equipment, etc.) then rig_hero_glbs.py (includes barbarian)
cd client && godot --headless --import
```

Update `character_barbarian_v0` and `character_base_humanoid_v0` `provenance.sha256` in
`assets/manifests/assets.v0.json` after regeneration.

## Swap in an AI-generated mesh

1. Export a **static** GLB from Meshy, Tripo, or similar (no embedded skin or animations).
2. Replace `goliath_barbarian.glb` (optionally rename current file to `goliath_barbarian_<label>.glb` for comparison).
3. Probe orientation: `python3 skills/3dmodel/scripts/create_model_probe.py --model assets/characters/barbarian/goliath_barbarian.glb --key barbarian_ai`
4. Re-rig: `make gen-assets` (or `PYTHONPATH=. python3 tools/assets/rig_hero_glbs.py`)
5. Update manifest `sha256`, `origin`, and `license` in `assets/manifests/assets.v0.json`.
6. `cd client && godot --headless --import`
7. Run verification commands below.

If the replacement mesh is **not** A-pose, tune landmark bands in `joint_globals_from_mesh()` or
add a pose-specific detector.

## Verify

```bash
make validate-assets
.venv/bin/pytest tools/assets/test_rig_canonical_hero.py tools/assets/test_rig_hero_glbs.py -q
python3 skills/showme/scripts/render_focus.py --focus skeleton --class-id barbarian
python3 skills/showme/scripts/render_focus.py --focus gear --class-id barbarian --items starter_barbarian_axe,helm,mail,boots
make bot-client SCENARIO=93_barbarian_tier3_visual HEADLESS=1
```

Visual replay: `make bot-visual scenario=93_barbarian_tier3_visual`
