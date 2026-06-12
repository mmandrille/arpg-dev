"""Debug progression payload helpers for protocol bot scenarios."""
from __future__ import annotations

from typing import Any


def debug_progression_body(progression: dict[str, Any]) -> dict[str, Any]:
    return {
        "level": int(progression.get("level", 1)),
        "experience": int(progression.get("experience", 0)),
        "unspent_stat_points": int(progression.get("unspent_stat_points", 0)),
        "unspent_skill_points": int(progression.get("unspent_skill_points", 0)),
        "stats": dict(progression.get("stats", {})),
        "gold": int(progression.get("gold", 0)),
        "skill_ranks": dict(progression.get("skill_ranks", {})),
    }
