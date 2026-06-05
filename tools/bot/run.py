#!/usr/bin/env python3
"""Headless Python protocol bot (ADR-0001 D8.5 layer 1).

Plays the full vertical slice through the same auth + WebSocket path as the
real client and asserts authoritative outcomes:

    dev-login -> create session -> move -> attack until dead -> pick up loot
    -> equip -> assert via /state -> reconnect and assert persisted inventory.

Progress goes to stderr; with --print-session-id the recorded session id is
written to stdout (and nothing else) so it can be captured for replay.

Usage:
    python -m tools.bot.run --base-url http://localhost:8080 \
        --dev-token local-dev-token --debug-token local-debug-token \
        [--print-session-id]
"""
from __future__ import annotations

import argparse
import asyncio
import json
import sys
from typing import Any

import httpx
import websockets

from tools.bot.protocol import make_envelope, to_ws_url

MONSTER_ID = "1002"
SLICE_TIMEOUT_S = 20.0


def log(*args: Any) -> None:
    print("[bot]", *args, file=sys.stderr, flush=True)


# --- HTTP steps -------------------------------------------------------------

def dev_login(client: httpx.Client, email: str, dev_token: str) -> tuple[str, str]:
    resp = client.post("/v0/auth/dev-login", json={"email": email, "dev_token": dev_token})
    resp.raise_for_status()
    body = resp.json()
    log("logged in as", body["account_id"])
    return body["account_id"], body["access_token"]


def create_session(client: httpx.Client, token: str) -> dict[str, Any]:
    resp = client.post("/v0/sessions", headers=auth(token), json={"mode": "solo"})
    resp.raise_for_status()
    body = resp.json()
    log("session", body["session_id"], "seed", body["seed"])
    return body


def fetch_state(client: httpx.Client, token: str, debug_token: str, session_id: str) -> dict[str, Any]:
    resp = client.get(
        f"/v0/sessions/{session_id}/state",
        headers={**auth(token), "X-Debug-Token": debug_token},
    )
    resp.raise_for_status()
    return resp.json()


def auth(token: str) -> dict[str, str]:
    return {"Authorization": f"Bearer {token}"}


# --- WebSocket slice --------------------------------------------------------

async def recv_json(ws) -> dict[str, Any]:
    return json.loads(await ws.recv())


async def drive_slice(base_url: str, token: str, sess: dict[str, Any]) -> str:
    """Play the slice over WebSocket; return the equipped item's instance id."""
    sid = sess["session_id"]
    uri = to_ws_url(base_url, sess["ws_url"])
    async with websockets.connect(uri, additional_headers=auth(token)) as ws:
        loop = asyncio.get_event_loop()

        first = await recv_json(ws)
        assert first["type"] == "session_snapshot", first["type"]
        last_tick = first["payload"]["server_tick"]
        log("connected; initial snapshot tick", last_tick)

        await ws.send(json.dumps(make_envelope(
            "client_ready", sid, last_tick, {"client_version": "bot", "last_seen_tick": last_tick})))

        # Move toward the monster (exercises movement + reconciliation surface).
        await ws.send(json.dumps(make_envelope(
            "move_intent", sid, last_tick, {"direction": {"x": 1, "y": 0}, "duration_ticks": 3})))

        async def attack() -> None:
            await ws.send(json.dumps(make_envelope("attack_intent", sid, last_tick, {"target_id": MONSTER_ID})))

        killed = picked_up = equip_sent = equipped = False
        loot_id: str | None = None
        item_id: str | None = None

        await attack()
        last_attack = loop.time()
        deadline = loop.time() + SLICE_TIMEOUT_S

        while not equipped:
            if loop.time() > deadline:
                raise TimeoutError(f"slice stalled: killed={killed} picked_up={picked_up} equipped={equipped}")
            try:
                msg = await asyncio.wait_for(ws.recv(), timeout=0.1)
            except asyncio.TimeoutError:
                if not killed and loop.time() - last_attack > 0.12:
                    await attack()
                    last_attack = loop.time()
                continue

            m = json.loads(msg)
            last_tick = max(last_tick, int(m.get("tick", 0)))
            if m["type"] != "state_delta":
                continue

            p = m["payload"]
            for ev in (p.get("events") or []):
                if ev["event_type"] == "monster_killed":
                    killed = True
                    log("monster killed at tick", p.get("server_tick"))
            for c in (p.get("changes") or []):
                if c["op"] == "entity_spawn" and c["entity"]["type"] == "loot":
                    loot_id = c["entity"]["id"]
                elif c["op"] == "inventory_add":
                    item_id = c["item"]["item_instance_id"]
                elif c["op"] == "equipped_update" and c.get("slot") == "weapon" and c.get("item_instance_id") == item_id:
                    equipped = True

            if killed and not picked_up and loot_id:
                await ws.send(json.dumps(make_envelope("pick_up_intent", sid, last_tick, {"entity_id": loot_id})))
                picked_up = True
                log("picking up loot", loot_id)
            if picked_up and item_id and not equip_sent:
                await ws.send(json.dumps(make_envelope(
                    "equip_intent", sid, last_tick, {"item_instance_id": item_id, "slot": "weapon"})))
                equip_sent = True
                log("equipping item", item_id)

        log("equipped item", item_id, "- slice complete over protocol")
        assert item_id is not None
        return item_id


