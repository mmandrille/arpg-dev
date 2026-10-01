"""Summarize per-frame live Godot first-spawn trials from scripts/benchmark.sh."""

from __future__ import annotations

import argparse
import json
import platform
import subprocess
from pathlib import Path
from typing import Any

from tools.bot.benchmark_client_stats import percentile

PREFIX = "[client-spawn-frame] "
REQUIRED_FIELDS = {
    "frame", "tick", "first_spawn", "monsters", "frame_interval_ms",
    "process_wall_ms", "phases_ms", "draw_calls", "resources", "nodes",
    "static_memory_mib", "engine_process_ms",
}


def parse_trace(path: Path) -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    for line in path.read_text(encoding="utf-8", errors="replace").splitlines():
        if not line.startswith(PREFIX):
            continue
        row = json.loads(line[len(PREFIX):])
        if not isinstance(row, dict) or not REQUIRED_FIELDS <= row.keys():
            raise ValueError(f"{path}: malformed trace row")
        if rows and row["frame"] != rows[-1]["frame"] + 1:
            raise ValueError(f"{path}: skipped or repeated frame after {rows[-1]['frame']}")
        rows.append(row)
    if not rows:
        raise ValueError(f"{path}: no live Godot frame trace")
    return rows


def _stats(values: list[float]) -> dict[str, float]:
    if not values:
        return {}
    return {
        "p50": percentile(values, 0.50),
        "p95": percentile(values, 0.95),
        "max": max(values),
    }


def summarize_trial(rows: list[dict[str, Any]], bot_log: str) -> dict[str, Any]:
    if "scenario done " not in bot_log or "scenario failed " in bot_log:
        raise ValueError("bot scenario did not complete successfully")
    markers = [i for i, row in enumerate(rows) if row["first_spawn"]]
    if len(markers) != 1:
        raise ValueError(f"expected one first-spawn marker, got {len(markers)}")
    index = markers[0]
    if index + 1 >= len(rows):
        raise ValueError("trace ends on first-spawn frame")
    spawn = rows[index]
    if spawn["monsters"] <= 0 or spawn["process_wall_ms"] <= 0:
        raise ValueError("first-spawn frame has no monsters or process timing")
    if not rows[0].get("renderer") or not rows[0].get("graphics_quality"):
        raise ValueError("missing renderer or graphics quality metadata")
    # The interval including spawn work is observed at the next frame start.
    following = rows[index + 1:index + 6]
    steady: list[dict[str, Any]] = []
    elapsed_ms = 0.0
    for row in rows[index + 1:]:
        elapsed_ms += max(0.0, float(row["frame_interval_ms"]))
        if elapsed_ms >= 3000.0:
            steady.append(row)
    if len(following) < 5 or not steady:
        raise ValueError("trace ends before catch-up and steady-state frames")
    phases = {key: float(value) for key, value in spawn["phases_ms"].items()}
    return {
        "frames": len(rows),
        "spawn_frame": spawn["frame"],
        "tick": spawn["tick"],
        "model_count": spawn["monsters"],
        "entity_count": spawn.get("entities"),
        "renderer": rows[0]["renderer"],
        "graphics_quality": rows[0]["graphics_quality"],
        "spawn_process_wall_ms": float(spawn["process_wall_ms"]),
        "spawn_engine_process_ms": float(spawn["engine_process_ms"]),
        "next_engine_process_ms": float(rows[index + 1]["engine_process_ms"]),
        "pre_spawn_interval_ms": float(spawn["frame_interval_ms"]),
        "spawn_frame_interval_ms": float(rows[index + 1]["frame_interval_ms"]),
        "spawn_draw_calls": int(rows[index + 1]["draw_calls"]),
        "following_intervals_ms": [float(row["frame_interval_ms"]) for row in following],
        "spawn_phases_ms": dict(sorted(phases.items(), key=lambda item: item[1], reverse=True)),
        "steady_frame_interval_ms": _stats([float(row["frame_interval_ms"]) for row in steady]),
        "steady_process_wall_ms": _stats([float(row["process_wall_ms"]) for row in steady]),
        "steady_draw_calls": _stats([float(row["draw_calls"]) for row in steady]),
        "resource_count": {"spawn": spawn["resources"], "max": max(row["resources"] for row in rows)},
        "node_count": {"spawn": spawn["nodes"], "max": max(row["nodes"] for row in rows)},
        "static_memory_mib": {"spawn": spawn["static_memory_mib"], "max": max(row["static_memory_mib"] for row in rows)},
    }


def _command_output(args: list[str]) -> str:
    try:
        return subprocess.check_output(args, text=True, stderr=subprocess.DEVNULL).strip()
    except (OSError, subprocess.CalledProcessError):
        return "unavailable"


