#!/usr/bin/env python3
"""Headless Python protocol bot (ADR-0001 D8.5 layer 1).

Plays discovered bot scenarios through the same auth + WebSocket path as the
real client and asserts authoritative outcomes:

    dev-login -> create session -> move -> attack until dead -> pick up loot
    -> equip -> assert via /state -> reconnect and assert reconstructed state.

Progress goes to stderr; with --print-session-id the recorded session id is
written to stdout (and nothing else) so it can be captured for replay.

Usage:
    python -m tools.bot.run --base-url http://localhost:8080 \
        --dev-token local-dev-token --debug-token local-debug-token \
        [--scenario all] [--write-manifest path] [--print-session-id]
"""
from __future__ import annotations

import argparse
import asyncio
from dataclasses import dataclass, field
from datetime import datetime, timezone
import json
from pathlib import Path
import sys
from typing import Any

import httpx
import websockets

from tools.bot.protocol import make_envelope, to_ws_url

MONSTER_ID = "1002"
SLICE_TIMEOUT_S = 20.0
ROOT = Path(__file__).resolve().parent.parent.parent
SCENARIO_DIR = Path(__file__).resolve().parent / "scenarios"


def log(*args: Any) -> None:
    print("[bot]", *args, file=sys.stderr, flush=True)


@dataclass(frozen=True)
class Scenario:
    id: str
    title: str
    description: str
    steps: list[dict[str, Any]]
    assertions: list[str]
    path: Path


@dataclass
class RuntimeState:
    last_tick: int = 0
    killed: bool = False
    loot_ids: list[str] = field(default_factory=list)
    item_id: str | None = None
    equipped_item_id: str | None = None
    seen_events: set[str] = field(default_factory=set)


def load_scenarios(scenario_dir: Path = SCENARIO_DIR) -> list[Scenario]:
    scenarios: list[Scenario] = []
    for path in sorted(scenario_dir.glob("*.json")):
        raw = json.loads(path.read_text())
        sid = raw.get("id", "")
        if not sid:
            raise ValueError(f"{path}: missing scenario id")
        scenarios.append(Scenario(
            id=sid,
            title=raw.get("title", sid),
            description=raw.get("description", ""),
            steps=list(raw.get("steps", [])),
            assertions=list(raw.get("assertions", [])),
            path=path,
        ))
    if not scenarios:
        raise ValueError(f"no scenarios found in {scenario_dir}")
    return scenarios


def select_scenarios(scenarios: list[Scenario], selected: str) -> list[Scenario]:
    if selected == "all":
        return scenarios
    wanted = {part.strip() for part in selected.split(",") if part.strip()}
    found = [s for s in scenarios if s.id in wanted]
    missing = wanted - {s.id for s in found}
    if missing:
        raise ValueError(f"unknown scenario(s): {', '.join(sorted(missing))}")
    return found


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


def fetch_replay(client: httpx.Client, token: str, debug_token: str, session_id: str) -> dict[str, Any]:
    resp = client.get(
        f"/v0/sessions/{session_id}/replay",
        headers={**auth(token), "X-Debug-Token": debug_token},
    )
    resp.raise_for_status()
    return resp.json()


def auth(token: str) -> dict[str, str]:
    return {"Authorization": f"Bearer {token}"}


# --- WebSocket slice --------------------------------------------------------

async def recv_json(ws) -> dict[str, Any]:
    return json.loads(await ws.recv())


async def drive_scenario(base_url: str, token: str, sess: dict[str, Any], scenario: Scenario) -> RuntimeState:
    """Play one scenario over WebSocket and return observed runtime state."""
    sid = sess["session_id"]
    uri = to_ws_url(base_url, sess["ws_url"])
    async with websockets.connect(uri, additional_headers=auth(token)) as ws:
        loop = asyncio.get_event_loop()

        first = await recv_json(ws)
        assert first["type"] == "session_snapshot", first["type"]
        state = RuntimeState(last_tick=first["payload"]["server_tick"])
        log("connected; initial snapshot tick", state.last_tick)

        await ws.send(json.dumps(make_envelope(
            "client_ready", sid, state.last_tick, {"client_version": "bot", "last_seen_tick": state.last_tick})))

        for step in scenario.steps:
            await execute_step(ws, sid, state, step, loop)

        if state.equipped_item_id:
            log("equipped item", state.equipped_item_id, "- scenario complete over protocol")
        else:
            log("scenario complete over protocol")
        return state


