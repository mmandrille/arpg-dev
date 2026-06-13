"""Skill-domain cross checks for shared validation."""
from __future__ import annotations

import math
from typing import Any


def skill_requirements_for_rank(requirements: dict[str, Any], rank: int) -> dict[str, Any]:
    rank_offset = max(0, int(rank) - 1)
    level = int(requirements.get("level", 0)) + int(requirements.get("level_per_rank", 0)) * rank_offset
    base_stats = requirements.get("stats", {})
    stats_per_rank = requirements.get("stats_per_rank", {})
    stats = {}
    for stat in ("str", "dex", "vit", "magic"):
        required = int(base_stats.get(stat, 0)) + int(stats_per_rank.get(stat, 0)) * rank_offset
        if required > 0:
            stats[stat] = required
    return {"level": level, "stats": stats}


def validate_skill_catalogs(
    report: Any,
    skills: dict[str, Any],
    skill_presentations: dict[str, Any],
    class_defs: dict[str, Any],
    skill_magic_golden: dict[str, Any],
    *,
    base_attack_interval: int,
    min_attack_speed: float,
    max_attack_speed: float,
) -> None:
    magic_bolt = skills.get("skills", {}).get("magic_bolt")
    skill_class_map = {skill_id: skill.get("class", "") for skill_id, skill in skills.get("skills", {}).items()}
    unknown_skill_classes = {skill_id: class_id for skill_id, class_id in skill_class_map.items() if class_id not in class_defs}
    if unknown_skill_classes:
        report.fail("skill classes", f"unknown classes: {unknown_skill_classes}")
    elif (
        skill_class_map.get("magic_bolt") != "sorcerer"
        or skill_class_map.get("rage") != "barbarian"
        or skill_class_map.get("heal") != "paladin"
        or skill_class_map.get("piercing_shot") != "ranger"
        or skill_class_map.get("pinning_shot") != "ranger"
        or skill_class_map.get("volley") != "ranger"
    ):
        report.fail("skill classes", "core class skills must map to their owning classes")
    else:
        report.ok("skill classes reference character classes")
    if magic_bolt is None:
        report.fail("skills magic_bolt", "missing magic_bolt")
    elif magic_bolt.get("kind") != "projectile_attack":
        report.fail("skills magic_bolt", "kind must be projectile_attack")
    elif int(magic_bolt.get("max_rank", 0)) <= 0:
        report.fail("skills magic_bolt", "max_rank must be positive")
    elif magic_bolt.get("targeting") != "direction_or_target":
        report.fail("skills magic_bolt", "targeting must be direction_or_target")
    elif int(magic_bolt.get("tree", {}).get("tier", 0)) <= 0 or int(magic_bolt.get("tree", {}).get("column", 0)) <= 0:
        report.fail("skills magic_bolt", "tree tier/column must be positive")
    elif int(magic_bolt.get("requirements", {}).get("stats", {}).get("magic", 0)) != 5:
        report.fail("skills magic_bolt requirements", "rank 1 magic requirement must be 5")
    elif int(magic_bolt.get("requirements", {}).get("level_per_rank", 0)) != 1:
        report.fail("skills magic_bolt requirements", "level requirement must increase by 1 per rank")
    elif int(magic_bolt.get("requirements", {}).get("stats_per_rank", {}).get("magic", 0)) != 3:
        report.fail("skills magic_bolt requirements", "single-stat requirement must increase by 3 per rank")
    elif float(magic_bolt.get("projectile", {}).get("range", 0)) <= 0 or float(magic_bolt.get("projectile", {}).get("speed", 0)) <= 0:
        report.fail("skills magic_bolt", "range/projectile_speed must be positive")
    elif magic_bolt.get("cooldown", {}).get("type") != "attack_interval_multiplier":
        report.fail("skills magic_bolt", "cooldown type must be attack_interval_multiplier")
    elif float(magic_bolt.get("cooldown", {}).get("multiplier", 0)) <= 0:
        report.fail("skills magic_bolt", "cooldown multiplier must be positive")
    else:
        dmg = magic_bolt["damage"]
        if dmg.get("type") != "rank_linear_range":
            report.fail("skills magic_bolt damage", "damage type must be rank_linear_range")
        else:
            rank_one_min = int(dmg["min_base"])
            rank_one_max = int(dmg["max_base"])
            rank_max_min = rank_one_min + int(dmg["min_per_rank"]) * (int(magic_bolt["max_rank"]) - 1)
            rank_max_max = rank_one_max + int(dmg["max_per_rank"]) * (int(magic_bolt["max_rank"]) - 1)
            if rank_one_max < rank_one_min or rank_max_max < rank_max_min:
                report.fail("skills magic_bolt damage", "damage max must be >= min at every rank")
            else:
                report.ok("skills magic_bolt declarative tuning is valid")

    missing_skill_presentations = sorted(set(skills.get("skills", {})) - set(skill_presentations.get("skills", {})))
    extra_skill_presentations = sorted(set(skill_presentations.get("skills", {})) - set(skills.get("skills", {})))
    if missing_skill_presentations:
        report.fail("skill_presentations coverage", f"missing presentations for {missing_skill_presentations}")
    elif extra_skill_presentations:
        report.fail("skill_presentations keys", f"unknown skills {extra_skill_presentations}")
    elif magic_bolt is not None:
        mismatched_projectiles = []
        for skill_id, skill in skills.get("skills", {}).items():
            projectile_visual = skill.get("projectile", {}).get("visual", "")
            if projectile_visual and skill_presentations["skills"].get(skill_id, {}).get("projectile_visual") != projectile_visual:
                mismatched_projectiles.append(skill_id)
        if mismatched_projectiles:
            report.fail("skill_presentations projectile visuals", f"mismatched skills {mismatched_projectiles}")
        else:
            report.ok("skill presentations cover skill rules")

    if magic_bolt is not None:
        for skill_id, skill in skills.get("skills", {}).items():
            for req in skill.get("requirements", {}).get("skills", []):
                required_id = req.get("skill_id", "")
                required_rank = int(req.get("rank", 0))
                if required_id not in skills.get("skills", {}):
                    report.fail("skills prerequisites", f"{skill_id} references unknown skill {required_id}")
                    break
                if required_rank > int(skills["skills"][required_id]["max_rank"]):
                    report.fail("skills prerequisites", f"{skill_id} requires {required_id} rank beyond max")
                    break
            else:
                continue
            break
        else:
            report.ok("skill prerequisites reference known skills")

    if magic_bolt is not None:
        skill_golden = skill_magic_golden.get("skill", {})
        if skill_golden.get("class") != magic_bolt.get("class"):
            report.fail("skill_points golden skill", "class must match skills.v0.json")
        elif skill_golden.get("tree") != magic_bolt.get("tree"):
            report.fail("skill_points golden skill", "tree must match skills.v0.json")
        elif skill_golden.get("kind") != magic_bolt.get("kind"):
            report.fail("skill_points golden skill", "kind must match skills.v0.json")
        elif skill_golden.get("requirements") != magic_bolt.get("requirements"):
            report.fail("skill_points golden skill", "requirements must match skills.v0.json")
        else:
            report.ok("skill_points golden skill catalog metadata matches rules")

    failed_skill_magic = False
    if magic_bolt is None:
        failed_skill_magic = True
    elif skill_magic_golden["skill"]["skill_id"] != "magic_bolt":
        report.fail("skill_points golden skill", "skill_id must be magic_bolt")
        failed_skill_magic = True
    elif int(skill_magic_golden["skill"]["max_rank"]) != int(magic_bolt["max_rank"]):
        report.fail("skill_points golden skill", "max_rank must match skills.v0.json")
        failed_skill_magic = True
    elif int(skill_magic_golden["attack_speed"]["base_attack_interval_ticks"]) != base_attack_interval:
        report.fail("skill_points golden attack_speed", "base interval must match combat.v0.json")
        failed_skill_magic = True
    elif not math.isclose(float(skill_magic_golden["attack_speed"]["min_effective_attack_speed"]), min_attack_speed, rel_tol=0, abs_tol=0.000001):
        report.fail("skill_points golden attack_speed", "min clamp must match combat.v0.json")
        failed_skill_magic = True
    elif not math.isclose(float(skill_magic_golden["attack_speed"]["max_effective_attack_speed"]), max_attack_speed, rel_tol=0, abs_tol=0.000001):
        report.fail("skill_points golden attack_speed", "max clamp must match combat.v0.json")
        failed_skill_magic = True
    if not failed_skill_magic and magic_bolt is not None:
        cooldown_multiplier = float(magic_bolt["cooldown"]["multiplier"])
        for case in skill_magic_golden["attack_speed"]["cases"]:
            raw_speed = float(case["dex_attack_speed"]) * float(case["weapon_attack_speed"]) * (1 + int(case["item_attack_speed_percent"]) / 100.0)
            effective_speed = min(max(raw_speed, min_attack_speed), max_attack_speed)
            attack_interval = int(math.ceil(base_attack_interval / effective_speed))
            cooldown_ticks = int(math.ceil(attack_interval * cooldown_multiplier))
            if not math.isclose(float(case["expected_effective_attack_speed"]), round(effective_speed, 6), rel_tol=0, abs_tol=0.000001):
                report.fail("skill_points golden attack_speed", f"{case['name']}: effective speed mismatch")
                failed_skill_magic = True
                break
            if int(case["expected_attack_interval_ticks"]) != attack_interval:
                report.fail("skill_points golden attack_speed", f"{case['name']}: attack interval mismatch")
                failed_skill_magic = True
                break
            if int(case["expected_magic_bolt_cooldown_ticks"]) != cooldown_ticks:
                report.fail("skill_points golden attack_speed", f"{case['name']}: cooldown mismatch")
                failed_skill_magic = True
                break
    if not failed_skill_magic and magic_bolt is not None:
        for case in skill_magic_golden["skill"]["rank_requirement_cases"]:
            rank = int(case["rank"])
            if rank < 1 or rank > int(magic_bolt["max_rank"]):
                report.fail("skill_points golden skill", f"rank {rank}: requirement case outside max_rank")
                failed_skill_magic = True
                break
            requirements = skill_requirements_for_rank(magic_bolt["requirements"], rank)
            if int(case["level"]) != requirements["level"] or case["stats"] != requirements["stats"]:
                report.fail("skill_points golden skill", f"rank {rank}: requirement mismatch")
                failed_skill_magic = True
                break
    if not failed_skill_magic and magic_bolt is not None:
        cost = magic_bolt["cost"]["mana"]
        dmg = magic_bolt["damage"]
        for case in skill_magic_golden["skill"]["rank_cases"]:
            rank = int(case["rank"])
            if rank < 1 or rank > int(magic_bolt["max_rank"]):
                report.fail("skill_points golden skill", f"rank {rank}: outside max_rank")
                failed_skill_magic = True
                break
            mana_cost = int(cost["base"]) + int(cost["per_rank"]) * (rank - 1)
            damage = {
                "min": int(dmg["min_base"]) + int(dmg["min_per_rank"]) * (rank - 1),
                "max": int(dmg["max_base"]) + int(dmg["max_per_rank"]) * (rank - 1),
            }
            if int(case["mana_cost"]) != mana_cost or case["damage"] != damage:
                report.fail("skill_points golden skill", f"rank {rank}: mana/damage mismatch")
                failed_skill_magic = True
                break
        if not failed_skill_magic:
            report.ok("skill_points golden matches combat and skills rules")
