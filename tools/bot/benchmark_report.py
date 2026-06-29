"""Generate a perf report from a benchmark bot run or a play-debug session.

Benchmark usage (called by scripts/benchmark.sh):
    python -m tools.bot.benchmark_report \\
        --server-log <path> --bot-log <path> [--client-log <path>] [--out <path>]

Play-debug analysis usage (called by make perf-analyze):
    python -m tools.bot.benchmark_report --play-log /tmp/arpg-perf.log [--out <path>]

  --play-log    Combined tee log from `make play-debug` (has [backend] and
                [client1] prefixes on each line; contains both backend_perf
                JSON and [client-perf] kv lines)

  --server-log  Raw server output (JSON structured logs, no prefix)
  --bot-log     Bot stderr (scenario begin/done boundary markers)
  --client-log  Godot stdout captured during benchmark Godot replay
  --out         Optional output file path (also prints to stdout)

The client section shows FPS distribution, frame budget, per-phase delta cost,
fog cost, and draw calls — the metrics that actually explain low FPS in real play.
"""

from __future__ import annotations

import argparse
import json
import re
import statistics
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

# ── bot boundary parser ────────────────────────────────────────────────────────

_BOT_SCENARIO_BEGIN = re.compile(r"\[bot (\d{2}:\d{2}:\d{2})\] scenario begin (\S+)")
_BOT_SCENARIO_DONE = re.compile(r"\[bot (\d{2}:\d{2}:\d{2})\] scenario (?:done|failed) (\S+)")


def _hms_to_seconds(hms: str, date: datetime) -> float:
    h, m, s = (int(x) for x in hms.split(":"))
    return date.replace(hour=h, minute=m, second=s, microsecond=0, tzinfo=timezone.utc).timestamp()


def parse_bot_boundaries(bot_log: Path) -> list[dict[str, Any]]:
    text = bot_log.read_text(encoding="utf-8", errors="replace")
    anchor_date = datetime.now(timezone.utc)
    pending: dict[str, float] = {}
    boundaries: list[dict[str, Any]] = []
    for line in text.splitlines():
        m = _BOT_SCENARIO_BEGIN.search(line)
        if m:
            pending[m.group(2)] = _hms_to_seconds(m.group(1), anchor_date)
            continue
        m = _BOT_SCENARIO_DONE.search(line)
        if m:
            sid = m.group(2)
            begin = pending.pop(sid, None)
            end = _hms_to_seconds(m.group(1), anchor_date)
            boundaries.append({"id": sid, "begin_ts": begin, "end_ts": end})
    return boundaries


# ── server log parser ──────────────────────────────────────────────────────────

_BACKEND_PREFIX = re.compile(r"^\[backend\]\s*")


def parse_perf_samples(server_log: Path) -> list[dict[str, Any]]:
    """Parse backend_perf JSON lines. Handles raw JSON or [backend]-prefixed lines."""
    samples: list[dict[str, Any]] = []
    for line in server_log.read_text(encoding="utf-8", errors="replace").splitlines():
        line = line.strip()
        line = _BACKEND_PREFIX.sub("", line)
        if not line or line[0] != "{":
            continue
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue
        if obj.get("message") != "backend_perf":
            continue
        ts_str = obj.get("ts", "")
        try:
            ts = datetime.fromisoformat(ts_str.replace("Z", "+00:00")).timestamp()
        except (ValueError, AttributeError):
            ts = 0.0
        obj["_ts"] = ts
        samples.append(obj)
    return samples


def assign_samples(
    samples: list[dict[str, Any]],
    boundaries: list[dict[str, Any]],
) -> dict[str, list[dict[str, Any]]]:
    result: dict[str, list[dict[str, Any]]] = {}
    for sample in samples:
        ts = sample["_ts"]
        assigned = False
        for b in boundaries:
            if b["begin_ts"] is not None and b["begin_ts"] <= ts <= b["end_ts"] + 2:
                result.setdefault(b["id"], []).append(sample)
                assigned = True
                break
        if not assigned:
            result.setdefault("_unassigned", []).append(sample)
    return result


