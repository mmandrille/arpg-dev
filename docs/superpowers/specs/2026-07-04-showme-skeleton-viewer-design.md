# Design: showme skeleton viewer (`--focus skeleton`)

Date: 2026-07-04

## Problem

The paladin's `hand_r` / `hand_l` bones are not physically located where the visible mesh
hands are, causing weapons and shields to float at the wrong position. Defensive gear
(helm, boots, chest) is also undersized on the paladin model. Before fixing socket offsets
or item scales, we need a tool that shows **exactly where every bone sits on the rendered
mesh** so corrections can be made by inspection rather than blind trial-and-error.

## Goal

Add `--focus skeleton` to the existing showme infrastructure. It renders a single
screenshot of the character in a forced spread pose with:

1. A **red sphere** at every bone in the `Skeleton3D` — no label, just position.
2. A **colored labeled sphere** at each active equipment socket (`right_hand_socket`,
   `head_socket`, `boots_socket`, etc.), color-coded by slot type.

This is a diagnostic/iteration tool for visual artists and engineers, not a player-facing
feature.

## Scope

- **In scope:** new `_setup_skeleton()` in `visual_capture.gd`; new `"skeleton"` choice in
  `render_focus.py`; SKILL.md catalog entry.
- **Out of scope:** fixing gear sizes, fixing socket offsets, any server/protocol/rules
  change.

## Architecture

No new files. Changes touch:

| File | Change |
|------|--------|
| `client/scripts/showme/visual_capture.gd` | Add `_setup_skeleton()` and `"skeleton"` match arm |
| `skills/showme/scripts/render_focus.py` | Add `"skeleton"` to `--focus` choices; add size override 800×600 |
| `.claude/skills/showme/SKILL.md` | Catalog entry |

The `--class-id` flag already exists — `_setup_skeleton()` reuses `_apply_class_model()`
exactly as `_setup_gear()` does. Default class when none is given: `paladin`.

## Spread pose

After the character is loaded and the `Skeleton3D` is found, override bone poses **before
the first rendered frame** using `Skeleton3D.set_bone_pose_rotation()`.

Target bone names and rotations (paladin has all of these):

| Bone | Spread rotation | Effect |
|------|-----------------|--------|
| `arm_r` | `Quaternion(Vector3.FORWARD, -PI/2)` | Right arm 90° out |
| `arm_l` | `Quaternion(Vector3.FORWARD, PI/2)` | Left arm 90° out |
| `leg_r` | `Quaternion(Vector3.RIGHT, PI/9)` | Right leg ~20° forward/out |
| `leg_l` | `Quaternion(Vector3.RIGHT, PI/9)` | Left leg ~20° forward/out |

Rotation is applied only if the bone exists (checked via `skel.find_bone(name) >= 0`).
The function is class-agnostic; it works on any class whose skeleton has these bone names.

`await process_frame` after pose override so transforms propagate before marker placement.

## Red bone dots

```
for i in range(skel.get_bone_count()):
    var pose = skel.get_bone_global_pose(i)
    var world_pos = skel.global_transform * pose.origin
    # place sphere at world_pos, radius 0.025, color red
```

Each sphere is a `MeshInstance3D` with a `SphereMesh` (radius 0.025) and a red
`StandardMaterial3D` (no shading, `albedo_color = Color("#e03030")`).

## Socket marker spheres

After bone dots, iterate `SOCKET_BY_SLOT` keys. For each socket name, call
`character.find_child(socket_name, true, false)`. If found:

- Sphere radius: 0.045
- Color by category:

| Sockets | Color |
|---------|-------|
| `right_hand_socket`, `off_hand_socket` | `#4488ff` (blue) |
| `head_socket` | `#ffee44` (yellow) |
| `chest_socket`, `belt_socket`, `amulet_socket` | `#44dddd` (cyan) |
| `boots_socket`, `gloves_socket` | `#ff8822` (orange) |
| `ring_left_socket`, `ring_right_socket` | `#cc44ff` (magenta) |

- `Label3D` positioned 0.12 above the sphere, billboard mode, font_size 28, text = socket
  name without the `_socket` suffix (e.g. `"right_hand"`, `"head"`, `"boots"`).

## Camera

`_add_camera(root, Vector3(3.2, 2.2, 4.5), Vector3(0.0, 1.1, 0.0), 3.2)` — pulled back
and raised slightly compared to `gear` focus to accommodate the spread arms.

## render_focus.py changes

- Add `"skeleton"` to the `choices` list.
- Add size override: if focus is `"skeleton"` and default size, bump to `800×600`.
- Pass `--class-id` through normally (already handled).

## SKILL.md catalog entry

Under **Equipment and models**:

| `skeleton` | Character in spread pose, red dots at every bone + labeled socket spheres. | `--class-id` | 800×600 |

## Acceptance criteria

1. `python3 skills/showme/scripts/render_focus.py --focus skeleton --class-id paladin`
   produces a screenshot showing the paladin in T-spread pose.
2. Every bone (17 for paladin) has a visible red dot at its position.
3. All 10 equipment sockets have a colored sphere and a readable label.
4. `--class-id barbarian` (8 bones, no elbow/knee/foot) also works without errors.
5. The screenshot reveals whether `hand_r` / `hand_l` are at the visual hand position.
