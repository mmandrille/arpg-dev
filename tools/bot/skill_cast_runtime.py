"""cast_skill bot step payload and execution."""

from __future__ import annotations

import json
import math
from typing import Any, Callable

from tools.bot.skill_rules_loader import skill_rule_targeting


def build_cast_skill_payload(
    step: dict[str, Any],
    state: Any,
    skill_id: str,
    find_player: Callable[[Any], dict[str, Any] | None],
    resolve_target: Callable[[Any, dict[str, Any]], dict[str, Any]],
) -> dict[str, Any]:
    payload: dict[str, Any] = {"skill_id": skill_id}
    if bool(step.get("target_self", False)):
        player = find_player(state)
        if player is None:
            raise AssertionError("cast_skill target_self: player not found")
        payload["target_id"] = str(player["id"])
        return payload

    if step.get("target_id") is not None:
        payload["target_id"] = str(step["target_id"])
        return payload

    if step.get("monster_def_id") is not None:
        target = resolve_target(state, step)
        targeting = skill_rule_targeting(skill_id)
        force_target_id = bool(step.get("use_target_id", False))
        if force_target_id or targeting != "direction":
            payload["target_id"] = str(target["id"])
            return payload

        player = find_player(state)
        if player is None:
            raise AssertionError("cast_skill direction: player not found")
        pos = player.get("position", {})
        tpos = target.get("position", {})
        dx = float(tpos.get("x", 0.0)) - float(pos.get("x", 0.0))
        dy = float(tpos.get("y", 0.0)) - float(pos.get("y", 0.0))
        length = math.hypot(dx, dy)
        if length <= 0.0001:
            payload["direction"] = {"x": 1.0, "y": 0.0}
        else:
            payload["direction"] = {"x": dx / length, "y": dy / length}
        return payload

    direction = step.get("direction", {"x": 1, "y": 0})
    payload["direction"] = {"x": float(direction.get("x", 0)), "y": float(direction.get("y", 0))}
    return payload


async def execute_cast_skill(
    ws: Any,
    session_id: str,
    state: Any,
    step: dict[str, Any],
    loop: Any,
    make_envelope: Callable[..., dict[str, Any]],
    wait_for_accept: Callable[..., Any],
    wait_for_reject: Callable[..., Any],
    wait_for_matching_event: Callable[..., Any],
    wait_for_skill_cooldown: Callable[..., Any],
    find_player: Callable[[Any], dict[str, Any] | None],
    resolve_target: Callable[[Any, dict[str, Any]], dict[str, Any]],
) -> None:
    skill_id = str(step.get("skill_id", "magic_bolt"))
    payload = build_cast_skill_payload(step, state, skill_id, find_player, resolve_target)
    event_start_index = len(state.events)
    env = make_envelope("cast_skill_intent", session_id, state.last_tick, payload)
    await ws.send(json.dumps(env))
    expect_reject = step.get("expect_reject")
    if expect_reject:
        await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
        return
    await wait_for_accept(ws, state, env["message_id"], loop)
    if step.get("event_type"):
        expected_event: dict[str, Any] = {"event_type": str(step["event_type"])}
        if step.get("skill_id") is not None:
            expected_event["skill_id"] = skill_id
        await wait_for_matching_event(ws, state, expected_event, loop, start_index=event_start_index)
    expected = step.get("expect_skill_cooldown")
    if isinstance(expected, dict):
        await wait_for_skill_cooldown(ws, state, expected, loop)