async def check_persistence(base_url: str, token: str, session_id: str, item_id: str) -> None:
    """Reconnect with a fresh sim; the snapshot must reload inventory from DB."""
    uri = to_ws_url(base_url, "/v0/ws?session_id=" + session_id)
    async with websockets.connect(uri, additional_headers=auth(token)) as ws:
        snap = await recv_json(ws)
        assert snap["type"] == "session_snapshot", snap["type"]
        inv = snap["payload"]["inventory"]
        equipped = snap["payload"]["equipped"]
        assert_equipped_sword(inv, equipped, item_id, "reconnect snapshot")
        log("persisted inventory survived reconnect (loaded from Postgres)")


# --- assertions -------------------------------------------------------------

def assert_equipped_sword(inventory: list[dict], equipped: dict, item_id: str, where: str) -> None:
    if len(inventory) != 1:
        raise AssertionError(f"{where}: inventory size {len(inventory)} != 1: {inventory}")
    item = inventory[0]
    if item["item_def_id"] != "rusty_sword" or not item["equipped"]:
        raise AssertionError(f"{where}: item not an equipped rusty_sword: {item}")
    if equipped.get("weapon") != item_id:
        raise AssertionError(f"{where}: equipped weapon {equipped.get('weapon')} != {item_id}")


def assert_player_damaged(entities: list[dict], where: str) -> None:
    players = [e for e in entities if e.get("type") == "player"]
    if len(players) != 1:
        raise AssertionError(f"{where}: expected one player entity, got {players}")
    hp = players[0].get("hp")
    if not isinstance(hp, int) or hp >= 10:
        raise AssertionError(f"{where}: player hp {hp} did not show retaliation damage")


# --- main -------------------------------------------------------------------

def main() -> int:
    parser = argparse.ArgumentParser(description="arpg headless protocol bot")
    parser.add_argument("--base-url", default="http://localhost:8080")
    parser.add_argument("--dev-token", default="local-dev-token")
    parser.add_argument("--debug-token", default="local-debug-token")
    parser.add_argument("--email", default="bot@example.test")
    parser.add_argument("--print-session-id", action="store_true")
    args = parser.parse_args()

    with httpx.Client(base_url=args.base_url, timeout=10.0) as client:
        _, token = dev_login(client, args.email, args.dev_token)
        sess = create_session(client, token)
        session_id = sess["session_id"]

        item_id = asyncio.run(drive_slice(args.base_url, token, sess))

        # Assert authoritative state through the inspection API.
        state = fetch_state(client, token, args.debug_token, session_id)
        assert_equipped_sword(state["inventory"], state["equipped"], item_id, "/state API")
        assert_player_damaged(state["entities"], "/state API")
        log("/state API confirms equipped inventory and player damage")

        # Assert persistence by reconnecting a fresh session loop.
        asyncio.run(check_persistence(args.base_url, token, session_id, item_id))

    log("BOT OK")
    if args.print_session_id:
        print(session_id)
    return 0


if __name__ == "__main__":
    sys.exit(main())
