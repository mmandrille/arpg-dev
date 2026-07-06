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
REQUIRED_BONES = [
    "root",
    "spine",
    "chest",
    "neck",
    "head",
    "arm_l",
    "elbow_l",
    "hand_l",
    "arm_r",
    "elbow_r",
    "hand_r",
    "leg_l",
    "knee_l",
    "foot_l",
    "leg_r",
    "knee_r",
    "foot_r",
]

HEROES = {
    "base_human": (
        "assets/characters/base_human/base_human_mesh.glb",
        "client/assets/characters/base_human/base_human.glb",
    ),
}
CANONICAL_RIG_IDS = frozenset({"base_human"})
RANGER_REST_POSE_DEGREES = 82.0
RANGER_REST_POSE_SHOULDER_RATIO = 0.12
HERO_TARGET_HEIGHTS: dict[str, float] = {
    "base_human": 1.97,
}


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


def _read_index_accessor(gltf: dict, bin_blob: bytes, accessor_index: int) -> list[int]:
    accessor = gltf["accessors"][accessor_index]
    if accessor.get("type") != "SCALAR":
        raise ValueError(f"index accessor {accessor_index} must be SCALAR")
    component_type = int(accessor["componentType"])
    if component_type not in (5123, 5125):
        raise ValueError(f"index accessor {accessor_index} must be uint16 or uint32")
    view = gltf["bufferViews"][accessor["bufferView"]]
    count = int(accessor["count"])
    elem_size = _component_size(component_type)
    stride = int(view.get("byteStride", elem_size))
    start = int(view.get("byteOffset", 0)) + int(accessor.get("byteOffset", 0))
    fmt = "<H" if component_type == 5123 else "<I"
    return [struct.unpack_from(fmt, bin_blob, start + i * stride)[0] for i in range(count)]


def _write_index_accessor(gltf: dict, bin_buf: bytearray, accessor_index: int, values: list[int]) -> None:
    accessor = gltf["accessors"][accessor_index]
    component_type = int(accessor["componentType"])
    if component_type not in (5123, 5125):
        raise ValueError(f"index accessor {accessor_index} must be uint16 or uint32")
    view = gltf["bufferViews"][accessor["bufferView"]]
    elem_size = _component_size(component_type)
    stride = int(view.get("byteStride", elem_size))
    start = int(view.get("byteOffset", 0)) + int(accessor.get("byteOffset", 0))
    fmt = "<H" if component_type == 5123 else "<I"
    for i, value in enumerate(values):
        struct.pack_into(fmt, bin_buf, start + i * stride, value)
    accessor["count"] = len(values)
    accessor["max"] = [max(values) if values else 0]


def _cull_apose_tpose_cap_triangles(
    gltf: dict,
    bin_buf: bytearray,
    positions_by_accessor: dict[int, list[tuple[float, float, float]]],
    primitives: list[dict],
    mins: list[float],
    maxs: list[float],
) -> None:
    """Drop leftover horizontal T-pose shoulder caps common on mastjie A-pose exports."""
    cx = (mins[0] + maxs[0]) * 0.5
    width = max(maxs[0] - mins[0], 0.001)
    height = max(maxs[1] - mins[1], 0.001)
    side_threshold = width * 0.42

    for primitive in primitives:
        if "indices" not in primitive:
            continue
        position_accessor = int(primitive["attributes"]["POSITION"])
        positions = positions_by_accessor[position_accessor]
        index_accessor = int(primitive["indices"])
        indices = _read_index_accessor(gltf, bytes(bin_buf), index_accessor)
        kept: list[int] = []
        for tri in range(0, len(indices), 3):
            verts = [positions[indices[tri + i]] for i in range(3)]
            centroid = (
                sum(v[0] for v in verts) / 3.0,
                sum(v[1] for v in verts) / 3.0,
                sum(v[2] for v in verts) / 3.0,
            )
            yn = (centroid[1] - mins[1]) / height
            if 0.47 <= yn <= 0.63 and abs(centroid[0] - cx) >= side_threshold:
                continue
            kept.extend(indices[tri : tri + 3])
        _write_index_accessor(gltf, bin_buf, index_accessor, kept)


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


