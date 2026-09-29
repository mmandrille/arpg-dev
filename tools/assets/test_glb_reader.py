"""Direct unit tests for the stdlib glTF structural reader (ADR-0018 P0)."""
from __future__ import annotations

import base64
import json
import struct
import zlib
from pathlib import Path

import pytest

from tools.assets import gen_glb, glb_reader


def _png_bytes(width: int, height: int) -> bytes:
    def chunk(kind: bytes, payload: bytes) -> bytes:
        crc = zlib.crc32(kind + payload) & 0xFFFFFFFF
        return struct.pack(">I", len(payload)) + kind + payload + struct.pack(">I", crc)

    header = struct.pack(">IIBBBBB", width, height, 8, 2, 0, 0, 0)
    raw = b"".join(b"\x00" + b"\x00\x00\x00" * width for _ in range(height))
    return b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", header) + chunk(b"IDAT", zlib.compress(raw)) + chunk(b"IEND", b"")


def write_gltf_fixture(directory: Path) -> Path:
    """Hand-built .gltf + external .bin: parent(T=+1x, S=2) -> child mesh, 1 triangle, 1 clip."""
    (directory / "fixture.bin").write_bytes(b"\x00" * 64)
    png_uri = "data:image/png;base64," + base64.b64encode(_png_bytes(3, 2)).decode("ascii")
    doc = {
        "asset": {"version": "2.0"},
        "scene": 0,
        "scenes": [{"nodes": [0]}],
        "nodes": [
            {"name": "Parent", "translation": [1.0, 0.0, 0.0], "scale": [2.0, 2.0, 2.0], "children": [1]},
            {"name": "Child", "mesh": 0},
        ],
        "meshes": [{"name": "Tri", "primitives": [{"attributes": {"POSITION": 0}, "indices": 1, "material": 0}]}],
        "materials": [{"name": "Mat"}],
        "images": [{"uri": png_uri}],
        "buffers": [{"uri": "fixture.bin", "byteLength": 64}],
        "accessors": [
            {"componentType": 5126, "count": 3, "type": "VEC3", "min": [0.0, 0.0, 0.0], "max": [1.0, 1.0, 0.0]},
            {"componentType": 5123, "count": 3, "type": "SCALAR"},
            {"componentType": 5126, "count": 2, "type": "SCALAR", "min": [0.0], "max": [1.25]},
        ],
        "animations": [{
            "name": "Wave",
            "samplers": [{"input": 2, "output": 0}],
            "channels": [{"sampler": 0, "target": {"node": 1, "path": "rotation"}}],
        }],
    }
    path = directory / "fixture.gltf"
    path.write_text(json.dumps(doc), encoding="utf-8")
    return path


def test_glb_skinned_fixture_reports_joints_and_triangles(tmp_path: Path) -> None:
    path = tmp_path / "dummy.glb"
    path.write_bytes(gen_glb.monster_dummy_glb())
    gltf = glb_reader.load_gltf(path)
    assert gltf.buffers and gltf.buffers[0] is not None
    assert glb_reader.skin_joint_names(gltf, 0) == ["root", "pivot"]
    cube_triangles = len(gen_glb._cube_geometry()[2]) // 3
    total = sum(glb_reader.mesh_triangle_count(gltf, i) for i in range(len(gltf.list("meshes"))))
    assert total == 2 * cube_triangles  # base slab + post


def test_glb_skinned_extents_use_bind_pose(tmp_path: Path) -> None:
    path = tmp_path / "dummy.glb"
    path.write_bytes(gen_glb.monster_dummy_glb())
    extents = glb_reader.scene_extents(glb_reader.load_gltf(path))
    assert extents is not None
    # Base slab is 0.9 wide; post top sits at 0.95 + 1.5/2 = 1.7 above the origin.
    assert extents["size"][0] == pytest.approx(0.9, abs=1e-4)
    assert extents["max"][1] == pytest.approx(1.7, abs=1e-4)


def test_gltf_external_buffer_transforms_images_and_animation(tmp_path: Path) -> None:
    gltf = glb_reader.load_gltf(write_gltf_fixture(tmp_path))
    assert gltf.buffers == [b"\x00" * 64]
    assert glb_reader.mesh_triangle_count(gltf, 0) == 1
    extents = glb_reader.scene_extents(gltf)
    assert extents is not None
    assert extents["min"] == pytest.approx([1.0, 0.0, 0.0])
    assert extents["max"] == pytest.approx([3.0, 2.0, 0.0])
    assert glb_reader.image_summaries(gltf) == [{"name": gltf.list("images")[0]["uri"], "mime": "", "width": 3, "height": 2}]
    [anim] = glb_reader.animation_summaries(gltf)
    assert anim["name"] == "Wave"
    assert anim["duration_s"] == pytest.approx(1.25)
    assert anim["target_nodes"] == ["Child"]
    assert anim["target_paths"] == ["rotation"]
    tree = glb_reader.node_tree(gltf)
    assert tree[0]["name"] == "Parent" and tree[0]["children"][0]["mesh"] == 0


def test_non_unit_scales_and_malformed_input(tmp_path: Path) -> None:
    gltf = glb_reader.load_gltf(write_gltf_fixture(tmp_path))
    assert glb_reader.non_unit_node_scales(gltf) == [("Parent", [2.0, 2.0, 2.0])]
    bad = tmp_path / "bad.glb"
    bad.write_bytes(b"not a model")
    with pytest.raises(ValueError):
        glb_reader.load_gltf(bad)


def test_non_triangle_primitive_is_not_counted(tmp_path: Path) -> None:
    gltf = glb_reader.load_gltf(write_gltf_fixture(tmp_path))
    assert glb_reader.primitive_triangle_count(gltf, {"mode": 1, "attributes": {"POSITION": 0}}) is None