# ── client perf log parser ─────────────────────────────────────────────────────

_CLIENT_PERF_LINE = re.compile(r"\[client-perf\]\s+(.+)")
_CLIENT_PREFIX = re.compile(r"^\[client\d*\]\s*")
_KV = re.compile(r"(\w+)=([-\d.]+)")


def parse_client_perf_lines(log_path: Path) -> list[dict[str, float]]:
    """Parse [client-perf] key=value lines from any log file (raw Godot output,
    tee'd play-debug log with [client1] prefix, or benchmark client log)."""
    samples: list[dict[str, float]] = []
    for line in log_path.read_text(encoding="utf-8", errors="replace").splitlines():
        # Strip [client1] / [client] prefix if present (play-debug format)
        line = _CLIENT_PREFIX.sub("", line.strip())
        m = _CLIENT_PERF_LINE.search(line)
        if not m:
            continue
        kv_str = m.group(1)
        row: dict[str, float] = {}
        for k, v in _KV.findall(kv_str):
            try:
                row[k] = float(v)
            except ValueError:
                pass
        if row:
            samples.append(row)
    return samples


# ── statistics helpers ─────────────────────────────────────────────────────────

def _vals(samples: list[dict[str, Any]], key: str) -> list[float]:
    return [float(s[key]) for s in samples if key in s]


def _pct(values: list[float], p: float) -> float:
    return sorted(values)[max(0, int(len(values) * p) - 1)] if values else 0.0


def _fmt(values: list[float], unit: str = "ms") -> str:
    if not values:
        return "n/a"
    avg = statistics.mean(values)
    p95 = _pct(values, 0.95)
    mx = max(values)
    return f"avg {avg:6.1f} {unit}  p95 {p95:6.1f} {unit}  max {mx:6.1f} {unit}"


def _fmt_int(values: list[float]) -> str:
    if not values:
        return "n/a"
    return f"avg {statistics.mean(values):5.0f}   max {max(values):5.0f}"


# ── server block renderer ──────────────────────────────────────────────────────

def render_server_block(scenario_id: str, samples: list[dict[str, Any]]) -> list[str]:
    lines: list[str] = []
    sep = "─" * 62
    lines.append(sep)
    lines.append(f"  SERVER — {scenario_id}")
    lines.append(f"  Samples: {len(samples)}")
    lines.append("")

    overruns = [s for s in samples if s.get("tick_over_budget")]
    overrun_pct = 100.0 * len(overruns) / len(samples) if samples else 0.0
    max_overrun = max((float(s.get("tick_overrun_ms", 0)) for s in overruns), default=0.0)
    lines.append("  TICK BUDGET")
    lines.append(f"    overruns        {len(overruns):4d}  ({overrun_pct:.1f}% of samples)")
    lines.append(f"    max overrun     {max_overrun:.1f} ms")
    lines.append("")

    lines.append("  SIMULATION (ms / sample)")
    for key, label in [
        ("total_ms",     "total_ms    "),
        ("sim_ms",       "sim_ms      "),
        ("ai_ms",        "ai_ms       "),
        ("pathfind_ms",  "pathfind_ms "),
        ("combat_ms",    "combat_ms   "),
        ("broadcast_ms", "broadcast_ms"),
        ("persist_ms",   "persist_ms  "),
    ]:
        vals = _vals(samples, key)
        if vals:
            lines.append(f"    {label}  {_fmt(vals)}")
    lines.append("")

    lines.append("  PATHFINDING (per sample)")
    req_vals = _vals(samples, "path_requests")
    hit_vals = _vals(samples, "path_cache_hits")
    node_vals = _vals(samples, "path_nodes_visited")
    if req_vals:
        lines.append(f"    requests        {_fmt_int(req_vals)}")
    if req_vals and hit_vals:
        ratios = [min(h / r * 100.0, 100.0) for h, r in zip(hit_vals, req_vals) if r > 0]
        if ratios:
            lines.append(f"    cache hit %     avg {statistics.mean(ratios):.0f}%")
    if node_vals:
        lines.append(f"    nodes visited   {_fmt_int(node_vals)}")
    lines.append("")

    lines.append("  ENTITY LOAD (per sample)")
    for key, label in [
        ("monsters_moved", "monsters_moved"),
        ("changes",        "changes       "),
        ("events",         "events        "),
        ("inputs",         "inputs        "),
    ]:
        vals = _vals(samples, key)
        if vals:
            lines.append(f"    {label}  {_fmt_int(vals)}")
    client_vals = _vals(samples, "clients")
    if client_vals:
        lines.append(f"    clients          {int(statistics.mean(client_vals))}")
    lines.append("")

    return lines


