#!/usr/bin/env python3
"""Delete all characters for a dev-login account.

This is intentionally account-scoped and uses the public character API so bot
wrappers can clean their own durable heroes without reaching into Postgres.
"""

from __future__ import annotations

import argparse
from typing import Any

import httpx


def auth(token: str) -> dict[str, str]:
    return {"Authorization": f"Bearer {token}"}


def cleanup_account(base_url: str, dev_token: str, email: str) -> int:
    with httpx.Client(base_url=base_url, timeout=10.0) as client:
        login = client.post("/v0/auth/dev-login", json={"email": email, "dev_token": dev_token})
        login.raise_for_status()
        token = str(login.json()["access_token"])
        listed = client.get("/v0/characters", headers=auth(token))
        listed.raise_for_status()
        characters: list[dict[str, Any]] = list(listed.json().get("characters", []))
        removed = 0
        for character in characters:
            character_id = str(character.get("character_id", ""))
            if not character_id:
                continue
            deleted = client.delete(f"/v0/characters/{character_id}", headers=auth(token))
            if deleted.status_code == 204:
                removed += 1
                continue
            deleted.raise_for_status()
        return removed


def main() -> int:
    parser = argparse.ArgumentParser(description="delete all characters for a dev-login account")
    parser.add_argument("--base-url", default="http://localhost:8888")
    parser.add_argument("--dev-token", default="local-dev-token")
    parser.add_argument("--email", required=True)
    args = parser.parse_args()

    removed = cleanup_account(args.base_url, args.dev_token, args.email)
    print(f"deleted {removed} character(s) for {args.email}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
