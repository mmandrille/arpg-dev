#!/usr/bin/env python3
"""Headless Python protocol bot (ADR-0001 D8.5 layer 1).

Plays discovered bot scenarios through the same auth + WebSocket path as the
real client and asserts authoritative outcomes:

    dev-login -> create session -> move -> attack until dead -> pick up loot
    -> equip -> assert via /state -> reconnect and assert reconstructed state.

Progress goes to stderr; with --print-session-id the recorded session id is
written to stdout (and nothing else) so it can be captured for replay.

Usage:
    python -m tools.bot.run --base-url http://localhost:8888 \
        --dev-token local-dev-token --debug-token local-debug-token \
        [--scenario all] [--write-manifest path] [--print-session-id]
"""
from __future__ import annotations

import argparse
import asyncio
from datetime import datetime, timezone
import json
import math
from pathlib import Path
import sys
import time
from typing import Any

import httpx
import websockets

from tools.bot.bot_types import CoopPeer, DEFAULT_WORLD_ID, RuntimeState, Scenario
from tools.bot.protocol import make_envelope, to_ws_url
from tools.bot.bot_context import BotContext
from tools.bot.runtime_queries import dict_distance, find_player
from tools.bot import skill_visual_runtime
from tools.bot.debug_progression import debug_progression_body
from tools.bot.stash_assertions import (
    assert_stash_capacity,
    assert_stash_event,
    assert_stash_gold,
    assert_stash_item_count,
    filtered_stash_items,
    find_stash_item_by_id,
    select_stash_item,
)
from tools.bot.unique_effect_assertions import assert_inventory_unique_effect_coverage
from tools.bot.weapon_set_runtime import equipped_slot_id
from tools.bot.skill_binding_runtime import assert_skill_bindings, set_skill_bindings
SLICE_TIMEOUT_S = 20.0
MAX_SCENARIO_ELAPSED_S = 15.0
WAIT_LOG_INTERVAL_S = 2.0
WALK_STOP_DISTANCE = 1.0
WALK_MAX_TICKS = 40
ROOT = Path(__file__).resolve().parent.parent.parent
SCENARIO_DIR = Path(__file__).resolve().parent / "scenarios"
BOT_RUN_ARTIFACT_DIR = ROOT / ".artifacts" / "bot-runs"


def load_known_world_ids() -> set[str]:
    worlds_path = ROOT / "shared" / "rules" / "worlds.v0.json"
    data = json.loads(worlds_path.read_text(encoding="utf-8"))
    return set(data["worlds"])


def monster_xp_reward(monster_def_id: str) -> int:
    monsters_path = ROOT / "shared" / "rules" / "monsters.v0.json"
    data = json.loads(monsters_path.read_text(encoding="utf-8"))
    return int(data.get("monsters", {}).get(monster_def_id, {}).get("xp_reward", 0))


KNOWN_WORLD_IDS = load_known_world_ids()
_SKILL_RULES: dict[str, Any] | None = None


def skill_rule_max_rank(skill_id: str) -> int:
    global _SKILL_RULES
    if _SKILL_RULES is None:
        skills_path = ROOT / "shared" / "rules" / "skills.v0.json"
        data = json.loads(skills_path.read_text(encoding="utf-8"))
        _SKILL_RULES = dict(data.get("skills", {}))
    skill = _SKILL_RULES.get(skill_id)
    if not isinstance(skill, dict):
        raise AssertionError(f"shared skill rule {skill_id} not found")
    return int(skill.get("max_rank", -1))


def log(*args: Any) -> None:
    stamp = datetime.now(timezone.utc).strftime("%H:%M:%S")
    print(f"[bot {stamp}]", *args, file=sys.stderr, flush=True)


def log_wait_progress(label: str, loop, started_at: float, **details: Any) -> None:
    elapsed = loop.time() - started_at
    parts = [f"{label} elapsed={elapsed:.1f}s"]
    for key, value in details.items():
        parts.append(f"{key}={value}")
    log(*parts)


def assert_scenario_elapsed_within_budget(scenario_id: str, elapsed_s: float) -> None:
    if elapsed_s > MAX_SCENARIO_ELAPSED_S:
        raise TimeoutError(
            f"protocol bot scenario {scenario_id} took {elapsed_s:.2f}s; "
            f"budget is {MAX_SCENARIO_ELAPSED_S:.2f}s. Shorten the scenario to its core proof."
        )




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
            character_class=str(raw.get("character_class", "")),
            debug_progression=dict(raw.get("debug_progression", {})),
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
        return [scenario for scenario in scenarios if scenario.id != "skill_visual"]
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


def create_session(client: httpx.Client, token: str, world_id: str, seed: str = "", character_id: str = "") -> dict[str, Any]:
    body: dict[str, Any] = {"mode": "solo", "world_id": world_id}
    if seed:
        body["seed"] = seed
    if character_id:
        body["character_id"] = character_id
    resp = client.post("/v0/sessions", headers=auth(token), json=body)
    resp.raise_for_status()
    body = resp.json()
    log("session", body["session_id"], "seed", body["seed"], "world", body.get("world_id"))
    return body


def list_characters(client: httpx.Client, token: str) -> list[dict[str, Any]]:
    resp = client.get("/v0/characters", headers=auth(token))
    resp.raise_for_status()
    return list(resp.json().get("characters", []))


def delete_character(client: httpx.Client, token: str, character_id: str) -> None:
    resp = client.delete(f"/v0/characters/{character_id}", headers=auth(token))
    resp.raise_for_status()


def cleanup_account_characters(client: httpx.Client, email: str, dev_token: str) -> int:
    _, token = dev_login(client, email, dev_token)
    removed = 0
    for char in list_characters(client, token):
        character_id = str(char.get("character_id", ""))
        if not character_id:
            continue
        delete_character(client, token, character_id)
        removed += 1
    return removed


def create_character(client: httpx.Client, token: str, name: str, character_class: str = "") -> dict[str, Any]:
    body = {"name": name}
    if character_class:
        body["character_class"] = character_class
    resp = client.post("/v0/characters", headers=auth(token), json=body)
    resp.raise_for_status()
    return resp.json()


def ensure_character(client: httpx.Client, token: str, name: str, character_class: str = "") -> str:
    chars = list_characters(client, token)
    if character_class:
        for char in chars:
            if str(char.get("character_class", "")) == character_class and str(char.get("name", "")) == name:
                return str(char["character_id"])
    elif chars:
        return str(chars[0]["character_id"])
    return str(create_character(client, token, name, character_class)["character_id"])


