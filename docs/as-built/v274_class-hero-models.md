# v274 As-Built - Class Hero Models

Date: 2026-06-18
Spec: [`docs/specs/v274_spec-class-hero-models.md`](../specs/v274_spec-class-hero-models.md)
Plan: [`docs/plans/v274_2026-06-18-class-hero-models.md`](../plans/v274_2026-06-18-class-hero-models.md)

## Shipped

- Barbarian, paladin, rogue, and sorcerer class presentation now use the supplied GLB assets through
  the existing stable `character_<class>_v0` manifest IDs.
- Runtime GLBs live under `client/assets/characters/<class>/` and Godot extracted one embedded
  texture/import sidecar per class.
- `shared/assets/class_presentations.v0.json` now supports optional class model `scale` and
  `height_offset` presentation metadata; paladin applies a `10.0` scale correction because the
  supplied GLB is authored around 0.18m tall.
- Static character assets can explicitly declare `required_nodes: []` in the asset manifest.
  Rigged generated class assets still require real skin joints when they list required nodes.
- `character_visual.gd` creates root-relative hand socket fallbacks when a class mesh has no
  `Skeleton3D`, preserving equipment and channel-effect attachment points for static hero models.
- The class model smoke now accepts either rigged character meshes with expected bones or static
  meshes with visible `MeshInstance3D` nodes, and it verifies sockets after replacing `ModelRoot`.

## Proof

```bash
python3 skills/3dmodel/scripts/create_model_probe.py --model assets/characters/barbarian/goliath_barbarian.glb --key v274_barbarian --yaw-degrees 0
python3 skills/3dmodel/scripts/create_model_probe.py --model assets/characters/paladin/knight.glb --key v274_paladin --yaw-degrees 0
python3 skills/3dmodel/scripts/create_model_probe.py --model assets/characters/rogue/assasine.glb --key v274_rogue --yaw-degrees 0
python3 skills/3dmodel/scripts/create_model_probe.py --model assets/characters/sorcerer/mage.glb --key v274_sorcerer --yaw-degrees 0
.venv/bin/python -m pytest tools/assets/test_validate_assets.py
make validate-shared
make validate-assets
godot --headless --path client --script res://tests/test_animation.gd
make client-unit
make ci
```

All commands above passed on 2026-06-18.

## Boundaries

- No server gameplay, class stats, combat, loot, protocol, persistence, or authority changed.
- The supplied GLBs have no skins or embedded animations; this slice keeps them as static class body
  visuals with root/socket fallback support.
- Ranger remains on the generated rigged humanoid asset.
- Visual orientation was probed at yaw `0`; further class-specific yaw tweaks can be handled as
  follow-up visual fixes if playtesting shows a model facing sideways or backward.
