"""Checks that first-spawn summaries use real adjacent frames and reject bad trials."""

from pathlib import Path

import pytest

from tools.bot.first_spawn_report import PREFIX, parse_trace, summarize_trial


def _row(frame: int, *, spawn: bool = False, monsters: int = 0) -> dict:
    row = {
        "frame": frame,
        "tick": 10 + frame,
        "first_spawn": spawn,
        "monsters": monsters,
        "entities": monsters + 1,
        "frame_interval_ms": 100.0 if frame == 2 else 16.0,
        "process_wall_ms": 80.0 if spawn else 3.0,
        "engine_process_ms": 75.0 if frame == 2 else 3.0,
        "phases_ms": {"d_upsert_m": 50.0} if spawn else {},
        "draw_calls": 100,
        "resources": 200,
        "nodes": 300,
        "static_memory_mib": 400.0,
    }
    if frame == 0:
        row.update(renderer="gl_compatibility", graphics_quality="balanced")
    return row


def test_spawn_interval_comes_from_next_frame() -> None:
    rows = [_row(0), _row(1, spawn=True, monsters=36)] + [_row(frame, monsters=36) for frame in range(2, 210)]
    trial = summarize_trial(rows, "[bot 12:00:00] scenario done sorcerer_multigroup_perf_probe")
    assert trial["spawn_process_wall_ms"] == 80.0
    assert trial["pre_spawn_interval_ms"] == 16.0
    assert trial["spawn_frame_interval_ms"] == 100.0
    assert trial["next_engine_process_ms"] == 75.0
    assert trial["model_count"] == 36
    assert trial["spawn_phases_ms"]["d_upsert_m"] == 50.0


@pytest.mark.parametrize("rows,bot_log", [
    ([_row(0), _row(1)], "scenario done probe"),
    ([_row(0), _row(1, spawn=True, monsters=36)], "scenario done probe"),
    ([_row(0), _row(1, spawn=True, monsters=36), _row(2, monsters=36)], "scenario done probe"),
    ([_row(0), _row(1, spawn=True, monsters=36), _row(2)], "scenario failed probe"),
])
def test_invalid_trials_fail(rows: list[dict], bot_log: str) -> None:
    with pytest.raises(ValueError):
        summarize_trial(rows, bot_log)


def test_parse_trace_rejects_frame_gap(tmp_path: Path) -> None:
    import json

    path = tmp_path / "client.log"
    path.write_text("\n".join(PREFIX + json.dumps(row) for row in [_row(0), _row(2)]))
    with pytest.raises(ValueError, match="skipped or repeated frame"):
        parse_trace(path)