def seed_debug_progression(
    client: httpx.Client, token: str, debug_token: str, character_id: str,
    progression: dict[str, Any],
) -> None:
    body = debug_progression_body(progression)
    resp = client.put(
        f"/v0/debug/characters/{character_id}/progression",
        headers={**auth(token), "X-Debug-Token": debug_token},
        json=body,
    )
    resp.raise_for_status()


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
        target_tick = state.last_tick + ticks
        deadline = loop.time() + float(step.get("timeout_s", SLICE_TIMEOUT_S))
        pulse_count = 0
        while state.last_tick < target_tick:
            before = state.last_tick
            env = make_envelope(
                "move_intent",
                session_id,
                state.last_tick,
                {"direction": {"x": 0, "y": 0}, "duration_ticks": 1},
            )
            await ws.send(json.dumps(env))
            await wait_for_accept(ws, state, env["message_id"], loop)
            pulse_deadline = min(deadline, loop.time() + 0.5)
            while state.last_tick <= before and loop.time() <= pulse_deadline:
                await pump_one(ws, state, timeout=0.1)
            if state.last_tick <= before and loop.time() <= deadline:
                direction = {"x": 1 if pulse_count % 2 == 0 else -1, "y": 0}
                pulse_count += 1
                env = make_envelope(
                    "move_intent",
                    session_id,
                    state.last_tick,
                    {"direction": direction, "duration_ticks": 1},
                )
                await ws.send(json.dumps(env))
                await wait_for_accept(ws, state, env["message_id"], loop)
                pulse_deadline = min(deadline, loop.time() + 0.5)
                while state.last_tick <= before and loop.time() <= pulse_deadline:
                    await pump_one(ws, state, timeout=0.1)
            if loop.time() > deadline:
                raise TimeoutError(f"stalled waiting for tick {target_tick}")
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

    if action == "assert_inventory_unique_effect_coverage":
        assert_inventory_unique_effect_coverage(state.inventory, step, "runtime protocol")
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

    if action == "assert_active_weapon_set":
        assert_count_matches(state.active_weapon_set, step, "assert_active_weapon_set")
        return

    if action == "assert_weapon_set_slot_def":
        weapon_set = int(step["weapon_set"])
        slot = str(step["slot"])
        expected_def = str(step["item_def_id"])
        item_id = equipped_slot_id(state, slot, weapon_set)
        if item_id is None:
            raise AssertionError(f"assert_weapon_set_slot_def: set {weapon_set} {slot} is empty, want {expected_def}")
        item = next((i for i in state.inventory if str(i.get("item_instance_id")) == str(item_id)), None)
        if item is None or item.get("item_def_id") != expected_def:
            raise AssertionError(f"assert_weapon_set_slot_def: set {weapon_set} {slot} {item_id} row={item}, want {expected_def}")
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
            monster_pack_leader=bool(step["monster_pack_leader"]) if step.get("monster_pack_leader") is not None else None,
            target_id=str(step["target_id"]) if step.get("target_id") else None,
            timeout_s=float(step.get("timeout_s", SLICE_TIMEOUT_S)),
            fresh_event=bool(step.get("fresh_event", False)),
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
            await wait_for_event(ws, state, str(event_type), loop, timeout_s=float(step.get("timeout_s", SLICE_TIMEOUT_S)))
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
        skipped_ids: set[str] = _remembered_excluded_ids(state, step)
        last_action = 0.0
        pending_message_id = ""
        start_index = len(state.combat_events)
        deadline = loop.time() + SLICE_TIMEOUT_S
        while not any(combat_event_matches(ev, step, state) for ev in state.combat_events[start_index:]):
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
            if reason in {"basic_attack_on_cooldown", "projectile_busy"}:
                continue
            if reason == "player_dead":
                raise AssertionError("action_until_combat_event: player died")
            raise AssertionError(f"action_intent for {monster_def_id or target_id} was rejected: {reason}")
        if not any(combat_event_matches(ev, step, state) for ev in state.combat_events[start_index:]):
            raise AssertionError(f"action_until_combat_event target ended before {combat_event_summary(step)}")
        return

    if action == "wait_for_combat_event":
        start_index = len(state.combat_events)
        deadline = loop.time() + float(step.get("timeout_s", SLICE_TIMEOUT_S))
        while not any(combat_event_matches(ev, step, state) for ev in state.combat_events[start_index:]):
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
            target_id = str(target["id"])
            if target_item_def_id == "gold":
                await wait_for_event(ws, state, "gold_picked_up", loop)
                return
            deadline = loop.time() + SLICE_TIMEOUT_S
            while target_id in state.entities:
                if loop.time() > deadline:
                    raise TimeoutError(f"action_entity stalled waiting for loot pickup {target_item_def_id}")
                await pump_one(ws, state, timeout=0.1)
            return
        event_type = step.get("event_type")
        if event_type:
            await wait_for_event(ws, state, str(event_type), loop, timeout_s=float(step.get("timeout_s", SLICE_TIMEOUT_S)))
        return

    if action == "kill_monsters":
        monster_def_id = str(step.get("monster_def_id", ""))
        monster_pack_leader = step.get("monster_pack_leader")
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
            candidates = find_live_monsters_sorted(state, monster_def_id, monster_pack_leader=monster_pack_leader, exclude_ids=skipped_ids)
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
                    live = find_live_monsters_sorted(state, monster_def_id, monster_pack_leader=monster_pack_leader)
                    if live and len(skipped_ids) >= len(live):
                        raise AssertionError(
                            f"kill_monsters: all live {monster_def_id} targets rejected {reason}; player="
                            f"{(find_player(state) or {}).get('position')}"
                        )
                    break
                if reason == "player_dead":
                    raise AssertionError(f"kill_monsters: player died after killing {killed}/{max_count}")
                if reason == "basic_attack_on_cooldown":
                    continue
                raise AssertionError(f"action_intent for {monster_def_id} {target_id} was rejected: {reason}")
        return
    if action == "move_until_in_range":
        target = resolve_target(state, step)
        stop_distance = float(step.get("stop_distance", WALK_STOP_DISTANCE))
        if target.get("type") == "monster":
            await move_until_entity_in_range(
                ws,
                session_id,
                state,
                str(target["id"]),
                loop,
                stop_distance=stop_distance,
                max_ticks=int(step.get("max_ticks", WALK_MAX_TICKS)),
            )
            return
        await walk_toward(ws, session_id, state, target["position"], loop, stop_distance=stop_distance, max_ticks=int(step.get("max_ticks", WALK_MAX_TICKS)))
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
            if item_def_id == "gold" and "gold_picked_up" in state.seen_events:
                return
            await pump_one(ws, state, timeout=0.1)
            loot = find_loot(state, item_def_id)
        if loot is None:
            if item_def_id == "gold" and "gold_picked_up" in state.seen_events:
                return
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
        loot_id = str(loot["id"])
        await ws.send(json.dumps(make_envelope(
            "action_intent", session_id, state.last_tick, {"target_id": loot_id})))
        log("picking up", item_def_id, "loot", loot_id)
        if item_def_id == "gold":
            await wait_for_event(ws, state, "gold_picked_up", loop)
            return
        wait_started = loop.time()
        last_wait_log = wait_started
        while loot_id in state.entities:
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
        payload = {"item_instance_id": item_id, "slot": slot}
        weapon_set = step.get("weapon_set")
        if weapon_set is not None:
            payload["weapon_set"] = int(weapon_set)
        env = make_envelope("equip_intent", session_id, state.last_tick, payload)
        await ws.send(json.dumps(env))
        log("equipping", item_def_id, item_id)
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        while equipped_slot_id(state, slot, weapon_set) != item_id:
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
        weapon_set = step.get("weapon_set")
        deadline = loop.time() + SLICE_TIMEOUT_S
        if equipped_slot_id(state, slot, weapon_set) is None:
            raise AssertionError(f"unequip_slot: slot {slot} is already empty")
        payload = {"slot": slot}
        if weapon_set is not None:
            payload["weapon_set"] = int(weapon_set)
        await ws.send(json.dumps(make_envelope("unequip_intent", session_id, state.last_tick, payload)))
        log("unequipping", slot)
        while equipped_slot_id(state, slot, weapon_set) is not None:
            if loop.time() > deadline:
                raise TimeoutError(f"unequip_slot stalled waiting for empty {slot}")
            await pump_one(ws, state, timeout=0.1)
        return

    if action == "swap_weapon_set":
        want = int(step.get("active_weapon_set", 1 if state.active_weapon_set == 0 else 0))
        env = make_envelope("swap_weapon_set_intent", session_id, state.last_tick, {})
        await ws.send(json.dumps(env))
        await wait_for_accept(ws, state, env["message_id"], loop)
        deadline = loop.time() + SLICE_TIMEOUT_S
        while state.active_weapon_set != want:
            if loop.time() > deadline:
                raise TimeoutError(f"swap_weapon_set stalled waiting for active set {want}")
            await pump_one(ws, state, timeout=0.1)
        return

    if action == "assign_hotbar":
        slot_index = int(step["slot_index"])
        item_def_id = step.get("item_def_id")
        item_id: str | None = None
        if item_def_id is not None:
            item = find_inventory_item(state.inventory, str(item_def_id))
            if item is None:
                slot = find_hotbar_slot_by_item_def(state.hotbar, str(item_def_id))
                if slot is None:
                    raise AssertionError(f"assign_hotbar: missing inventory or hotbar item {item_def_id}")
                item_id = str(slot["item_instance_id"])
            else:
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
        item_def_id = step.get("item_def_id")
        if item_def_id is not None:
            slot = find_hotbar_slot_by_item_def(state.hotbar, str(item_def_id))
            if slot is None:
                raise AssertionError(f"use_hotbar_slot: missing hotbar item {item_def_id}")
            slot_index = int(slot.get("slot_index", slot_index))
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

    if action == "assert_player_mana":
        player = find_player(state)
        if player is None:
            raise AssertionError("assert_player_mana: player not found in runtime state")
        assert_count_matches(int(player.get("mana", -1)), step, "assert_player_mana")
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

    if action == "allocate_skill_point":
        skill_id = str(step.get("skill_id", "magic_bolt"))
        env = make_envelope(
            "allocate_skill_point_intent",
            session_id,
            state.last_tick,
            {"skill_id": skill_id},
        )
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        expected = step.get("expect_skill_progression")
        if isinstance(expected, dict):
            await wait_for_skill_progression(ws, state, expected, loop)
        return

    if action == "cast_skill":
        skill_id = str(step.get("skill_id", "magic_bolt"))
        payload: dict[str, Any] = {"skill_id": skill_id}
        if bool(step.get("target_self", False)):
            player = find_player(state)
            if player is None:
                raise AssertionError("cast_skill target_self: player not found")
            payload["target_id"] = str(player["id"])
        elif step.get("target_id") is not None:
            payload["target_id"] = str(step["target_id"])
        elif step.get("monster_def_id") is not None:
            target = resolve_target(state, step)
            payload["target_id"] = str(target["id"])
        else:
            direction = step.get("direction", {"x": 1, "y": 0})
            payload["direction"] = {"x": float(direction.get("x", 0)), "y": float(direction.get("y", 0))}
        event_start_index = len(state.events)
        env = make_envelope("cast_skill_intent", session_id, state.last_tick, payload)
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        if step.get("event_type"):
            expected_event: dict[str, Any] = {"event_type": str(step["event_type"])}
            if step.get("skill_id") is not None:
                expected_event["skill_id"] = skill_id
            await wait_for_matching_event(ws, state, expected_event, loop, start_index=event_start_index)
        expected = step.get("expect_skill_cooldown")
        if isinstance(expected, dict):
            await wait_for_skill_cooldown(ws, state, expected, loop)
        return

    if action == "assert_skill_progression":
        assert_skill_progression(state.skill_progression, step, "runtime protocol")
        return

    if action == "wait_skill_progression":
        await wait_for_skill_progression(ws, state, step, loop)
        return

    if action == "assert_skill_cooldown":
        assert_skill_cooldown(state.skill_cooldowns, step, "runtime protocol")
        return

    if action == "wait_skill_cooldown":
        await wait_for_skill_cooldown(ws, state, step, loop)
        return

    if action == "set_skill_bindings": await set_skill_bindings(ws, session_id, state, step, loop, make_envelope, wait_for_accept, wait_for_reject, pump_one, SLICE_TIMEOUT_S); return
    if action == "assert_skill_bindings": assert_skill_bindings(state, step, assert_count_matches); return

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

    if action == "open_bishop_service":
        target = find_interactable(state, str(step.get("interactable_def_id", "town_bishop")))
        if target is None:
            raise AssertionError(f"open_bishop_service: missing town_bishop on level {state.current_level}")
        await walk_toward(
            ws,
            session_id,
            state,
            target["position"],
            loop,
            stop_distance=float(step.get("stop_distance", WALK_STOP_DISTANCE)),
            max_ticks=derived_walk_max_ticks(state, target["position"], int(step.get("max_ticks", WALK_MAX_TICKS))),
        )
        start_index = len(state.events)
        env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": str(target["id"])})
        await ws.send(json.dumps(env))
        await wait_for_accept(ws, state, env["message_id"], loop)
        await wait_for_event(ws, state, "bishop_service_opened", loop, start_index=start_index)
        return

    if action == "respec_bishop":
        target = find_interactable(state, str(step.get("interactable_def_id", "town_bishop")))
        if target is None:
            raise AssertionError(f"respec_bishop: missing town_bishop on level {state.current_level}")
        before_gold = state.gold
        start_index = len(state.events)
        env = make_envelope(
            "bishop_respec_intent",
            session_id,
            state.last_tick,
            {"bishop_entity_id": str(target["id"])},
        )
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        await wait_for_event(ws, state, "bishop_respec", loop, start_index=start_index)
        state.last_gold_before_action = before_gold
        state.last_gold_after_action = state.gold
        expected_progression = step.get("expect_progression")
        if isinstance(expected_progression, dict):
            await wait_for_character_progression(ws, state, expected_progression, loop)
        expected_skill = step.get("expect_skill_progression")
        if isinstance(expected_skill, dict):
            await wait_for_skill_progression(ws, state, expected_skill, loop)
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
        assert_shop_event_details([event], step, "buy_shop_offer")
        if len(state.inventory) <= before_inventory_count:
            raise AssertionError(f"buy_shop_offer: inventory did not grow after {offer}")
        return

    if action == "reroll_shop":
        shop_id = str(step.get("shop_id", "town_mystery_seller"))
        target = find_interactable(state, str(step.get("interactable_def_id", "town_mystery_seller")))
        if target is None:
            raise AssertionError(f"reroll_shop: missing shop entity on level {state.current_level}")
        before_gold = state.gold
        start_index = len(state.shop_events)
        env = make_envelope(
            "shop_reroll_intent",
            session_id,
            state.last_tick,
            {"shop_entity_id": str(target["id"])},
        )
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        event = await wait_for_shop_event(ws, state, "shop_reroll", loop, shop_id=shop_id, start_index=start_index)
        state.last_gold_before_action = before_gold
        state.last_gold_after_action = state.gold
        if int(event.get("price", 0)) <= 0:
            raise AssertionError(f"reroll_shop: event missing positive price: {event}")
        if int(event.get("total_gold", state.gold)) != state.gold:
            raise AssertionError(f"reroll_shop: event total_gold {event.get('total_gold')} != state gold {state.gold}")
        offers = list(state.shop_offers.get(shop_id, {}).keys())
        if not offers:
            raise AssertionError(f"reroll_shop: cached offers missing after event {event}")
        if "|reroll:" not in str(event.get("refresh_key", "")):
            raise AssertionError(f"reroll_shop: event missing reroll refresh key: {event}")
        assert_shop_event_details([event], step, "reroll_shop")
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
        matches = filtered_shop_events(state, step)
        assert_count_matches(len(matches), step, "assert_shop_event", f": {matches}")
        assert_shop_event_details(matches, step, "assert_shop_event")
        return

    if action == "open_stash":
        stash_id = str(step.get("stash_id", "account_stash"))
        interactable_def_id = str(step.get("interactable_def_id", "town_stash"))
        target = find_interactable(state, interactable_def_id)
        if target is None:
            raise AssertionError(f"open_stash: missing {interactable_def_id} on level {state.current_level}")
        await walk_toward(
            ws,
            session_id,
            state,
            target["position"],
            loop,
            stop_distance=float(step.get("stop_distance", WALK_STOP_DISTANCE)),
            max_ticks=derived_walk_max_ticks(state, target["position"], int(step.get("max_ticks", WALK_MAX_TICKS))),
        )
        start_index = len(state.stash_events)
        env = make_envelope("action_intent", session_id, state.last_tick, {"target_id": str(target["id"])})
        await ws.send(json.dumps(env))
        await wait_for_accept(ws, state, env["message_id"], loop)
        await wait_for_stash_event(ws, state, "stash_opened", loop, stash_id=stash_id, start_index=start_index)
        return

    if action == "deposit_stash_item":
        stash_id = str(step.get("stash_id", "account_stash"))
        target = find_interactable(state, str(step.get("interactable_def_id", "town_stash")))
        if target is None:
            raise AssertionError(f"deposit_stash_item: missing stash entity on level {state.current_level}")
        item = select_inventory_item(state.inventory, step)
        item_id = str(item["item_instance_id"])
        before_stash_count = len(state.stash_items)
        start_index = len(state.stash_events)
        env = make_envelope(
            "stash_deposit_item_intent",
            session_id,
            state.last_tick,
            {"stash_entity_id": str(target["id"]), "item_instance_id": item_id},
        )
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        event = await wait_for_stash_event(ws, state, "stash_item_deposited", loop, stash_id=stash_id, start_index=start_index)
        state.last_gold_before_action = None
        state.last_gold_after_action = None
        if str(event.get("item_instance_id")) != item_id:
            raise AssertionError(f"deposit_stash_item: event item {event.get('item_instance_id')} != {item_id}")
        if find_inventory_item_by_instance(state.inventory, item_id) is not None:
            raise AssertionError(f"deposit_stash_item: item {item_id} still in inventory")
        if len(state.stash_items) <= before_stash_count:
            raise AssertionError(f"deposit_stash_item: stash did not grow after depositing {item_id}: {state.stash_items}")
        return

    if action == "withdraw_stash_item":
        stash_id = str(step.get("stash_id", "account_stash"))
        target = find_interactable(state, str(step.get("interactable_def_id", "town_stash")))
        if target is None:
            raise AssertionError(f"withdraw_stash_item: missing stash entity on level {state.current_level}")
        item = select_stash_item(state, step)
        stash_item_id = str(item["stash_item_id"])
        before_inventory_count = len(state.inventory)
        start_index = len(state.stash_events)
        env = make_envelope(
            "stash_withdraw_item_intent",
            session_id,
            state.last_tick,
            {"stash_entity_id": str(target["id"]), "stash_item_id": stash_item_id},
        )
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        event = await wait_for_stash_event(ws, state, "stash_item_withdrawn", loop, stash_id=stash_id, start_index=start_index)
        state.last_gold_before_action = None
        state.last_gold_after_action = None
        if str(event.get("stash_item_id")) != stash_item_id:
            raise AssertionError(f"withdraw_stash_item: event stash item {event.get('stash_item_id')} != {stash_item_id}")
        if find_stash_item_by_id(state.stash_items, stash_item_id) is not None:
            raise AssertionError(f"withdraw_stash_item: stash item {stash_item_id} still in stash")
        if len(state.inventory) <= before_inventory_count:
            raise AssertionError(f"withdraw_stash_item: inventory did not grow after withdrawing {stash_item_id}")
        return

    if action == "take_unique_chest_item":
        chest_entity_id = step.get("chest_entity_id")
        target = find_interactable(state, str(step.get("interactable_def_id", "town_unique_chest"))) if chest_entity_id is None else {"id": str(chest_entity_id)}
        if target is None:
            raise AssertionError(f"take_unique_chest_item: missing unique chest on level {state.current_level}")
        item = select_stash_item(state, step)
        chest_item_id = str(item["stash_item_id"])
        before_inventory_count = len(state.inventory)
        start_index = len(state.stash_events)
        env = make_envelope(
            "unique_chest_take_item_intent",
            session_id,
            state.last_tick,
            {"chest_entity_id": str(target["id"]), "chest_item_id": chest_item_id},
        )
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        event = await wait_for_stash_event(ws, state, "unique_chest_item_taken", loop, stash_id="unique_test_chest", start_index=start_index)
        if str(event.get("stash_item_id")) != chest_item_id:
            raise AssertionError(f"take_unique_chest_item: event chest item {event.get('stash_item_id')} != {chest_item_id}")
        if len(state.inventory) <= before_inventory_count:
            raise AssertionError(f"take_unique_chest_item: inventory did not grow after taking {chest_item_id}")
        return

    if action == "deposit_stash_gold":
        amount = int(step["amount"])
        stash_id = str(step.get("stash_id", "account_stash"))
        target = find_interactable(state, str(step.get("interactable_def_id", "town_stash")))
        if target is None:
            raise AssertionError(f"deposit_stash_gold: missing stash entity on level {state.current_level}")
        before_gold = state.gold
        start_index = len(state.stash_events)
        env = make_envelope(
            "stash_deposit_gold_intent",
            session_id,
            state.last_tick,
            {"stash_entity_id": str(target["id"]), "amount": amount},
        )
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        event = await wait_for_stash_event(ws, state, "stash_gold_deposited", loop, stash_id=stash_id, start_index=start_index)
        state.last_gold_before_action = before_gold
        state.last_gold_after_action = state.gold
        if int(event.get("amount", 0)) != amount:
            raise AssertionError(f"deposit_stash_gold: event amount {event.get('amount')} != {amount}")
        return

    if action == "withdraw_stash_gold":
        amount = int(step["amount"])
        stash_id = str(step.get("stash_id", "account_stash"))
        target = find_interactable(state, str(step.get("interactable_def_id", "town_stash")))
        if target is None:
            raise AssertionError(f"withdraw_stash_gold: missing stash entity on level {state.current_level}")
        before_gold = state.gold
        start_index = len(state.stash_events)
        env = make_envelope(
            "stash_withdraw_gold_intent",
            session_id,
            state.last_tick,
            {"stash_entity_id": str(target["id"]), "amount": amount},
        )
        await ws.send(json.dumps(env))
        expect_reject = step.get("expect_reject")
        if expect_reject:
            await wait_for_reject(ws, state, env["message_id"], str(expect_reject), loop)
            return
        await wait_for_accept(ws, state, env["message_id"], loop)
        event = await wait_for_stash_event(ws, state, "stash_gold_withdrawn", loop, stash_id=stash_id, start_index=start_index)
        state.last_gold_before_action = before_gold
        state.last_gold_after_action = state.gold
        if int(event.get("amount", 0)) != amount:
            raise AssertionError(f"withdraw_stash_gold: event amount {event.get('amount')} != {amount}")
        return

    if action == "assert_stash_item_count":
        items = filtered_stash_items(state.stash_items, step)
        assert_count_matches(len(items), step, "assert_stash_item_count", f": {items}")
        return

    if action == "assert_stash_gold":
        assert_count_matches(state.stash_gold, step, "assert_stash_gold")
        return

    if action == "assert_stash_event":
        event_type = str(step["event_type"])
        stash_id = str(step.get("stash_id", "account_stash"))
        matches = [
            event for event in state.stash_events
            if event.get("event_type") == event_type and str(event.get("stash_id", "")) == stash_id
        ]
        if step.get("stash_item_id") is not None:
            matches = [event for event in matches if str(event.get("stash_item_id", "")) == str(step["stash_item_id"])]
        assert_count_matches(len(matches), step, "assert_stash_event", f": {matches}")
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


