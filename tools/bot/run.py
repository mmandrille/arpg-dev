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
import time
from typing import Any

import httpx
import websockets

from tools.bot.protocol import make_envelope, to_ws_url

SLICE_TIMEOUT_S = 20.0
WAIT_LOG_INTERVAL_S = 2.0
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
    stamp = datetime.now(timezone.utc).strftime("%H:%M:%S")
    print(f"[bot {stamp}]", *args, file=sys.stderr, flush=True)


def log_wait_progress(label: str, loop, started_at: float, **details: Any) -> None:
    elapsed = loop.time() - started_at
    parts = [f"{label} elapsed={elapsed:.1f}s"]
    for key, value in details.items():
        parts.append(f"{key}={value}")
    log(*parts)


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
    min_player_monster_distance: dict[str, float] = field(default_factory=dict)
    initial_monster_positions: dict[str, dict[str, float]] = field(default_factory=dict)
    current_level: int = 0
    visited_levels: set[int] = field(default_factory=lambda: {0})
    last_delta_level: int | None = None
    pending_level_load: int | None = None
    used_stair_positions: dict[str, dict[str, float]] = field(default_factory=dict)
    discovered_teleporters: dict[int, bool] = field(default_factory=dict)
    used_teleporter_positions: dict[int, dict[str, float]] = field(default_factory=dict)


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
    wanted = {normalize_scenario_selector(part.strip()) for part in selected.split(",") if part.strip()}
    found = [s for s in scenarios if scenario_matches_selector(s, wanted)]
    matched: set[str] = set()
    for scenario in found:
        aliases = {
            normalize_scenario_selector(scenario.id),
            normalize_scenario_selector(scenario.path.name),
            normalize_scenario_selector(scenario.path.stem),
        }
        matched.update(wanted & aliases)
    missing = wanted - matched
    if missing:
        raise ValueError(f"unknown scenario(s): {', '.join(sorted(missing))}")
    return found


def normalize_scenario_selector(value: str) -> str:
    if value.endswith(".json"):
        value = value[:-5]
    return value


def scenario_matches_selector(scenario: Scenario, wanted: set[str]) -> bool:
    return bool({
        normalize_scenario_selector(scenario.id),
        normalize_scenario_selector(scenario.path.name),
        normalize_scenario_selector(scenario.path.stem),
    } & wanted)


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

        total_steps = len(scenario.steps)
        for index, step in enumerate(scenario.steps):
            action = step.get("action", "?")
            log(f"step {index + 1}/{total_steps} begin action={action}")
            step_started = loop.time()
            await execute_step(ws, sid, state, step, loop, index=index, total=total_steps)
            log(
                f"step {index + 1}/{total_steps} done action={action}",
                f"elapsed={loop.time() - step_started:.2f}s",
                f"tick={state.last_tick}",
            )

        if state.equipped_item_id:
            log("equipped item", state.equipped_item_id, "- scenario complete over protocol")
        else:
            log("scenario complete over protocol")
        return state


