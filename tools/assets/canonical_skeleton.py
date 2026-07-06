"""Frozen 17-bone humanoid bind pose — shared by gen_glb and canonical hero re-skin."""
from __future__ import annotations

import math
import statistics

# Matches gen_glb._full_humanoid_glb joints (translation-only bind pose).
CANONICAL_JOINTS: list[tuple[str, int, tuple[float, float, float]]] = [
    ("root", -1, (0.0, 0.0, 0.1588)),
    ("spine", 0, (0.0, 1.1327, 0.0)),
    ("chest", 1, (0.0, 0.2068, 0.0)),
    ("neck", 2, (0.0, 0.3349, 0.0)),
    ("head", 3, (0.0, 0.1379, 0.0)),
    ("arm_l", 2, (-0.3506, 0.1009, -0.129)),
    ("elbow_l", 5, (0.07, -0.3099, 0.0099)),
    ("hand_l", 6, (0.058, -0.2539, 0.0079)),
    ("arm_r", 2, (0.3506, 0.1009, -0.129)),
    ("elbow_r", 8, (-0.07, -0.3099, 0.0099)),
    ("hand_r", 9, (-0.058, -0.2539, 0.0079)),
    ("leg_l", 0, (-0.1551, 0.8865, 0.0)),
    ("knee_l", 11, (0.0, -0.394, 0.0)),
    ("foot_l", 12, (-0.077, -0.454, -0.054)),
    ("leg_r", 0, (0.1551, 0.8865, 0.0)),
    ("knee_r", 14, (0.0, -0.394, 0.0)),
    ("foot_r", 15, (0.077, -0.454, -0.054)),
]

BONE_SEGMENTS: list[tuple[int, int]] = [
    (0, 1), (1, 2), (2, 3), (3, 4),
    (2, 5), (5, 6), (6, 7),
    (2, 8), (8, 9), (9, 10),
    (0, 11), (11, 12), (12, 13),
    (0, 14), (14, 15), (15, 16),
]

_INFLUENCE_RADIUS = 0.025


def joint_globals() -> list[tuple[float, float, float]]:
    """World-space bind positions for all 17 joints."""
    globals_: list[tuple[float, float, float]] = []
    for _name, parent, local in CANONICAL_JOINTS:
        if parent < 0:
            globals_.append(local)
            continue

        px, py, pz = globals_[parent]
        lx, ly, lz = local
        globals_.append((px + lx, py + ly, pz + lz))

    return globals_


def _closest_on_segment(
    pos: tuple[float, float, float],
    a: tuple[float, float, float],
    b: tuple[float, float, float],
) -> tuple[float, float]:
    """Return param t in [0,1] and squared distance from pos to segment a→b."""
    abx = b[0] - a[0]
    aby = b[1] - a[1]
    abz = b[2] - a[2]
    apx = pos[0] - a[0]
    apy = pos[1] - a[1]
    apz = pos[2] - a[2]
    h2 = abx * abx + aby * aby + abz * abz
    if h2 < 1e-12:
        dx = pos[0] - a[0]
        dy = pos[1] - a[1]
        dz = pos[2] - a[2]
        return 0.0, dx * dx + dy * dy + dz * dz

    t = (apx * abx + apy * aby + apz * abz) / h2
    t = max(0.0, min(1.0, t))
    cx = a[0] + t * abx
    cy = a[1] + t * aby
    cz = a[2] + t * abz
    dx = pos[0] - cx
    dy = pos[1] - cy
    dz = pos[2] - cz
    return t, dx * dx + dy * dy + dz * dz


