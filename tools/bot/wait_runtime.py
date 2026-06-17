from __future__ import annotations

import asyncio
import json
from typing import Any

from tools.bot.bot_context import BotContext
from tools.bot.bot_types import RuntimeState
from tools.bot.runtime_queries import event_matches, find_player

SLICE_TIMEOUT_S = 20.0


def _require_helpers(helpers: dict[str, Any] | None) -> dict[str, Any]:
    if helpers is None:
        raise AssertionError("wait runtime helpers require helper bindings")
    return helpers


def _pump_one(ctx: BotContext | None, helpers: dict[str, Any] | None):
    if ctx is not None:
        return ctx.pump_one
    return _require_helpers(helpers)["pump_one"]


async def wait_for_tick_advance(ws, state: RuntimeState, loop, helpers: dict[str, Any] | None = None, ctx: BotContext | None = None) -> None:
    pump_one = _pump_one(ctx, helpers)
    start = state.last_tick
    deadline = loop.time() + SLICE_TIMEOUT_S
    while state.last_tick <= start:
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for tick advance from {start}")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_tick(ws, state: RuntimeState, target_tick: int, loop, helpers: dict[str, Any] | None = None, ctx: BotContext | None = None) -> None:
    pump_one = _pump_one(ctx, helpers)
    deadline = loop.time() + SLICE_TIMEOUT_S
    while state.last_tick < target_tick:
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for tick {target_tick}")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_accept(ws, state: RuntimeState, message_id: str, loop, helpers: dict[str, Any] | None = None, ctx: BotContext | None = None) -> None:
    pump_one = _pump_one(ctx, helpers)
    deadline = loop.time() + SLICE_TIMEOUT_S
    while message_id not in state.accepted_message_ids:
        reason = state.rejected_message_reasons.get(message_id)
        if reason is not None:
            raise AssertionError(f"intent {message_id} rejected while waiting for accept: {reason}")
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for accept {message_id}")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_reject(ws, state: RuntimeState, message_id: str, reason: str, loop, helpers: dict[str, Any] | None = None, ctx: BotContext | None = None) -> None:
    pump_one = _pump_one(ctx, helpers)
    deadline = loop.time() + SLICE_TIMEOUT_S
    while message_id not in state.rejected_message_reasons:
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for reject {message_id}")
        await pump_one(ws, state, timeout=0.1)
    got = state.rejected_message_reasons[message_id]
    if got != reason:
        raise AssertionError(f"reject {message_id} reason={got}, want {reason}")


async def wait_for_event(ws, state: RuntimeState, event_type: str, loop, *, timeout_s: float = SLICE_TIMEOUT_S, start_index: int = 0, helpers: dict[str, Any] | None = None, ctx: BotContext | None = None) -> None:
    pump_one = _pump_one(ctx, helpers)
    deadline = loop.time() + timeout_s
    while not any(ev.get("event_type") == event_type for ev in state.events[start_index:]):
        if loop.time() > deadline:
            player = find_player(state)
            raise TimeoutError(
                f"stalled waiting for event {event_type}; "
                f"level={state.current_level} tick={state.last_tick} "
                f"player_hp={(player or {}).get('hp')} "
                f"seen_events={sorted(state.seen_events)} "
                f"recent_combat={state.combat_events[-5:]}"
            )
        await pump_one(ws, state, timeout=0.1)


async def wait_for_matching_event(
    ws,
    state: RuntimeState,
    expected: dict[str, Any],
    loop,
    *,
    timeout_s: float = SLICE_TIMEOUT_S,
    start_index: int = 0,
    helpers: dict[str, Any] | None = None,
    ctx: BotContext | None = None,
) -> None:
    pump_one = _pump_one(ctx, helpers)
    event_type = str(expected.get("event_type", ""))
    deadline = loop.time() + timeout_s
    while not any(event_matches(ev, expected) for ev in state.events[start_index:]):
        if loop.time() > deadline:
            player = find_player(state)
            raise TimeoutError(
                f"stalled waiting for event {event_type or expected}; "
                f"level={state.current_level} tick={state.last_tick} "
                f"player_hp={(player or {}).get('hp')} "
                f"seen_events={sorted(state.seen_events)} "
                f"recent_combat={state.combat_events[-5:]}"
            )
        await pump_one(ws, state, timeout=0.1)