# ── client block renderer ──────────────────────────────────────────────────────

def render_client_block(label: str, samples: list[dict[str, float]]) -> list[str]:
    if not samples:
        return []
    lines: list[str] = []
    sep = "─" * 62
    lines.append(sep)
    lines.append(f"  CLIENT — {label}")
    lines.append(f"  Samples: {len(samples)}")
    lines.append("")

    fps_vals = _vals(samples, "fps")
    frame_vals = _vals(samples, "avg_frame_ms")
    if fps_vals:
        avg_fps = statistics.mean(fps_vals)
        p5_fps = _pct(fps_vals, 0.05)   # worst 5% — the "low" tail
        min_fps = min(fps_vals)
        lines.append("  FPS")
        lines.append(f"    avg {avg_fps:5.1f}  p5 (worst tail) {p5_fps:5.1f}  min {min_fps:5.1f}")
        lines.append("")

    if frame_vals:
        lines.append("  FRAME TIME (ms)")
        lines.append(f"    {_fmt(frame_vals)}")
        lines.append("")

    lines.append("  CLIENT DELTA PHASES (ms accumulated / sample)")
    for key, label_ in [
        ("delta",          "delta total "),
        ("d_chg",          "d_chg       "),
        ("d_upsert",       "d_upsert    "),
        ("d_upsert_player","d_upsert_pl "),
        ("d_upsert_m",     "d_upsert_m  "),
        ("d_evt",          "d_evt       "),
        ("d_ui",           "d_ui        "),
        ("d_recon",        "d_recon     "),
    ]:
        vals = _vals(samples, key)
        if vals and max(vals) > 0.01:
            lines.append(f"    {label_}  {_fmt(vals)}")
    lines.append("")

    lines.append("  RENDERING (per sample)")
    for key, label_ in [
        ("fog",        "fog ms      "),
        ("draw_calls", "draw_calls  "),
        ("primitives", "primitives  "),
        ("nodes",      "scene nodes "),
        ("objects",    "objects     "),
    ]:
        vals = _vals(samples, key)
        if vals:
            if key == "fog":
                lines.append(f"    {label_}  {_fmt(vals)}")
            else:
                lines.append(f"    {label_}  {_fmt_int(vals)}")
    lines.append("")

    entity_vals = _vals(samples, "live_monsters")
    proj_vals = _vals(samples, "projectiles")
    if entity_vals:
        lines.append("  SCENE ENTITIES (per sample)")
        lines.append(f"    live_monsters   {_fmt_int(entity_vals)}")
        if proj_vals:
            lines.append(f"    projectiles     {_fmt_int(proj_vals)}")
        lines.append("")

    return lines


# ── report entry point ─────────────────────────────────────────────────────────

