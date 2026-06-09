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
import math
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
BOT_RUN_ARTIFACT_DIR = ROOT / ".artifacts" / "bot-runs"


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
    seed: str
    peer_count: int
    title: str
    description: str
    steps: list[dict[str, Any]]
    assertions: list[Any]
    fresh_session_checks: list[dict[str, Any]]
    path: Path


@dataclass
class RuntimeState:
    world_id: str = DEFAULT_WORLD_ID
    local_player_id: str = ""
    party: list[dict[str, Any]] = field(default_factory=list)
    last_tick: int = 0
    killed: bool = False
    walls: list[dict[str, Any]] = field(default_factory=list)
    entities: dict[str, dict[str, Any]] = field(default_factory=dict)
    inventory: list[dict[str, Any]] = field(default_factory=list)
    equipped: dict[str, Any] = field(default_factory=dict)
    hotbar_capacity: int = 2
    hotbar: list[dict[str, Any]] = field(default_factory=list)
    inventory_rows: int = 3
    inventory_capacity: int = 15
    gold: int = 0
    loot_ids: list[str] = field(default_factory=list)
    item_id: str | None = None
    equipped_item_id: str | None = None
    seen_events: set[str] = field(default_factory=set)
    combat_events: list[dict[str, Any]] = field(default_factory=list)
    pending_attack_monsters: dict[str, str] = field(default_factory=dict)
    accepted_attack_counts: dict[str, int] = field(default_factory=dict)
    killed_monster_def_ids: set[str] = field(default_factory=set)
    max_monster_damage_by_def: dict[str, int] = field(default_factory=dict)
    accepted_message_ids: set[str] = field(default_factory=set)
    rejected_message_reasons: dict[str, str] = field(default_factory=dict)
    min_player_monster_distance: dict[str, float] = field(default_factory=dict)
    initial_monster_positions: dict[str, dict[str, float]] = field(default_factory=dict)
    initial_entity_positions: dict[str, dict[str, float]] = field(default_factory=dict)
    recorded_player_hp: int | None = None
    current_level: int = 0
    visited_levels: set[int] = field(default_factory=lambda: {0})
    last_delta_level: int | None = None
    pending_level_load: int | None = None
    used_stair_positions: dict[str, dict[str, float]] = field(default_factory=dict)
    discovered_teleporters: dict[int, bool] = field(default_factory=dict)
    used_teleporter_positions: dict[int, dict[str, float]] = field(default_factory=dict)
    character_progression: dict[str, Any] = field(default_factory=dict)
    shop_offers: dict[str, dict[str, dict[str, Any]]] = field(default_factory=dict)
    shop_sell_appraisals: dict[str, dict[str, dict[str, Any]]] = field(default_factory=dict)
    shop_events: list[dict[str, Any]] = field(default_factory=list)
    last_shop_event: dict[str, Any] | None = None
    last_gold_before_action: int | None = None
    last_gold_after_action: int | None = None


@dataclass
class CoopPeer:
    label: str
    token: str
    session: dict[str, Any]
    state: RuntimeState
    ws: Any


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
            seed=str(raw.get("seed", "")),
            peer_count=int(raw.get("peer_count", 2)),
            title=raw.get("title", sid),
            description=raw.get("description", ""),
            steps=list(raw.get("steps", [])),
            assertions=list(raw.get("assertions", [])),
            fresh_session_checks=list(raw.get("fresh_session_checks", [])),
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


def create_session(client: httpx.Client, token: str, world_id: str, seed: str = "") -> dict[str, Any]:
    body: dict[str, Any] = {"mode": "solo", "world_id": world_id}
    if seed:
        body["seed"] = seed
    resp = client.post("/v0/sessions", headers=auth(token), json=body)
    resp.raise_for_status()
    body = resp.json()
    log("session", body["session_id"], "seed", body["seed"], "world", body.get("world_id"))
    return body


def list_characters(client: httpx.Client, token: str) -> list[dict[str, Any]]:
    resp = client.get("/v0/characters", headers=auth(token))
    resp.raise_for_status()
    return list(resp.json().get("characters", []))


def create_character(client: httpx.Client, token: str, name: str) -> dict[str, Any]:
    resp = client.post("/v0/characters", headers=auth(token), json={"name": name})
    resp.raise_for_status()
    return resp.json()


def ensure_character(client: httpx.Client, token: str, name: str) -> str:
    chars = list_characters(client, token)
    if chars:
        return str(chars[0]["character_id"])
    return str(create_character(client, token, name)["character_id"])


def create_coop_session(
    client: httpx.Client,
    token: str,
    world_id: str,
    character_id: str,
    seed: str = "",
) -> dict[str, Any]:
    body: dict[str, Any] = {"mode": "coop", "world_id": world_id, "character_id": character_id}
    if seed:
        body["seed"] = seed
    resp = client.post("/v0/sessions", headers=auth(token), json=body)
    resp.raise_for_status()
    body = resp.json()
    if not body.get("join_code"):
        raise AssertionError(f"co-op create did not return join_code: {body}")
    log("co-op session", body["session_id"], "seed", body["seed"], "world", body.get("world_id"))
    return body


def create_listed_coop_session(
    client: httpx.Client,
    token: str,
    world_id: str,
    character_id: str,
    seed: str = "",
) -> dict[str, Any]:
    body: dict[str, Any] = {"mode": "coop", "listed": True, "world_id": world_id, "character_id": character_id}
    if seed:
        body["seed"] = seed
    resp = client.post("/v0/sessions", headers=auth(token), json=body)
    resp.raise_for_status()
    body = resp.json()
    if not body.get("listed"):
        raise AssertionError(f"listed co-op create did not return listed=true: {body}")
    log("listed co-op session", body["session_id"], "seed", body["seed"], "world", body.get("world_id"))
    return body


def list_active_sessions(client: httpx.Client, token: str) -> list[dict[str, Any]]:
    resp = client.get("/v0/sessions/active", headers=auth(token))
    resp.raise_for_status()
    return list(resp.json().get("sessions", []))


def join_coop_session(client: httpx.Client, token: str, session_id: str, join_code: str, character_id: str) -> dict[str, Any]:
    resp = client.post(
        f"/v0/sessions/{session_id}/join",
        headers=auth(token),
        json={"join_code": join_code, "character_id": character_id},
    )
    resp.raise_for_status()
    body = resp.json()
    log("joined co-op session", body["session_id"], "character", character_id)
    return body


