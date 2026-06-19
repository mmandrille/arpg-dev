from __future__ import annotations

import json
import struct

from tools.assets.rig_hero_glbs import REQUIRED_BONES, parse_glb, rig_glb_bytes
from tools.assets.validate_assets import parse_glb_skin_joint_names


def _minimal_static_glb() -> bytes:
    positions = [
        (-0.5, 0.0, 0.0),
        (0.5, 0.0, 0.0),
        (0.0, 1.0, 0.0),
    ]
    normals = [(0.0, 0.0, 1.0)] * 3
    indices = [0, 1, 2]
    bin_buf = bytearray()
    pos_off = len(bin_buf)
    for p in positions:
        bin_buf += struct.pack("<fff", *p)
    nrm_off = len(bin_buf)
    for n in normals:
        bin_buf += struct.pack("<fff", *n)
    idx_off = len(bin_buf)
    for i in indices:
        bin_buf += struct.pack("<H", i)
    while len(bin_buf) % 4:
        bin_buf.append(0)
    gltf = {
        "asset": {"version": "2.0"},
        "scene": 0,
        "scenes": [{"nodes": [0]}],
        "nodes": [{"name": "world", "children": [1]}, {"name": "Hero", "mesh": 0}],
        "meshes": [{
            "primitives": [{
                "attributes": {"POSITION": 0, "NORMAL": 1},
                "indices": 2,
                "mode": 4,
            }],
        }],
        "accessors": [
            {
                "bufferView": 0,
                "componentType": 5126,
                "count": 3,
                "type": "VEC3",
                "min": [-0.5, 0.0, 0.0],
                "max": [0.5, 1.0, 0.0],
            },
            {"bufferView": 1, "componentType": 5126, "count": 3, "type": "VEC3"},
            {"bufferView": 2, "componentType": 5123, "count": 3, "type": "SCALAR"},
        ],
        "bufferViews": [
            {"buffer": 0, "byteOffset": pos_off, "byteLength": nrm_off - pos_off, "target": 34962},
            {"buffer": 0, "byteOffset": nrm_off, "byteLength": idx_off - nrm_off, "target": 34962},
            {"buffer": 0, "byteOffset": idx_off, "byteLength": len(indices) * 2, "target": 34963},
        ],
        "buffers": [{"byteLength": len(bin_buf)}],
    }
    json_bytes = bytearray(json.dumps(gltf, sort_keys=True, separators=(",", ":")).encode("utf-8"))
    while len(json_bytes) % 4:
        json_bytes.append(0x20)
    json_chunk = struct.pack("<II", len(json_bytes), 0x4E4F534A) + bytes(json_bytes)
    bin_chunk = struct.pack("<II", len(bin_buf), 0x004E4942) + bytes(bin_buf)
    total = 12 + len(json_chunk) + len(bin_chunk)
    return b"glTF" + struct.pack("<II", 2, total) + json_chunk + bin_chunk


def test_rig_glb_bytes_adds_humanoid_skin_and_weights(tmp_path):
    out = rig_glb_bytes(_minimal_static_glb())
    parsed = parse_glb(out)
    assert [parsed.gltf["nodes"][i]["name"] for i in parsed.gltf["skins"][0]["joints"]] == REQUIRED_BONES
    root_node = parsed.gltf["nodes"][parsed.gltf["skins"][0]["joints"][0]]
    assert [parsed.gltf["nodes"][i]["name"] for i in root_node["children"]] == ["spine", "leg_l", "leg_r"]
    primitive = parsed.gltf["meshes"][0]["primitives"][0]
    assert "JOINTS_0" in primitive["attributes"]
    assert "WEIGHTS_0" in primitive["attributes"]
    assert parsed.gltf["nodes"][1]["skin"] == 0

    rigged_path = tmp_path / "rigged.glb"
    rigged_path.write_bytes(out)
    assert parse_glb_skin_joint_names(rigged_path) == set(REQUIRED_BONES)


def test_rig_glb_bytes_rejects_already_skinned_source():
    out = rig_glb_bytes(_minimal_static_glb())
    try:
        rig_glb_bytes(out)
    except ValueError as exc:
        assert "already skinned" in str(exc)
    else:
        raise AssertionError("expected already-skinned GLB to be rejected")
