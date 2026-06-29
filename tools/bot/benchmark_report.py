"""Generate a perf report from a benchmark bot run.

Usage (called by scripts/benchmark.sh):
    python -m tools.bot.benchmark_report --server-log <path> --bot-log <path> [--out <path>]

Reads:
  --server-log  Raw server output (JSON structured logs with backend_perf lines)
  --bot-log     Bot stderr (scenario begin/done boundary markers)
  --out         Optional output file path (default: print to stdout)

Produces a per-scenario performance summary with tick budget, sim phases,
pathfinding, and entity load statistics. Any future scenario file with
ci_tier=benchmark is included automatically.
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

# ── parser helpers ─────────────────────────────────────────────────────────────

_BOT_SCENARIO_BEGIN = re.compile(r"\[bot (\d{2}:\d{2}:\d{2})\] scenario begin (\S+)")
_BOT_SCENARIO_DONE = re.compile(r"\[bot (\d{2}:\d{2}:\d{2})\] scenario (?:done|failed) (\S+)")


def _hms_to_seconds(hms: str, date: datetime) -> float:
    """Convert HH:MM:SS (UTC wall clock, same date as `date`) to a unix timestamp."""
    h, m, s = (int(x) for x in hms.split(":"))
    return date.replace(hour=h, minute=m, second=s, microsecond=0, tzinfo=timezone.utc).timestamp()


def parse_bot_boundaries(bot_log: Path) -> list[dict[str, Any]]:
    """Return list of {id, begin_ts, end_ts} from bot scenario markers."""
    text = bot_log.read_text(encoding="utf-8", errors="replace")
    # Use the first timestamp we see to anchor the date
    first_ts = re.search(r"\[bot (\d{2}:\d{2}:\d{2})\]", text)
    if not first_ts:
        return []
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


def parse_perf_samples(server_log: Path) -> list[dict[str, Any]]:
    """Return all backend_perf JSON objects from the server log, with a `ts` field."""
    samples: list[dict[str, Any]] = []
    for line in server_log.read_text(encoding="utf-8", errors="replace").splitlines():
        line = line.strip()
        if not line or line[0] != "{":
            continue
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue
        if obj.get("message") != "backend_perf":
            continue
        # Parse timestamp from the structured log
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
    """Map samples to scenarios by timestamp window. Unassigned → '_unassigned'."""
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


# ── statistics helpers ─────────────────────────────────────────────────────────

def _vals(samples: list[dict[str, Any]], key: str) -> list[float]:
    return [float(s[key]) for s in samples if key in s]


def _fmt(values: list[float], unit: str = "ms") -> str:
    if not values:
        return "n/a"
    avg = statistics.mean(values)
    p95 = sorted(values)[int(len(values) * 0.95)]
    mx = max(values)
    return f"avg {avg:6.1f} {unit}  p95 {p95:6.1f} {unit}  max {mx:6.1f} {unit}"


def _fmt_int(values: list[float]) -> str:
    if not values:
        return "n/a"
    avg = statistics.mean(values)
    mx = max(values)
    return f"avg {avg:5.0f}   max {mx:5.0f}"


# ── report rendering ───────────────────────────────────────────────────────────

def render_scenario_block(scenario_id: str, samples: list[dict[str, Any]]) -> list[str]:
    lines: list[str] = []
    sep = "─" * 62
    lines.append(sep)
    lines.append(f"  {scenario_id}")
    lines.append(f"  Samples: {len(samples)}")
    lines.append("")

    # Tick budget
    overruns = [s for s in samples if s.get("tick_over_budget")]
    overrun_pct = 100.0 * len(overruns) / len(samples) if samples else 0.0
    overrun_ms_vals = [float(s.get("tick_overrun_ms", 0)) for s in overruns]
    max_overrun = max(overrun_ms_vals) if overrun_ms_vals else 0.0
    lines.append("  TICK BUDGET")
    lines.append(f"    overruns        {len(overruns):4d}  ({overrun_pct:.1f}% of samples)")
    lines.append(f"    max overrun     {max_overrun:.1f} ms")
    lines.append("")

    # Sim phases
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

    # Pathfinding
    lines.append("  PATHFINDING (per sample)")
    req_vals = _vals(samples, "path_requests")
    hit_vals = _vals(samples, "path_cache_hits")
    node_vals = _vals(samples, "path_nodes_visited")
    if req_vals:
        lines.append(f"    requests        {_fmt_int(req_vals)}")
    if req_vals and hit_vals:
        ratios = [
            min(h / r * 100.0, 100.0)
            for h, r in zip(hit_vals, req_vals)
            if r > 0
        ]
        if ratios:
            lines.append(f"    cache hit %     avg {statistics.mean(ratios):.0f}%")
    if node_vals:
        lines.append(f"    nodes visited   {_fmt_int(node_vals)}")
    lines.append("")

    # Entity load
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


def render_report(
    scenario_samples: dict[str, list[dict[str, Any]]],
    boundaries: list[dict[str, Any]],
    generated_at: str,
) -> str:
    ordered_ids = [b["id"] for b in boundaries if b["id"] in scenario_samples]
    # Any that weren't in boundaries (e.g. boundary detection failed)
    for sid in scenario_samples:
        if sid not in ordered_ids and sid != "_unassigned":
            ordered_ids.append(sid)

    total_samples = sum(len(v) for k, v in scenario_samples.items() if k != "_unassigned")

    header = [
        "╔══════════════════════════════════════════════════════════════╗",
        "║           ARPG PERFORMANCE BENCHMARK REPORT                  ║",
        "╚══════════════════════════════════════════════════════════════╝",
        f"  Generated : {generated_at}",
        f"  Scenarios : {len(ordered_ids)}",
        f"  Total samples: {total_samples}",
        "",
    ]

    body: list[str] = []
    for sid in ordered_ids:
        body.extend(render_scenario_block(sid, scenario_samples[sid]))

    unassigned = scenario_samples.get("_unassigned", [])
    if unassigned:
        body.append("─" * 62)
        body.append(f"  (unassigned samples: {len(unassigned)} — scenario boundary detection may have failed)")
        body.append("")

    return "\n".join(header + body)


# ── entry point ────────────────────────────────────────────────────────────────

def main() -> None:
    parser = argparse.ArgumentParser(description="Generate ARPG benchmark perf report")
    parser.add_argument("--server-log", type=Path, required=True, help="Server log file")
    parser.add_argument("--bot-log", type=Path, required=True, help="Bot stderr log file")
    parser.add_argument("--out", type=Path, help="Write report to file (also prints to stdout)")
    args = parser.parse_args()

    if not args.server_log.exists():
        sys.exit(f"server log not found: {args.server_log}")
    if not args.bot_log.exists():
        sys.exit(f"bot log not found: {args.bot_log}")

    generated_at = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")
    boundaries = parse_bot_boundaries(args.bot_log)
    samples = parse_perf_samples(args.server_log)
    scenario_samples = assign_samples(samples, boundaries)

    if not samples:
        print("WARNING: no backend_perf lines found in server log — was ARPG_PERF_DEBUG=1 set?")

    report = render_report(scenario_samples, boundaries, generated_at)
    print(report)

    if args.out:
        args.out.parent.mkdir(parents=True, exist_ok=True)
        args.out.write_text(report, encoding="utf-8")
        print(f"\nReport written to: {args.out}")


if __name__ == "__main__":
    main()
