#!/usr/bin/env python3
"""Inject the shared humanoid rig into supplied static hero GLBs.

The v274 hero models arrive as static meshes. This tool preserves their mesh,
material, texture, and node structure, then appends the same eight skin joints
used by the generated humanoid so the existing Godot character animation clips
can drive them.
"""
from __future__ import annotations

import json
import math
import struct
import sys
from dataclasses import dataclass
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
REQUIRED_BONES = ["root", "spine", "arm_l", "hand_l", "arm_r", "hand_r", "leg_l", "leg_r"]

HEROES = {
    "barbarian": (
        "assets/characters/barbarian/goliath_barbarian.glb",
        "client/assets/characters/barbarian/barbarian.glb",
    ),
    "paladin": (
        "assets/characters/paladin/knight.glb",
        "client/assets/characters/paladin/paladin.glb",
    ),
    "rogue": (
        "assets/characters/rogue/assasine.glb",
        "client/assets/characters/rogue/rogue.glb",
    ),
    "ranger": (
        "assets/characters/ranger/green_hood.glb",
        "client/assets/characters/ranger/ranger.glb",
    ),
    "sorcerer": (
        "assets/characters/sorcerer/mage.glb",
        "client/assets/characters/sorcerer/sorcerer.glb",
    ),
}
RANGER_REST_POSE_DEGREES = 82.0
RANGER_REST_POSE_SHOULDER_RATIO = 0.12


@dataclass(frozen=True)
class ChunkedGlb:
    gltf: dict
    bin_blob: bytes


def _pad_bytes(data: bytearray, fill: int = 0) -> None:
    while len(data) % 4 != 0:
        data.append(fill)


def parse_glb(data: bytes) -> ChunkedGlb:
    if len(data) < 20 or data[0:4] != b"glTF":
        raise ValueError("not a GLB file")
    version, _length = struct.unpack_from("<II", data, 4)
    if version != 2:
        raise ValueError(f"unsupported GLB version {version}")
    offset = 12
    gltf: dict | None = None
    bin_blob = b""
    while offset + 8 <= len(data):
        chunk_len, chunk_type = struct.unpack_from("<II", data, offset)
        offset += 8
        chunk = data[offset : offset + chunk_len]
        offset += chunk_len
        if chunk_type == 0x4E4F534A:  # JSON
            gltf = json.loads(chunk.decode("utf-8"))
        elif chunk_type == 0x004E4942:  # BIN
            bin_blob = bytes(chunk)
    if gltf is None:
        raise ValueError("GLB has no JSON chunk")
    if len(gltf.get("buffers", [])) != 1:
        raise ValueError("only single-buffer GLBs are supported")
    return ChunkedGlb(gltf=gltf, bin_blob=bin_blob)


def write_glb(gltf: dict, bin_blob: bytes) -> bytes:
    json_bytes = bytearray(json.dumps(gltf, sort_keys=True, separators=(",", ":")).encode("utf-8"))
    _pad_bytes(json_bytes, 0x20)
    bin_bytes = bytearray(bin_blob)
    _pad_bytes(bin_bytes)
    json_chunk = struct.pack("<II", len(json_bytes), 0x4E4F534A) + bytes(json_bytes)
    bin_chunk = struct.pack("<II", len(bin_bytes), 0x004E4942) + bytes(bin_bytes)
    total = 12 + len(json_chunk) + len(bin_chunk)
    return b"glTF" + struct.pack("<II", 2, total) + json_chunk + bin_chunk


def _accessor_element_count(accessor: dict) -> int:
    return {
        "SCALAR": 1,
        "VEC2": 2,
        "VEC3": 3,
        "VEC4": 4,
        "MAT4": 16,
    }[str(accessor["type"])]


def _component_size(component_type: int) -> int:
    return {
        5120: 1,
        5121: 1,
        5122: 2,
        5123: 2,
        5125: 4,
        5126: 4,
    }[component_type]


def read_position_accessor(gltf: dict, bin_blob: bytes, accessor_index: int) -> list[tuple[float, float, float]]:
    accessor = gltf["accessors"][accessor_index]
    if accessor.get("componentType") != 5126 or accessor.get("type") != "VEC3":
        raise ValueError(f"POSITION accessor {accessor_index} must be float VEC3")
    view = gltf["bufferViews"][accessor["bufferView"]]
    count = int(accessor["count"])
    elem_size = _component_size(5126) * _accessor_element_count(accessor)
    stride = int(view.get("byteStride", elem_size))
    start = int(view.get("byteOffset", 0)) + int(accessor.get("byteOffset", 0))
    out: list[tuple[float, float, float]] = []
    for i in range(count):
        off = start + i * stride
        out.append(struct.unpack_from("<fff", bin_blob, off))
    return out


