"""Co-op combat soak scenarios for skill budget and session stability proofs."""
from __future__ import annotations

import json
import os
import statistics
from pathlib import Path
from typing import Any

import httpx

from tools.bot.bot_types import CoopPeer, RuntimeState, Scenario


def _require_helpers(helpers: dict[str, Any] | None) -> dict[str, Any]:
    if helpers is None:
        raise AssertionError("combat soak runtime requires helper bindings")
    return helpers


def parse_server_perf_log(log_path: Path) -> tuple[int, list[float], list[float]]:
    overrun_count = 0
    persist_samples: list[float] = []
    overrun_samples: list[float] = []
    for line in log_path.read_text(encoding="utf-8", errors="replace").splitlines():
        line = line.strip()
        if not line or line[0] != "{":
            continue
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue
        message = str(obj.get("message", ""))
        if message == "session_tick_budget_overrun":
            overrun_count += 1
            overrun_samples.append(float(obj.get("tick_overrun_ms", 0.0)))
            continue
        if message != "backend_perf":
            continue
        persist_samples.append(float(obj.get("persist_ms", 0.0)))
    return overrun_count, persist_samples, overrun_samples


def assert_soak_server_perf(scenario_id: str, soak_config: dict[str, Any], *, helpers: dict[str, Any]) -> None:
    log = helpers["log"]
    log_path = os.environ.get("ARPG_BOT_SERVER_LOG", "").strip()
    if not log_path:
        log("soak perf skip", scenario_id, "ARPG_BOT_SERVER_LOG unset")
        return
    path = Path(log_path)
    if not path.is_file():
        raise AssertionError(f"{scenario_id}: missing server perf log at {log_path}")
    overrun_count, persist_samples, overrun_samples = parse_server_perf_log(path)
    max_overruns = int(soak_config.get("max_tick_overruns", 12))
    if overrun_count > max_overruns:
        raise AssertionError(
            f"{scenario_id}: session_tick_budget_overrun count={overrun_count}, want <= {max_overruns}"
        )
    if persist_samples:
        persist_p95 = statistics.quantiles(persist_samples, n=20)[-1] if len(persist_samples) >= 2 else persist_samples[0]
        max_persist_p95 = float(soak_config.get("max_persist_ms_p95", 100.0))
        if persist_p95 > max_persist_p95:
            raise AssertionError(
                f"{scenario_id}: persist_ms p95={persist_p95:.1f}, want <= {max_persist_p95}"
            )
    if overrun_samples:
        overrun_p95 = statistics.quantiles(overrun_samples, n=20)[-1] if len(overrun_samples) >= 2 else overrun_samples[0]
        max_overrun_p95 = float(soak_config.get("max_tick_overrun_ms_p95", 50.0))
        if overrun_p95 > max_overrun_p95:
            raise AssertionError(
                f"{scenario_id}: tick_overrun_ms p95={overrun_p95:.1f}, want <= {max_overrun_p95}"
            )
    log(
        "soak perf ok",
        scenario_id,
        f"overruns={overrun_count}",
        f"persist_samples={len(persist_samples)}",
    )


async def run_crowded_skill_overlap_lab(
    *,
    client: httpx.Client,
    base_url: str,
    tokens: list[str],
    debug_token: str,
    scenario: Scenario,
    character_ids: list[str],
    helpers: dict[str, Any] | None = None,
) -> tuple[dict[str, Any], RuntimeState]:
    helpers = _require_helpers(helpers)
    log = helpers["log"]
    seed_debug_progression = helpers["seed_debug_progression"]
    create_coop_session = helpers["create_coop_session"]
    join_coop_session = helpers["join_coop_session"]
    connect_coop_peer = helpers["connect_coop_peer"]
    wait_coop_until = helpers["wait_coop_until"]
    send_coop_intent = helpers["send_coop_intent"]
    pump_coop = helpers["pump_coop"]
    close_coop_peer = helpers["close_coop_peer"]

    peer_count = max(2, scenario.peer_count)
    if len(tokens) < peer_count or len(character_ids) < peer_count:
        raise AssertionError(f"{scenario.id}: requires {peer_count} peers")

    for index in range(peer_count):
        seed_debug_progression(client, tokens[index], debug_token, character_ids[index], scenario.debug_progression)

    sess = create_coop_session(client, tokens[0], scenario.world_id, character_ids[0], scenario.seed)
    session_id = str(sess["session_id"])
    host = await connect_coop_peer(base_url, tokens[0], sess, "host", scenario.world_id)
    peers = [host]
    volley_casts = 0
    try:
        for index in range(1, peer_count):
            joined = join_coop_session(client, tokens[index], session_id, str(sess["join_code"]), character_ids[index])
            peers.append(await connect_coop_peer(base_url, tokens[index], joined, f"peer-{index}", scenario.world_id))

        await wait_coop_until(
            peers,
            "overlap peers ready",
            lambda: all(len(peer.state.party) >= peer_count for peer in peers),
            timeout_s=20.0,
        )

        chasers = sum(
            1
            for ent in host.state.entities.values()
            if ent.get("type") == "monster" and ent.get("monster_def_id") == "crowded_probe_chaser" and ent.get("hp", 0) > 0
        )
        if chasers < 25:
            raise AssertionError(f"{scenario.id}: crowded chasers={chasers}, want >= 25")

        for _round in range(3):
            pending: list[tuple[CoopPeer, str]] = []
            for peer in peers:
                message_id = await send_coop_intent(
                    peer,
                    "cast_skill_intent",
                    {"skill_id": "volley", "direction": {"x": 1, "y": 0}},
                )
                pending.append((peer, message_id))
            for peer, message_id in pending:
                await wait_coop_until(
                    peers,
                    f"{peer.label} overlap cast {message_id}",
                    lambda peer=peer, message_id=message_id: message_id in peer.state.accepted_message_ids
                    or message_id in peer.state.rejected_message_reasons,
                    timeout_s=10.0,
                )
                if message_id in peer.state.accepted_message_ids:
                    volley_casts += 1
            for _ in range(45):
                await pump_coop(peers, timeout=0.12)

        for peer in peers:
            if peer.ws.close_code is not None:
                raise AssertionError(f"{scenario.id}: {peer.label} websocket closed")

        if volley_casts < 2:
            raise AssertionError(f"{scenario.id}: accepted volley casts={volley_casts}, want >= 2")

        assert_soak_server_perf(
            scenario.id,
            {
                "max_tick_overruns": 8,
                "max_persist_ms_p95": 100.0,
                "max_tick_overrun_ms_p95": 50.0,
            },
            helpers=helpers,
        )
    finally:
        for peer in peers:
            try:
                await close_coop_peer(peer)
            except Exception:
                pass

    log("crowded skill overlap lab matched", session_id, f"volley_casts={volley_casts}")
    return sess, host.state