async def execute_step(ws, session_id: str, state: RuntimeState, step: dict[str, Any], loop) -> None:
    action = step.get("action")
    if action == "move":
        direction = step["direction"]
        duration = int(step.get("duration_ticks", 1))
        await ws.send(json.dumps(make_envelope(
            "move_intent",
            session_id,
            state.last_tick,
            {"direction": direction, "duration_ticks": duration},
        )))
        await wait_for_tick(ws, state, state.last_tick + 1, loop)
        return

    if action == "attack_until_event":
        target_id = str(step.get("target_id", MONSTER_ID))
        event_type = str(step["event_type"])
        last_attack = 0.0
        deadline = loop.time() + SLICE_TIMEOUT_S
        while event_type not in state.seen_events:
            if loop.time() > deadline:
                raise TimeoutError(f"attack_until_event stalled waiting for {event_type}")
            if loop.time() - last_attack > 0.12:
                await ws.send(json.dumps(make_envelope(
                    "attack_intent", session_id, state.last_tick, {"target_id": target_id})))
                last_attack = loop.time()
            await pump_one(ws, state, timeout=0.1)
        return

    if action == "pick_up_first_loot":
        deadline = loop.time() + SLICE_TIMEOUT_S
        while not state.loot_ids:
            if loop.time() > deadline:
                raise TimeoutError("pick_up_first_loot stalled waiting for loot")
            await pump_one(ws, state, timeout=0.1)
        loot_id = state.loot_ids[0]
        await ws.send(json.dumps(make_envelope(
            "pick_up_intent", session_id, state.last_tick, {"entity_id": loot_id})))
        log("picking up loot", loot_id)
        while state.item_id is None:
            if loop.time() > deadline:
                raise TimeoutError("pick_up_first_loot stalled waiting for inventory_add")
            await pump_one(ws, state, timeout=0.1)
        return

    if action == "equip_first_inventory_item":
        deadline = loop.time() + SLICE_TIMEOUT_S
        while state.item_id is None:
            if loop.time() > deadline:
                raise TimeoutError("equip_first_inventory_item stalled waiting for inventory")
            await pump_one(ws, state, timeout=0.1)
        slot = str(step.get("slot", "weapon"))
        await ws.send(json.dumps(make_envelope(
            "equip_intent", session_id, state.last_tick, {"item_instance_id": state.item_id, "slot": slot})))
        log("equipping item", state.item_id)
        while state.equipped_item_id != state.item_id:
            if loop.time() > deadline:
                raise TimeoutError("equip_first_inventory_item stalled waiting for equipped_update")
            await pump_one(ws, state, timeout=0.1)
        return

    raise ValueError(f"unsupported scenario action: {action}")


async def wait_for_tick(ws, state: RuntimeState, target_tick: int, loop) -> None:
    deadline = loop.time() + SLICE_TIMEOUT_S
    while state.last_tick < target_tick:
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for tick {target_tick}")
        await pump_one(ws, state, timeout=0.1)


async def pump_one(ws, state: RuntimeState, timeout: float) -> None:
    try:
        msg = await asyncio.wait_for(ws.recv(), timeout=timeout)
    except asyncio.TimeoutError:
        return
    ingest_message(json.loads(msg), state)


def ingest_message(m: dict[str, Any], state: RuntimeState) -> None:
    state.last_tick = max(state.last_tick, int(m.get("tick", 0)))
    if m.get("type") != "state_delta":
        return

    p = m["payload"]
    for ev in (p.get("events") or []):
        event_type = ev["event_type"]
        state.seen_events.add(event_type)
        if event_type == "monster_killed":
            state.killed = True
            log("monster killed at tick", p.get("server_tick"))
    for c in (p.get("changes") or []):
        if c["op"] == "entity_spawn" and c["entity"]["type"] == "loot":
            loot_id = c["entity"]["id"]
            if loot_id not in state.loot_ids:
                state.loot_ids.append(loot_id)
        elif c["op"] == "inventory_add":
            state.item_id = c["item"]["item_instance_id"]
        elif c["op"] == "equipped_update" and c.get("slot") == "weapon":
            state.equipped_item_id = c.get("item_instance_id")


