#!/usr/bin/env python3
"""Compile shared rules/assets into codex_index.v0.json for the Godot client."""

from __future__ import annotations

import json
import math
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
CONTENT = ROOT / "shared" / "content"
RULES = ROOT / "shared" / "rules"
ASSETS = ROOT / "shared" / "assets"
OUT_PATH = CONTENT / "codex_index.v0.json"


def load(path: Path) -> dict:
    with path.open(encoding="utf-8") as handle:
        return json.load(handle)


def write_json(path: Path, payload: dict) -> None:
    path.write_text(json.dumps(payload, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")


def section(heading: str = "", lines: list[str] | None = None, bullets: list[str] | None = None) -> dict:
    out: dict = {}
    if heading:
        out["heading"] = heading
    if lines:
        out["lines"] = lines
    if bullets:
        out["bullets"] = bullets
    return out


def page(page_id: str, title: str, sections: list[dict], subtitle: str = "") -> dict:
    payload = {"id": page_id, "title": title, "sections": sections}
    if subtitle:
        payload["subtitle"] = subtitle
    return payload


def chapter(chapter_id: str, title: str, pages: list[dict]) -> dict:
    return {"id": chapter_id, "title": title, "pages": pages}


def max_item_level_for_depth(depth: int, levels_per_tier: int) -> int:
    return max(1, math.floor(depth / levels_per_tier))


def format_stat_line(stats: dict) -> str:
    return "STR %d  DEX %d  VIT %d  MAGIC %d" % (
        int(stats.get("str", 0)),
        int(stats.get("dex", 0)),
        int(stats.get("vit", 0)),
        int(stats.get("magic", 0)),
    )


def skill_display_name(skill_id: str, skill_def: dict) -> str:
    return str(skill_def.get("name", skill_id.replace("_", " ").title()))


def skill_detail_lines(skill_id: str, skill_def: dict) -> list[str]:
    lines = [f"Class: {skill_def.get('class', 'any')}", f"Kind: {skill_def.get('kind', 'unknown')}"]
    tree = skill_def.get("tree", {})
    if tree:
        lines.append("Tree tier %d, column %d" % (int(tree.get("tier", 0)), int(tree.get("column", 0))))
    lines.append("Max rank: %d" % int(skill_def.get("max_rank", 1)))
    cost = skill_def.get("cost", {})
    mana = cost.get("mana", {}) if isinstance(cost, dict) else {}
    if isinstance(mana, dict) and "base" in mana:
        lines.append("Mana cost base: %d (+%d per rank)" % (int(mana.get("base", 0)), int(mana.get("per_rank", 0))))
    cooldown = skill_def.get("cooldown", {})
    if isinstance(cooldown, dict) and cooldown.get("type"):
        lines.append("Cooldown: %s" % cooldown.get("type"))
    reqs = skill_def.get("requirements", {})
    if isinstance(reqs, dict):
        lines.append("Requires character level %d" % int(reqs.get("level", 1)))
    return lines


def build_concepts_chapter(overlays: dict) -> dict:
    pages = []
    for concept_id, entry in sorted(overlays.get("concepts", {}).items()):
        pages.append(
            page(
                f"concept:{concept_id}",
                str(entry.get("title", concept_id)),
                [section(lines=[str(entry.get("body", ""))])],
            )
        )
    return chapter("concepts", "Game Concepts", pages)


def build_classes_chapter(progression: dict, class_presentations: dict, skills: dict) -> dict:
    classes = progression.get("classes", {})
    presentations = class_presentations.get("classes", {})
    skill_defs = skills.get("skills", {})
    pages = []
    for class_id in sorted(classes.keys()):
        class_def = classes[class_id]
        presentation = presentations.get(class_id, {})
        icon = presentation.get("icon", {})
        stats = class_def.get("base_stats", {})
        actives_by_tier: dict[int, list[str]] = {}
        passives: list[str] = []
        for skill_id, skill_def in sorted(skill_defs.items()):
            if str(skill_def.get("class", "")) != class_id:
                continue
            display = skill_display_name(skill_id, skill_def)
            if skill_def.get("kind") == "passive_stat_bonus":
                passives.append(display)
                continue
            tier = int(skill_def.get("tree", {}).get("tier", 1))
            actives_by_tier.setdefault(tier, []).append(display)
        sections = [
            section("Starting Stats", lines=[format_stat_line(stats)]),
            section(
                "Identity",
                lines=[
                    "Movement speed: %.0f%%" % (float(class_def.get("base_movement_speed", 0.0)) * 100.0),
                    "Light radius: %.0f" % float(class_def.get("light_radius", 0.0)),
                ],
            ),
        ]
        if icon:
            sections.append(
                section(
                    "Icon",
                    lines=["Label: %s  Shape: %s" % (icon.get("label", ""), icon.get("shape", ""))],
                )
            )
        for tier in sorted(actives_by_tier.keys()):
            sections.append(section(f"Active Skills (Tier {tier})", bullets=sorted(actives_by_tier[tier])))
        if passives:
            sections.append(section("Passives", bullets=sorted(passives)))
        pages.append(
            page(
                f"class:{class_id}",
                str(class_def.get("name", class_id.title())),
                sections,
                subtitle=class_id,
            )
        )
    return chapter("classes", "Classes", pages)


def build_skills_chapter(skills: dict, skill_presentations: dict) -> dict:
    skill_defs = skills.get("skills", {})
    presentations = skill_presentations.get("skills", {})
    pages = []
    for skill_id in sorted(skill_defs.keys()):
        skill_def = skill_defs[skill_id]
        presentation = presentations.get(skill_id, {})
        summary = str(presentation.get("summary", skill_def.get("description", "")))
        sections = [section(lines=skill_detail_lines(skill_id, skill_def))]
        if summary:
            sections.append(section("Summary", lines=[summary]))
        pages.append(page(f"skill:{skill_id}", skill_display_name(skill_id, skill_def), sections))
    return chapter("skills", "Skills", pages)


def stat_label(stat: str) -> str:
    labels = {
        "damage_percent": "damage",
        "attack_speed_percent": "attack speed",
        "reach_percent": "reach",
        "max_mana_percent": "max mana",
    }
    return labels.get(stat, stat.replace("_", " "))


def format_affinity_row(row: dict) -> str:
    class_name = str(row.get("class", "")).title()
    stat = stat_label(str(row.get("stat", "")))
    lo = int(row.get("min", 0))
    hi = int(row.get("max", 0))
    mode = str(row.get("mode", ""))
    if mode == "penalty_if_not_class":
        return "Non-%s: %+d%% to %+d%% %s when equipped" % (class_name, lo, hi, stat)
    return "%s: +%d%% to +%d%% %s when equipped" % (class_name, lo, hi, stat)


def build_item_families_chapter(item_templates: dict, item_presentations: dict) -> dict:
    templates = item_templates.get("templates", {})
    families = item_presentations.get("families", {})
    by_type: dict[str, list[dict]] = {}
    for template in templates.values():
        item_type = str(template.get("item_type", ""))
        if not item_type:
            continue
        by_type.setdefault(item_type, []).append(template)
    pages = []
    for family_id in sorted(families.keys()):
        family = families[family_id]
        icon = family.get("icon", {})
        matching = by_type.get(family_id, [])
        affinity_rows: dict[tuple, dict] = {}
        attack_speeds: list[float] = []
        slots: set[str] = set()
        for template in matching:
            slots.add(str(template.get("slot", "")))
            if "attack_speed" in template:
                attack_speeds.append(float(template.get("attack_speed", 0.0)))
            for row in template.get("class_affinities", []):
                if not isinstance(row, dict):
                    continue
                key = (
                    str(row.get("class", "")),
                    str(row.get("stat", "")),
                    str(row.get("mode", "")),
                    int(row.get("min", 0)),
                    int(row.get("max", 0)),
                )
                affinity_rows[key] = row
        sections = []
        if icon:
            sections.append(
                section(
                    "Presentation",
                    lines=["Icon label: %s  Shape: %s" % (icon.get("label", ""), icon.get("shape", ""))],
                )
            )
        if slots:
            sections.append(section("Typical Slot", lines=[", ".join(sorted(s for s in slots if s))]))
        if attack_speeds:
            lo = min(attack_speeds)
            hi = max(attack_speeds)
            if lo == hi:
                sections.append(section("Attack Speed Baseline", lines=["%.2fx" % lo]))
            else:
                sections.append(section("Attack Speed Baseline", lines=["%.2fx to %.2fx" % (lo, hi)]))
        affinities = [format_affinity_row(row) for row in sorted(affinity_rows.values(), key=lambda r: (r.get("class", ""), r.get("stat", "")))]
        if affinities:
            sections.append(section("Class Affinities", bullets=affinities))
        else:
            sections.append(section("Class Affinities", lines=["No class-specific affinity on current templates."]))
        if not sections:
            sections.append(section(lines=["Family metadata only."]))
        pages.append(page(f"family:{family_id}", family_id.replace("_", " ").title(), sections))
    return chapter("item_families", "Item Families", pages)


def build_resources_chapter(items: dict, main_config: dict, dungeon_generation: dict, overlays: dict) -> dict:
    gameplay = main_config.get("gameplay", {})
    levels_per_tier = int(dungeon_generation.get("item_level_tiers", {}).get("levels_per_tier", 10))
    upgrade_max = int(gameplay.get("item_upgrade_max_level", 1))
    item_defs = items.get("items", {})
    mechanics = overlays.get("mechanics", {})
    pages = []
    for resource_id in ("upgrade_shard", "renew_stone"):
        item_def = item_defs.get(resource_id, {})
        pages.append(
            page(
                f"resource:{resource_id}",
                str(item_def.get("name", resource_id)),
                [
                    section(
                        lines=[
                            "Leveled consumable stored in your bag.",
                            "Merge three stones of the same level at the blacksmith to create one stone of level +1.",
                        ]
                    )
                ],
            )
        )
    resource_intro = mechanics.get("resource_drops", {})
    drop_cfg = gameplay.get("resource_loot_drops", {})
    pool_lines = []
    for entry in drop_cfg.get("pool", []):
        pool_lines.append("%s (weight %d)" % (entry.get("item_def_id", "?"), int(entry.get("weight", 0))))
    pages.append(
        page(
            "mechanic:resource_drops",
            str(resource_intro.get("title", "Resource Drops")),
            [
                section(lines=[str(resource_intro.get("intro", ""))]),
                section(
                    "Drop Chances (%)",
                    bullets=[
                        "Common/rare monster kill: %d" % int(drop_cfg.get("monster_common_rare_chance_percent", 0)),
                        "Champion monster kill: %d" % int(drop_cfg.get("monster_champion_chance_percent", 0)),
                        "Unique monster kill: %d" % int(drop_cfg.get("monster_unique_chance_percent", 0)),
                        "Boss kill: %d" % int(drop_cfg.get("boss_kill_chance_percent", 0)),
                        "Regular chest: %d" % int(drop_cfg.get("chest_regular_chance_percent", 0)),
                        "Boss chest: %d" % int(drop_cfg.get("chest_boss_chance_percent", 0)),
                    ],
                ),
                section("Weighted Pool", bullets=pool_lines or ["(empty)"]),
                section(
                    "Shard Level",
                    lines=[
                        "Dropped resource level rolls uniformly from 1..max(1, floor(depth / %d))." % levels_per_tier
                    ],
                ),
            ],
        )
    )
    upgrade_intro = mechanics.get("blacksmith_upgrade", {})
    pages.append(
        page(
            "mechanic:blacksmith_upgrade",
            str(upgrade_intro.get("title", "Upgrade Item")),
            [
                section(lines=[str(upgrade_intro.get("intro", ""))]),
                section(
                    "Caps",
                    bullets=[
                        "Config max upgrade steps per item: %d" % upgrade_max,
                        "Depth cap at deepest cleared floor: item_level + 1 <= max(1, floor(depth / %d))" % levels_per_tier,
                        "Example at depth 25: max item level %d" % max_item_level_for_depth(25, levels_per_tier),
                    ],
                ),
            ],
        )
    )
    renew_intro = mechanics.get("blacksmith_renew", {})
    pages.append(
        page(
            "mechanic:blacksmith_renew",
            str(renew_intro.get("title", "Renew Item")),
            [section(lines=[str(renew_intro.get("intro", ""))])],
        )
    )
    merge_intro = mechanics.get("shard_merge", {})
    pages.append(
        page(
            "mechanic:shard_merge",
            str(merge_intro.get("title", "Merge Resources")),
            [section(lines=[str(merge_intro.get("intro", ""))])],
        )
    )
    return chapter("resources", "Resources & Crafting", pages)


def build_loot_chapter(treasure_classes: dict, loot_tables: dict, overlays: dict) -> dict:
    mechanics = overlays.get("mechanics", {})
    intro = mechanics.get("treasure_classes", {})
    tc_defs = treasure_classes.get("classes", {})
    example_id = "dungeon_mob_tc_1" if "dungeon_mob_tc_1" in tc_defs else next(iter(sorted(tc_defs.keys())), "")
    pages = [
        page(
            "mechanic:treasure_classes",
            str(intro.get("title", "Treasure Classes")),
            [
                section(lines=[str(intro.get("intro", ""))]),
                section(
                    "Flow",
                    bullets=[
                        "Loot profile selects a treasure_class_id for the event.",
                        "Each attempt rolls success_weight vs no_drop_weight.",
                        "On success, one entry is picked by weight from the table.",
                        "Entries may be fixed items (item_def_id) or equipment templates (item_template_id).",
                    ],
                ),
            ],
        )
    ]
    if example_id:
        example = tc_defs[example_id]
        attempt = example.get("attempts", [{}])[0]
        entry_lines = []
        for entry in attempt.get("entries", [])[:8]:
            if "item_def_id" in entry:
                entry_lines.append("%s (weight %d)" % (entry["item_def_id"], int(entry.get("weight", 0))))
            elif "item_template_id" in entry:
                entry_lines.append("%s template (weight %d)" % (entry["item_template_id"], int(entry.get("weight", 0))))
        pages.append(
            page(
                f"treasure_class:{example_id}",
                f"Example: {example_id}",
                [
                    section(
                        "Attempt",
                        lines=[
                            "Success weight: %d" % int(attempt.get("success_weight", 0)),
                            "No-drop weight: %d" % int(attempt.get("no_drop_weight", 0)),
                        ],
                    ),
                    section("Sample Entries", bullets=entry_lines),
                ],
            )
        )
    profiles = loot_tables.get("loot_tables", {})
    profile_lines = []
    for profile_id in sorted(profiles.keys())[:6]:
        profile = profiles[profile_id]
        if profile.get("treasure_class_id"):
            profile_lines.append("%s -> %s" % (profile_id, profile.get("treasure_class_id", "?")))
    if profile_lines:
        pages[0]["sections"].append(section("Sample Loot Profiles", bullets=profile_lines))
    return chapter("loot", "Loot & Treasure Classes", pages)


def build_index(chapters: list[str] | None = None) -> dict:
    overlays = load(CONTENT / "codex_overlays.v0.json")
    progression = load(RULES / "character_progression.v0.json")
    skills = load(RULES / "skills.v0.json")
    class_presentations = load(ASSETS / "class_presentations.v0.json")
    skill_presentations = load(ASSETS / "skill_presentations.v0.json")
    item_templates = load(RULES / "item_templates.v0.json")
    item_presentations = load(ASSETS / "item_presentations.v0.json")
    items = load(RULES / "items.v0.json")
    main_config = load(RULES / "main_config.v0.json")
    dungeon_generation = load(RULES / "dungeon_generation.v0.json")
    treasure_classes = load(RULES / "treasure_classes.v0.json")
    loot_tables = load(RULES / "loot_tables.v0.json")

    builders = {
        "concepts": lambda: build_concepts_chapter(overlays),
        "classes": lambda: build_classes_chapter(progression, class_presentations, skills),
        "skills": lambda: build_skills_chapter(skills, skill_presentations),
        "item_families": lambda: build_item_families_chapter(item_templates, item_presentations),
        "resources": lambda: build_resources_chapter(items, main_config, dungeon_generation, overlays),
        "loot": lambda: build_loot_chapter(treasure_classes, loot_tables, overlays),
    }
    selected = chapters or list(builders.keys())
    unknown = [name for name in selected if name not in builders]
    if unknown:
        raise SystemExit("unknown codex chapters: %s" % ", ".join(unknown))
    return {"version": 0, "chapters": [builders[name]() for name in selected]}


def main(argv: list[str] | None = None) -> int:
    argv = argv if argv is not None else sys.argv[1:]
    chapters = None
    if argv:
        if argv[0] == "--chapters":
            chapters = argv[1].split(",") if len(argv) > 1 else None
        else:
            chapters = argv[0].split(",")
    payload = build_index(chapters)
    write_json(OUT_PATH, payload)
    page_count = sum(len(chapter.get("pages", [])) for chapter in payload.get("chapters", []))
    print("wrote %s (%d chapters, %d pages)" % (OUT_PATH.relative_to(ROOT), len(payload["chapters"]), page_count))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