async def wait_for_shop_event(
    ws,
    state: RuntimeState,
    event_type: str,
    loop,
    *,
    shop_id: str = "town_vendor",
    start_index: int = 0,
    helpers: dict[str, Any] | None = None,
    ctx: BotContext | None = None,
) -> dict[str, Any]:
    pump_one = _pump_one(ctx, helpers)
    deadline = loop.time() + SLICE_TIMEOUT_S
    while True:
        for event in state.shop_events[start_index:]:
            if event.get("event_type") == event_type and str(event.get("shop_id", "")) == shop_id:
                return event
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for shop event {event_type} shop_id={shop_id}")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_stash_event(
    ws,
    state: RuntimeState,
    event_type: str,
    loop,
    *,
    stash_id: str = "account_stash",
    start_index: int = 0,
    helpers: dict[str, Any] | None = None,
    ctx: BotContext | None = None,
) -> dict[str, Any]:
    pump_one = _pump_one(ctx, helpers)
    deadline = loop.time() + SLICE_TIMEOUT_S
    while True:
        for event in state.stash_events[start_index:]:
            if event.get("event_type") == event_type and str(event.get("stash_id", "")) == stash_id:
                return event
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for stash event {event_type} stash_id={stash_id}")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_level_change(ws, state: RuntimeState, previous_level: int, loop, helpers: dict[str, Any] | None = None, ctx: BotContext | None = None) -> None:
    pump_one = _pump_one(ctx, helpers)
    deadline = loop.time() + SLICE_TIMEOUT_S
    while state.current_level == previous_level or state.pending_level_load is not None:
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for level change from {previous_level}")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_teleporter_discovery(ws, state: RuntimeState, level: int, loop, helpers: dict[str, Any] | None = None, ctx: BotContext | None = None) -> None:
    pump_one = _pump_one(ctx, helpers)
    deadline = loop.time() + SLICE_TIMEOUT_S
    while not state.discovered_teleporters.get(level, False):
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for teleporter discovery level {level}")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_character_progression(ws, state: RuntimeState, expected: dict[str, Any], loop, helpers: dict[str, Any] | None = None) -> None:
    helpers = _require_helpers(helpers)
    assert_character_progression = helpers["assert_character_progression"]
    pump_one = helpers["pump_one"]
    deadline = loop.time() + SLICE_TIMEOUT_S
    while True:
        try:
            assert_character_progression(state.character_progression, expected, "runtime protocol")
            return
        except AssertionError:
            pass
        if loop.time() > deadline:
            assert_character_progression(state.character_progression, expected, "runtime protocol")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_skill_progression(ws, state: RuntimeState, expected: dict[str, Any], loop, helpers: dict[str, Any] | None = None) -> None:
    helpers = _require_helpers(helpers)
    assert_skill_progression = helpers["assert_skill_progression"]
    pump_one = helpers["pump_one"]
    deadline = loop.time() + float(expected.get("timeout_s", SLICE_TIMEOUT_S))
    while True:
        try:
            assert_skill_progression(state.skill_progression, expected, "runtime protocol")
            return
        except AssertionError:
            pass
        if loop.time() > deadline:
            assert_skill_progression(state.skill_progression, expected, "runtime protocol")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_skill_cooldown(ws, state: RuntimeState, expected: dict[str, Any], loop, helpers: dict[str, Any] | None = None) -> None:
    helpers = _require_helpers(helpers)
    assert_skill_cooldown = helpers["assert_skill_cooldown"]
    pump_one = helpers["pump_one"]
    deadline = loop.time() + float(expected.get("timeout_s", SLICE_TIMEOUT_S))
    while True:
        try:
            assert_skill_cooldown(state.skill_cooldowns, expected, "runtime protocol")
            return
        except AssertionError:
            pass
        if loop.time() > deadline:
            assert_skill_cooldown(state.skill_cooldowns, expected, "runtime protocol")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_player_position(
    ws,
    state: RuntimeState,
    x: float,
    y: float,
    tolerance: float,
    loop,
    helpers: dict[str, Any] | None = None,
) -> None:
    helpers = _require_helpers(helpers)
    assert_player_position = helpers["assert_player_position"]
    pump_one = helpers["pump_one"]
    deadline = loop.time() + SLICE_TIMEOUT_S
    while True:
        try:
            assert_player_position(state, x, y, tolerance, "runtime protocol")
            return
        except AssertionError:
            pass
        if loop.time() > deadline:
            assert_player_position(state, x, y, tolerance, "runtime protocol")
        await pump_one(ws, state, timeout=0.1)


async def pump_one(ws, state: RuntimeState, timeout: float, helpers: dict[str, Any] | None = None) -> None:
    helpers = _require_helpers(helpers)
    ingest_message = helpers["ingest_message"]
    try:
        msg = await asyncio.wait_for(ws.recv(), timeout=timeout)
    except asyncio.TimeoutError:
        return
    ingest_message(json.loads(msg), state)
