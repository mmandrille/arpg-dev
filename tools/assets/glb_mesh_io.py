"""Minimal single-buffer GLB read/write helpers for mesh-editing asset tools.

Moved verbatim from the retired ``rig_hero_glbs.py`` (ADR-0018 P3c, v484) because the Poly Pizza
equipment importer (``import_equipment_glb.py``) still rescales positions with them. Serialization
is unchanged (sorted-key compact JSON), so re-imported equipment GLBs stay byte-identical.
"""
from __future__ import annotations

import json
import struct
from dataclasses import dataclass

JSON_CHUNK = 0x4E4F534A
BIN_CHUNK = 0x004E4942
FLOAT = 5126


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
        if chunk_type == JSON_CHUNK:
            gltf = json.loads(chunk.decode("utf-8"))
        elif chunk_type == BIN_CHUNK:
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
    json_chunk = struct.pack("<II", len(json_bytes), JSON_CHUNK) + bytes(json_bytes)
    bin_chunk = struct.pack("<II", len(bin_bytes), BIN_CHUNK) + bytes(bin_bytes)
    total = 12 + len(json_chunk) + len(bin_chunk)
    return b"glTF" + struct.pack("<II", 2, total) + json_chunk + bin_chunk


def _vec3_layout(gltf: dict, accessor_index: int) -> tuple[dict, int, int]:
    accessor = gltf["accessors"][accessor_index]
    if accessor.get("componentType") != FLOAT or accessor.get("type") != "VEC3":
        raise ValueError(f"accessor {accessor_index} must be float VEC3")
    view = gltf["bufferViews"][accessor["bufferView"]]
    stride = int(view.get("byteStride", 12))
    start = int(view.get("byteOffset", 0)) + int(accessor.get("byteOffset", 0))
    return accessor, stride, start


def read_position_accessor(gltf: dict, bin_blob: bytes, accessor_index: int) -> list[tuple[float, float, float]]:
    accessor, stride, start = _vec3_layout(gltf, accessor_index)
    return [struct.unpack_from("<fff", bin_blob, start + i * stride) for i in range(int(accessor["count"]))]


def write_vec3_accessor(gltf: dict, bin_buf: bytearray, accessor_index: int, values: list[tuple[float, float, float]]) -> None:
    accessor, stride, start = _vec3_layout(gltf, accessor_index)
    if int(accessor["count"]) != len(values):
        raise ValueError(f"accessor {accessor_index} count mismatch")
    for i, value in enumerate(values):
        struct.pack_into("<fff", bin_buf, start + i * stride, *value)
    accessor["min"] = [min(v[i] for v in values) for i in range(3)]
    accessor["max"] = [max(v[i] for v in values) for i in range(3)]


def bounds(positions_by_accessor: dict[int, list[tuple[float, float, float]]]) -> tuple[list[float], list[float]]:
    positions = [p for values in positions_by_accessor.values() for p in values]
    if not positions:
        raise ValueError("GLB has no POSITION data")
    mins = [min(p[i] for p in positions) for i in range(3)]
    maxs = [max(p[i] for p in positions) for i in range(3)]
    return mins, maxs
