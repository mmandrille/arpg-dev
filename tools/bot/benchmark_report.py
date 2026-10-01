"""Generate a perf report from a benchmark bot run or a play-debug session.

Benchmark usage (called by scripts/benchmark.sh):
    python -m tools.bot.benchmark_report \\
        --server-log <path> --bot-log <path> \\
        [--scenario-client-log <scenario_id>=<path> ...] [--out <path>]

Play-debug analysis usage (called by make perf-analyze):
    python -m tools.bot.benchmark_report --play-log /tmp/arpg-perf.log [--out <path>]

  --play-log    Combined tee log from `make play-debug` (has [backend] and
                [client1] prefixes on each line; contains both backend_perf
                JSON and [client-perf] kv lines)

  --server-log  Raw server output (JSON structured logs, no prefix)
  --bot-log     Bot stderr (scenario begin/done boundary markers)
  --client-log  Godot stdout captured during a benchmark (single, unlabelled)
  --scenario-client-log  <scenario_id>=<path> Godot observer log for one
                scenario; repeatable — client stats are reported per scenario
  --out         Optional output file path (also prints to stdout)

Client sections (tools/bot/benchmark_client_stats.py) report the first-spawn
hitch separately, drop warmup samples, and show true frame intervals when
bounded frame batches are present. Older logs retain one-second average frame
percentiles plus process_ms and draw_calls/primitives. FPS alone is misleading
when the observer is vsync-capped.
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

from tools.bot.benchmark_client_stats import DEFAULT_SETTLE_SAMPLES, percentile, render_client_block

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
        if obj.get("message", obj.get("msg")) != "backend_perf":
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
_CLIENT_FRAME_BATCH_LINE = re.compile(r"\[client-frame-batch\]\s+(.+)")
_CLIENT_PREFIX = re.compile(r"^\[client\d*\]\s*")
_KV = re.compile(r"(\w+)=([-\d.]+)")
_MAX_FRAME_BATCH = 2048


def parse_client_perf_lines(log_path: Path) -> list[dict[str, Any]]:
    """Parse [client-perf] key=value lines from any log file (raw Godot output,
    tee'd play-debug log with [client1] prefix, or benchmark client log)."""
    samples: list[dict[str, Any]] = []
    for line in log_path.read_text(encoding="utf-8", errors="replace").splitlines():
        # Strip [client1] / [client] prefix if present (play-debug format)
        line = _CLIENT_PREFIX.sub("", line.strip())
        m = _CLIENT_PERF_LINE.search(line)
        if not m:
            batch = _CLIENT_FRAME_BATCH_LINE.search(line)
            if batch and samples:
                _attach_frame_batch(samples[-1], batch.group(1))
            continue
        kv_str = m.group(1)
        row: dict[str, Any] = {}
        for k, v in _KV.findall(kv_str):
            # Phase timers can reuse a base field name (e.g. `entities` count vs
            # `entities` phase ms); keep the base value and prefix the phase.
            if k in row:
                k = f"phase_{k}"
            try:
                row[k] = float(v)
            except ValueError:
                pass
        if row:
            samples.append(row)
    return samples


def _attach_frame_batch(row: dict[str, Any], fields_text: str) -> None:
    fields = dict(part.split("=", 1) for part in fields_text.split() if "=" in part)
    try:
        count = int(fields.get("n", "-1"))
        dropped = int(fields.get("dropped", "-1"))
        tick = int(fields.get("tick", "-1"))
        raw = fields.get("us", "")
        values = [int(item) for item in raw.split(",")] if raw else []
        if ("frame_us" in row or count <= 0 or count > _MAX_FRAME_BATCH or dropped < 0
                or tick != int(row.get("tick", -2)) or len(values) != count
                or any(value <= 0 for value in values)):
            raise ValueError("invalid frame batch shape")
    except ValueError:
        row["frame_batch_error"] = "invalid frame batch"
        return
    row["frame_us"] = values
    row["frame_dropped"] = dropped
    for key in ("renderer", "quality", "camera", "world", "seed"):
        row[key] = fields.get(key, "")
    for key in ("width", "height"):
        try:
            row[key] = int(fields.get(key, "0"))
        except ValueError:
            row["frame_batch_error"] = "invalid viewport"


def parse_client_sections(client_log: Path | None, scenario_logs: list[str]) -> list[tuple[str, list[dict[str, Any]]]]:
    sections: list[tuple[str, list[dict[str, Any]]]] = []
    if client_log and client_log.exists():
        sections.append(("live benchmark session", parse_client_perf_lines(client_log)))
    for entry in scenario_logs:
        scenario_id, sep, path = entry.partition("=")
        if not sep or not scenario_id or not path:
            sys.exit(f"--scenario-client-log expects SCENARIO_ID=PATH, got {entry!r}")
        log_path = Path(path)
        if log_path.exists():
            sections.append((scenario_id, parse_client_perf_lines(log_path)))
    return sections


# ── statistics helpers ─────────────────────────────────────────────────────────

def _vals(samples: list[dict[str, Any]], key: str) -> list[float]:
    return [float(s[key]) for s in samples if key in s]


def _pct(values: list[float], p: float) -> float:
    return percentile(values, p)


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


# ── report entry point ─────────────────────────────────────────────────────────

def render_report(
    scenario_samples: dict[str, list[dict[str, Any]]],
    boundaries: list[dict[str, Any]],
    client_sections: list[tuple[str, list[dict[str, Any]]]],
    generated_at: str,
    mode: str = "benchmark",
    settle_samples: int = DEFAULT_SETTLE_SAMPLES,
) -> str:
    ordered_ids = [b["id"] for b in boundaries if b["id"] in scenario_samples]
    for sid in scenario_samples:
        if sid not in ordered_ids and sid != "_unassigned":
            ordered_ids.append(sid)

    total_server = sum(len(v) for k, v in scenario_samples.items() if k != "_unassigned")
    total_client = sum(len(samples) for _, samples in client_sections)

    header = [
        "╔══════════════════════════════════════════════════════════════╗",
        "║           ARPG PERFORMANCE REPORT                            ║",
        "╚══════════════════════════════════════════════════════════════╝",
        f"  Generated : {generated_at}",
        f"  Mode      : {mode}",
        f"  Server samples: {total_server}  |  Client samples: {total_client}",
        "",
    ]

    if not total_client and not total_server:
        header.append("  WARNING: no perf data found. Check ARPG_PERF_DEBUG=1 was set")
        header.append("  and that the log files are from the correct session.")
        header.append("")

    body: list[str] = []

    # Client blocks: one per benchmark scenario observer, or one for play-debug.
    for label, samples in client_sections:
        body.extend(render_client_block(label, samples, settle_samples))

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
    parser.add_argument("--client-log", type=Path, help="Godot stdout captured during benchmark (single, unlabelled)")
    parser.add_argument("--scenario-client-log", action="append", default=[], metavar="SCENARIO_ID=PATH",
                        help="Per-scenario Godot observer log; repeatable")
    parser.add_argument("--warmup-settle-samples", type=int, default=DEFAULT_SETTLE_SAMPLES,
                        help="Client samples dropped after the first-spawn hitch before steady-state stats")
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
        client_sections = [("live session", parse_client_perf_lines(args.play_log))]
        scenario_samples = assign_samples(server_samples, [])
        boundaries: list[dict[str, Any]] = []
        mode = "play-debug"
    else:
        if not args.server_log or not args.server_log.exists():
            sys.exit(f"server log not found: {args.server_log}")
        server_samples = parse_perf_samples(args.server_log)
        boundaries = parse_bot_boundaries(args.bot_log) if args.bot_log and args.bot_log.exists() else []
        scenario_samples = assign_samples(server_samples, boundaries)
        client_sections = parse_client_sections(args.client_log, args.scenario_client_log)
        mode = "benchmark"

    if not server_samples:
        print("WARNING: no backend_perf lines found in server log — was ARPG_PERF_DEBUG=1 set?")
    if not any(samples for _, samples in client_sections):
        print("WARNING: no [client-perf] lines found in client log")

    report = render_report(scenario_samples, boundaries, client_sections, generated_at, mode, args.warmup_settle_samples)
    print(report)

    if args.out:
        args.out.parent.mkdir(parents=True, exist_ok=True)
        args.out.write_text(report, encoding="utf-8")
        print(f"\nReport written to: {args.out}")


if __name__ == "__main__":
    main()