def _runtime_context() -> BotContext:
    return BotContext(pump_one=pump_one)


async def walk_toward(
    ws,
    session_id: str,
    state: RuntimeState,
    target_pos: dict[str, Any],
    loop,
    max_ticks: int = WALK_MAX_TICKS,
    stop_distance: float = WALK_STOP_DISTANCE,
) -> None:
    from tools.bot.movement_runtime import walk_toward as walk_toward_impl

    await walk_toward_impl(ws, session_id, state, target_pos, loop, max_ticks, stop_distance, ctx=_runtime_context())


async def move_until_entity_in_range(
    ws,
    session_id: str,
    state: RuntimeState,
    target_id: str,
    loop,
    *,
    stop_distance: float,
    max_ticks: int = WALK_MAX_TICKS,
) -> None:
    from tools.bot.movement_runtime import move_until_entity_in_range as move_until_entity_in_range_impl

    await move_until_entity_in_range_impl(
        ws,
        session_id,
        state,
        target_id,
        loop,
        stop_distance=stop_distance,
        max_ticks=max_ticks,
        ctx=_runtime_context(),
    )


def range_candidate_positions(
    player_pos: dict[str, Any],
    target_pos: dict[str, Any],
    stop_distance: float,
) -> list[dict[str, float]]:
    from tools.bot.movement_runtime import range_candidate_positions as range_candidate_positions_impl

    return range_candidate_positions_impl(player_pos, target_pos, stop_distance)


def derived_walk_max_ticks(state: RuntimeState, target_pos: dict[str, Any], requested: int) -> int:
    from tools.bot.movement_runtime import derived_walk_max_ticks as derived_walk_max_ticks_impl

    return derived_walk_max_ticks_impl(state, target_pos, requested)


async def move_to_position(
    ws,
    session_id: str,
    state: RuntimeState,
    target_pos: dict[str, Any],
    loop,
    max_ticks: int = WALK_MAX_TICKS,
    stop_distance: float = WALK_STOP_DISTANCE,
) -> None:
    from tools.bot.movement_runtime import move_to_position as move_to_position_impl

    await move_to_position_impl(ws, session_id, state, target_pos, loop, max_ticks, stop_distance, ctx=_runtime_context())


