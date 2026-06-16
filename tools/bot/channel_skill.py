"""Channel-skill bot steps."""

from __future__ import annotations

import json
from typing import Any


async def execute_channel_skill_path(ws, session_id: str, state, step: dict[str, Any], loop, helpers: dict[str, Any]) -> None:
    skill_id = str(step.get("skill_id", "charge"))
    segments = step.get("segments", [])
    if not isinstance(segments, list) or not segments:
        raise AssertionError("channel_skill_path requires non-empty segments")

    make_envelope = helpers["make_envelope"]
    wait_for_accept = helpers["wait_for_accept"]
    wait_for_matching_event = helpers["wait_for_matching_event"]
    pump_one = helpers["pump_one"]
    event_start_index = len(state.events)

    first = segments[0]
    direction = first.get("direction", {"x": 1, "y": 0})
    env = make_envelope(
        "channel_skill_intent",
        session_id,
        state.last_tick,
        {
            "skill_id": skill_id,
            "phase": "start",
            "direction": {"x": float(direction.get("x", 0)), "y": float(direction.get("y", 0))},
        },
    )
    await ws.send(json.dumps(env))
    await wait_for_accept(ws, state, env["message_id"], loop)
    await wait_for_matching_event(ws, state, {"event_type": "skill_channel_started", "skill_id": skill_id}, loop, start_index=event_start_index)

    for index, segment in enumerate(segments):
        direction = segment.get("direction", {"x": 1, "y": 0})
        if index > 0:
            env = make_envelope(
                "channel_skill_intent",
                session_id,
                state.last_tick,
                {
                    "skill_id": skill_id,
                    "phase": "update",
                    "direction": {"x": float(direction.get("x", 0)), "y": float(direction.get("y", 0))},
                },
            )
            await ws.send(json.dumps(env))
            await wait_for_accept(ws, state, env["message_id"], loop)
        for _ in range(max(1, int(segment.get("ticks", 1)))):
            before = state.last_tick
            env = make_envelope(
                "move_intent",
                session_id,
                state.last_tick,
                {"direction": {"x": 0, "y": 0}, "duration_ticks": 1},
            )
            await ws.send(json.dumps(env))
            await wait_for_accept(ws, state, env["message_id"], loop)
            deadline = loop.time() + 1.0
            while state.last_tick <= before and loop.time() <= deadline:
                await pump_one(ws, state, timeout=0.1)

    env = make_envelope("channel_skill_intent", session_id, state.last_tick, {"skill_id": skill_id, "phase": "stop"})
    await ws.send(json.dumps(env))
    await wait_for_accept(ws, state, env["message_id"], loop)
    await wait_for_matching_event(ws, state, {"event_type": "skill_channel_ended", "skill_id": skill_id}, loop, start_index=event_start_index)