def write_vec3_accessor(gltf: dict, bin_buf: bytearray, accessor_index: int, values: list[tuple[float, float, float]]) -> None:
    accessor = gltf["accessors"][accessor_index]
    if accessor.get("componentType") != 5126 or accessor.get("type") != "VEC3":
        raise ValueError(f"accessor {accessor_index} must be float VEC3")
    if int(accessor["count"]) != len(values):
        raise ValueError(f"accessor {accessor_index} count mismatch")
    view = gltf["bufferViews"][accessor["bufferView"]]
    elem_size = _component_size(5126) * _accessor_element_count(accessor)
    stride = int(view.get("byteStride", elem_size))
    start = int(view.get("byteOffset", 0)) + int(accessor.get("byteOffset", 0))
    for i, value in enumerate(values):
        struct.pack_into("<fff", bin_buf, start + i * stride, *value)
    accessor["min"] = [min(v[i] for v in values) for i in range(3)]
    accessor["max"] = [max(v[i] for v in values) for i in range(3)]


def _bounds(positions_by_accessor: dict[int, list[tuple[float, float, float]]]) -> tuple[list[float], list[float]]:
    positions = [p for values in positions_by_accessor.values() for p in values]
    if not positions:
        raise ValueError("GLB has no POSITION data")
    mins = [min(p[i] for p in positions) for i in range(3)]
    maxs = [max(p[i] for p in positions) for i in range(3)]
    return mins, maxs


def _apply_ranger_rest_pose(gltf: dict, bin_buf: bytearray) -> None:
    positions_by_accessor: dict[int, list[tuple[float, float, float]]] = {}
    normal_accessors: list[tuple[int, int]] = []
    for mesh in gltf.get("meshes", []):
        for primitive in mesh.get("primitives", []):
            attrs = primitive.get("attributes", {})
            if "POSITION" not in attrs:
                continue
            position_accessor = int(attrs["POSITION"])
            positions_by_accessor.setdefault(position_accessor, read_position_accessor(gltf, bytes(bin_buf), position_accessor))
            if "NORMAL" in attrs:
                normal_accessors.append((position_accessor, int(attrs["NORMAL"])))
    mins, maxs = _bounds(positions_by_accessor)
    angles_by_accessor: dict[int, list[float | None]] = {}
    for accessor_index, positions in positions_by_accessor.items():
        transformed: list[tuple[float, float, float]] = []
        angles: list[float | None] = []
        for pos in positions:
            angle = _ranger_arm_fold_angle(pos, mins, maxs)
            angles.append(angle)
            transformed.append(_rotate_ranger_arm_position(pos, mins, maxs, angle) if angle is not None else pos)
        angles_by_accessor[accessor_index] = angles
        write_vec3_accessor(gltf, bin_buf, accessor_index, transformed)

    for position_accessor, normal_accessor in normal_accessors:
        angles = angles_by_accessor[position_accessor]
        normals = read_position_accessor(gltf, bytes(bin_buf), normal_accessor)
        if len(normals) != len(angles):
            continue
        transformed_normals = [
            _rotate_xy_vec(normal, angle) if angle is not None else normal
            for normal, angle in zip(normals, angles)
        ]
        write_vec3_accessor(gltf, bin_buf, normal_accessor, transformed_normals)


def _ranger_arm_fold_angle(pos: tuple[float, float, float], mins: list[float], maxs: list[float]) -> float | None:
    x, y, _z = pos
    cx = (mins[0] + maxs[0]) * 0.5
    width = max(maxs[0] - mins[0], 0.001)
    height = max(maxs[1] - mins[1], 0.001)
    yn = (y - mins[1]) / height
    side = x - cx
    if 0.60 <= yn <= 0.84 and abs(side) >= width * 0.20:
        return -math.copysign(math.radians(RANGER_REST_POSE_DEGREES), side)
    return None


