"""Protocol-bot movement toward a server-authored boss lane."""

from __future__ import annotations

import math
from typing import Any, Callable

from tools.bot.bot_types import RuntimeState


def boss_lane_target(state: RuntimeState, pattern_id: str, lane_kind: str) -> dict[str, float]:
    if lane_kind not in {"safe", "danger"}:
        raise AssertionError(f"boss lane movement: invalid lane kind {lane_kind}")
    for event in reversed(state.events):
        if event.get("event_type") != "boss_phase_started" or event.get("pattern_id") != pattern_id:
            continue
        lane = event.get("lane")
        if not isinstance(lane, dict):
            continue
        try:
            count = int(lane["count"])
            safe_index = int(lane["safe_index"])
            width = float(lane["width"])
            length = float(lane["length"])
            origin = lane["origin"]
            forward = lane["forward"]
            right = lane["right"]
            danger = lane["danger_lanes"]
            index = safe_index if lane_kind == "safe" else int(danger[-1])
            if not 3 <= count <= 9 or not 0 <= safe_index < count or not 0 <= index < count or width <= 0 or length <= 0:
                continue
            if lane_kind == "danger" and (not isinstance(danger, list) or index == safe_index):
                continue
            along = length / 3.0
            across = (index - (count - 1) / 2.0) * width
            target = {
                "x": float(origin["x"]) + float(forward["x"]) * along + float(right["x"]) * across,
                "y": float(origin["y"]) + float(forward["y"]) * along + float(right["y"]) * across,
            }
            if all(math.isfinite(value) for value in target.values()):
                return target
        except (IndexError, KeyError, TypeError, ValueError):
            continue
    raise AssertionError(f"boss lane movement: no valid {lane_kind} lane for {pattern_id}")


async def move_to_boss_lane_or_position(
    ws: Any,
    session_id: str,
    state: RuntimeState,
    step: dict[str, Any],
    loop: Any,
    walk_toward: Callable[..., Any],
    move_to_position: Callable[..., Any],
    default_max_ticks: int,
) -> None:
    if step["action"] == "move_to_boss_lane":
        target = boss_lane_target(state, str(step.get("pattern_id", "shifting_bulwark")), str(step.get("lane_kind", "safe")))
    else:
        target = {"x": float(step["x"]), "y": float(step["y"])}
    move_fn = move_to_position if bool(step.get("pathfind")) else walk_toward
    await move_fn(
        ws, session_id, state, target, loop,
        stop_distance=float(step.get("tolerance", 0.25)),
        max_ticks=int(step.get("max_ticks", default_max_ticks)),
    )
