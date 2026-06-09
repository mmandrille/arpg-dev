"""Pure helpers for the protocol bot: envelope construction and URL handling.

Kept side-effect free so they can be unit-tested without a running server.
"""
from __future__ import annotations

import itertools
from typing import Any

_message_counter = itertools.count(1)


def next_message_id() -> str:
    """Return a unique, non-empty message id for an outbound envelope."""
    return f"msg-{next(_message_counter)}"


def make_envelope(
    msg_type: str,
    session_id: str,
    tick: int,
    payload: dict[str, Any],
    message_id: str | None = None,
    correlation_id: str | None = None,
) -> dict[str, Any]:
    """Build a v0 WebSocket envelope."""
    env: dict[str, Any] = {
        "type": msg_type,
        "message_id": message_id or next_message_id(),
        "session_id": session_id,
        "tick": tick,
        "payload": payload,
    }
    if correlation_id:
        env["correlation_id"] = correlation_id
    return env


def to_ws_url(base_url: str, ws_path: str) -> str:
    """Combine an http(s) base URL and a ws path into a ws(s) URL.

    >>> to_ws_url("http://localhost:8888", "/v0/ws?session_id=s")
    'ws://localhost:8888/v0/ws?session_id=s'
    """
    if base_url.startswith("https"):
        scheme, rest = "wss", base_url[len("https"):]
    elif base_url.startswith("http"):
        scheme, rest = "ws", base_url[len("http"):]
    else:
        raise ValueError(f"unsupported base url scheme: {base_url}")
    return scheme + rest.rstrip("/") + ws_path
