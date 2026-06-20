"""Class-foundation scenario coverage checks.

These helpers keep class/skill coverage derived from shared rules instead of
hardcoding the current class catalog in tests.
"""
from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any

from tools.bot.bot_types import Scenario


@dataclass(frozen=True)
class CoverageResult:
    missing_scenarios: dict[str, str]
    missing_skills: dict[str, list[str]]

    def assert_complete(self) -> None:
        messages: list[str] = []
        if self.missing_scenarios:
            missing = ", ".join(f"{class_id}->{scenario_id}" for class_id, scenario_id in sorted(self.missing_scenarios.items()))
            messages.append(f"missing class-foundation scenarios: {missing}")
        if self.missing_skills:
            missing = ", ".join(
                f"{class_id}: {', '.join(skill_ids)}"
                for class_id, skill_ids in sorted(self.missing_skills.items())
            )
            messages.append(f"missing class skill coverage: {missing}")
        if messages:
            raise AssertionError("; ".join(messages))


def load_class_ids(rules_dir: Path) -> set[str]:
    raw = json.loads((rules_dir / "character_progression.v0.json").read_text())
    classes = raw.get("classes", {})
    if not isinstance(classes, dict):
        raise AssertionError("character_progression.v0.json classes must be an object")
    return set(str(class_id) for class_id in classes)


def load_skills_by_class(rules_dir: Path, class_ids: set[str]) -> dict[str, set[str]]:
    raw = json.loads((rules_dir / "skills.v0.json").read_text())
    skills = raw.get("skills", {})
    if not isinstance(skills, dict):
        raise AssertionError("skills.v0.json skills must be an object")
    by_class: dict[str, set[str]] = {class_id: set() for class_id in class_ids}
    for skill_id, skill in skills.items():
        if not isinstance(skill, dict):
            continue
        if skill.get("kind") == "passive_stat_bonus":
            continue
        class_id = str(skill.get("class", ""))
        if class_id in by_class:
            by_class[class_id].add(str(skill_id))
    return by_class


def validate_class_foundation_coverage(rules_dir: Path, scenarios: list[Scenario]) -> CoverageResult:
    class_ids = load_class_ids(rules_dir)
    skills_by_class = load_skills_by_class(rules_dir, class_ids)
    scenarios_by_id = {scenario.id: scenario for scenario in scenarios}
    missing_scenarios: dict[str, str] = {}
    missing_skills: dict[str, list[str]] = {}

    for class_id in sorted(class_ids):
        scenario_id = f"{class_id}_class_foundation"
        scenario = scenarios_by_id.get(scenario_id)
        if scenario is None:
            missing_scenarios[class_id] = scenario_id
            continue
        covered = _referenced_skill_ids(scenario)
        missing = sorted(skills_by_class.get(class_id, set()) - covered)
        if missing:
            missing_skills[class_id] = missing

    return CoverageResult(missing_scenarios=missing_scenarios, missing_skills=missing_skills)


def _referenced_skill_ids(scenario: Scenario) -> set[str]:
    referenced: set[str] = set()
    skill_ranks = scenario.debug_progression.get("skill_ranks")
    if isinstance(skill_ranks, dict):
        referenced.update(str(skill_id) for skill_id in skill_ranks)
    for item in [*scenario.steps, *scenario.assertions]:
        _collect_skill_ids(item, referenced)
    for check in scenario.fresh_session_checks:
        for key in ("steps", "assertions"):
            raw_items = check.get(key, [])
            if isinstance(raw_items, list):
                for item in raw_items:
                    _collect_skill_ids(item, referenced)
    return referenced


def _collect_skill_ids(value: Any, out: set[str]) -> None:
    if isinstance(value, dict):
        skill_id = value.get("skill_id")
        if isinstance(skill_id, str) and skill_id:
            out.add(skill_id)
        for child in value.values():
            _collect_skill_ids(child, out)
    elif isinstance(value, list):
        for child in value:
            _collect_skill_ids(child, out)
