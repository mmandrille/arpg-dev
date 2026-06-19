#!/usr/bin/env python3
"""Inject a simple quadruped rig into supplied static monster GLBs."""
from __future__ import annotations

import struct
import sys
from pathlib import Path

from tools.assets.rig_hero_glbs import (
    _append_accessor,
    _append_buffer_view,
    _bounds,
    _inverse_translation_matrix,
    parse_glb,
    read_position_accessor,
    write_glb,
)
from tools.assets.validate_assets import parse_glb_skin_joint_names

ROOT = Path(__file__).resolve().parents[2]
QUADRUPED_BONES = ["root", "spine", "head", "tail", "leg_fl", "leg_fr", "leg_bl", "leg_br"]

QUADRUPED_MONSTERS = {
    "quadruped_predator": (
        "assets/monsters/purple_fantasy/evil_fox_monster.glb",
        "client/assets/monsters/purple_fantasy/evil_fox_monster.glb",
    ),
}


def _joint_globals(mins: list[float], maxs: list[float]) -> list[tuple[float, float, float]]:
    cx = (mins[0] + maxs[0]) * 0.5
    cy = (mins[1] + maxs[1]) * 0.5
    cz = (mins[2] + maxs[2]) * 0.5
    width = maxs[0] - mins[0]
    depth = maxs[2] - mins[2]
    return [
        (cx, mins[1], cz),
        (cx, cy, cz),
        (cx, cy + (maxs[1] - mins[1]) * 0.15, mins[2] + depth * 0.12),
        (cx, cy, maxs[2] - depth * 0.08),
        (cx - width * 0.28, cy * 0.55, mins[2] + depth * 0.28),
        (cx + width * 0.28, cy * 0.55, mins[2] + depth * 0.28),
        (cx - width * 0.28, cy * 0.55, maxs[2] - depth * 0.25),
        (cx + width * 0.28, cy * 0.55, maxs[2] - depth * 0.25),
    ]


def _joint_nodes(joints: list[tuple[float, float, float]], first_index: int) -> list[dict]:
    return [
        {"name": "root", "translation": list(joints[0]), "children": [first_index + 1, first_index + 4, first_index + 5, first_index + 6, first_index + 7]},
        {"name": "spine", "translation": [0.0, joints[1][1] - joints[0][1], 0.0], "children": [first_index + 2, first_index + 3]},
        {"name": "head", "translation": [joints[2][0] - joints[1][0], joints[2][1] - joints[1][1], joints[2][2] - joints[1][2]]},
        {"name": "tail", "translation": [joints[3][0] - joints[1][0], joints[3][1] - joints[1][1], joints[3][2] - joints[1][2]]},
        {"name": "leg_fl", "translation": [joints[4][0] - joints[0][0], joints[4][1] - joints[0][1], joints[4][2] - joints[0][2]]},
        {"name": "leg_fr", "translation": [joints[5][0] - joints[0][0], joints[5][1] - joints[0][1], joints[5][2] - joints[0][2]]},
        {"name": "leg_bl", "translation": [joints[6][0] - joints[0][0], joints[6][1] - joints[0][1], joints[6][2] - joints[0][2]]},
        {"name": "leg_br", "translation": [joints[7][0] - joints[0][0], joints[7][1] - joints[0][1], joints[7][2] - joints[0][2]]},
    ]


def _joint_for_vertex(pos: tuple[float, float, float], mins: list[float], maxs: list[float]) -> int:
    x, y, z = pos
    height = max(maxs[1] - mins[1], 0.001)
    yn = (y - mins[1]) / height
    if yn > 0.62:
        return 2 if z < (mins[2] + (maxs[2] - mins[2]) * 0.36) else 3
    if yn < 0.42:
        front = z < (mins[2] + maxs[2]) * 0.5
        left = x < (mins[0] + maxs[0]) * 0.5
        if front and left:
            return 4
        if front:
            return 5
        if left:
            return 6
        return 7
    return 1