async def attack_until_monster_event(
    ws,
    session_id: str,
    state: RuntimeState,
    loop,
    event_type: str,
    *,
    monster_def_id: str | None = None, rarity: str | None = None, is_boss: bool | None = None,
    monster_pack_leader: bool | None = None, target_id: str | None = None, timeout_s: float = SLICE_TIMEOUT_S,
    fresh_event: bool = False,
) -> None:
    deadline = loop.time() + timeout_s; start_event_count = sum(1 for ev in state.events if ev.get("event_type") == event_type)
    skipped_ids: set[str] = set()
    active_target_id = target_id
    pending_message_id = ""
    last_action = 0.0
    for _ in range(5):
        if find_player(state) is not None:
            break
        await pump_one(ws, state, timeout=0.1)
    while event_type not in state.seen_events or (fresh_event and sum(1 for ev in state.events if ev.get("event_type") == event_type) <= start_event_count):
        if loop.time() > deadline:
            raise TimeoutError(f"attack_until_monster_event stalled waiting for {event_type}")
        if (monster_def_id or is_boss is not None or monster_pack_leader is not None) and (active_target_id is None or active_target_id in skipped_ids):
            candidates = find_live_monsters_sorted(
                state,
                monster_def_id or "",
                rarity=rarity,
                is_boss=is_boss,
                monster_pack_leader=monster_pack_leader,
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
        if reason in {"basic_attack_on_cooldown", "projectile_busy"}:
            continue
        if reason == "invalid_target" and monster_def_id and monster_def_id in state.killed_monster_def_ids:
            continue
        if reason == "player_dead":
            raise AssertionError(f"attack_until_monster_event: player died waiting for {event_type}")
        label = monster_def_id or active_target_id
        raise AssertionError(f"action_intent for {label} was rejected: {reason}")


async def wait_for_player_move_or_accept(ws, state: RuntimeState, before: dict[str, Any], message_id: str, loop) -> bool:
    from tools.bot.movement_runtime import wait_for_player_move_or_accept as wait_for_player_move_or_accept_impl

    return await wait_for_player_move_or_accept_impl(ws, state, before, message_id, loop, ctx=_runtime_context())


async def wait_for_tick_advance(ws, state: RuntimeState, loop) -> None:
    from tools.bot.wait_runtime import wait_for_tick_advance as wait_for_tick_advance_impl

    await wait_for_tick_advance_impl(ws, state, loop, helpers=globals())


async def wait_for_tick(ws, state: RuntimeState, target_tick: int, loop) -> None:
    from tools.bot.wait_runtime import wait_for_tick as wait_for_tick_impl

    await wait_for_tick_impl(ws, state, target_tick, loop, helpers=globals())


async def wait_for_accept(ws, state: RuntimeState, message_id: str, loop) -> None:
    from tools.bot.wait_runtime import wait_for_accept as wait_for_accept_impl

    await wait_for_accept_impl(ws, state, message_id, loop, helpers=globals())


async def wait_for_reject(ws, state: RuntimeState, message_id: str, reason: str, loop) -> None:
    from tools.bot.wait_runtime import wait_for_reject as wait_for_reject_impl

    await wait_for_reject_impl(ws, state, message_id, reason, loop, helpers=globals())


async def wait_for_event(ws, state: RuntimeState, event_type: str, loop, *, timeout_s: float = SLICE_TIMEOUT_S, start_index: int = 0) -> None:
    from tools.bot.wait_runtime import wait_for_event as wait_for_event_impl

    await wait_for_event_impl(ws, state, event_type, loop, timeout_s=timeout_s, start_index=start_index, helpers=globals())


async def wait_for_matching_event(
    ws,
    state: RuntimeState,
    expected: dict[str, Any],
    loop,
    *,
    timeout_s: float = SLICE_TIMEOUT_S,
    start_index: int = 0,
) -> None:
    from tools.bot.wait_runtime import wait_for_matching_event as wait_for_matching_event_impl

    await wait_for_matching_event_impl(
        ws,
        state,
        expected,
        loop,
        timeout_s=timeout_s,
        start_index=start_index,
        helpers=globals(),
    )


async def wait_for_shop_event(
    ws,
    state: RuntimeState,
    event_type: str,
    loop,
    *,
    shop_id: str = "town_vendor",
    start_index: int = 0,
) -> dict[str, Any]:
    from tools.bot.wait_runtime import wait_for_shop_event as wait_for_shop_event_impl

    return await wait_for_shop_event_impl(
        ws, state, event_type, loop, shop_id=shop_id, start_index=start_index, helpers=globals()
    )


async def wait_for_stash_event(
    ws,
    state: RuntimeState,
    event_type: str,
    loop,
    *,
    stash_id: str = "account_stash",
    start_index: int = 0,
) -> dict[str, Any]:
    from tools.bot.wait_runtime import wait_for_stash_event as wait_for_stash_event_impl

    return await wait_for_stash_event_impl(
        ws, state, event_type, loop, stash_id=stash_id, start_index=start_index, helpers=globals()
    )


def combat_event_matches(event: dict[str, Any], expected: dict[str, Any], state: RuntimeState | None = None) -> bool:
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
        "skill_id",
        "weapon_slot",
        "damage_type",
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
    for key, event_key in (
        ("max_damage", "damage"),
        ("max_raw_damage", "raw_damage"),
        ("max_mitigated_damage", "mitigated_damage"),
    ):
        if key in expected and int(event.get(event_key, 999999)) > int(expected[key]):
            return False
    if not combat_event_entity_matches(event, expected, state, "source"):
        return False
    if not combat_event_entity_matches(event, expected, state, "target"):
        return False
    return True


def event_matches(event: dict[str, Any], expected: dict[str, Any]) -> bool:
    for key in (
        "event_type",
        "entity_id",
        "source_entity_id",
        "target_entity_id",
        "boss_template_id",
        "skill_id",
        "damage_type",
        "pattern_id",
        "phase_kind",
        "reason",
        "state",
        "service",
    ):
        if key in expected and str(event.get(key, "")) != str(expected[key]):
            return False
    for key in (
        "phase_index",
        "duration_ticks",
        "from_level",
        "to_level",
        "rank",
        "mana",
        "heal",
        "damage",
        "raw_damage",
        "mitigated_damage",
        "amount",
        "remaining_ticks",
        "total_ticks",
        "price",
        "total_gold",
        "unspent_stat_points",
        "unspent_skill_points",
    ):
        if key in expected and int(event.get(key, -999999)) != int(expected[key]):
            return False
    if "affordable" in expected and bool(event.get("affordable", False)) != bool(expected["affordable"]):
        return False
    return True


def event_summary(expected: dict[str, Any]) -> str:
    parts = []
    for key in (
        "event_type",
        "entity_id",
        "source_entity_id",
        "target_entity_id",
        "boss_template_id",
        "skill_id",
        "damage_type",
        "rank",
        "mana",
        "heal",
        "damage",
        "raw_damage",
        "mitigated_damage",
        "amount",
        "remaining_ticks",
        "total_ticks",
        "pattern_id",
        "phase_kind",
        "phase_index",
        "duration_ticks",
        "reason",
        "state",
        "service",
        "price",
        "total_gold",
        "unspent_stat_points",
        "unspent_skill_points",
        "affordable",
        "from_level",
        "to_level",
    ):
        if key in expected:
            parts.append(f"{key}={expected[key]}")
    return ", ".join(parts) or str(expected)


def combat_event_entity_matches(
    event: dict[str, Any],
    expected: dict[str, Any],
    state: RuntimeState | None,
    prefix: str,
) -> bool:
    selector: dict[str, Any] = {}
    for expected_key, selector_key in (
        (f"{prefix}_entity_type", "entity_type"),
        (f"{prefix}_monster_def_id", "monster_def_id"),
        (f"{prefix}_is_boss", "is_boss"),
        (f"{prefix}_boss_template_id", "boss_template_id"), (f"{prefix}_monster_pack_leader", "monster_pack_leader"),
    ):
        if expected_key in expected:
            selector[selector_key] = expected[expected_key]
    if not selector:
        return True
    if state is None:
        return False
    entity_id = str(event.get(f"{prefix}_entity_id", ""))
    event_monster_def_id = str(event.get(f"{prefix}_monster_def_id", ""))
    if len(selector) == 1 and "monster_def_id" in selector and event_monster_def_id:
        return event_monster_def_id == str(selector["monster_def_id"])
    entity = state.entities.get(entity_id)
    if entity is None:
        return False
    return entity_matches_selector(entity, selector)


def combat_event_summary(expected: dict[str, Any]) -> str:
    parts = []
    for key in (
        "event_type",
        "outcome",
        "damage",
        "min_damage",
        "blocked",
        "critical",
        "source_entity_type",
        "source_monster_def_id",
        "target_entity_type",
        "target_monster_def_id",
    ):
        if key in expected:
            parts.append(f"{key}={expected[key]}")
    return ", ".join(parts) or str(expected)


async def wait_for_level_change(ws, state: RuntimeState, previous_level: int, loop) -> None:
    from tools.bot.wait_runtime import wait_for_level_change as wait_for_level_change_impl

    await wait_for_level_change_impl(ws, state, previous_level, loop, helpers=globals())


async def wait_for_teleporter_discovery(ws, state: RuntimeState, level: int, loop) -> None:
    from tools.bot.wait_runtime import wait_for_teleporter_discovery as wait_for_teleporter_discovery_impl

    await wait_for_teleporter_discovery_impl(ws, state, level, loop, helpers=globals())


async def wait_for_character_progression(ws, state: RuntimeState, expected: dict[str, Any], loop) -> None:
    from tools.bot.wait_runtime import wait_for_character_progression as wait_for_character_progression_impl

    await wait_for_character_progression_impl(ws, state, expected, loop, helpers=globals())


async def wait_for_skill_progression(ws, state: RuntimeState, expected: dict[str, Any], loop) -> None:
    from tools.bot.wait_runtime import wait_for_skill_progression as wait_for_skill_progression_impl

    await wait_for_skill_progression_impl(ws, state, expected, loop, helpers=globals())


async def wait_for_skill_cooldown(ws, state: RuntimeState, expected: dict[str, Any], loop) -> None:
    from tools.bot.wait_runtime import wait_for_skill_cooldown as wait_for_skill_cooldown_impl

    await wait_for_skill_cooldown_impl(ws, state, expected, loop, helpers=globals())


async def wait_for_player_position(
    ws,
    state: RuntimeState,
    x: float,
    y: float,
    tolerance: float,
    loop,
) -> None:
    from tools.bot.wait_runtime import wait_for_player_position as wait_for_player_position_impl

    await wait_for_player_position_impl(ws, state, x, y, tolerance, loop, helpers=globals())


async def pump_one(ws, state: RuntimeState, timeout: float) -> None:
    from tools.bot.wait_runtime import pump_one as pump_one_impl

    await pump_one_impl(ws, state, timeout, helpers=globals())


def ingest_message(m: dict[str, Any], state: RuntimeState) -> None:
    from tools.bot.state_ingest import ingest_message as ingest_message_impl

    ingest_message_impl(m, state, helpers=globals())


def ingest_snapshot(payload: dict[str, Any], state: RuntimeState) -> None:
    from tools.bot.state_ingest import ingest_snapshot as ingest_snapshot_impl

    ingest_snapshot_impl(payload, state, helpers=globals())


def parse_discovered_teleporters(payload: dict[str, Any]) -> dict[int, bool]:
    from tools.bot.state_ingest import parse_discovered_teleporters as parse_discovered_teleporters_impl

    return parse_discovered_teleporters_impl(payload, helpers=globals())


def upsert_hotbar(state: RuntimeState, slot_index: int, item_instance_id: Any, item: dict[str, Any] | None = None) -> None:
    from tools.bot.state_ingest import upsert_hotbar as upsert_hotbar_impl

    upsert_hotbar_impl(state, slot_index, item_instance_id, item, helpers=globals())


def decay_skill_cooldowns(state: RuntimeState, ticks: int) -> None:
    from tools.bot.state_ingest import decay_skill_cooldowns as decay_skill_cooldowns_impl

    decay_skill_cooldowns_impl(state, ticks, helpers=globals())


def hotbar_item_id(hotbar: list[dict[str, Any]], slot_index: int) -> str | None:
    for slot in hotbar:
        if int(slot.get("slot_index", -1)) == slot_index:
            raw = slot.get("item_instance_id")
            return None if raw is None else str(raw)
    return None

def find_hotbar_slot_by_item_def(hotbar: list[dict[str, Any]], item_def_id: str) -> dict[str, Any] | None:
    for slot in hotbar:
        item = slot.get("item")
        if isinstance(item, dict) and str(item.get("item_def_id", "")) == item_def_id:
            return slot
    return None


def clear_active_level_state(state: RuntimeState) -> None:
    from tools.bot.state_ingest import clear_active_level_state as clear_active_level_state_impl

    clear_active_level_state_impl(state, helpers=globals())


def track_initial_entity_position(state: RuntimeState, entity: dict[str, Any]) -> None:
    from tools.bot.state_ingest import track_initial_entity_position as track_initial_entity_position_impl

    track_initial_entity_position_impl(state, entity, helpers=globals())


def track_initial_monster_position(state: RuntimeState, entity: dict[str, Any]) -> None:
    from tools.bot.state_ingest import track_initial_monster_position as track_initial_monster_position_impl

    track_initial_monster_position_impl(state, entity, helpers=globals())


def update_runtime_distances(state: RuntimeState) -> None:
    from tools.bot.state_ingest import update_runtime_distances as update_runtime_distances_impl

    update_runtime_distances_impl(state, helpers=globals())


def upsert_inventory(state: RuntimeState, item: dict[str, Any]) -> None:
    from tools.bot.state_ingest import upsert_inventory as upsert_inventory_impl

    upsert_inventory_impl(state, item, helpers=globals())


def remove_inventory_item(state: RuntimeState, item_instance_id: str) -> None:
    from tools.bot.state_ingest import remove_inventory_item as remove_inventory_item_impl

    remove_inventory_item_impl(state, item_instance_id, helpers=globals())


def upsert_stash_item(state: RuntimeState, item: dict[str, Any]) -> None:
    from tools.bot.state_ingest import upsert_stash_item as upsert_stash_item_impl

    upsert_stash_item_impl(state, item, helpers=globals())


def remove_stash_item(state: RuntimeState, stash_item_id: str) -> None:
    from tools.bot.state_ingest import remove_stash_item as remove_stash_item_impl

    remove_stash_item_impl(state, stash_item_id, helpers=globals())


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
    alive: bool | None = True,
    exclude_ids: set[str] | None = None,
    skip: int = 0,
) -> dict[str, Any] | None:
    monsters = find_live_monsters_sorted(state, monster_def_id, rarity=rarity, is_boss=is_boss, alive=alive, exclude_ids=exclude_ids)
    skip = max(skip, 0)
    if len(monsters) <= skip:
        return None
    return monsters[skip]
