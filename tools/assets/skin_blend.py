"""Per-vertex skin weight helpers for procedural skinned meshes."""
from __future__ import annotations

from typing import Callable

WeightFn = Callable[[float, float, float], tuple[tuple[int, int, int, int], tuple[float, float, float, float]]]


def _pack(ja: int, jb: int, wa: float, wb: float) -> tuple[tuple[int, int, int, int], tuple[float, float, float, float]]:
    total = wa + wb
    if total < 1e-9:
        return (ja, 0, 0, 0), (1.0, 0.0, 0.0, 0.0)

    wa /= total
    wb /= total
    return (ja, jb, 0, 0), (wa, wb, 0.0, 0.0)


def _pack3(
    ja: int, jb: int, jc: int,
    wa: float, wb: float, wc: float,
) -> tuple[tuple[int, int, int, int], tuple[float, float, float, float]]:
    total = wa + wb + wc
    if total < 1e-9:
        return (ja, 0, 0, 0), (1.0, 0.0, 0.0, 0.0)

    wa /= total
    wb /= total
    wc /= total
    return (ja, jb, jc, 0), (wa, wb, wc, 0.0)


def blend_segment(
    p0: tuple[float, float, float],
    p1: tuple[float, float, float],
    ja: int,
    jb: int,
    split: float = 0.35,
) -> WeightFn:
    """Blend jb→ja along segment p0→p1; jb dominates near p0, ja dominates after `split`."""
    dx = p1[0] - p0[0]
    dy = p1[1] - p0[1]
    dz = p1[2] - p0[2]
    h2 = dx * dx + dy * dy + dz * dz

    def _fn(px: float, py: float, pz: float):
        if h2 < 1e-12:
            return (ja, 0, 0, 0), (1.0, 0.0, 0.0, 0.0)

        t = ((px - p0[0]) * dx + (py - p0[1]) * dy + (pz - p0[2]) * dz) / h2
        t = max(0.0, min(1.0, t))
        if t >= split:
            return (ja, 0, 0, 0), (1.0, 0.0, 0.0, 0.0)

        w_ja = t / split if split > 0 else 1.0
        w_jb = 1.0 - w_ja
        return _pack(ja, jb, w_ja, w_jb)

    return _fn


def blend_segment_triple(
    p0: tuple[float, float, float],
    p1: tuple[float, float, float],
    ja: int,
    jb: int,
    jc: int,
    zone_a: float = 0.28,
    zone_c: float = 0.72,
) -> WeightFn:
    """Along p0→p1: ja↔jb in the first zone, jb alone mid, jb↔jc in the last zone."""
    dx = p1[0] - p0[0]
    dy = p1[1] - p0[1]
    dz = p1[2] - p0[2]
    h2 = dx * dx + dy * dy + dz * dz

    def _fn(px: float, py: float, pz: float):
        if h2 < 1e-12:
            return (jb, 0, 0, 0), (1.0, 0.0, 0.0, 0.0)

        t = ((px - p0[0]) * dx + (py - p0[1]) * dy + (pz - p0[2]) * dz) / h2
        t = max(0.0, min(1.0, t))
        if t <= zone_a:
            w_jb = 1.0 - t / zone_a if zone_a > 0 else 1.0
            w_ja = t / zone_a if zone_a > 0 else 0.0
            return _pack3(ja, jb, jc, w_ja, w_jb, 0.0)

        if t >= zone_c:
            w_jc = (t - zone_c) / (1.0 - zone_c) if zone_c < 1.0 else 1.0
            w_jb = 1.0 - w_jc
            return _pack3(ja, jb, jc, 0.0, w_jb, w_jc)

        return (ja, jb, jc, 0), (0.0, 1.0, 0.0, 0.0)

    return _fn


def blend_lateral(
    ja: int,
    jb: int,
    axis: str,
    center: float,
    inner: float,
    outer: float,
) -> WeightFn:
    """Blend jb→ja as |axis coordinate| grows past `inner` toward `outer`."""
    def _fn(px: float, py: float, pz: float):
        coord = px if axis == "x" else py if axis == "y" else pz
        dist = abs(coord - center)
        if dist <= inner:
            return (jb, 0, 0, 0), (1.0, 0.0, 0.0, 0.0)
        if dist >= outer:
            return (ja, 0, 0, 0), (1.0, 0.0, 0.0, 0.0)

        t = (dist - inner) / (outer - inner)
        return _pack(ja, jb, t, 1.0 - t)

    return _fn
