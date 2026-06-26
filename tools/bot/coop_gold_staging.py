from __future__ import annotations

from collections.abc import Awaitable, Callable
from typing import Any

from tools.bot.bot_types import CoopPeer, RuntimeState

MoveCoopPeerTo = Callable[..., Awaitable[None]]
PlayerPosition = Callable[[RuntimeState], dict[str, Any]]
EntityVisibleAsLoot = Callable[[RuntimeState, str, str | None], bool]


def dominant_step_direction(from_pos: dict[str, Any], to_pos: dict[str, Any]) -> dict[str, int]:
    dx = float(to_pos["x"]) - float(from_pos["x"])
    dy = float(to_pos["y"]) - float(from_pos["y"])
    if abs(dx) >= abs(dy):
        return {"x": 1 if dx >= 0 else -1, "y": 0}

    return {"x": 0, "y": 1 if dy >= 0 else -1}


async def stage_peers_for_same_tick_gold_pickup(
    peers: list[CoopPeer],
    host: CoopPeer,
    guest: CoopPeer,
    gold: dict[str, Any],
    *,
    move_coop_peer_to: MoveCoopPeerTo,
    player_position: PlayerPosition,
    entity_visible_as_loot: EntityVisibleAsLoot,
) -> dict[str, int]:
    gold_pos = dict(gold["position"])
    direction = dominant_step_direction(player_position(host.state), gold_pos)
    # Outside resting pickup range; one min-momentum tick enters range (gold_auto_pickup_test.go).
    stage_distance = 1.7
    stage_pos = {
        "x": float(gold_pos["x"]) - float(direction["x"]) * stage_distance,
        "y": float(gold_pos["y"]) - float(direction["y"]) * stage_distance,
    }
    await move_coop_peer_to(peers, host, stage_pos, stop_distance=0.05, max_ticks=320)
    await move_coop_peer_to(peers, guest, stage_pos, stop_distance=0.05, max_ticks=320)
    gold_id = str(gold["id"])
    if not entity_visible_as_loot(host.state, gold_id, "gold") or not entity_visible_as_loot(guest.state, gold_id, "gold"):
        raise AssertionError(f"gold {gold_id} was consumed before same-tick staging")

    return direction
