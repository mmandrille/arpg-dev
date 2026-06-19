# v275 Spec: Rigged Hero Models

## Summary

Turn the four supplied static hero GLBs from v274 into skinned GLBs with the same bone names used by
the original generated humanoid rig, so the existing `character_anims.tres` clips drive the new
barbarian, paladin, rogue, and sorcerer models.

## Goals

- Keep the v274 hero body assets as the visible meshes for barbarian, paladin, rogue, and sorcerer.
- Add a deterministic asset tool that injects the shared humanoid skin contract into those GLBs:
  `root`, `spine`, `arm_l`, `hand_l`, `arm_r`, `hand_r`, `leg_l`, and `leg_r`.
- Restore non-empty `required_nodes` for the four hero manifest entries so `make validate-assets`
  proves they are rigged skin joints again.
- Make `make gen-assets` regenerate the supplied hero runtime GLBs through the rigging pass, so the
  class models do not regress to the old generated placeholders.
- Tighten the Godot class model smoke to prove the class GLBs expose a `Skeleton3D`, hand sockets are
  `BoneAttachment3D`s again, and the existing walk/attack clips change the expected bone poses.

## Non-Goals

- Do not alter server gameplay, class stats, combat, loot, protocol, persistence, or authority.
- Do not add a Blender/manual skinning dependency in this slice.
- Do not promise production-quality deformation. The slice may use automatic rigid region weights
  to make the existing animations work; final artist-quality weights remain future art work.
- Do not replace the ranger model.

## Asset Decision

- Adopt: the v274 user-provided local GLBs as the source meshes.
- Borrow: the generated humanoid rig joint names, animation library, `BoneAttachment3D` socket
  contract, asset manifest validation, and Godot class presentation loader.
- Reject: a parallel class animation system, client-authoritative presentation state, or manual
  Blender-only workflow that agents cannot reproduce in CI.

## Acceptance

- The four runtime hero GLBs import in Godot with a `Skeleton3D` and all required humanoid bones.
- `make validate-assets` fails if any required class bone is missing from the GLB skin.
- The existing class walk/attack/off-hand animation tracks apply to the new class models.
- `make gen-assets` preserves the rigged hero GLBs by running the deterministic rigging tool.
- Manual visual check remains available with `make bot-visual scenario=20_menu_create_join_flow`.
