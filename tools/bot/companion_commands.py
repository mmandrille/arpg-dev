"""Protocol bot helpers for companion command intents."""

from __future__ import annotations

import json
from typing import Any


async def set_companion_stance(ws: Any, session_id: str, state: Any, step: dict[str, Any], loop: Any, helpers: dict[str, Any]) -> None:
    env = helpers["make_envelope"](
        "companion_command_intent",
        session_id,
        state.last_tick,
        {"stance": str(step["stance"])},
    )
    await ws.send(json.dumps(env))
    expect_reject = step.get("expect_reject")
    if expect_reject:
        await helpers["wait_for_reject"](ws, state, env["message_id"], str(expect_reject), loop)
        return
    await helpers["wait_for_accept"](ws, state, env["message_id"], loop)