def join_listed_session(client: httpx.Client, token: str, session_id: str, character_id: str) -> dict[str, Any]:
    resp = client.post(
        f"/v0/sessions/{session_id}/join",
        headers=auth(token),
        json={"character_id": character_id},
    )
    resp.raise_for_status()
    body = resp.json()
    if body.get("join_code"):
        raise AssertionError(f"listed join leaked join_code: {body}")
    log("joined listed co-op session", body["session_id"], "character", character_id)
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
        state.world_id = scenario.world_id
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
        for _ in range(ticks):
            env = make_envelope(
                "move_intent",
                session_id,
                state.last_tick,
                {"direction": {"x": 0, "y": 0}, "duration_ticks": 1},
            )
            await ws.send(json.dumps(env))
            await wait_for_accept(ws, state, env["message_id"], loop)
        return

    if action == "wait_until_assertion":
        assertion = step.get("assertion")
        if not isinstance(assertion, dict):
            raise AssertionError(f"wait_until_assertion: assertion must be object: {step}")
        timeout_s = float(step.get("timeout_s", SLICE_TIMEOUT_S))
        deadline = loop.time() + timeout_s
        last_error: AssertionError | None = None
        while loop.time() <= deadline:
            try:
                run_runtime_assertions([assertion], state, "runtime protocol")
                return
            except AssertionError as exc:
                last_error = exc
            env = make_envelope(
                "move_intent",
                session_id,
                state.last_tick,
                {"direction": {"x": 0, "y": 0}, "duration_ticks": 1},
            )
            await ws.send(json.dumps(env))
            await wait_for_accept(ws, state, env["message_id"], loop)
        detail = f": {last_error}" if last_error is not None else ""
        raise TimeoutError(f"wait_until_assertion timed out after {timeout_s:.1f}s{detail}")

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
        move_fn = move_to_position if bool(step.get("pathfind")) else walk_toward
        await move_fn(
            ws,
            session_id,
            state,
            {"x": float(step["x"]), "y": float(step["y"])},
            loop,
            stop_distance=float(step.get("tolerance", 0.25)),
            max_ticks=int(step.get("max_ticks", WALK_MAX_TICKS)),
        )
        return

    if action == "assert_player_position":
        assert_player_position(state, float(step["x"]), float(step["y"]), float(step.get("tolerance", 0.001)), "runtime protocol")
        return

    if action == "record_player_hp":
        player = find_player(state)
        if player is None or not isinstance(player.get("hp"), int):
            raise AssertionError(f"record_player_hp: missing player hp: {player}")
        state.recorded_player_hp = int(player["hp"])
        return

    if action == "assert_player_at_used_stair":
        direction = str(step["direction"])
        pos = state.used_stair_positions.get(direction)
        if pos is None:
            raise AssertionError(f"assert_player_at_used_stair: no recorded {direction} stair")
        deadline = loop.time() + float(step.get("wait_for_player_s", 3.0))
        last_error: AssertionError | None = None
        while loop.time() < deadline:
            try:
                assert_player_adjacent_to_position(
                    state,
                    float(pos["x"]),
                    float(pos["y"]),
                    float(step.get("tolerance", 0.001)),
                    float(step.get("max_distance", 1.415)),
                    "runtime protocol",
                )
                return
            except AssertionError as exc:
                last_error = exc
                await pump_one(ws, state, timeout=0.1)
        if last_error is not None:
            raise last_error
        return

    if action == "assert_teleporter_discovered":
        level = int(step["level"])
        want = bool(step.get("discovered", True))
        got = bool(state.discovered_teleporters.get(level, False))
        if got != want:
            raise AssertionError(f"assert_teleporter_discovered: level {level} discovered={got}, want {want}")
        return

    if action == "assert_inventory_contains":
        assert_inventory_contains(
            state.inventory,
            str(step["item_def_id"]),
            step.get("equipped"),
            "runtime protocol",
        )
        return

    if action == "assert_inventory_count":
        matches = state.inventory
        if step.get("item_def_id") is not None:
            matches = [item for item in matches if str(item.get("item_def_id", "")) == str(step["item_def_id"])]
        if step.get("item_template_id") is not None:
            matches = [item for item in matches if str(item.get("item_template_id", "")) == str(step["item_template_id"])]
        if step.get("equipped") is not None:
            matches = [item for item in matches if bool(item.get("equipped")) == bool(step["equipped"])]
        assert_count_matches(len(matches), step, "assert_inventory_count", f": {matches}")
        return

    if action == "assert_rolled_inventory_item":
        assert_rolled_inventory_item(state.inventory, step, "runtime protocol")
        return

    if action == "assert_inventory_requirement_status":
        assert_inventory_requirement_status(state.inventory, step, "runtime protocol")
        return

    if action == "assert_equipped_weapon_def":
        expected_def = str(step["item_def_id"])
        slot = str(step.get("slot", "main_hand"))
        weapon_id = state.equipped.get(slot)
        if weapon_id is None:
            raise AssertionError(f"assert_equipped_weapon_def: equipped {slot} is empty, want {expected_def}")
        item = next((i for i in state.inventory if str(i.get("item_instance_id")) == str(weapon_id)), None)
        if item is None or item.get("item_def_id") != expected_def:
            raise AssertionError(f"assert_equipped_weapon_def: {slot} {weapon_id} row={item}, want {expected_def}")
        return

    if action == "assert_equipped_slot_def":
        slot = str(step["slot"])
        expected_def = str(step["item_def_id"])
        item_id = state.equipped.get(slot)
        if item_id is None:
            raise AssertionError(f"assert_equipped_slot_def: equipped {slot} is empty, want {expected_def}")
        item = next((i for i in state.inventory if str(i.get("item_instance_id")) == str(item_id)), None)
        if item is None or item.get("item_def_id") != expected_def:
            raise AssertionError(f"assert_equipped_slot_def: {slot} {item_id} row={item}, want {expected_def}")
        return

    if action == "assert_equipped_slot_empty":
        slot = str(step["slot"])
        if state.equipped.get(slot) is not None:
            raise AssertionError(f"assert_equipped_slot_empty: {slot}={state.equipped.get(slot)}, want empty")
        return

    if action == "assert_hotbar_slot":
        assert_hotbar_slot(state.hotbar, int(step["slot_index"]), step.get("item_def_id"), "runtime protocol", state.inventory)
        return

    if action == "assert_hotbar_capacity":
        assert_count_matches(state.hotbar_capacity, step, "assert_hotbar_capacity")
        return

    if action == "assert_inventory_capacity":
        if "rows" in step:
            assert_count_matches(state.inventory_rows, {"equals": int(step["rows"])}, "assert_inventory_capacity rows")
        assert_count_matches(state.inventory_capacity, step, "assert_inventory_capacity")
        return

    if action == "assert_entity_count":
        assert_entity_count(list(state.entities.values()), step, "runtime protocol")
        return

    if action == "assert_player_at_discovered_teleporter":
        level = int(step.get("level", state.current_level))
        pos = state.used_teleporter_positions.get(level)
        if pos is None:
            raise AssertionError(f"assert_player_at_discovered_teleporter: no recorded teleporter for level {level}")
        assert_player_adjacent_to_position(
            state,
            float(pos["x"]),
            float(pos["y"]),
            float(step.get("tolerance", 0.001)),
            float(step.get("max_distance", 1.415)),
            "runtime protocol",
        )
        return

    if action == "attack_until_event":
        event_type = str(step["event_type"])
        await attack_until_monster_event(
            ws,
            session_id,
            state,
            loop,
            event_type,
            monster_def_id=str(step["monster_def_id"]) if step.get("monster_def_id") else None,
            rarity=str(step["rarity"]) if step.get("rarity") is not None else None,
            is_boss=bool(step["is_boss"]) if step.get("is_boss") is not None else None,
            target_id=str(step["target_id"]) if step.get("target_id") else None,
            timeout_s=float(step.get("timeout_s", SLICE_TIMEOUT_S)),
        )
        return

    if action == "directional_attack":
        direction = directional_attack_direction(state, step)
        env = make_envelope(
            "directional_attack_intent",
            session_id,
            state.last_tick,
            {"direction": direction},
        )
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

    if action == "action_until_event":
        event_type = str(step["event_type"])
        if step.get("monster_def_id") or step.get("target_id") or step.get("is_boss") is not None:
            await attack_until_monster_event(
                ws,
                session_id,
                state,
                loop,
                event_type,
                monster_def_id=str(step["monster_def_id"]) if step.get("monster_def_id") else None,
                rarity=str(step["rarity"]) if step.get("rarity") is not None else None,
                is_boss=bool(step["is_boss"]) if step.get("is_boss") is not None else None,
                target_id=str(step["target_id"]) if step.get("target_id") else None,
                timeout_s=float(step.get("timeout_s", SLICE_TIMEOUT_S)),
            )
            return
        target = resolve_target(state, step)
        target_id = str(target["id"])
        last_action = 0.0
        deadline = loop.time() + SLICE_TIMEOUT_S
        while event_type not in state.seen_events:
            if loop.time() > deadline:
                raise TimeoutError(f"action_until_event stalled waiting for {event_type}")
            if loop.time() - last_action > 0.12:
                env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": target_id})
                await ws.send(json.dumps(env))
                last_action = loop.time()
            await pump_one(ws, state, timeout=0.1)
        return

    if action == "action_until_combat_event":
        target = resolve_target(state, step)
        target_id = str(target["id"])
        monster_def_id = str(step.get("monster_def_id") or target.get("monster_def_id") or "")
        rarity = str(step["rarity"]) if step.get("rarity") is not None else None
        is_monster = target.get("type") == "monster"
        skipped_ids: set[str] = set()
        last_action = 0.0
        pending_message_id = ""
        start_index = len(state.combat_events)
        deadline = loop.time() + SLICE_TIMEOUT_S
        while not any(combat_event_matches(ev, step) for ev in state.combat_events[start_index:]):
            if loop.time() > deadline:
                raise TimeoutError(f"action_until_combat_event stalled waiting for {combat_event_summary(step)}")
            if is_monster and monster_def_id and target_id in skipped_ids:
                candidates = find_live_monsters_sorted(state, monster_def_id, rarity=rarity, exclude_ids=skipped_ids)
                if not candidates:
                    if skipped_ids:
                        skipped_ids.clear()
                        continue
                    raise AssertionError(f"action_until_combat_event: no live {monster_def_id} targets")
                target_id = str(candidates[0]["id"])
            if loop.time() - last_action > 0.12:
                current = state.entities.get(target_id)
                if current is not None and current.get("type") == "monster" and int(current.get("hp", 0)) <= 0:
                    break
                env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": target_id})
                pending_message_id = str(env["message_id"])
                await ws.send(json.dumps(env))
                last_action = loop.time()
            await pump_one(ws, state, timeout=0.1)
            if not pending_message_id:
                continue
            if pending_message_id in state.accepted_message_ids:
                pending_message_id = ""
                continue
            reason = state.rejected_message_reasons.pop(pending_message_id, None)
            if reason is None:
                continue
            pending_message_id = ""
            if is_monster and reason in {"no_path", "path_too_long"}:
                skipped_ids.add(target_id)
                continue
            if reason == "invalid_target" and monster_def_id and monster_def_id in state.killed_monster_def_ids:
                continue
            if reason == "projectile_busy":
                continue
            if reason == "player_dead":
                raise AssertionError("action_until_combat_event: player died")
            raise AssertionError(f"action_intent for {monster_def_id or target_id} was rejected: {reason}")
        if not any(combat_event_matches(ev, step) for ev in state.combat_events[start_index:]):
            raise AssertionError(f"action_until_combat_event target ended before {combat_event_summary(step)}")
        return

    if action == "wait_for_combat_event":
        start_index = len(state.combat_events)
        deadline = loop.time() + float(step.get("timeout_s", SLICE_TIMEOUT_S))
        while not any(combat_event_matches(ev, step) for ev in state.combat_events[start_index:]):
            if loop.time() > deadline:
                raise TimeoutError(f"wait_for_combat_event stalled waiting for {combat_event_summary(step)}")
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
            if target_item_def_id == "gold":
                await wait_for_event(ws, state, "gold_picked_up", loop)
                return
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
        step_timeout_s = float(step.get("timeout_s", SLICE_TIMEOUT_S * max(1, max_count)))
        killed = 0
        skipped_ids: set[str] = set()
        step_deadline = loop.time() + step_timeout_s
        for _ in range(5):
            if find_player(state) is not None:
                break
            await pump_one(ws, state, timeout=0.1)
        while killed < max_count:
            if loop.time() > step_deadline:
                raise TimeoutError(f"kill_monsters timed out after killing {killed}/{max_count} {monster_def_id}")
            if find_player(state) is None:
                await pump_one(ws, state, timeout=0.1)
                continue
            candidates = find_live_monsters_sorted(state, monster_def_id, exclude_ids=skipped_ids)
            if not candidates:
                if skipped_ids:
                    skipped_ids.clear()
                    continue
                if killed == 0:
                    raise AssertionError(f"kill_monsters: no live {monster_def_id} targets")
                return
            target = candidates[0]
            target_id = str(target["id"])
            target_deadline = loop.time() + SLICE_TIMEOUT_S
            last_action = 0.0
            pending_message_id = ""
            while True:
                current = state.entities.get(target_id)
                if current is None or int(current.get("hp", 0)) <= 0:
                    killed += 1
                    skipped_ids.discard(target_id)
                    break
                if loop.time() > target_deadline:
                    skipped_ids.add(target_id)
                    break
                if loop.time() - last_action > 0.12:
                    env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": target_id})
                    pending_message_id = str(env["message_id"])
                    await ws.send(json.dumps(env))
                    last_action = loop.time()
                await pump_one(ws, state, timeout=0.1)
                if not pending_message_id:
                    continue
                if pending_message_id in state.accepted_message_ids:
                    pending_message_id = ""
                    continue
                reason = state.rejected_message_reasons.pop(pending_message_id, None)
                if reason is None:
                    continue
                pending_message_id = ""
                if reason == "invalid_target" and monster_def_id in state.killed_monster_def_ids:
                    killed += 1
                    break
                if reason in {"no_path", "path_too_long"}:
                    skipped_ids.add(target_id)
                    live = find_live_monsters_sorted(state, monster_def_id)
                    if live and len(skipped_ids) >= len(live):
                        raise AssertionError(
                            f"kill_monsters: all live {monster_def_id} targets rejected {reason}; player="
                            f"{(find_player(state) or {}).get('position')}"
                        )
                    break
                if reason == "player_dead":
                    raise AssertionError(f"kill_monsters: player died after killing {killed}/{max_count}")
                raise AssertionError(f"action_intent for {monster_def_id} {target_id} was rejected: {reason}")
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
        deadline = loop.time() + float(step.get("wait_for_stair_s", 3.0))
        while target is None and loop.time() < deadline:
            await pump_one(ws, state, timeout=0.1)
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
            max_ticks=derived_walk_max_ticks(state, target["position"], int(step.get("max_ticks", WALK_MAX_TICKS))),
        )
        msg_type = "descend_intent" if direction == "down" else "ascend_intent"
        previous_level = state.current_level
        env = make_envelope(msg_type, session_id, state.last_tick, {})
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
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
            max_ticks=derived_walk_max_ticks(state, target["position"], int(step.get("max_ticks", WALK_MAX_TICKS))),
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
        if bool(step.get("pathfind")):
            await move_to_position(
                ws,
                session_id,
                state,
                monster["position"],
                loop,
                max_ticks=int(step.get("max_ticks", WALK_MAX_TICKS)),
                stop_distance=float(step.get("stop_distance", WALK_STOP_DISTANCE)),
            )
            return
        await walk_toward(
            ws,
            session_id,
            state,
            monster["position"],
            loop,
            max_ticks=int(step.get("max_ticks", WALK_MAX_TICKS)),
        )
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

    if action == "pick_up_first_rolled_loot":
        deadline = loop.time() + SLICE_TIMEOUT_S
        loot = find_first_rolled_loot(state)
        while loot is None:
            if loop.time() > deadline:
                raise TimeoutError("pick_up_first_rolled_loot stalled waiting for rolled loot")
            await pump_one(ws, state, timeout=0.1)
            loot = find_first_rolled_loot(state)
        loot_id = str(loot["id"])
        before_inventory_count = len(state.inventory)
        await ws.send(json.dumps(make_envelope(
            "action_intent", session_id, state.last_tick, {"target_id": loot_id})))
        log("picking up rolled loot", loot.get("item_template_id"), loot_id)
        while len(state.inventory) <= before_inventory_count:
            if loop.time() > deadline:
                raise TimeoutError("pick_up_first_rolled_loot stalled waiting for inventory_add")
            await pump_one(ws, state, timeout=0.1)
        state.item_id = str(state.inventory[-1].get("item_instance_id", state.item_id))
        return

    if action == "pick_up_loot":
        item_def_id = str(step["item_def_id"])
        bag_index = int(step.get("bag_index", 0))
        loot = find_loot(state, item_def_id)
        deadline = loop.time() + float(step.get("wait_for_loot_s", 3.0))
        while loot is None and loop.time() < deadline:
            await pump_one(ws, state, timeout=0.1)
            loot = find_loot(state, item_def_id)
        if loot is None:
            raise AssertionError(f"pick_up_loot: loot not found for item_def_id={item_def_id}")
        deadline = loop.time() + SLICE_TIMEOUT_S
        await walk_toward(
            ws,
            session_id,
            state,
            loot["position"],
            loop,
            max_ticks=int(step.get("max_ticks", WALK_MAX_TICKS)),
        )
        await ws.send(json.dumps(make_envelope(
            "action_intent", session_id, state.last_tick, {"target_id": loot["id"]})))
        log("picking up", item_def_id, "loot", loot["id"])
        if item_def_id == "gold":
            await wait_for_event(ws, state, "gold_picked_up", loop)
            return
        wait_started = loop.time()
        last_wait_log = wait_started
        while find_inventory_item(state.inventory, item_def_id, bag_index) is None:
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
        item = find_inventory_item(state.inventory, item_def_id, bag_index)
        if item is not None:
            state.item_id = item["item_instance_id"]
        return

    if action == "equip_first_inventory_item":
        deadline = loop.time() + SLICE_TIMEOUT_S
        while state.item_id is None:
            if loop.time() > deadline:
                raise TimeoutError("equip_first_inventory_item stalled waiting for inventory")
            await pump_one(ws, state, timeout=0.1)
        item = find_inventory_item_by_instance(state.inventory, state.item_id)
        slot = str(step.get("slot", (item or {}).get("slot", "main_hand")))
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
        bag_index = int(step.get("bag_index", 0))
        deadline = loop.time() + SLICE_TIMEOUT_S
        item = find_inventory_item(state.inventory, item_def_id, bag_index)
        while item is None:
            if loop.time() > deadline:
                raise TimeoutError(f"equip_inventory_item stalled waiting for {item_def_id}")
            await pump_one(ws, state, timeout=0.1)
            item = find_inventory_item(state.inventory, item_def_id, bag_index)
        slot = str(step.get("slot", item.get("slot", "main_hand")))
        item_id = str(item["item_instance_id"])
        env = make_envelope(
            "equip_intent",
            session_id,
            state.last_tick,
            {"item_instance_id": item_id, "slot": slot},
        )
        await ws.send(json.dumps(env))
        log("equipping", item_def_id, item_id)
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        while state.equipped.get(slot) != item_id:
            if loop.time() > deadline:
                raise TimeoutError(f"equip_inventory_item stalled waiting for equipped_update for {item_def_id}")
            await pump_one(ws, state, timeout=0.1)
        state.equipped_item_id = item_id
        return

    if action == "equip_last_inventory_item":
        deadline = loop.time() + SLICE_TIMEOUT_S
        while state.item_id is None:
            if loop.time() > deadline:
                raise TimeoutError("equip_last_inventory_item stalled waiting for inventory")
            await pump_one(ws, state, timeout=0.1)
        item = find_inventory_item_by_instance(state.inventory, state.item_id)
        if item is None:
            raise AssertionError(f"equip_last_inventory_item: missing inventory item {state.item_id}")
        slot = str(step.get("slot", item.get("slot", "main_hand")))
        await ws.send(json.dumps(make_envelope(
            "equip_intent", session_id, state.last_tick, {"item_instance_id": state.item_id, "slot": slot})))
        log("equipping last item", state.item_id, "slot", slot)
        while state.equipped.get(slot) != state.item_id:
            if loop.time() > deadline:
                raise TimeoutError(f"equip_last_inventory_item stalled waiting for equipped_update for {state.item_id}")
            await pump_one(ws, state, timeout=0.1)
        state.equipped_item_id = state.item_id
        return


    if action == "unequip_slot":
        slot = str(step.get("slot", "main_hand"))
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

    if action == "assign_hotbar":
        slot_index = int(step["slot_index"])
        item_def_id = step.get("item_def_id")
        item_id: str | None = None
        if item_def_id is not None:
            item = find_inventory_item(state.inventory, str(item_def_id))
            if item is None:
                raise AssertionError(f"assign_hotbar: missing inventory item {item_def_id}")
            item_id = str(item["item_instance_id"])
        await ws.send(json.dumps(make_envelope(
            "assign_hotbar_intent",
            session_id,
            state.last_tick,
            {"slot_index": slot_index, "item_instance_id": item_id},
        )))
        deadline = loop.time() + SLICE_TIMEOUT_S
        while hotbar_item_id(state.hotbar, slot_index) != item_id:
            if loop.time() > deadline:
                raise TimeoutError(f"assign_hotbar stalled waiting for slot {slot_index}={item_id}")
            await pump_one(ws, state, timeout=0.1)
        return

    if action == "use_hotbar_slot":
        slot_index = int(step["slot_index"])
        env = make_envelope(
            "use_hotbar_intent",
            session_id,
            state.last_tick,
            {"slot_index": slot_index},
        )
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        deadline = loop.time() + SLICE_TIMEOUT_S
        if expect_reject:
            await wait_for_reject(ws, state, str(env["message_id"]), str(expect_reject), loop)
            return
        while hotbar_item_id(state.hotbar, slot_index) is not None:
            if loop.time() > deadline:
                raise TimeoutError(f"use_hotbar_slot stalled waiting for slot {slot_index} clear")
            await pump_one(ws, state, timeout=0.1)
        return

    if action == "assert_player_hp":
        player = find_player(state)
        if player is None:
            raise AssertionError("assert_player_hp: player not found in runtime state")
        assert_count_matches(int(player.get("hp", -1)), step, "assert_player_hp")
        return

    if action == "allocate_stat":
        stat = str(step["stat"])
        points = int(step.get("points", 1))
        env = make_envelope(
            "allocate_stat_intent",
            session_id,
            state.last_tick,
            {"stat": stat, "points": points},
        )
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        expected = step.get("expect_progression")
        if isinstance(expected, dict):
            await wait_for_character_progression(ws, state, expected, loop)
        return

    if action == "assert_character_progression":
        assert_character_progression(state.character_progression, step, "runtime protocol")
        return

    if action == "open_shop":
        shop_id = str(step.get("shop_id", "town_vendor"))
        interactable_def_id = str(step.get("interactable_def_id", "town_vendor"))
        target = find_interactable(state, interactable_def_id)
        if target is None:
            raise AssertionError(f"open_shop: missing {interactable_def_id} on level {state.current_level}")
        await walk_toward(
            ws,
            session_id,
            state,
            target["position"],
            loop,
            stop_distance=float(step.get("stop_distance", WALK_STOP_DISTANCE)),
            max_ticks=derived_walk_max_ticks(state, target["position"], int(step.get("max_ticks", WALK_MAX_TICKS))),
        )
        start_index = len(state.shop_events)
        env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": str(target["id"])})
        await ws.send(json.dumps(env))
        await wait_for_accept(ws, state, env["message_id"], loop)
        await wait_for_shop_event(ws, state, "shop_opened", loop, shop_id=shop_id, start_index=start_index)
        return

    if action == "assert_shop_offer_count":
        offers = filtered_shop_offers(state, step)
        assert_count_matches(len(offers), step, "assert_shop_offer_count", f": {offers}")
        return

    if action == "assert_shop_offer_details":
        offers = filtered_shop_offers(state, step)
        assert_shop_detail_rows(offers, step, "assert_shop_offer_details")
        return

    if action == "assert_shop_sell_appraisal_count":
        rows = filtered_shop_sell_appraisals(state, step)
        assert_count_matches(len(rows), step, "assert_shop_sell_appraisal_count", f": {rows}")
        return

    if action == "assert_shop_sell_appraisal_details":
        rows = filtered_shop_sell_appraisals(state, step)
        step = dict(step)
        step.setdefault("price_key", "sell_price")
        assert_shop_detail_rows(rows, step, "assert_shop_sell_appraisal_details")
        return

    if action == "buy_shop_offer":
        shop_id = str(step.get("shop_id", "town_vendor"))
        target = find_interactable(state, str(step.get("interactable_def_id", "town_vendor")))
        if target is None:
            raise AssertionError(f"buy_shop_offer: missing shop entity on level {state.current_level}")
        offer = select_shop_offer(state, step)
        before_gold = state.gold
        before_inventory_count = len(state.inventory)
        start_index = len(state.shop_events)
        env = make_envelope(
            "shop_buy_intent",
            session_id,
            state.last_tick,
            {"shop_entity_id": str(target["id"]), "offer_id": str(offer["offer_id"])},
        )
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        event = await wait_for_shop_event(ws, state, "shop_purchase", loop, shop_id=shop_id, start_index=start_index)
        state.last_gold_before_action = before_gold
        state.last_gold_after_action = state.gold
        if int(event.get("price", 0)) != int(offer.get("buy_price", 0)):
            raise AssertionError(f"buy_shop_offer: event price {event.get('price')} != offer price {offer.get('buy_price')}")
        if len(state.inventory) <= before_inventory_count:
            raise AssertionError(f"buy_shop_offer: inventory did not grow after {offer}")
        return

    if action == "sell_inventory_item":
        shop_id = str(step.get("shop_id", "town_vendor"))
        target = find_interactable(state, str(step.get("interactable_def_id", "town_vendor")))
        if target is None:
            raise AssertionError(f"sell_inventory_item: missing shop entity on level {state.current_level}")
        candidates = state.inventory
        if step.get("item_def_id") is not None:
            candidates = [item for item in candidates if str(item.get("item_def_id", "")) == str(step["item_def_id"])]
        if step.get("item_template_id") is not None:
            candidates = [item for item in candidates if str(item.get("item_template_id", "")) == str(step["item_template_id"])]
        if step.get("rolled") is not None:
            want_rolled = bool(step["rolled"])
            candidates = [item for item in candidates if bool(item.get("item_template_id")) == want_rolled]
        if step.get("equipped") is not None:
            candidates = [item for item in candidates if bool(item.get("equipped")) == bool(step["equipped"])]
        candidates = sorted(candidates, key=lambda item: str(item.get("item_instance_id", "")))
        if not candidates:
            raise AssertionError(f"sell_inventory_item: no matching inventory item for {step}; inventory={state.inventory}")
        item = candidates[int(step.get("bag_index", 0))]
        item_id = str(item["item_instance_id"])
        before_gold = state.gold
        start_index = len(state.shop_events)
        env = make_envelope(
            "shop_sell_intent",
            session_id,
            state.last_tick,
            {"shop_entity_id": str(target["id"]), "item_instance_id": item_id},
        )
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        event = await wait_for_shop_event(ws, state, "shop_sale", loop, shop_id=shop_id, start_index=start_index)
        state.last_gold_before_action = before_gold
        state.last_gold_after_action = state.gold
        if str(event.get("item_instance_id")) != item_id:
            raise AssertionError(f"sell_inventory_item: event item {event.get('item_instance_id')} != {item_id}")
        if find_inventory_item_by_instance(state.inventory, item_id) is not None:
            raise AssertionError(f"sell_inventory_item: item {item_id} still in inventory")
        return

    if action == "assert_gold_changed":
        before = state.last_gold_before_action
        after = state.last_gold_after_action
        if before is None or after is None:
            raise AssertionError("assert_gold_changed: no recorded shop gold change")
        direction = str(step.get("direction", ""))
        if direction == "decrease" and not after < before:
            raise AssertionError(f"assert_gold_changed: gold {after} did not decrease from {before}")
        if direction == "increase" and not after > before:
            raise AssertionError(f"assert_gold_changed: gold {after} did not increase from {before}")
        if "by_at_least" in step and abs(after - before) < int(step["by_at_least"]):
            raise AssertionError(f"assert_gold_changed: |{after} - {before}| < {step['by_at_least']}")
        return

    if action == "assert_deepest_dungeon_depth":
        got = int(state.character_progression.get("deepest_dungeon_depth", 0))
        assert_count_matches(got, step, "assert_deepest_dungeon_depth")
        return

    if action == "assert_shop_event":
        event_type = str(step["event_type"])
        shop_id = str(step.get("shop_id", "town_vendor"))
        matches = [
            event for event in state.shop_events
            if event.get("event_type") == event_type and str(event.get("shop_id", "")) == shop_id
        ]
        if step.get("offer_id") is not None:
            matches = [event for event in matches if str(event.get("offer_id", "")) == str(step["offer_id"])]
        assert_count_matches(len(matches), step, "assert_shop_event", f": {matches}")
        return

    if action == "use_inventory_item":
        item_def_id = str(step["item_def_id"])
        bag_index = int(step.get("bag_index", 0))
        deadline = loop.time() + SLICE_TIMEOUT_S
        item = find_inventory_item(state.inventory, item_def_id, bag_index)
        if item is None:
            raise AssertionError(f"use_inventory_item: missing inventory item {item_def_id} at bag_index={bag_index}")
        item_id = str(item["item_instance_id"])
        env = make_envelope("use_intent", session_id, state.last_tick, {"item_instance_id": item_id})
        await ws.send(json.dumps(env))
        log("using", item_def_id, item_id)
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
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
    if state.world_id == "dungeon_levels" and state.current_level < 0:
        await move_to_position(ws, session_id, state, target_pos, loop, max_ticks=max_ticks, stop_distance=stop_distance)
        return
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