async def execute_step(
    ws,
    session_id: str,
    state: RuntimeState,
    step: dict[str, Any],
    loop,
    *,
    index: int = 0,
    total: int = 0,
) -> None:
    action = step.get("action")
    if action == "wait_ticks":
        ticks = int(step["ticks"])
        deadline = loop.time() + (ticks * 0.05) + 0.15
        while loop.time() < deadline:
            await pump_one(ws, state, timeout=0.05)

        return

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
        await walk_toward(
            ws,
            session_id,
            state,
            {"x": float(step["x"]), "y": float(step["y"])},
            loop,
            stop_distance=float(step.get("tolerance", 0.25)),
        )
        return

    if action == "assert_player_position":
        assert_player_position(state, float(step["x"]), float(step["y"]), float(step.get("tolerance", 0.001)), "runtime protocol")
        return

    if action == "assert_player_at_used_stair":
        direction = str(step["direction"])
        pos = state.used_stair_positions.get(direction)
        if pos is None:
            raise AssertionError(f"assert_player_at_used_stair: no recorded {direction} stair")
        assert_player_position(
            state,
            float(pos["x"]),
            float(pos["y"]),
            float(step.get("tolerance", 0.001)),
            "runtime protocol",
        )
        return

    if action == "assert_teleporter_discovered":
        level = int(step["level"])
        want = bool(step.get("discovered", True))
        got = bool(state.discovered_teleporters.get(level, False))
        if got != want:
            raise AssertionError(f"assert_teleporter_discovered: level {level} discovered={got}, want {want}")
        return

    if action == "assert_player_at_discovered_teleporter":
        level = int(step.get("level", state.current_level))
        pos = state.used_teleporter_positions.get(level)
        if pos is None:
            raise AssertionError(f"assert_player_at_discovered_teleporter: no recorded teleporter for level {level}")
        assert_player_position(
            state,
            float(pos["x"]),
            float(pos["y"]),
            float(step.get("tolerance", 0.001)),
            "runtime protocol",
        )
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
        wait_started = loop.time()
        last_wait_log = wait_started
        deadline = loop.time() + SLICE_TIMEOUT_S
        while event_type not in state.seen_events:
            if loop.time() > deadline:
                raise TimeoutError(f"attack_until_event stalled waiting for {event_type}")
            if loop.time() - last_wait_log >= WAIT_LOG_INTERVAL_S:
                monster = state.entities.get(target_id)
                log_wait_progress(
                    f"attack_until_event waiting for {event_type}",
                    loop,
                    wait_started,
                    monster_hp=(monster or {}).get("hp"),
                    seen_events=sorted(state.seen_events),
                    tick=state.last_tick,
                )
                last_wait_log = loop.time()
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

    if action == "action_once_until_event":
        target = resolve_target(state, step)
        event_type = str(step["event_type"])
        target_id = str(target["id"])
        monster_def_id = str(step.get("monster_def_id") or target.get("monster_def_id") or "")
        env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": target_id})
        if monster_def_id:
            state.pending_attack_monsters[env["message_id"]] = monster_def_id
        await ws.send(json.dumps(env))
        await wait_for_accept(ws, state, env["message_id"], loop)
        await wait_for_event(ws, state, event_type, loop)
        return

    if action == "action_entity":
        target = resolve_target(state, step)
        target_type = str(target.get("type", ""))
        target_item_def_id = str(target.get("item_def_id", step.get("item_def_id", "")))
        env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": str(target["id"])})
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        if target_type == "loot" and target_item_def_id:
            deadline = loop.time() + SLICE_TIMEOUT_S
            while find_inventory_item(state.inventory, target_item_def_id) is None:
                if loop.time() > deadline:
                    raise TimeoutError(f"action_entity stalled waiting for loot pickup {target_item_def_id}")
                await pump_one(ws, state, timeout=0.1)
            return
        event_type = step.get("event_type")
        if event_type:
            await wait_for_event(ws, state, str(event_type), loop)
        return

    if action == "kill_monsters":
        monster_def_id = str(step["monster_def_id"])
        max_count = int(step.get("count", 99))
        killed = 0
        while killed < max_count:
            target = find_monster(state, monster_def_id)
            if target is None:
                return
            target_id = str(target["id"])
            deadline = loop.time() + SLICE_TIMEOUT_S
            last_action = 0.0
            while True:
                current = state.entities.get(target_id)
                if current is None or int(current.get("hp", 0)) <= 0:
                    killed += 1
                    break
                if loop.time() > deadline:
                    raise TimeoutError(f"kill_monsters stalled waiting for {monster_def_id} {target_id}")
                if loop.time() - last_action > 0.12:
                    env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": target_id})
                    state.pending_attack_monsters[env["message_id"]] = monster_def_id
                    await ws.send(json.dumps(env))
                    last_action = loop.time()
                await pump_one(ws, state, timeout=0.1)
        return

    if action == "move_until_in_range":
        target = resolve_target(state, step)
        await walk_toward(ws, session_id, state, target["position"], loop, stop_distance=float(step.get("stop_distance", WALK_STOP_DISTANCE)))
        return

    if action == "use_stair":
        direction = str(step["direction"])
        if direction not in {"down", "up"}:
            raise AssertionError(f"use_stair: direction must be down/up: {step}")
        stair_def_id = "stairs_down" if direction == "down" else "stairs_up"
        target = find_interactable(state, stair_def_id)
        if target is None:
            raise AssertionError(f"use_stair: missing {stair_def_id} on level {state.current_level}")
        target_pos = target.get("position", {})
        state.used_stair_positions[direction] = {
            "x": float(target_pos.get("x", "nan")),
            "y": float(target_pos.get("y", "nan")),
        }
        await walk_toward(
            ws,
            session_id,
            state,
            target["position"],
            loop,
            stop_distance=float(step.get("stop_distance", WALK_STOP_DISTANCE)),
            max_ticks=int(step.get("max_ticks", WALK_MAX_TICKS)),
        )
        msg_type = "descend_intent" if direction == "down" else "ascend_intent"
        previous_level = state.current_level
        env = make_envelope(msg_type, session_id, state.last_tick, {})
        await ws.send(json.dumps(env))
        await wait_for_accept(ws, state, env["message_id"], loop)
        await wait_for_level_change(ws, state, previous_level, loop)
        return

    if action == "discover_teleporter":
        target = find_interactable(state, "teleporter")
        if target is None:
            raise AssertionError(f"discover_teleporter: missing teleporter on level {state.current_level}")
        await walk_toward(
            ws,
            session_id,
            state,
            target["position"],
            loop,
            stop_distance=float(step.get("stop_distance", WALK_STOP_DISTANCE)),
            max_ticks=int(step.get("max_ticks", WALK_MAX_TICKS)),
        )
        target_pos = target.get("position", {})
        state.used_teleporter_positions[state.current_level] = {
            "x": float(target_pos.get("x", "nan")),
            "y": float(target_pos.get("y", "nan")),
        }
        env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": target["id"]})
        await ws.send(json.dumps(env))
        await wait_for_accept(ws, state, env["message_id"], loop)
        if bool(step.get("wait_discovered", True)):
            await wait_for_teleporter_discovery(ws, state, state.current_level, loop)
        return

    if action == "teleport_to_level":
        target_level = int(step["target_level"])
        previous_level = state.current_level
        env = make_envelope("teleport_intent", session_id, state.last_tick, {"target_level": target_level})
        await ws.send(json.dumps(env))
        await wait_for_accept(ws, state, env["message_id"], loop)
        await wait_for_level_change(ws, state, previous_level, loop)
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
        wait_started = loop.time()
        last_wait_log = wait_started
        while find_inventory_item(state.inventory, item_def_id) is None:
            if loop.time() > deadline:
                raise TimeoutError(f"pick_up_loot stalled waiting for {item_def_id}")
            if loop.time() - last_wait_log >= WAIT_LOG_INTERVAL_S:
                log_wait_progress(
                    f"pick_up_loot waiting for {item_def_id}",
                    loop,
                    wait_started,
                    inventory_count=len(state.inventory),
                    loot_ids=state.loot_ids,
                    tick=state.last_tick,
                )
                last_wait_log = loop.time()
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

    if action == "unequip_slot":
        slot = str(step.get("slot", "weapon"))
        deadline = loop.time() + SLICE_TIMEOUT_S
        if state.equipped.get(slot) is None:
            raise AssertionError(f"unequip_slot: slot {slot} is already empty")
        await ws.send(json.dumps(make_envelope(
            "unequip_intent", session_id, state.last_tick, {"slot": slot})))
        log("unequipping", slot)
        while state.equipped.get(slot) is not None:
            if loop.time() > deadline:
                raise TimeoutError(f"unequip_slot stalled waiting for empty {slot}")
            await pump_one(ws, state, timeout=0.1)
        return

    if action == "assert_player_hp":
        want = int(step["equals"])
        player = find_player(state)
        if player is None:
            raise AssertionError("assert_player_hp: player not found in runtime state")
        assert_player_hp_equals([player], want, "runtime protocol")
        return

    if action == "use_inventory_item":
        item_def_id = str(step["item_def_id"])
        bag_index = int(step.get("bag_index", 0))
        deadline = loop.time() + SLICE_TIMEOUT_S
        item = find_inventory_item(state.inventory, item_def_id, bag_index)
        if item is None:
            raise AssertionError(f"use_inventory_item: missing inventory item {item_def_id} at bag_index={bag_index}")
        item_id = str(item["item_instance_id"])
        await ws.send(json.dumps(make_envelope(
            "use_intent", session_id, state.last_tick, {"item_instance_id": item_id})))
        log("using", item_def_id, item_id)
        wait_started = loop.time()
        last_wait_log = wait_started
        while any(str(i.get("item_instance_id")) == item_id for i in state.inventory):
            if loop.time() > deadline:
                raise TimeoutError(f"use_inventory_item stalled waiting for removal of {item_id}")
            if loop.time() - last_wait_log >= WAIT_LOG_INTERVAL_S:
                player = find_player(state)
                log_wait_progress(
                    f"use_inventory_item waiting for removal of {item_id}",
                    loop,
                    wait_started,
                    player_hp=(player or {}).get("hp"),
                    inventory_count=len(state.inventory),
                    seen_events=sorted(state.seen_events),
                    tick=state.last_tick,
                )
                last_wait_log = loop.time()
            await pump_one(ws, state, timeout=0.1)
        return

    if action == "drop_inventory_item":
        item_def_id = str(step["item_def_id"])
        deadline = loop.time() + SLICE_TIMEOUT_S
        item = find_inventory_item(state.inventory, item_def_id)
        if item is None:
            raise AssertionError(f"drop_inventory_item: missing inventory item {item_def_id}")
        item_id = str(item["item_instance_id"])
        await ws.send(json.dumps(make_envelope(
            "drop_intent", session_id, state.last_tick, {"item_instance_id": item_id})))
        log("dropping", item_def_id, item_id)
        while any(str(i.get("item_instance_id")) == item_id for i in state.inventory):
            if loop.time() > deadline:
                raise TimeoutError(f"drop_inventory_item stalled waiting for removal of {item_id}")
            await pump_one(ws, state, timeout=0.1)
        while find_loot(state, item_def_id) is None:
            if loop.time() > deadline:
                raise TimeoutError(f"drop_inventory_item stalled waiting for loot {item_def_id}")
            await pump_one(ws, state, timeout=0.1)
        if state.item_id == item_id:
            state.item_id = None
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

        candidates: list[dict[str, int]] = []
        if abs(dx) > 0:
            candidates.append({"x": 1 if dx > 0 else -1, "y": 0})
        if abs(dy) > 0:
            candidates.append({"x": 0, "y": 1 if dy > 0 else -1})
        if abs(dx) > 0 and abs(dy) > 0:
            candidates.append({"x": 1 if dx > 0 else -1, "y": 1 if dy > 0 else -1})

        moved = False
        for direction in candidates:
            before = {"x": player_pos["x"], "y": player_pos["y"]}
            env = make_envelope(
                "move_intent",
                session_id,
                state.last_tick,
                {"direction": direction, "duration_ticks": 1},
            )
            await ws.send(json.dumps(env))
            if await wait_for_player_move_or_accept(ws, state, before, env["message_id"], loop):
                moved = True
                break
        if moved:
            continue

    raise TimeoutError(f"walk_toward exhausted {max_ticks} ticks toward {target_pos}")


