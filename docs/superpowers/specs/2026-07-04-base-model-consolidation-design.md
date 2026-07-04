# Design: Base Model Consolidation + Skeleton Fix

Date: 2026-07-04

## Problem

1. **Five redundant GLBs.** Each class (barbarian, sorcerer, paladin, rogue, ranger) has its own
   model GLB with the same 17-bone skeleton. Maintaining five near-identical files is wasteful.
   `base_humanoid.glb` (8 bones, simple mannequin) is the nominal shared base but is unused in
   practice for rendered classes.

2. **Arm/hand bones end too early.** In `barbarian.glb`, `hand_r`/`hand_l` each have
   `local_translation = (0, -0.3349, 0.0489)` — the same offset as `elbow_r`/`elbow_l`. The arm
   chain has three equal segments: shoulder → arm_r → elbow_r → hand_r, with the hand bone ending
   at mid-forearm level rather than at the wrist/hand. Equipment sockets attached to `hand_r`
   float inside the forearm instead of at the hand.

## Goal

- Replace `base_humanoid.glb` with `barbarian.glb` as the single shared humanoid base. All five
  classes point at the same runtime GLB (differentiation via scale/stance data, not separate
  models). The `character_base_humanoid_v0` asset ID is preserved — only the path changes.
- Fix the `hand_r`/`hand_l` (and `foot_r`/`foot_l`) bone positions in `barbarian.glb` via a
  Python script that edits the GLB binary directly. Iteration driven by user Godot screenshots.

## Scope

**Part 1 — code cleanup (all files, no Blender):**

| File | Change |
|------|--------|
| `assets/manifests/assets.v0.json` | `character_base_humanoid_v0.runtime_path` → `client/assets/characters/barbarian/barbarian.glb` |
| `client/scenes/character.tscn` | `[ext_resource]` path → `res://assets/characters/barbarian/barbarian.glb` |
| `client/scripts/class_presentations_loader.gd` | Two hardcoded fallback paths → `barbarian.glb` |
| `client/tools/inspect_rig.gd` | Bone list → barbarian's 17 bones; delete base_humanoid check |
| `client/assets/characters/base_humanoid/base_humanoid.glb` | **Delete** + `.import` |
| `tools/assets/gen_glb.py` | Remove base_humanoid entry |
| `client/tests/test_animation.gd` | Update any base_humanoid path references |

Keep the `assets/characters/base_humanoid/` source directory and README (historical source art).

**Part 2 — GLB bone fix (iterative Python script):**

File: `tools/assets/fix_skeleton.py`

The script:
1. Parses `client/assets/characters/barbarian/barbarian.glb` (binary glTF format: 12-byte header,
   JSON chunk, binary chunk).
2. Accepts target local translations for named bones as a dict argument.
3. For each modified bone, also recomputes its inverse bind matrix in the `skins[0].inverseBindMatrices`
   accessor, so mesh skinning remains correct.
4. Writes the patched GLB back in place.

Iteration loop:
1. Run `fix_skeleton.py` with proposed bone offsets.
2. User opens Godot, takes viewport screenshot of skeleton overlay.
3. We measure visual gap → adjust offsets → repeat.

Initial target offsets (first iteration, to be tuned by screenshots):
- `hand_r` / `hand_l`: extend downward another `0.20` units (i.e., new local_translation
  `(0, -0.53, 0.07)` instead of `(0, -0.335, 0.049)`).
- `foot_r` / `foot_l`: extend downward `0.10` units past current position.
- All other bones: unchanged.

## Acceptance criteria

1. `make validate-assets` passes (barbarian.glb still validates).
2. `godot --headless --path client --script res://tools/inspect_rig.gd` passes with barbarian bones.
3. `python3 skills/showme/scripts/render_focus.py --focus skeleton` shows barbarian model.
4. After the bone fix: skeleton viewer shows `right_hand_socket` at the visual wrist/hand position
   of the ghost mesh.
5. `make ci` green.
