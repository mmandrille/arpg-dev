from __future__ import annotations

import json
from typing import Any, Callable

from tools.bot.bot_context import BotContext
from tools.bot.bot_types import RuntimeState
from tools.bot.runtime_queries import find_player


async def wait_for_event_with_action_retries(
    ws,
    state: RuntimeState,
    session_id: str,
    *,
    target_id: str,
    event_type: str,
    timeout_s: float,
    loop,
    make_envelope: Callable[..., dict[str, Any]],
    wait_for_accept,
    pump_one,
    ctx: BotContext | None = None,
) -> None:
    deadline = loop.time() + timeout_s
    start_index = len(state.events)
    last_resend = loop.time()
    while loop.time() < deadline:
        if any(ev.get("event_type") == event_type for ev in state.events[start_index:]):
            return
        if loop.time() - last_resend >= 3.0:
            env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": target_id})
            await ws.send(json.dumps(env))
            await wait_for_accept(ws, state, env["message_id"], loop)
            last_resend = loop.time()
        if ctx is not None:
            await ctx.pump_one(ws, state, timeout=0.1)
        else:
            await pump_one(ws, state, timeout=0.1)

    player = find_player(state)
    raise TimeoutError(
        f"stalled waiting for event {event_type}; "
        f"level={state.current_level} tick={state.last_tick} "
        f"player_hp={(player or {}).get('hp')} "
        f"seen_events={sorted(state.seen_events)} "
        f"recent_combat={state.combat_events[-5:]}"
    )
