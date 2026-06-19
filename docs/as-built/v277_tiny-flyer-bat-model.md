# v277 As-Built: Tiny Flyer Bat Model

## Shipped

- Replaced `client/assets/monsters/tiny_flyer/monster_tiny_flyer.glb` with the user-provided
  `assets/monsters/tiny_flyer/bat_low_poly.glb` runtime bytes.
- Updated `monster_tiny_flyer_v0` manifest metadata with source path, provenance, SHA-256, and
  real bat skin joints.
- Added `client/animations/monster_tiny_flyer_anims.tres` and rewired
  `client/scenes/monster_tiny_flyer.tscn` to use a `ModelRoot/Model` child transform pattern.
- Kept the existing `dungeon_bat` visual metadata unchanged: `monster_tiny_flyer`, hover offset,
  hover flyer animation profile, and data-driven resolution all remain intact.
- Added client animation coverage proving tiny flyer yaw correction remains zero, the imported
  child model carries the `0.56` source-scale correction, and walk animation bobs the child model
  without overwriting `ModelRoot`.
- Added mirrored wing-bone rotation tracks for `rechterVleugel_09` and `linkerVleugel_013` in the
  tiny flyer `idle` and `walk` clips so the bat flaps while hovering or moving.

## Verification

- `make validate-shared`
- `make validate-assets`
- `GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd`
- `GODOT=/opt/homebrew/bin/godot make client-unit`
- `GODOT=/opt/homebrew/bin/godot HEADLESS=1 make bot-visual scenario=41_monster_visual_catalog`

Manual visual replay:

```bash
make bot-visual scenario=41_monster_visual_catalog
```
