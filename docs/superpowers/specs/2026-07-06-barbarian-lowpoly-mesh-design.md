# Design: Low-Poly Humanoid Base Mesh (Barbarian)

**Date:** 2026-07-06
**Scope:** `tools/assets/gen_glb.py` — `barbarian_glb()` only
**Constraint:** 17-bone skeleton, bone translations, and skinning weights are frozen. Only geometry changes.

---

## Goal

Replace the current block-cube body segments of the barbarian with low-poly frustum/prism shapes to achieve the faceted humanoid silhouette from the reference (visible face normals, tapered torso, cylindrical limbs). All other characters, monsters, and equipment are untouched.

---

## Frozen: 17-bone skeleton

The joints list in `_full_humanoid_glb` stays byte-for-byte identical. Global bone world positions for reference:

| Bone | Index | World position |
|------|-------|----------------|
| root | 0 | (0.000, 0.000, 0.159) |
| spine | 1 | (0.000, 1.133, 0.159) |
| chest | 2 | (0.000, 1.340, 0.159) |
| neck | 3 | (0.000, 1.674, 0.159) |
| head | 4 | (0.000, 1.812, 0.159) |
| arm_l | 5 | (-0.351, 1.440, 0.030) |
| elbow_l | 6 | (-0.281, 1.131, 0.040) |
| hand_l | 7 | (-0.223, 0.877, 0.048) |
| arm_r | 8 | (0.351, 1.440, 0.030) |
| elbow_r | 9 | (0.281, 1.131, 0.040) |
| hand_r | 10 | (0.223, 0.877, 0.048) |
| leg_l | 11 | (-0.155, 0.887, 0.159) |
| knee_l | 12 | (-0.155, 0.493, 0.159) |
| foot_l | 13 | (-0.232, 0.039, 0.105) |
| leg_r | 14 | (0.155, 0.887, 0.159) |
| knee_r | 15 | (0.155, 0.493, 0.159) |
| foot_r | 16 | (0.232, 0.039, 0.105) |

---

## New primitive: `_prism_geom(n, r_bot, r_top, h)`

Generates an n-sided frustum centered at origin on the Y axis:
- Bottom cap at y = -h/2, radius r_bot
- Top cap at y = +h/2, radius r_top
- Each side face has its own flat normal (produces faceted low-poly shading)
- Returns `(positions, normals, indices)` — same contract as `_cube_geometry()`

`r_bot == r_top` → regular prism. `r_bot != r_top` → frustum (taper).

n=8 for body/head, n=6 for limbs.

---

## Part format extension in `_build_skinned_glb`

Existing 3-element and 4-element (with color) box parts are untouched.

New 5-element format adds a custom geometry tuple as the last element:
```
(joint_idx, (tx, ty, tz), None, color, (positions, normals, indices))
```
When the 5th element is present, the function uses it directly instead of the unit cube. `tx,ty,tz` still apply as mesh-space translation offset.

---

## Helper: `_place_geom(geom, tx, ty, tz)`

Translates all positions in `(positions, normals, indices)` by `(tx, ty, tz)` and returns the translated copy. Used to place centered frustums at their bone-midpoint mesh positions.

---

## Body segments — centers and sizes preserved

Each frustum center matches the midpoint between the controlling bone pair (same as current box center). Radii are derived from the current box half-widths so the overall silhouette envelope is the same.

| Segment | Bone | Center (mesh space) | Shape | n | r_bot | r_top | h |
|---------|------|--------------------|----|---|-------|-------|---|
| Head | 4 | (0.000, 1.762, 0.159) | prism | 8 | 0.13 | 0.13 | 0.28 |
| Upper torso | 2 | (0.000, 1.507, 0.159) | frustum | 8 | 0.19 | 0.23 | 0.33 |
| Lower torso | 1 | (0.000, 1.236, 0.159) | frustum | 8 | 0.17 | 0.19 | 0.32 |
| Hips | 1 | (0.000, 0.980, 0.159) | frustum | 8 | 0.15 | 0.17 | 0.18 |
| Upper arm R | 8 | (0.316, 1.285, 0.035) | frustum | 6 | 0.045 | 0.055 | 0.32 |
| Forearm R | 9 | (0.252, 1.004, 0.044) | frustum | 6 | 0.036 | 0.045 | 0.26 |
| Hand R | 10 | (0.223, 0.877, 0.048) | box | — | — | — | — |
| Upper arm L | 5 | (-0.316, 1.285, 0.035) | frustum | 6 | 0.045 | 0.055 | 0.32 |
| Forearm L | 6 | (-0.252, 1.004, 0.044) | frustum | 6 | 0.036 | 0.045 | 0.26 |
| Hand L | 7 | (-0.223, 0.877, 0.048) | box | — | — | — | — |
| Thigh R | 14 | (0.155, 0.690, 0.159) | frustum | 6 | 0.062 | 0.075 | 0.40 |
| Shin R | 15 | (0.194, 0.266, 0.132) | frustum | 6 | 0.048 | 0.062 | 0.46 |
| Foot R | 16 | (0.232, 0.039, 0.105) | box | — | — | — | — |
| Thigh L | 11 | (-0.155, 0.690, 0.159) | frustum | 6 | 0.062 | 0.075 | 0.40 |
| Shin L | 12 | (-0.194, 0.266, 0.132) | frustum | 6 | 0.048 | 0.062 | 0.46 |
| Foot L | 13 | (-0.232, 0.039, 0.105) | box | — | — | — | — |

Color: `(0.66, 0.36, 0.25, 1.0)` — unchanged.

---

## What does NOT change

- `_full_humanoid_glb` joints list — untouched
- All other characters (sorcerer, paladin, rogue, ranger)
- All monsters and equipment
- `_build_glb`, `_humanoid_glb`, `base_humanoid_glb` — untouched
- Skinning weights (still 100% to single bone per part)
- Inverse bind matrices
- glTF schema structure

---

## Validation

After regeneration: `make validate-assets` must pass. Manual check in Godot that bones sit inside mesh and BoneAttachment3D sockets (hand_r, hand_l) are at correct world positions.
