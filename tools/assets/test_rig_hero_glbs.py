from __future__ import annotations

import json
import struct

from tools.assets.rig_hero_glbs import HEROES, REQUIRED_BONES, ROOT, parse_glb, read_position_accessor, rig_glb_bytes
from tools.assets.validate_assets import parse_glb_skin_joint_names


def _minimal_static_glb(positions: list[tuple[float, float, float]] | None = None) -> bytes:
    positions = positions or [
        (-0.5, 0.0, 0.0),
        (0.5, 0.0, 0.0),
        (0.0, 1.0, 0.0),
    ]
    normals = [(0.0, 0.0, 1.0)] * len(positions)
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
                "count": len(positions),
                "type": "VEC3",
                "min": [min(p[i] for p in positions) for i in range(3)],
                "max": [max(p[i] for p in positions) for i in range(3)],
            },
            {"bufferView": 1, "componentType": 5126, "count": len(positions), "type": "VEC3"},
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


def test_ranger_runtime_glb_skips_vertex_rest_pose_bake():
    """Tier-3 ranger meshes must not pre-bend vertices; shared anims rotate bones instead."""
    out = rig_glb_bytes(
        _minimal_static_glb([
            (-10.0, 7.5, 0.0),
            (10.0, 7.5, 0.0),
            (0.0, 0.0, 0.0),
            (0.0, 10.0, 0.0),
        ]),
        hero_id="ranger",
    )
    parsed = parse_glb(out)
    accessor = parsed.gltf["meshes"][0]["primitives"][0]["attributes"]["POSITION"]
    positions = read_position_accessor(parsed.gltf, parsed.bin_blob, accessor)
    left, right = positions[0], positions[1]
    assert abs(left[1] - right[1]) < 1e-5, "arms must stay level (no vertex rest-pose fold)"
    assert abs(left[0] + right[0]) < 1e-5, "arms must stay symmetric after rig scale"


def test_hero_rig_sources_include_all_class_models():
    assert set(HEROES.keys()) == {"barbarian", "paladin", "rogue", "ranger", "sorcerer"}
    assert HEROES["ranger"] == (
        "assets/characters/ranger/green_hood.glb",
        "client/assets/characters/ranger/ranger.glb",
    )


def test_paladin_runtime_glb_has_required_bones():
    runtime = ROOT / HEROES["paladin"][1]
    assert runtime.is_file(), f"missing rigged paladin runtime GLB: {runtime}"
    assert parse_glb_skin_joint_names(runtime) == set(REQUIRED_BONES)


def test_barbarian_runtime_glb_has_required_bones():
    runtime = ROOT / HEROES["barbarian"][1]
    assert runtime.is_file(), f"missing rigged barbarian runtime GLB: {runtime}"
    assert parse_glb_skin_joint_names(runtime) == set(REQUIRED_BONES)


def test_rogue_runtime_glb_has_required_bones():
    runtime = ROOT / HEROES["rogue"][1]
    assert runtime.is_file(), f"missing rigged rogue runtime GLB: {runtime}"
    assert parse_glb_skin_joint_names(runtime) == set(REQUIRED_BONES)


def test_sorcerer_runtime_glb_has_required_bones():
    runtime = ROOT / HEROES["sorcerer"][1]
    assert runtime.is_file(), f"missing rigged sorcerer runtime GLB: {runtime}"
    assert parse_glb_skin_joint_names(runtime) == set(REQUIRED_BONES)


def test_ranger_runtime_glb_has_required_bones():
    runtime = ROOT / HEROES["ranger"][1]
    assert runtime.is_file(), f"missing rigged ranger runtime GLB: {runtime}"
    assert parse_glb_skin_joint_names(runtime) == set(REQUIRED_BONES)
