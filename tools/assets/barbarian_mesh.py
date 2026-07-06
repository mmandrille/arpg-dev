"""Procedural barbarian body mesh — parts list for _full_humanoid_glb."""
from __future__ import annotations

from tools.assets.geom_primitives import (
    _extrapolate,
    _lerp3,
    _prism_between,
    _prism_ellipse_geom,
    _prism_geom,
)
from tools.assets.skin_blend import blend_lateral, blend_segment, blend_segment_triple, blend_shoulder

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
    n_head = max(n, 16)
    return [
        _part(_NECK, (0.0, neck_cy, _Z), _SKIN,
              _prism_ellipse_geom(n_head, 0.12, 0.10, 0.14, 0.11, neck_top - neck_bot, False, False),
              blend_segment((0.0, neck_bot, _Z), (0.0, neck_top, _Z), _NECK, _CHEST, split=0.38)),
        _part(_HEAD, (0.0, skull_cy, _Z), _SKIN,
              _prism_ellipse_geom(n_head, 0.138, 0.112, 0.118, 0.098, skull_top - skull_bot, False, True),
              blend_segment((0.0, skull_bot, _Z), (0.0, skull_top, _Z), _HEAD, _NECK, split=0.32)),
        # Jaw/chin — shifted forward (+Z) for a readable face silhouette.
        _part(_HEAD, (0.0, 1.718, 0.188), _SKIN,
              _prism_ellipse_geom(10, 0.112, 0.078, 0.098, 0.068, 0.105, False, True)),
        # Chin point — stronger lower-face read.
        _part(_HEAD, (0.0, 1.698, 0.198), _SKIN,
              _prism_ellipse_geom(6, 0.072, 0.052, 0.058, 0.042, 0.065, False, True)),
        # Brow ridge — shallow cap above eyes.
        _part(_HEAD, (0.0, 1.852, 0.168), _SKIN,
              _prism_ellipse_geom(10, 0.125, 0.095, 0.110, 0.085, 0.055, False, True)),
    ]


def _head_eyes():
    """Eye-socket recesses — lateral indent at brow height."""
    eye_y = 1.818
    eye_z = 0.178
    return [
        _part(_HEAD, (0.052, eye_y, eye_z), _SKIN,
              _prism_ellipse_geom(6, 0.028, 0.022, 0.022, 0.018, 0.038, False, False)),
        _part(_HEAD, (-0.052, eye_y, eye_z), _SKIN,
              _prism_ellipse_geom(6, 0.028, 0.022, 0.022, 0.018, 0.038, False, False)),
    ]


def _head_facial(side_sign: float = 1.0):
    """Nose bridge and ear nubs — readable face at gameplay scale."""
    ear_x = 0.128 * side_sign
    return [
        _part(_HEAD, (0.0, 1.792, 0.208), _SKIN,
              _prism_ellipse_geom(6, 0.032, 0.028, 0.022, 0.018, 0.072, False, True)),
        _part(_HEAD, (ear_x, 1.778, 0.152), _SKIN,
              _prism_ellipse_geom(5, 0.028, 0.038, 0.022, 0.032, 0.055, False, False)),
    ]


