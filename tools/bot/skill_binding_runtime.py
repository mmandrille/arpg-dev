from __future__ import annotations

import json
from typing import Any, Callable

from tools.bot.bot_types import RuntimeState


async def set_skill_bindings(
    ws: Any,
    session_id: str,
    state: RuntimeState,
    step: dict[str, Any],
    loop: Any,
    make_envelope: Callable[..., dict[str, Any]],
    wait_for_accept: Callable[..., Any],
    wait_for_reject: Callable[..., Any],
    pump_one: Callable[..., Any],
    timeout_s: float,
) -> None:
    keys = [str(row) for row in step.get("function_keys", [])]
    while len(keys) < 16:
        keys.append("")
    keys = keys[:16]
    env = make_envelope("set_skill_bindings_intent", session_id, state.last_tick, {
        "function_keys": keys,
        "right_click_skill_id": str(step.get("right_click_skill_id", "")),
    })
    await ws.send(json.dumps(env))
    expect_reject = step.get("expect_reject")
    if expect_reject:
        await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
        return
    await wait_for_accept(ws, state, env["message_id"], loop)
    deadline = loop.time() + timeout_s
    while state.skill_function_keys != keys:
        if loop.time() > deadline:
            raise TimeoutError("set_skill_bindings stalled waiting for skill_bindings_update")
        await pump_one(ws, state, timeout=0.1)


def assert_skill_bindings(
    state: RuntimeState,
    step: dict[str, Any],
    assert_count_matches: Callable[..., None],
) -> None:
    if step.get("length") is not None:
        assert_count_matches(len(state.skill_function_keys), {"equals": int(step["length"])}, "assert_skill_bindings.length")
    for slot, skill_id in (step.get("slots") or {}).items():
        index = int(slot)
        got = state.skill_function_keys[index] if 0 <= index < len(state.skill_function_keys) else None
        if got != str(skill_id):
            raise AssertionError(f"assert_skill_bindings slot {index} = {got!r}, want {skill_id!r}: {state.skill_function_keys}")
    if step.get("right_click_skill_id") is not None and state.right_click_skill_id != str(step["right_click_skill_id"]):
        raise AssertionError(f"assert_skill_bindings right_click = {state.right_click_skill_id!r}, want {step['right_click_skill_id']!r}")
