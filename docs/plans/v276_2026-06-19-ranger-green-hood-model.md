# v276 Plan: Ranger Green Hood Model

## Context

`green_hood.glb` is a static Sketchfab-sourced character GLB. It has no skin or animations, so the
right local boundary is the existing deterministic hero rigging tool from v275. The source is also
authored at a much larger unit scale, so scale belongs in shared class presentation data, not code.

## Tasks

- [x] Inspect `assets/characters/ranger/green_hood.glb` for skins, animations, source metadata, and
  bounds.
- [x] Add ranger to `tools/assets/rig_hero_glbs.py`.
- [x] Regenerate `client/assets/characters/ranger/ranger.glb` through `make gen-assets`.
- [x] Update `assets/manifests/assets.v0.json` with green hood provenance, source path, hash, and
  required humanoid bones.
- [x] Add a ranger class presentation scale override in `shared/assets/class_presentations.v0.json`.
- [x] Bake a ranger-only source rest-pose correction so arms/hands are not left in the source
  T-pose.
- [x] Extend the Godot class animation test to include ranger and prove scaled class sockets keep
  mounted equipment at normal world scale.
- [x] Run focused asset/client/bot verification and update as-built/progress docs.

## Verification

- `.venv/bin/python -m pytest tools/assets/test_rig_hero_glbs.py tools/assets/test_validate_assets.py -q`
- `make validate-shared`
- `make validate-assets`
- `godot --headless --path client --script res://tests/test_animation.gd`
- `make client-unit`
- `HEADLESS=1 make bot-visual scenario=ranger_class_foundation`

Manual visual check:

```bash
make bot-visual scenario=ranger_class_foundation
```
