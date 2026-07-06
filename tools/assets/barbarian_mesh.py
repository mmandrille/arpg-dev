"""Procedural barbarian body mesh — parts list for _full_humanoid_glb."""
from __future__ import annotations

from tools.assets.geom_primitives import (
    _extrapolate,
    _prism_between,
    _prism_ellipse_geom,
    _prism_geom,
)
from tools.assets.skin_blend import blend_lateral, blend_segment, blend_segment_triple

_SKIN = (0.66, 0.36, 0.25, 1.0)

# Frozen bind-pose world positions (matches _full_humanoid_glb joints).
_WP_ARM_R = (0.351, 1.440, 0.030)
_WP_ELBOW_R = (0.281, 1.131, 0.040)
_WP_HAND_R = (0.223, 0.877, 0.048)
_WP_ARM_L = (-0.351, 1.440, 0.030)
_WP_ELBOW_L = (-0.281, 1.131, 0.040)
_WP_HAND_L = (-0.223, 0.877, 0.048)
_WP_LEG_R = (0.155, 0.887, 0.159)
_WP_KNEE_R = (0.155, 0.493, 0.159)
_WP_FOOT_R = (0.232, 0.039, 0.105)
_WP_LEG_L = (-0.155, 0.887, 0.159)
_WP_KNEE_L = (-0.155, 0.493, 0.159)
_WP_FOOT_L = (-0.232, 0.039, 0.105)
_HIP_Y = 0.889
_Z = 0.159

# Bone indices (frozen _full_humanoid_glb map).
_SPINE = 1
_CHEST = 2
_NECK = 3
_HEAD = 4
_ARM_L = 5
_ARM_R = 8
_ELBOW_L_I = 6
_ELBOW_R_I = 9
_HAND_L_I = 7
_HAND_R_I = 10
_LEG_L_I = 11
_LEG_R_I = 14
_KNEE_L_I = 12
_KNEE_R_I = 15
_FOOT_L_I = 13
_FOOT_R_I = 16

_CHEST_ANCHOR_R = (0.255, 1.485, 0.125)
_CHEST_ANCHOR_L = (-0.255, 1.485, 0.125)
_HIP_ANCHOR_R = (_WP_LEG_R[0], _HIP_Y, 0.145)
_HIP_ANCHOR_L = (_WP_LEG_L[0], _HIP_Y, 0.145)


def _part(joint: int, offset, color, geom, weight_fn=None):
    if weight_fn is None:
        return (joint, offset, None, color, geom)
    return (joint, offset, None, color, geom, weight_fn)


def _torso_stack(n_body: int):
    """Elliptical torso: wider side-to-side (rx) than front-to-back (rz)."""
    return [
        _part(_SPINE, (0.0, 0.970, _Z), _SKIN,
              _prism_ellipse_geom(n_body, 0.18, 0.13, 0.18, 0.13, 0.161, True, False)),
        _part(_SPINE, (0.0, 1.195, _Z), _SKIN,
              _prism_ellipse_geom(n_body, 0.18, 0.13, 0.20, 0.14, 0.290, False, False)),
        _part(_CHEST, (0.0, 1.445, _Z), _SKIN,
              _prism_ellipse_geom(n_body, 0.20, 0.14, 0.25, 0.16, 0.210, False, False),
              blend_lateral(_ARM_R, _CHEST, "x", 0.0, 0.14, 0.24)),
    ]


def _head_neck(n: int):
    """Skull + jaw read as a face; neck blends into chest."""
    neck_bot = 1.550
    neck_top = 1.700
    neck_cy = (neck_bot + neck_top) / 2.0
    skull_bot = 1.695
    skull_top = 1.905
    skull_cy = (skull_bot + skull_top) / 2.0
    return [
        _part(_NECK, (0.0, neck_cy, _Z), _SKIN,
              _prism_ellipse_geom(n, 0.12, 0.10, 0.14, 0.11, neck_top - neck_bot, False, False),
              blend_segment((0.0, neck_bot, _Z), (0.0, neck_top, _Z), _NECK, _CHEST, split=0.38)),
        _part(_HEAD, (0.0, skull_cy, _Z), _SKIN,
              _prism_ellipse_geom(n, 0.138, 0.112, 0.118, 0.098, skull_top - skull_bot, False, True),
              blend_segment((0.0, skull_bot, _Z), (0.0, skull_top, _Z), _HEAD, _NECK, split=0.32)),
        # Jaw/chin — shifted forward (+Z) for a readable face silhouette.
        _part(_HEAD, (0.0, 1.728, 0.182), _SKIN,
              _prism_ellipse_geom(8, 0.105, 0.075, 0.090, 0.065, 0.095, False, True)),
        # Brow ridge — shallow cap above eyes.
        _part(_HEAD, (0.0, 1.852, 0.168), _SKIN,
              _prism_ellipse_geom(8, 0.125, 0.095, 0.110, 0.085, 0.055, False, True)),
    ]


