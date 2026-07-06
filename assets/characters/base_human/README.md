# Base Human Character (Tier 3)

Shared player body mesh — starting point for per-class customization.

- **Source mesh:** `assets/characters/base_human/base_human_mesh.glb` (Tier-3 external GLB; CC0 Male Fighter from Poly Pizza)
- **Legacy mesh:** `assets/characters/base_human/base_human_mesh_legacy.glb` (pre-v430 goliath import; for before/after comparison)
- **Runtime model:** `client/assets/characters/base_human/base_human.glb`
- **Rig pipeline:** `make gen-assets` → `rig_hero_glbs.py` → `rig_canonical_hero.py`
- **Manifest:** `character_base_human_v0` in `assets/manifests/assets.v0.json`
- **Class binding:** `shared/assets/class_presentations.v0.json` — each player class uses its own `character_{class}_v0` fork (v444); `character_base_human_v0` remains fallback for unknown classes
- **License:** CC0-1.0 (mastjie / Poly Pizza — see manifest provenance)

## Canonical re-skin (v443)

`base_human` uses a **hybrid rig** that keeps the tier-3 art mesh while fixing skeleton/weight
problems that forced the interim procedural body.

| Piece | File | Role |
|-------|------|------|
| Frozen bind pose | `tools/assets/canonical_skeleton.py` | 17-bone hierarchy + locals (same as `gen_glb._full_humanoid_glb`) |
| A-pose landmarks | `joint_globals_from_mesh()` | Shoulder/elbow/hand/knee/foot joints from mesh vertex clusters after height normalize |
| Segment weights | `vertex_weights()` | Up to 4 influences per vertex via inverse-distance bone segments |
| Rig entrypoint | `tools/assets/rig_canonical_hero.py` | Skins static GLB → runtime `base_human.glb` |
| Hero orchestration | `tools/assets/rig_hero_glbs.py` | `HEROES` contains only `base_human`; canonical re-skin path |

**Why landmarks?** The source mesh is **A-pose** (arms horizontal). Frozen canonical hand joints
assume **arms-down** bind pose — correct for procedural mesh + shared clips, but wrong for this art.
`joint_globals_from_mesh()` keeps canonical torso/head/legs and places arm/foot joints on detected
mesh landmarks so bones sit inside the mesh in the skeleton viewer and deform cleanly.

**Procedural fallback (dev only):** `tools/assets/gen_glb.py::base_human_glb()` + `base_human_mesh.py`
— not written by `make gen-assets`; useful for weight/skeleton experiments.

## Regenerate

```bash
make gen-assets   # gen_glb.py (equipment, etc.) then rig_hero_glbs.py (base_human)
cd client && godot --headless --import
```

Update `character_base_human_v0` `provenance.sha256` in `assets/manifests/assets.v0.json` after regeneration.

## Swap in an AI-generated mesh

1. Export a **static** GLB from Meshy, Tripo, or similar (no embedded skin or animations).
2. Replace `base_human_mesh.glb` (optionally archive current file as `base_human_mesh_<label>.glb`).
3. Probe orientation: `python3 skills/3dmodel/scripts/create_model_probe.py --model assets/characters/base_human/base_human_mesh.glb --key base_human_ai`
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
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin --items helm,mail,boots
make model model=character_base_human_v0 CHECK=1
```

Visual replay (barbarian class scenario): `make bot-visual scenario=93_barbarian_tier3_visual`
