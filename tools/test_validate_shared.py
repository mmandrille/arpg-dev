from __future__ import annotations

import copy
import json
from pathlib import Path

import pytest
from jsonschema import Draft202012Validator

from tools.content_manifest import ManifestError, merge_catalog_files, skill_rule_entries
from tools.validate_main_config import validate_main_config_gameplay
from tools.validate_skills import validate_skill_catalogs, validate_skill_synergies


ROOT = Path(__file__).resolve().parent.parent
MANIFEST_SCHEMA = ROOT / "shared/content/content_libraries.v0.schema.json"


def load(path: Path):
    with path.open(encoding="utf-8") as fh:
        return json.load(fh)


class CapturingReport:
    def __init__(self) -> None:
        self.failures: list[str] = []

    def ok(self, _label: str) -> None:
        pass

    def fail(self, label: str, detail: str) -> None:
        self.failures.append(f"{label}: {detail}")


def test_content_manifest_schema_rejects_unknown_top_level_group() -> None:
    manifest = {
        "version": 0,
        "rules": {"skills": [{"group": "mage", "path": "../rules/skills.v0.json"}]},
        "assets": {
            "skills": {
                "presentations": [{"group": "default", "path": "../assets/skill_presentations.v0.json"}]
            }
        },
        "runtime": {},
    }
    errors = list(Draft202012Validator(load(MANIFEST_SCHEMA)).iter_errors(manifest))
    assert any("Additional properties" in error.message and "runtime" in error.message for error in errors)


def test_content_manifest_merge_rejects_duplicate_skill_ids(tmp_path: Path) -> None:
    content_dir = tmp_path / "content"
    rules_dir = tmp_path / "rules"
    content_dir.mkdir()
    rules_dir.mkdir()
    (rules_dir / "a.v0.json").write_text('{"version": 0, "skills": {"magic_bolt": {}}}\n', encoding="utf-8")
    (rules_dir / "b.v0.json").write_text('{"version": 0, "skills": {"magic_bolt": {}}}\n', encoding="utf-8")
    manifest_path = content_dir / "content_libraries.v0.json"
    manifest_path.write_text(
        json.dumps(
            {
                "version": 0,
                "rules": {
                    "skills": [
                        {"group": "a", "path": "../rules/a.v0.json"},
                        {"group": "b", "path": "../rules/b.v0.json"},
                    ]
                },
                "assets": {"skills": {"presentations": [{"group": "default", "path": "../rules/a.v0.json"}]}},
            }
        )
        + "\n",
        encoding="utf-8",
    )

    with pytest.raises(ManifestError, match="duplicate skills id magic_bolt"):
        merge_catalog_files(manifest_path, skill_rule_entries(load(manifest_path)), "skills")


def _run_skill_catalog_validation(skills: dict) -> CapturingReport:
    class_defs = load(ROOT / "shared/rules/character_progression.v0.json")["classes"]
    combat = load(ROOT / "shared/rules/combat.v0.json")
    report = CapturingReport()
    validate_skill_catalogs(
        report,
        skills,
        load(ROOT / "shared/assets/skill_presentations.v0.json"),
        class_defs,
        load(ROOT / "shared/golden/skill_points_and_magic_bolt.json"),
        base_attack_interval=int(combat["base_attack_interval_ticks"]),
        min_attack_speed=float(combat["min_effective_attack_speed"]),
        max_attack_speed=float(combat["max_effective_attack_speed"]),
    )
    return report


def test_skill_synergy_validator_requires_synergy_for_prerequisite() -> None:
    skills = copy.deepcopy(load(ROOT / "shared/rules/skills.v0.json"))
    skills["skills"]["snipe"].pop("synergies", None)
    report = CapturingReport()
    validate_skill_synergies(report, skills)
    assert any("skill synergies:" in failure and "snipe: missing synergies" in failure for failure in report.failures)


def test_skill_synergy_validator_rejects_orphan_synergy_source() -> None:
    skills = copy.deepcopy(load(ROOT / "shared/rules/skills.v0.json"))
    skills["skills"]["snipe"]["synergies"] = [
        {
            "source_skill_id": "volley",
            "modifier": "damage_percent",
            "percent_per_source_rank": 10,
        }
    ]
    report = CapturingReport()
    validate_skill_synergies(report, skills)
    assert any(
        "skill synergies:" in failure and "snipe: synergy source volley is not a prerequisite" in failure
        for failure in report.failures
    )


def test_skill_catalog_synergy_coverage_is_valid() -> None:
    skills = load(ROOT / "shared/rules/skills.v0.json")
    report = CapturingReport()
    validate_skill_synergies(report, skills)
    assert report.failures == []


def test_skill_validator_reports_unknown_skill_class() -> None:
    skills = copy.deepcopy(load(ROOT / "shared/rules/skills.v0.json"))
    skills["skills"]["magic_bolt"]["class"] = "ghost"
    report = _run_skill_catalog_validation(skills)

    assert any("skill classes: unknown classes" in failure and "ghost" in failure for failure in report.failures)


def test_main_config_validator_reports_invalid_drop_rate() -> None:
    main_config = copy.deepcopy(load(ROOT / "shared/rules/main_config.v0.json"))
    main_config["gameplay"]["base_drop_rate_percent"] = 101
    dungeon_generation = load(ROOT / "shared/rules/dungeon_generation.v0.json")
    treasure_class_defs = load(ROOT / "shared/rules/treasure_classes.v0.json")["classes"]
    loot = load(ROOT / "shared/rules/loot_tables.v0.json")
    report = CapturingReport()

    def treasure_class_id_for_table(table_id: str) -> str | None:
        loot_def = loot["loot_tables"].get(table_id)
        if loot_def and loot_def.get("mode") == "treasure_class":
            return str(loot_def.get("treasure_class_id"))
        if table_id in treasure_class_defs:
            return table_id
        return None

    validate_main_config_gameplay(
        report,
        main_config["gameplay"],
        dungeon_generation,
        treasure_class_defs,
        treasure_class_id_for_table,
    )

    assert any(
        "main_config gameplay: base_drop_rate_percent must be within [0,100]" in failure
        for failure in report.failures
    )