def render_report(
    scenario_samples: dict[str, list[dict[str, Any]]],
    boundaries: list[dict[str, Any]],
    client_samples: list[dict[str, float]],
    generated_at: str,
    mode: str = "benchmark",
) -> str:
    ordered_ids = [b["id"] for b in boundaries if b["id"] in scenario_samples]
    for sid in scenario_samples:
        if sid not in ordered_ids and sid != "_unassigned":
            ordered_ids.append(sid)

    total_server = sum(len(v) for k, v in scenario_samples.items() if k != "_unassigned")

    header = [
        "╔══════════════════════════════════════════════════════════════╗",
        "║           ARPG PERFORMANCE REPORT                            ║",
        "╚══════════════════════════════════════════════════════════════╝",
        f"  Generated : {generated_at}",
        f"  Mode      : {mode}",
        f"  Server samples: {total_server}  |  Client samples: {len(client_samples)}",
        "",
    ]

    if not client_samples and not total_server:
        header.append("  WARNING: no perf data found. Check ARPG_PERF_DEBUG=1 was set")
        header.append("  and that the log files are from the correct session.")
        header.append("")

    body: list[str] = []

    # Client block (global — not sliced per scenario in play-debug mode)
    if client_samples:
        body.extend(render_client_block("live session" if mode == "play-debug" else "live benchmark session", client_samples))

    # Server blocks per scenario
    for sid in ordered_ids:
        body.extend(render_server_block(sid, scenario_samples[sid]))

    # Global server block for play-debug (no scenario boundaries)
    if not ordered_ids and scenario_samples.get("_unassigned"):
        body.extend(render_server_block("play session", scenario_samples["_unassigned"]))

    unassigned = scenario_samples.get("_unassigned", [])
    if unassigned and ordered_ids:
        body.append("─" * 62)
        body.append(f"  (unassigned server samples: {len(unassigned)})")
        body.append("")

    return "\n".join(header + body)


def main() -> None:
    parser = argparse.ArgumentParser(description="ARPG perf report — benchmark or play-debug")
    parser.add_argument("--server-log", type=Path)
    parser.add_argument("--bot-log", type=Path)
    parser.add_argument("--client-log", type=Path, help="Godot stdout captured during benchmark")
    parser.add_argument("--play-log", type=Path,
                        help="Combined tee log from make play-debug (has [backend]/[client1] prefixes)")
    parser.add_argument("--out", type=Path)
    args = parser.parse_args()

    if not args.play_log and not args.server_log:
        sys.exit("provide --play-log (play-debug analysis) or --server-log + --bot-log (benchmark)")

    generated_at = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")

    if args.play_log:
        if not args.play_log.exists():
            sys.exit(f"play log not found: {args.play_log}")
        # play-debug log has both server and client lines with prefixes
        server_samples = parse_perf_samples(args.play_log)
        client_samples = parse_client_perf_lines(args.play_log)
        scenario_samples = assign_samples(server_samples, [])
        boundaries: list[dict[str, Any]] = []
        mode = "play-debug"
    else:
        if not args.server_log or not args.server_log.exists():
            sys.exit(f"server log not found: {args.server_log}")
        server_samples = parse_perf_samples(args.server_log)
        boundaries = parse_bot_boundaries(args.bot_log) if args.bot_log and args.bot_log.exists() else []
        scenario_samples = assign_samples(server_samples, boundaries)
        client_samples = parse_client_perf_lines(args.client_log) if args.client_log and args.client_log.exists() else []
        mode = "benchmark"

    if not server_samples:
        print("WARNING: no backend_perf lines found in server log — was ARPG_PERF_DEBUG=1 set?")
    if not client_samples:
        print("WARNING: no [client-perf] lines found in client log")

    report = render_report(scenario_samples, boundaries, client_samples, generated_at, mode)
    print(report)

    if args.out:
        args.out.parent.mkdir(parents=True, exist_ok=True)
        args.out.write_text(report, encoding="utf-8")
        print(f"\nReport written to: {args.out}")


if __name__ == "__main__":
    main()
