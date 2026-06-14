from __future__ import annotations

import math
import json
from typing import Any

from tools.bot.bot_types import RuntimeState
from tools.bot.protocol import make_envelope

SLICE_TIMEOUT_S = 20.0
WALK_STOP_DISTANCE = 1.0
WALK_MAX_TICKS = 40


def _require_helpers(helpers: dict[str, Any] | None) -> dict[str, Any]:
    if helpers is None:
        raise AssertionError("movement runtime helpers require helper bindings")
    return helpers


async def walk_toward(
    ws,
    session_id: str,
    state: RuntimeState,
    target_pos: dict[str, Any],
    loop,
    max_ticks: int = WALK_MAX_TICKS,
    stop_distance: float = WALK_STOP_DISTANCE,
    helpers: dict[str, Any] | None = None,
) -> None:
    helpers = _require_helpers(helpers)
    find_player = helpers["find_player"]
    move_to_position = helpers["move_to_position"]
    wait_for_player_move_or_accept = helpers["wait_for_player_move_or_accept"]
    pump_one = helpers["pump_one"]
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


async def move_until_entity_in_range(
    ws,
    session_id: str,
    state: RuntimeState,
    target_id: str,
    loop,
    *,
    stop_distance: float,
    max_ticks: int = WALK_MAX_TICKS,
    helpers: dict[str, Any] | None = None,
) -> None:
    helpers = _require_helpers(helpers)
    find_player = helpers["find_player"]
    dict_distance = helpers["dict_distance"]
    range_candidate_positions = helpers["range_candidate_positions"]
    walk_toward = helpers["walk_toward"]
    last_error: Exception | None = None
    attempts = max(1, max_ticks // 20)
    for _ in range(attempts):
        target = state.entities.get(target_id)
        player = find_player(state)
        if target is None:
            raise AssertionError(f"move_until_entity_in_range: target vanished: {target_id}")
        if player is None:
            raise AssertionError("move_until_entity_in_range: player not found")
        if dict_distance(player["position"], target["position"]) <= stop_distance:
            return

        for candidate in range_candidate_positions(player["position"], target["position"], stop_distance):
            try:
                await walk_toward(
                    ws,
                    session_id,
                    state,
                    candidate,
                    loop,
                    max_ticks=min(max_ticks, 120),
                    stop_distance=0.75,
                )
            except AssertionError as exc:
                if "no_path" not in str(exc) and "path_too_long" not in str(exc):
                    raise
                last_error = exc
                continue
            except TimeoutError as exc:
                last_error = exc
                continue

            target = state.entities.get(target_id)
            player = find_player(state)
            if target is None:
                raise AssertionError(f"move_until_entity_in_range: target vanished: {target_id}")
            if player is None:
                raise AssertionError("move_until_entity_in_range: player not found")
            if dict_distance(player["position"], target["position"]) <= stop_distance:
                return

    target = state.entities.get(target_id)
    player = find_player(state)
    detail = f"; last_error={last_error}" if last_error is not None else ""
    raise TimeoutError(
        f"move_until_entity_in_range exhausted {max_ticks} ticks toward {target_id}; "
        f"player={(player or {}).get('position')} target={(target or {}).get('position')}{detail}"
    )


def range_candidate_positions(
    player_pos: dict[str, Any],
    target_pos: dict[str, Any],
    stop_distance: float,
    helpers: dict[str, Any] | None = None,
) -> list[dict[str, float]]:
    helpers = _require_helpers(helpers)
    dict_distance = helpers["dict_distance"]
    radius = max(1.25, stop_distance * 0.85)
    player_x = float(player_pos["x"])
    player_y = float(player_pos["y"])
    target_x = float(target_pos["x"])
    target_y = float(target_pos["y"])
    base_x = player_x - target_x
    base_y = player_y - target_y
    base_len = math.hypot(base_x, base_y)
    if base_len <= 0.000001:
        base_x, base_y, base_len = 1.0, 0.0, 1.0
    base_angle = math.atan2(base_y, base_x)
    offsets = [0.0, math.pi / 4, -math.pi / 4, math.pi / 2, -math.pi / 2, math.pi, 3 * math.pi / 4, -3 * math.pi / 4]
    candidates: list[dict[str, float]] = []
    seen: set[tuple[int, int]] = set()
    for offset in offsets:
        angle = base_angle + offset
        pos = {
            "x": round(target_x + math.cos(angle) * radius, 3),
            "y": round(target_y + math.sin(angle) * radius, 3),
        }
        key = (int(round(pos["x"] * 1000)), int(round(pos["y"] * 1000)))
        if key in seen:
            continue
        seen.add(key)
        candidates.append(pos)
    candidates.sort(key=lambda pos: dict_distance(player_pos, pos))
    return candidates


def derived_walk_max_ticks(
    state: RuntimeState,
    target_pos: dict[str, Any],
    requested: int,
    helpers: dict[str, Any] | None = None,
) -> int:
    helpers = _require_helpers(helpers)
    find_player = helpers["find_player"]
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
    helpers: dict[str, Any] | None = None,
) -> None:
    helpers = _require_helpers(helpers)
    find_player = helpers["find_player"]
    wait_for_player_move_or_accept = helpers["wait_for_player_move_or_accept"]
    pump_one = helpers["pump_one"]
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


async def wait_for_player_move_or_accept(
    ws,
    state: RuntimeState,
    before: dict[str, Any],
    message_id: str,
    loop,
    helpers: dict[str, Any] | None = None,
) -> bool:
    helpers = _require_helpers(helpers)
    find_player = helpers["find_player"]
    pump_one = helpers["pump_one"]
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