def derived_walk_max_ticks(state: RuntimeState, target_pos: dict[str, Any], requested: int) -> int:
    player = find_player(state)
    if player is None:
        return requested
    player_pos = player.get("position", {})
    dx = abs(float(target_pos.get("x", 0.0)) - float(player_pos.get("x", 0.0)))
    dy = abs(float(target_pos.get("y", 0.0)) - float(player_pos.get("y", 0.0)))
    distance_ticks = int(max(dx, dy) * 20) + 160
    return max(requested, distance_ticks, WALK_MAX_TICKS)


async def move_to_position(
    ws,
    session_id: str,
    state: RuntimeState,
    target_pos: dict[str, Any],
    loop,
    max_ticks: int = WALK_MAX_TICKS,
    stop_distance: float = WALK_STOP_DISTANCE,
) -> None:
    player = find_player(state)
    if player is None:
        raise AssertionError("move_to_position: player not found")
    player_pos = player["position"]
    dx = float(target_pos["x"]) - float(player_pos["x"])
    dy = float(target_pos["y"]) - float(player_pos["y"])
    if max(abs(dx), abs(dy)) <= stop_distance:
        return

    env = make_envelope("move_to_intent", session_id, state.last_tick, {"position": target_pos})
    await ws.send(json.dumps(env))
    before = {"x": player_pos["x"], "y": player_pos["y"]}
    await wait_for_player_move_or_accept(ws, state, before, env["message_id"], loop)
    unchanged_ticks = 0
    stalled_reissues = 0
    last_pos = before
    for _ in range(max_ticks):
        player = find_player(state)
        if player is None:
            raise AssertionError("move_to_position: player not found")
        player_pos = player["position"]
        dx = float(target_pos["x"]) - float(player_pos["x"])
        dy = float(target_pos["y"]) - float(player_pos["y"])
        if max(abs(dx), abs(dy)) <= stop_distance:
            return
        current_pos = {"x": player_pos["x"], "y": player_pos["y"]}
        if current_pos == last_pos:
            unchanged_ticks += 1
            if unchanged_ticks >= 120:
                stalled_reissues += 1
                if stalled_reissues > 3:
                    raise TimeoutError(
                        f"move_to_position made no progress after {stalled_reissues} move_to_intent attempts "
                        f"toward {target_pos}; player={current_pos}"
                    )
                env = make_envelope("move_to_intent", session_id, state.last_tick, {"position": target_pos})
                await ws.send(json.dumps(env))
                await wait_for_player_move_or_accept(ws, state, current_pos, env["message_id"], loop)
                unchanged_ticks = 0
                last_pos = current_pos
        else:
            unchanged_ticks = 0
            stalled_reissues = 0
            last_pos = current_pos
        await pump_one(ws, state, timeout=0.05)
    player = find_player(state)
    player_pos = (player or {}).get("position")
    raise TimeoutError(f"move_to_position exhausted {max_ticks} ticks toward {target_pos}; player={player_pos}")


