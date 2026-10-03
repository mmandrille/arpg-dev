"""Discover data-driven showme screenshot jobs from shared catalogs."""
from __future__ import annotations

import json
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[2]

CLASS_PROGRESSION_REL = "shared/rules/character_progression.v0.json"
SKILLS_REL = "shared/rules/skills.v0.json"
ITEM_PRESENTATIONS_REL = "shared/assets/item_presentations.v0.json"
ITEM_VISUALS_REL = "shared/assets/item_visuals.v0.json"
ARMOR_LOOK_REL = "shared/assets/armor_look.v0.json"
ASSETS_MANIFEST_REL = "assets/manifests/assets.v0.json"
DUNGEON_GENERATION_REL = "shared/rules/dungeon_generation.v0.json"

# Existing showme scene focuses captured as-is by the scenes suite (ADR-0018 D9 baseline).
SCENE_FOCUSES: tuple[str, ...] = ("town", "monsters", "chests", "stairs", "eye-view", "heal-rain")
TOWN_PLAY_VIEWS: tuple[tuple[str, str], ...] = (("plaza", "normal"), ("plaza", "max"), ("vendor", "normal"), ("gate", "max"), ("west", "max"), ("north", "max"))
DUNGEON_ROOM_FOCUS = "dungeon-room"
# v516: panel variants captured for every class; dual-wield layout proven for two classes.
CHARACTER_SCREEN_VARIANTS: tuple[str, ...] = ("points", "nopoints", "paper-doll")
CHARACTER_SCREEN_DUAL_CLASSES: tuple[str, ...] = ("barbarian", "rogue")
ITEM_TOOLTIP_RARITIES: tuple[str, ...] = ("common", "magic", "rare", "unique", "set")
SKILL_TREE_VARIANTS: tuple[str, ...] = ("points", "nopoints", "hover")
HUD_STATES: tuple[str, ...] = ("full", "half", "low", "empty")
HUD_SIZES: tuple[tuple[int, int], ...] = ((1280, 720), (1920, 1080))


@dataclass(frozen=True)
class CaptureJob:
    suite: str
    focus: str
    slug: str
    output_rel: str
    extra_args: tuple[str, ...] = field(default_factory=tuple)


@dataclass(frozen=True)
class SuiteSpec:
    name: str
    focus: str
    description: str


SUITE_SPECS: dict[str, SuiteSpec] = {
    "skeleton": SuiteSpec("skeleton", "skeleton", "Bone and socket markers per player class"),
    "gear": SuiteSpec("gear", "gear", "Default equipped gear set per player class"),
    "skill-icon": SuiteSpec("skill-icon", "skill-icon", "Individual skill icon per skill id"),
    "item-icon": SuiteSpec("item-icon", "item-icon", "Individual item icon per presentation family"),
    "item-icons": SuiteSpec("item-icons", "item-icons", "Grid of all item icon families"),
    "floor-item": SuiteSpec("floor-item", "floor-item", "Ground loot model per item with visuals"),
    "item-asset": SuiteSpec("item-asset", "item-asset", "Isolated 3D asset per item_visuals asset_id"),
    "character-screen": SuiteSpec("character-screen", "character-screen", "Character stats panel and paper-doll backdrop per class (v516)"),
    "item-tooltip": SuiteSpec("item-tooltip", "item-tooltip", "Shared text-only item tooltip per rarity (v532)"),
    "skill-tree": SuiteSpec("skill-tree", "skills", "Skill tree progression, points, hover, and selection states (v519)"),
    "hud": SuiteSpec("hud", "hud", "HUD fixture: vitals globes, hotbar cluster, minimap frame, boss bar (full/half/low/empty at two sizes)"),
    "scenes": SuiteSpec("scenes", "scene", "Town/monster/prop scenes + runtime-lit dungeon room per biome"),
}

DEFAULT_SUITES: tuple[str, ...] = tuple(SUITE_SPECS.keys())


def _read_json(rel_path: str) -> dict[str, Any]:
    path = ROOT / rel_path
    with path.open(encoding="utf-8") as handle:
        parsed = json.load(handle)
    if not isinstance(parsed, dict):
        raise ValueError(f"{rel_path} is not a JSON object")
    return parsed


