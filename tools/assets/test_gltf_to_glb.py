"""Tests for the deterministic .gltf -> .glb packer."""
from __future__ import annotations

import json
from pathlib import Path

import pytest

from tools.assets import glb_reader, gltf_to_glb
from tools.assets.test_glb_reader import _png_bytes, write_gltf_fixture


def _external_image_fixture(directory: Path) -> Path:
    path = write_gltf_fixture(directory)
    doc = json.loads(path.read_text(encoding="utf-8"))
    (directory / "atlas.png").write_bytes(_png_bytes(5, 4))
    doc["images"] = [{"uri": "atlas.png", "name": "atlas"}]
    path.write_text(json.dumps(doc), encoding="utf-8")
    return path


def test_pack_embeds_buffer_and_image_and_round_trips(tmp_path: Path) -> None:
    source = _external_image_fixture(tmp_path)
    out = tmp_path / "packed.glb"
    out.write_bytes(gltf_to_glb.pack(source))
    packed = glb_reader.load_gltf(out)
    original = glb_reader.load_gltf(source)
    assert "uri" not in packed.doc["buffers"][0]
    assert all("uri" not in image for image in packed.doc["images"])
    assert glb_reader.image_summaries(packed)[0]["width"] == 5
    assert glb_reader.mesh_triangle_count(packed, 0) == glb_reader.mesh_triangle_count(original, 0)
    assert glb_reader.scene_extents(packed) == glb_reader.scene_extents(original)
    assert [n["name"] for n in packed.list("nodes")] == [n["name"] for n in original.list("nodes")]


def test_pack_is_deterministic(tmp_path: Path) -> None:
    source = _external_image_fixture(tmp_path)
    assert gltf_to_glb.pack(source) == gltf_to_glb.pack(source)


def test_pack_rejects_missing_image(tmp_path: Path) -> None:
    source = _external_image_fixture(tmp_path)
    (tmp_path / "atlas.png").unlink()
    with pytest.raises(ValueError):
        gltf_to_glb.pack(source)