def find_live_monsters_sorted(
    state: RuntimeState,
    monster_def_id: str,
    rarity: str | None = None,
    is_boss: bool | None = None,
    monster_pack_leader: bool | None = None,
    alive: bool | None = True,
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
            or (monster_pack_leader is not None and bool(entity.get("monster_pack_leader", False)) != bool(monster_pack_leader))
            or entity_id in excluded
        ):
            continue
        if alive is not None and (not isinstance(entity.get("hp"), int) or entity["hp"] > 0) != bool(alive):
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
            if bool(step.get("allow_missing_target")):
                return {"id": str(step["target_id"])}
            raise AssertionError(f"{step.get('action')}: target not found: {step['target_id']}")
        return target
    if step.get("monster_def_id"):
        rarity = str(step["rarity"]) if step.get("rarity") is not None else None
        exclude_ids = _remembered_excluded_ids(state, step)
        target = find_nearest_monster(state, str(step["monster_def_id"]), rarity,
                                      bool(step["is_boss"]) if step.get("is_boss") is not None else None,
                                      bool(step["alive"]) if step.get("alive") is not None else True,
                                      exclude_ids=exclude_ids, skip=int(step.get("target_skip", 0)))
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


def _remembered_excluded_ids(state: RuntimeState, step: dict[str, Any]) -> set[str]:
    excluded: set[str] = set()
    raw_keys = step.get("exclude_remembered", [])
    if not isinstance(raw_keys, list):
        raw_keys = [raw_keys]
    for key in raw_keys:
        remembered_id = state.remembered_entity_ids.get(str(key), "")
        if remembered_id:
            excluded.add(remembered_id)
    return excluded


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


def filtered_inventory_items(inventory: list[dict[str, Any]], step: dict[str, Any]) -> list[dict[str, Any]]:
    items = list(inventory)
    if step.get("item_instance_id") is not None:
        items = [item for item in items if str(item.get("item_instance_id", "")) == str(step["item_instance_id"])]
    if step.get("item_def_id") is not None:
        items = [item for item in items if str(item.get("item_def_id", "")) == str(step["item_def_id"])]
    if step.get("item_template_id") is not None:
        items = [item for item in items if str(item.get("item_template_id", "")) == str(step["item_template_id"])]
    if step.get("display_name") is not None:
        items = [item for item in items if str(item.get("display_name", "")) == str(step["display_name"])]
    if step.get("rolled") is not None:
        want_rolled = bool(step["rolled"])
        items = [item for item in items if bool(item.get("item_template_id")) == want_rolled]
    if step.get("equipped") is not None:
        items = [item for item in items if bool(item.get("equipped")) == bool(step["equipped"])]
    items.sort(key=lambda item: str(item.get("item_instance_id", "")))
    return items


def select_inventory_item(inventory: list[dict[str, Any]], step: dict[str, Any]) -> dict[str, Any]:
    items = filtered_inventory_items(inventory, step)
    if not items:
        raise AssertionError(f"{step.get('action')}: no matching inventory item for {step}; inventory={inventory}")
    index = int(step.get("bag_index", 0))
    if index < 0 or index >= len(items):
        raise AssertionError(f"{step.get('action')}: bag_index {index} out of range for {items}")
    return items[index]


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
    if step.get("source_depth_min") is not None:
        offers = [offer for offer in offers if shop_row_source_depth_bounds(offer)[0] >= int(step["source_depth_min"])]
    if step.get("source_depth_max") is not None:
        offers = [offer for offer in offers if shop_row_source_depth_bounds(offer)[1] <= int(step["source_depth_max"])]
    if step.get("concealed") is not None:
        want_concealed = bool(step["concealed"])
        offers = [offer for offer in offers if bool(offer.get("concealed")) == want_concealed]
    if bool(step.get("affordable")):
        reserve_gold = int(step.get("reserve_gold", 0))
        budget = max(0, state.gold - reserve_gold)
        offers = [offer for offer in offers if int(offer.get("buy_price", 0)) <= budget]
    offers.sort(key=lambda offer: (int(offer.get("buy_price", 0)), str(offer.get("offer_id", ""))))
    return offers


def shop_row_source_depth_bounds(row: dict[str, Any]) -> tuple[int, int]:
    if row.get("source_depth") is not None:
        depth = int(row.get("source_depth", 0))
        return depth, depth
    return int(row.get("source_depth_min", 0)), int(row.get("source_depth_max", 0))


SHOP_IDENTITY_FIELDS = (
    "item_def_id",
    "item_template_id",
    "display_name",
    "rarity",
    "rolled_stats",
    "requirements",
    "requirement_status",
    "requirements_met",
    "effect_ids",
    "summary_lines",
    "comparison",
    "equip_preview",
)


def filtered_shop_events(state: RuntimeState, step: dict[str, Any]) -> list[dict[str, Any]]:
    event_type = str(step["event_type"])
    shop_id = str(step.get("shop_id", "town_vendor"))
    matches = [
        event for event in state.shop_events
        if event.get("event_type") == event_type and str(event.get("shop_id", "")) == shop_id
    ]
    if step.get("offer_id") is not None:
        matches = [event for event in matches if str(event.get("offer_id", "")) == str(step["offer_id"])]
    if step.get("offer_kind") is not None:
        matches = [event for event in matches if shop_event_offer_kind(event) == str(step["offer_kind"])]
    if step.get("requires_revealed_item"):
        matches = [event for event in matches if shop_event_revealed_item(event)]
    if step.get("rarity_in") is not None:
        allowed = {str(rarity) for rarity in step["rarity_in"]}
        matches = [event for event in matches if str((event.get("item") or {}).get("rarity", "")) in allowed]
    if step.get("item_slot") is not None:
        matches = [event for event in matches if str((event.get("item") or {}).get("slot", "")) == str(step["item_slot"])]
    if step.get("item_category") is not None:
        matches = [event for event in matches if str((event.get("item") or {}).get("category", "")) == str(step["item_category"])]
    return matches


def shop_event_offer_kind(event: dict[str, Any]) -> str:
    explicit = str(event.get("offer_kind", ""))
    if explicit:
        return explicit
    offer_id = str(event.get("offer_id", ""))
    if ":" in offer_id:
        return offer_id.split(":", 1)[0]
    return ""


def shop_event_revealed_item(event: dict[str, Any]) -> bool:
    item = event.get("item")
    if not isinstance(item, dict):
        return False
    return bool(item.get("item_instance_id") and item.get("item_template_id") and item.get("display_name") and item.get("rarity"))


def assert_shop_event_details(events: list[dict[str, Any]], step: dict[str, Any], label: str) -> None:
    if step.get("requires_revealed_item"):
        missing = [event for event in events if not shop_event_revealed_item(event)]
        if missing:
            raise AssertionError(f"{label}: events missing revealed item payload: {missing}")
    if step.get("rarity_in") is not None:
        allowed = {str(rarity) for rarity in step["rarity_in"]}
        mismatched = [
            event for event in events
            if str((event.get("item") or {}).get("rarity", "")) not in allowed
        ]
        if mismatched:
            raise AssertionError(f"{label}: revealed rarity not in {sorted(allowed)}: {mismatched}")


def filtered_shop_sell_appraisals(state: RuntimeState, step: dict[str, Any]) -> list[dict[str, Any]]:
    shop_id = str(step.get("shop_id", "town_vendor"))
    rows = list(state.shop_sell_appraisals.get(shop_id, {}).values())
    if step.get("item_instance_id") is not None:
        rows = [row for row in rows if str(row.get("item_instance_id")) == str(step["item_instance_id"])]
    if step.get("item_def_id") is not None:
        rows = [row for row in rows if str(row.get("item_def_id")) == str(step["item_def_id"])]
    if step.get("item_template_id") is not None:
        rows = [row for row in rows if str(row.get("item_template_id")) == str(step["item_template_id"])]
    if step.get("rolled") is not None:
        want_rolled = bool(step["rolled"])
        rows = [row for row in rows if bool(row.get("item_template_id")) == want_rolled]
    rows.sort(key=lambda row: (int(row.get("sell_price", 0)), str(row.get("item_instance_id", ""))))
    return rows


def select_shop_offer(state: RuntimeState, step: dict[str, Any]) -> dict[str, Any]:
    offers = filtered_shop_offers(state, step)
    if not offers:
        shop_id = str(step.get("shop_id", "town_vendor"))
        reserve_gold = int(step.get("reserve_gold", 0))
        budget = max(0, state.gold - reserve_gold)
        raise AssertionError(
            f"{step.get('action')}: no matching shop offers in {shop_id}: {step}; "
            f"gold={state.gold} budget={budget}; have={state.shop_offers.get(shop_id, {})}"
        )
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
    if step.get("requires_concealed"):
        missing = [row for row in rows if row.get("concealed") is not True]
        if missing:
            raise AssertionError(f"{label}: rows are not concealed: {missing}")
    if step.get("requires_mystery_label"):
        missing = [row for row in rows if not row.get("mystery_label")]
        if missing:
            raise AssertionError(f"{label}: rows missing mystery_label: {missing}")
    if step.get("forbids_item_identity"):
        leaking = [
            row for row in rows
            if any(field in row for field in SHOP_IDENTITY_FIELDS)
        ]
        if leaking:
            raise AssertionError(f"{label}: rows expose hidden item fields: {leaking}")
    if step.get("source_depth_min") is not None:
        minimum = int(step["source_depth_min"])
        missing = [row for row in rows if shop_row_source_depth_bounds(row)[0] < minimum]
        if missing:
            raise AssertionError(f"{label}: rows below source_depth_min {minimum}: {missing}")
    if step.get("source_depth_max") is not None:
        maximum = int(step["source_depth_max"])
        missing = [row for row in rows if shop_row_source_depth_bounds(row)[1] > maximum]
        if missing:
            raise AssertionError(f"{label}: rows above source_depth_max {maximum}: {missing}")
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
            stash_items=payload.get("stash_items", []),
            stash_gold=int(payload.get("stash_gold", 0)),
            stash_capacity=int(payload.get("stash_capacity", 50)),
            skill_progression=payload.get("skill_progression", {}),
            skill_cooldowns=payload.get("skill_cooldowns", []),
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
    slot = next((s for s in hotbar if int(s.get("slot_index", -1)) == slot_index), {})
    item = slot.get("item")
    if item is None:
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
    if assertion.get("display_name_suffix") is not None:
        suffix = str(assertion["display_name_suffix"])
        if not str(item.get("display_name", "")).endswith(suffix):
            raise AssertionError(f"{where}: display_name missing suffix {suffix}: {item}")
    stats = item.get("rolled_stats", {})
    if not isinstance(stats, dict):
        raise AssertionError(f"{where}: rolled_stats is not an object: {item}")
    for key in assertion.get("stat_keys", []):
        if key not in stats:
            raise AssertionError(f"{where}: missing rolled stat {key}: {item}")
    if (stat_count_min := assertion.get("stat_count_min")) is not None and len(stats) < int(stat_count_min): raise AssertionError(f"{where}: rolled stat count {len(stats)} < {stat_count_min}: {item}")
    if (stat_count_max := assertion.get("stat_count_max")) is not None and len(stats) > int(stat_count_max): raise AssertionError(f"{where}: rolled stat count {len(stats)} > {stat_count_max}: {item}")
    if (min_damage := assertion.get("min_damage")) is not None and int(stats.get("damage_min", -1)) < int(min_damage):
        raise AssertionError(f"{where}: damage_min {stats.get('damage_min')} < {min_damage}: {item}")
    req = item.get("requirements", {})
    if int(req.get("level", 0)) != int(assertion.get("required_level", 1)):
        raise AssertionError(f"{where}: required level {req.get('level')} mismatch: {item}")
    if assertion.get("effect_ids") is not None:
        expected_effects = [str(effect_id) for effect_id in assertion["effect_ids"]]
        if item.get("effect_ids", []) != expected_effects:
            raise AssertionError(f"{where}: effect_ids {item.get('effect_ids', [])} != {expected_effects}: {item}")
    elif item.get("rarity") != "unique" and item.get("effect_ids", []) != []:
        raise AssertionError(f"{where}: non-unique effect_ids should be empty: {item}")


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
    for bool_key in ("is_boss", "monster_pack_leader"):
        if selector.get(bool_key) is not None and bool(entity.get(bool_key, False)) != bool(selector[bool_key]):
            return False
    if selector.get("visual_scale") is not None:
        if abs(float(entity.get("visual_scale", 1.0)) - float(selector["visual_scale"])) > 0.000001:
            return False
    if selector.get("hp") is not None and int(entity.get("hp", -999999)) != int(selector["hp"]): return False
    if selector.get("max_hp") is not None and int(entity.get("max_hp", -999999)) != int(selector["max_hp"]): return False
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


