"""Client-side ([client-perf]) statistics for the benchmark report.

The Godot observer prints one ``[client-perf]`` sample per second. Two things
make naive aggregation misleading:

* The first sample(s) after joining cover snapshot apply + model instantiation
  (``process_ms`` in the hundreds). Folding them into p95/max hides steady-state
  cost behind a one-off load spike, so they are split out: the **first-spawn
  hitch** (first sample with entities on screen) is reported on its own and the
  next ``settle_samples`` are dropped as warmup.
* A capped observer (vsync on, or macOS Metal/MoltenVK blocking on drawables
  even with ``--disable-vsync``) pins FPS at the refresh rate, so FPS says
  nothing about headroom. Frame/process time percentiles and draw load are the
  primary signal; the report flags a cap detected from the frame-time floor.
"""

from __future__ import annotations

import math
import statistics
from dataclasses import dataclass

DEFAULT_SETTLE_SAMPLES = 2


@dataclass(frozen=True)
class ClientWarmupSplit:
    hitch: dict[str, float] | None
    warmup: list[dict[str, float]]
    steady: list[dict[str, float]]


def split_warmup(samples: list[dict[str, float]], settle_samples: int = DEFAULT_SETTLE_SAMPLES) -> ClientWarmupSplit:
    """Split samples into first-spawn hitch, warmup (hitch + settle), and steady state."""
    if not samples:
        return ClientWarmupSplit(None, [], [])
    hitch_index = next((i for i, s in enumerate(samples) if s.get("entities", 0) > 0), 0)
    cut = hitch_index + 1 + max(0, settle_samples)
    steady = samples[cut:]
    if not steady:
        # Too short to settle: keep everything after the hitch rather than report nothing.
        cut = hitch_index + 1
        steady = samples[cut:]
    return ClientWarmupSplit(samples[hitch_index], samples[:cut], steady)


def values(samples: list[dict[str, float]], key: str) -> list[float]:
    return [float(s[key]) for s in samples if key in s]


def percentile(vals: list[float], p: float) -> float:
    """Nearest-rank percentile (p in 0..1)."""
    if not vals:
        return 0.0
    ordered = sorted(vals)
    rank = max(1, math.ceil(p * len(ordered)))
    return ordered[min(rank, len(ordered)) - 1]


def fmt_percentiles(vals: list[float], unit: str = "ms") -> str:
    if not vals:
        return "n/a"
    return (
        f"p50 {percentile(vals, 0.50):7.2f}  p95 {percentile(vals, 0.95):7.2f}  "
        f"p99 {percentile(vals, 0.99):7.2f}  max {max(vals):7.2f} {unit}"
    )


def fmt_count(vals: list[float]) -> str:
    if not vals:
        return "n/a"
    return f"p50 {percentile(vals, 0.50):8.0f}  p95 {percentile(vals, 0.95):8.0f}  max {max(vals):8.0f}"


def fmt_avg_p95_max(vals: list[float], unit: str = "ms") -> str:
    if not vals:
        return "n/a"
    return f"avg {statistics.mean(vals):6.1f} {unit}  p95 {percentile(vals, 0.95):6.1f} {unit}  max {max(vals):6.1f} {unit}"


def frame_time_floor_capped(samples: list[dict[str, float]]) -> bool:
    """True when the fast quarter of frame windows is no faster than the median.

    Uncapped rendering shows frame times scattered below the median; a display
    or compositor cap (vsync, or macOS Metal blocking on drawables even with
    --disable-vsync) pins the floor at the refresh interval. p25 rather than
    min: 1 s sample windows occasionally land one frame short (15.5 ms at 60 Hz).
    """
    frames = values(samples, "avg_frame_ms")
    if len(frames) < 3:
        return False
    median = percentile(frames, 0.50)
    return median > 0 and (median - percentile(frames, 0.25)) / median < 0.01


def vsync_label(samples: list[dict[str, float]]) -> str:
    modes = {int(s["vsync"]) for s in samples if "vsync" in s}
    if not modes:
        requested = "unknown (sample predates vsync field)"
    elif modes == {0}:
        requested = "off"
    else:
        requested = "on"
    if frame_time_floor_capped(samples):
        return f"{requested}; frame-time floor = median → FPS is capped, read process_ms for headroom"
    return requested


_DELTA_PHASES = [
    ("delta", "delta total "),
    ("d_chg", "d_chg       "),
    ("d_upsert", "d_upsert    "),
    ("d_upsert_player", "d_upsert_pl "),
    ("d_upsert_m", "d_upsert_m  "),
    ("d_evt", "d_evt       "),
    ("d_ui", "d_ui        "),
    ("d_recon", "d_recon     "),
]

_HITCH_KEYS = ["avg_frame_ms", "process_ms", "net_poll", "d_upsert", "d_upsert_m", "delta"]


def render_client_block(label: str, samples: list[dict[str, float]], settle_samples: int = DEFAULT_SETTLE_SAMPLES) -> list[str]:
    if not samples:
        return []
    split = split_warmup(samples, settle_samples)
    steady = split.steady
    lines = [
        "─" * 62,
        f"  CLIENT — {label}",
        f"  Samples: {len(samples)}  (steady {len(steady)}, warmup excluded {len(split.warmup)})",
        f"  vsync  : {vsync_label(steady or samples)}",
        "",
    ]

    if split.hitch is not None:
        parts = [f"{key} {split.hitch[key]:.1f}" for key in _HITCH_KEYS if key in split.hitch]
        lines.append("  FIRST-SPAWN HITCH (first sample with entities; excluded below)")
        lines.append(f"    {'  '.join(parts)}")
        lines.append("")

    lines.append("  FRAME TIME (steady state)")
    lines.append(f"    avg_frame_ms  {fmt_percentiles(values(steady, 'avg_frame_ms'))}")
    lines.append(f"    process_ms    {fmt_percentiles(values(steady, 'process_ms'))}")
    lines.append(f"    physics_ms    {fmt_percentiles(values(steady, 'physics_ms'))}")
    fps_vals = values(steady, "fps")
    if fps_vals:
        lines.append(
            f"    fps           avg {statistics.mean(fps_vals):5.1f}  p5 {percentile(fps_vals, 0.05):5.1f}  min {min(fps_vals):5.1f}"
        )
    lines.append("")

    lines.append("  DRAW LOAD (steady state, per sample)")
    for key, label_ in [
        ("draw_calls", "draw_calls  "),
        ("primitives", "primitives  "),
        ("objects", "objects     "),
        ("nodes", "scene nodes "),
    ]:
        vals = values(steady, key)
        if vals:
            lines.append(f"    {label_}  {fmt_count(vals)}")
    fog_vals = values(steady, "fog")
    if fog_vals:
        lines.append(f"    fog ms        {fmt_avg_p95_max(fog_vals)}")
    lines.append("")

    lines.append("  CLIENT DELTA PHASES (steady state, ms accumulated / sample)")
    for key, label_ in _DELTA_PHASES:
        vals = values(steady, key)
        if vals and max(vals) > 0.01:
            lines.append(f"    {label_}  {fmt_avg_p95_max(vals)}")
    lines.append("")

    monster_vals = values(steady, "live_monsters")
    if monster_vals:
        lines.append("  SCENE ENTITIES (steady state, per sample)")
        lines.append(f"    live_monsters   {fmt_count(monster_vals)}")
        proj_vals = values(steady, "projectiles")
        if proj_vals:
            lines.append(f"    projectiles     {fmt_count(proj_vals)}")
        lines.append("")

    return lines
