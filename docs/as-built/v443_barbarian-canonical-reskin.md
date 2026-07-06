# v443 As Built — Barbarian Canonical Re-skin

Date: 2026-07-06

## Problem

Procedural `barbarian_mesh.py` achieved correct 17-bone bind pose and multi-bone weights but looked
blocky. Re-enabling the v430 goliath mesh via `rig_hero_glbs.py` restored art quality but regressed:

1. **Single-bone heuristic weights** — arms bound 100% to chest; bad deformation.
2. **Frozen canonical joint positions on A-pose art** — hand/hip bones inside torso while mesh limbs
   extended horizontally (skeleton viewer mismatch).

## Shipped

- `tools/assets/canonical_skeleton.py` — frozen joint locals/globals, segment distance weights,
  `joint_globals_from_mesh()` A-pose landmark placement.
- `tools/assets/rig_canonical_hero.py` — skins static hero GLBs onto canonical skeleton with mesh-fitted globals.
- `tools/assets/rig_hero_glbs.py` — barbarian back in `HEROES`; `CANONICAL_RIG_IDS` routes to canonical rig.
- `tools/assets/gen_glb.py` — barbarian removed from `TARGETS` (runtime owned by rig step).
- `tools/assets/test_rig_canonical_hero.py` — landmark hand proximity + dual-weight contracts.
- Runtime `client/assets/characters/barbarian/barbarian.glb` regenerated from `goliath_barbarian.glb`.
- Manifest `sha256` + provenance updated for `character_barbarian_v0` / `character_base_humanoid_v0`.
- `assets/characters/barbarian/README.md` — pipeline architecture and verification.

Procedural `barbarian_glb()` / `barbarian_mesh.py` retained for dev reference; not merge output.

## Verification

```bash
.venv/bin/pytest tools/assets/test_rig_canonical_hero.py tools/assets/test_rig_hero_glbs.py -q
make validate-assets
make model model=character_barbarian_v0 CHECK=1
python3 skills/showme/scripts/render_focus.py --focus skeleton --class-id barbarian
python3 skills/showme/scripts/render_focus.py --focus gear --class-id barbarian --items starter_barbarian_axe,helm,mail,boots
```

Visual replay: `make bot-visual scenario=93_barbarian_tier3_visual`