def _torso_anatomy():
    """Pecs, clavicles, lats, and a front abdominal ridge."""
    clav_r0 = (0.055, 1.565, 0.170)
    clav_r1 = (0.230, 1.525, 0.152)
    clav_l0 = (-0.055, 1.565, 0.170)
    clav_l1 = (-0.230, 1.525, 0.152)
    return [
        # Pectorals — forward (+Z) chest mass.
        _part(_CHEST, (0.105, 1.418, 0.200), _SKIN,
              _prism_ellipse_geom(8, 0.095, 0.072, 0.085, 0.062, 0.115, False, True)),
        _part(_CHEST, (-0.105, 1.418, 0.200), _SKIN,
              _prism_ellipse_geom(8, 0.095, 0.072, 0.085, 0.062, 0.115, False, True)),
        # Clavicles — collar line into the shoulder.
        _part(_CHEST, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(clav_r0, clav_r1, 5, 0.038, 0.032, False, False),
              blend_segment(clav_r0, clav_r1, _CHEST, _ARM_R, split=0.65)),
        _part(_CHEST, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(clav_l0, clav_l1, 5, 0.038, 0.032, False, False),
              blend_segment(clav_l0, clav_l1, _CHEST, _ARM_L, split=0.65)),
        # Lat flanks — side taper on lower ribcage.
        _part(_SPINE, (0.145, 1.220, 0.150), _SKIN,
              _prism_geom(6, 0.068, 0.056, 0.220, False, False)),
        _part(_SPINE, (-0.145, 1.220, 0.150), _SKIN,
              _prism_geom(6, 0.068, 0.056, 0.220, False, False)),
        # Lower belly ridge.
        _part(_SPINE, (0.0, 1.085, 0.180), _SKIN,
              _prism_ellipse_geom(6, 0.058, 0.048, 0.048, 0.038, 0.130, False, False)),
        # Rib hints — shallow side arcs on upper abdomen.
        _part(_SPINE, (0.118, 1.268, 0.162), _SKIN,
              _prism_geom(5, 0.042, 0.034, 0.095, False, False)),
        _part(_SPINE, (-0.118, 1.268, 0.162), _SKIN,
              _prism_geom(5, 0.042, 0.034, 0.095, False, False)),
    ]


def _waist_bridge():
    """Pelvis ring — closes the torso-to-hip gap."""
    return [
        _part(_SPINE, (0.0, 0.978, 0.148), _SKIN,
              _prism_ellipse_geom(10, 0.165, 0.135, 0.172, 0.142, 0.095, False, False),
              blend_lateral(_LEG_R_I, _SPINE, "x", 0.0, 0.06, 0.14)),
    ]


def _limb_joint_pads():
    """Bridge prisms at elbows/knees — spans open-cap seams along the bone axis."""
    def _bridge(joint: int, p_in: tuple, p_out: tuple, r: float):
        return _part(
            joint, (0.0, 0.0, 0.0), _SKIN,
            _prism_between(p_in, p_out, 8, r, r, False, False),
        )

    elbow_r_in = _lerp3(_WP_ELBOW_R, _CHEST_ANCHOR_R, 0.10)
    elbow_r_out = _lerp3(_WP_ELBOW_R, _WP_HAND_R, 0.12)
    elbow_l_in = _lerp3(_WP_ELBOW_L, _CHEST_ANCHOR_L, 0.10)
    elbow_l_out = _lerp3(_WP_ELBOW_L, _WP_HAND_L, 0.12)
    knee_r_in = _lerp3(_WP_KNEE_R, _HIP_ANCHOR_R, 0.10)
    knee_r_out = _lerp3(_WP_KNEE_R, _WP_FOOT_R, 0.12)
    knee_l_in = _lerp3(_WP_KNEE_L, _HIP_ANCHOR_L, 0.10)
    knee_l_out = _lerp3(_WP_KNEE_L, _WP_FOOT_L, 0.12)
    return [
        _bridge(_ELBOW_R_I, elbow_r_in, elbow_r_out, 0.058),
        _bridge(_ELBOW_L_I, elbow_l_in, elbow_l_out, 0.058),
        _bridge(_KNEE_R_I, knee_r_in, knee_r_out, 0.070),
        _bridge(_KNEE_L_I, knee_l_in, knee_l_out, 0.070),
    ]