async def attack_until_monster_event(
    ws,
    session_id: str,
    state: RuntimeState,
    loop,
    event_type: str,
    *,
    monster_def_id: str | None = None,
    rarity: str | None = None,
    is_boss: bool | None = None,
    target_id: str | None = None,
    timeout_s: float = SLICE_TIMEOUT_S,
) -> None:
    deadline = loop.time() + timeout_s
    skipped_ids: set[str] = set()
    active_target_id = target_id
    pending_message_id = ""
    last_action = 0.0
    for _ in range(5):
        if find_player(state) is not None:
            break
        await pump_one(ws, state, timeout=0.1)
    while event_type not in state.seen_events:
        if loop.time() > deadline:
            raise TimeoutError(f"attack_until_monster_event stalled waiting for {event_type}")
        if (monster_def_id or is_boss is not None) and (active_target_id is None or active_target_id in skipped_ids):
            candidates = find_live_monsters_sorted(
                state,
                monster_def_id or "",
                rarity=rarity,
                is_boss=is_boss,
                exclude_ids=skipped_ids,
            )
            if not candidates:
                if skipped_ids:
                    skipped_ids.clear()
                    continue
                raise AssertionError(f"attack_until_monster_event: no live {monster_def_id or 'boss'} targets")
            active_target_id = str(candidates[0]["id"])
        if active_target_id is None:
            raise AssertionError("attack_until_monster_event: no target")
        current = state.entities.get(active_target_id)
        if current is not None and current.get("type") == "monster" and int(current.get("hp", 0)) <= 0:
            await pump_one(ws, state, timeout=0.1)
            continue
        if loop.time() - last_action > 0.12:
            env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": active_target_id})
            pending_message_id = str(env["message_id"])
            await ws.send(json.dumps(env))
            last_action = loop.time()
        await pump_one(ws, state, timeout=0.1)
        if not pending_message_id:
            continue
        if pending_message_id in state.accepted_message_ids:
            pending_message_id = ""
            continue
        reason = state.rejected_message_reasons.pop(pending_message_id, None)
        if reason is None:
            continue
        pending_message_id = ""
        if reason in {"no_path", "path_too_long"}:
            skipped_ids.add(active_target_id)
            active_target_id = None
            continue
        if reason == "invalid_target" and monster_def_id and monster_def_id in state.killed_monster_def_ids:
            continue
        if reason == "player_dead":
            raise AssertionError(f"attack_until_monster_event: player died waiting for {event_type}")
        label = monster_def_id or active_target_id
        raise AssertionError(f"action_intent for {label} was rejected: {reason}")


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
        reason = state.rejected_message_reasons.get(message_id)
        if reason is not None:
            raise AssertionError(f"intent {message_id} rejected while waiting for accept: {reason}")
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


async def wait_for_shop_event(
    ws,
    state: RuntimeState,
    event_type: str,
    loop,
    *,
    shop_id: str = "town_vendor",
    start_index: int = 0,
) -> dict[str, Any]:
    deadline = loop.time() + SLICE_TIMEOUT_S
    while True:
        for event in state.shop_events[start_index:]:
            if event.get("event_type") == event_type and str(event.get("shop_id", "")) == shop_id:
                return event
        if loop.time() > deadline:
            raise TimeoutError(f"stalled waiting for shop event {event_type} shop_id={shop_id}")
        await pump_one(ws, state, timeout=0.1)


def combat_event_matches(event: dict[str, Any], expected: dict[str, Any]) -> bool:
    for key in (
        "event_type",
        "outcome",
        "damage",
        "raw_damage",
        "mitigated_damage",
        "blocked",
        "critical",
        "source_entity_id",
        "target_entity_id",
        "entity_id",
    ):
        if key in expected and event.get(key) != expected[key]:
            return False
    for key, event_key in (
        ("min_damage", "damage"),
        ("min_raw_damage", "raw_damage"),
        ("min_mitigated_damage", "mitigated_damage"),
    ):
        if key in expected and int(event.get(event_key, -999999)) < int(expected[key]):
            return False
    return True


def combat_event_summary(expected: dict[str, Any]) -> str:
    parts = []
    for key in ("event_type", "outcome", "damage", "min_damage", "blocked", "critical"):
        if key in expected:
            parts.append(f"{key}={expected[key]}")
    return ", ".join(parts) or str(expected)


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


async def wait_for_character_progression(ws, state: RuntimeState, expected: dict[str, Any], loop) -> None:
    deadline = loop.time() + SLICE_TIMEOUT_S
    while True:
        try:
            assert_character_progression(state.character_progression, expected, "runtime protocol")
            return
        except AssertionError:
            pass
        if loop.time() > deadline:
            assert_character_progression(state.character_progression, expected, "runtime protocol")
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
            if reason == "invalid_target" and monster_def_id in state.killed_monster_def_ids:
                return
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
        if event_type in {"shop_opened", "shop_purchase", "shop_sale"}:
            shop_event = dict(ev)
            state.shop_events.append(shop_event)
            state.last_shop_event = shop_event
            shop_id = str(ev.get("shop_id", ""))
            if event_type == "shop_opened" and shop_id:
                state.shop_offers[shop_id] = {
                    str(offer["offer_id"]): dict(offer)
                    for offer in ev.get("offers", [])
                }
                state.shop_sell_appraisals[shop_id] = {
                    str(appraisal["item_instance_id"]): dict(appraisal)
                    for appraisal in ev.get("sell_appraisals", [])
                }
        if event_type in {"monster_damaged", "player_damaged", "player_killed", "attack_missed", "attack_blocked"}:
            state.combat_events.append(dict(ev))
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
        if event_type == "monster_damaged":
            entity = state.entities.get(str(ev.get("entity_id")))
            if entity is not None and entity.get("monster_def_id") and isinstance(ev.get("damage"), int):
                monster_def_id = str(entity["monster_def_id"])
                state.max_monster_damage_by_def[monster_def_id] = max(
                    state.max_monster_damage_by_def.get(monster_def_id, 0),
                    int(ev["damage"]),
                )
    apply_level_entities = delta_level == state.current_level
    for c in (p.get("changes") or []):
        if c["op"] == "wall_layout_update":
            if apply_level_entities:
                state.walls = [dict(wall) for wall in c.get("walls", [])]
        elif c["op"] in {"entity_spawn", "entity_update"}:
            if not apply_level_entities:
                continue
            entity = c["entity"]
            existing = state.entities.get(entity["id"], {})
            existing.update(entity)
            state.entities[entity["id"]] = existing
            track_initial_entity_position(state, existing)
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
        elif c["op"] == "equipped_update":
            state.equipped_item_id = c.get("item_instance_id")
            state.equipped[c["slot"]] = c.get("item_instance_id")
            if "hotbar_capacity" in c:
                state.hotbar_capacity = int(c["hotbar_capacity"])
            if "inventory_rows" in c:
                state.inventory_rows = int(c["inventory_rows"])
            if "inventory_capacity" in c:
                state.inventory_capacity = int(c["inventory_capacity"])
        elif c["op"] == "hotbar_update":
            upsert_hotbar(state, int(c["slot_index"]), c.get("item_instance_id"))
            if "inventory_rows" in c:
                state.inventory_rows = int(c["inventory_rows"])
            if "inventory_capacity" in c:
                state.inventory_capacity = int(c["inventory_capacity"])
        elif c["op"] == "gold_update":
            state.gold = int(c.get("gold", state.gold))
        elif c["op"] == "teleporter_discovery_update":
            state.discovered_teleporters[int(c["level"])] = bool(c["discovered"])
        elif c["op"] == "character_progression_update":
            state.character_progression = dict(c.get("character_progression") or {})
            if "gold" in state.character_progression:
                state.gold = int(state.character_progression["gold"])
    if state.pending_level_load is not None and delta_level == state.pending_level_load and find_player(state) is not None:
        state.pending_level_load = None
    update_runtime_distances(state)


def ingest_snapshot(payload: dict[str, Any], state: RuntimeState) -> None:
    state.last_tick = max(state.last_tick, int(payload.get("server_tick", 0)))
    state.current_level = int(payload.get("current_level", 0))
    state.local_player_id = str(payload.get("local_player_id", state.local_player_id))
    state.party = [dict(row) for row in payload.get("party", [])]
    state.visited_levels.add(state.current_level)
    state.last_delta_level = state.current_level
    state.pending_level_load = None
    state.walls = [dict(wall) for wall in payload.get("walls", [])]
    state.entities = {str(e["id"]): dict(e) for e in payload.get("entities", [])}
    state.inventory = [dict(i) for i in payload.get("inventory", [])]
    state.equipped = dict(payload.get("equipped", {}))
    state.hotbar_capacity = int(payload.get("hotbar_capacity", 2))
    state.hotbar = [dict(slot) for slot in payload.get("hotbar", [])]
    state.inventory_rows = int(payload.get("inventory_rows", 3))
    state.inventory_capacity = int(payload.get("inventory_capacity", state.inventory_rows * 5))
    state.character_progression = dict(payload.get("character_progression", {}))
    state.gold = int(payload.get("gold", state.character_progression.get("gold", 0)))
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
        track_initial_entity_position(state, entity)
        track_initial_monster_position(state, entity)
    update_runtime_distances(state)


def parse_discovered_teleporters(payload: dict[str, Any]) -> dict[int, bool]:
    return {
        int(row["level"]): bool(row["discovered"])
        for row in payload.get("discovered_teleporters", [])
    }


def upsert_hotbar(state: RuntimeState, slot_index: int, item_instance_id: Any) -> None:
    while len(state.hotbar) <= slot_index:
        state.hotbar.append({"slot_index": len(state.hotbar), "item_instance_id": None})
    state.hotbar[slot_index] = {"slot_index": slot_index, "item_instance_id": item_instance_id}


def hotbar_item_id(hotbar: list[dict[str, Any]], slot_index: int) -> str | None:
    for slot in hotbar:
        if int(slot.get("slot_index", -1)) == slot_index:
            raw = slot.get("item_instance_id")
            return None if raw is None else str(raw)
    return None


def clear_active_level_state(state: RuntimeState) -> None:
    state.walls.clear()
    state.entities.clear()
    state.loot_ids.clear()
    state.min_player_monster_distance.clear()
    state.initial_monster_positions.clear()
    state.initial_entity_positions.clear()


def track_initial_entity_position(state: RuntimeState, entity: dict[str, Any]) -> None:
    entity_id = str(entity.get("id", ""))
    if not entity_id or entity_id in state.initial_entity_positions:
        return
    pos = entity.get("position")
    if not isinstance(pos, dict):
        return
    state.initial_entity_positions[entity_id] = {
        "x": float(pos.get("x", 0.0)),
        "y": float(pos.get("y", 0.0)),
    }


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


def find_first_rolled_loot(state: RuntimeState) -> dict[str, Any] | None:
    for loot_id in state.loot_ids:
        entity = state.entities.get(loot_id)
        if entity is not None and entity.get("type") == "loot" and entity.get("item_template_id"):
            return entity
    return None


def find_monster(state: RuntimeState, monster_def_id: str, rarity: str | None = None) -> dict[str, Any] | None:
    for entity in state.entities.values():
        if (
            entity.get("type") == "monster"
            and (not monster_def_id or entity.get("monster_def_id") == monster_def_id)
            and (rarity is None or entity.get("rarity") == rarity)
            and int(entity.get("hp", 0)) > 0
        ):
            return entity
    return None


def find_nearest_monster(
    state: RuntimeState,
    monster_def_id: str,
    rarity: str | None = None,
    is_boss: bool | None = None,
    exclude_ids: set[str] | None = None,
) -> dict[str, Any] | None:
    monsters = find_live_monsters_sorted(state, monster_def_id, rarity=rarity, is_boss=is_boss, exclude_ids=exclude_ids)
    if not monsters:
        return None
    return monsters[0]


def find_live_monsters_sorted(
    state: RuntimeState,
    monster_def_id: str,
    rarity: str | None = None,
    is_boss: bool | None = None,
    exclude_ids: set[str] | None = None,
) -> list[dict[str, Any]]:
    excluded = exclude_ids or set()
    player = find_player(state)
    if player is None:
        return []
    player_pos = player.get("position", {})
    player_x = float(player_pos.get("x", 0.0))
    player_y = float(player_pos.get("y", 0.0))
    ranked: list[tuple[float, str, dict[str, Any]]] = []
    for entity in state.entities.values():
        entity_id = str(entity.get("id", ""))
        if (
            entity.get("type") != "monster"
            or (monster_def_id and entity.get("monster_def_id") != monster_def_id)
            or (rarity is not None and entity.get("rarity") != rarity)
            or (is_boss is not None and bool(entity.get("is_boss", False)) != is_boss)
            or int(entity.get("hp", 0)) <= 0
            or entity_id in excluded
        ):
            continue
        pos = entity.get("position", {})
        distance = max(abs(float(pos.get("x", 0.0)) - player_x), abs(float(pos.get("y", 0.0)) - player_y))
        ranked.append((distance, entity_id, entity))
    ranked.sort(key=lambda item: (item[0], item[1]))
    return [entity for _, _, entity in ranked]


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
        rarity = str(step["rarity"]) if step.get("rarity") is not None else None
        target = find_nearest_monster(state, str(step["monster_def_id"]), rarity, bool(step["is_boss"]) if step.get("is_boss") is not None else None)
        if target is None:
            raise AssertionError(f"{step.get('action')}: monster not found: {step}")
        return target
    if step.get("is_boss") is not None:
        target = find_nearest_monster(state, "", None, bool(step["is_boss"]))
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


