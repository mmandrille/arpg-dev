"""Static mana-sustainability check for skill-spam bot scenarios.

Benchmark scenarios pin ``debug_progression`` stats and then spam skills. When
skill costs or the mana formulas in ``shared/rules`` change, a pinned ``magic``
value can silently stop covering the loop and the scenario dies mid-run with
``not_enough_mana`` (v469-v475 ``paladin_charge_loop_perf_probe``).

This module projects the mana pool across a scenario's steps using only the
loaded rules — cast costs (``skills.v0.json`` + ``skill_mana_scaling``), charge
channel drain (``mobility.channel_mana_per_10_seconds``), ``max_mana`` and
``mana_regen_per_second`` (``character_progression.v0.json`` derived stats).

The projection is deliberately conservative: regen is credited only for steps
with a fixed tick duration (``wait_ticks``, ``move.duration_ticks``, channel
segment ticks); equipment bonuses and time spent in open-ended waits are
ignored. A scenario that passes here has headroom on the live server too.
"""

from __future__ import annotations

import json
import math
from dataclasses import dataclass
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[2]
RULES_DIR = ROOT / "shared" / "rules"

# Mirrors server/internal/game/sim.go tickDuration: seconds of game time per
# deterministic sim tick. Engine constant, not a tuning value.
SIM_TICK_SECONDS = 0.05


@dataclass(frozen=True)
class ManaShortfall:
    step_index: int
    skill_id: str
    needed: float
    available: float


@dataclass(frozen=True)
class ManaProjection:
    max_mana: float
    regen_per_second: float
    lowest_mana: float
    shortfalls: list[ManaShortfall]


def load_mana_rules(rules_dir: Path = RULES_DIR) -> dict[str, Any]:
    skills = json.loads((rules_dir / "skills.v0.json").read_text(encoding="utf-8"))["skills"]
    progression = json.loads((rules_dir / "character_progression.v0.json").read_text(encoding="utf-8"))
    return {
        "skills": skills,
        "derived_stats": progression["derived_stats"],
        "skill_mana_scaling": progression.get("skill_mana_scaling", {}),
    }


def eval_linear_stat_formula(formula: dict[str, Any], stats: dict[str, Any]) -> float:
    """Python mirror of evalProgressionFormula for the linear formula shape."""
    kind = str(formula.get("type", "linear"))
    if kind != "linear":
        raise ValueError(f"mana budget only understands linear stat formulas, got {kind!r}")
    value = float(formula.get("base", 0))
    for stat in ("str", "dex", "vit", "magic"):
        value += float(formula.get(f"per_{stat}", 0)) * float(stats.get(stat, 0))
    if "min" in formula and value < float(formula["min"]):
        value = float(formula["min"])
    if "max" in formula and value > float(formula["max"]):
        value = float(formula["max"])
    return value


def rank_scaled_mana_cost(curve: dict[str, Any], base: int, per_rank: int, rank: int) -> int:
    """Python mirror of rankScaledInt (server/internal/game/skill_rank_scaling.go)."""
    rank = max(1, rank)
    if curve.get("type") == "linear":
        return max(0, base + per_rank * (rank - 1))
    pct = max(0.0, float(curve.get("percent_per_rank", 10)))
    factor = (1.0 + pct / 100.0) ** (rank - 1)
    # Go's math.Round rounds half away from zero; costs are non-negative.
    return max(0, int(math.floor(base * factor + per_rank * (rank - 1) + 0.5)))


def skill_cast_cost(rules: dict[str, Any], skill_id: str, rank: int) -> int:
    skill = rules["skills"].get(skill_id)
    if not isinstance(skill, dict):
        raise AssertionError(f"shared skill rule {skill_id} not found")
    mana = skill.get("cost", {}).get("mana", {})
    return rank_scaled_mana_cost(rules["skill_mana_scaling"], int(mana.get("base", 0)), int(mana.get("per_rank", 0)), rank)


def channel_drain_per_tick(rules: dict[str, Any], skill_id: str) -> float:
    per_10s = float(rules["skills"].get(skill_id, {}).get("mobility", {}).get("channel_mana_per_10_seconds", 0))
    return per_10s * SIM_TICK_SECONDS / 10.0


def project_scenario_mana(raw: dict[str, Any], rules: dict[str, Any] | None = None) -> ManaProjection:
    rules = rules or load_mana_rules()
    progression = raw.get("debug_progression", {})
    stats = progression.get("stats", {})
    ranks = progression.get("skill_ranks", {})
    max_mana = eval_linear_stat_formula(rules["derived_stats"]["max_mana"], stats)
    regen_per_second = eval_linear_stat_formula(rules["derived_stats"]["mana_regen_per_second"], stats)
    regen_per_tick = regen_per_second * SIM_TICK_SECONDS

    mana = max_mana
    lowest = mana
    shortfalls: list[ManaShortfall] = []

    def pass_ticks(ticks: int, drain_per_tick: float = 0.0) -> None:
        nonlocal mana, lowest
        for _ in range(max(0, ticks)):
            mana = min(max_mana, mana + regen_per_tick) - drain_per_tick
            lowest = min(lowest, mana)

    def spend(index: int, skill_id: str, cost: float) -> None:
        nonlocal mana, lowest
        if mana < cost:
            shortfalls.append(ManaShortfall(index, skill_id, cost, mana))
        mana -= cost
        lowest = min(lowest, mana)

    for index, step in enumerate(raw.get("steps", [])):
        action = step.get("action")
        skill_id = str(step.get("skill_id", ""))
        if action == "cast_skill":
            spend(index, skill_id, skill_cast_cost(rules, skill_id, int(ranks.get(skill_id, 1))))
        elif action == "channel_skill_path":
            spend(index, skill_id, skill_cast_cost(rules, skill_id, int(ranks.get(skill_id, 1))))
            ticks = sum(max(1, int(seg.get("ticks", 1))) for seg in step.get("segments", []))
            pass_ticks(ticks, channel_drain_per_tick(rules, skill_id))
        elif action == "wait_ticks":
            pass_ticks(int(step.get("ticks", 0)))
        elif action == "move":
            pass_ticks(int(step.get("duration_ticks", 0)))

    return ManaProjection(max_mana, regen_per_second, lowest, shortfalls)


def format_shortfalls(scenario_id: str, projection: ManaProjection) -> str:
    rows = ", ".join(
        f"step {s.step_index} {s.skill_id} needs {s.needed:.1f} has {s.available:.1f}"
        for s in projection.shortfalls
    )
    return (
        f"{scenario_id}: pinned debug_progression cannot sustain its skill loop under current rules "
        f"(max_mana {projection.max_mana:.1f}, regen {projection.regen_per_second:.2f}/s): {rows}. "
        "Raise debug_progression.stats.magic."
    )
