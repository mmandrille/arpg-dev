"""Pure read-only queries over RuntimeState.

Extracted from run.py so runtime modules (movement, waits, assertions) can
import these helpers directly instead of receiving them through run.py's
``globals()``. These functions have no side effects and no dependency on the
WebSocket driver, so any module that uses them stays importable and
unit-testable without importing ``tools.bot.run``.
"""
from __future__ import annotations

import math
from typing import Any

from tools.bot.bot_types import RuntimeState


def find_player(state: RuntimeState) -> dict[str, Any] | None:
    if state.local_player_id:
        player = state.entities.get(state.local_player_id)
        if player is not None and player.get("type") == "player":
            return player
    for entity in state.entities.values():
        if entity.get("type") == "player":
            return entity
    return None


def dict_distance(a: dict[str, Any], b: dict[str, Any]) -> float:
    return math.hypot(float(a["x"]) - float(b["x"]), float(a["y"]) - float(b["y"]))


def event_matches(event: dict[str, Any], expected: dict[str, Any]) -> bool:
    for key in (
        "event_type",
        "entity_id",
        "source_entity_id",
        "target_entity_id",
        "boss_template_id",
        "skill_id",
        "damage_type",
        "pattern_id",
        "phase_kind",
        "reason",
        "state",
        "stance",
        "service",
    ):
        if key in expected and str(event.get(key, "")) != str(expected[key]):
            return False
    for key in (
        "phase_index",
        "duration_ticks",
        "from_level",
        "to_level",
        "rank",
        "mana",
        "heal",
        "damage",
        "raw_damage",
        "mitigated_damage",
        "amount",
        "remaining_ticks",
        "total_ticks",
        "price",
        "total_gold",
        "unspent_stat_points",
        "unspent_skill_points",
    ):
        if key in expected and int(event.get(key, -999999)) != int(expected[key]):
            return False
    if "affordable" in expected and bool(event.get("affordable", False)) != bool(expected["affordable"]):
        return False
    return True


def event_summary(expected: dict[str, Any]) -> str:
    parts = []
    for key in (
        "event_type",
        "entity_id",
        "source_entity_id",
        "target_entity_id",
        "boss_template_id",
        "skill_id",
        "damage_type",
        "rank",
        "mana",
        "heal",
        "damage",
        "raw_damage",
        "mitigated_damage",
        "amount",
        "remaining_ticks",
        "total_ticks",
        "pattern_id",
        "phase_kind",
        "phase_index",
        "duration_ticks",
        "reason",
        "state",
        "service",
        "price",
        "total_gold",
        "unspent_stat_points",
        "unspent_skill_points",
        "affordable",
        "from_level",
        "to_level",
    ):
        if key in expected:
            parts.append(f"{key}={expected[key]}")
    return ", ".join(parts) or str(expected)
