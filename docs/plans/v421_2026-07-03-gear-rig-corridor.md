# v421 Plan — Gear Rig Corridor

Spec: [`docs/specs/v421_spec-gear-rig-corridor.md`](../specs/v421_spec-gear-rig-corridor.md)

## Tasks

- [x] Add deterministic equipment GLB generators in `gen_glb.py`
- [x] Point fallback manifest entries at armor GLBs (not rusty sword)
- [x] Skip procedural fallback when dedicated equipment GLB is registered
- [x] Paladin mesh height normalization in `rig_hero_glbs.py`; class scale → 1.0
- [x] Regenerate assets + Godot import sidecars
- [x] Update animation + equipment probe tests
- [ ] `/showme gear --class-id paladin` capture
- [x] Docs: as-built, lifecycle

## Verify

```bash
make gen-assets
make validate-assets
make validate-shared
godot --headless --path client --script res://tests/test_animation.gd
godot --headless --path client --script res://tests/test_item_visuals.gd
.venv/bin/python -m pytest tools/assets/test_rig_hero_glbs.py -q
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin
```
