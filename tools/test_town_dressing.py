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


def _xy(p: dict | tuple[float, float]) -> tuple[float, float]:
    if isinstance(p, dict):
        return float(p["x"]), float(p["y"])
    return float(p[0]), float(p[1])


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
CLASS_PRESENTATIONS = json.loads((ROOT / "shared/assets/class_presentations.v0.json").read_text(encoding="utf-8"))["classes"]


def _all_ground_asset_ids() -> list[str]:
    ids = [v["asset_id"] for v in DRESSING["plaza"]["tile_variants"]]
    ids += [v["asset_id"] for v in DRESSING["plaza"]["rim"]["tile_variants"]]
    ids += [v["asset_id"] for v in DRESSING["edge"]["tile_variants"]]
    ids += [v["asset_id"] for v in DRESSING["scatter"]["patches"]]
    ids += [v["asset_id"] for v in DRESSING["scatter"]["rocks"]]
    return ids


def _point_segment_distance(point: tuple[float, float], a: tuple[float, float], b: tuple[float, float]) -> float:
    dx, dy = b[0] - a[0], b[1] - a[1]
    length2 = dx * dx + dy * dy
    if length2 == 0.0:
        return math.dist(point, a)
    t = max(0.0, min(1.0, ((point[0] - a[0]) * dx + (point[1] - a[1]) * dy) / length2))
    return math.dist(point, (a[0] + t * dx, a[1] + t * dy))


def _orientation(a: tuple[float, float], b: tuple[float, float], c: tuple[float, float]) -> float:
    return (b[0] - a[0]) * (c[1] - a[1]) - (b[1] - a[1]) * (c[0] - a[0])


def _segments_intersect(a: tuple[float, float], b: tuple[float, float], c: tuple[float, float], d: tuple[float, float]) -> bool:
    o1, o2, o3, o4 = _orientation(a, b, c), _orientation(a, b, d), _orientation(c, d, a), _orientation(c, d, b)
    if ((o1 > 1e-9 and o2 < -1e-9) or (o1 < -1e-9 and o2 > 1e-9)) and ((o3 > 1e-9 and o4 < -1e-9) or (o3 < -1e-9 and o4 > 1e-9)):
        return True

    def on_segment(point: tuple[float, float], start: tuple[float, float], end: tuple[float, float]) -> bool:
        return min(start[0], end[0]) - 1e-9 <= point[0] <= max(start[0], end[0]) + 1e-9 and min(start[1], end[1]) - 1e-9 <= point[1] <= max(start[1], end[1]) + 1e-9

    return ((abs(o1) <= 1e-9 and on_segment(c, a, b)) or (abs(o2) <= 1e-9 and on_segment(d, a, b))
            or (abs(o3) <= 1e-9 and on_segment(a, c, d)) or (abs(o4) <= 1e-9 and on_segment(b, c, d)))


def _segment_distance(a: tuple[float, float], b: tuple[float, float], c: tuple[float, float], d: tuple[float, float]) -> float:
    if _segments_intersect(a, b, c, d):
        return 0.0
    return min(_point_segment_distance(a, c, d), _point_segment_distance(b, c, d),
               _point_segment_distance(c, a, b), _point_segment_distance(d, a, b))


def _ambient_segments(actor: dict) -> list[tuple[tuple[float, float], tuple[float, float]]]:
    if actor["mode"] == "idle":
        point = _xy(actor["position"])
        return [(point, point)]
    points = [_xy(p) for p in actor["path"]]
    return list(zip(points, points[1:] + points[:1]))