def _scale_position_accessor(
    gltf: dict,
    bin_buf: bytearray,
    accessor_index: int,
    scale: float,
    origin_y: float,
) -> None:
    positions = read_position_accessor(gltf, bytes(bin_buf), accessor_index)
    accessor = gltf["accessors"][accessor_index]
    view = gltf["bufferViews"][accessor["bufferView"]]
    offset = int(view["byteOffset"])
    scaled: list[tuple[float, float, float]] = []
    for x, y, z in positions:
        sx = x * scale
        sy = origin_y + (y - origin_y) * scale
        sz = z * scale
        scaled.append((sx, sy, sz))
        struct.pack_into("<fff", bin_buf, offset, sx, sy, sz)
        offset += 12
    accessor["min"] = [min(p[i] for p in scaled) for i in range(3)]
    accessor["max"] = [max(p[i] for p in scaled) for i in range(3)]


def _normalize_mesh_height(
    gltf: dict,
    bin_buf: bytearray,
    positions_by_accessor: dict[int, list[tuple[float, float, float]]],
    target_height: float,
) -> None:
    mins, maxs = _bounds(positions_by_accessor)
    height = max(maxs[1] - mins[1], 0.001)
    if abs(height - target_height) < 0.05:
        return
    scale = target_height / height
    origin_y = mins[1]
    for accessor_index in positions_by_accessor.keys():
        _scale_position_accessor(gltf, bin_buf, accessor_index, scale, origin_y)
        positions_by_accessor[accessor_index] = read_position_accessor(gltf, bytes(bin_buf), accessor_index)


def _lerp3(
    a: tuple[float, float, float],
    b: tuple[float, float, float],
    t: float,
) -> tuple[float, float, float]:
    return (
        a[0] + (b[0] - a[0]) * t,
        a[1] + (b[1] - a[1]) * t,
        a[2] + (b[2] - a[2]) * t,
    )


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

    root = (cx, cy, cz)
    spine = (cx, cy + height * 0.575, cz)
    chest = (cx, cy + height * 0.68, cz)
    neck = (cx, cy + height * 0.85, cz)
    head = (cx, cy + height * 0.92, cz)
    arm_l = (cx - shoulder_x, cy + height * 0.75, cz)
    hand_l = (cx - shoulder_x, cy + height * 0.41, cz + hand_z)
    arm_r = (cx + shoulder_x, cy + height * 0.75, cz)
    hand_r = (cx + shoulder_x, cy + height * 0.41, cz + hand_z)
    leg_l = (cx - leg_x, cy + height * 0.45, cz)
    leg_r = (cx + leg_x, cy + height * 0.45, cz)
    foot_l = (cx - leg_x, cy + height * 0.05, cz)
    foot_r = (cx + leg_x, cy + height * 0.05, cz)

    return [
        root,
        spine,
        chest,
        neck,
        head,
        arm_l,
        _lerp3(arm_l, hand_l, 0.5),
        hand_l,
        arm_r,
        _lerp3(arm_r, hand_r, 0.5),
        hand_r,
        leg_l,
        _lerp3(leg_l, foot_l, 0.5),
        foot_l,
        leg_r,
        _lerp3(leg_r, foot_r, 0.5),
        foot_r,
    ]


