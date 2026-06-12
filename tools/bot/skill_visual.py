"""Validated skill visual replay wrapper."""
from __future__ import annotations

from dataclasses import dataclass
import argparse
import os
from pathlib import Path
import subprocess
from typing import Sequence

from tools.bot.skill_demo import SkillDemoEntry, skill_demo_entry


ROOT = Path(__file__).resolve().parents[2]

SKILL_SCENARIOS = {
    "magic_bolt": "skill_points_and_magic_bolt",
    "rage": "rage_and_heal_skills",
    "heal": "paladin_heal_skill",
    "holy_shield": "paladin_holy_shield",
}


@dataclass(frozen=True)
class SkillVisualPlan:
    skill: SkillDemoEntry
    scenario_id: str
    command: list[str]


def build_plan(skill_id: str, root: Path = ROOT) -> SkillVisualPlan:
    entry = skill_demo_entry(skill_id)
    scenario_id = SKILL_SCENARIOS.get(skill_id)
    if scenario_id is None:
        raise ValueError(f"skill {skill_id!r} has no visual scenario mapping")
    return SkillVisualPlan(
        skill=entry,
        scenario_id=scenario_id,
        command=[str(root / "scripts" / "bot_visual.sh")],
    )


def run_skill_visual(skill_id: str, *, dry_run: bool = False, root: Path = ROOT) -> int:
    if not skill_id:
        raise ValueError("missing required skill id; use skill=<skill_id>")
    plan = build_plan(skill_id, root)
    print(
        "skill_visual skill=%s category=%s ranks=%s scenario=%s"
        % (plan.skill.skill_id, plan.skill.category, plan.skill.rank_targets, plan.scenario_id)
    )
    print("delegates: ARPG_BOT_SCENARIO=%s %s" % (plan.scenario_id, " ".join(plan.command)))
    if dry_run:
        return 0
    env = os.environ.copy()
    env["ARPG_BOT_SCENARIO"] = plan.scenario_id
    return subprocess.call(plan.command, cwd=root, env=env)


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Run a visual replay for a shared skill id.")
    parser.add_argument("skill_id", nargs="?", help="Skill id to visualize.")
    parser.add_argument("--dry-run", action="store_true", help="Print the delegated command without running it.")
    args = parser.parse_args(argv)
    try:
        return run_skill_visual(str(args.skill_id or ""), dry_run=args.dry_run)
    except (KeyError, ValueError) as exc:
        parser.exit(2, f"skill_visual: {exc}\n")


if __name__ == "__main__":
    raise SystemExit(main())