def _torso_silhouette():
    """Shoulder blades, glutes, and hip sockets — fills back/side gaps."""
    bridge_r_a = (0.195, 1.520, 0.148)
    bridge_l_a = (-0.195, 1.520, 0.148)
    bridge_r_b = (0.170, 1.505, 0.112)
    bridge_l_b = (-0.170, 1.505, 0.112)
    hip_bridge_r = (0.095, 0.915, 0.135)
    hip_bridge_l = (-0.095, 0.915, 0.135)
    return [
        _part(_CHEST, (0.185, 1.485, _Z), _SKIN,
              _prism_geom(6, 0.135, 0.115, 0.130, False, False),
              blend_lateral(_ARM_R, _CHEST, "x", 0.0, 0.10, 0.20)),
        _part(_CHEST, (-0.185, 1.485, _Z), _SKIN,
              _prism_geom(6, 0.135, 0.115, 0.130, False, False),
              blend_lateral(_ARM_L, _CHEST, "x", 0.0, 0.10, 0.20)),
        _part(_CHEST, (0.155, 1.500, 0.108), _SKIN,
              _prism_geom(6, 0.120, 0.095, 0.110, False, False),
              blend_lateral(_ARM_R, _CHEST, "x", 0.0, 0.08, 0.18)),
        _part(_CHEST, (-0.155, 1.500, 0.108), _SKIN,
              _prism_geom(6, 0.120, 0.095, 0.110, False, False),
              blend_lateral(_ARM_L, _CHEST, "x", 0.0, 0.08, 0.18)),
        _part(_CHEST, (0.0, 1.520, 0.105), _SKIN,
              _prism_ellipse_geom(6, 0.16, 0.10, 0.20, 0.12, 0.120, False, False)),
        _part(_CHEST, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(bridge_r_a, _WP_ARM_R, 6, 0.130, 0.092, False, False),
              blend_segment(bridge_r_a, _WP_ARM_R, _ARM_R, _CHEST, split=0.55)),
        _part(_CHEST, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(bridge_l_a, _WP_ARM_L, 6, 0.130, 0.092, False, False),
              blend_segment(bridge_l_a, _WP_ARM_L, _ARM_L, _CHEST, split=0.55)),
        _part(_CHEST, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(bridge_r_b, _WP_ARM_R, 6, 0.110, 0.090, False, False),
              blend_segment(bridge_r_b, _WP_ARM_R, _ARM_R, _CHEST, split=0.50)),
        _part(_CHEST, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(bridge_l_b, _WP_ARM_L, 6, 0.110, 0.090, False, False),
              blend_segment(bridge_l_b, _WP_ARM_L, _ARM_L, _CHEST, split=0.50)),
        _part(_SPINE, (0.0, 0.925, 0.118), _SKIN,
              _prism_ellipse_geom(8, 0.17, 0.14, 0.15, 0.12, 0.150, False, False),
              blend_lateral(_LEG_R_I, _SPINE, "x", 0.0, 0.08, 0.16)),
        _part(_SPINE, (0.105, 0.935, _Z), _SKIN,
              _prism_geom(6, 0.135, 0.100, 0.125, False, False),
              blend_segment((0.105, 0.935, _Z), _WP_LEG_R, _LEG_R_I, _SPINE, split=0.45)),
        _part(_SPINE, (-0.105, 0.935, _Z), _SKIN,
              _prism_geom(6, 0.135, 0.100, 0.125, False, False),
              blend_segment((-0.105, 0.935, _Z), _WP_LEG_L, _LEG_L_I, _SPINE, split=0.45)),
        _part(_SPINE, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(hip_bridge_r, _WP_LEG_R, 6, 0.115, 0.082, False, False),
              blend_segment(hip_bridge_r, _WP_LEG_R, _LEG_R_I, _SPINE, split=0.50)),
        _part(_SPINE, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(hip_bridge_l, _WP_LEG_L, 6, 0.115, 0.082, False, False),
              blend_segment(hip_bridge_l, _WP_LEG_L, _LEG_L_I, _SPINE, split=0.50)),
    ]