def _rotate_ranger_arm_position(
    pos: tuple[float, float, float],
    mins: list[float],
    maxs: list[float],
    angle: float,
) -> tuple[float, float, float]:
    x, y, z = pos
    cx = (mins[0] + maxs[0]) * 0.5
    width = max(maxs[0] - mins[0], 0.001)
    height = max(maxs[1] - mins[1], 0.001)
    side = 1.0 if x >= cx else -1.0
    pivot = (cx + side * width * RANGER_REST_POSE_SHOULDER_RATIO, mins[1] + height * 0.76)
    rx, ry = _rotate_xy((x, y), pivot, angle)
    return (rx, ry, z)


def _rotate_xy(point: tuple[float, float], pivot: tuple[float, float], angle: float) -> tuple[float, float]:
    c = math.cos(angle)
    s = math.sin(angle)
    dx = point[0] - pivot[0]
    dy = point[1] - pivot[1]
    return (pivot[0] + dx * c - dy * s, pivot[1] + dx * s + dy * c)


def _rotate_xy_vec(vec: tuple[float, float, float], angle: float) -> tuple[float, float, float]:
    c = math.cos(angle)
    s = math.sin(angle)
    return (vec[0] * c - vec[1] * s, vec[0] * s + vec[1] * c, vec[2])


def _joint_globals(mins: list[float], maxs: list[float]) -> list[tuple[float, float, float]]:
    cx = (mins[0] + maxs[0]) * 0.5
    cy = mins[1]
    cz = (mins[2] + maxs[2]) * 0.5
    width = max(maxs[0] - mins[0], 0.001)
    height = max(maxs[1] - mins[1], 0.001)
    depth = max(maxs[2] - mins[2], 0.001)
    shoulder_x = width * 0.30
    leg_x = width * 0.12
    hand_z = depth * 0.18
    return [
        (cx, cy, cz),
        (cx, cy + height * 0.575, cz),
        (cx - shoulder_x, cy + height * 0.75, cz),
        (cx - shoulder_x, cy + height * 0.41, cz + hand_z),
        (cx + shoulder_x, cy + height * 0.75, cz),
        (cx + shoulder_x, cy + height * 0.41, cz + hand_z),
        (cx - leg_x, cy + height * 0.45, cz),
        (cx + leg_x, cy + height * 0.45, cz),
    ]


def _joint_nodes(joint_globals: list[tuple[float, float, float]], offset: int) -> list[dict]:
    root, spine, arm_l, hand_l, arm_r, hand_r, leg_l, leg_r = joint_globals
    return [
        {"name": "root", "translation": list(root), "children": [offset + 1, offset + 6, offset + 7]},
        {"name": "spine", "translation": _delta(root, spine), "children": [offset + 2, offset + 4]},
        {"name": "arm_l", "translation": _delta(spine, arm_l), "children": [offset + 3]},
        {"name": "hand_l", "translation": _delta(arm_l, hand_l)},
        {"name": "arm_r", "translation": _delta(spine, arm_r), "children": [offset + 5]},
        {"name": "hand_r", "translation": _delta(arm_r, hand_r)},
        {"name": "leg_l", "translation": _delta(root, leg_l)},
        {"name": "leg_r", "translation": _delta(root, leg_r)},
    ]


def _delta(a: tuple[float, float, float], b: tuple[float, float, float]) -> list[float]:
    return [b[0] - a[0], b[1] - a[1], b[2] - a[2]]


def _inverse_translation_matrix(global_pos: tuple[float, float, float]) -> list[float]:
    x, y, z = global_pos
    return [
        1.0, 0.0, 0.0, 0.0,
        0.0, 1.0, 0.0, 0.0,
        0.0, 0.0, 1.0, 0.0,
        -x, -y, -z, 1.0,
    ]


def _joint_for_vertex(pos: tuple[float, float, float], mins: list[float], maxs: list[float]) -> int:
    x, y, _z = pos
    cx = (mins[0] + maxs[0]) * 0.5
    width = max(maxs[0] - mins[0], 0.001)
    height = max(maxs[1] - mins[1], 0.001)
    yn = (y - mins[1]) / height
    side = x - cx
    arm_threshold = width * 0.20
    if 0.32 <= yn <= 0.82 and abs(side) >= arm_threshold:
        return 4 if side > 0.0 else 2
    if yn < 0.52:
        return 7 if side > 0.0 else 6
    return 1


