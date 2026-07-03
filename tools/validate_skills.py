"""Skill-domain cross checks for shared validation."""
from __future__ import annotations

import math
from typing import Any

SKILL_SYNERGY_MODIFIERS = frozenset(
    {
        "damage_percent",
        "cone_size_percent",
        "volley_spread_percent",
        "projectile_range_percent",
        "buff_duration_percent",
        "buff_power_percent",
        "area_radius_percent",
        "root_duration_percent",
        "slow_duration_percent",
        "mark_duration_percent",
        "bleed_duration_percent",
        "revive_power_percent",
        "passive_stat_percent",
        "execute_threshold_percent",
    }
)


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


def rank_scaled_int(curve: dict[str, Any], base: int, per_rank: int, rank: int) -> int:
    if rank < 1:
        rank = 1
    curve_type = curve.get("type", "compound_percent")
    if curve_type == "linear":
        return max(0, base + per_rank * (rank - 1))
    pct = max(0, int(curve.get("percent_per_rank", 8)))
    factor = (1.0 + pct / 100.0) ** (rank - 1)
    return max(0, int(round(base * factor + per_rank * (rank - 1))))


def validate_skill_tree_layout_hints(report: Any, skills: dict[str, Any]) -> None:
    errors: list[str] = []
    catalog = skills.get("skills", {})
    for skill_id, skill in catalog.items():
        tree = skill.get("tree", {})
        if str(tree.get("branch", "")) == "survival":
            continue
        if str(skill.get("kind", "")) in {"passive_stat_bonus", "passive_execute"}:
            continue
        prereqs = skill.get("requirements", {}).get("skills", [])
        if len(prereqs) != 1:
            continue
        parent_id = str(prereqs[0].get("skill_id", ""))
        parent = catalog.get(parent_id)
        if parent is None:
            continue
        child_tier = int(tree.get("tier", 0))
        parent_tier = int(parent.get("tree", {}).get("tier", 0))
        child_col = int(tree.get("column", 0))
        parent_col = int(parent.get("tree", {}).get("column", 0))
        if child_col == parent_col and child_tier <= parent_tier:
            errors.append(f"{skill_id}: chain column hint must be below parent tier")
        if child_col == parent_col and str(prereqs[0].get("skill_id", "")) != parent_id:
            errors.append(f"{skill_id}: mismatched chain parent")
    if errors:
        report.fail("skill tree layout hints", "; ".join(errors))
    else:
        report.ok("skill tree layout hints are coherent for chain children")


def validate_skill_synergies(report: Any, skills: dict[str, Any]) -> None:
    catalog = skills.get("skills", {})
    errors: list[str] = []

    for skill_id, skill in sorted(catalog.items()):
        required_skills = skill.get("requirements", {}).get("skills", [])
        synergies = skill.get("synergies", [])
        if not required_skills:
            if synergies:
                errors.append(f"{skill_id}: synergies defined without skill prerequisites")
            continue
        if not synergies:
            errors.append(f"{skill_id}: missing synergies for required skills")
            continue

        required_ids = [str(req.get("skill_id", "")) for req in required_skills]
        if any(not source_id for source_id in required_ids):
            errors.append(f"{skill_id}: prerequisite entry missing source_skill_id")
            continue

        synergy_by_source: dict[str, list[dict[str, Any]]] = {}
        for idx, synergy in enumerate(synergies):
            source_id = str(synergy.get("source_skill_id", ""))
            modifier = str(synergy.get("modifier", ""))
            percent = synergy.get("percent_per_source_rank")
            if not source_id:
                errors.append(f"{skill_id}: synergies[{idx}] missing source_skill_id")
                continue
            if source_id not in catalog:
                errors.append(f"{skill_id}: synergies[{idx}] references unknown skill {source_id}")
                continue
            if catalog[source_id].get("class") != skill.get("class"):
                errors.append(f"{skill_id}: synergy source {source_id} must match class {skill.get('class')}")
            if modifier not in SKILL_SYNERGY_MODIFIERS:
                errors.append(f"{skill_id}: synergies[{idx}] uses unknown modifier {modifier!r}")
            if not isinstance(percent, (int, float)) or float(percent) <= 0:
                errors.append(f"{skill_id}: synergies[{idx}] percent_per_source_rank must be positive")
            synergy_by_source.setdefault(source_id, []).append(synergy)

        for source_id in required_ids:
            if source_id not in synergy_by_source:
                errors.append(f"{skill_id}: missing synergy for prerequisite {source_id}")

        for source_id, entries in synergy_by_source.items():
            if source_id not in required_ids:
                errors.append(f"{skill_id}: synergy source {source_id} is not a prerequisite")
            if len(entries) != 1:
                errors.append(f"{skill_id}: duplicate synergy entries for source {source_id}")

    if errors:
        report.fail("skill synergies", "; ".join(errors))
    else:
        report.ok("skills with prerequisites declare synergies")