def _arm_chain(arm_bone: int, elbow_bone: int, hand_bone: int,
               arm: tuple, elbow: tuple, hand: tuple, chest_anchor: tuple):
    """Upper arm overlaps chest; forearm + palm follow bone line with joint blends."""
    n_limb = 8
    palm = _extrapolate(elbow, hand, 0.40)
    palm_stub = _extrapolate(elbow, hand, 0.12)
    return [
        _part(arm_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(chest_anchor, elbow, n_limb, 0.095, 0.046, False, True),
              blend_segment(chest_anchor, elbow, arm_bone, _CHEST, split=0.42)),
        _part(elbow_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(elbow, hand, n_limb, 0.046, 0.038, False, True),
              blend_segment_triple(elbow, hand, arm_bone, elbow_bone, hand_bone)),
        _part(hand_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(hand, palm, 6, 0.038, 0.032, True, False),
              blend_segment(hand, palm, hand_bone, elbow_bone, split=0.35)),
        _part(hand_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(hand, palm_stub, 6, 0.048, 0.042, False, False)),
    ]


def _leg_chain(leg_bone: int, knee_bone: int, foot_bone: int,
               leg: tuple, knee: tuple, foot: tuple, hip_anchor: tuple):
    """Thigh from hip socket; two-part shin for calf bulge with hip/knee blends."""
    n_limb = 8
    mid_shin = (
        (knee[0] + foot[0]) / 2.0 + 0.018,
        (knee[1] + foot[1]) / 2.0,
        (knee[2] + foot[2]) / 2.0,
    )
    toe = _extrapolate(knee, foot, 0.30)
    return [
        _part(leg_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(hip_anchor, knee, n_limb, 0.082, 0.068, False, True),
              blend_segment(hip_anchor, knee, leg_bone, _SPINE, split=0.40)),
        _part(knee_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(knee, mid_shin, n_limb, 0.068, 0.072, False, False),
              blend_segment_triple(knee, mid_shin, leg_bone, knee_bone, foot_bone, 0.32, 0.68)),
        _part(knee_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(mid_shin, foot, n_limb, 0.072, 0.052, False, True),
              blend_segment_triple(mid_shin, foot, leg_bone, knee_bone, foot_bone, 0.25, 0.70)),
        _part(foot_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(foot, toe, 6, 0.052, 0.044, True, False),
              blend_segment(foot, toe, foot_bone, knee_bone, split=0.40)),
    ]


def barbarian_parts() -> list:
    """All skinned mesh parts for the procedural barbarian."""
    n_body = 12
    parts = _torso_stack(n_body) + _head_neck(n_body) + _torso_silhouette()
    parts += _arm_chain(_ARM_R, _ELBOW_R_I, _HAND_R_I, _WP_ARM_R, _WP_ELBOW_R, _WP_HAND_R, _CHEST_ANCHOR_R)
    parts += _arm_chain(_ARM_L, _ELBOW_L_I, _HAND_L_I, _WP_ARM_L, _WP_ELBOW_L, _WP_HAND_L, _CHEST_ANCHOR_L)
    parts += _leg_chain(_LEG_R_I, _KNEE_R_I, _FOOT_R_I, _WP_LEG_R, _WP_KNEE_R, _WP_FOOT_R, _HIP_ANCHOR_R)
    parts += _leg_chain(_LEG_L_I, _KNEE_L_I, _FOOT_L_I, _WP_LEG_L, _WP_KNEE_L, _WP_FOOT_L, _HIP_ANCHOR_L)
    return parts
