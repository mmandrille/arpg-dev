from __future__ import annotations

import pytest

from tools.bot.skill_demo import all_skill_demo_entries
from tools.bot.skill_visual import SKILL_VISUAL_SCENARIO, build_plan, print_skill_visual_matrix, run_skill_visual, skill_visual_matrix


def test_current_skills_use_single_visual_scenario() -> None:
    for entry in all_skill_demo_entries():
        skill_id = entry.skill_id
        plan = build_plan(skill_id)
        assert plan.skill.skill_id == skill_id
        assert plan.scenario_id == SKILL_VISUAL_SCENARIO
        assert plan.command[-1].endswith("scripts/bot_visual.sh")


def test_run_skill_visual_dry_run_does_not_launch_process(capsys: pytest.CaptureFixture[str]) -> None:
    assert run_skill_visual("holy_shield", dry_run=True) == 0

    out = capsys.readouterr().out
    assert "skill=holy_shield" in out
    assert "category=stat_buff" in out
    assert "scenario=skill_visual" in out


def test_run_all_skill_visual_dry_run_uses_catalog(capsys: pytest.CaptureFixture[str]) -> None:
    entries = all_skill_demo_entries()

    assert run_skill_visual("all", dry_run=True) == 0

    out = capsys.readouterr().out
    assert f"skill_visual all count={len(entries)} scenario=skill_visual" in out
    for entry in entries:
        assert f"skill={entry.skill_id}" in out


def test_missing_skill_fails_clearly() -> None:
    with pytest.raises(ValueError, match="missing required skill id"):
        run_skill_visual("", dry_run=True)


def test_unknown_skill_fails_before_scenario_mapping() -> None:
    with pytest.raises(KeyError, match="unknown skill_id 'meteor'"):
        build_plan("meteor")


def test_skill_visual_matrix_reports_current_coverage() -> None:
    rows = {str(row["skill_id"]): row for row in skill_visual_matrix()}

    assert set(rows) == {entry.skill_id for entry in all_skill_demo_entries()}
    for row in rows.values():
        assert row["scenario_id"] == SKILL_VISUAL_SCENARIO
        assert row["rank1_visual"] is True
        assert row["rank5_visual"] is False
        assert row["buff_stat_delta_visual"] is False


def test_print_skill_visual_matrix(capsys: pytest.CaptureFixture[str]) -> None:
    print_skill_visual_matrix()

    out = capsys.readouterr().out
    assert "skill_id\tclass\tcategory\ticon\tranks\tscenario\trank1\trank5\tbuff_stats" in out
    assert "holy_shield\tpaladin\tstat_buff\tS\t1,5\tskill_visual\tyes\tno\tno" in out