def validate_skill_catalogs(
    report: Any,
    skills: dict[str, Any],
    skill_presentations: dict[str, Any],
    class_defs: dict[str, Any],
    skill_magic_golden: dict[str, Any],
    *,
    character_progression: dict[str, Any] | None = None,
    base_attack_interval: int,
    min_attack_speed: float,
    max_attack_speed: float,
) -> None:
    rank_curve = (character_progression or {}).get("skill_rank_scaling", {"type": "compound_percent", "percent_per_rank": 8})
    mana_curve = (character_progression or {}).get("skill_mana_scaling", {"type": "compound_percent", "percent_per_rank": 10})
    magic_bolt = skills.get("skills", {}).get("magic_bolt")
    skill_class_map = {skill_id: skill.get("class", "") for skill_id, skill in skills.get("skills", {}).items()}
    unknown_skill_classes = {skill_id: class_id for skill_id, class_id in skill_class_map.items() if class_id not in class_defs}
    if unknown_skill_classes:
        report.fail("skill classes", f"unknown classes: {unknown_skill_classes}")
    elif (
        skill_class_map.get("magic_bolt") != "sorcerer"
        or skill_class_map.get("arcane_barrage") != "sorcerer"
        or skill_class_map.get("rage") != "barbarian"
        or skill_class_map.get("earthbreaker") != "barbarian"
        or skill_class_map.get("heal") != "paladin"
        or skill_class_map.get("sanctuary") != "paladin"
        or skill_class_map.get("shadow_flurry") != "rogue"
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
        if dmg.get("type") != "weapon_multiplier_range":
            report.fail("skills magic_bolt damage", "damage type must be weapon_multiplier_range")
        else:
            rank_one_min = int(dmg["min_base"])
            rank_one_max = int(dmg["max_base"])
            max_rank = int(magic_bolt["max_rank"])
            rank_max_min = rank_scaled_int(rank_curve, rank_one_min, int(dmg["min_per_rank"]), max_rank)
            rank_max_max = rank_scaled_int(rank_curve, rank_one_max, int(dmg["max_per_rank"]), max_rank)
            if rank_one_min < 1 or rank_one_max < 1 or rank_max_min < 1 or rank_max_max < 1:
                report.fail("skills magic_bolt damage", "weapon multiplier percents must be positive at every rank")
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

    passive_chains = {
        "sorcerer": [
            ("arcane_focus", 1, None),
            ("mana_weaving", 5, "arcane_focus"),
            ("spell_dynamo", 10, "mana_weaving"),
            ("arcane_reservoir", 15, "spell_dynamo"),
        ],
        "barbarian": [
            ("iron_hide", 1, None),
            ("battle_tempo", 5, "iron_hide"),
            ("crushing_force", 10, "battle_tempo"),
            ("unstoppable_heart", 15, "crushing_force"),
        ],
        "paladin": [
            ("vigilant_guard", 1, None),
            ("faithful_bulwark", 5, "vigilant_guard"),
            ("consecrated_vitality", 10, "faithful_bulwark"),
            ("oathbound_resolve", 15, "consecrated_vitality"),
        ],
        "rogue": [
            ("quick_hands", 1, None),
            ("killer_instinct", 5, "quick_hands"),
            ("evasive_footwork", 10, "killer_instinct"),
        ],
        "ranger": [
            ("trail_sense", 1, None),
            ("precision_draw", 5, "trail_sense"),
            ("deadeye", 10, "precision_draw"),
            ("wildborn_endurance", 15, "deadeye"),
        ],
    }
    passive_errors = []
    for class_id, rows in passive_chains.items():
        for tier, (skill_id, level, prereq_id) in enumerate(rows, start=1):
            skill = skills.get("skills", {}).get(skill_id)
            if skill is None:
                passive_errors.append(f"{skill_id}: missing")
                continue
            req = skill.get("requirements", {})
            prereqs = req.get("skills", [])
            expected_prereqs = [] if prereq_id is None else [{"skill_id": prereq_id, "rank": 1}]
            if skill.get("class") != class_id:
                passive_errors.append(f"{skill_id}: class {skill.get('class')} != {class_id}")
            if skill.get("kind") != "passive_stat_bonus" or skill.get("targeting") != "self":
                passive_errors.append(f"{skill_id}: must be self passive_stat_bonus")
            if int(skill.get("max_rank", 0)) != 1:
                passive_errors.append(f"{skill_id}: max_rank must be 1")
            tree = skill.get("tree", {})
            if int(tree.get("tier", 0)) != tier or int(tree.get("column", 0)) < 5:
                passive_errors.append(f"{skill_id}: tree must be tier {tier} in the right-side column")
            if int(req.get("level", 0)) != level or int(req.get("level_per_rank", -1)) != 0:
                passive_errors.append(f"{skill_id}: level gate must be {level} with no per-rank scaling")
            if req.get("stats", {}) != {} or req.get("stats_per_rank", {}) != {}:
                passive_errors.append(f"{skill_id}: must not have stat requirements")
            if prereqs != expected_prereqs:
                passive_errors.append(f"{skill_id}: prereqs {prereqs} != {expected_prereqs}")
            if not skill.get("passive_stats", {}).get("stats", {}):
                passive_errors.append(f"{skill_id}: missing passive stat effect")
            if skill.get("cost", {}).get("mana", {}) != {"base": 0, "per_rank": 0}:
                passive_errors.append(f"{skill_id}: passive mana cost must be zero")
            if skill.get("cooldown", {}) != {"type": "none"}:
                passive_errors.append(f"{skill_id}: passive cooldown must be none")
    if passive_errors:
        report.fail("class passive column", "; ".join(passive_errors))
    else:
        report.ok("class passive column chains are valid")

    mobility_kinds = {"mobility"}
    passive_kinds = {"passive_stat_bonus", "passive_execute"}
    survival_kinds = {"survival_autocast"}
    decuple_classes = ("barbarian", "paladin", "sorcerer", "ranger", "rogue")
    expected_actives = {
        "barbarian": 7,
        "paladin": 7,
        "sorcerer": 7,
        "ranger": 7,
        "rogue": 7,
    }
    decuple_errors = []
    for class_id in decuple_classes:
        actives = []
        movement = []
        passives = []
        survival = []
        for skill_id, skill in skills.get("skills", {}).items():
            if skill.get("class") != class_id:
                continue
            kind = skill.get("kind", "")
            if kind in mobility_kinds:
                movement.append(skill_id)
            elif kind in passive_kinds:
                passives.append(skill_id)
            elif kind in survival_kinds:
                survival.append(skill_id)
            else:
                actives.append(skill_id)
        if len(actives) != expected_actives[class_id]:
            decuple_errors.append(f"{class_id}: actives={len(actives)} want {expected_actives[class_id]} ({actives})")
        if len(movement) != 1:
            decuple_errors.append(f"{class_id}: movement={len(movement)} want 1 ({movement})")
        if len(passives) != 4:
            decuple_errors.append(f"{class_id}: passives={len(passives)} want 4 ({passives})")
        if len(survival) != 1:
            decuple_errors.append(f"{class_id}: survival={len(survival)} want 1 ({survival})")
    if decuple_errors:
        report.fail("class skill decuple", "; ".join(decuple_errors))
    else:
        report.ok("each class has expected actives, 1 movement, 4 passives, and 1 survival")

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

    validate_skill_tree_layout_hints(report, skills)
    validate_skill_synergies(report, skills)

    skill_prereqs = {
        skill_id: {
            str(req.get("skill_id", "")): int(req.get("rank", 0))
            for req in skill.get("requirements", {}).get("skills", [])
        }
        for skill_id, skill in skills.get("skills", {}).items()
    }
    expected_prereqs = {
        "arcane_barrage": {"lightning": 1},
        "earthbreaker": {"cleave": 1},
        "sanctuary": {"holy_shield": 1},
        "shadow_flurry": {"dash": 1},
        "pinning_shot": {"piercing_shot": 1},
        "volley": {"piercing_shot": 1},
        "skullcrusher": {"ground_slam": 1},
        "consecrated_smite": {"radiant_bolt": 1},
        "eviscerate": {"poison_stab": 1},
    }
    mismatched_prereqs = {
        skill_id: skill_prereqs.get(skill_id, {})
        for skill_id, prereqs in expected_prereqs.items()
        if skill_prereqs.get(skill_id, {}) != prereqs
    }
    if mismatched_prereqs:
        report.fail("class third skill prerequisites", f"unexpected prerequisites: {mismatched_prereqs}")
    else:
        report.ok("class third skill prerequisites are stable")

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
            mana_cost = rank_scaled_int(mana_curve, int(cost["base"]), int(cost["per_rank"]), rank)
            damage = {
                "min_percent": rank_scaled_int(rank_curve, int(dmg["min_base"]), int(dmg["min_per_rank"]), rank),
                "max_percent": rank_scaled_int(rank_curve, int(dmg["max_base"]), int(dmg["max_per_rank"]), rank),
            }
            if int(case["mana_cost"]) != mana_cost or case["damage"] != damage:
                report.fail("skill_points golden skill", f"rank {rank}: mana/damage mismatch")
                failed_skill_magic = True
                break
        if not failed_skill_magic:
            report.ok("skill_points golden matches combat and skills rules")
