# Design: Connected Barbarian Mesh (Cap Suppression + Shoulder Raise)

**Date:** 2026-07-06
**Scope:** `tools/assets/gen_glb.py` — `_prism_geom` + `barbarian_glb()` only
**Constraint:** 17-bone skeleton frozen. All other characters/monsters/equipment untouched.

---

## Goal

Eliminate the visible flat ring separations between adjacent body segments by suppressing cap faces at internal joints and aligning boundary Y positions and radii. Raise the shoulder placement so arms visually connect to the chest.

---

## Change 1: `_prism_geom` cap parameters

Add two boolean keyword arguments:

```python
def _prism_geom(n: int, r_bot: float, r_top: float, h: float,
                cap_bot: bool = True, cap_top: bool = True):
```

- `cap_bot=False` → skip generating the bottom disc faces entirely
- `cap_top=False` → skip generating the top disc faces entirely
- Default `True` preserves existing behaviour for all current callers

Vertex/index counts change accordingly:
- With both caps: `6n + 2` verts, `12n` indices (unchanged)
- Missing one cap: `5n + 1` verts, `9n` indices
- Missing both caps: `4n` verts, `6n` indices (side faces only)

Tests must be updated to account for the new parameters without breaking the default-cap assertions.

---

## Change 2: Flush boundary rules

At every internal joint between two adjacent frustums, both conditions must hold:

1. **Y flush**: `lower_frustum_top_Y == upper_frustum_bottom_Y` (same value, no gap, no overlap)
2. **Radius match**: `lower_frustum.r_top == upper_frustum.r_bot`
3. **Cap suppression**: `lower_frustum cap_top=False`, `upper_frustum cap_bot=False`

Exposed surfaces (top of head, bottom of hips, shoulder top, wrist, ankle, knee) keep `cap=True`.

---

## Change 3: New torso stack (connected)

Defined by shared boundary Y positions:

| Boundary | Y value | Shared radius |
|----------|---------|---------------|
| Head top (exposed) | 1.902 | — |
| Head/neck | 1.700 | 0.16 |
| Neck/chest | 1.550 | 0.23 |
| Chest/spine | 1.340 | 0.19 |
| Spine/hips | 1.050 | 0.17 |
| Hips bottom (exposed) | 0.889 | — |

Derived segment parameters (cy = midpoint, h = span):

| Part | Bone | cy | h | r_bot | r_top | cap_bot | cap_top |
|------|------|----|---|-------|-------|---------|---------|
| Hips | 1 | 0.970 | 0.161 | 0.17 | 0.17 | True | False |
| Lower torso | 1 | 1.195 | 0.290 | 0.17 | 0.19 | False | False |
| Upper torso | 2 | 1.445 | 0.210 | 0.19 | 0.23 | False | False |
| Neck | 3 | 1.625 | 0.150 | 0.23 | 0.16 | False | False |
| Head | 4 | 1.801 | 0.202 | 0.16 | 0.13 | False | True |

All use n=8. Z offset stays 0.159 throughout (matches existing skeleton Z).

---

## Change 4: Connected arm stack (shoulder raise)

Arm boundaries:

| Boundary | Y value |
|----------|---------|
| Shoulder top (exposed) | 1.550 (= neck/chest boundary) |
| Elbow | 1.131 (elbow_r/elbow_l bone world Y) |
| Wrist (exposed) | 0.877 (hand_r/hand_l bone world Y) |

| Part | Bone R/L | cy | h | r_bot | r_top | cap_bot | cap_top |
|------|----------|----|---|-------|-------|---------|---------|
| Upper arm R | 8 | (0.316, 1.341, 0.035) | 0.419 | 0.042 | 0.060 | False | True |
| Upper arm L | 5 | (-0.316, 1.341, 0.035) | 0.419 | 0.042 | 0.060 | False | True |
| Forearm R | 9 | (0.252, 1.004, 0.044) | 0.254 | 0.036 | 0.042 | True | False |
| Forearm L | 6 | (-0.252, 1.004, 0.044) | 0.254 | 0.036 | 0.042 | True | False |
| Hand R/L | 10/7 | unchanged | box | — | — | — | — |

Elbow joint: upper arm `cap_bot=False`, forearm `cap_top=False`, radii match at 0.042.

---

## Change 5: Connected leg stack (knee joint)

Leg boundaries:

| Boundary | Y value |
|----------|---------|
| Hip joint (exposed) | 0.889 (= hips bottom) |
| Knee | 0.493 (knee_r/knee_l bone world Y) |
| Ankle (exposed) | 0.039 (foot_r/foot_l bone world Y) |

| Part | Bone R/L | cy | h | r_bot | r_top | cap_bot | cap_top |
|------|----------|----|---|-------|-------|---------|---------|
| Thigh R | 14 | (0.155, 0.691, 0.159) | 0.396 | 0.062 | 0.075 | False | True |
| Thigh L | 11 | (-0.155, 0.691, 0.159) | 0.396 | 0.062 | 0.075 | False | True |
| Shin R | 15 | (0.194, 0.266, 0.132) | 0.454 | 0.048 | 0.062 | True | False |
| Shin L | 12 | (-0.194, 0.266, 0.132) | 0.454 | 0.048 | 0.062 | True | False |
| Foot R/L | 16/13 | unchanged | box | — | — | — | — |

Knee joint: thigh `cap_bot=False`, shin `cap_top=False`, radii match at 0.062.

---

## What does NOT change

- `_full_humanoid_glb` joints list (17 bones, all translations frozen)
- All other characters, monsters, equipment
- `_build_skinned_glb`, `_build_glb`, `_humanoid_glb`
- Skin color `(0.66, 0.36, 0.25, 1.0)`
- Maintainability baseline (line count stays within ±25 of current 642)

---

## Validation

- `make test-py` (5 tests pass, update vertex/index count tests for cap params)
- `make validate-assets`
- `make ci`
- Visual: `make model model=character_barbarian_v0` — body segments flow without horizontal rings; arm tops are at chest level