def test_ambient_roster_keeps_footprints_and_patrols_clear() -> None:
    life = DRESSING["ambient_life"]
    assert life["enabled"]
    assert 0 < len(life["actors"]) <= 4
    ids = [actor["id"] for actor in life["actors"]]
    assert len(ids) == len(set(ids)), "ambient actor ids must be unique"
    anchors = {name: _xy(pos) for name, pos in _gameplay_points().items()}
    center, gate = _xy(TOWN["center"]), _xy(TOWN["gate_position"])
    gate_clearance = float(DRESSING["plaza"]["path_width_m"]) / 2.0 + float(life["actor_clearance_m"]) + float(life["gate_approach_clearance_m"])
    service_half_width = float(DRESSING["service_paths"]["path_width_m"]) / 2.0 + float(life["actor_clearance_m"]) + float(life["service_path_margin_m"])
    for actor in life["actors"]:
        resolved = CLASS_PRESENTATIONS[actor["class_id"]]["model"]["asset_id"]
        assert MANIFEST[resolved]["type"] == "character", f"{resolved} must be a registered character"
        segments = _ambient_segments(actor)
        vertices = [point for segment in segments for point in segment]
        for point in vertices:
            assert math.dist(point, center) <= float(TOWN["radius_m"]) - float(life["fence_clearance_m"]) - float(life["actor_clearance_m"]), f"{actor['id']} is too close to palisade"
        for obstacle, point in anchors.items():
            distance = min(_point_segment_distance(point, *segment) for segment in segments)
            assert distance >= float(life["anchor_clearance_m"]), f"{actor['id']} route approaches gameplay anchor {obstacle} ({distance:.2f} m)"
        for prop in DRESSING["props"]:
            point = _xy(prop["position"])
            distance = min(_point_segment_distance(point, *segment) for segment in segments)
            assert distance >= float(life["prop_clearance_m"]), f"{actor['id']} route approaches prop {prop['asset_id']} ({distance:.2f} m)"
        for target in DRESSING["service_paths"]["targets"]:
            service = anchors[target]
            distance = min(_segment_distance(*route, center, service) for route in segments)
            assert distance >= service_half_width, f"{actor['id']} route blocks service path {target} ({distance:.2f} m)"
        distance = min(_segment_distance(*route, center, gate) for route in segments)
        assert distance >= gate_clearance, f"{actor['id']} route blocks the gate approach ({distance:.2f} m)"
    for index, actor in enumerate(life["actors"]):
        for other in life["actors"][index + 1:]:
            distance = min(_segment_distance(*first, *second)
                           for first in _ambient_segments(actor) for second in _ambient_segments(other))
            minimum = 2.0 * float(life["actor_clearance_m"])
            assert distance >= minimum, f"{actor['id']} and {other['id']} routes overlap ({distance:.2f} m)"


def test_ambient_clearance_rejects_an_unsafe_anchor_fixture() -> None:
    life = DRESSING["ambient_life"]
    unsafe = {"id": "unsafe_fixture", "mode": "idle", "position": {"x": 20.0, "y": 12.0}}
    vendor = _gameplay_points()["town_vendor"]
    separation = min(_point_segment_distance(vendor, *segment) for segment in _ambient_segments(unsafe))
    assert separation < float(life["anchor_clearance_m"]), "negative fixture must overlap the vendor's reserved space"


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


def test_nature_asset_ids_exist_in_the_manifest() -> None:
    nature = DRESSING["nature"]
    for group in nature["groups"]:
        for item in [*group["landmarks"], *group["variants"]]:
            asset_id = item["asset_id"]
            assert asset_id in MANIFEST, f"{asset_id} missing from the asset manifest"
            assert MANIFEST[asset_id]["type"] == "environment"


def test_nature_landmarks_stay_outside_fence_and_gate() -> None:
    nature = DRESSING["nature"]
    center = _xy(TOWN["center"])
    gate = _xy(TOWN["gate_position"])
    end = _xy(nature["gate_approach"]["end_position"])
    for group in nature["groups"]:
        for landmark in group["landmarks"]:
            point = _xy(landmark["position"])
            radius = float(landmark["footprint_radius_m"])
            assert math.dist(point, center) >= float(TOWN["radius_m"]) + float(nature["fence_clearance_m"]) + radius
            segment = (end[0] - gate[0], end[1] - gate[1])
            length2 = segment[0] ** 2 + segment[1] ** 2
            t = max(0.0, min(1.0, ((point[0] - gate[0]) * segment[0] + (point[1] - gate[1]) * segment[1]) / length2))
            nearest = (gate[0] + t * segment[0], gate[1] + t * segment[1])
            assert math.dist(point, nearest) > float(nature["gate_approach"]["half_width_m"]) + radius


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
