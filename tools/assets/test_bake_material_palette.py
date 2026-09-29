"""Tests for the colour-material -> palette-texture GLB bake (ADR-0018 P4b)."""
from __future__ import annotations

import struct
import zlib
from pathlib import Path

import pytest

from tools.assets import bake_material_palette as bake
from tools.assets import glb_reader
from tools.assets.gltf_to_glb import write_glb

RED = [1.0, 0.0, 0.0, 1.0]
GREY = [0.2, 0.2, 0.2, 1.0]


def _material(name: str, factor: list[float]) -> dict:
    return {"name": name, "pbrMetallicRoughness": {"baseColorFactor": factor, "metallicFactor": 0.4}}


def _colour_glb(path: Path, factors: list[list[float]], primitive_materials: list[int]) -> Path:
    """One triangle per primitive; primitive i uses material primitive_materials[i]."""
    binary = bytearray(struct.pack("<9f", 0, 0, 0, 1, 0, 0, 0, 1, 0))
    doc = {
        "asset": {"version": "2.0"},
        "bufferViews": [{"buffer": 0, "byteOffset": 0, "byteLength": 36}],
        "accessors": [{"bufferView": 0, "componentType": 5126, "count": 3, "type": "VEC3",
                       "min": [0, 0, 0], "max": [1, 1, 0]}],
        "materials": [_material(f"m{i}", f) for i, f in enumerate(factors)],
        "meshes": [{"name": "body", "primitives": [
            {"attributes": {"POSITION": 0}, "material": m} for m in primitive_materials
        ]}],
        "nodes": [{"name": "body", "mesh": 0}],
        "scenes": [{"nodes": [0]}],
        "scene": 0,
    }
    path.write_bytes(write_glb(doc, binary))
    return path


def _decode_png_rgb(data: bytes) -> tuple[int, int, list[bytes]]:
    width, height = struct.unpack(">II", data[16:24])
    offset, idat = 8, b""
    while offset < len(data):
        length = struct.unpack(">I", data[offset:offset + 4])[0]
        tag = data[offset + 4:offset + 8]
        if tag == b"IDAT":
            idat += data[offset + 8:offset + 8 + length]
        offset += 12 + length
    raw = zlib.decompress(idat)
    stride = 1 + width * 3
    return width, height, [raw[r * stride + 1:(r + 1) * stride] for r in range(height)]


def _buffer(glb: glb_reader.Gltf) -> bytes:
    data = glb.buffers[0]
    assert data is not None
    return data


def _texel(glb: glb_reader.Gltf, u: float) -> tuple[int, ...]:
    view = glb.doc["bufferViews"][glb.doc["images"][0]["bufferView"]]
    png = _buffer(glb)[view["byteOffset"]:view["byteOffset"] + view["byteLength"]]
    width, height, rows = _decode_png_rgb(png)
    x = int(u * width)
    return tuple(rows[height // 2][x * 3:x * 3 + 3])


def _uv(glb: glb_reader.Gltf, primitive: dict) -> tuple[float, float]:
    accessor = glb.doc["accessors"][primitive["attributes"]["TEXCOORD_0"]]
    view = glb.doc["bufferViews"][accessor["bufferView"]]
    return struct.unpack_from("<ff", _buffer(glb), view["byteOffset"])


def test_bake_collapses_to_one_textured_kit_material(tmp_path: Path) -> None:
    source = _colour_glb(tmp_path / "in.glb", [RED, GREY], [0, 1])
    out = tmp_path / "out.glb"
    out.write_bytes(bake.bake(source))
    glb = glb_reader.load_gltf(out)
    assert len(glb.doc["materials"]) == 1
    pbr = glb.doc["materials"][0]["pbrMetallicRoughness"]
    assert pbr["baseColorTexture"] == {"index": 0}
    assert pbr["metallicFactor"] == bake.KIT_METALLIC
    assert all(p["material"] == 0 for p in glb.doc["meshes"][0]["primitives"])


def test_each_primitive_samples_its_own_srgb_colour(tmp_path: Path) -> None:
    source = _colour_glb(tmp_path / "in.glb", [RED, GREY], [0, 1])
    out = tmp_path / "out.glb"
    out.write_bytes(bake.bake(source))
    glb = glb_reader.load_gltf(out)
    red, grey = glb.doc["meshes"][0]["primitives"]
    assert glb.doc["accessors"][red["attributes"]["TEXCOORD_0"]]["count"] == 3
    assert _texel(glb, _uv(glb, red)[0]) == (255, 0, 0)
    expected_grey = bake.srgb_byte(0.2)
    assert _texel(glb, _uv(glb, grey)[0]) == (expected_grey,) * 3
    assert expected_grey > 51  # linear 0.2 must be sRGB-encoded, not copied


def test_identical_colours_share_a_palette_slot(tmp_path: Path) -> None:
    source = _colour_glb(tmp_path / "in.glb", [RED, GREY, RED], [0, 2])
    glb_path = tmp_path / "out.glb"
    glb_path.write_bytes(bake.bake(source))
    glb = glb_reader.load_gltf(glb_path)
    first, second = glb.doc["meshes"][0]["primitives"]
    assert _uv(glb, first) == _uv(glb, second)
    assert glb_reader.image_summaries(glb)[0]["width"] == 2 * bake.BLOCK


def test_bake_is_deterministic_and_keeps_geometry(tmp_path: Path) -> None:
    source = _colour_glb(tmp_path / "in.glb", [RED, GREY], [0, 1])
    assert bake.bake(source) == bake.bake(source)
    out = tmp_path / "out.glb"
    out.write_bytes(bake.bake(source))
    original, baked = glb_reader.load_gltf(source), glb_reader.load_gltf(out)
    assert glb_reader.mesh_triangle_count(baked, 0) == glb_reader.mesh_triangle_count(original, 0)
    assert glb_reader.scene_extents(baked) == glb_reader.scene_extents(original)


def test_bake_rejects_textured_input(tmp_path: Path) -> None:
    source = tmp_path / "in.glb"
    first = _colour_glb(tmp_path / "first.glb", [RED], [0])
    source.write_bytes(bake.bake(first))
    with pytest.raises(ValueError):
        bake.bake(source)