def _append_buffer_view(gltf: dict, bin_buf: bytearray, payload: bytes, *, target: int | None = None) -> int:
    _pad_bytes(bin_buf)
    offset = len(bin_buf)
    bin_buf.extend(payload)
    view: dict = {"buffer": 0, "byteOffset": offset, "byteLength": len(payload)}
    if target is not None:
        view["target"] = target
    gltf.setdefault("bufferViews", []).append(view)
    return len(gltf["bufferViews"]) - 1


def _append_accessor(gltf: dict, accessor: dict) -> int:
    gltf.setdefault("accessors", []).append(accessor)
    return len(gltf["accessors"]) - 1


def rig_glb_bytes(data: bytes, *, hero_id: str = "") -> bytes:
    parsed = parse_glb(data)
    gltf = parsed.gltf
    if gltf.get("skins"):
        raise ValueError("source GLB is already skinned")
    bin_buf = bytearray(parsed.bin_blob)
    if hero_id == "ranger":
        _apply_ranger_rest_pose(gltf, bin_buf)

    positions_by_accessor: dict[int, list[tuple[float, float, float]]] = {}
    primitives: list[dict] = []
    for mesh in gltf.get("meshes", []):
        for primitive in mesh.get("primitives", []):
            attrs = primitive.get("attributes", {})
            if "POSITION" not in attrs:
                continue
            position_accessor = int(attrs["POSITION"])
            positions_by_accessor.setdefault(
                position_accessor,
                read_position_accessor(gltf, bytes(bin_buf), position_accessor),
            )
            primitives.append(primitive)
    mins, maxs = _bounds(positions_by_accessor)
    joint_globals = _joint_globals(mins, maxs)

    first_joint_node = len(gltf.setdefault("nodes", []))
    gltf["nodes"].extend(_joint_nodes(joint_globals, first_joint_node))
    joint_indices = list(range(first_joint_node, first_joint_node + len(REQUIRED_BONES)))

    ibm_payload = bytearray()
    for pos in joint_globals:
        ibm_payload.extend(struct.pack("<16f", *_inverse_translation_matrix(pos)))
    ibm_view = _append_buffer_view(gltf, bin_buf, bytes(ibm_payload))
    ibm_accessor = _append_accessor(
        gltf,
        {"bufferView": ibm_view, "componentType": 5126, "count": len(REQUIRED_BONES), "type": "MAT4"},
    )

    skin_index = len(gltf.setdefault("skins", []))
    gltf["skins"].append({
        "joints": joint_indices,
        "inverseBindMatrices": ibm_accessor,
        "skeleton": first_joint_node,
    })

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
        joints_accessor = _append_accessor(
            gltf,
            {
                "bufferView": joints_view,
                "componentType": 5123,
                "count": len(positions),
                "type": "VEC4",
                "min": [0, 0, 0, 0],
                "max": [len(REQUIRED_BONES) - 1, 0, 0, 0],
            },
        )
        weights_accessor = _append_accessor(
            gltf,
            {
                "bufferView": weights_view,
                "componentType": 5126,
                "count": len(positions),
                "type": "VEC4",
                "min": [0.0, 0.0, 0.0, 0.0],
                "max": [1.0, 0.0, 0.0, 0.0],
            },
        )
        primitive["attributes"]["JOINTS_0"] = joints_accessor
        primitive["attributes"]["WEIGHTS_0"] = weights_accessor

    for node in gltf.get("nodes", [])[:first_joint_node]:
        if "mesh" in node:
            node["skin"] = skin_index

    scene_index = int(gltf.get("scene", 0))
    scenes = gltf.setdefault("scenes", [{"nodes": []}])
    scene_nodes = scenes[scene_index].setdefault("nodes", [])
    if first_joint_node not in scene_nodes:
        scene_nodes.append(first_joint_node)
    gltf["buffers"][0]["byteLength"] = len(bin_buf)
    gltf["asset"]["generator"] = "arpg-dev/tools/assets/rig_hero_glbs.py"
    return write_glb(gltf, bytes(bin_buf))


def rig_file(source: Path, target: Path) -> None:
    target.parent.mkdir(parents=True, exist_ok=True)
    hero_id = target.parent.name
    target.write_bytes(rig_glb_bytes(source.read_bytes(), hero_id=hero_id))


def main() -> int:
    for class_id, (source_rel, target_rel) in HEROES.items():
        source = ROOT / source_rel
        target = ROOT / target_rel
        rig_file(source, target)
        print(f"rigged {class_id}: {source_rel} -> {target_rel}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
