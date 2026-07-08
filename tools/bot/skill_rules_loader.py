"""Shared skill rule lookups for protocol bot steps."""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[2]
_SKILL_RULES: dict[str, Any] | None = None


def _ensure_skill_rules() -> dict[str, Any]:
    global _SKILL_RULES
    if _SKILL_RULES is None:
        skills_path = ROOT / "shared" / "rules" / "skills.v0.json"
        data = json.loads(skills_path.read_text(encoding="utf-8"))
        _SKILL_RULES = dict(data.get("skills", {}))
    return _SKILL_RULES


def skill_rule_max_rank(skill_id: str) -> int:
    skill = _ensure_skill_rules().get(skill_id)
    if not isinstance(skill, dict):
        raise AssertionError(f"shared skill rule {skill_id} not found")
    return int(skill.get("max_rank", -1))


def skill_rule_targeting(skill_id: str) -> str:
    skill = _ensure_skill_rules().get(skill_id)
    if not isinstance(skill, dict):
        raise AssertionError(f"shared skill rule {skill_id} not found")
    return str(skill.get("targeting", "direction"))