async def wait_for_player_move_or_accept(ws, state: RuntimeState, before: dict[str, Any], message_id: str, loop) -> bool:
    deadline = loop.time() + SLICE_TIMEOUT_S
    while True:
        player = find_player(state)
        if player is not None:
            pos = player["position"]
            if pos.get("x") != before.get("x") or pos.get("y") != before.get("y"):
                return True
        if message_id in state.accepted_message_ids:
            return False
        if message_id in state.rejected_message_reasons:
            raise AssertionError(f"move_intent rejected: {state.rejected_message_reasons[message_id]}")
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for player movement from {before}")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_tick_advance(ws, state: RuntimeState, loop) -> None:
    start = state.last_tick
    deadline = loop.time() + SLICE_TIMEOUT_S
    while state.last_tick <= start:
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for tick advance from {start}")
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


async def wait_for_level_change(ws, state: RuntimeState, previous_level: int, loop) -> None:
    deadline = loop.time() + SLICE_TIMEOUT_S
    while state.current_level == previous_level or state.pending_level_load is not None:
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for level change from {previous_level}")
        await pump_one(ws, state, timeout=0.1)


async def wait_for_teleporter_discovery(ws, state: RuntimeState, level: int, loop) -> None:
    deadline = loop.time() + SLICE_TIMEOUT_S
    while not state.discovered_teleporters.get(level, False):
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for teleporter discovery level {level}")
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
    delta_level = int(p.get("level", state.current_level))
    state.last_delta_level = delta_level
    state.visited_levels.add(delta_level)
    for ev in (p.get("events") or []):
        event_type = ev["event_type"]
        state.seen_events.add(event_type)
        if event_type == "level_changed":
            state.current_level = int(ev["to_level"])
            state.pending_level_load = state.current_level
            clear_active_level_state(state)
            state.visited_levels.add(int(ev["from_level"]))
            state.visited_levels.add(int(ev["to_level"]))
        if event_type == "monster_killed":
            state.killed = True
            entity = state.entities.get(str(ev.get("entity_id")))
            if entity is not None and entity.get("monster_def_id"):
                state.killed_monster_def_ids.add(str(entity["monster_def_id"]))
            log("monster killed at tick", p.get("server_tick"))
    apply_level_entities = delta_level == state.current_level
    for c in (p.get("changes") or []):
        if c["op"] in {"entity_spawn", "entity_update"}:
            if not apply_level_entities:
                continue
            entity = c["entity"]
            existing = state.entities.get(entity["id"], {})
            existing.update(entity)
            state.entities[entity["id"]] = existing
            track_initial_monster_position(state, existing)
            if c["op"] == "entity_spawn" and entity["type"] == "loot":
                loot_id = entity["id"]
                if loot_id not in state.loot_ids:
                    state.loot_ids.append(loot_id)
        elif c["op"] == "entity_remove":
            if not apply_level_entities:
                continue
            entity_id = c["entity_id"]
            state.entities.pop(entity_id, None)
            if entity_id in state.loot_ids:
                state.loot_ids.remove(entity_id)
        elif c["op"] == "inventory_add":
            upsert_inventory(state, c["item"])
            state.item_id = c["item"]["item_instance_id"]
        elif c["op"] == "inventory_update":
            upsert_inventory(state, c["item"])
        elif c["op"] == "inventory_remove":
            remove_inventory_item(state, str(c["item_instance_id"]))
        elif c["op"] == "equipped_update" and c.get("slot") == "weapon":
            state.equipped_item_id = c.get("item_instance_id")
            state.equipped[c["slot"]] = c.get("item_instance_id")
        elif c["op"] == "teleporter_discovery_update":
            state.discovered_teleporters[int(c["level"])] = bool(c["discovered"])
    if state.pending_level_load is not None and delta_level == state.pending_level_load:
        state.pending_level_load = None
    update_runtime_distances(state)


