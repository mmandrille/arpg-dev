from __future__ import annotations

import pytest

from tools.bot.skill_visual import SKILL_SCENARIOS, build_plan, run_skill_visual


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
