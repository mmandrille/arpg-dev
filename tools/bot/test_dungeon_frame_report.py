"""Focused guards for the live dungeon trace format and fixture failure modes."""

from __future__ import annotations

import json
from pathlib import Path

import pytest

from tools.bot.benchmark_client_stats import frame_intervals_ms, split_warmup
from tools.bot.benchmark_report import parse_client_perf_lines, parse_perf_samples
from tools.bot.dungeon_frame_report import build_report


def _trace(tmp_path: Path, *, changed_quality: bool = False, missing_batch: bool = False) -> tuple[Path, Path, Path]:
    scenario = tmp_path / "scenario.json"
    scenario.write_text(json.dumps({
        "id": "dungeon_frame_pacing_probe", "runner": "godot_client",
        "world_id": "generated_wall_lab", "seed": "pinned",
        "render_fixture": {
            "quality": "balanced", "window_size": "1280x720", "renderer": "forward_plus",
            "vsync": 1, "camera": "isometric", "camera_route": "stationary_isometric",
            "current_level": -1, "wall_count": 31, "live_monsters": 24, "interactables": 2,
            "min_steady_batches": 2, "min_steady_frames": 4, "min_steady_tick_span": 1,
        },
    }), encoding="utf-8")
    client = tmp_path / "client.log"
    lines = ["Godot Engine v4.7.2", "Metal - Using Device #0: Test GPU",
             "[bot-client] session ok id=sess_test", "[bot-client] PASS dungeon_frame_pacing_probe"]
    for tick in range(1, 6):
        lines.append(f"[client-perf] tick={tick} ws=1 entities=26 live_monsters=24 vsync=1 "
                     "interactables=2 avg_frame_ms=20.0 process_ms=5.0 draw_calls=80 "
                     "primitives=100000 objects=380")
        if not missing_batch or tick != 5:
            quality = "performance" if changed_quality and tick >= 4 else "balanced"
            intervals = "100000,100000" if tick == 1 else ("10000,100000" if tick == 5 else "10000,10000")
            lines.append(f"[client-frame-batch] tick={tick} n=2 dropped=0 renderer=forward_plus "
                         f"quality={quality} camera=isometric width=1280 height=720 "
                         f"world=generated_wall_lab seed=pinned us={intervals}")
    client.write_text("\n".join(lines), encoding="utf-8")
    server = tmp_path / "server.log"
    server.write_text("\n".join(json.dumps({
        "msg": "backend_perf", "tick": tick, "game_level": -1, "walls": 31,
        "live_monsters": 24, "interactables": 2, "total_ms": 4.0,
    }) for tick in range(4, 6)), encoding="utf-8")
    return scenario, client, server


def test_frame_percentiles_use_raw_steady_intervals_and_parse_server_msg(tmp_path: Path):
    scenario, client, server = _trace(tmp_path)
    rows = parse_client_perf_lines(client)
    steady = split_warmup(rows).steady
    frames, missing, dropped = frame_intervals_ms(steady)
    assert (frames, missing, dropped) == ([10.0, 10.0, 10.0, 100.0], 0, 0)
    assert len(parse_perf_samples(server)) == 2
    report = build_report(scenario, client, server)
    assert "Frame interval: n=4 p50=10.00ms p95=100.00ms p99=100.00ms" in report
    assert "Warmup: 3 one-second batches excluded (pre-spawn 0, first spawn 1, settle 2)" in report


@pytest.mark.parametrize("change,expected", [
    ({"changed_quality": True}, "quality"),
    ({"missing_batch": True}, "incomplete frame batches"),
])
def test_report_rejects_changed_or_incomplete_fixture(tmp_path: Path, change: dict, expected: str):
    scenario, client, server = _trace(tmp_path, **change)
    with pytest.raises(ValueError, match=expected):
        build_report(scenario, client, server)


def test_quality_override_still_checks_every_steady_batch(tmp_path: Path):
    scenario, client, server = _trace(tmp_path, changed_quality=True)
    report = build_report(scenario, client, server, quality="performance")
    assert "quality: performance" in report