def directional_attack_direction(state: RuntimeState, step: dict[str, Any]) -> dict[str, float]:
    if isinstance(step.get("direction"), dict):
        raw = step["direction"]
        return {"x": float(raw.get("x", 0.0)), "y": float(raw.get("y", 0.0))}
    target = resolve_target(state, step)
    player = find_player(state)
    if player is None:
        raise AssertionError("directional_attack: missing player")
    ppos = player.get("position", {})
    tpos = target.get("position", {})
    dx = float(tpos.get("x", 0.0)) - float(ppos.get("x", 0.0))
    dy = float(tpos.get("y", 0.0)) - float(ppos.get("y", 0.0))
    length = (dx * dx + dy * dy) ** 0.5
    if length <= 0.000001:
        raise AssertionError(f"directional_attack: target overlaps player, no direction: {target}")
    return {"x": dx / length, "y": dy / length}


def find_inventory_item(
    inventory: list[dict[str, Any]],
    item_def_id: str,
    bag_index: int = 0,
) -> dict[str, Any] | None:
    matches = [item for item in inventory if item.get("item_def_id") == item_def_id]
    if bag_index < 0 or bag_index >= len(matches):
        return None
    return matches[bag_index]


def find_inventory_item_by_instance(inventory: list[dict[str, Any]], item_instance_id: str | None) -> dict[str, Any] | None:
    if item_instance_id is None:
        return None
    return next((item for item in inventory if str(item.get("item_instance_id")) == str(item_instance_id)), None)


def filtered_shop_offers(state: RuntimeState, step: dict[str, Any]) -> list[dict[str, Any]]:
    shop_id = str(step.get("shop_id", "town_vendor"))
    offers = list(state.shop_offers.get(shop_id, {}).values())
    if step.get("offer_id") is not None:
        offers = [offer for offer in offers if str(offer.get("offer_id")) == str(step["offer_id"])]
    if step.get("offer_kind") is not None:
        offers = [offer for offer in offers if str(offer.get("kind")) == str(step["offer_kind"])]
    if step.get("item_def_id") is not None:
        offers = [offer for offer in offers if str(offer.get("item_def_id")) == str(step["item_def_id"])]
    if step.get("item_template_id") is not None:
        offers = [offer for offer in offers if str(offer.get("item_template_id")) == str(step["item_template_id"])]
    if bool(step.get("affordable")):
        reserve_gold = int(step.get("reserve_gold", 0))
        budget = max(0, state.gold - reserve_gold)
        offers = [offer for offer in offers if int(offer.get("buy_price", 0)) <= budget]
    offers.sort(key=lambda offer: (int(offer.get("buy_price", 0)), str(offer.get("offer_id", ""))))
    return offers


def filtered_shop_sell_appraisals(state: RuntimeState, step: dict[str, Any]) -> list[dict[str, Any]]:
    shop_id = str(step.get("shop_id", "town_vendor"))
    rows = list(state.shop_sell_appraisals.get(shop_id, {}).values())
    if step.get("item_instance_id") is not None:
        rows = [row for row in rows if str(row.get("item_instance_id")) == str(step["item_instance_id"])]
    if step.get("item_def_id") is not None:
        rows = [row for row in rows if str(row.get("item_def_id")) == str(step["item_def_id"])]
    if step.get("item_template_id") is not None:
        rows = [row for row in rows if str(row.get("item_template_id")) == str(step["item_template_id"])]
    rows.sort(key=lambda row: (int(row.get("sell_price", 0)), str(row.get("item_instance_id", ""))))
    return rows


def select_shop_offer(state: RuntimeState, step: dict[str, Any]) -> dict[str, Any]:
    offers = filtered_shop_offers(state, step)
    if not offers:
        shop_id = str(step.get("shop_id", "town_vendor"))
        raise AssertionError(f"{step.get('action')}: no matching shop offers in {shop_id}: {step}; have={state.shop_offers.get(shop_id, {})}")
    index = int(step.get("offer_index", 0))
    if index < 0 or index >= len(offers):
        raise AssertionError(f"{step.get('action')}: offer_index {index} out of range for {offers}")
    return offers[index]


def assert_shop_detail_rows(rows: list[dict[str, Any]], step: dict[str, Any], label: str) -> None:
    assert_count_matches(len(rows), step, label, f": {rows}")
    if not rows:
        return
    if step.get("requires_summary", True):
        missing = [row for row in rows if not row.get("summary_lines")]
        if missing:
            raise AssertionError(f"{label}: rows missing summary_lines: {missing}")
    if step.get("requires_price", True):
        price_key = str(step.get("price_key", "buy_price"))
        missing = [row for row in rows if int(row.get(price_key, 0)) <= 0]
        if missing:
            raise AssertionError(f"{label}: rows missing positive {price_key}: {missing}")
    if step.get("requires_slot"):
        missing = [row for row in rows if not row.get("slot")]
        if missing:
            raise AssertionError(f"{label}: rows missing slot: {missing}")
    if step.get("requires_category"):
        missing = [row for row in rows if not row.get("category")]
        if missing:
            raise AssertionError(f"{label}: rows missing category: {missing}")
    if step.get("requires_comparison"):
        missing = [row for row in rows if not row.get("comparison", {}).get("deltas")]
        if missing:
            raise AssertionError(f"{label}: rows missing comparison deltas: {missing}")
    if step.get("requires_requirement_status"):
        missing = [row for row in rows if not row.get("requirement_status") or "requirements_met" not in row]
        if missing:
            raise AssertionError(f"{label}: rows missing requirement status: {missing}")
    if step.get("requires_equip_preview"):
        missing = [
            row for row in rows
            if not isinstance(row.get("equip_preview"), dict) or "requirements_met" not in row.get("equip_preview", {})
        ]
        if missing:
            raise AssertionError(f"{label}: rows missing equip preview: {missing}")
    if step.get("summary_contains") is not None:
        needle = str(step["summary_contains"])
        if not any(needle in str(line) for row in rows for line in row.get("summary_lines", [])):
            raise AssertionError(f"{label}: no summary line contains {needle!r}: {rows}")


def find_player(state: RuntimeState) -> dict[str, Any] | None:
    if state.local_player_id:
        player = state.entities.get(state.local_player_id)
        if player is not None and player.get("type") == "player":
            return player
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
            walls=payload.get("walls", []),
            discovered_teleporters=parse_discovered_teleporters(payload),
            character_progression=payload.get("character_progression", {}),
            hotbar_capacity=int(payload.get("hotbar_capacity", 2)),
            hotbar=payload.get("hotbar", []),
            inventory_rows=int(payload.get("inventory_rows", 3)),
            inventory_capacity=int(payload.get("inventory_capacity", int(payload.get("inventory_rows", 3)) * 5)),
            gold=int(payload.get("gold", 0)),
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
    if equipped.get("main_hand") != expected_id:
        raise AssertionError(f"{where}: equipped main_hand {equipped.get('main_hand')} != {item_id}")


def assert_player_hp_equals(entities: list[dict], want: int, where: str) -> None:
    player = find_player_entities(entities)
    if player is None:
        raise AssertionError(f"{where}: player not found")
    hp = player.get("hp")
    if hp != want:
        raise AssertionError(f"{where}: player hp {hp} != {want}")


def assert_player_max_hp_equals(entities: list[dict], want: int, where: str) -> None:
    player = find_player_entities(entities)
    if player is None:
        raise AssertionError(f"{where}: player not found")
    max_hp = player.get("max_hp")
    if max_hp != want:
        raise AssertionError(f"{where}: player max_hp {max_hp} != {want}")


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


def assert_player_adjacent_to_position(state: RuntimeState, x: float, y: float, tolerance: float, max_distance: float, where: str) -> None:
    player = find_player(state)
    if player is None:
        raise AssertionError(f"{where}: missing player entity")
    pos = player.get("position", {})
    got_x = float(pos.get("x", "nan"))
    got_y = float(pos.get("y", "nan"))
    dist = math.hypot(got_x - x, got_y - y)
    if dist <= tolerance:
        raise AssertionError(f"{where}: player position ({got_x}, {got_y}) is on marker ({x}, {y})")
    if dist > max_distance:
        raise AssertionError(f"{where}: player position ({got_x}, {got_y}) is not adjacent to marker ({x}, {y}); distance={dist}")


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


def assert_inventory_requirement_status(inventory: list[dict], assertion: dict[str, Any], where: str) -> None:
    item_def_id = str(assertion["item_def_id"])
    bag_index = int(assertion.get("bag_index", 0))
    item = find_inventory_item(inventory, item_def_id, bag_index)
    if item is None:
        raise AssertionError(f"{where}: missing inventory item {item_def_id}: {inventory}")
    if assertion.get("equipped") is not None and bool(item.get("equipped")) != bool(assertion["equipped"]):
        raise AssertionError(f"{where}: inventory {item_def_id} equipped={item.get('equipped')} want {assertion['equipped']}: {item}")
    assert_requirement_payload(item, assertion, f"{where}: inventory {item_def_id}")


def assert_loot_requirement_status(entities: list[dict], assertion: dict[str, Any], where: str) -> None:
    item_def_id = str(assertion["item_def_id"])
    matches = [
        entity for entity in entities
        if entity.get("type") == "loot" and str(entity.get("item_def_id", "")) == item_def_id
    ]
    if not matches:
        raise AssertionError(f"{where}: missing loot {item_def_id}: {entities}")
    assert_requirement_payload(matches[0], assertion, f"{where}: loot {item_def_id}")


def assert_requirement_payload(row: dict[str, Any], assertion: dict[str, Any], where: str) -> None:
    if "requirements_met" in assertion:
        want = bool(assertion["requirements_met"])
        got = bool(row.get("requirements_met"))
        if got != want:
            raise AssertionError(f"{where}: requirements_met {got} != {want}: {row}")
    expected_status = assertion.get("status", [])
    if expected_status:
        actual_status = row.get("requirement_status", [])
        if not isinstance(actual_status, list):
            raise AssertionError(f"{where}: requirement_status is not a list: {row}")
        for expected in expected_status:
            stat = str(expected["stat"])
            actual = next((entry for entry in actual_status if str(entry.get("stat", "")) == stat), None)
            if actual is None:
                raise AssertionError(f"{where}: missing requirement status {stat}: {actual_status}")
            for key in ("required", "current"):
                if key in expected and int(actual.get(key, -1)) != int(expected[key]):
                    raise AssertionError(f"{where}: requirement {stat}.{key} {actual.get(key)} != {expected[key]}: {actual}")
            if "met" in expected and bool(actual.get("met")) != bool(expected["met"]):
                raise AssertionError(f"{where}: requirement {stat}.met {actual.get('met')} != {expected['met']}: {actual}")
    if assertion.get("requires_requirement_status"):
        if not row.get("requirement_status") or "requirements_met" not in row:
            raise AssertionError(f"{where}: missing requirement status payload: {row}")
    if assertion.get("requires_equip_preview"):
        preview = row.get("equip_preview")
        if not isinstance(preview, dict):
            raise AssertionError(f"{where}: missing equip_preview: {row}")
        if "requirements_met" not in preview:
            raise AssertionError(f"{where}: equip_preview missing requirements_met: {preview}")
    preview = row.get("equip_preview")
    if isinstance(preview, dict):
        if assertion.get("preview_slot") is not None and str(preview.get("slot", "")) != str(assertion["preview_slot"]):
            raise AssertionError(f"{where}: preview slot {preview.get('slot')} != {assertion['preview_slot']}: {preview}")
        if assertion.get("preview_requirements_met") is not None:
            want = bool(assertion["preview_requirements_met"])
            got = bool(preview.get("requirements_met"))
            if got != want:
                raise AssertionError(f"{where}: preview requirements_met {got} != {want}: {preview}")
        delta_stats = [str(delta.get("stat", "")) for delta in preview.get("deltas", []) if isinstance(delta, dict)]
        for stat in assertion.get("preview_delta_stats", []):
            if str(stat) not in delta_stats:
                raise AssertionError(f"{where}: preview missing delta stat {stat}: {preview}")


def assert_hotbar_slot(hotbar: list[dict], slot_index: int, item_def_id: Any, where: str, inventory: list[dict] | None = None) -> None:
    assigned_id = hotbar_item_id(hotbar, slot_index)
    if item_def_id is None:
        if assigned_id is not None:
            raise AssertionError(f"{where}: hotbar[{slot_index}]={assigned_id}, want empty")
        return
    if assigned_id is None:
        raise AssertionError(f"{where}: hotbar[{slot_index}] is empty, want {item_def_id}")
    if inventory is None:
        return
    item = next((i for i in inventory if str(i.get("item_instance_id")) == assigned_id), None)
    if item is None or item.get("item_def_id") != str(item_def_id):
        raise AssertionError(f"{where}: hotbar[{slot_index}] item={item}, want {item_def_id}")


