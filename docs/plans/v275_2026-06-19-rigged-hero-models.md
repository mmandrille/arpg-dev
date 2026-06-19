# v275 Plan: Rigged Hero Models

## Context

v274 intentionally integrated the supplied class hero GLBs as static meshes. The old player
animation library targets bone-pose paths under `Skeleton3D`, so static meshes cannot use those
clips directly. This slice adds a reproducible GLB rigging pass that preserves the supplied mesh,
material, and embedded texture while adding the shared humanoid joint hierarchy and automatic rigid
weights.

## Tasks

- [x] Add `tools/assets/rig_hero_glbs.py` to inject the shared humanoid skin into the four class GLBs.
- [x] Add focused Python tests for the rigging tool.
- [x] Wire `make gen-assets` to run the hero rigging pass after generated placeholder assets.
- [x] Regenerate the four runtime class GLBs and update manifest provenance/hashes and required nodes.
- [x] Update the Godot class animation test to require class skeletons and bone-driven walk/attack poses.
- [x] Reimport assets in Godot and run focused validation/client checks.
- [x] Record as-built proof, update progress docs, run final CI, and commit.

## Verification

- `.venv/bin/python -m pytest tools/assets/test_rig_hero_glbs.py tools/assets/test_validate_assets.py`
- `make gen-assets`
- `make validate-assets`
- `godot --headless --path client --script res://tests/test_animation.gd`
- `make client-unit`
- `make ci`
- Manual visual check available: `make bot-visual scenario=20_menu_create_join_flow`
