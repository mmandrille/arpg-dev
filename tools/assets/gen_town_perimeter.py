#!/usr/bin/env python3
"""Emit deterministic town perimeter wall entities for worlds.v0.json."""

from __future__ import annotations

import json
import math
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
PRESENTATION_PATH = REPO_ROOT / "shared/assets/town_presentation.v0.json"

HEADING_INDEX = {
    "east": 0,
    "south": 1,
    "west": 2,
    "north": 3,
}


def load_presentation() -> dict:
    with PRESENTATION_PATH.open(encoding="utf-8") as handle:
        return json.load(handle)


def segment_indices_for_heading(segment_count: int, heading: str, gap: int) -> set[int]:
    quarter = segment_count // 4
    center = HEADING_INDEX[heading] * quarter
    half_gap = gap // 2
    blocked = set()
    for offset in range(gap):
        blocked.add((center - half_gap + offset) % segment_count)
    return blocked


def wall_entities(presentation: dict) -> list[dict]:
    center = presentation["center"]
    cx = float(center["x"])
    cy = float(center["y"])
    radius = float(presentation["radius_m"])
    segment_count = int(presentation["segment_count"])
    thickness = float(presentation["wall_thickness"])
    kind = str(presentation["wall_kind"])
    gap = int(presentation["gate_gap_segments"])
    heading = str(presentation["gate_heading"])
    skip = segment_indices_for_heading(segment_count, heading, gap)

    walls: list[dict] = []
    for index in range(segment_count):
        if index in skip:
            continue
        angle_start = (2.0 * math.pi * index) / segment_count
        angle_end = (2.0 * math.pi * (index + 1)) / segment_count
        mid_angle = (angle_start + angle_end) * 0.5
        arc_len = radius * (angle_end - angle_start)
        px = cx + radius * math.cos(mid_angle)
        py = cy + radius * math.sin(mid_angle)
        cos_a = math.cos(mid_angle)
        sin_a = math.sin(mid_angle)
        if abs(cos_a) >= abs(sin_a):
            size_x = max(thickness, arc_len)
            size_y = thickness
        else:
            size_x = thickness
            size_y = max(thickness, arc_len)
        walls.append(
            {
                "type": "wall",
                "kind": kind,
                "position": {"x": round(px, 4), "y": round(py, 4)},
                "size": {"x": round(size_x, 4), "y": round(size_y, 4)},
            }
        )
    return walls


def gate_entity(presentation: dict) -> dict:
    gate = presentation["gate_position"]
    return {
        "type": "interactable",
        "interactable_def_id": "town_exit_gate",
        "position": {"x": float(gate["x"]), "y": float(gate["y"])},
    }


def main() -> None:
    presentation = load_presentation()
    payload = {
        "walls": wall_entities(presentation),
        "gate": gate_entity(presentation),
        "wall_count": len(wall_entities(presentation)),
    }
    print(json.dumps(payload, indent=2))


if __name__ == "__main__":
    main()