def assert_rolled_inventory_item(inventory: list[dict], assertion: dict[str, Any], where: str) -> None:
    item_def_id = str(assertion["item_def_id"])
    item = find_inventory_item(inventory, item_def_id)
    if item is None:
        raise AssertionError(f"{where}: missing rolled inventory item {item_def_id}: {inventory}")
    template_id = str(assertion.get("item_template_id", item_def_id))
    if item.get("item_template_id") != template_id:
        raise AssertionError(f"{where}: item_template_id {item.get('item_template_id')} != {template_id}: {item}")
    if assertion.get("equipped") is not None and bool(item.get("equipped")) != bool(assertion["equipped"]):
        raise AssertionError(f"{where}: rolled item equipped={item.get('equipped')} want {assertion['equipped']}: {item}")
    rarity = assertion.get("rarity")
    if rarity is not None and item.get("rarity") != str(rarity):
        raise AssertionError(f"{where}: rarity {item.get('rarity')} != {rarity}: {item}")
    suffix = str(assertion.get("display_name_suffix", "Cave Blade"))
    if not str(item.get("display_name", "")).endswith(suffix):
        raise AssertionError(f"{where}: display_name missing suffix {suffix}: {item}")
    stats = item.get("rolled_stats", {})
    if not isinstance(stats, dict):
        raise AssertionError(f"{where}: rolled_stats is not an object: {item}")
    for key in assertion.get("stat_keys", []):
        if key not in stats:
            raise AssertionError(f"{where}: missing rolled stat {key}: {item}")
    min_damage = assertion.get("min_damage")
    if min_damage is not None and int(stats.get("damage_min", -1)) < int(min_damage):
        raise AssertionError(f"{where}: damage_min {stats.get('damage_min')} < {min_damage}: {item}")
    req = item.get("requirements", {})
    if int(req.get("level", 0)) != int(assertion.get("required_level", 1)):
        raise AssertionError(f"{where}: required level {req.get('level')} mismatch: {item}")
    if item.get("effect_ids", []) != []:
        raise AssertionError(f"{where}: effect_ids should be empty in v23: {item}")


def assert_rolled_inventory_any(inventory: list[dict], equipped: bool | None, where: str) -> None:
    matches = [
        item for item in inventory
        if item.get("item_template_id") and (equipped is None or bool(item.get("equipped")) == equipped)
    ]
    if not matches:
        raise AssertionError(f"{where}: missing rolled inventory item equipped={equipped}: {inventory}")
    item = matches[0]
    stats = item.get("rolled_stats", {})
    if not isinstance(stats, dict) or not stats:
        raise AssertionError(f"{where}: rolled item missing stats: {item}")
    req = item.get("requirements", {})
    if int(req.get("level", 0)) != 1:
        raise AssertionError(f"{where}: rolled item required level mismatch: {item}")
    if item.get("effect_ids", []) != []:
        raise AssertionError(f"{where}: rolled item effect_ids should be empty: {item}")


def assert_entity_count(entities: list[dict], assertion: dict[str, Any], where: str) -> None:
    matches: list[dict] = []
    for entity in entities:
        if entity_matches_selector(entity, assertion):
            matches.append(entity)
    assert_count_matches(len(matches), assertion, f"{where}: entity count", f" for {assertion}: {matches}")


def entity_matches_selector(entity: dict[str, Any], selector: dict[str, Any]) -> bool:
    string_filters = {
        "entity_type": "type",
        "monster_def_id": "monster_def_id",
        "interactable_def_id": "interactable_def_id",
        "item_def_id": "item_def_id",
        "item_template_id": "item_template_id",
        "rarity": "rarity",
        "state": "state",
        "boss_template_id": "boss_template_id",
        "visual_model": "visual_model",
    }
    for selector_key, entity_key in string_filters.items():
        if selector_key in selector and selector[selector_key] is not None:
            if str(entity.get(entity_key, "")) != str(selector[selector_key]):
                return False
    if "level" in selector and selector["level"] is not None:
        if int(entity.get("level", -999999)) != int(selector["level"]):
            return False
    if selector.get("is_boss") is not None:
        if bool(entity.get("is_boss", False)) != bool(selector["is_boss"]):
            return False
    if selector.get("visual_scale") is not None:
        if abs(float(entity.get("visual_scale", 1.0)) - float(selector["visual_scale"])) > 0.000001:
            return False
    if selector.get("alive") is not None:
        hp = entity.get("hp")
        is_alive = not isinstance(hp, int) or hp > 0
        if is_alive != bool(selector["alive"]):
            return False
    return True


def assert_count_matches(got: int, assertion: dict[str, Any], label: str, suffix: str = "") -> None:
    if "equals" in assertion:
        want = int(assertion["equals"])
        if got != want:
            raise AssertionError(f"{label} {got} != {want}{suffix}")
    if "at_least" in assertion:
        want = int(assertion["at_least"])
        if got < want:
            raise AssertionError(f"{label} {got} < {want}{suffix}")
    if "at_most" in assertion:
        want = int(assertion["at_most"])
        if got > want:
            raise AssertionError(f"{label} {got} > {want}{suffix}")
    if "between" in assertion:
        bounds = assertion["between"]
        if not isinstance(bounds, list) or len(bounds) != 2:
            raise AssertionError(f"{label}: between must be [min, max]{suffix}")
        low = int(bounds[0])
        high = int(bounds[1])
        if got < low or got > high:
            raise AssertionError(f"{label} {got} not between {low} and {high}{suffix}")


def assert_character_progression(progression: dict[str, Any], assertion: dict[str, Any], where: str) -> None:
    if not progression:
        raise AssertionError(f"{where}: missing character_progression")
    for key in ("level", "experience", "unspent_stat_points", "gold", "deepest_dungeon_depth"):
        if key in assertion:
            want = int(assertion[key])
            got = int(progression.get(key, -1))
            if got != want:
                raise AssertionError(f"{where}: character_progression.{key} {got} != {want}: {progression}")
    base_stats = progression.get("base_stats", {})
    for key in ("str", "dex", "vit", "magic"):
        if key in assertion:
            want = int(assertion[key])
            got = int(base_stats.get(key, -1))
            if got != want:
                raise AssertionError(f"{where}: base_stats.{key} {got} != {want}: {progression}")
    derived = progression.get("derived_stats", {})
    expected_derived = assertion.get("derived_stats", {})
    if isinstance(expected_derived, dict):
        for key, want_raw in expected_derived.items():
            got = float(derived.get(key, -999999))
            want = float(want_raw)
            if abs(got - want) > 0.000001:
                raise AssertionError(f"{where}: derived_stats.{key} {got} != {want}: {progression}")
    expected_breakdowns = assertion.get("stat_breakdowns", [])
    if isinstance(expected_breakdowns, list):
        assert_stat_breakdowns(progression.get("stat_breakdowns", []), expected_breakdowns, where)


def assert_stat_breakdowns(actual: Any, expected_rows: list[dict[str, Any]], where: str) -> None:
    if not isinstance(actual, list):
        raise AssertionError(f"{where}: stat_breakdowns is not a list: {actual}")
    for expected in expected_rows:
        key = str(expected["key"])
        row = next((r for r in actual if isinstance(r, dict) and str(r.get("key")) == key), None)
        if row is None:
            raise AssertionError(f"{where}: missing stat_breakdown {key}: {actual}")
        for number_key in ("value", "uncapped_value", "cap"):
            if number_key not in expected:
                continue
            want_raw = expected[number_key]
            got_raw = row.get(number_key)
            if want_raw is None:
                if got_raw is not None:
                    raise AssertionError(f"{where}: stat_breakdown {key}.{number_key} {got_raw} != None: {row}")
                continue
            got = float(got_raw)
            want = float(want_raw)
            if abs(got - want) > 0.000001:
                raise AssertionError(f"{where}: stat_breakdown {key}.{number_key} {got} != {want}: {row}")
        min_value = expected.get("min_value")
        if min_value is not None and float(row.get("value", -999999)) < float(min_value):
            raise AssertionError(f"{where}: stat_breakdown {key}.value {row.get('value')} < {min_value}: {row}")
        min_uncapped = expected.get("min_uncapped_value")
        if min_uncapped is not None and float(row.get("uncapped_value", -999999)) < float(min_uncapped):
            raise AssertionError(f"{where}: stat_breakdown {key}.uncapped_value {row.get('uncapped_value')} < {min_uncapped}: {row}")
        sources = row.get("sources", [])
        if not isinstance(sources, list):
            raise AssertionError(f"{where}: stat_breakdown {key}.sources is not a list: {row}")
        source_kinds = {str(source.get("kind")) for source in sources if isinstance(source, dict)}
        for kind in expected.get("source_kinds", []):
            if str(kind) not in source_kinds:
                raise AssertionError(f"{where}: stat_breakdown {key} missing source kind {kind}: {row}")