def vertex_weights(
    pos: tuple[float, float, float],
    globals_: list[tuple[float, float, float]] | None = None,
) -> tuple[tuple[int, int, int, int], tuple[float, float, float, float]]:
    """Up to four joint influences from inverse-distance segment weighting."""
    if globals_ is None:
        globals_ = joint_globals()

    scores = [0.0] * len(globals_)
    for ja, jb in BONE_SEGMENTS:
        t, dist2 = _closest_on_segment(pos, globals_[ja], globals_[jb])
        influence = 1.0 / (math.sqrt(dist2) + _INFLUENCE_RADIUS) ** 2
        scores[ja] += influence * (1.0 - t)
        scores[jb] += influence * t

    ranked = sorted(enumerate(scores), key=lambda item: -item[1])
    top = [(joint, score) for joint, score in ranked if score > 1e-9][:4]
    if not top:
        return (0, 0, 0, 0), (1.0, 0.0, 0.0, 0.0)

    total = sum(score for _, score in top)
    joints = [0, 0, 0, 0]
    weights = [0.0, 0.0, 0.0, 0.0]
    for idx, (joint, score) in enumerate(top):
        joints[idx] = joint
        weights[idx] = score / total

    return tuple(joints), tuple(weights)


def bind_pose_translation(
    mins: list[float],
    maxs: list[float],
    globals_: list[tuple[float, float, float]] | None = None,
) -> tuple[float, float, float]:
    """Translate mesh bbox to align with canonical root XZ and foot Y."""
    if globals_ is None:
        globals_ = joint_globals()

    target_cx = globals_[0][0]
    target_cz = globals_[0][2]
    target_foot_y = min(globals_[13][1], globals_[16][1])
    mesh_cx = (mins[0] + maxs[0]) * 0.5
    mesh_cz = (mins[2] + maxs[2]) * 0.5
    return (
        target_cx - mesh_cx,
        target_foot_y - mins[1],
        target_cz - mesh_cz,
    )


def _cluster_mean(
    positions: list[tuple[float, float, float]],
    predicate,
) -> tuple[float, float, float] | None:
    pts = [p for p in positions if predicate(p)]
    if not pts:
        return None

    return (
        statistics.mean(p[0] for p in pts),
        statistics.mean(p[1] for p in pts),
        statistics.mean(p[2] for p in pts),
    )


def joint_globals_from_mesh(
    all_positions: list[tuple[float, float, float]],
    mins: list[float],
    maxs: list[float],
) -> list[tuple[float, float, float]]:
    """Canonical torso/legs with A-pose arm and foot landmarks from mesh vertices."""
    g = list(joint_globals())
    cx = (mins[0] + maxs[0]) * 0.5
    w = max(maxs[0] - mins[0], 1e-6)
    h = max(maxs[1] - mins[1], 1e-6)
    y0 = mins[1]

    def _side_x(p: tuple[float, float, float], side: int) -> float:
        return (p[0] - cx) * side

    for side, arm_i, elbow_i, hand_i in ((-1, 5, 6, 7), (1, 8, 9, 10)):
        shoulder = _cluster_mean(
            all_positions,
            lambda p, s=side: _side_x(p, s) > w * 0.20 and y0 + h * 0.68 < p[1] < y0 + h * 0.84,
        )
        elbow = _cluster_mean(
            all_positions,
            lambda p, s=side: _side_x(p, s) > w * 0.28 and y0 + h * 0.53 < p[1] < y0 + h * 0.69,
        )
        hand = _cluster_mean(
            all_positions,
            lambda p, s=side: _side_x(p, s) > w * 0.32 and y0 + h * 0.38 < p[1] < y0 + h * 0.58,
        )
        if shoulder is not None:
            g[arm_i] = shoulder
        if elbow is not None:
            g[elbow_i] = elbow
        if hand is not None:
            g[hand_i] = hand

    for side, knee_i, foot_i in ((-1, 12, 13), (1, 15, 16)):
        knee = _cluster_mean(
            all_positions,
            lambda p, s=side: _side_x(p, s) > w * 0.02 and y0 + h * 0.18 < p[1] < y0 + h * 0.32,
        )
        foot = _cluster_mean(
            all_positions,
            lambda p, s=side: _side_x(p, s) > 0.0 and p[1] < y0 + h * 0.08,
        )
        if knee is not None:
            g[knee_i] = knee
        if foot is not None:
            g[foot_i] = foot

    return g