def class_ids() -> list[str]:
    data = _read_json(CLASS_PROGRESSION_REL)
    classes = data.get("classes", {})
    if not isinstance(classes, dict):
        return []
    return sorted(str(class_id) for class_id in classes)


def skill_ids() -> list[str]:
    data = _read_json(SKILLS_REL)
    skills = data.get("skills", {})
    if not isinstance(skills, dict):
        return []
    return sorted(str(skill_id) for skill_id in skills)


def item_family_ids() -> list[str]:
    data = _read_json(ITEM_PRESENTATIONS_REL)
    families = data.get("families", {})
    if not isinstance(families, dict):
        return []
    return sorted(str(family_id) for family_id in families)


def item_def_ids_with_visuals() -> list[str]:
    """Equipment with a world look: item_visuals (weapons) plus armor_look items (ADR-0018 P3b)."""
    visuals = _read_json(ITEM_VISUALS_REL).get("item_visuals", {})
    armor = _read_json(ARMOR_LOOK_REL).get("items", {})
    ids = set(visuals) if isinstance(visuals, dict) else set()
    if isinstance(armor, dict):
        ids |= set(armor)
    return sorted(str(item_def_id) for item_def_id in ids)


def item_asset_ids() -> list[str]:
    data = _read_json(ITEM_VISUALS_REL)
    visuals = data.get("item_visuals", {})
    if not isinstance(visuals, dict):
        return []
    asset_ids: set[str] = set()
    for entry in visuals.values():
        if not isinstance(entry, dict):
            continue
        asset_id = str(entry.get("asset_id", "")).strip()
        if asset_id:
            asset_ids.add(asset_id)
    return sorted(asset_ids)


def biome_palette_levels() -> list[tuple[str, int]]:
    """Return (palette_id, level) for each biome palette, at its shallowest depth."""
    data = _read_json(DUNGEON_GENERATION_REL)
    palettes = data.get("biome_palettes", [])
    if not isinstance(palettes, list):
        return []
    levels: list[tuple[str, int]] = []
    for palette in palettes:
        if not isinstance(palette, dict):
            continue
        palette_id = str(palette.get("id", "")).strip()
        min_depth = palette.get("min_depth")
        if not palette_id or not isinstance(min_depth, int) or min_depth < 1:
            continue
        levels.append((palette_id, -min_depth))
    return levels


def manifest_asset_ids() -> set[str]:
    data = _read_json(ASSETS_MANIFEST_REL)
    assets = data.get("assets", {})
    if not isinstance(assets, dict):
        return set()
    return {str(asset_id) for asset_id in assets}


