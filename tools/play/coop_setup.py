#!/usr/bin/env python3
"""Create a local co-op session for explicit development/debug use.

`make play N` now launches independent menu clients. This helper is kept for
manual protocol/debug setup when a pre-created private join-code session is
useful.
"""

from __future__ import annotations

import argparse
import json
import time
import urllib.error
import urllib.request
from typing import Any


def request(base_url: str, method: str, path: str, body: dict[str, Any] | None = None, token: str = "") -> dict[str, Any]:
    data = json.dumps(body or {}).encode("utf-8") if body is not None else None
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    req = urllib.request.Request(f"{base_url.rstrip('/')}{path}", data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            raw = resp.read().decode("utf-8")
    except urllib.error.HTTPError as exc:
        raw = exc.read().decode("utf-8")
        raise RuntimeError(f"{method} {path} failed HTTP {exc.code}: {raw}") from exc
    return json.loads(raw) if raw else {}


def dev_login(base_url: str, email: str, dev_token: str) -> str:
    body = request(base_url, "POST", "/v0/auth/dev-login", {"email": email, "dev_token": dev_token})
    return str(body["access_token"])


def ensure_character(base_url: str, token: str, name: str) -> str:
    chars = request(base_url, "GET", "/v0/characters", token=token).get("characters", [])
    if chars:
        return str(chars[0]["character_id"])
    created = request(base_url, "POST", "/v0/characters", {"name": name}, token)
    return str(created["character_id"])


def main() -> int:
    parser = argparse.ArgumentParser(description="prepare a local co-op play session")
    parser.add_argument("--base-url", default="http://localhost:8888")
    parser.add_argument("--dev-token", default="local-dev-token")
    parser.add_argument("--clients", type=int, required=True)
    parser.add_argument("--email-prefix", default="player")
    parser.add_argument("--world-id", default="dungeon_levels")
    args = parser.parse_args()

    if args.clients < 2:
        raise SystemExit("--clients must be at least 2")

    stamp = int(time.time() * 1000)
    clients: list[dict[str, str]] = []
    for idx in range(1, args.clients + 1):
        email = f"{args.email_prefix}{idx}+play-{stamp}@example.test"
        token = dev_login(args.base_url, email, args.dev_token)
        character_id = ensure_character(args.base_url, token, f"Player {idx}")
        clients.append({"email": email, "token": token, "character_id": character_id})

    host = clients[0]
    session = request(
        args.base_url,
        "POST",
        "/v0/sessions",
        {"mode": "coop", "world_id": args.world_id, "character_id": host["character_id"]},
        host["token"],
    )
    session_id = str(session["session_id"])
    join_code = str(session["join_code"])

    for guest in clients[1:]:
        request(
            args.base_url,
            "POST",
            f"/v0/sessions/{session_id}/join",
            {"join_code": join_code, "character_id": guest["character_id"]},
            guest["token"],
        )

    print(json.dumps({"session_id": session_id, "world_id": args.world_id, "clients": clients}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
