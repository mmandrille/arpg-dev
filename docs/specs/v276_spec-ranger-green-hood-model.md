# v276 Spec: Ranger Green Hood Model

## Summary

Replace the generated ranger hero placeholder with the supplied `green_hood.glb` model while
keeping the same class presentation, humanoid animation, and equipment socket contracts used by
the other hero classes.

## Goals

- Use `assets/characters/ranger/green_hood.glb` as the source model for `character_ranger_v0`.
- Generate `client/assets/characters/ranger/ranger.glb` through the deterministic hero rigging
  pass so the existing walk/attack/off-hand animation clips drive the ranger.
- Preserve the asset manifest contract: runtime path, provenance, hash, and required humanoid
  skin joints.
- Add a ranger class presentation scale override because the source model is authored in much
  larger units than the other hero models.
- Bake a ranger-specific rest-pose correction so the source mesh's T-pose arms and hands hang near
  the body before the shared skeleton is applied.
- Keep hand-mounted gear at authored world scale even when a class model needs a presentation scale
  correction.

## Non-Goals

- Do not change ranger gameplay, skills, starter loadout, server authority, combat, loot, or
  protocol.
- Do not add a Blender/manual skinning dependency in this slice.
- Do not claim production-quality deformation; the generated rig uses the current automatic rigid
  region weights.

## Asset Decision

- Adopt: `assets/characters/ranger/green_hood.glb`, supplied locally by the user.
- Borrow: the v275 deterministic hero rigging tool, shared humanoid bone names, class presentation
  loader, and Godot animation smoke.
- Reject: client-only gameplay shortcuts, hardcoded model paths in runtime code, and a separate
  ranger animation system.

## Acceptance

- `character_ranger_v0` resolves to the green hood runtime GLB and imports in Godot.
- The runtime ranger GLB contains `root`, `spine`, `arm_l`, `hand_l`, `arm_r`, `hand_r`, `leg_l`,
  and `leg_r` as skin joints.
- The existing class animation test includes ranger and proves the walk/attack/off-hand clips
  rotate the expected bones.
- The ranger body no longer appears in the source T-pose during gameplay.
- Ranger class presentation scale does not shrink or enlarge mounted weapons.
- Manual visual check remains available with `make bot-visual scenario=ranger_class_foundation`.