def discover_jobs(suites: list[str] | None = None) -> list[CaptureJob]:
    selected = list(DEFAULT_SUITES if suites is None else suites)
    unknown = [suite for suite in selected if suite not in SUITE_SPECS]
    if unknown:
        known = ", ".join(sorted(SUITE_SPECS))
        raise ValueError(f"unknown suite(s): {', '.join(unknown)}; known: {known}")

    jobs: list[CaptureJob] = []
    manifest_ids = manifest_asset_ids()

    for suite_name in selected:
        spec = SUITE_SPECS[suite_name]
        if suite_name == "skeleton":
            for class_id in class_ids():
                jobs.append(CaptureJob(
                    suite=spec.name,
                    focus=spec.focus,
                    slug=class_id,
                    output_rel=f"{spec.name}/{class_id}.png",
                    extra_args=("--class-id", class_id),
                ))
        elif suite_name == "gear":
            for class_id in class_ids():
                jobs.append(CaptureJob(
                    suite=spec.name,
                    focus=spec.focus,
                    slug=class_id,
                    output_rel=f"{spec.name}/{class_id}.png",
                    extra_args=("--class-id", class_id),
                ))
        elif suite_name == "skill-icon":
            for skill_id in skill_ids():
                jobs.append(CaptureJob(
                    suite=spec.name,
                    focus=spec.focus,
                    slug=skill_id,
                    output_rel=f"{spec.name}/{skill_id}.png",
                    extra_args=("--skill-id", skill_id),
                ))
        elif suite_name == "item-icon":
            for family_id in item_family_ids():
                jobs.append(CaptureJob(
                    suite=spec.name,
                    focus=spec.focus,
                    slug=family_id,
                    output_rel=f"{spec.name}/{family_id}.png",
                    extra_args=("--family-id", family_id),
                ))
        elif suite_name == "item-icons":
            jobs.append(CaptureJob(
                suite=spec.name,
                focus=spec.focus,
                slug="all-families",
                output_rel=f"{spec.name}/all-families.png",
            ))
        elif suite_name == "floor-item":
            for item_def_id in item_def_ids_with_visuals():
                jobs.append(CaptureJob(
                    suite=spec.name,
                    focus=spec.focus,
                    slug=item_def_id,
                    output_rel=f"{spec.name}/{item_def_id}.png",
                    extra_args=("--items", item_def_id),
                ))
        elif suite_name == "item-asset":
            for asset_id in item_asset_ids():
                if asset_id not in manifest_ids:
                    continue
                jobs.append(CaptureJob(
                    suite=spec.name,
                    focus=spec.focus,
                    slug=asset_id,
                    output_rel=f"{spec.name}/{asset_id}.png",
                    extra_args=("--asset-id", asset_id),
                ))
        elif suite_name == "character-screen":
            for class_id in class_ids():
                variants = list(CHARACTER_SCREEN_VARIANTS)
                if class_id in CHARACTER_SCREEN_DUAL_CLASSES:
                    variants.append("nopoints-dual")
                for variant in variants:
                    slug = f"{class_id}-{variant}"
                    jobs.append(CaptureJob(
                        suite=spec.name,
                        focus=spec.focus,
                        slug=slug,
                        output_rel=f"{spec.name}/{slug}.png",
                        extra_args=("--class-id", class_id, "--variant", variant),
                    ))
        elif suite_name == "item-tooltip":
            for rarity in ITEM_TOOLTIP_RARITIES:
                jobs.append(CaptureJob(
                    suite=spec.name,
                    focus=spec.focus,
                    slug=rarity,
                    output_rel=f"{spec.name}/{rarity}.png",
                    extra_args=("--variant", rarity),
                ))
        elif suite_name == "skill-tree":
            for variant in SKILL_TREE_VARIANTS:
                jobs.append(CaptureJob(
                    suite=spec.name,
                    focus=spec.focus,
                    slug=variant,
                    output_rel=f"{spec.name}/{variant}.png",
                    extra_args=("--variant", variant),
                ))
        elif suite_name == "hud":
            for state in HUD_STATES:
                for width, height in HUD_SIZES:
                    slug = f"hud-{state}-{width}x{height}"
                    jobs.append(CaptureJob(
                        suite=spec.name,
                        focus="hud",
                        slug=slug,
                        output_rel=f"{spec.name}/{slug}.png",
                        extra_args=("--hud-state", state, "--width", str(width), "--height", str(height)),
                    ))
        elif suite_name == "scenes":
            for focus in SCENE_FOCUSES:
                jobs.append(CaptureJob(
                    suite=spec.name,
                    focus=focus,
                    slug=focus,
                    output_rel=f"{spec.name}/{focus}.png",
                ))
            for view, zoom in TOWN_PLAY_VIEWS:
                slug = f"town-play-{view}-{zoom}"
                jobs.append(CaptureJob(
                    suite=spec.name,
                    focus="town-play",
                    slug=slug,
                    output_rel=f"{spec.name}/{slug}.png",
                    extra_args=("--town-view", view, "--town-zoom", zoom),
                ))
            for palette_id, level in biome_palette_levels():
                slug = f"{DUNGEON_ROOM_FOCUS}-{palette_id}"
                jobs.append(CaptureJob(
                    suite=spec.name,
                    focus=DUNGEON_ROOM_FOCUS,
                    slug=slug,
                    output_rel=f"{spec.name}/{slug}.png",
                    extra_args=("--level", str(level)),
                ))
        else:
            raise ValueError(f"unhandled suite: {suite_name}")

    return jobs


def suite_summary() -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    for suite_name in DEFAULT_SUITES:
        spec = SUITE_SPECS[suite_name]
        count = len([job for job in discover_jobs([suite_name])])
        rows.append({
            "suite": suite_name,
            "focus": spec.focus,
            "description": spec.description,
            "count": count,
        })
    return rows
