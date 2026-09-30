"""v491: town dressing is presentation only (the server never sees it), so its placement must keep
clear of every gameplay position in the town world preset, stay off the plaza road, and stay inside
the fence. Also cross-checks the plaza centre and gate against the world preset."""
from __future__ import annotations

import json
import math
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
TOWN = json.loads((ROOT / "shared/assets/town_presentation.v0.json").read_text(encoding="utf-8"))
WORLD = json.loads((ROOT / "shared/rules/worlds.v0.json").read_text(encoding="utf-8"))["worlds"]["dungeon_levels"]
DRESSING = TOWN["dressing"]


def _xy(p: dict) -> tuple[float, float]:
    return float(p["x"]), float(p["y"])


def _gameplay_points() -> dict[str, tuple[float, float]]:
    points = {"player_spawn": _xy(WORLD["player"]["position"])}
    for entity in WORLD["entities"]:
        if entity["type"] in ("interactable", "monster"):
            name = entity.get("interactable_def_id") or entity.get("monster_def_id")
            points[name] = _xy(entity["position"])
    return points


def test_plaza_centre_and_gate_match_world_preset() -> None:
    points = _gameplay_points()
    assert _xy(TOWN["center"]) == points["stairs_down"], "plaza centre must be the stairs down"
    assert _xy(TOWN["gate_position"]) == points["town_exit_gate"]


def test_props_keep_clear_of_gameplay_positions() -> None:
    clearance = float(DRESSING["min_clearance_m"])
    for prop in DRESSING["props"]:
        px, py = _xy(prop["position"])
        for name, (gx, gy) in _gameplay_points().items():
            assert math.dist((px, py), (gx, gy)) >= clearance, f"{prop['asset_id']} at {px},{py} blocks {name}"


def test_props_stay_off_the_road_and_inside_the_fence() -> None:
    cx, cy = _xy(TOWN["center"])
    gx, gy = _xy(TOWN["gate_position"])
    half = float(DRESSING["plaza"]["path_width_m"]) / 2.0
    for prop in DRESSING["props"]:
        px, py = _xy(prop["position"])
        on_road = abs(px - cx) < half and min(cy, gy) <= py <= max(cy, gy)
        assert not on_road, f"{prop['asset_id']} at {px},{py} sits on the road to the gate"
        assert math.dist((px, py), (cx, cy)) <= float(TOWN["radius_m"]) - 0.5, f"{prop['asset_id']} outside the fence"
