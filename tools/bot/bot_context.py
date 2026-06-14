"""Typed runtime context for extracted protocol-bot modules.

This is the decoupling primitive that replaces ``helpers=globals()`` laundering
(v145-v149). A module that needs a *stateful* runtime service (e.g. the
WebSocket pump) declares a ``BotContext`` parameter instead of reaching back
into run.py's module namespace. The module then:

  * imports only leaf modules (bot_types, protocol, runtime_queries), never
    ``tools.bot.run``; and
  * can be unit-tested by constructing a ``BotContext`` with stub callables.

Pure helpers (find_player, dict_distance) do NOT belong here — they are
imported directly from ``runtime_queries``. Only inject services that carry
real runtime/connection state.
"""
from __future__ import annotations

from dataclasses import dataclass
from typing import Awaitable, Callable


@dataclass(frozen=True)
class BotContext:
    """Narrow set of runtime services injected into bot runtime modules.

    pump_one: ``async (ws, state, timeout) -> None`` — pump exactly one inbound
    WebSocket message into the runtime state.
    """

    pump_one: Callable[..., Awaitable[None]]
