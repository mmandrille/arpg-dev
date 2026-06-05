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

SLICE_TIMEOUT_S = 20.0
DEFAULT_WORLD_ID = "vertical_slice"
WALK_STOP_DISTANCE = 1.0
WALK_MAX_TICKS = 40
ROOT = Path(__file__).resolve().parent.parent.parent
SCENARIO_DIR = Path(__file__).resolve().parent / "scenarios"


def load_known_world_ids() -> set[str]:
    worlds_path = ROOT / "shared" / "rules" / "worlds.v0.json"
    data = json.loads(worlds_path.read_text(encoding="utf-8"))
    return set(data["worlds"])


KNOWN_WORLD_IDS = load_known_world_ids()


def log(*args: Any) -> None:
    print("[bot]", *args, file=sys.stderr, flush=True)


@dataclass(frozen=True)
class Scenario:
    id: str
    world_id: str
    title: str
    description: str
    steps: list[dict[str, Any]]
    assertions: list[Any]
    path: Path


@dataclass
class RuntimeState:
    last_tick: int = 0
    killed: bool = False
    entities: dict[str, dict[str, Any]] = field(default_factory=dict)
    inventory: list[dict[str, Any]] = field(default_factory=list)
    equipped: dict[str, Any] = field(default_factory=dict)
    loot_ids: list[str] = field(default_factory=list)
    item_id: str | None = None
    equipped_item_id: str | None = None
    seen_events: set[str] = field(default_factory=set)
    pending_attack_monsters: dict[str, str] = field(default_factory=dict)
    accepted_attack_counts: dict[str, int] = field(default_factory=dict)
    killed_monster_def_ids: set[str] = field(default_factory=set)
    accepted_message_ids: set[str] = field(default_factory=set)
    rejected_message_reasons: dict[str, str] = field(default_factory=dict)