def ingest_snapshot(payload: dict[str, Any], state: RuntimeState) -> None:
    state.last_tick = max(state.last_tick, int(payload.get("server_tick", 0)))
    state.current_level = int(payload.get("current_level", 0))
    state.visited_levels.add(state.current_level)
    state.last_delta_level = state.current_level
    state.pending_level_load = None
    state.entities = {str(e["id"]): dict(e) for e in payload.get("entities", [])}
    state.inventory = [dict(i) for i in payload.get("inventory", [])]
    state.equipped = dict(payload.get("equipped", {}))
    state.discovered_teleporters = parse_discovered_teleporters(payload)
    state.loot_ids = [
        entity_id
        for entity_id, entity in state.entities.items()
        if entity.get("type") == "loot"
    ]
    for item in state.inventory:
        if item.get("equipped"):
            state.equipped_item_id = item.get("item_instance_id")
    for entity in state.entities.values():
        track_initial_monster_position(state, entity)
    update_runtime_distances(state)


def parse_discovered_teleporters(payload: dict[str, Any]) -> dict[int, bool]:
    return {
        int(row["level"]): bool(row["discovered"])
        for row in payload.get("discovered_teleporters", [])
    }


def clear_active_level_state(state: RuntimeState) -> None:
    state.entities.clear()
    state.loot_ids.clear()
    state.min_player_monster_distance.clear()
    state.initial_monster_positions.clear()


