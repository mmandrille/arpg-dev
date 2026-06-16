"""Skill visual-demo metadata derived from shared skill catalogs."""
from __future__ import annotations

from dataclasses import asdict, dataclass
import argparse
import json
from pathlib import Path
from typing import Any

from tools.content_manifest import (
    load_json,
    merge_catalog_files,
    skill_presentation_entries,
    skill_rule_entries,
)


ROOT = Path(__file__).resolve().parents[2]
MANIFEST_PATH = ROOT / "shared" / "content" / "content_libraries.v0.json"


@dataclass(frozen=True)
class SkillDemoEntry:
    skill_id: str
    name: str
    class_id: str
    kind: str
    category: str
    max_rank: int
    rank_targets: list[int]
    targeting: str
    icon_label: str
    icon_shape: str
    icon_color: str
    icon_accent: str
    summary: str

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


def load_skill_rules(manifest_path: Path = MANIFEST_PATH) -> dict[str, Any]:
    manifest = load_json(manifest_path)
    return merge_catalog_files(manifest_path, skill_rule_entries(manifest), "skills")


def load_skill_presentations(manifest_path: Path = MANIFEST_PATH) -> dict[str, Any]:
    manifest = load_json(manifest_path)
    return merge_catalog_files(manifest_path, skill_presentation_entries(manifest), "skills")


def skill_demo_entry(skill_id: str, manifest_path: Path = MANIFEST_PATH) -> SkillDemoEntry:
    skills = load_skill_rules(manifest_path)
    presentations = load_skill_presentations(manifest_path)
    if skill_id not in skills:
        known = ", ".join(sorted(skills))
        raise KeyError(f"unknown skill_id {skill_id!r}; known skills: {known}")

    skill = skills[skill_id]
    presentation = presentations.get(skill_id, {})
    icon = presentation.get("icon", {})
    max_rank = int(skill.get("max_rank", 1))
    rank_targets = [1]
    if max_rank > 1:
        rank_targets.append(max_rank)

    return SkillDemoEntry(
        skill_id=skill_id,
        name=str(skill.get("name", skill_id)),
        class_id=str(skill.get("class", "")),
        kind=str(skill.get("kind", "")),
        category=demo_category(skill),
        max_rank=max_rank,
        rank_targets=rank_targets,
        targeting=str(skill.get("targeting", "")),
        icon_label=str(icon.get("label", "")),
        icon_shape=str(icon.get("shape", "")),
        icon_color=str(icon.get("color", "")),
        icon_accent=str(icon.get("accent", "")),
        summary=str(presentation.get("summary", "")),
    )


def all_skill_demo_entries(manifest_path: Path = MANIFEST_PATH) -> list[SkillDemoEntry]:
    skills = load_skill_rules(manifest_path)
    return [skill_demo_entry(skill_id, manifest_path) for skill_id in sorted(skills)]


def demo_category(skill: dict[str, Any]) -> str:
    kind = str(skill.get("kind", ""))
    if kind in {"projectile_attack", "cold_projectile_attack", "chain_projectile_attack"}:
        return "attack"
    if kind == "cone_attack":
        return "attack"
    if kind == "mobility":
        return "mobility"
    if kind == "area_heal":
        return "heal"
    if kind == "self_buff":
        return "self_buff"
    if kind == "area_stat_buff":
        return "stat_buff"
    if kind == "passive_execute":
        return "passive"
    return "unknown"


def main() -> int:
    parser = argparse.ArgumentParser(description="Print skill visual-demo metadata.")
    parser.add_argument("skill_id", nargs="?", help="Skill id to inspect. Omit to list all skills.")
    args = parser.parse_args()

    if args.skill_id:
        print(json.dumps(skill_demo_entry(args.skill_id).to_dict(), indent=2, sort_keys=True))
    else:
        print(json.dumps([entry.to_dict() for entry in all_skill_demo_entries()], indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
