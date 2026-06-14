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
