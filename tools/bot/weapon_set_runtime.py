from __future__ import annotations

from typing import Any

from tools.bot.bot_types import RuntimeState


def equipped_slot_id(state: RuntimeState, slot: str, weapon_set: Any | None = None) -> Any | None:
    if weapon_set is None or slot not in {"main_hand", "off_hand"}:
        return state.equipped.get(slot)
    wanted = int(weapon_set)
    for row in state.weapon_sets:
        if int(row.get("index", -1)) == wanted:
            return row.get(slot)
    return None
