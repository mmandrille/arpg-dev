from __future__ import annotations

import pytest

from tools.bot.skill_visual import SKILL_SCENARIOS, build_plan, print_skill_visual_matrix, run_skill_visual, skill_visual_matrix


def test_current_skills_have_visual_scenario_mappings() -> None:
    assert SKILL_SCENARIOS == {
        "magic_bolt": "skill_points_and_magic_bolt",
        "rage": "rage_and_heal_skills",
        "heal": "paladin_heal_skill",
        "holy_shield": "paladin_holy_shield",
    }

    for skill_id, scenario_id in SKILL_SCENARIOS.items():
        plan = build_plan(skill_id)
        assert plan.skill.skill_id == skill_id
        assert plan.scenario_id == scenario_id
        assert plan.command[-1].endswith("scripts/bot_visual.sh")


def test_run_skill_visual_dry_run_does_not_launch_process(capsys: pytest.CaptureFixture[str]) -> None:
    assert run_skill_visual("holy_shield", dry_run=True) == 0

    out = capsys.readouterr().out
    assert "skill=holy_shield" in out
    assert "category=stat_buff" in out
    assert "scenario=paladin_holy_shield" in out


def test_missing_skill_fails_clearly() -> None:
    with pytest.raises(ValueError, match="missing required skill id"):
        run_skill_visual("", dry_run=True)


def test_unknown_skill_fails_before_scenario_mapping() -> None:
    with pytest.raises(KeyError, match="unknown skill_id 'meteor'"):
        build_plan("meteor")


def test_skill_visual_matrix_reports_current_coverage() -> None:
    rows = {str(row["skill_id"]): row for row in skill_visual_matrix()}

    assert set(rows) == set(SKILL_SCENARIOS)
    for skill_id, row in rows.items():
        assert row["scenario_id"] == SKILL_SCENARIOS[skill_id]
        assert row["rank1_visual"] is True
        assert row["rank5_visual"] is False
        assert row["buff_stat_delta_visual"] is False


def test_print_skill_visual_matrix(capsys: pytest.CaptureFixture[str]) -> None:
    print_skill_visual_matrix()

    out = capsys.readouterr().out
    assert "skill_id\tclass\tcategory\ticon\tranks\tscenario\trank1\trank5\tbuff_stats" in out
    assert "holy_shield\tpaladin\tstat_buff\tS\t1,5\tpaladin_holy_shield\tyes\tno\tno" in out
