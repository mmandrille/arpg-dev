from __future__ import annotations

import pytest

from tools.bot.skill_demo import all_skill_demo_entries
from tools.bot.skill_visual import SKILL_VISUAL_SCENARIO, build_plan, print_skill_visual_matrix, run_skill_visual, skill_visual_matrix
from tools.bot.skill_visual_runtime import (
    POST_CAST_HOLD_TICKS,
    build_assertions,
    build_steps,
    magic_required_for_mana,
    seed_skill_visual_character,
    selected_skill_visual_level,
    skill_mana_cost,
    skill_required_stats,
)


def test_current_skills_use_single_visual_scenario() -> None:
    for entry in all_skill_demo_entries():
        skill_id = entry.skill_id
        plan = build_plan(skill_id)
        assert plan.skill.skill_id == skill_id
        assert plan.scenario_id == SKILL_VISUAL_SCENARIO
        assert plan.command[-1].endswith("scripts/bot_visual.sh")


def test_run_skill_visual_dry_run_does_not_launch_process(capsys: pytest.CaptureFixture[str]) -> None:
    assert run_skill_visual("holy_shield", dry_run=True, rank=5) == 0

    out = capsys.readouterr().out
    assert "skill=holy_shield" in out
    assert "category=stat_buff" in out
    assert "rank=5" in out
    assert "level=auto" in out
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


def test_rank_must_be_within_skill_max_rank() -> None:
    with pytest.raises(ValueError, match="between 1 and 5"):
        run_skill_visual("holy_shield", dry_run=True, rank=6)


def test_skill_visual_matrix_reports_current_coverage() -> None:
    rows = {str(row["skill_id"]): row for row in skill_visual_matrix()}

    assert set(rows) == {entry.skill_id for entry in all_skill_demo_entries()}
    for row in rows.values():
        assert row["scenario_id"] == SKILL_VISUAL_SCENARIO
        assert row["rank1_visual"] is True
        assert row["rank5_visual"] is False
        assert row["buff_stat_delta_visual"] is False


def test_skill_visual_steps_hold_after_cast_for_two_seconds() -> None:
    for entry in all_skill_demo_entries():
        steps = build_steps(entry)
        assert not any(step.get("action") == "attack_until_event" for step in steps)
        assert not any(step.get("action") == "allocate_skill_point" for step in steps)
        assert steps[-1] == {"action": "wait_ticks", "ticks": POST_CAST_HOLD_TICKS}
        assert POST_CAST_HOLD_TICKS >= 40


def test_skill_visual_assertions_use_seeded_rank_without_rank_update_event() -> None:
    entry = next(entry for entry in all_skill_demo_entries() if entry.skill_id == "holy_shield")
    assertions = build_assertions(entry, rank=5)

    assert {"type": "event_seen", "event_type": "skill_cast", "skill_id": "holy_shield", "rank": 5} in assertions
    assert not any(assertion.get("event_type") == "skill_rank_updated" for assertion in assertions)
    assert any(assertion.get("type") == "skill_progression" and assertion.get("rank") == 5 for assertion in assertions)


def test_skill_visual_rank_requirements_are_derived_from_rules(monkeypatch: pytest.MonkeyPatch) -> None:
    entry = next(entry for entry in all_skill_demo_entries() if entry.skill_id == "holy_shield")
    monkeypatch.setenv("ARPG_SKILL_VISUAL_LEVEL", "")

    assert selected_skill_visual_level(entry, 5) == 5
    assert skill_required_stats("holy_shield", 5) == {"vit": 13, "magic": 13}


def test_skill_visual_seed_payload_sets_rank_level_and_required_stats() -> None:
    class Response:
        def raise_for_status(self) -> None:
            pass

    class Client:
        def __init__(self) -> None:
            self.url = ""
            self.headers: dict[str, str] = {}
            self.payload: dict[str, object] = {}

        def put(self, url: str, *, headers: dict[str, str], json: dict[str, object]) -> Response:
            self.url = url
            self.headers = headers
            self.payload = json
            return Response()

    entry = next(entry for entry in all_skill_demo_entries() if entry.skill_id == "holy_shield")
    client = Client()

    seed_skill_visual_character(client, "access", "debug", "char_123", entry, rank=5, level=5)  # type: ignore[arg-type]

    assert client.url == "/v0/debug/characters/char_123/progression"
    assert client.headers == {"Authorization": "Bearer access", "X-Debug-Token": "debug"}
    assert client.payload["level"] == 5
    assert client.payload["experience"] == 0
    assert client.payload["skill_ranks"] == {"holy_shield": 5}
    assert client.payload["stats"] == {"str": 6, "dex": 4, "vit": 13, "magic": 13}


def test_skill_visual_seed_payload_covers_ranked_mana_cost() -> None:
    class Response:
        def raise_for_status(self) -> None:
            pass

    class Client:
        def __init__(self) -> None:
            self.payload: dict[str, object] = {}

        def put(self, url: str, *, headers: dict[str, str], json: dict[str, object]) -> Response:
            self.payload = json
            return Response()

    entry = next(entry for entry in all_skill_demo_entries() if entry.skill_id == "cleave")
    client = Client()

    seed_skill_visual_character(client, "access", "debug", "char_123", entry, rank=5, level=5)  # type: ignore[arg-type]

    stats = client.payload["stats"]
    assert isinstance(stats, dict)
    mana_cost = skill_mana_cost("cleave", 5)
    assert mana_cost == 12
    assert stats["magic"] >= magic_required_for_mana(mana_cost)


def test_print_skill_visual_matrix(capsys: pytest.CaptureFixture[str]) -> None:
    print_skill_visual_matrix()

    out = capsys.readouterr().out
    assert "skill_id\tclass\tcategory\ticon\tranks\tscenario\trank1\trank5\tbuff_stats" in out
    assert "holy_shield\tpaladin\tstat_buff\tS\t1,5\tskill_visual\tyes\tno\tno" in out