def _shoulder_deltoids(side_sign: float = 1.0):
    """Outer deltoid caps — triple-weighted shoulder deformation."""
    arm = _ARM_R if side_sign > 0 else _ARM_L
    anchor = _CHEST_ANCHOR_R if side_sign > 0 else _CHEST_ANCHOR_L
    arm_pos = _WP_ARM_R if side_sign > 0 else _WP_ARM_L
    x = 0.198 * side_sign
    shoulder_blend = blend_shoulder(
        arm, _CHEST, _SPINE, "x", 0.0, 0.06, 0.20, back_z=0.118,
    )
    delt_top = _lerp3(anchor, arm_pos, 0.35)
    delt_bot = _lerp3(anchor, arm_pos, 0.62)
    return [
        _part(_CHEST, (x, 1.508, 0.138), _SKIN,
              _prism_geom(8, 0.118, 0.102, 0.115, False, False),
              shoulder_blend),
        _part(_CHEST, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(delt_top, delt_bot, 8, 0.105, 0.088, False, False),
              blend_segment(delt_top, delt_bot, arm, _CHEST, split=0.42)),
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
              _prism_geom(8, 0.135, 0.115, 0.130, False, False),
              blend_shoulder(_ARM_R, _CHEST, _SPINE, "x", 0.0, 0.08, 0.20, back_z=0.120)),
        _part(_CHEST, (-0.185, 1.485, _Z), _SKIN,
              _prism_geom(8, 0.135, 0.115, 0.130, False, False),
              blend_shoulder(_ARM_L, _CHEST, _SPINE, "x", 0.0, 0.08, 0.20, back_z=0.120)),
        _part(_CHEST, (0.155, 1.500, 0.108), _SKIN,
              _prism_geom(8, 0.120, 0.095, 0.110, False, False),
              blend_shoulder(_ARM_R, _CHEST, _SPINE, "x", 0.0, 0.06, 0.18, back_z=0.112)),
        _part(_CHEST, (-0.155, 1.500, 0.108), _SKIN,
              _prism_geom(8, 0.120, 0.095, 0.110, False, False),
              blend_shoulder(_ARM_L, _CHEST, _SPINE, "x", 0.0, 0.06, 0.18, back_z=0.112)),
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
               arm: tuple, elbow: tuple, hand: tuple, chest_anchor: tuple,
               side_sign: float = 1.0):
    """Upper arm with bicep bulge; forearm + palm + thumb."""
    n_limb = 10
    mid_upper = _lerp3(chest_anchor, elbow, 0.58)
    bicep = (
        mid_upper[0] + side_sign * 0.028,
        mid_upper[1] - 0.018,
        mid_upper[2] + 0.012,
    )
    palm = _extrapolate(elbow, hand, 0.40)
    palm_stub = _extrapolate(elbow, hand, 0.12)
    thumb_tip = (hand[0] + side_sign * 0.042, hand[1] - 0.018, hand[2] + 0.028)
    shoulder_out = _lerp3(chest_anchor, arm, 0.18)
    return [
        _part(arm_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(chest_anchor, shoulder_out, 8, 0.088, 0.082, False, False),
              blend_shoulder(arm_bone, _CHEST, _SPINE, "x", 0.0, 0.04, 0.16, back_z=0.122)),
        _part(arm_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(shoulder_out, mid_upper, n_limb, 0.092, 0.078, False, False),
              blend_shoulder(arm_bone, _CHEST, _SPINE, "x", 0.0, 0.05, 0.18, back_z=0.125)),
        _part(arm_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(mid_upper, elbow, n_limb, 0.080, 0.052, False, False),
              blend_segment(mid_upper, elbow, arm_bone, _CHEST, split=0.35)),
        # Bicep peak — lateral outer bulge on upper arm.
        _part(arm_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(bicep, _lerp3(bicep, elbow, 0.55), 6, 0.048, 0.038, False, False)),
        _part(elbow_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(elbow, hand, n_limb, 0.052, 0.040, False, False),
              blend_segment_triple(elbow, hand, arm_bone, elbow_bone, hand_bone)),
        _part(hand_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(hand, palm, 6, 0.042, 0.034, True, False),
              blend_segment(hand, palm, hand_bone, elbow_bone, split=0.35)),
        _part(hand_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(hand, palm_stub, 6, 0.050, 0.044, False, False)),
        _part(hand_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(hand, thumb_tip, 5, 0.022, 0.018, True, False)),
    ] + _hand_fingers(hand_bone, palm, side_sign)


def _hand_fingers(hand_bone: int, palm: tuple, side_sign: float):
    """Three knuckle stubs — breaks the mitten silhouette."""
    fingers = []
    for spread in (-0.024, 0.0, 0.024):
        tip = (palm[0] + side_sign * spread, palm[1] - 0.022, palm[2] + 0.058)
        fingers.append(
            _part(hand_bone, (0.0, 0.0, 0.0), _SKIN,
                  _prism_between(palm, tip, 5, 0.020, 0.014, True, False)),
        )
    return fingers


