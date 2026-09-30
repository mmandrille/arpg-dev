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


MANIFEST = json.loads((ROOT / "assets/manifests/assets.v0.json").read_text(encoding="utf-8"))["assets"]


def _all_ground_asset_ids() -> list[str]:
    ids = [v["asset_id"] for v in DRESSING["plaza"]["tile_variants"]]
    ids += [v["asset_id"] for v in DRESSING["plaza"]["rim"]["tile_variants"]]
    ids += [v["asset_id"] for v in DRESSING["edge"]["tile_variants"]]
    ids += [v["asset_id"] for v in DRESSING["scatter"]["patches"]]
    ids += [v["asset_id"] for v in DRESSING["scatter"]["rocks"]]
    return ids


def test_anchors_match_the_world_preset_exactly() -> None:
    ids = [a["id"] for a in DRESSING["anchors"]]
    assert len(ids) == len(set(ids)), "dressing.anchors ids must be unique (a dict would hide duplicates)"
    anchors = {a["id"]: _xy(a["position"]) for a in DRESSING["anchors"]}
    assert anchors == _gameplay_points(), "dressing.anchors must list every world-preset gameplay point"


def test_service_path_targets_are_known_anchors() -> None:
    anchors = {a["id"] for a in DRESSING["anchors"]}
    # An empty list is valid: service paths are off when no target is configured.
    for target in DRESSING["service_paths"]["targets"]:
        assert target in anchors, f"service path target {target} is not an anchor"


def test_ground_asset_ids_exist_in_the_manifest() -> None:
    for asset_id in _all_ground_asset_ids():
        assert asset_id in MANIFEST, f"{asset_id} missing from the asset manifest"
        assert MANIFEST[asset_id]["type"] == "environment"


def test_scatter_fence_clearance_is_positive() -> None:
    # The planner excludes the ring [fence - clearance, fence + clearance] (unit-tested in GDScript),
    # so the only thing to pin in data is that the clearance exists and the scatter radius is sane.
    scatter = DRESSING["scatter"]
    assert float(scatter["fence_clearance_m"]) > 0.0
    assert float(scatter["radius_m"]) > 0.0


def test_scatter_variants_are_well_formed() -> None:
    for group in ("patches", "rocks"):
        for v in DRESSING["scatter"][group]:
            assert v["weight"] >= 1 and v["radius_m"] > 0.0
            assert 0.0 < v["scale_min"] <= v["scale_max"], f"{v['asset_id']} scale range"
