# v276 As-Built - Ranger Green Hood Model

Date: 2026-06-19
Spec: [`docs/specs/v276_spec-ranger-green-hood-model.md`](../specs/v276_spec-ranger-green-hood-model.md)
Plan: [`docs/plans/v276_2026-06-19-ranger-green-hood-model.md`](../plans/v276_2026-06-19-ranger-green-hood-model.md)

## Shipped

- Added `assets/characters/ranger/green_hood.glb` as the source asset for `character_ranger_v0`.
- Added ranger to `tools/assets/rig_hero_glbs.py`, so `make gen-assets` replaces the generated
  placeholder with a rigged green hood runtime GLB.
- Regenerated `client/assets/characters/ranger/ranger.glb` and imported its embedded texture PNGs
  under `client/assets/characters/ranger/`.
- Updated the asset manifest with Sketchfab/CC-BY provenance, source URL, runtime hash, and the
  required humanoid skin-joint list.
- Added a ranger class presentation scale override (`0.0092`) to normalize the source model's large
  authoring units.
- Added a ranger-only rest-pose bake to `tools/assets/rig_hero_glbs.py`, folding the source
  T-pose arms/hands downward before skinning so the ranger stands in a usable in-game pose.
- Extended the class animation smoke to include ranger and prove scaled class hand sockets keep
  mounted gear at normal world scale.

## Proof

```bash
make gen-assets
.venv/bin/python -m pytest tools/assets/test_rig_hero_glbs.py tools/assets/test_validate_assets.py -q
make validate-shared
make validate-assets
godot --headless --path client --script res://tests/test_animation.gd
make client-unit
HEADLESS=1 make bot-visual scenario=ranger_class_foundation
```

All commands above passed on 2026-06-19.

## Boundaries

- No ranger gameplay, skills, starter loadout, server authority, combat, loot, persistence, or
  protocol changed.
- The green hood model uses the current automatic rigid-region weights. Final artist-quality
  deformation remains future art/tech-art polish.
- Hand socket scale compensation now prevents class model scale corrections from resizing mounted
  weapons or shields.