def rig_quadruped_glb_bytes(data: bytes) -> bytes:
    parsed = parse_glb(data)
    gltf = parsed.gltf
    if gltf.get("skins"):
        raise ValueError("source GLB is already skinned")
    bin_buf = bytearray(parsed.bin_blob)
    positions_by_accessor: dict[int, list[tuple[float, float, float]]] = {}
    primitives: list[dict] = []
    for mesh in gltf.get("meshes", []):
        for primitive in mesh.get("primitives", []):
            attrs = primitive.get("attributes", {})
            if "POSITION" not in attrs:
                continue
            position_accessor = int(attrs["POSITION"])
            positions_by_accessor.setdefault(position_accessor, read_position_accessor(gltf, bytes(bin_buf), position_accessor))
            primitives.append(primitive)
    mins, maxs = _bounds(positions_by_accessor)
    joint_globals = _joint_globals(mins, maxs)
    first_joint_node = len(gltf.setdefault("nodes", []))
    gltf["nodes"].extend(_joint_nodes(joint_globals, first_joint_node))
    joint_indices = list(range(first_joint_node, first_joint_node + len(QUADRUPED_BONES)))

    ibm_payload = bytearray()
    for pos in joint_globals:
        ibm_payload.extend(struct.pack("<16f", *_inverse_translation_matrix(pos)))
    ibm_view = _append_buffer_view(gltf, bin_buf, bytes(ibm_payload))
    ibm_accessor = _append_accessor(gltf, {"bufferView": ibm_view, "componentType": 5126, "count": len(QUADRUPED_BONES), "type": "MAT4"})
    skin_index = len(gltf.setdefault("skins", []))
    gltf["skins"].append({"joints": joint_indices, "inverseBindMatrices": ibm_accessor, "skeleton": first_joint_node})

    for primitive in primitives:
        positions = positions_by_accessor[int(primitive["attributes"]["POSITION"])]
        joints_payload = bytearray()
        weights_payload = bytearray()
        for pos in positions:
            joint = _joint_for_vertex(pos, mins, maxs)
            joints_payload.extend(struct.pack("<HHHH", joint, 0, 0, 0))
            weights_payload.extend(struct.pack("<ffff", 1.0, 0.0, 0.0, 0.0))
        joints_view = _append_buffer_view(gltf, bin_buf, bytes(joints_payload), target=34962)
        weights_view = _append_buffer_view(gltf, bin_buf, bytes(weights_payload), target=34962)
        primitive["attributes"]["JOINTS_0"] = _append_accessor(gltf, {"bufferView": joints_view, "componentType": 5123, "count": len(positions), "type": "VEC4"})
        primitive["attributes"]["WEIGHTS_0"] = _append_accessor(gltf, {"bufferView": weights_view, "componentType": 5126, "count": len(positions), "type": "VEC4"})

    for node in gltf.get("nodes", [])[:first_joint_node]:
        if "mesh" in node:
            node["skin"] = skin_index
    scene_nodes = gltf.setdefault("scenes", [{"nodes": []}])[int(gltf.get("scene", 0))].setdefault("nodes", [])
    if first_joint_node not in scene_nodes:
        scene_nodes.append(first_joint_node)
    gltf["buffers"][0]["byteLength"] = len(bin_buf)
    gltf["asset"]["generator"] = "arpg-dev/tools/assets/rig_quadruped_monster_glbs.py"
    return write_glb(gltf, bytes(bin_buf))


def rig_quadruped_file(source: Path, target: Path) -> None:
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes(rig_quadruped_glb_bytes(source.read_bytes()))


def validate_target(target: Path) -> None:
    joints = parse_glb_skin_joint_names(target)
    missing = sorted(set(QUADRUPED_BONES) - (joints or set()))
    if missing:
        raise ValueError(f"{target}: missing joints {missing}")


def main() -> int:
    for monster_id, (source_rel, target_rel) in QUADRUPED_MONSTERS.items():
        source = ROOT / source_rel
        target = ROOT / target_rel
        rig_quadruped_file(source, target)
        validate_target(target)
        print(f"rigged {monster_id}: {source_rel} -> {target_rel}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
