# v444 Plan — Class Body Forks

Status: Complete  
Spec: [`docs/specs/v444_spec-class-body-forks.md`](../specs/v444_spec-class-body-forks.md)

## Approach

- Reproducible **programmatic vertex morph** from `base_human_mesh.glb` (`tools/assets/class_body_morph.py`).
- Canonical rig path for all five player classes + `base_human` fallback.
- Keep `character_base_human_v0` as `FALLBACK_ASSET_ID` for unknown classes.

## Tasks

- [x] Add `tools/assets/class_body_morph.py` + unit tests
- [x] Wire `class_body_morph.py generate` into `make gen-assets`
- [x] Register five classes in `rig_hero_glbs.py` (`HEROES`, `HERO_TARGET_HEIGHTS`, `CANONICAL_RIG_IDS`)
- [x] Generate `{class}_mesh.glb` sources and rigged runtime GLBs
- [x] Add manifest entries `character_{class}_v0` with provenance sha256
- [x] Point `class_presentations.v0.json` at per-class assets
- [x] Regenerate `model_preview_catalog.v0.json`
- [x] Extend rig landmark tests per class; update `test_animation.gd` asset assertions
- [x] READMEs under `assets/characters/{class}/`
- [x] As-built + PROGRESS lifecycle

## Verification

```bash
.venv/bin/pytest tools/assets/test_class_body_morph.py tools/assets/test_rig_canonical_hero.py tools/assets/test_rig_hero_glbs.py tools/assets/test_model_catalog.py -q
make validate-shared validate-assets
make client-unit
make model model=character_barbarian_v0 CHECK=1
make model model=character_rogue_v0 CHECK=1
```