def track_initial_monster_position(state: RuntimeState, entity: dict[str, Any]) -> None:
    if entity.get("type") != "monster":
        return
    monster_def_id = str(entity.get("monster_def_id", ""))
    if not monster_def_id or monster_def_id in state.initial_monster_positions:
        return
    pos = entity.get("position", {})
    state.initial_monster_positions[monster_def_id] = {
        "x": float(pos.get("x", 0.0)),
        "y": float(pos.get("y", 0.0)),
    }


def update_runtime_distances(state: RuntimeState) -> None:
    player = find_player(state)
    if player is None:
        return
    ppos = player.get("position", {})
    px = float(ppos.get("x", 0.0))
    py = float(ppos.get("y", 0.0))
    for entity in state.entities.values():
        if entity.get("type") != "monster":
            continue
        if int(entity.get("hp", 0)) <= 0:
            continue
        monster_def_id = str(entity.get("monster_def_id", ""))
        if not monster_def_id:
            continue
        mpos = entity.get("position", {})
        dx = px - float(mpos.get("x", 0.0))
        dy = py - float(mpos.get("y", 0.0))
        dist = (dx * dx + dy * dy) ** 0.5
        current = state.min_player_monster_distance.get(monster_def_id)
        if current is None or dist < current:
            state.min_player_monster_distance[monster_def_id] = dist


