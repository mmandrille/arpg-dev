"""Extraction-independence proof for movement_runtime (v152).

These tests enforce the property the coupling ratchet's regex cannot: the
extracted module imports and unit-tests WITHOUT importing tools.bot.run, and
its runtime services arrive through a typed BotContext that a test can stub.
"""
from __future__ import annotations

import asyncio
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]


def test_movement_runtime_imports_without_run() -> None:
    code = (
        "import sys, tools.bot.movement_runtime as m; "
        "assert 'tools.bot.run' not in sys.modules, "
        "sorted(k for k in sys.modules if k.startswith('tools.bot')); "
        "assert hasattr(m, 'walk_toward')"
    )
    result = subprocess.run(
        [sys.executable, "-c", code], cwd=ROOT, text=True, capture_output=True
    )
    assert result.returncode == 0, result.stderr


def test_range_candidate_positions_orders_closest_first() -> None:
    from tools.bot.movement_runtime import range_candidate_positions
    from tools.bot.runtime_queries import dict_distance

    player = {"x": 0.0, "y": 0.0}
    target = {"x": 6.0, "y": 0.0}
    candidates = range_candidate_positions(player, target, stop_distance=1.5)
    assert candidates, "expected at least one ring candidate"
    distances = [dict_distance(player, c) for c in candidates]
    assert distances == sorted(distances)


def test_derived_walk_max_ticks_scales_with_distance() -> None:
    from tools.bot.bot_types import RuntimeState
    from tools.bot.movement_runtime import derived_walk_max_ticks

    state = RuntimeState(local_player_id="p1")
    state.entities["p1"] = {"type": "player", "position": {"x": 0.0, "y": 0.0}}
    near = derived_walk_max_ticks(state, {"x": 1.0, "y": 0.0}, requested=10)
    far = derived_walk_max_ticks(state, {"x": 50.0, "y": 0.0}, requested=10)
    assert far > near >= 40


def test_wait_for_player_move_uses_stub_context() -> None:
    from tools.bot.bot_context import BotContext
    from tools.bot.bot_types import RuntimeState
    from tools.bot.movement_runtime import wait_for_player_move_or_accept

    pumped = {"count": 0}

    async def fake_pump(ws, state, timeout):  # noqa: ANN001
        pumped["count"] += 1

    ctx = BotContext(pump_one=fake_pump)
    state = RuntimeState(local_player_id="p1")
    state.entities["p1"] = {"type": "player", "position": {"x": 5.0, "y": 5.0}}

    async def drive() -> bool:
        loop = asyncio.get_running_loop()
        return await wait_for_player_move_or_accept(
            None, state, {"x": 0.0, "y": 0.0}, "m1", loop, ctx=ctx
        )

    # Player already differs from `before`, so it returns True without pumping.
    assert asyncio.run(drive()) is True
    assert pumped["count"] == 0
