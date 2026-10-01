"""Validate and summarize a pinned live dungeon client/server frame trace.

Run only with logs retained from one ``dungeon_frame_pacing_probe`` scenario.
The report deliberately fails when the scene or timing samples are incomplete.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import platform
import re
import subprocess
from pathlib import Path
from typing import Any

from tools.bot.benchmark_client_stats import frame_intervals_ms, percentile, split_warmup, values
from tools.bot.benchmark_report import parse_client_perf_lines, parse_perf_samples

ROOT = Path(__file__).resolve().parents[2]
DEFAULT_SCENARIO = ROOT / "tools/bot/scenarios/client/108_dungeon_frame_pacing_probe.json"


def _require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def _expected_size(label: str) -> tuple[int, int]:
    match = re.fullmatch(r"(\d+)x(\d+)", label)
    _require(match is not None, f"invalid fixture window_size: {label!r}")
    return int(match.group(1)), int(match.group(2))


def _git_revision() -> str:
    result = subprocess.run(["git", "rev-parse", "--short", "HEAD"], cwd=ROOT,
                            check=False, capture_output=True, text=True)
    return result.stdout.strip() if result.returncode == 0 else "unknown"


def _catalog_sha256() -> str:
    digest = hashlib.sha256()
    for group in (ROOT / "shared/assets", ROOT / "shared/rules"):
        for catalog in sorted(group.rglob("*.json")):
            digest.update(str(catalog.relative_to(ROOT)).encode("utf-8"))
            digest.update(catalog.read_bytes())
    return digest.hexdigest()[:16]


def _pct_line(label: str, data: list[float], unit: str) -> str:
    _require(bool(data), f"no {label} samples")
    return (f"{label}: n={len(data)} p50={percentile(data, 0.50):.2f}{unit} "
            f"p95={percentile(data, 0.95):.2f}{unit} "
            f"p99={percentile(data, 0.99):.2f}{unit}")


def build_report(scenario_path: Path, client_log: Path, server_log: Path,
                 quality: str | None = None, window_size: str | None = None) -> str:
    scenario = json.loads(scenario_path.read_text(encoding="utf-8"))
    fixture: dict[str, Any] = dict(scenario["render_fixture"])
    if quality is not None:
        _require(quality in ("balanced", "performance"), "quality must be balanced or performance")
        fixture["quality"] = quality
    if window_size is not None:
        fixture["window_size"] = window_size
    _require(scenario.get("runner") == "godot_client", "scenario must use the live Godot client")
    raw_client = client_log.read_text(encoding="utf-8", errors="replace")
    _require(f"[bot-client] PASS {scenario['id']}" in raw_client,
             "client log does not contain a passing scenario result")
    client_samples = parse_client_perf_lines(client_log)
    split = split_warmup(client_samples)
    steady = split.steady
    _require(len(steady) >= fixture["min_steady_batches"],
             f"only {len(steady)} steady client batches; need {fixture['min_steady_batches']}")
    frames, missing, dropped = frame_intervals_ms(steady)
    _require(not missing and not dropped, f"incomplete frame batches: missing={missing}, dropped={dropped}")
    _require(len(frames) >= fixture["min_steady_frames"],
             f"only {len(frames)} steady frames; need {fixture['min_steady_frames']}")
    ticks = [int(row["tick"]) for row in steady]
    _require(max(ticks) - min(ticks) >= fixture["min_steady_tick_span"],
             "steady client tick span is too short")
    width, height = _expected_size(fixture["window_size"])
    expected = {
        "renderer": fixture["renderer"], "quality": fixture["quality"],
        "camera": fixture["camera"], "width": width, "height": height,
        "world": scenario["world_id"], "seed": scenario["seed"],
        "live_monsters": fixture["live_monsters"], "ws": 1.0,
        "vsync": float(fixture["vsync"]),
    }
    if "interactables" in fixture:
        expected["interactables"] = fixture["interactables"]
    for index, row in enumerate(steady):
        for key, wanted in expected.items():
            _require(row.get(key) == wanted,
                     f"client batch {index} {key}: expected {wanted!r}, got {row.get(key)!r}")
    server_samples = [row for row in parse_perf_samples(server_log)
                      if min(ticks) <= int(row.get("tick", -1)) <= max(ticks)]
    # backend_perf is emitted about every ten server ticks, not every tick.
    _require(len(server_samples) >= max(1, fixture["min_steady_tick_span"] // 12),
             f"only {len(server_samples)} matching server ticks")
    server_expected = {
        "game_level": fixture["current_level"], "walls": fixture["wall_count"],
        "live_monsters": fixture["live_monsters"],
    }
    if "interactables" in fixture:
        server_expected["interactables"] = fixture["interactables"]
    for index, row in enumerate(server_samples):
        for key, wanted in server_expected.items():
            _require(row.get(key) == wanted,
                     f"server tick {index} {key}: expected {wanted!r}, got {row.get(key)!r}")
    raw_lines = raw_client.splitlines()
    godot = next((line for line in raw_lines if line.startswith("Godot Engine")), "unknown")
    device = next((line for line in raw_lines if "Using Device" in line), "unknown")
    pre_spawn = client_samples.index(split.hitch) if split.hitch is not None else 0
    settle = len(split.warmup) - pre_spawn - 1
    lines = [
        f"Dungeon frame pacing: {scenario['id']}",
        "Session: live bot PASS sentinel verified; session ID omitted from retained counters",
        f"Host: {platform.platform()} | {platform.machine()}",
        f"Godot: {godot}", f"Renderer device: {device}",
        f"Renderer method: {fixture['renderer']} | quality: {fixture['quality']} | size: {width}x{height}",
        f"Camera route: {fixture['camera_route']} | world: {scenario['world_id']} | seed: {scenario['seed']}",
        f"Level: {fixture['current_level']} | walls: {fixture['wall_count']} | live monsters: {fixture['live_monsters']}",
        f"Catalog SHA256: {_catalog_sha256()} | checkout: {_git_revision()}",
        f"Vsync mode: {int(steady[0]['vsync']) if 'vsync' in steady[0] else 'unknown'}",
        f"Warmup: {len(split.warmup)} one-second batches excluded "
        f"(pre-spawn {pre_spawn}, first spawn 1, settle {settle})",
        f"First-spawn process: {float(split.hitch.get('process_ms', 0)):.2f}ms" if split.hitch else "",
        f"Steady: {len(steady)} client batches, {len(frames)} frame intervals, {len(server_samples)} server ticks; client ticks {min(ticks)}..{max(ticks)}",
        _pct_line("Frame interval", frames, "ms"),
        _pct_line("One-second average frame", values(steady, "avg_frame_ms"), "ms"),
        _pct_line("Process time", values(steady, "process_ms"), "ms"),
        _pct_line("Draw calls", values(steady, "draw_calls"), ""),
        _pct_line("Primitives", values(steady, "primitives"), ""),
        _pct_line("Rendered objects", values(steady, "objects"), ""),
        _pct_line("Server tick total", values(server_samples, "total_ms"), "ms"),
        f"Raw client log: {client_log.resolve()}",
        f"Raw server log: {server_log.resolve()}",
    ]
    return "\n".join(lines) + "\n"


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--scenario", type=Path, default=DEFAULT_SCENARIO)
    parser.add_argument("--client-log", type=Path, required=True)
    parser.add_argument("--server-log", type=Path, required=True)
    parser.add_argument("--quality", choices=("balanced", "performance"),
                        help="Expected quality when BOT_CLIENT_RENDER_QUALITY overrides the fixture")
    parser.add_argument("--window-size", help="Expected widthxheight when BOT_CLIENT_RENDER_SIZE overrides the fixture")
    parser.add_argument("--out", type=Path)
    args = parser.parse_args()
    try:
        report = build_report(args.scenario, args.client_log, args.server_log,
                              quality=args.quality, window_size=args.window_size)
    except (ValueError, KeyError, OSError) as exc:
        parser.exit(1, f"dungeon frame report: {exc}\n")
    print(report, end="")
    if args.out:
        args.out.parent.mkdir(parents=True, exist_ok=True)
        args.out.write_text(report, encoding="utf-8")


if __name__ == "__main__":
    main()
