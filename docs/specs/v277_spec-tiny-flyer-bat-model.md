# v277 Spec: Tiny Flyer Bat Model

## Summary

Replace the generated tiny flyer placeholder with the supplied `bat_low_poly.glb` model while
preserving the existing `dungeon_bat` monster definition, `monster_tiny_flyer` scene key, hover
metadata, and client-only animation boundary.

## Goals

- Use `assets/monsters/tiny_flyer/bat_low_poly.glb` as the source for `monster_tiny_flyer_v0`.
- Copy the runtime bytes to `client/assets/monsters/tiny_flyer/monster_tiny_flyer.glb`.
- Update asset manifest provenance, hash, source path, and rig joint requirements.
- Keep `shared/assets/monster_visuals.v0.json` unchanged so `dungeon_bat` still resolves through
  the existing data-driven visual path.
- Add a tiny flyer scene animation library that animates the child model, not the correction root.
- Add looped skeletal wing-flap tracks for the imported bat while preserving whole-model hover bob.
- Preserve the bat's local `+Z` front and apply a larger source scale correction in the scene.

## Non-Goals

- Do not change bat gameplay, combat tuning, true flying movement, pathing, AI, loot, server
  authority, or protocol.
- Do not introduce a new asset pipeline or hardcode model paths in gameplay code.
- Do not claim final production animation quality; this is a runtime model replacement with the
  current presentation clips.

## Asset Decision

- Adopt: `assets/monsters/tiny_flyer/bat_low_poly.glb`, supplied locally by the user.
- Borrow: existing monster visual catalog, asset manifest validation, Godot GLB import path, and
  node-root animation pattern from imported wolf/quadruped monsters.
- Reject: generated placeholder tiny flyer art, hardcoded client model paths, and a separate
  gameplay-specific flyer implementation.

## Acceptance

- `monster_tiny_flyer_v0` resolves to the supplied bat runtime GLB and its provenance hash matches.
- `dungeon_bat` continues to resolve to `scene: monster_tiny_flyer` with `animation_profile:
  hover_flyer`.
- `client/scenes/monster_tiny_flyer.tscn` instantiates the imported bat under `ModelRoot/Model`,
  applies no yaw correction, and scales the child model to `0.56`.
- The scene-level `AnimationPlayer` exposes `idle`, `walk`, `hit`, and `death`; walk bobs the
  child model without overwriting `ModelRoot.rotation.y`.
- The `idle` and `walk` clips rotate `rechterVleugel_09` and `linkerVleugel_013` so the bat flaps
  its wings while hovering or moving.
- Manual visual replay remains available with `make bot-visual scenario=14_dungeon_wall_rendering`
  after entering a dungeon level that can spawn `dungeon_bat`.