def assert_skill_progression(progression: dict[str, Any], assertion: dict[str, Any], where: str) -> None:
    if not progression:
        raise AssertionError(f"{where}: missing skill_progression")
    if "unspent_skill_points" in assertion:
        want = int(assertion["unspent_skill_points"])
        got = int(progression.get("unspent_skill_points", -1))
        if got != want:
            raise AssertionError(f"{where}: skill_progression.unspent_skill_points {got} != {want}: {progression}")
    skill_id = str(assertion.get("skill_id", "magic_bolt"))
    skill = next((row for row in progression.get("skills", []) if str(row.get("skill_id")) == skill_id), None)
    if skill is None:
        raise AssertionError(f"{where}: missing skill {skill_id}: {progression}")
    for key in ("rank", "max_rank"):
        if key in assertion:
            want_raw = assertion[key]
            if key == "max_rank" and str(want_raw) == "from_rules":
                want = skill_rule_max_rank(skill_id)
            else:
                want = int(want_raw)
            got = int(skill.get(key, -1))
            if got != want:
                raise AssertionError(f"{where}: skill {skill_id}.{key} {got} != {want}: {progression}")
    if "can_spend" in assertion:
        want = bool(assertion["can_spend"])
        got = bool(skill.get("can_spend", False))
        if got != want:
            raise AssertionError(f"{where}: skill {skill_id}.can_spend {got} != {want}: {progression}")


def assert_skill_cooldown(cooldowns: list[dict[str, Any]], assertion: dict[str, Any], where: str) -> None:
    skill_id = str(assertion.get("skill_id", "magic_bolt"))
    row = next((cooldown for cooldown in cooldowns if str(cooldown.get("skill_id")) == skill_id), None)
    if bool(assertion.get("absent", False)):
        if row is not None:
            raise AssertionError(f"{where}: cooldown {skill_id} present, want absent: {cooldowns}")
        return
    if row is None:
        raise AssertionError(f"{where}: missing cooldown {skill_id}: {cooldowns}")
    for field in ("remaining_ticks", "total_ticks"):
        if field in assertion:
            assert_count_matches(int(row.get(field, -1)), assertion[field] if isinstance(assertion[field], dict) else {"equals": int(assertion[field])}, f"{where}: cooldown {skill_id}.{field}")


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
    stash_items: list[dict[str, Any]] | None = None,
    stash_gold: int | None = None,
    stash_capacity: int | None = None,
    skill_progression: dict[str, Any] | None = None,
    skill_cooldowns: list[dict[str, Any]] | None = None,
) -> None:
    from tools.bot.runtime_assertions import run_assertions as run_snapshot_assertions

    run_snapshot_assertions(
        assertions,
        entities,
        inventory,
        equipped,
        item_id,
        where,
        current_level,
        walls,
        discovered_teleporters,
        character_progression,
        hotbar_capacity,
        hotbar,
        inventory_rows,
        inventory_capacity,
        gold,
        stash_items,
        stash_gold,
        stash_capacity,
        skill_progression,
        skill_cooldowns,
        globals(),
    )


def run_runtime_assertions(assertions: list[Any], state: RuntimeState, where: str) -> None:
    from tools.bot.runtime_assertions import run_runtime_assertions as run_runtime_assertions_impl

    run_runtime_assertions_impl(assertions, state, where, globals())


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
    debug_progression: dict[str, Any] | None = None,
) -> tuple[dict[str, Any], RuntimeState]:
    active_debug_progression = scenario.debug_progression if debug_progression is None else debug_progression
    character_id = ""
    if scenario.character_class or active_debug_progression:
        character_name = f"{scenario.character_class.title()} Bot" if scenario.character_class else f"{scenario.title} Bot"
        character_id = ensure_character(client, token, character_name, scenario.character_class)
    if active_debug_progression:
        seed_debug_progression(client, token, debug_token, character_id, active_debug_progression)
    sess = create_session(client, token, world_id, seed, character_id)
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
        character_class=scenario.character_class,
        debug_progression=active_debug_progression,
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
        stash_items=state.get("stash_items", []),
        stash_gold=int(state.get("stash_gold", 0)),
        stash_capacity=int(state.get("stash_capacity", 50)),
        skill_progression=state.get("skill_progression", {}),
        skill_cooldowns=state.get("skill_cooldowns", []),
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
    from tools.bot.coop_runtime import connect_coop_peer as connect_coop_peer_impl

    return await connect_coop_peer_impl(base_url, token, sess, label, world_id, helpers=globals())


async def close_coop_peer(peer: CoopPeer) -> None:
    from tools.bot.coop_runtime import close_coop_peer as close_coop_peer_impl

    await close_coop_peer_impl(peer, helpers=globals())


async def pump_coop(peers: list[CoopPeer], timeout: float = 0.1) -> None:
    from tools.bot.coop_runtime import pump_coop as pump_coop_impl

    await pump_coop_impl(peers, timeout, helpers=globals())


async def wait_coop_until(peers: list[CoopPeer], label: str, predicate, timeout_s: float = SLICE_TIMEOUT_S) -> None:
    from tools.bot.coop_runtime import wait_coop_until as wait_coop_until_impl

    await wait_coop_until_impl(peers, label, predicate, timeout_s, helpers=globals())


async def send_coop_intent(peer: CoopPeer, msg_type: str, payload: dict[str, Any]) -> str:
    from tools.bot.coop_runtime import send_coop_intent as send_coop_intent_impl

    return await send_coop_intent_impl(peer, msg_type, payload, helpers=globals())


async def wait_coop_accept(peers: list[CoopPeer], peer: CoopPeer, message_id: str) -> None:
    from tools.bot.coop_runtime import wait_coop_accept as wait_coop_accept_impl

    await wait_coop_accept_impl(peers, peer, message_id, helpers=globals())


def player_position(state: RuntimeState) -> dict[str, Any]:
    from tools.bot.coop_runtime import player_position as player_position_impl

    return player_position_impl(state, helpers=globals())


def player_entity_ids(state: RuntimeState) -> set[str]:
    from tools.bot.coop_runtime import player_entity_ids as player_entity_ids_impl

    return player_entity_ids_impl(state, helpers=globals())


def state_experience(state: RuntimeState) -> int:
    return int(state.character_progression.get("experience", 0))


def state_gold(state: RuntimeState) -> int:
    return int(state.gold)


def assert_party_contains_roles(state: RuntimeState, where: str) -> None:
    from tools.bot.coop_runtime import assert_party_contains_roles as assert_party_contains_roles_impl

    assert_party_contains_roles_impl(state, where, helpers=globals())


def find_non_gold_loot(state: RuntimeState) -> dict[str, Any] | None:
    for loot_id in list(state.loot_ids):
        loot = state.entities.get(str(loot_id))
        if loot is None or loot.get("type") != "loot":
            continue
        if str(loot.get("item_def_id", "")) != "gold":
            return loot
    return None


def find_new_non_gold_loot(state: RuntimeState, before_ids: set[str]) -> dict[str, Any] | None:
    for loot_id in sorted(str(loot_id) for loot_id in state.loot_ids):
        if loot_id in before_ids:
            continue
        loot = state.entities.get(loot_id)
        if loot is None or loot.get("type") != "loot":
            continue
        if str(loot.get("item_def_id", "")) != "gold":
            return loot
    return None


def entity_visible_as_loot(state: RuntimeState, entity_id: str, item_def_id: str | None = None) -> bool:
    entity = state.entities.get(str(entity_id))
    if entity is None or entity.get("type") != "loot":
        return False
    if item_def_id is not None and str(entity.get("item_def_id", "")) != item_def_id:
        return False
    return True


def dominant_step_direction(from_pos: dict[str, Any], to_pos: dict[str, Any]) -> dict[str, int]:
    dx = float(to_pos["x"]) - float(from_pos["x"])
    dy = float(to_pos["y"]) - float(from_pos["y"])
    if abs(dx) >= abs(dy):
        return {"x": 1 if dx >= 0 else -1, "y": 0}
    return {"x": 0, "y": 1 if dy >= 0 else -1}


async def stage_peers_for_same_tick_gold_pickup(peers: list[CoopPeer], host: CoopPeer, guest: CoopPeer, gold: dict[str, Any]) -> dict[str, int]:
    gold_pos = dict(gold["position"])
    direction = dominant_step_direction(player_position(host.state), gold_pos)
    stage_pos = {
        "x": float(gold_pos["x"]) - float(direction["x"]) * 2.0,
        "y": float(gold_pos["y"]) - float(direction["y"]) * 2.0,
    }
    await move_coop_peer_to(peers, host, stage_pos, stop_distance=0.25, max_ticks=320)
    await move_coop_peer_to(peers, guest, stage_pos, stop_distance=0.25, max_ticks=320)
    gold_id = str(gold["id"])
    if not entity_visible_as_loot(host.state, gold_id, "gold") or not entity_visible_as_loot(guest.state, gold_id, "gold"):
        raise AssertionError(f"gold {gold_id} was consumed before same-tick staging")
    return direction