def _joint_nodes(joint_globals: list[tuple[float, float, float]], offset: int) -> list[dict]:
    (
        root,
        spine,
        chest,
        neck,
        head,
        arm_l,
        elbow_l,
        hand_l,
        arm_r,
        elbow_r,
        hand_r,
        leg_l,
        knee_l,
        foot_l,
        leg_r,
        knee_r,
        foot_r,
    ) = joint_globals

    return [
        {
            "name": "root",
            "translation": list(root),
            "children": [offset + 1, offset + 11, offset + 14],
        },
        {
            "name": "spine",
            "translation": _delta(root, spine),
            "children": [offset + 2],
        },
        {
            "name": "chest",
            "translation": _delta(spine, chest),
            "children": [offset + 3, offset + 5, offset + 8],
        },
        {"name": "neck", "translation": _delta(chest, neck), "children": [offset + 4]},
        {"name": "head", "translation": _delta(neck, head)},
        {
            "name": "arm_l",
            "translation": _delta(chest, arm_l),
            "children": [offset + 6],
        },
        {"name": "elbow_l", "translation": _delta(arm_l, elbow_l), "children": [offset + 7]},
        {"name": "hand_l", "translation": _delta(elbow_l, hand_l)},
        {
            "name": "arm_r",
            "translation": _delta(chest, arm_r),
            "children": [offset + 9],
        },
        {"name": "elbow_r", "translation": _delta(arm_r, elbow_r), "children": [offset + 10]},
        {"name": "hand_r", "translation": _delta(elbow_r, hand_r)},
        {
            "name": "leg_l",
            "translation": _delta(root, leg_l),
            "children": [offset + 12],
        },
        {"name": "knee_l", "translation": _delta(leg_l, knee_l), "children": [offset + 13]},
        {"name": "foot_l", "translation": _delta(knee_l, foot_l)},
        {
            "name": "leg_r",
            "translation": _delta(root, leg_r),
            "children": [offset + 15],
        },
        {"name": "knee_r", "translation": _delta(leg_r, knee_r), "children": [offset + 16]},
        {"name": "foot_r", "translation": _delta(knee_r, foot_r)},
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
    right = side > 0.0

    # Mastjie A-pose meshes often retain horizontal T-pose shoulder caps around yn~0.5–0.65.
    # Bind those to chest so shared character_anims arm rotations do not stretch them.
    if 0.52 <= yn < 0.68 and abs(side) >= arm_threshold * 0.85:
        return 2

    if yn >= 0.90:
        return 4

    if yn >= 0.84:
        return 3

    if 0.68 <= yn < 0.84 and abs(side) < arm_threshold * 0.75:
        return 2

    if 0.68 <= yn <= 0.88 and abs(side) >= arm_threshold:
        return 8 if right else 5

    if 0.52 <= yn < 0.68 and abs(side) >= arm_threshold * 0.85:
        return 9 if right else 6

    if 0.35 <= yn < 0.52 and abs(side) >= arm_threshold * 0.85:
        return 10 if right else 7

    if yn < 0.12:
        return 16 if right else 13

    if yn < 0.35:
        return 15 if right else 12

    if yn < 0.52:
        return 14 if right else 11

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


def rig_glb_bytes(data: bytes, *, hero_id: str = "", target_height: float | None = None) -> bytes:
    parsed = parse_glb(data)
    gltf = parsed.gltf
    if gltf.get("skins"):
        raise ValueError("source GLB is already skinned")
    gltf.pop("animations", None)
    bin_buf = bytearray(parsed.bin_blob)
    # Ranger vertex rest-pose was for the legacy Sketchfab green-hood T-pose source.
    # Tier-3 mastjie meshes share character_anims bone rotations; pre-bending vertices
    # fights those clips and stretches arms at runtime.

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
    target_height = target_height if target_height is not None else HERO_TARGET_HEIGHTS.get(hero_id)
    if target_height is not None:
        _normalize_mesh_height(gltf, bin_buf, positions_by_accessor, target_height)
    mins, maxs = _bounds(positions_by_accessor)
    _cull_apose_tpose_cap_triangles(gltf, bin_buf, positions_by_accessor, primitives, mins, maxs)
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
    source_bytes = source.read_bytes()
    if hero_id in CANONICAL_RIG_IDS:
        from tools.assets.rig_canonical_hero import rig_glb_canonical_bytes

        target.write_bytes(rig_glb_canonical_bytes(source_bytes, hero_id=hero_id))
        return

    target.write_bytes(rig_glb_bytes(source_bytes, hero_id=hero_id))


def main() -> int:
    for class_id, (source_rel, target_rel) in HEROES.items():
        source = ROOT / source_rel
        target = ROOT / target_rel
        rig_file(source, target)
        print(f"rigged {class_id}: {source_rel} -> {target_rel}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