async def run_six_player_boss_combat_soak(
    *,
    client: httpx.Client,
    base_url: str,
    tokens: list[str],
    debug_token: str,
    scenario: Scenario,
    character_ids: list[str],
    helpers: dict[str, Any] | None = None,
) -> tuple[dict[str, Any], RuntimeState]:
    helpers = _require_helpers(helpers)
    log = helpers["log"]
    seed_debug_progression = helpers["seed_debug_progression"]
    create_coop_session = helpers["create_coop_session"]
    join_coop_session = helpers["join_coop_session"]
    connect_coop_peer = helpers["connect_coop_peer"]
    wait_coop_until = helpers["wait_coop_until"]
    send_coop_intent = helpers["send_coop_intent"]
    pump_coop = helpers["pump_coop"]
    close_coop_peer = helpers["close_coop_peer"]
    player_entity_ids = helpers["player_entity_ids"]

    raw = json.loads(scenario.path.read_text())
    loadouts = list(raw.get("peer_loadouts", []))
    soak_config = dict(raw.get("soak_config", {}))
    peer_count = scenario.peer_count
    if len(tokens) < peer_count or len(character_ids) < peer_count:
        raise AssertionError(f"{scenario.id}: peer_count={peer_count} requires enough tokens/characters")
    if len(loadouts) < peer_count:
        raise AssertionError(f"{scenario.id}: peer_loadouts length={len(loadouts)}, want >= {peer_count}")

    for index in range(peer_count):
        loadout = loadouts[index]
        progression = dict(loadout.get("debug_progression", scenario.debug_progression))
        seed_debug_progression(client, tokens[index], debug_token, character_ids[index], progression)

    sess = create_coop_session(client, tokens[0], scenario.world_id, character_ids[0], scenario.seed)
    session_id = str(sess["session_id"])
    host = await connect_coop_peer(base_url, tokens[0], sess, "host", scenario.world_id)
    peers = [host]
    casts_seen = 0
    try:
        for index in range(1, peer_count):
            joined = join_coop_session(client, tokens[index], session_id, str(sess["join_code"]), character_ids[index])
            peers.append(await connect_coop_peer(base_url, tokens[index], joined, f"peer-{index}", scenario.world_id))

        await wait_coop_until(
            peers,
            f"{peer_count} peers connected on same level",
            lambda: all(len(peer.state.party) >= peer_count for peer in peers)
            and all(player_entity_ids(peer.state) >= {p.state.local_player_id for p in peers} for peer in peers),
            timeout_s=30.0,
        )

        cast_cycles = int(soak_config.get("cast_cycles", 8))
        ticks_between = int(soak_config.get("ticks_between_cycles", 18))
        for _cycle in range(cast_cycles):
            for index, peer in enumerate(peers):
                skill_id = str(loadouts[index].get("cast_skill", loadouts[index].get("skill_id", "magic_bolt")))
                message_id = await send_coop_intent(
                    peer,
                    "cast_skill_intent",
                    {"skill_id": skill_id, "direction": {"x": 1, "y": 0}},
                )
                await wait_coop_until(
                    peers,
                    f"{peer.label} cast response {message_id}",
                    lambda peer=peer, message_id=message_id: message_id in peer.state.accepted_message_ids
                    or message_id in peer.state.rejected_message_reasons,
                    timeout_s=8.0,
                )
                if message_id in peer.state.accepted_message_ids:
                    casts_seen += 1
            for _ in range(ticks_between):
                await pump_coop(peers, timeout=0.12)

        for peer in peers:
            if peer.ws.close_code is not None:
                raise AssertionError(f"{scenario.id}: {peer.label} websocket closed with code {peer.ws.close_code}")

        min_casts = max(peer_count, cast_cycles)
        if casts_seen < min_casts:
            raise AssertionError(f"{scenario.id}: accepted casts={casts_seen}, want >= {min_casts}")

        assert_soak_server_perf(scenario.id, soak_config, helpers=helpers)
    finally:
        for peer in peers:
            try:
                await close_coop_peer(peer)
            except Exception:
                pass

    log("six-player boss combat soak matched", session_id, f"casts={casts_seen}")
    return sess, host.state