async def open_chest_for_non_gold_loot(peers: list[CoopPeer], host: CoopPeer) -> dict[str, Any] | None:
    chest = find_interactable(host.state, "treasure_chest")
    if chest is None or str(chest.get("state", "closed")) == "open":
        return None
    before_ids = {str(loot_id) for loot_id in host.state.loot_ids}
    await move_coop_peer_to(peers, host, dict(chest["position"]), stop_distance=1.0, max_ticks=320)
    message_id = await send_coop_intent(host, "action_intent", {"target_id": str(chest["id"])})
    await wait_coop_accept(peers, host, message_id)
    await pump_coop(peers, timeout=0.3)
    return find_new_non_gold_loot(host.state, before_ids) or find_non_gold_loot(host.state)


async def move_coop_peer_to(
    peers: list[CoopPeer],
    peer: CoopPeer,
    target_pos: dict[str, Any],
    *,
    stop_distance: float = WALK_STOP_DISTANCE,
    max_ticks: int = 240,
) -> None:
    if await try_move_coop_peer_to(peers, peer, target_pos, stop_distance=stop_distance, max_ticks=max_ticks):
        return
    player = find_player(peer.state)
    raise TimeoutError(f"{peer.label}: did not reach {target_pos}; player={(player or {}).get('position')}")


async def try_move_coop_peer_to(
    peers: list[CoopPeer],
    peer: CoopPeer,
    target_pos: dict[str, Any],
    *,
    stop_distance: float = WALK_STOP_DISTANCE,
    max_ticks: int = 240,
) -> bool:
    player = find_player(peer.state)
    if player is None:
        raise AssertionError(f"{peer.label}: missing local player")
    if dict_distance(player["position"], target_pos) <= stop_distance:
        return True
    message_id = await send_coop_intent(peer, "move_to_intent", {"position": target_pos})
    await wait_coop_until(
        peers,
        f"{peer.label} move_to {message_id}",
        lambda: message_id in peer.state.accepted_message_ids or message_id in peer.state.rejected_message_reasons,
    )
    reason = peer.state.rejected_message_reasons.pop(message_id, None)
    if reason is not None:
        if reason in {"no_path", "path_too_long"}:
            return False
        raise AssertionError(f"{peer.label} move_to rejected: {reason}")
    for _ in range(max_ticks):
        player = find_player(peer.state)
        if player is None:
            raise AssertionError(f"{peer.label}: missing local player while moving")
        if dict_distance(player["position"], target_pos) <= stop_distance:
            return True
        await pump_coop(peers, timeout=0.05)
    return False


async def move_coop_peer_near(
    peers: list[CoopPeer],
    peer: CoopPeer,
    center: dict[str, Any],
    *,
    offsets: list[tuple[float, float]],
    stop_distance: float = 1.2,
    max_ticks: int = 260,
) -> dict[str, Any]:
    for dx, dy in offsets:
        target = {"x": float(center["x"]) + dx, "y": float(center["y"]) + dy}
        if await try_move_coop_peer_to(peers, peer, target, stop_distance=stop_distance, max_ticks=max_ticks):
            return target
    raise TimeoutError(f"{peer.label}: could not reach an approach point near {center}")


async def move_coop_peer(peers: list[CoopPeer], peer: CoopPeer, direction: dict[str, int]) -> None:
    before = player_position(peer.state)
    message_id = await send_coop_intent(peer, "move_intent", {"direction": direction, "duration_ticks": 1})
    await wait_coop_accept(peers, peer, message_id)
    await wait_coop_until(
        peers,
        f"{peer.label} local movement",
        lambda: player_position(peer.state) != before,
    )


async def wait_for_coop_damage_signal(
    peers: list[CoopPeer],
    driver: CoopPeer,
    *,
    min_raw_damage: int,
    timeout_s: float = 16.0,
) -> None:
    loop = asyncio.get_event_loop()
    start_indexes = {peer.label: len(peer.state.combat_events) for peer in peers}
    deadline = loop.time() + timeout_s
    while loop.time() <= deadline:
        for peer in peers:
            for ev in peer.state.combat_events[start_indexes[peer.label]:]:
                if ev.get("event_type") == "player_damaged" and int(ev.get("raw_damage", 0)) >= min_raw_damage:
                    return
        message_id = await send_coop_intent(driver, "move_intent", {"direction": {"x": 0, "y": 0}, "duration_ticks": 1})
        await wait_coop_accept(peers, driver, message_id)
        await pump_coop(peers, timeout=0.1)
    raise TimeoutError(f"co-op damage signal did not reach raw_damage >= {min_raw_damage}")


async def coop_attack_until_kill(
    peers: list[CoopPeer],
    attacker: CoopPeer,
    monster_def_id: str,
    *,
    companions: list[CoopPeer] | None = None,
    timeout_s: float = 20.0,
) -> None:
    loop = asyncio.get_event_loop()
    deadline = loop.time() + timeout_s
    last_action = 0.0
    companion_offsets = [(-3.0, 0.0), (3.0, 0.0), (0.0, -3.0), (0.0, 3.0), (-4.0, 0.0), (4.0, 0.0)]
    while loop.time() <= deadline:
        if "monster_killed" in attacker.state.seen_events:
            return
        monster = find_monster(attacker.state, monster_def_id)
        if monster is None:
            if "monster_killed" in attacker.state.seen_events:
                return
            await pump_coop(peers, timeout=0.1)
            continue
        for companion in companions or []:
            if companion.state.current_level != attacker.state.current_level:
                continue
            if dict_distance(player_position(companion.state), monster["position"]) > 6.0:
                await move_coop_peer_near(peers, companion, monster["position"], offsets=companion_offsets, stop_distance=1.4, max_ticks=180)
        if loop.time() - last_action >= 0.12:
            message_id = await send_coop_intent(attacker, "action_intent", {"target_id": str(monster["id"])})
            await wait_coop_until(
                peers,
                f"{attacker.label} attack {message_id}",
                lambda: message_id in attacker.state.accepted_message_ids or message_id in attacker.state.rejected_message_reasons,
            )
            reason = attacker.state.rejected_message_reasons.pop(message_id, None)
            if reason is not None:
                if reason == "invalid_target" and "monster_killed" in attacker.state.seen_events:
                    return
                if reason not in {"basic_attack_on_cooldown", "projectile_busy"}:
                    raise AssertionError(f"{attacker.label} attack rejected: {reason}")
            last_action = loop.time()
        await pump_coop(peers, timeout=0.1)
    raise TimeoutError(f"co-op attack did not kill {monster_def_id}")


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


async def run_coop_rewards_and_scaling(
    *,
    client: httpx.Client,
    base_url: str,
    tokens: list[str],
    debug_token: str,
    scenario: Scenario,
    character_ids: list[str],
) -> tuple[dict[str, Any], RuntimeState]:
    if len(tokens) < 3 or len(character_ids) < 3:
        raise AssertionError(f"{scenario.id}: requires host, nearby guest, and excluded guest")
    sess = create_coop_session(client, tokens[0], scenario.world_id, character_ids[0], scenario.seed)
    session_id = str(sess["session_id"])

    host = await connect_coop_peer(base_url, tokens[0], sess, "host", scenario.world_id)
    peers = [host]
    joined_sessions = [sess]
    try:
        for index, label in ((1, "nearby-guest"), (2, "excluded-guest")):
            joined = join_coop_session(client, tokens[index], session_id, str(sess["join_code"]), character_ids[index])
            joined_sessions.append(joined)
            peers.append(await connect_coop_peer(base_url, tokens[index], joined, label, scenario.world_id))

        nearby = peers[1]
        excluded = peers[2]
        await wait_coop_until(peers, "all peers have party rows", lambda: all(len(peer.state.party) >= 3 for peer in peers))

        reward_peers = [host, nearby]
        monster_def_id = "dungeon_mob"
        if scenario.world_id == "dungeon_levels":
            await execute_step(host.ws, session_id, host.state, {"action": "use_stair", "direction": "down", "max_ticks": 240}, asyncio.get_event_loop())
            await execute_step(nearby.ws, session_id, nearby.state, {"action": "use_stair", "direction": "down", "max_ticks": 240}, asyncio.get_event_loop())
            await wait_coop_until(
                peers,
                "host and nearby guest share dungeon while excluded guest stays town",
                lambda: host.state.current_level == -1 and nearby.state.current_level == -1 and excluded.state.current_level == 0,
            )
            monster_def_id = "dungeon_archer"
        else:
            await wait_coop_until(
                peers,
                "all compact reward peers share level",
                lambda: host.state.current_level == nearby.state.current_level == excluded.state.current_level,
            )
            await move_coop_peer_to(peers, excluded, {"x": 2, "y": 2}, stop_distance=0.5, max_ticks=180)
        await wait_coop_until(
            reward_peers,
            "reward peers see each other",
            lambda: all(player_entity_ids(peer.state) >= {host.state.local_player_id, nearby.state.local_player_id} for peer in reward_peers),
        )

        monster = find_monster(host.state, monster_def_id)
        if monster is None:
            raise AssertionError(f"{scenario.id}: missing {monster_def_id} in host state")
        monster_pos = dict(monster["position"])
        approach_offsets = [(-1.5, 0.0), (1.5, 0.0), (0.0, -1.5), (0.0, 1.5), (-2.5, 0.0), (2.5, 0.0)]
        nearby_offsets = [(-4.0, 0.0), (4.0, 0.0), (0.0, -4.0), (0.0, 4.0), (-3.0, 0.0), (3.0, 0.0)]
        await move_coop_peer_near(reward_peers, nearby, monster_pos, offsets=nearby_offsets, stop_distance=1.2, max_ticks=260)
        await move_coop_peer_near(reward_peers, host, monster_pos, offsets=approach_offsets, stop_distance=1.0, max_ticks=260)
        if dict_distance(player_position(nearby.state), monster_pos) > 10.0:
            raise AssertionError(f"nearby guest too far from monster: {player_position(nearby.state)} monster={monster_pos}")

        before_host_xp = state_experience(host.state)
        before_nearby_xp = state_experience(nearby.state)
        before_excluded_xp = state_experience(excluded.state)
        expected_xp = monster_xp_reward(monster_def_id)

        await coop_attack_until_kill(reward_peers, nearby, monster_def_id, companions=[host])
        await wait_coop_until(
            peers,
            "shared xp applied to nearby only",
            lambda: state_experience(host.state) >= before_host_xp + expected_xp
            and state_experience(nearby.state) >= before_nearby_xp + expected_xp
            and state_experience(excluded.state) == before_excluded_xp,
        )

        for peer in list(peers):
            await close_coop_peer(peer)
        peers.clear()

        replay = fetch_replay(client, tokens[0], debug_token, session_id)
        if not replay.get("match", False):
            raise AssertionError(f"co-op rewards replay mismatch for {session_id}: {replay.get('mismatch')}")

        fresh_host = create_session(client, tokens[0], scenario.world_id)
        fresh_host_state = fetch_state(client, tokens[0], debug_token, str(fresh_host["session_id"]))
        host_fresh_xp = int(fresh_host_state.get("character_progression", {}).get("experience", 0))
        if host_fresh_xp < before_host_xp + expected_xp:
            raise AssertionError(f"host fresh xp={host_fresh_xp}, want >= {before_host_xp + expected_xp}")

        fresh_nearby = create_session(client, tokens[1], scenario.world_id)
        fresh_nearby_state = fetch_state(client, tokens[1], debug_token, str(fresh_nearby["session_id"]))
        nearby_fresh_xp = int(fresh_nearby_state.get("character_progression", {}).get("experience", 0))
        if nearby_fresh_xp < before_nearby_xp + expected_xp:
            raise AssertionError(f"nearby fresh xp={nearby_fresh_xp}, want >= {before_nearby_xp + expected_xp}")

        fresh_excluded = create_session(client, tokens[2], scenario.world_id)
        fresh_excluded_state = fetch_state(client, tokens[2], debug_token, str(fresh_excluded["session_id"]))
        excluded_fresh_xp = int(fresh_excluded_state.get("character_progression", {}).get("experience", 0))
        if excluded_fresh_xp != before_excluded_xp:
            raise AssertionError(f"excluded fresh xp={excluded_fresh_xp}, want {before_excluded_xp}")
    finally:
        for peer in peers:
            try:
                await close_coop_peer(peer)
            except Exception:
                pass

    _ = debug_token
    log("co-op rewards/scaling protocol checks matched", session_id)
    return sess, host.state