def run_assertions(
    assertions: list[Any],
    entities: list[dict],
    inventory: list[dict],
    equipped: dict,
    item_id: str | None,
    where: str,
    current_level: int | None = None,
    walls: list[dict] | None = None,
    discovered_teleporters: dict[int, bool] | None = None,
    character_progression: dict[str, Any] | None = None,
    hotbar_capacity: int | None = None,
    hotbar: list[dict] | None = None,
    inventory_rows: int | None = None,
    inventory_capacity: int | None = None,
    gold: int | None = None,
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
            matches = inventory
            if assertion.get("item_def_id") is not None:
                matches = [item for item in matches if str(item.get("item_def_id", "")) == str(assertion["item_def_id"])]
            if assertion.get("item_template_id") is not None:
                matches = [item for item in matches if str(item.get("item_template_id", "")) == str(assertion["item_template_id"])]
            if assertion.get("equipped") is not None:
                matches = [item for item in matches if bool(item.get("equipped")) == bool(assertion["equipped"])]
            assert_count_matches(len(matches), assertion, f"{where}: inventory count", f": {matches}")
        elif typ == "inventory_contains":
            expected_equipped = assertion.get("equipped")
            assert_inventory_contains(inventory, str(assertion["item_def_id"]), expected_equipped, where)
        elif typ == "inventory_requirement_status":
            assert_inventory_requirement_status(inventory, assertion, where)
        elif typ == "loot_requirement_status":
            assert_loot_requirement_status(entities, assertion, where)
        elif typ == "gold":
            assert_count_matches(int(gold or 0), assertion, f"{where}: gold")
        elif typ == "wall_count":
            wall_rows = list(walls or [])
            if assertion.get("source") is not None:
                source = str(assertion["source"])
                wall_rows = [wall for wall in wall_rows if str(wall.get("source", "")) == source]
            assert_count_matches(len(wall_rows), assertion, f"{where}: wall count", f": {wall_rows}")
        elif typ == "non_perimeter_wall_exists":
            wall_rows = list(walls or [])
            if not any(str(wall.get("source", "")) != "perimeter" for wall in wall_rows):
                raise AssertionError(f"{where}: no non-perimeter wall in {wall_rows}")
        elif typ == "rolled_inventory_item":
            assert_rolled_inventory_item(inventory, assertion, where)
        elif typ == "rolled_inventory_any":
            expected_equipped = assertion.get("equipped")
            assert_rolled_inventory_any(inventory, expected_equipped, where)
        elif typ == "entity_count":
            assert_entity_count(entities, assertion, where)
        elif typ == "equipped_weapon_def":
            expected_def = str(assertion["item_def_id"])
            slot = str(assertion.get("slot", "main_hand"))
            weapon_id = equipped.get(slot)
            if weapon_id is None:
                raise AssertionError(f"{where}: equipped {slot} is empty, want {expected_def}")
            item = next((i for i in inventory if str(i.get("item_instance_id")) == str(weapon_id)), None)
            if item is None or item.get("item_def_id") != expected_def:
                raise AssertionError(f"{where}: equipped {slot} {weapon_id} row = {item}, want {expected_def}")
        elif typ == "equipped_slot_def":
            slot = str(assertion["slot"])
            expected_def = str(assertion["item_def_id"])
            item_id = equipped.get(slot)
            if item_id is None:
                raise AssertionError(f"{where}: equipped {slot} is empty, want {expected_def}")
            item = next((i for i in inventory if str(i.get("item_instance_id")) == str(item_id)), None)
            if item is None or item.get("item_def_id") != expected_def:
                raise AssertionError(f"{where}: equipped {slot} {item_id} row = {item}, want {expected_def}")
        elif typ == "equipped_slot_empty":
            slot = str(assertion["slot"])
            if equipped.get(slot) is not None:
                raise AssertionError(f"{where}: equipped {slot}={equipped.get(slot)}, want empty")
        elif typ == "hotbar_capacity":
            assert_count_matches(int(hotbar_capacity or 0), assertion, f"{where}: hotbar_capacity")
        elif typ == "hotbar_slot":
            assert_hotbar_slot(hotbar or [], int(assertion["slot_index"]), assertion.get("item_def_id"), where, inventory)
        elif typ == "inventory_capacity":
            if "rows" in assertion:
                assert_count_matches(int(inventory_rows or 0), {"equals": int(assertion["rows"])}, f"{where}: inventory_rows")
            assert_count_matches(int(inventory_capacity or 0), assertion, f"{where}: inventory_capacity")
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
        elif typ in {
            "monster_moved",
            "entity_moved",
            "monster_within_player_distance",
            "monster_near_spawn",
            "event_seen",
            "combat_event_seen",
            "player_never_in_melee_range_of",
            "monster_damage_at_least",
            "player_hp_decreased_from_recorded",
            "shop_offer_count",
            "shop_offer_details",
            "shop_sell_appraisal_count",
            "shop_sell_appraisal_details",
            "shop_event",
        }:
            continue
        elif typ == "player_hp_equals":
            assert_player_hp_equals(entities, int(assertion["equals"]), where)
        elif typ == "player_max_hp_equals":
            assert_player_max_hp_equals(entities, int(assertion["equals"]), where)
        elif typ == "character_progression":
            assert_character_progression(character_progression or {}, assertion, where)
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
        if typ == "combat_event_seen":
            if not any(combat_event_matches(event, assertion) for event in state.combat_events):
                raise AssertionError(
                    f"{where}: combat event {combat_event_summary(assertion)} not seen; have {state.combat_events}"
                )
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
        if typ == "entity_moved":
            min_distance = float(assertion["min_distance"])
            matches = [entity for entity in state.entities.values() if entity_matches_selector(entity, assertion)]
            if not matches:
                raise AssertionError(f"{where}: no entity matched {assertion}")
            matches.sort(key=lambda entity: str(entity.get("id", "")))
            entity = matches[0]
            entity_id = str(entity.get("id", ""))
            initial = state.initial_entity_positions.get(entity_id)
            if initial is None:
                raise AssertionError(f"{where}: missing initial position for entity {entity_id}")
            pos = entity.get("position", {})
            dx = float(pos.get("x", 0.0)) - initial["x"]
            dy = float(pos.get("y", 0.0)) - initial["y"]
            dist = (dx * dx + dy * dy) ** 0.5
            if dist < min_distance:
                raise AssertionError(
                    f"{where}: entity {entity_id} moved {dist:.3f} < min_distance {min_distance}"
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
        if typ == "player_hp_decreased_from_recorded":
            if state.recorded_player_hp is None:
                raise AssertionError(f"{where}: no recorded player hp")
            player = find_player(state)
            if player is None or not isinstance(player.get("hp"), int):
                raise AssertionError(f"{where}: missing current player hp: {player}")
            current_hp = int(player["hp"])
            if current_hp >= state.recorded_player_hp:
                raise AssertionError(
                    f"{where}: player hp {current_hp} did not decrease below recorded {state.recorded_player_hp}"
                )
            continue
        if typ == "current_level":
            want = int(assertion["equals"])
            if state.current_level != want:
                raise AssertionError(f"{where}: current_level {state.current_level} != {want}")
            continue
        if typ == "wall_count":
            walls = state.walls
            if assertion.get("source") is not None:
                source = str(assertion["source"])
                walls = [wall for wall in walls if str(wall.get("source", "")) == source]
            assert_count_matches(len(walls), assertion, f"{where}: wall count", f": {walls}")
            continue
        if typ == "non_perimeter_wall_exists":
            if not any(str(wall.get("source", "")) != "perimeter" for wall in state.walls):
                raise AssertionError(f"{where}: no non-perimeter wall in {state.walls}")
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
        if typ == "inventory_requirement_status":
            assert_inventory_requirement_status(state.inventory, assertion, where)
            continue
        if typ == "loot_requirement_status":
            assert_loot_requirement_status(list(state.entities.values()), assertion, where)
            continue
        if typ == "inventory_count":
            matches = state.inventory
            if assertion.get("item_def_id") is not None:
                matches = [item for item in matches if str(item.get("item_def_id", "")) == str(assertion["item_def_id"])]
            if assertion.get("item_template_id") is not None:
                matches = [item for item in matches if str(item.get("item_template_id", "")) == str(assertion["item_template_id"])]
            if assertion.get("equipped") is not None:
                matches = [item for item in matches if bool(item.get("equipped")) == bool(assertion["equipped"])]
            assert_count_matches(len(matches), assertion, f"{where}: inventory count", f": {matches}")
            continue
        if typ == "gold":
            assert_count_matches(state.gold, assertion, f"{where}: gold")
            continue
        if typ == "shop_offer_count":
            offers = filtered_shop_offers(state, assertion)
            assert_count_matches(len(offers), assertion, f"{where}: shop_offer_count", f": {offers}")
            continue
        if typ == "shop_offer_details":
            offers = filtered_shop_offers(state, assertion)
            assert_shop_detail_rows(offers, assertion, f"{where}: shop_offer_details")
            continue
        if typ == "shop_sell_appraisal_count":
            rows = filtered_shop_sell_appraisals(state, assertion)
            assert_count_matches(len(rows), assertion, f"{where}: shop_sell_appraisal_count", f": {rows}")
            continue
        if typ == "shop_sell_appraisal_details":
            rows = filtered_shop_sell_appraisals(state, assertion)
            assertion = dict(assertion)
            assertion.setdefault("price_key", "sell_price")
            assert_shop_detail_rows(rows, assertion, f"{where}: shop_sell_appraisal_details")
            continue
        if typ == "shop_event":
            shop_id = str(assertion.get("shop_id", "town_vendor"))
            event_type = str(assertion["event_type"])
            matches = [
                event for event in state.shop_events
                if event.get("event_type") == event_type and str(event.get("shop_id", "")) == shop_id
            ]
            assert_count_matches(len(matches), assertion, f"{where}: shop_event", f": {matches}")
            continue
        if typ == "rolled_inventory_item":
            assert_rolled_inventory_item(state.inventory, assertion, where)
        if typ == "rolled_inventory_any":
            expected_equipped = assertion.get("equipped")
            assert_rolled_inventory_any(state.inventory, expected_equipped, where)
            continue
        if typ == "equipped_slot_def":
            slot = str(assertion["slot"])
            expected_def = str(assertion["item_def_id"])
            item_id = state.equipped.get(slot)
            item = next((i for i in state.inventory if str(i.get("item_instance_id")) == str(item_id)), None)
            if item_id is None or item is None or item.get("item_def_id") != expected_def:
                raise AssertionError(f"{where}: equipped {slot} {item_id} row={item}, want {expected_def}")
            continue
        if typ == "equipped_slot_empty":
            slot = str(assertion["slot"])
            if state.equipped.get(slot) is not None:
                raise AssertionError(f"{where}: equipped {slot}={state.equipped.get(slot)}, want empty")
            continue
        if typ == "hotbar_capacity":
            assert_count_matches(state.hotbar_capacity, assertion, f"{where}: hotbar_capacity")
            continue
        if typ == "hotbar_slot":
            assert_hotbar_slot(state.hotbar, int(assertion["slot_index"]), assertion.get("item_def_id"), where, state.inventory)
            continue
        if typ == "inventory_capacity":
            if "rows" in assertion:
                assert_count_matches(state.inventory_rows, {"equals": int(assertion["rows"])}, f"{where}: inventory_rows")
            assert_count_matches(state.inventory_capacity, assertion, f"{where}: inventory_capacity")
            continue
        if typ == "entity_count":
            assert_entity_count(list(state.entities.values()), assertion, where)
            continue
        if typ == "monster_damage_at_least":
            monster_def_id = str(assertion["monster_def_id"])
            got = state.max_monster_damage_by_def.get(monster_def_id, 0)
            want = int(assertion["damage"])
            if got < want:
                raise AssertionError(f"{where}: max damage to {monster_def_id} = {got}, want at least {want}")
            continue
        if typ == "character_progression":
            assert_character_progression(state.character_progression, assertion, where)
            continue
        if typ == "player_max_hp_equals":
            player = find_player(state)
            if player is None:
                raise AssertionError(f"{where}: player not found")
            got = int(player.get("max_hp", -1))
            want = int(assertion["equals"])
            if got != want:
                raise AssertionError(f"{where}: player max_hp {got} != {want}")
            continue


def default_manifest_path() -> Path:
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    return BOT_RUN_ARTIFACT_DIR / f"{stamp}.json"


def should_clean_bot_run_artifacts(manifest_path: Path) -> bool:
    return manifest_path.parent.resolve() == BOT_RUN_ARTIFACT_DIR.resolve()


def clean_bot_run_artifacts(artifact_dir: Path = BOT_RUN_ARTIFACT_DIR) -> int:
    if not artifact_dir.exists():
        return 0

    removed = 0
    for path in artifact_dir.glob("*.json"):
        if path.is_file():
            path.unlink()
            removed += 1
    return removed


def write_manifest(path: Path, base_url: str, results: list[dict[str, Any]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    body = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "base_url": base_url,
        "scenarios": results,
    }
    path.write_text(json.dumps(body, indent=2) + "\n")


def scenario_email(base_email: str, scenario_id: str) -> str:
    if "@" not in base_email:
        return base_email
    local, domain = base_email.split("@", 1)
    safe_id = "".join(ch if ch.isalnum() else "-" for ch in scenario_id)
    return f"{local}+{safe_id}-{int(time.time() * 1000)}@{domain}"


def run_verified_session(
    *,
    client: httpx.Client,
    base_url: str,
    token: str,
    debug_token: str,
    scenario: Scenario,
    world_id: str,
    steps: list[dict[str, Any]],
    assertions: list[Any],
    seed: str = "",
) -> tuple[dict[str, Any], RuntimeState]:
    sess = create_session(client, token, world_id, seed)
    session_id = sess["session_id"]
    log("session created", session_id, f"seed={sess.get('seed')}")

    phase_started = time.monotonic()
    run_scenario = Scenario(
        id=scenario.id,
        world_id=world_id,
        seed=seed,
        peer_count=scenario.peer_count,
        title=scenario.title,
        description=scenario.description,
        steps=steps,
        assertions=assertions,
        fresh_session_checks=[],
        path=scenario.path,
    )
    observed = asyncio.run(drive_scenario(base_url, token, sess, run_scenario))
    log("phase drive done", f"elapsed={time.monotonic() - phase_started:.2f}s")
    run_runtime_assertions(assertions, observed, "runtime protocol")

    phase_started = time.monotonic()
    state = fetch_state(client, token, debug_token, session_id)
    run_assertions(
        assertions,
        state["entities"],
        state["inventory"],
        state["equipped"],
        observed.item_id,
        "/state API",
        current_level=int(state.get("current_level", 0)),
        walls=state.get("walls", []),
        discovered_teleporters=parse_discovered_teleporters(state),
        character_progression=state.get("character_progression", {}),
        hotbar_capacity=int(state.get("hotbar_capacity", 2)),
        hotbar=state.get("hotbar", []),
        inventory_rows=int(state.get("inventory_rows", 3)),
        inventory_capacity=int(state.get("inventory_capacity", int(state.get("inventory_rows", 3)) * 5)),
        gold=int(state.get("gold", 0)),
    )
    log("phase /state done", f"elapsed={time.monotonic() - phase_started:.2f}s")

    phase_started = time.monotonic()
    asyncio.run(check_persistence(base_url, token, session_id, observed.item_id, assertions))
    log("phase reconnect done", f"elapsed={time.monotonic() - phase_started:.2f}s")

    phase_started = time.monotonic()
    replay = fetch_replay(client, token, debug_token, session_id)
    if not replay.get("match", False):
        raise AssertionError(f"replay mismatch for {session_id}: {replay.get('mismatch')}")
    log("phase replay done", f"elapsed={time.monotonic() - phase_started:.2f}s", session_id)
    return sess, observed


async def connect_coop_peer(base_url: str, token: str, sess: dict[str, Any], label: str, world_id: str) -> CoopPeer:
    uri = to_ws_url(base_url, sess["ws_url"])
    ws = await websockets.connect(uri, additional_headers=auth(token))
    first = await recv_json(ws)
    assert first["type"] == "session_snapshot", first["type"]
    state = RuntimeState(world_id=world_id)
    ingest_snapshot(first["payload"], state)
    await ws.send(json.dumps(make_envelope(
        "client_ready",
        sess["session_id"],
        state.last_tick,
        {"client_version": f"bot-{label}", "last_seen_tick": state.last_tick},
    )))
    log("co-op peer connected", label, "local_player_id", state.local_player_id, "level", state.current_level)
    return CoopPeer(label=label, token=token, session=sess, state=state, ws=ws)


async def close_coop_peer(peer: CoopPeer) -> None:
    await peer.ws.close()


async def pump_coop(peers: list[CoopPeer], timeout: float = 0.1) -> None:
    tasks = {asyncio.create_task(peer.ws.recv()): peer for peer in peers}
    done, pending = await asyncio.wait(tasks.keys(), timeout=timeout, return_when=asyncio.FIRST_COMPLETED)
    for task in pending:
        task.cancel()
    for task in done:
        peer = tasks[task]
        ingest_message(json.loads(task.result()), peer.state)


async def wait_coop_until(peers: list[CoopPeer], label: str, predicate, timeout_s: float = SLICE_TIMEOUT_S) -> None:
    loop = asyncio.get_event_loop()
    deadline = loop.time() + timeout_s
    while not predicate():
        if loop.time() > deadline:
            raise TimeoutError(f"co-op wait timed out: {label}")
        await pump_coop(peers, timeout=0.1)


async def send_coop_intent(peer: CoopPeer, msg_type: str, payload: dict[str, Any]) -> str:
    env = make_envelope(msg_type, peer.session["session_id"], peer.state.last_tick, payload)
    await peer.ws.send(json.dumps(env))
    return str(env["message_id"])


async def wait_coop_accept(peers: list[CoopPeer], peer: CoopPeer, message_id: str) -> None:
    await wait_coop_until(
        peers,
        f"{peer.label} accept {message_id}",
        lambda: message_id in peer.state.accepted_message_ids or message_id in peer.state.rejected_message_reasons,
    )
    if message_id in peer.state.rejected_message_reasons:
        raise AssertionError(f"{peer.label} intent {message_id} rejected: {peer.state.rejected_message_reasons[message_id]}")


def player_position(state: RuntimeState) -> dict[str, Any]:
    player = find_player(state)
    if player is None:
        raise AssertionError(f"missing local player {state.local_player_id}")
    return dict(player.get("position", {}))


def player_entity_ids(state: RuntimeState) -> set[str]:
    return {str(entity_id) for entity_id, entity in state.entities.items() if entity.get("type") == "player"}


def assert_party_contains_roles(state: RuntimeState, where: str) -> None:
    roles = {str(row.get("role", "")) for row in state.party}
    if not {"host", "guest"} <= roles:
        raise AssertionError(f"{where}: party roles {roles}, want host+guest; party={state.party}")


async def move_coop_peer(peers: list[CoopPeer], peer: CoopPeer, direction: dict[str, int]) -> None:
    before = player_position(peer.state)
    message_id = await send_coop_intent(peer, "move_intent", {"direction": direction, "duration_ticks": 1})
    await wait_coop_accept(peers, peer, message_id)
    await wait_coop_until(
        peers,
        f"{peer.label} local movement",
        lambda: player_position(peer.state) != before,
    )


async def run_true_coop_session(
    *,
    client: httpx.Client,
    base_url: str,
    host_token: str,
    guest_token: str,
    debug_token: str,
    scenario: Scenario,
    host_character_id: str,
    guest_character_id: str,
) -> tuple[dict[str, Any], RuntimeState]:
    sess = create_coop_session(client, host_token, scenario.world_id, host_character_id, scenario.seed)
    session_id = sess["session_id"]
    host = await connect_coop_peer(base_url, host_token, sess, "host", scenario.world_id)
    try:
        await execute_step(host.ws, session_id, host.state, {"action": "use_stair", "direction": "down", "max_ticks": 240}, asyncio.get_event_loop())
        if host.state.current_level != -1:
            raise AssertionError(f"host level after descend = {host.state.current_level}, want -1")

        joined = join_coop_session(client, guest_token, session_id, str(sess["join_code"]), guest_character_id)
        guest = await connect_coop_peer(base_url, guest_token, joined, "guest", scenario.world_id)
        try:
            if guest.state.current_level != 0:
                raise AssertionError(f"guest joined level {guest.state.current_level}, want town level 0")
            if host.state.local_player_id == guest.state.local_player_id:
                raise AssertionError(f"host and guest local_player_id match: {host.state.local_player_id}")
            assert_party_contains_roles(guest.state, "guest initial snapshot")

            await execute_step(guest.ws, session_id, guest.state, {"action": "use_stair", "direction": "down", "max_ticks": 240}, asyncio.get_event_loop())
            peers = [host, guest]
            await wait_coop_until(
                peers,
                "both clients see both player entities on shared level",
                lambda: player_entity_ids(host.state) >= {host.state.local_player_id, guest.state.local_player_id}
                and player_entity_ids(guest.state) >= {host.state.local_player_id, guest.state.local_player_id},
            )

            host_before = player_position(host.state)
            guest_before = player_position(guest.state)
            await move_coop_peer(peers, host, {"x": 1, "y": 0})
            await move_coop_peer(peers, guest, {"x": 0, "y": 1})
            if player_position(host.state) == host_before:
                raise AssertionError("host did not move its local player")
            if player_position(guest.state) == guest_before:
                raise AssertionError("guest did not move its local player")

            await close_coop_peer(guest)
            await wait_coop_until(
                [host],
                "guest entity removed after disconnect",
                lambda: guest.state.local_player_id not in player_entity_ids(host.state),
            )
            moved_after_disconnect = player_position(host.state)
            await move_coop_peer([host], host, {"x": -1, "y": 0})
            if player_position(host.state) == moved_after_disconnect:
                raise AssertionError("host stopped moving after guest disconnected")

            guest_reconnect = await connect_coop_peer(base_url, guest_token, joined, "guest-reconnect", scenario.world_id)
            try:
                if guest_reconnect.state.local_player_id != guest.state.local_player_id:
                    raise AssertionError(
                        f"guest reconnect local_player_id={guest_reconnect.state.local_player_id}, want {guest.state.local_player_id}"
                    )
                await wait_coop_until(
                    [host, guest_reconnect],
                    "guest reconnect town snapshot",
                    lambda: guest_reconnect.state.current_level == 0,
                    timeout_s=3.0,
                )
                if guest_reconnect.state.current_level != 0:
                    raise AssertionError(f"guest reconnect level {guest_reconnect.state.current_level}, want town")
                assert_party_contains_roles(guest_reconnect.state, "guest reconnect snapshot")
            finally:
                await close_coop_peer(guest_reconnect)
        finally:
            try:
                await close_coop_peer(guest)
            except Exception:
                pass
    finally:
        try:
            await close_coop_peer(host)
        except Exception:
            pass

    replay = fetch_replay(client, host_token, debug_token, session_id)
    if not replay.get("match", False):
        raise AssertionError(f"co-op replay mismatch for {session_id}: {replay.get('mismatch')}")
    log("co-op replay matched", session_id)
    return sess, host.state


def assert_active_session_row(rows: list[dict[str, Any]], session_id: str, *, min_members: int = 1) -> None:
    row = next((row for row in rows if str(row.get("session_id")) == session_id), None)
    if row is None:
        raise AssertionError(f"active sessions missing {session_id}: {rows}")
    if row.get("join_code") is not None:
        raise AssertionError(f"active session row leaked join_code: {row}")
    if row.get("mode") != "coop" or row.get("listed") is not True:
        raise AssertionError(f"active session row shape mismatch: {row}")
    if int(row.get("member_count", 0)) < min_members:
        raise AssertionError(f"active session member_count={row.get('member_count')}, want >= {min_members}: {row}")
    if not row.get("host_display_name"):
        raise AssertionError(f"active session missing host_display_name: {row}")


async def run_session_browser_uncapped_coop(
    *,
    client: httpx.Client,
    base_url: str,
    tokens: list[str],
    debug_token: str,
    scenario: Scenario,
    character_ids: list[str],
) -> tuple[dict[str, Any], RuntimeState]:
    if len(tokens) < scenario.peer_count or len(character_ids) < scenario.peer_count:
        raise AssertionError(f"{scenario.id}: peer_count={scenario.peer_count} requires enough tokens/characters")
    sess = create_listed_coop_session(client, tokens[0], scenario.world_id, character_ids[0], scenario.seed)
    session_id = str(sess["session_id"])

    host = await connect_coop_peer(base_url, tokens[0], sess, "host", scenario.world_id)
    peers = [host]
    joined_sessions = [sess]
    try:
        assert_active_session_row(list_active_sessions(client, tokens[1]), session_id, min_members=1)
        for index in range(1, scenario.peer_count):
            joined = join_listed_session(client, tokens[index], session_id, character_ids[index])
            joined_sessions.append(joined)
            assert_active_session_row(list_active_sessions(client, tokens[index]), session_id, min_members=index + 1)
            peers.append(await connect_coop_peer(base_url, tokens[index], joined, f"peer-{index}", scenario.world_id))

        await wait_coop_until(
            peers,
            f"{scenario.peer_count} peers see same-level players",
            lambda: all(len(peer.state.party) >= scenario.peer_count for peer in peers)
            and all(player_entity_ids(peer.state) >= {p.state.local_player_id for p in peers} for peer in peers),
        )
        if len({peer.state.local_player_id for peer in peers}) != scenario.peer_count:
            raise AssertionError(f"local_player_id values are not distinct: {[peer.state.local_player_id for peer in peers]}")
        for peer in peers:
            assert_party_contains_roles(peer.state, f"{peer.label} party")

        for index, peer in enumerate(list(peers)):
            before = {other.label: player_position(other.state) for other in peers}
            await move_coop_peer(peers, peer, {"x": 1 if index % 2 == 0 else 0, "y": 0 if index % 2 == 0 else 1})
            after = {other.label: player_position(other.state) for other in peers}
            if after[peer.label] == before[peer.label]:
                raise AssertionError(f"{peer.label} did not move independently")
            for other in peers:
                if other is not peer and after[other.label] != before[other.label]:
                    raise AssertionError(f"{other.label} local player moved after {peer.label} input")

        disconnected = peers[-1]
        disconnected_id = disconnected.state.local_player_id
        await close_coop_peer(disconnected)
        remaining = peers[:-1]
        await wait_coop_until(
            remaining,
            "disconnected peer removed from remaining clients",
            lambda: all(disconnected_id not in player_entity_ids(peer.state) for peer in remaining),
        )
        moved_after_disconnect = player_position(remaining[0].state)
        await move_coop_peer(remaining, remaining[0], {"x": -1, "y": 0})
        if player_position(remaining[0].state) == moved_after_disconnect:
            raise AssertionError("remaining peer stopped moving after disconnect")

        reconnect = await connect_coop_peer(base_url, tokens[-1], joined_sessions[-1], "peer-reconnect", scenario.world_id)
        try:
            if reconnect.state.local_player_id != disconnected_id:
                raise AssertionError(f"reconnect local_player_id={reconnect.state.local_player_id}, want {disconnected_id}")
        finally:
            await close_coop_peer(reconnect)
    finally:
        for peer in peers:
            try:
                await close_coop_peer(peer)
            except Exception:
                pass

    replay = fetch_replay(client, tokens[0], debug_token, session_id)
    if not replay.get("match", False):
        raise AssertionError(f"listed co-op replay mismatch for {session_id}: {replay.get('mismatch')}")
    log("listed uncapped co-op replay matched", session_id)
    return sess, host.state


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
        for scenario in selected:
            scenario_started = time.monotonic()
            log("scenario begin", scenario.id, "-", scenario.title, f"world={scenario.world_id}")
            if scenario.id == "true_coop_session":
                replay_email = scenario_email(args.email, scenario.id + "-host")
                guest_email = scenario_email(args.email, scenario.id + "-guest")
                _, host_token = dev_login(client, replay_email, args.dev_token)
                _, guest_token = dev_login(client, guest_email, args.dev_token)
                host_character_id = ensure_character(client, host_token, "Coop Host")
                guest_character_id = ensure_character(client, guest_token, "Coop Guest")
                sess, observed = asyncio.run(run_true_coop_session(
                    client=client,
                    base_url=args.base_url,
                    host_token=host_token,
                    guest_token=guest_token,
                    debug_token=args.debug_token,
                    scenario=scenario,
                    host_character_id=host_character_id,
                    guest_character_id=guest_character_id,
                ))
                token = host_token
            elif scenario.id == "session_browser_uncapped_coop":
                tokens: list[str] = []
                character_ids: list[str] = []
                replay_email = scenario_email(args.email, f"{scenario.id}-peer-0")
                for index in range(scenario.peer_count):
                    email = replay_email if index == 0 else scenario_email(args.email, f"{scenario.id}-peer-{index}")
                    _, peer_token = dev_login(client, email, args.dev_token)
                    tokens.append(peer_token)
                    character_ids.append(ensure_character(client, peer_token, f"Coop Peer {index + 1}"))
                sess, observed = asyncio.run(run_session_browser_uncapped_coop(
                    client=client,
                    base_url=args.base_url,
                    tokens=tokens,
                    debug_token=args.debug_token,
                    scenario=scenario,
                    character_ids=character_ids,
                ))
                token = tokens[0]
            else:
                replay_email = scenario_email(args.email, scenario.id)
                _, token = dev_login(client, replay_email, args.dev_token)
                sess, observed = run_verified_session(
                    client=client,
                    base_url=args.base_url,
                    token=token,
                    debug_token=args.debug_token,
                    scenario=scenario,
                    world_id=scenario.world_id,
                    steps=scenario.steps,
                    assertions=scenario.assertions,
                    seed=scenario.seed,
                )
            session_id = sess["session_id"]
            last_session_id = session_id

            for idx, check in enumerate(scenario.fresh_session_checks, start=1):
                if scenario.id == "true_coop_session":
                    break
                log("fresh session check begin", scenario.id, f"#{idx}")
                check_world = str(check.get("world_id", scenario.world_id))
                check_steps = list(check.get("steps", []))
                check_assertions = list(check.get("assertions", []))
                check_sess, check_observed = run_verified_session(
                    client=client,
                    base_url=args.base_url,
                    token=token,
                    debug_token=args.debug_token,
                    scenario=scenario,
                    world_id=check_world,
                    steps=check_steps,
                    assertions=check_assertions,
                    seed=str(check.get("seed", "")),
                )
                last_session_id = check_sess["session_id"]
                observed = check_observed
                log("fresh session check done", scenario.id, f"#{idx}")

            log("scenario done", scenario.id, f"elapsed={time.monotonic() - scenario_started:.2f}s")

            scenario_visual = json.loads(scenario.path.read_text()).get("visual")
            entry = {
                "id": scenario.id,
                "title": scenario.title,
                "description": scenario.description,
                "world_id": scenario.world_id,
                "session_id": session_id,
                "replay_email": replay_email,
                "seed": sess["seed"],
                "final_tick": observed.last_tick,
                "status": "passed",
                "replay_match": True,
            }
            if isinstance(scenario_visual, dict):
                entry["visual"] = scenario_visual
            results.append(entry)

    if args.write_manifest:
        if should_clean_bot_run_artifacts(args.write_manifest):
            removed = clean_bot_run_artifacts(args.write_manifest.parent)
            if removed:
                log("deleted old bot run artifacts", removed)
        write_manifest(args.write_manifest, args.base_url, results)
        log("wrote manifest", args.write_manifest)

    log("BOT OK", "- scenarios:", ", ".join(r["id"] for r in results))
    if args.print_session_id:
        print(last_session_id)
    return 0


if __name__ == "__main__":
    sys.exit(main())
