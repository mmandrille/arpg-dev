"""Low-poly frustum primitives for deterministic GLB generation."""
from __future__ import annotations

import math


def _prism_geom(n: int, r_bot: float, r_top: float, h: float,
                cap_bot: bool = True, cap_top: bool = True):
    """Circular frustum — delegates to ellipse with rx=rz=r."""
    return _prism_ellipse_geom(n, r_bot, r_bot, r_top, r_top, h, cap_bot, cap_top)


def _prism_ellipse_geom(
    n: int,
    rx_bot: float, rz_bot: float,
    rx_top: float, rz_top: float,
    h: float,
    cap_bot: bool = True,
    cap_top: bool = True,
):
    """Elliptical frustum on Y (x=rx*cos, z=rz*sin) — deeper side profile than front/back."""
    pos, nrm, idx = [], [], []
    hh = h / 2.0
    for i in range(n):
        a0 = 2 * math.pi * i / n
        a1 = 2 * math.pi * (i + 1) / n
        am = (a0 + a1) / 2.0
        bl = (rx_bot * math.cos(a0), -hh, rz_bot * math.sin(a0))
        br = (rx_bot * math.cos(a1), -hh, rz_bot * math.sin(a1))
        tr = (rx_top * math.cos(a1),  hh, rz_top * math.sin(a1))
        tl = (rx_top * math.cos(a0),  hh, rz_top * math.sin(a0))
        drx = (rx_top - rx_bot) / h if h else 0.0
        drz = (rz_top - rz_bot) / h if h else 0.0
        nx = math.cos(am)
        nz = math.sin(am)
        raw = (nx, -(nx * drx + nz * drz), nz)
        nl = math.sqrt(sum(v * v for v in raw))
        fn = tuple(v / nl for v in raw)
        b = len(pos)
        for v in (bl, tl, tr, br):
            pos.append(v)
            nrm.append(fn)
        idx += [b, b + 1, b + 2, b, b + 2, b + 3]
    if cap_bot:
        cy = -hh
        cn = (0.0, -1.0, 0.0)
        c_bot = len(pos)
        pos.append((0.0, cy, 0.0))
        nrm.append(cn)
        bot_edges = []
        for i in range(n):
            a = 2 * math.pi * i / n
            bot_edges.append(len(pos))
            pos.append((rx_bot * math.cos(a), cy, rz_bot * math.sin(a)))
            nrm.append(cn)
        for i in range(n):
            idx += [c_bot, bot_edges[(i + 1) % n], bot_edges[i]]
    if cap_top:
        cy = hh
        cn = (0.0, 1.0, 0.0)
        c_top = len(pos)
        pos.append((0.0, cy, 0.0))
        nrm.append(cn)
        top_edges = []
        for i in range(n):
            a = 2 * math.pi * i / n
            top_edges.append(len(pos))
            pos.append((rx_top * math.cos(a), cy, rz_top * math.sin(a)))
            nrm.append(cn)
        for i in range(n):
            idx += [c_top, top_edges[i], top_edges[(i + 1) % n]]
    return pos, nrm, idx



def _rotate_to_y_dir(
    pos: list[tuple[float, float, float]],
    nrm: list[tuple[float, float, float]],
    ux: float, uy: float, uz: float,
) -> tuple[list[tuple[float, float, float]], list[tuple[float, float, float]]]:
    """Rotate geometry so local +Y aligns with unit direction (ux, uy, uz)."""
    if uy > 0.999:
        return pos, nrm

    if uy < -0.999:
        rot = [( -px, -py, -pz) for px, py, pz in pos]
        rnm = [( -nx, -ny, -nz) for nx, ny, nz in nrm]
        return rot, rnm

    # Y → dir via rotation in the plane spanned by Y and dir (axis = Y × dir).
    ax, ay, az = -uz, 0.0, ux
    al = math.sqrt(ax * ax + az * az)
    ax, az = ax / al, az / al
    ang = math.acos(max(-1.0, min(1.0, uy)))
    c, s = math.cos(ang), math.sin(ang)
    rot_pos, rot_nrm = [], []
    for px, py, pz in pos:
        y1 = c * py - s * pz
        z1 = s * py + c * pz
        x2 = c * px + s * z1
        z2 = -s * px + c * z1
        rot_pos.append((x2, y1, z2))
    for nx, ny, nz in nrm:
        y1 = c * ny - s * nz
        z1 = s * ny + c * nz
        x2 = c * nx + s * z1
        z2 = -s * nx + c * z1
        rot_nrm.append((x2, y1, z2))
    return rot_pos, rot_nrm


def _prism_between(
    p0: tuple[float, float, float],
    p1: tuple[float, float, float],
    n: int,
    r0: float,
    r1: float,
    cap_bot: bool = True,
    cap_top: bool = True,
):
    """Frustum from p0 (radius r0) to p1 (radius r1) in world space."""
    dx = p1[0] - p0[0]
    dy = p1[1] - p0[1]
    dz = p1[2] - p0[2]
    h = math.sqrt(dx * dx + dy * dy + dz * dz)
    if h < 1e-9:
        h = 1e-9
        ux, uy, uz = 0.0, 1.0, 0.0
    else:
        ux, uy, uz = dx / h, dy / h, dz / h

    pos, nrm, idx = _prism_geom(n, r0, r1, h, cap_bot=cap_bot, cap_top=cap_top)
    pos, nrm = _rotate_to_y_dir(pos, nrm, ux, uy, uz)
    cx = (p0[0] + p1[0]) / 2.0
    cy = (p0[1] + p1[1]) / 2.0
    cz = (p0[2] + p1[2]) / 2.0
    pos = [(px + cx, py + cy, pz + cz) for px, py, pz in pos]
    return pos, nrm, idx


def _extrapolate(p0: tuple[float, float, float], p1: tuple[float, float, float], t: float):
    """Point p1 + t * (p1 - p0)."""
    return (
        p1[0] + t * (p1[0] - p0[0]),
        p1[1] + t * (p1[1] - p0[1]),
        p1[2] + t * (p1[2] - p0[2]),
    )