def load_scenarios(scenario_dir: Path = SCENARIO_DIR) -> list[Scenario]:
    scenarios: list[Scenario] = []
    for path in sorted(scenario_dir.glob("*.json")):
        raw = json.loads(path.read_text())
        sid = raw.get("id", "")
        if not sid:
            raise ValueError(f"{path}: missing scenario id")
        world_id = raw.get("world_id", DEFAULT_WORLD_ID)
        if world_id not in KNOWN_WORLD_IDS:
            raise ValueError(f"{path}: unknown world_id {world_id}")
        scenarios.append(Scenario(
            id=sid,
            world_id=world_id,
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


def create_session(client: httpx.Client, token: str, world_id: str) -> dict[str, Any]:
    resp = client.post("/v0/sessions", headers=auth(token), json={"mode": "solo", "world_id": world_id})
    resp.raise_for_status()
    body = resp.json()
    log("session", body["session_id"], "seed", body["seed"], "world", body.get("world_id"))
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
        state = RuntimeState()
        ingest_snapshot(first["payload"], state)
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
        env = make_envelope(
            "move_intent",
            session_id,
            state.last_tick,
            {"direction": direction, "duration_ticks": duration},
        )
        await ws.send(json.dumps(env))
        await wait_for_accept(ws, state, env["message_id"], loop)
        return

    if action == "move_until_player_position":
        direction = step["direction"]
        duration = int(step.get("duration_ticks", 1))
        env = make_envelope(
            "move_intent",
            session_id,
            state.last_tick,
            {"direction": direction, "duration_ticks": duration},
        )
        await ws.send(json.dumps(env))
        await wait_for_accept(ws, state, env["message_id"], loop)
        await wait_for_player_position(
            ws,
            state,
            float(step["x"]),
            float(step["y"]),
            float(step.get("tolerance", 0.001)),
            loop,
        )
        return

    if action == "assert_player_position":
        assert_player_position(state, float(step["x"]), float(step["y"]), float(step.get("tolerance", 0.001)), "runtime protocol")
        return

    if action == "attack_until_event":
        target_id = str(step["target_id"]) if step.get("target_id") else None
        if target_id is None:
            monster = find_monster(state, str(step.get("monster_def_id", "training_dummy")))
            if monster is None:
                raise AssertionError(f"attack_until_event: monster not found: {step}")
            target_id = str(monster["id"])
        event_type = str(step["event_type"])
        last_attack = 0.0
        deadline = loop.time() + SLICE_TIMEOUT_S
        while event_type not in state.seen_events:
            if loop.time() > deadline:
                raise TimeoutError(f"attack_until_event stalled waiting for {event_type}")
            if loop.time() - last_attack > 0.12:
                monster = state.entities.get(target_id)
                if monster is not None and monster.get("hp") == 0:
                    break
                monster_def_id = str(step.get("monster_def_id") or (monster or {}).get("monster_def_id") or "")
                env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": target_id})
                if monster_def_id:
                    state.pending_attack_monsters[env["message_id"]] = monster_def_id
                await ws.send(json.dumps(env))
                last_attack = loop.time()
            await pump_one(ws, state, timeout=0.1)
        return

    if action == "action_until_event":
        target = resolve_target(state, step)
        event_type = str(step["event_type"])
        target_id = str(target["id"])
        last_action = 0.0
        deadline = loop.time() + SLICE_TIMEOUT_S
        while event_type not in state.seen_events:
            if loop.time() > deadline:
                raise TimeoutError(f"action_until_event stalled waiting for {event_type}")
            if loop.time() - last_action > 0.12:
                current = state.entities.get(target_id)
                if current is not None and current.get("type") == "monster" and current.get("hp") == 0:
                    break
                monster_def_id = str(step.get("monster_def_id") or (current or {}).get("monster_def_id") or "")
                env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": target_id})
                if monster_def_id:
                    state.pending_attack_monsters[env["message_id"]] = monster_def_id
                await ws.send(json.dumps(env))
                last_action = loop.time()
            await pump_one(ws, state, timeout=0.1)
        return

    if action == "action_entity":
        target = resolve_target(state, step)
        env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": str(target["id"])})
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        event_type = step.get("event_type")
        if event_type:
            await wait_for_event(ws, state, str(event_type), loop)
        return

    if action == "move_until_in_range":
        target = resolve_target(state, step)
        await walk_toward(ws, session_id, state, target["position"], loop, stop_distance=float(step.get("stop_distance", WALK_STOP_DISTANCE)))
        return

    if action == "walk_to_loot":
        loot = find_loot(state, str(step["item_def_id"]))
        if loot is None:
            raise AssertionError(f"walk_to_loot: loot not found: {step}")
        await walk_toward(ws, session_id, state, loot["position"], loop)
        return

    if action == "walk_to_monster":
        monster = find_monster(state, str(step["monster_def_id"]))
        if monster is None:
            raise AssertionError(f"walk_to_monster: monster not found: {step}")
        await walk_toward(ws, session_id, state, monster["position"], loop)
        return

    if action == "pick_up_first_loot":
        deadline = loop.time() + SLICE_TIMEOUT_S
        while not state.loot_ids:
            if loop.time() > deadline:
                raise TimeoutError("pick_up_first_loot stalled waiting for loot")
            await pump_one(ws, state, timeout=0.1)
        loot_id = state.loot_ids[0]
        await ws.send(json.dumps(make_envelope(
            "action_intent", session_id, state.last_tick, {"target_id": loot_id})))
        log("picking up loot", loot_id)
        while state.item_id is None:
            if loop.time() > deadline:
                raise TimeoutError("pick_up_first_loot stalled waiting for inventory_add")
            await pump_one(ws, state, timeout=0.1)
        return

    if action == "pick_up_loot":
        item_def_id = str(step["item_def_id"])
        loot = find_loot(state, item_def_id)
        if loot is None:
            raise AssertionError(f"pick_up_loot: loot not found for item_def_id={item_def_id}")
        deadline = loop.time() + SLICE_TIMEOUT_S
        await ws.send(json.dumps(make_envelope(
            "action_intent", session_id, state.last_tick, {"target_id": loot["id"]})))
        log("picking up", item_def_id, "loot", loot["id"])
        while find_inventory_item(state.inventory, item_def_id) is None:
            if loop.time() > deadline:
                raise TimeoutError(f"pick_up_loot stalled waiting for {item_def_id}")
            await pump_one(ws, state, timeout=0.1)
        item = find_inventory_item(state.inventory, item_def_id)
        if item is not None:
            state.item_id = item["item_instance_id"]
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

    if action == "equip_inventory_item":
        item_def_id = str(step["item_def_id"])
        deadline = loop.time() + SLICE_TIMEOUT_S
        item = find_inventory_item(state.inventory, item_def_id)
        while item is None:
            if loop.time() > deadline:
                raise TimeoutError(f"equip_inventory_item stalled waiting for {item_def_id}")
            await pump_one(ws, state, timeout=0.1)
            item = find_inventory_item(state.inventory, item_def_id)
        slot = str(step.get("slot", item.get("slot", "weapon")))
        item_id = str(item["item_instance_id"])
        await ws.send(json.dumps(make_envelope(
            "equip_intent", session_id, state.last_tick, {"item_instance_id": item_id, "slot": slot})))
        log("equipping", item_def_id, item_id)
        while state.equipped.get(slot) != item_id:
            if loop.time() > deadline:
                raise TimeoutError(f"equip_inventory_item stalled waiting for equipped_update for {item_def_id}")
            await pump_one(ws, state, timeout=0.1)
        state.equipped_item_id = item_id
        return

    raise ValueError(f"unsupported scenario action: {action}")


async def walk_toward(
    ws,
    session_id: str,
    state: RuntimeState,
    target_pos: dict[str, Any],
    loop,
    max_ticks: int = WALK_MAX_TICKS,
    stop_distance: float = WALK_STOP_DISTANCE,
) -> None:
    for _ in range(max_ticks):
        player = find_player(state)
        if player is None:
            raise AssertionError("walk_toward: player not found")
        player_pos = player["position"]
        dx = float(target_pos["x"]) - float(player_pos["x"])
        dy = float(target_pos["y"]) - float(player_pos["y"])
        if max(abs(dx), abs(dy)) <= stop_distance:
            return

        direction = {"x": 0, "y": 0}
        if abs(dx) > 0:
            direction["x"] = 1 if dx > 0 else -1
        elif abs(dy) > 0:
            direction["y"] = 1 if dy > 0 else -1

        before = {"x": player_pos["x"], "y": player_pos["y"]}
        await ws.send(json.dumps(make_envelope(
            "move_intent",
            session_id,
            state.last_tick,
            {"direction": direction, "duration_ticks": 1},
        )))
        await wait_for_player_move(ws, state, before, loop)

    raise TimeoutError(f"walk_toward exhausted {max_ticks} ticks toward {target_pos}")


async def wait_for_player_move(ws, state: RuntimeState, before: dict[str, Any], loop) -> None:
    deadline = loop.time() + SLICE_TIMEOUT_S
    while True:
        player = find_player(state)
        if player is not None:
            pos = player["position"]
            if pos.get("x") != before.get("x") or pos.get("y") != before.get("y"):
                return
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for player movement from {before}")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_tick(ws, state: RuntimeState, target_tick: int, loop) -> None:
    deadline = loop.time() + SLICE_TIMEOUT_S
    while state.last_tick < target_tick:
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for tick {target_tick}")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_accept(ws, state: RuntimeState, message_id: str, loop) -> None:
    deadline = loop.time() + SLICE_TIMEOUT_S
    while message_id not in state.accepted_message_ids:
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for accept {message_id}")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_reject(ws, state: RuntimeState, message_id: str, reason: str, loop) -> None:
    deadline = loop.time() + SLICE_TIMEOUT_S
    while message_id not in state.rejected_message_reasons:
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for reject {message_id}")
        await pump_one(ws, state, timeout=0.1)
    got = state.rejected_message_reasons[message_id]
    if got != reason:
        raise AssertionError(f"reject {message_id} reason={got}, want {reason}")


async def wait_for_event(ws, state: RuntimeState, event_type: str, loop) -> None:
    deadline = loop.time() + SLICE_TIMEOUT_S
    while event_type not in state.seen_events:
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for event {event_type}")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_player_position(
    ws,
    state: RuntimeState,
    x: float,
    y: float,
    tolerance: float,
    loop,
) -> None:
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


async def pump_one(ws, state: RuntimeState, timeout: float) -> None:
    try:
        msg = await asyncio.wait_for(ws.recv(), timeout=timeout)
    except asyncio.TimeoutError:
        return
    ingest_message(json.loads(msg), state)


def ingest_message(m: dict[str, Any], state: RuntimeState) -> None:
    state.last_tick = max(state.last_tick, int(m.get("tick", 0)))
    if m.get("type") == "session_snapshot":
        ingest_snapshot(m["payload"], state)
        return
    if m.get("type") == "intent_accepted":
        accepted_id = str(m.get("payload", {}).get("accepted_message_id", ""))
        state.accepted_message_ids.add(accepted_id)
        monster_def_id = state.pending_attack_monsters.pop(accepted_id, None)
        if monster_def_id:
            state.accepted_attack_counts[monster_def_id] = state.accepted_attack_counts.get(monster_def_id, 0) + 1
        return
    if m.get("type") == "intent_rejected":
        rejected_id = str(m.get("payload", {}).get("rejected_message_id", ""))
        state.rejected_message_reasons[rejected_id] = str(m.get("payload", {}).get("reason", "unknown"))
        monster_def_id = state.pending_attack_monsters.pop(rejected_id, None)
        if monster_def_id:
            reason = m.get("payload", {}).get("reason", "unknown")
            raise AssertionError(f"action_intent for {monster_def_id} was rejected: {reason}")
        return
    if m.get("type") != "state_delta":
        return

    p = m["payload"]
    state.last_tick = max(state.last_tick, int(p.get("server_tick", state.last_tick)))
    for ev in (p.get("events") or []):
        event_type = ev["event_type"]
        state.seen_events.add(event_type)
        if event_type == "monster_killed":
            state.killed = True
            entity = state.entities.get(str(ev.get("entity_id")))
            if entity is not None and entity.get("monster_def_id"):
                state.killed_monster_def_ids.add(str(entity["monster_def_id"]))
            log("monster killed at tick", p.get("server_tick"))
    for c in (p.get("changes") or []):
        if c["op"] in {"entity_spawn", "entity_update"}:
            entity = c["entity"]
            existing = state.entities.get(entity["id"], {})
            existing.update(entity)
            state.entities[entity["id"]] = existing
            if c["op"] == "entity_spawn" and entity["type"] == "loot":
                loot_id = entity["id"]
                if loot_id not in state.loot_ids:
                    state.loot_ids.append(loot_id)
        elif c["op"] == "entity_remove":
            entity_id = c["entity_id"]
            state.entities.pop(entity_id, None)
            if entity_id in state.loot_ids:
                state.loot_ids.remove(entity_id)
        elif c["op"] == "inventory_add":
            upsert_inventory(state, c["item"])
            state.item_id = c["item"]["item_instance_id"]
        elif c["op"] == "inventory_update":
            upsert_inventory(state, c["item"])
        elif c["op"] == "equipped_update" and c.get("slot") == "weapon":
            state.equipped_item_id = c.get("item_instance_id")
            state.equipped[c["slot"]] = c.get("item_instance_id")


def ingest_snapshot(payload: dict[str, Any], state: RuntimeState) -> None:
    state.last_tick = max(state.last_tick, int(payload.get("server_tick", 0)))
    state.entities = {str(e["id"]): dict(e) for e in payload.get("entities", [])}
    state.inventory = [dict(i) for i in payload.get("inventory", [])]
    state.equipped = dict(payload.get("equipped", {}))
    state.loot_ids = [
        entity_id
        for entity_id, entity in state.entities.items()
        if entity.get("type") == "loot"
    ]
    for item in state.inventory:
        if item.get("equipped"):
            state.equipped_item_id = item.get("item_instance_id")


def upsert_inventory(state: RuntimeState, item: dict[str, Any]) -> None:
    item_id = item["item_instance_id"]
    for i, current in enumerate(state.inventory):
        if current.get("item_instance_id") == item_id:
            merged = dict(current)
            merged.update(item)
            state.inventory[i] = merged
            return
    state.inventory.append(dict(item))


def find_loot(state: RuntimeState, item_def_id: str) -> dict[str, Any] | None:
    for entity in state.entities.values():
        if entity.get("type") == "loot" and entity.get("item_def_id") == item_def_id:
            return entity
    return None


def find_monster(state: RuntimeState, monster_def_id: str) -> dict[str, Any] | None:
    for entity in state.entities.values():
        if entity.get("type") == "monster" and entity.get("monster_def_id") == monster_def_id:
            return entity
    return None


def find_interactable(state: RuntimeState, interactable_def_id: str) -> dict[str, Any] | None:
    for entity in state.entities.values():
        if entity.get("type") == "interactable" and entity.get("interactable_def_id") == interactable_def_id:
            return entity
    return None


def resolve_target(state: RuntimeState, step: dict[str, Any]) -> dict[str, Any]:
    if step.get("target_id"):
        target = state.entities.get(str(step["target_id"]))
        if target is None:
            raise AssertionError(f"{step.get('action')}: target not found: {step['target_id']}")
        return target
    if step.get("monster_def_id"):
        target = find_monster(state, str(step["monster_def_id"]))
        if target is None:
            raise AssertionError(f"{step.get('action')}: monster not found: {step}")
        return target
    if step.get("item_def_id"):
        target = find_loot(state, str(step["item_def_id"]))
        if target is None:
            raise AssertionError(f"{step.get('action')}: loot not found: {step}")
        return target
    if step.get("interactable_def_id"):
        target = find_interactable(state, str(step["interactable_def_id"]))
        if target is None:
            raise AssertionError(f"{step.get('action')}: interactable not found: {step}")
        return target
    raise AssertionError(f"{step.get('action')}: no target selector in {step}")


def find_inventory_item(inventory: list[dict[str, Any]], item_def_id: str) -> dict[str, Any] | None:
    for item in inventory:
        if item.get("item_def_id") == item_def_id:
            return item
    return None


def find_player(state: RuntimeState) -> dict[str, Any] | None:
    for entity in state.entities.values():
        if entity.get("type") == "player":
            return entity
    return None


async def check_persistence(base_url: str, token: str, session_id: str, item_id: str | None, assertions: list[Any]) -> None:
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

def assert_equipped_sword(inventory: list[dict], equipped: dict, item_id: str | None, where: str) -> None:
    item = find_inventory_item(inventory, "rusty_sword")
    if item is None:
        raise AssertionError(f"{where}: missing rusty_sword in inventory: {inventory}")
    if item["item_def_id"] != "rusty_sword" or not item["equipped"]:
        raise AssertionError(f"{where}: item not an equipped rusty_sword: {item}")
    expected_id = item_id or item["item_instance_id"]
    if equipped.get("weapon") != expected_id:
        raise AssertionError(f"{where}: equipped weapon {equipped.get('weapon')} != {item_id}")


def assert_player_damaged(entities: list[dict], where: str) -> None:
    players = [e for e in entities if e.get("type") == "player"]
    if len(players) != 1:
        raise AssertionError(f"{where}: expected one player entity, got {players}")
    hp = players[0].get("hp")
    if not isinstance(hp, int) or hp >= 10:
        raise AssertionError(f"{where}: player hp {hp} did not show retaliation damage")


def assert_player_position(state: RuntimeState, x: float, y: float, tolerance: float, where: str) -> None:
    player = find_player(state)
    if player is None:
        raise AssertionError(f"{where}: missing player entity")
    pos = player.get("position", {})
    got_x = float(pos.get("x", "nan"))
    got_y = float(pos.get("y", "nan"))
    if abs(got_x - x) > tolerance or abs(got_y - y) > tolerance:
        raise AssertionError(f"{where}: player position ({got_x}, {got_y}) != ({x}, {y})")


def assert_monster_dead(entities: list[dict], where: str, monster_def_id: str = "training_dummy") -> None:
    monsters = [e for e in entities if e.get("monster_def_id") == monster_def_id and e.get("type") == "monster"]
    if len(monsters) != 1:
        raise AssertionError(f"{where}: expected monster {monster_def_id}, got {monsters}")
    hp = monsters[0].get("hp")
    if hp != 0:
        raise AssertionError(f"{where}: monster {monster_def_id} hp {hp} != 0")


def assert_inventory_contains(inventory: list[dict], item_def_id: str, equipped: bool | None, where: str) -> None:
    item = find_inventory_item(inventory, item_def_id)
    if item is None:
        raise AssertionError(f"{where}: missing inventory item {item_def_id}: {inventory}")
    if equipped is not None and bool(item.get("equipped")) != equipped:
        raise AssertionError(f"{where}: {item_def_id} equipped={item.get('equipped')} want {equipped}")


def run_assertions(
    assertions: list[Any],
    entities: list[dict],
    inventory: list[dict],
    equipped: dict,
    item_id: str | None,
    where: str,
) -> None:
    for assertion in assertions:
        if isinstance(assertion, str):
            if assertion == "equipped_rusty_sword":
                assert_equipped_sword(inventory, equipped, item_id, where)
            elif assertion == "player_damaged":
                assert_player_damaged(entities, where)
            elif assertion == "monster_dead":
                assert_monster_dead(entities, where)
            else:
                raise AssertionError(f"{where}: unknown assertion {assertion}")
            continue

        if not isinstance(assertion, dict):
            raise AssertionError(f"{where}: assertion must be string or object: {assertion}")

        typ = assertion.get("type")
        if typ == "inventory_count":
            want = int(assertion["equals"])
            if len(inventory) != want:
                raise AssertionError(f"{where}: inventory count {len(inventory)} != {want}: {inventory}")
        elif typ == "inventory_contains":
            expected_equipped = assertion.get("equipped")
            assert_inventory_contains(inventory, str(assertion["item_def_id"]), expected_equipped, where)
        elif typ == "monster_dead":
            assert_monster_dead(entities, where, str(assertion["monster_def_id"]))
        elif typ == "interactable_state":
            expected_id = str(assertion["interactable_def_id"])
            expected_state = str(assertion["state"])
            matches = [
                e for e in entities
                if e.get("type") == "interactable" and e.get("interactable_def_id") == expected_id
            ]
            if len(matches) != 1 or matches[0].get("state") != expected_state:
                raise AssertionError(f"{where}: interactable {expected_id} state = {matches}, want {expected_state}")
        elif typ == "monster_killed_in_attacks":
            continue
        else:
            raise AssertionError(f"{where}: unknown assertion type {typ}")


def run_runtime_assertions(assertions: list[Any], state: RuntimeState, where: str) -> None:
    for assertion in assertions:
        if not isinstance(assertion, dict):
            continue
        typ = assertion.get("type")
        if typ != "monster_killed_in_attacks":
            continue
        monster_def_id = str(assertion["monster_def_id"])
        max_attacks = int(assertion["max_attacks"])
        count = state.accepted_attack_counts.get(monster_def_id, 0)
        if monster_def_id not in state.killed_monster_def_ids:
            raise AssertionError(f"{where}: monster {monster_def_id} was not observed killed at runtime")
        if count < 1:
            raise AssertionError(f"{where}: no accepted action_intent observed for {monster_def_id}")
        if count > max_attacks:
            raise AssertionError(
                f"{where}: monster {monster_def_id} killed in {count} accepted attacks, max {max_attacks}"
            )


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
            sess = create_session(client, token, scenario.world_id)
            session_id = sess["session_id"]
            last_session_id = session_id

            observed = asyncio.run(drive_scenario(args.base_url, token, sess, scenario))
            run_runtime_assertions(scenario.assertions, observed, "runtime protocol")

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
            asyncio.run(check_persistence(args.base_url, token, session_id, observed.item_id, scenario.assertions))

            replay = fetch_replay(client, token, args.debug_token, session_id)
            if not replay.get("match", False):
                raise AssertionError(f"replay mismatch for {session_id}: {replay.get('mismatch')}")
            log("replay verified for", session_id)

            results.append({
                "id": scenario.id,
                "title": scenario.title,
                "description": scenario.description,
                "world_id": scenario.world_id,
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
