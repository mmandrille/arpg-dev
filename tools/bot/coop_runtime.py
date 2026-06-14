from __future__ import annotations

import asyncio
import json
from typing import Any

import websockets

from tools.bot.bot_types import CoopPeer, RuntimeState
from tools.bot.protocol import make_envelope, to_ws_url

SLICE_TIMEOUT_S = 20.0


def _require_helpers(helpers: dict[str, Any] | None) -> dict[str, Any]:
    if helpers is None:
        raise AssertionError("co-op runtime helpers require helper bindings")
    return helpers


async def connect_coop_peer(base_url: str, token: str, sess: dict[str, Any], label: str, world_id: str, helpers: dict[str, Any] | None = None) -> CoopPeer:
    helpers = _require_helpers(helpers)
    auth = helpers["auth"]
    recv_json = helpers["recv_json"]
    ingest_snapshot = helpers["ingest_snapshot"]
    log = helpers["log"]
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


async def close_coop_peer(peer: CoopPeer, helpers: dict[str, Any] | None = None) -> None:
    helpers = _require_helpers(helpers)
    await peer.ws.close()


async def pump_coop(peers: list[CoopPeer], timeout: float = 0.1, helpers: dict[str, Any] | None = None) -> None:
    helpers = _require_helpers(helpers)
    ingest_message = helpers["ingest_message"]
    tasks = {asyncio.create_task(peer.ws.recv()): peer for peer in peers}
    done, pending = await asyncio.wait(tasks.keys(), timeout=timeout, return_when=asyncio.FIRST_COMPLETED)
    for task in pending:
        task.cancel()
    for task in done:
        peer = tasks[task]
        ingest_message(json.loads(task.result()), peer.state)


async def wait_coop_until(peers: list[CoopPeer], label: str, predicate, timeout_s: float = SLICE_TIMEOUT_S, helpers: dict[str, Any] | None = None) -> None:
    helpers = _require_helpers(helpers)
    pump_coop = helpers["pump_coop"]
    loop = asyncio.get_event_loop()
    deadline = loop.time() + timeout_s
    while not predicate():
        if loop.time() > deadline:
            raise TimeoutError(f"co-op wait timed out: {label}")
        await pump_coop(peers, timeout=0.1)


async def send_coop_intent(peer: CoopPeer, msg_type: str, payload: dict[str, Any], helpers: dict[str, Any] | None = None) -> str:
    helpers = _require_helpers(helpers)
    env = make_envelope(msg_type, peer.session["session_id"], peer.state.last_tick, payload)
    await peer.ws.send(json.dumps(env))
    return str(env["message_id"])


async def wait_coop_accept(peers: list[CoopPeer], peer: CoopPeer, message_id: str, helpers: dict[str, Any] | None = None) -> None:
    helpers = _require_helpers(helpers)
    wait_coop_until = helpers["wait_coop_until"]
    await wait_coop_until(
        peers,
        f"{peer.label} accept {message_id}",
        lambda: message_id in peer.state.accepted_message_ids or message_id in peer.state.rejected_message_reasons,
    )
    if message_id in peer.state.rejected_message_reasons:
        raise AssertionError(f"{peer.label} intent {message_id} rejected: {peer.state.rejected_message_reasons[message_id]}")


def player_position(state: RuntimeState, helpers: dict[str, Any] | None = None) -> dict[str, Any]:
    helpers = _require_helpers(helpers)
    find_player = helpers["find_player"]
    player = find_player(state)
    if player is None:
        raise AssertionError(f"missing local player {state.local_player_id}")
    return dict(player.get("position", {}))


def player_entity_ids(state: RuntimeState, helpers: dict[str, Any] | None = None) -> set[str]:
    helpers = _require_helpers(helpers)
    return {str(entity_id) for entity_id, entity in state.entities.items() if entity.get("type") == "player"}


def assert_party_contains_roles(state: RuntimeState, where: str, helpers: dict[str, Any] | None = None) -> None:
    helpers = _require_helpers(helpers)
    roles = {str(row.get("role", "")) for row in state.party}
    if not {"host", "guest"} <= roles:
        raise AssertionError(f"{where}: party roles {roles}, want host+guest; party={state.party}")
