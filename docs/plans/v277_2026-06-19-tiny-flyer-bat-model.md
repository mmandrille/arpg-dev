# v277 Plan: Tiny Flyer Bat Model

## Context

`bat_low_poly.glb` is a user-provided rigged bat model under the existing tiny flyer source
directory. The existing runtime contract already maps `dungeon_bat` to `monster_tiny_flyer`, so the
implementation should replace the asset and scene presentation only.

The probe found local `+Z` as the model front. The scene should therefore keep
`ModelRoot.rotation.y == 0` and scale the imported child model to `0.56`, which is 7x larger than
the initial integrated `0.08` child-model scale.

## Tasks

- [x] Read the project baseline, ADR-0001, ADR-0006, ADR-0007, and existing monster visual files.
- [x] Probe `assets/monsters/tiny_flyer/bat_low_poly.glb` for bounds, skins, animation count, and
  facing direction.
- [x] Copy the bat GLB to `client/assets/monsters/tiny_flyer/monster_tiny_flyer.glb` and import it
  through Godot.
- [x] Update `assets/manifests/assets.v0.json` provenance, source path, hash, and bat rig joints.
- [x] Replace `monster_tiny_flyer.tscn` with the correction-root/imported-child pattern and add a
  tiny flyer animation library.
- [x] Extend the Godot animation test to assert tiny flyer yaw preservation, source scale
  correction, and child bobbing.
- [x] Add mirrored skeletal wing-flap tracks to the tiny flyer idle/walk clips and assert the wing
  bone poses change during idle playback.
- [x] Run focused asset/shared/client verification and update as-built/progress docs.

## Verification

- `make validate-shared`
- `make validate-assets`
- `GODOT=/opt/homebrew/bin/godot make client-unit`

Manual visual check:

```bash
make bot-visual scenario=14_dungeon_wall_rendering
```
