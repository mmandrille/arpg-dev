"""Validated skill visual replay wrapper."""
from __future__ import annotations

from dataclasses import dataclass
import argparse
import os
from pathlib import Path
import subprocess
from typing import Sequence

from tools.bot.skill_demo import SkillDemoEntry, all_skill_demo_entries, skill_demo_entry


ROOT = Path(__file__).resolve().parents[2]
SKILL_VISUAL_SCENARIO = "skill_visual"


@dataclass(frozen=True)
class SkillVisualPlan:
    skill: SkillDemoEntry
    scenario_id: str
    command: list[str]


def selected_skill_entries(skill_id: str) -> list[SkillDemoEntry]:
    if not skill_id:
        raise ValueError("missing required skill id; use skill=<skill_id>")
    if skill_id == "all":
        return all_skill_demo_entries()
    return [skill_demo_entry(skill_id)]


def build_plan(skill_id: str, root: Path = ROOT) -> SkillVisualPlan:
    entry = skill_demo_entry(skill_id)
    return SkillVisualPlan(
        skill=entry,
        scenario_id=SKILL_VISUAL_SCENARIO,
        command=[str(root / "scripts" / "bot_visual.sh")],
    )


def run_skill_visual(skill_id: str, *, dry_run: bool = False, root: Path = ROOT) -> int:
    entries = selected_skill_entries(skill_id)
    if skill_id == "all":
        print("skill_visual all count=%d scenario=%s" % (len(entries), SKILL_VISUAL_SCENARIO))
    for entry in entries:
        plan = SkillVisualPlan(
            skill=entry,
            scenario_id=SKILL_VISUAL_SCENARIO,
            command=[str(root / "scripts" / "bot_visual.sh")],
        )
        plan_exit = run_skill_visual_plan(plan, dry_run=dry_run, root=root)
        if plan_exit != 0:
            return plan_exit
    return 0


def run_skill_visual_plan(plan: SkillVisualPlan, *, dry_run: bool = False, root: Path = ROOT) -> int:
    print(
        "skill_visual skill=%s category=%s ranks=%s scenario=%s"
        % (plan.skill.skill_id, plan.skill.category, plan.skill.rank_targets, plan.scenario_id)
    )
    print("delegates: ARPG_BOT_SCENARIO=%s %s" % (plan.scenario_id, " ".join(plan.command)))
    if dry_run:
        return 0
    env = os.environ.copy()
    env["ARPG_BOT_SCENARIO"] = plan.scenario_id
    env["ARPG_SKILL_VISUAL_SKILL_ID"] = plan.skill.skill_id
    return subprocess.call(plan.command, cwd=root, env=env)


def skill_visual_matrix() -> list[dict[str, object]]:
    rows: list[dict[str, object]] = []
    for entry in all_skill_demo_entries():
        rows.append({
            "skill_id": entry.skill_id,
            "class_id": entry.class_id,
            "category": entry.category,
            "icon": entry.icon_label,
            "rank_targets": entry.rank_targets,
            "scenario_id": SKILL_VISUAL_SCENARIO,
            "rank1_visual": True,
            "rank5_visual": False,
            "buff_stat_delta_visual": False,
        })
    return rows


def print_skill_visual_matrix() -> None:
    headers = [
        "skill_id",
        "class",
        "category",
        "icon",
        "ranks",
        "scenario",
        "rank1",
        "rank5",
        "buff_stats",
    ]
    print("\t".join(headers))
    for row in skill_visual_matrix():
        print("\t".join([
            str(row["skill_id"]),
            str(row["class_id"]),
            str(row["category"]),
            str(row["icon"]),
            ",".join(str(rank) for rank in row["rank_targets"]),
            str(row["scenario_id"]),
            _yes_no(bool(row["rank1_visual"])),
            _yes_no(bool(row["rank5_visual"])),
            _yes_no(bool(row["buff_stat_delta_visual"])),
        ]))


def _yes_no(value: bool) -> str:
    return "yes" if value else "no"


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Run a visual replay for a shared skill id.")
    parser.add_argument("skill_id", nargs="?", help="Skill id to visualize.")
    parser.add_argument("--dry-run", action="store_true", help="Print the delegated command without running it.")
    parser.add_argument("--list", action="store_true", help="Print skill visual coverage matrix.")
    args = parser.parse_args(argv)
    if args.list:
        print_skill_visual_matrix()
        return 0
    try:
        return run_skill_visual(str(args.skill_id or ""), dry_run=args.dry_run)
    except (KeyError, ValueError) as exc:
        parser.exit(2, f"skill_visual: {exc}\n")


if __name__ == "__main__":
    raise SystemExit(main())