async def check_persistence(base_url: str, token: str, session_id: str, item_id: str, assertions: list[str]) -> None:
    """Reconnect and assert the snapshot was reconstructed from recorded inputs."""
    uri = to_ws_url(base_url, "/v0/ws?session_id=" + session_id)
    async with websockets.connect(uri, additional_headers=auth(token)) as ws:
        snap = await recv_json(ws)
        assert snap["type"] == "session_snapshot", snap["type"]
        payload = snap["payload"]
        inv = payload["inventory"]
        equipped = payload["equipped"]
        run_assertions(assertions, payload["entities"], inv, equipped, item_id, "reconnect snapshot")
        log("reconnect snapshot restored expected scenario state")


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


def assert_monster_dead(entities: list[dict], where: str) -> None:
    monsters = [e for e in entities if e.get("id") == MONSTER_ID and e.get("type") == "monster"]
    if len(monsters) != 1:
        raise AssertionError(f"{where}: expected training dummy entity, got {monsters}")
    hp = monsters[0].get("hp")
    if hp != 0:
        raise AssertionError(f"{where}: training dummy hp {hp} != 0")


def run_assertions(
    assertions: list[str],
    entities: list[dict],
    inventory: list[dict],
    equipped: dict,
    item_id: str | None,
    where: str,
) -> None:
    for assertion in assertions:
        if assertion == "equipped_rusty_sword":
            if item_id is None:
                raise AssertionError(f"{where}: scenario did not observe an item to equip")
            assert_equipped_sword(inventory, equipped, item_id, where)
        elif assertion == "player_damaged":
            assert_player_damaged(entities, where)
        elif assertion == "monster_dead":
            assert_monster_dead(entities, where)
        else:
            raise AssertionError(f"{where}: unknown assertion {assertion}")


def default_manifest_path() -> Path:
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    return ROOT / ".artifacts" / "bot-runs" / f"{stamp}.json"


def write_manifest(path: Path, base_url: str, results: list[dict[str, Any]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    body = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "base_url": base_url,
        "scenarios": results,
    }
    path.write_text(json.dumps(body, indent=2) + "\n")


# --- main -------------------------------------------------------------------

def main() -> int:
    parser = argparse.ArgumentParser(description="arpg headless protocol bot")
    parser.add_argument("--base-url", default="http://localhost:8080")
    parser.add_argument("--dev-token", default="local-dev-token")
    parser.add_argument("--debug-token", default="local-debug-token")
    parser.add_argument("--email", default="bot@example.test")
    parser.add_argument("--scenario", default="all", help="scenario id, comma-separated ids, or all")
    parser.add_argument("--list-scenarios", action="store_true")
    parser.add_argument("--write-manifest", type=Path)
    parser.add_argument("--print-session-id", action="store_true")
    args = parser.parse_args()

    scenarios = load_scenarios()
    if args.list_scenarios:
        for scenario in scenarios:
            print(f"{scenario.id}\t{scenario.title}")
        return 0
    selected = select_scenarios(scenarios, args.scenario)
    results: list[dict[str, Any]] = []
    last_session_id = ""

    with httpx.Client(base_url=args.base_url, timeout=10.0) as client:
        _, token = dev_login(client, args.email, args.dev_token)
        for scenario in selected:
            log("scenario", scenario.id, "-", scenario.title)
            sess = create_session(client, token)
            session_id = sess["session_id"]
            last_session_id = session_id

            observed = asyncio.run(drive_scenario(args.base_url, token, sess, scenario))

            # Assert authoritative state through the inspection API.
            state = fetch_state(client, token, args.debug_token, session_id)
            run_assertions(
                scenario.assertions,
                state["entities"],
                state["inventory"],
                state["equipped"],
                observed.item_id,
                "/state API",
            )
            log("/state API confirms expected scenario state")

            # Assert replay reconstruction by reconnecting a fresh session loop.
            if observed.item_id is not None:
                asyncio.run(check_persistence(args.base_url, token, session_id, observed.item_id, scenario.assertions))

            replay = fetch_replay(client, token, args.debug_token, session_id)
            if not replay.get("match", False):
                raise AssertionError(f"replay mismatch for {session_id}: {replay.get('mismatch')}")
            log("replay verified for", session_id)

            results.append({
                "id": scenario.id,
                "title": scenario.title,
                "description": scenario.description,
                "session_id": session_id,
                "seed": sess["seed"],
                "status": "passed",
                "replay_match": True,
            })

    if args.write_manifest:
        write_manifest(args.write_manifest, args.base_url, results)
        log("wrote manifest", args.write_manifest)

    log("BOT OK", "- scenarios:", ", ".join(r["id"] for r in results))
    if args.print_session_id:
        print(last_session_id)
    return 0


if __name__ == "__main__":
    sys.exit(main())
