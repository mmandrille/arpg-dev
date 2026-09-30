"""Tests for tools/assets/glb_mesh_io.py (helpers kept from the retired hero rig tool)."""
from __future__ import annotations

import struct

import pytest

from tools.assets import glb_mesh_io


def _triangle_glb() -> bytes:
    positions = [(0.0, 0.0, 0.0), (2.0, 0.0, 0.0), (0.0, 3.0, 1.0)]
    gltf = {
        "asset": {"version": "2.0"},
        "buffers": [{"byteLength": 36}],
        "bufferViews": [{"buffer": 0, "byteOffset": 0, "byteLength": 36}],
        "accessors": [{"bufferView": 0, "componentType": 5126, "count": 3, "type": "VEC3"}],
    }
    return glb_mesh_io.write_glb(gltf, b"".join(struct.pack("<fff", *p) for p in positions))


def test_round_trip_reads_positions_and_bounds() -> None:
    glb = glb_mesh_io.parse_glb(_triangle_glb())
    positions = glb_mesh_io.read_position_accessor(glb.gltf, glb.bin_blob, 0)
    assert positions[2] == (0.0, 3.0, 1.0)
    assert glb_mesh_io.bounds({0: positions}) == ([0.0, 0.0, 0.0], [2.0, 3.0, 1.0])


def test_write_vec3_accessor_updates_data_and_min_max() -> None:
    glb = glb_mesh_io.parse_glb(_triangle_glb())
    buf = bytearray(glb.bin_blob)
    scaled = [(x * 0.5, y * 0.5, z * 0.5) for x, y, z in glb_mesh_io.read_position_accessor(glb.gltf, buf, 0)]
    glb_mesh_io.write_vec3_accessor(glb.gltf, buf, 0, scaled)
    assert glb.gltf["accessors"][0]["max"] == [1.0, 1.5, 0.5]
    reparsed = glb_mesh_io.parse_glb(glb_mesh_io.write_glb(glb.gltf, bytes(buf)))
    assert glb_mesh_io.read_position_accessor(reparsed.gltf, reparsed.bin_blob, 0) == scaled


def test_write_is_deterministic_and_rejects_bad_input() -> None:
    assert _triangle_glb() == _triangle_glb()
    with pytest.raises(ValueError):
        glb_mesh_io.parse_glb(b"not a glb at all, too short?")
    glb = glb_mesh_io.parse_glb(_triangle_glb())
    with pytest.raises(ValueError):
        glb_mesh_io.write_vec3_accessor(glb.gltf, bytearray(glb.bin_blob), 0, [(0.0, 0.0, 0.0)])