def upsert_inventory(state: RuntimeState, item: dict[str, Any]) -> None:
    item_id = item["item_instance_id"]
    for i, current in enumerate(state.inventory):
        if current.get("item_instance_id") == item_id:
            merged = dict(current)
            merged.update(item)
            state.inventory[i] = merged
            return
    state.inventory.append(dict(item))


def remove_inventory_item(state: RuntimeState, item_instance_id: str) -> None:
    state.inventory = [
        item for item in state.inventory
        if str(item.get("item_instance_id")) != item_instance_id
    ]
    if state.item_id == item_instance_id:
        state.item_id = None
    if state.equipped_item_id == item_instance_id:
        state.equipped_item_id = None


def find_loot(state: RuntimeState, item_def_id: str) -> dict[str, Any] | None:
    for entity in state.entities.values():
        if entity.get("type") == "loot" and entity.get("item_def_id") == item_def_id:
            return entity
    return None


def find_monster(state: RuntimeState, monster_def_id: str) -> dict[str, Any] | None:
    for entity in state.entities.values():
        if (
            entity.get("type") == "monster"
            and entity.get("monster_def_id") == monster_def_id
            and int(entity.get("hp", 0)) > 0
        ):
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


def find_inventory_item(
    inventory: list[dict[str, Any]],
    item_def_id: str,
    bag_index: int = 0,
) -> dict[str, Any] | None:
    matches = [item for item in inventory if item.get("item_def_id") == item_def_id]
    if bag_index < 0 or bag_index >= len(matches):
        return None
    return matches[bag_index]


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
        run_assertions(
            assertions,
            payload["entities"],
            inv,
            equipped,
            item_id,
            "reconnect snapshot",
            current_level=int(payload.get("current_level", 0)),
            discovered_teleporters=parse_discovered_teleporters(payload),
        )
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


def assert_player_hp_equals(entities: list[dict], want: int, where: str) -> None:
    player = find_player_entities(entities)
    if player is None:
        raise AssertionError(f"{where}: player not found")
    hp = player.get("hp")
    if hp != want:
        raise AssertionError(f"{where}: player hp {hp} != {want}")


def find_player_entities(entities: list[dict]) -> dict | None:
    for entity in entities:
        if entity.get("type") == "player":
            return entity
    return None


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
    current_level: int | None = None,
    discovered_teleporters: dict[int, bool] | None = None,
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
        elif typ == "equipped_weapon_def":
            expected_def = str(assertion["item_def_id"])
            weapon_id = equipped.get("weapon")
            if weapon_id is None:
                raise AssertionError(f"{where}: equipped weapon is empty, want {expected_def}")
            item = next((i for i in inventory if str(i.get("item_instance_id")) == str(weapon_id)), None)
            if item is None or item.get("item_def_id") != expected_def:
                raise AssertionError(f"{where}: equipped weapon {weapon_id} row = {item}, want {expected_def}")
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
        elif typ in {"monster_moved", "monster_within_player_distance", "monster_near_spawn", "event_seen", "player_never_in_melee_range_of"}:
            continue
        elif typ == "player_hp_equals":
            assert_player_hp_equals(entities, int(assertion["equals"]), where)
        elif typ == "current_level":
            if current_level is None:
                raise AssertionError(f"{where}: current_level unavailable")
            want = int(assertion["equals"])
            if current_level != want:
                raise AssertionError(f"{where}: current_level {current_level} != {want}")
        elif typ == "teleporter_discovered":
            if discovered_teleporters is None:
                raise AssertionError(f"{where}: discovered_teleporters unavailable")
            level = int(assertion["level"])
            want = bool(assertion.get("discovered", True))
            got = bool(discovered_teleporters.get(level, False))
            if got != want:
                raise AssertionError(f"{where}: teleporter level {level} discovered={got}, want {want}")
        elif typ == "teleporter_list_contains":
            if discovered_teleporters is None:
                raise AssertionError(f"{where}: discovered_teleporters unavailable")
            level = int(assertion["level"])
            if level not in discovered_teleporters:
                raise AssertionError(f"{where}: teleporter level {level} missing; have {sorted(discovered_teleporters)}")
        elif typ == "visited_levels_contain":
            continue
        else:
            raise AssertionError(f"{where}: unknown assertion type {typ}")


