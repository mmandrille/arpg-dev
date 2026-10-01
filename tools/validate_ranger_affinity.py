"""Cross-check Deadeye's shared damage rule, skill tags, and golden cases."""
from __future__ import annotations

from typing import Any


def validate_ranger_affinity(report: Any, skills: dict[str, Any], golden: dict[str, Any]) -> None:
    catalog = skills["skills"]
    rule = catalog["deadeye"]["passive_stats"].get("ranger_affinity_damage")
    if rule != golden["rule"]:
        report.fail("ranger affinity golden rule", "must match Deadeye's shared rule")
    else:
        report.ok("ranger affinity golden rule matches Deadeye")

    invalid_tags = [skill_id for skill_id, skill in catalog.items()
                    if "ranger_affinity_eligible" in skill
                    and (skill["class"] != "ranger" or skill["kind"] != "projectile_attack"
                         or skill["ranger_affinity_eligible"] is not True)]
    if invalid_tags:
        report.fail("ranger affinity skill tags", f"unsupported tags: {invalid_tags}")
    else:
        report.ok("ranger affinity skill tags target direct Ranger projectiles")

    for case in golden["cases"]:
        cfg = case.get("rule_override", golden["rule"])
        count = min(case["affinities"], cfg["max_active_affinities"])
        stat = max(0, case["effective_stat"])
        bonus = 0 if case["rank"] <= 0 else min(
            cfg["max_bonus_percent"],
            count * (cfg["base_percent_per_affinity"] + stat // cfg["stat_points_per_extra_percent"]),
        )
        expected_min = case["raw_min"] * (100 + bonus) // 100
        expected_max = case["raw_max"] * (100 + bonus) // 100
        if (bonus, expected_min, expected_max) != (
            case["expected_percent"], case["expected_min"], case["expected_max"]
        ):
            report.fail("ranger affinity golden case", case["name"])
        else:
            report.ok(f"ranger affinity golden {case['name']}")