def _leg_chain(leg_bone: int, knee_bone: int, foot_bone: int,
               leg: tuple, knee: tuple, foot: tuple, hip_anchor: tuple,
               side_sign: float = 1.0):
    """Thigh quad bulge; calf + heel/toe foot."""
    n_limb = 10
    mid_thigh = _lerp3(hip_anchor, knee, 0.52)
    mid_shin = (
        (knee[0] + foot[0]) / 2.0 + 0.018 * side_sign,
        (knee[1] + foot[1]) / 2.0,
        (knee[2] + foot[2]) / 2.0,
    )
    toe = _extrapolate(knee, foot, 0.32)
    heel = (foot[0] - 0.028 * side_sign, foot[1] + 0.048, foot[2] - 0.038)
    ankle_in = _lerp3(foot, mid_shin, 0.10)
    return [
        _part(leg_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(hip_anchor, mid_thigh, n_limb, 0.086, 0.080, False, False),
              blend_segment(hip_anchor, mid_thigh, leg_bone, _SPINE, split=0.42)),
        _part(leg_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(mid_thigh, knee, n_limb, 0.082, 0.072, False, False),
              blend_segment(mid_thigh, knee, leg_bone, _SPINE, split=0.30)),
        _part(knee_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(knee, mid_shin, n_limb, 0.072, 0.078, False, False),
              blend_segment_triple(knee, mid_shin, leg_bone, knee_bone, foot_bone, 0.32, 0.68)),
        _part(knee_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(mid_shin, foot, n_limb, 0.078, 0.056, False, False),
              blend_segment_triple(mid_shin, foot, leg_bone, knee_bone, foot_bone, 0.25, 0.70)),
        _part(foot_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(ankle_in, foot, 8, 0.058, 0.054, False, False),
              blend_segment(ankle_in, foot, foot_bone, knee_bone, split=0.45)),
        # Calf bulge — back-of-shin mass.
        _part(knee_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(
                  (mid_shin[0] - 0.012 * side_sign, mid_shin[1] + 0.018, mid_shin[2] - 0.022),
                  (mid_shin[0] - 0.010 * side_sign, mid_shin[1] - 0.042, mid_shin[2] - 0.028),
                  6, 0.048, 0.040, False, False),
              blend_segment_triple(knee, foot, leg_bone, knee_bone, foot_bone, 0.35, 0.65)),
        _part(foot_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(foot, toe, 8, 0.058, 0.048, True, False),
              blend_segment(foot, toe, foot_bone, knee_bone, split=0.40)),
        _part(foot_bone, (0.0, 0.0, 0.0), _SKIN,
              _prism_between(heel, foot, 6, 0.052, 0.056, False, False)),
    ]


def barbarian_parts() -> list:
    """All skinned mesh parts for the procedural barbarian."""
    n_body = 16
    parts = (
        _torso_stack(n_body)
        + _head_neck(n_body)
        + _head_eyes()
        + _head_facial(1.0)
        + _head_facial(-1.0)
        + _torso_anatomy()
        + _waist_bridge()
        + _shoulder_deltoids(1.0)
        + _shoulder_deltoids(-1.0)
        + _torso_silhouette()
        + _limb_joint_pads()
    )
    parts += _arm_chain(
        _ARM_R, _ELBOW_R_I, _HAND_R_I,
        _WP_ARM_R, _WP_ELBOW_R, _WP_HAND_R, _CHEST_ANCHOR_R, side_sign=1.0,
    )
    parts += _arm_chain(
        _ARM_L, _ELBOW_L_I, _HAND_L_I,
        _WP_ARM_L, _WP_ELBOW_L, _WP_HAND_L, _CHEST_ANCHOR_L, side_sign=-1.0,
    )
    parts += _leg_chain(
        _LEG_R_I, _KNEE_R_I, _FOOT_R_I,
        _WP_LEG_R, _WP_KNEE_R, _WP_FOOT_R, _HIP_ANCHOR_R, side_sign=1.0,
    )
    parts += _leg_chain(
        _LEG_L_I, _KNEE_L_I, _FOOT_L_I,
        _WP_LEG_L, _WP_KNEE_L, _WP_FOOT_L, _HIP_ANCHOR_L, side_sign=-1.0,
    )
    return parts
