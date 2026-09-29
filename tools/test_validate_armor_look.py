"""Tests for tools/validate_armor_look.py (ADR-0018 D5 / P3b)."""
from __future__ import annotations

from tools.validate_armor_look import validate_armor_look


class Report:
    def __init__(self) -> None:
        self.failures: list[str] = []

    def ok(self, _label: str) -> None:
        pass

    def fail(self, label: str, detail: str) -> None:
        self.failures.append(f"{label}: {detail}")


def _slot_matches(rule_slot: str | None, visual_slot: str) -> bool:
    if rule_slot == "ring":
        return visual_slot in {"ring_left", "ring_right"}
    return rule_slot == visual_slot


def _look(**overrides) -> dict:
    look = {
        "regions": {"body": ["_Body"], "headgear": ["_Hat"]},
        "slots": {"chest": {"region": "body", "mode": "tint"}, "head": {"region": "headgear", "mode": "headgear"}},
        "body_precedence": ["chest"],
        "no_world_visual_slots": ["ring_left", "ring_right", "amulet"],
        "items": {"mail": {"color": "#999999"}, "helm": {"color": "#aaaaaa"}},
    }
    look.update(overrides)
    return look


EQUIPPABLES = {"mail": "chest", "helm": "head", "ring": "ring", "sword": "main_hand"}


def _run(look: dict, visuals: dict | None = None, equippables: dict | None = None) -> list[str]:
    report = Report()
    validate_armor_look(report, look, visuals if visuals is not None else {"sword": {}}, equippables or EQUIPPABLES, _slot_matches)
    return report.failures


def test_weapons_armor_and_jewelry_together_cover_every_equippable() -> None:
    assert _run(_look()) == []


def test_armor_without_a_colour_is_uncovered() -> None:
    look = _look(items={"mail": {"color": "#999999"}})
    assert any("helm" in f and "coverage" in f for f in _run(look))


def test_weapon_missing_from_item_visuals_is_uncovered() -> None:
    assert any("sword" in f for f in _run(_look(), visuals={}))


def test_colour_for_a_non_armor_slot_is_rejected() -> None:
    look = _look(items={"mail": {"color": "#999999"}, "helm": {"color": "#aaaaaa"}, "sword": {"color": "#111111"}})
    assert any("sword" in f and "not a tint/headgear slot" in f for f in _run(look))


def test_unknown_region_and_precedence_slot_are_rejected() -> None:
    look = _look(slots={"chest": {"region": "cape", "mode": "tint"}}, body_precedence=["belt"])
    failures = _run(look, equippables={"mail": "chest"})
    assert any("unknown region cape" in f for f in failures)
    assert any("belt" in f and "precedence" in f for f in failures)
