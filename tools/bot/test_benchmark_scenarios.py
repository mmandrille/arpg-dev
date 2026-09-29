"""Fast guards so ci_tier=benchmark scenarios cannot silently rot.

Benchmark scenarios are excluded from the `ci` / `all` packs, so rule changes
used to break them unnoticed (v469-v475: paladin charge loop ran out of mana).
These checks run in `make test-py` / `make ci` without a server.
"""

from __future__ import annotations

import copy
import json
from pathlib import Path

from tools.bot.benchmark_client_stats import percentile, render_client_block, split_warmup
from tools.bot.benchmark_mana_budget import format_shortfalls, load_mana_rules, project_scenario_mana
from tools.bot.benchmark_report import parse_client_perf_lines, render_report
from tools.bot.run import load_scenarios, select_scenarios


def _benchmark_raw() -> list[tuple[str, dict]]:
    scenarios = select_scenarios(load_scenarios(), "benchmark")
    return [(s.id, json.loads(s.path.read_text(encoding="utf-8"))) for s in scenarios]


def test_benchmark_tier_is_not_empty():
    assert _benchmark_raw(), "no ci_tier=benchmark scenarios found"


def test_benchmark_scenarios_pinned_stats_sustain_their_skill_loop():
    rules = load_mana_rules()
    failures = []
    for scenario_id, raw in _benchmark_raw():
        projection = project_scenario_mana(raw, rules)
        if projection.shortfalls:
            failures.append(format_shortfalls(scenario_id, projection))
    assert not failures, "\n".join(failures)


def _synthetic_rules(cost: int, per_magic_mana: float) -> dict:
    return {
        "skills": {
            "zap": {"cost": {"mana": {"base": cost, "per_rank": 0}}},
            "dash": {"cost": {"mana": {"base": cost, "per_rank": 0}}, "mobility": {"channel_mana_per_10_seconds": 0}},
        },
        "derived_stats": {
            "max_mana": {"type": "linear", "base": 0, "per_magic": per_magic_mana, "min": 0},
            "mana_regen_per_second": {"type": "linear", "base": 0, "per_magic": 0, "min": 0},
        },
        "skill_mana_scaling": {"type": "compound_percent", "percent_per_rank": 10},
    }


def _loop_scenario(magic: int, casts: int) -> dict:
    return {
        "debug_progression": {"stats": {"magic": magic}, "skill_ranks": {"zap": 1}},
        "steps": [{"action": "cast_skill", "skill_id": "zap"} for _ in range(casts)],
    }


def test_mana_projection_follows_rule_costs_not_constants():
    scenario = _loop_scenario(magic=10, casts=3)

    assert not project_scenario_mana(scenario, _synthetic_rules(cost=3, per_magic_mana=1)).shortfalls
    pricier = project_scenario_mana(scenario, _synthetic_rules(cost=4, per_magic_mana=1))
    assert [s.step_index for s in pricier.shortfalls] == [2]
    smaller_pool = project_scenario_mana(scenario, _synthetic_rules(cost=3, per_magic_mana=0.5))
    assert smaller_pool.shortfalls


def test_mana_projection_credits_regen_only_for_fixed_duration_steps():
    rules = _synthetic_rules(cost=5, per_magic_mana=1)
    rules["derived_stats"]["mana_regen_per_second"] = {"type": "linear", "base": 1, "min": 0}
    base = {"debug_progression": {"stats": {"magic": 5}}, "steps": [
        {"action": "cast_skill", "skill_id": "zap"},
        {"action": "wait_until_assertion", "timeout_s": 30},
        {"action": "cast_skill", "skill_id": "zap"},
    ]}
    assert project_scenario_mana(base, rules).shortfalls

    waited = copy.deepcopy(base)
    waited["steps"][1] = {"action": "wait_ticks", "ticks": 200}
    assert not project_scenario_mana(waited, rules).shortfalls


def test_mana_projection_charges_channel_start_cost_and_drain():
    rules = _synthetic_rules(cost=2, per_magic_mana=1)
    rules["skills"]["dash"]["mobility"]["channel_mana_per_10_seconds"] = 200
    scenario = {
        "debug_progression": {"stats": {"magic": 10}},
        "steps": [{"action": "channel_skill_path", "skill_id": "dash", "segments": [{"ticks": 8}]}] * 2,
    }
    projection = project_scenario_mana(scenario, rules)
    assert projection.shortfalls and projection.shortfalls[0].step_index == 1


def _sample(**fields: float) -> dict[str, float]:
    base = {"fps": 60, "vsync": 0, "avg_frame_ms": 16.7, "process_ms": 5.0, "entities": 10}
    base.update(fields)
    return base


def test_split_warmup_reports_hitch_and_drops_settle_samples():
    samples = [
        _sample(entities=0, process_ms=1.0),
        _sample(process_ms=700.0),
        _sample(process_ms=40.0),
        _sample(process_ms=35.0),
        _sample(process_ms=5.0),
        _sample(process_ms=6.0),
    ]
    split = split_warmup(samples, settle_samples=2)

    assert split.hitch is samples[1]
    assert len(split.warmup) == 4
    assert [s["process_ms"] for s in split.steady] == [5.0, 6.0]


def test_split_warmup_short_run_keeps_post_hitch_samples():
    samples = [_sample(process_ms=700.0), _sample(process_ms=9.0)]
    split = split_warmup(samples, settle_samples=5)
    assert [s["process_ms"] for s in split.steady] == [9.0]


def test_percentile_is_nearest_rank():
    vals = [float(v) for v in range(1, 101)]
    assert percentile(vals, 0.5) == 50.0
    assert percentile(vals, 0.95) == 95.0
    assert percentile(vals, 1.0) == 100.0
    assert percentile([], 0.5) == 0.0


def test_client_block_excludes_hitch_from_frame_percentiles_and_flags_vsync():
    samples = [_sample(process_ms=700.0, vsync=1)] + [_sample(process_ms=5.0, vsync=1) for _ in range(5)]
    text = "\n".join(render_client_block("probe", samples, settle_samples=0))

    assert "FIRST-SPAWN HITCH" in text and "process_ms 700.0" in text
    process_line = next(line for line in text.splitlines() if line.strip().startswith("process_ms"))
    assert "700" not in process_line
    assert "FPS is capped" in text


def test_client_block_does_not_flag_cap_when_frames_run_faster_than_median():
    frames = [4.0, 6.0, 8.0, 9.0, 12.0]
    samples = [_sample(process_ms=700.0)] + [_sample(avg_frame_ms=f) for f in frames]
    text = "\n".join(render_client_block("probe", samples, settle_samples=0))
    assert "vsync  : off" in text and "capped" not in text


def test_parse_client_perf_lines_keeps_base_count_when_phase_reuses_name(tmp_path: Path):
    log = tmp_path / "client.log"
    log.write_text(
        "noise\n[client-perf] fps=60 vsync=0 avg_frame_ms=16.70 entities=39 draw_calls=600 entities=5.01 d_ui=2.0\n",
        encoding="utf-8",
    )
    [row] = parse_client_perf_lines(log)
    assert row["entities"] == 39.0
    assert row["phase_entities"] == 5.01
    assert row["draw_calls"] == 600.0


def test_report_renders_one_client_section_per_scenario():
    sections = [("probe_a", [_sample()] * 4), ("probe_b", [_sample()] * 4)]
    report = render_report({}, [], sections, "now")
    assert "CLIENT — probe_a" in report and "CLIENT — probe_b" in report
    assert "Client samples: 8" in report