async def run_non_gold_click_required_item_proof(
    *,
    client: httpx.Client,
    base_url: str,
    token: str,
    debug_token: str,
) -> None:
    sess = create_session(client, token, "coop_loot_lab", "chest_seed_22")
    session_id = str(sess["session_id"])
    peer = await connect_coop_peer(base_url, token, sess, "item-proof", "coop_loot_lab")
    peers = [peer]
    try:
        non_gold = find_non_gold_loot(peer.state) or await open_chest_for_non_gold_loot(peers, peer)
        if non_gold is None:
            raise AssertionError("item proof: chest did not produce non-gold loot")

        item_loot_id = str(non_gold["id"])
        before_inventory_count = len(peer.state.inventory)
        await move_coop_peer_to(peers, peer, dict(non_gold["position"]), stop_distance=1.25, max_ticks=260)
        await pump_coop(peers, timeout=0.2)
        if len(peer.state.inventory) != before_inventory_count:
            raise AssertionError(
                f"non-gold loot auto-picked without click: before={before_inventory_count} after={len(peer.state.inventory)}"
            )
        if not entity_visible_as_loot(peer.state, item_loot_id):
            raise AssertionError(f"non-gold loot {item_loot_id} disappeared before explicit pickup")

        pick_msg = await send_coop_intent(peer, "action_intent", {"target_id": item_loot_id})
        await wait_coop_accept(peers, peer, pick_msg)
        await wait_coop_until(
            peers,
            "explicit item pickup adds inventory and removes loot",
            lambda: len(peer.state.inventory) > before_inventory_count and item_loot_id not in peer.state.entities,
        )
    finally:
        for item_peer in peers:
            try:
                await close_coop_peer(item_peer)
            except Exception:
                pass

    replay = fetch_replay(client, token, debug_token, session_id)
    if not replay.get("match", False):
        raise AssertionError(f"non-gold item proof replay mismatch for {session_id}: {replay.get('mismatch')}")


async def run_gold_autopickup_shared_loot(
    *,
    client: httpx.Client,
    base_url: str,
    tokens: list[str],
    debug_token: str,
    scenario: Scenario,
    character_ids: list[str],
) -> tuple[dict[str, Any], RuntimeState]:
    if len(tokens) < 2 or len(character_ids) < 2:
        raise AssertionError(f"{scenario.id}: requires host and guest")
    sess = create_coop_session(client, tokens[0], scenario.world_id, character_ids[0], scenario.seed)
    session_id = str(sess["session_id"])

    host = await connect_coop_peer(base_url, tokens[0], sess, "host", scenario.world_id)
    guest: CoopPeer | None = None
    peers = [host]
    try:
        joined = join_coop_session(client, tokens[1], session_id, str(sess["join_code"]), character_ids[1])
        guest = await connect_coop_peer(base_url, tokens[1], joined, "guest", scenario.world_id)
        peers.append(guest)
        await wait_coop_until(peers, "both peers have party rows", lambda: all(len(peer.state.party) >= 2 for peer in peers))

        if scenario.world_id == "dungeon_levels":
            for peer in (host, guest):
                await execute_step(peer.ws, session_id, peer.state, {"action": "use_stair", "direction": "down", "max_ticks": 260}, asyncio.get_event_loop())
                await execute_step(peer.ws, session_id, peer.state, {"action": "use_stair", "direction": "down", "max_ticks": 260}, asyncio.get_event_loop())
            await wait_coop_until(
                peers,
                "both peers share level -2",
                lambda: host.state.current_level == -2 and guest.state.current_level == -2
                and player_entity_ids(host.state) >= {host.state.local_player_id, guest.state.local_player_id}
                and player_entity_ids(guest.state) >= {host.state.local_player_id, guest.state.local_player_id},
            )
        else:
            await wait_coop_until(
                peers,
                "both peers share compact loot level",
                lambda: host.state.current_level == guest.state.current_level
                and player_entity_ids(host.state) >= {host.state.local_player_id, guest.state.local_player_id}
                and player_entity_ids(guest.state) >= {host.state.local_player_id, guest.state.local_player_id},
            )

        gold = find_loot(host.state, "gold")
        if gold is None:
            raise AssertionError(f"{scenario.id}: missing generated dungeon gold in host state")
        gold_id = str(gold["id"])
        if not entity_visible_as_loot(guest.state, gold_id, "gold"):
            raise AssertionError(f"guest does not see shared gold {gold_id}: {guest.state.entities}")

        before_host_gold = state_gold(host.state)
        before_guest_gold = state_gold(guest.state)
        direction = await stage_peers_for_same_tick_gold_pickup(peers, host, guest, gold)
        host_msg = await send_coop_intent(host, "move_intent", {"direction": direction, "duration_ticks": 1})
        guest_msg = await send_coop_intent(guest, "move_intent", {"direction": direction, "duration_ticks": 1})
        await wait_coop_accept(peers, host, host_msg)
        await wait_coop_accept(peers, guest, guest_msg)
        await wait_coop_until(
            peers,
            "shared gold removed and awarded to lowest player id",
            lambda: gold_id not in host.state.entities
            and gold_id not in guest.state.entities
            and state_gold(host.state) > before_host_gold
            and state_gold(guest.state) == before_guest_gold,
        )
        if "gold_picked_up" not in host.state.seen_events:
            raise AssertionError("host did not receive gold_picked_up")
        if "gold_picked_up" in guest.state.seen_events:
            raise AssertionError("guest received another player's private gold_picked_up")

        for peer in list(peers):
            await close_coop_peer(peer)
        peers.clear()

        state_body = fetch_state(client, tokens[0], debug_token, session_id)
        if int(state_body.get("gold", 0)) < state_gold(host.state):
            raise AssertionError(f"/state gold={state_body.get('gold')} below observed {state_gold(host.state)}")

        replay = fetch_replay(client, tokens[0], debug_token, session_id)
        if not replay.get("match", False):
            raise AssertionError(f"gold auto-pickup replay mismatch for {session_id}: {replay.get('mismatch')}")

        await run_non_gold_click_required_item_proof(
            client=client,
            base_url=base_url,
            token=tokens[0],
            debug_token=debug_token,
        )

        fresh_host = create_session(client, tokens[0], scenario.world_id)
        fresh_host_state = fetch_state(client, tokens[0], debug_token, str(fresh_host["session_id"]))
        if int(fresh_host_state.get("gold", 0)) < state_gold(host.state):
            raise AssertionError(f"fresh host gold={fresh_host_state.get('gold')}, want >= {state_gold(host.state)}")
        fresh_guest = create_session(client, tokens[1], scenario.world_id)
        fresh_guest_state = fetch_state(client, tokens[1], debug_token, str(fresh_guest["session_id"]))
        if int(fresh_guest_state.get("gold", 0)) != before_guest_gold:
            raise AssertionError(f"fresh guest gold={fresh_guest_state.get('gold')}, want unchanged {before_guest_gold}")
    finally:
        for peer in peers:
            try:
                await close_coop_peer(peer)
            except Exception:
                pass

    log("gold auto-pickup/shared loot protocol checks matched", session_id)
    return sess, host.state


# --- main -------------------------------------------------------------------

def main() -> int:
    parser = argparse.ArgumentParser(description="arpg headless protocol bot")
    parser.add_argument("--base-url", default="http://localhost:8888")
    parser.add_argument("--dev-token", default="local-dev-token")
    parser.add_argument("--debug-token", default="local-debug-token")
    parser.add_argument("--email", default="bot@example.test")
    parser.add_argument("--scenario", default="all", help="scenario id, comma-separated ids, or all")
    parser.add_argument("--list-scenarios", action="store_true")
    parser.add_argument("--write-manifest", type=Path)
    parser.add_argument("--print-session-id", action="store_true")
    parser.add_argument("--cleanup-characters", action="store_true")
    args = parser.parse_args()

    scenarios = load_scenarios()
    if args.list_scenarios:
        for scenario in scenarios:
            print(f"{scenario.id}\t{scenario.title}")
        return 0
    selected = select_scenarios(scenarios, args.scenario)
    results: list[dict[str, Any]] = []
    cleanup_emails: set[str] = set()
    last_session_id = ""

    with httpx.Client(base_url=args.base_url, timeout=10.0) as client:
        for scenario in selected:
            scenario_started = time.monotonic()
            log("scenario begin", scenario.id, "-", scenario.title, f"world={scenario.world_id}")
            if scenario.id == "skill_visual":
                scenario, replay_email, token, sess, observed = skill_visual_runtime.run_selected(globals(), args, client, scenario)
            elif scenario.id == "true_coop_session":
                replay_email = scenario_email(args.email, scenario.id + "-host")
                guest_email = scenario_email(args.email, scenario.id + "-guest")
                cleanup_emails.update({replay_email, guest_email})
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
                    cleanup_emails.add(email)
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
            elif scenario.id == "coop_rewards_and_scaling":
                tokens = []
                character_ids = []
                peer_count = max(3, scenario.peer_count)
                replay_email = scenario_email(args.email, f"{scenario.id}-peer-0")
                for index in range(peer_count):
                    email = replay_email if index == 0 else scenario_email(args.email, f"{scenario.id}-peer-{index}")
                    cleanup_emails.add(email)
                    _, peer_token = dev_login(client, email, args.dev_token)
                    tokens.append(peer_token)
                    character_ids.append(ensure_character(client, peer_token, f"Reward Peer {index + 1}"))
                sess, observed = asyncio.run(run_coop_rewards_and_scaling(
                    client=client,
                    base_url=args.base_url,
                    tokens=tokens,
                    debug_token=args.debug_token,
                    scenario=scenario,
                    character_ids=character_ids,
                ))
                token = tokens[0]
            elif scenario.id == "gold_autopickup_shared_loot":
                tokens = []
                character_ids = []
                peer_count = max(2, scenario.peer_count)
                replay_email = scenario_email(args.email, f"{scenario.id}-peer-0")
                for index in range(peer_count):
                    email = replay_email if index == 0 else scenario_email(args.email, f"{scenario.id}-peer-{index}")
                    cleanup_emails.add(email)
                    _, peer_token = dev_login(client, email, args.dev_token)
                    tokens.append(peer_token)
                    character_ids.append(ensure_character(client, peer_token, f"Gold Peer {index + 1}"))
                sess, observed = asyncio.run(run_gold_autopickup_shared_loot(
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
                cleanup_emails.add(replay_email)
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
                    debug_progression=dict(check.get("debug_progression", scenario.debug_progression)),
                )
                last_session_id = check_sess["session_id"]
                observed = check_observed
                log("fresh session check done", scenario.id, f"#{idx}")

            scenario_elapsed = time.monotonic() - scenario_started
            assert_scenario_elapsed_within_budget(scenario.id, scenario_elapsed)
            log("scenario done", scenario.id, f"elapsed={scenario_elapsed:.2f}s")

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

        if args.cleanup_characters:
            removed = 0
            for email in sorted(cleanup_emails):
                removed += cleanup_account_characters(client, email, args.dev_token)
            if removed:
                log("deleted bot account characters", removed)

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