def build_report(run_dir: Path, scenario: str, runs: int, godot: str, baseline_label: str) -> dict[str, Any]:
    scenario_path = next((path for path in Path("tools/bot/scenarios").glob("*.json")
                          if json.loads(path.read_text(encoding="utf-8")).get("id") == scenario), None)
    if scenario_path is None:
        raise ValueError(f"scenario {scenario!r} was not found")
    fixture = json.loads(scenario_path.read_text(encoding="utf-8"))
    trials: list[dict[str, Any]] = []
    errors: list[str] = []
    for number in range(1, runs + 1):
        stem = f"{scenario}-run{number:02d}"
        client_path = run_dir / f"{stem}-client.log"
        bot_path = run_dir / f"{stem}-bot.log"
        try:
            trial = summarize_trial(parse_trace(client_path), bot_path.read_text(encoding="utf-8", errors="replace"))
            trial["id"] = stem
            trial["client_log"] = str(client_path)
            trial["bot_log"] = str(bot_path)
            trials.append(trial)
        except (OSError, ValueError, json.JSONDecodeError) as exc:
            errors.append(f"{stem}: {exc}")
    result: dict[str, Any] = {
        "baseline_label": baseline_label,
        "scenario": scenario,
        "seed": fixture.get("seed"),
        "world_id": fixture.get("world_id"),
        "requested_runs": runs,
        "valid_runs": len(trials),
        "cold_definition": "fresh Godot observer process per run; OS and GPU caches may persist",
        "host": platform.platform(),
        "machine": platform.machine(),
        "cpu": _command_output(["sysctl", "-n", "machdep.cpu.brand_string"]) if platform.system() == "Darwin" else platform.processor(),
        "godot_version": _command_output([godot, "--version"]),
        "git_head": _command_output(["git", "rev-parse", "HEAD"]),
        "trials": trials,
        "errors": errors,
    }
    if trials:
        for key in ("spawn_process_wall_ms", "pre_spawn_interval_ms", "spawn_frame_interval_ms", "spawn_draw_calls"):
            result[key] = _stats([trial[key] for trial in trials])
        result["model_counts"] = sorted({trial["model_count"] for trial in trials})
        result["renderers"] = sorted({trial["renderer"] for trial in trials})
        result["graphics_qualities"] = sorted({trial["graphics_quality"] for trial in trials})
        result["following_intervals_ms"] = [trial["following_intervals_ms"] for trial in trials]
    if len(trials) != runs:
        errors.append(f"only {len(trials)}/{runs} valid live-client trials")
    if len(result.get("model_counts", [])) > 1:
        errors.append("monster model count varied between runs")
    if len(result.get("renderers", [])) > 1 or len(result.get("graphics_qualities", [])) > 1:
        errors.append("renderer or quality tier varied between runs")
    return result


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--run-dir", type=Path, required=True)
    parser.add_argument("--scenario", required=True)
    parser.add_argument("--runs", type=int, required=True)
    parser.add_argument("--godot", default="godot")
    parser.add_argument("--baseline-label", default="current-checkout")
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    report = build_report(args.run_dir, args.scenario, args.runs, args.godot, args.baseline_label)
    report_path = args.run_dir / "first-spawn-summary.json"
    report_path.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    lines = [
        f"First-spawn live Godot: {report['valid_runs']}/{args.runs} valid trials",
        f"Scenario={args.scenario} seed={report['seed']} baseline={args.baseline_label}",
        f"Host={report['host']} CPU={report['cpu']} Godot={report['godot_version']}",
        f"Renderer={report.get('renderers', [])} quality={report.get('graphics_qualities', [])} models={report.get('model_counts', [])}",
    ]
    for key in ("spawn_process_wall_ms", "pre_spawn_interval_ms", "spawn_frame_interval_ms", "spawn_draw_calls"):
        if key in report:
            lines.append(f"{key}: {report[key]}")
    for trial in report["trials"]:
        top = list(trial["spawn_phases_ms"].items())[:5]
        lines.append(f"{trial['id']}: spawn process={trial['spawn_process_wall_ms']:.2f} ms "
                     f"frame={trial['spawn_frame_interval_ms']:.2f} ms models={trial['model_count']} top phases={top}")
    lines.extend(f"ERROR: {error}" for error in report["errors"])
    text = "\n".join(lines) + "\n"
    args.out.write_text(text, encoding="utf-8")
    print(text, end="")
    print(f"Raw summary: {report_path}")
    return 1 if report["errors"] else 0


if __name__ == "__main__":
    raise SystemExit(main())
