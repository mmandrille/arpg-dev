# v444 As Built — Class Body Forks

Date: 2026-07-06

## Shipped

- `tools/assets/class_body_morph.py` — reproducible vertex morph from `base_human_mesh.glb` into five class silhouettes.
- `make gen-assets` runs morph generation before rig pass.
- `tools/assets/rig_hero_glbs.py` — all five player classes in `HEROES`, `HERO_TARGET_HEIGHTS`, `CANONICAL_RIG_IDS`.
- Static sources: `assets/characters/{class}/{class}_mesh.glb` (committed).
- Runtime GLBs regenerated under `client/assets/characters/{class}/`.
- Manifest entries `character_{class}_v0` with provenance sha256.
- `shared/assets/class_presentations.v0.json` — per-class `character_{class}_v0` bindings.
- `shared/assets/model_preview_catalog.v0.json` regenerated.
- Tests: `test_class_body_morph.py`, per-class landmark + bone gates, `test_animation.gd` asset assertions, `test_model_viewer.gd` catalog update.
- Class READMEs under `assets/characters/{class}/`.
- `character_base_human_v0` kept as `FALLBACK_ASSET_ID` for unknown classes.

## Morphology (data-driven in `CLASS_MORPHS`)

| Class | Height (m) | Silhouette |
|-------|------------|------------|
| barbarian | 1.97 | Wide shoulders, thick chest/arms/legs |
| paladin | 1.88 | Broad armored frame |
| ranger | 1.82 | Lean, longer torso |
| sorcerer | 1.86 | Thin, taller torso/neck |
| rogue | 1.70 | Compact, slight forward mass |

## Verification

```bash
.venv/bin/pytest tools/assets/test_class_body_morph.py tools/assets/test_rig_canonical_hero.py tools/assets/test_rig_hero_glbs.py tools/assets/test_model_catalog.py -q
make validate-shared
make client-unit
make model model=character_barbarian_v0 CHECK=1
```

Visual: `python3 skills/showme/scripts/render_focus.py --focus skeleton --class-id barbarian` (and per-class gear/skeleton captures).

## Deferred

- Blender/AI mesh art pass replacing programmatic morph (morph params remain tunable data).
- Per-class landmark band overrides if future art breaks clusters.
- Extended-only bot visual scenarios per class (not CI pack).