def run_runtime_assertions(assertions: list[Any], state: RuntimeState, where: str) -> None:
    for assertion in assertions:
        if not isinstance(assertion, dict):
            continue
        typ = assertion.get("type")
        if typ == "player_never_in_melee_range_of":
            monster_def_id = str(assertion["monster_def_id"])
            min_distance = state.min_player_monster_distance.get(monster_def_id)
            if min_distance is None:
                raise AssertionError(f"{where}: no distance samples for monster {monster_def_id}")
            reach = float(assertion.get("reach", 1.5))
            monster_radius = float(assertion.get("monster_radius", 0.45))
            threshold = reach + monster_radius + 0.000001
            if min_distance <= threshold:
                raise AssertionError(
                    f"{where}: player entered melee range of {monster_def_id}: min_distance={min_distance:.3f}, threshold={threshold:.3f}"
                )
            continue
        if typ == "monster_killed_in_attacks":
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
            continue
        if typ == "event_seen":
            event_type = str(assertion["event_type"])
            if event_type not in state.seen_events:
                raise AssertionError(f"{where}: event {event_type} not seen; have {sorted(state.seen_events)}")
            continue
        if typ == "monster_moved":
            monster_def_id = str(assertion["monster_def_id"])
            min_distance = float(assertion["min_distance"])
            initial = state.initial_monster_positions.get(monster_def_id)
            monster = find_monster(state, monster_def_id)
            if initial is None or monster is None:
                raise AssertionError(f"{where}: missing initial/current monster {monster_def_id}")
            pos = monster.get("position", {})
            dx = float(pos.get("x", 0.0)) - initial["x"]
            dy = float(pos.get("y", 0.0)) - initial["y"]
            dist = (dx * dx + dy * dy) ** 0.5
            if dist < min_distance:
                raise AssertionError(
                    f"{where}: monster {monster_def_id} moved {dist:.3f} < min_distance {min_distance}"
                )
            continue
        if typ == "monster_within_player_distance":
            monster_def_id = str(assertion["monster_def_id"])
            max_distance = float(assertion["max_distance"])
            player = find_player(state)
            monster = find_monster(state, monster_def_id)
            if player is None or monster is None:
                raise AssertionError(f"{where}: missing player or monster {monster_def_id}")
            ppos = player.get("position", {})
            mpos = monster.get("position", {})
            dx = float(ppos.get("x", 0.0)) - float(mpos.get("x", 0.0))
            dy = float(ppos.get("y", 0.0)) - float(mpos.get("y", 0.0))
            dist = (dx * dx + dy * dy) ** 0.5
            if dist > max_distance:
                raise AssertionError(
                    f"{where}: monster {monster_def_id} distance {dist:.3f} > max_distance {max_distance}"
                )
            continue
        if typ == "monster_near_spawn":
            monster_def_id = str(assertion["monster_def_id"])
            max_distance_from_spawn = float(assertion["max_distance_from_spawn"])
            spawn = state.initial_monster_positions.get(monster_def_id)
            monster = find_monster(state, monster_def_id)
            if spawn is None or monster is None:
                raise AssertionError(f"{where}: missing spawn/current monster {monster_def_id}")
            mpos = monster.get("position", {})
            dx = float(mpos.get("x", 0.0)) - spawn["x"]
            dy = float(mpos.get("y", 0.0)) - spawn["y"]
            dist = (dx * dx + dy * dy) ** 0.5
            if dist > max_distance_from_spawn:
                raise AssertionError(
                    f"{where}: monster {monster_def_id} spawn distance {dist:.3f} > {max_distance_from_spawn}"
                )
            continue
        if typ == "current_level":
            want = int(assertion["equals"])
            if state.current_level != want:
                raise AssertionError(f"{where}: current_level {state.current_level} != {want}")
            continue
        if typ == "visited_levels_contain":
            want = int(assertion["level"])
            if want not in state.visited_levels:
                raise AssertionError(f"{where}: level {want} not visited; have {sorted(state.visited_levels)}")
            continue
        if typ == "teleporter_discovered":
            level = int(assertion["level"])
            want = bool(assertion.get("discovered", True))
            got = bool(state.discovered_teleporters.get(level, False))
            if got != want:
                raise AssertionError(f"{where}: teleporter level {level} discovered={got}, want {want}")
            continue
        if typ == "teleporter_list_contains":
            level = int(assertion["level"])
            if level not in state.discovered_teleporters:
                raise AssertionError(
                    f"{where}: teleporter level {level} missing; have {sorted(state.discovered_teleporters)}"
                )
            continue
        if typ == "inventory_contains":
            expected_equipped = assertion.get("equipped")
            assert_inventory_contains(state.inventory, str(assertion["item_def_id"]), expected_equipped, where)
            continue


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
            scenario_started = time.monotonic()
            log("scenario begin", scenario.id, "-", scenario.title, f"world={scenario.world_id}")
            sess = create_session(client, token, scenario.world_id)
            session_id = sess["session_id"]
            last_session_id = session_id
            log("session created", session_id, f"seed={sess.get('seed')}")

            phase_started = time.monotonic()
            observed = asyncio.run(drive_scenario(args.base_url, token, sess, scenario))
            log("phase drive done", f"elapsed={time.monotonic() - phase_started:.2f}s")
            run_runtime_assertions(scenario.assertions, observed, "runtime protocol")

            # Assert authoritative state through the inspection API.
            phase_started = time.monotonic()
            state = fetch_state(client, token, args.debug_token, session_id)
            run_assertions(
                scenario.assertions,
                state["entities"],
                state["inventory"],
                state["equipped"],
                observed.item_id,
                "/state API",
                current_level=int(state.get("current_level", 0)),
                discovered_teleporters=parse_discovered_teleporters(state),
            )
            log("phase /state done", f"elapsed={time.monotonic() - phase_started:.2f}s")

            # Assert replay reconstruction by reconnecting a fresh session loop.
            phase_started = time.monotonic()
            asyncio.run(check_persistence(args.base_url, token, session_id, observed.item_id, scenario.assertions))
            log("phase reconnect done", f"elapsed={time.monotonic() - phase_started:.2f}s")

            phase_started = time.monotonic()
            replay = fetch_replay(client, token, args.debug_token, session_id)
            if not replay.get("match", False):
                raise AssertionError(f"replay mismatch for {session_id}: {replay.get('mismatch')}")
            log("phase replay done", f"elapsed={time.monotonic() - phase_started:.2f}s", session_id)
            log("scenario done", scenario.id, f"elapsed={time.monotonic() - scenario_started:.2f}s")

            scenario_visual = json.loads(scenario.path.read_text()).get("visual")
            entry = {
                "id": scenario.id,
                "title": scenario.title,
                "description": scenario.description,
                "world_id": scenario.world_id,
                "session_id": session_id,
                "seed": sess["seed"],
                "final_tick": observed.last_tick,
                "status": "passed",
                "replay_match": True,
            }
            if isinstance(scenario_visual, dict):
                entry["visual"] = scenario_visual
            results.append(entry)

    if args.write_manifest:
        write_manifest(args.write_manifest, args.base_url, results)
        log("wrote manifest", args.write_manifest)

    log("BOT OK", "- scenarios:", ", ".join(r["id"] for r in results))
    if args.print_session_id:
        print(last_session_id)
    return 0


if __name__ == "__main__":
    sys.exit(main())
