from __future__ import annotations

import math
import struct

from tools.assets.canonical_skeleton import joint_globals, joint_globals_from_mesh, vertex_weights
from tools.assets.rig_canonical_hero import rig_glb_canonical_bytes
from tools.assets.rig_hero_glbs import REQUIRED_BONES, ROOT, parse_glb
from tools.assets.test_rig_hero_glbs import _minimal_static_glb
from tools.assets.validate_assets import parse_glb_skin_joint_names


def test_joint_globals_match_gen_glb_bind_pose():
    g = joint_globals()
    assert len(g) == 17
    assert abs(g[0][2] - 0.1588) < 1e-4
    assert abs(g[4][1] - 1.812) < 0.01
    assert abs(g[10][0] - 0.223) < 0.01


def test_vertex_weights_are_normalized():
    g = joint_globals()
    _, w = vertex_weights(g[8], g)
    assert abs(sum(w) - 1.0) < 1e-6
    assert w[0] > 0.5


def test_vertex_weights_arm_segment_is_dual():
    g = joint_globals()
    mid = (
        (g[8][0] + g[9][0]) / 2.0,
        (g[8][1] + g[9][1]) / 2.0,
        (g[8][2] + g[9][2]) / 2.0,
    )
    j, w = vertex_weights(mid, g)
    assert w[0] > 0.01 and w[1] > 0.01
    assert j[0] in (8, 9) and j[1] in (8, 9)


def test_rig_glb_canonical_bytes_adds_skin(tmp_path):
    out = rig_glb_canonical_bytes(_minimal_static_glb(), hero_id="barbarian")
    parsed = parse_glb(out)
    assert [parsed.gltf["nodes"][i]["name"] for i in parsed.gltf["skins"][0]["joints"]] == REQUIRED_BONES
    primitive = parsed.gltf["meshes"][0]["primitives"][0]
    assert "JOINTS_0" in primitive["attributes"]
    assert "WEIGHTS_0" in primitive["attributes"]
    rigged_path = tmp_path / "rigged.glb"
    rigged_path.write_bytes(out)
    assert parse_glb_skin_joint_names(rigged_path) == set(REQUIRED_BONES)


def test_barbarian_landmark_hands_near_mesh_extents():
    from pathlib import Path
    import statistics

    from tools.assets.rig_hero_glbs import _bounds, _normalize_mesh_height, parse_glb, read_position_accessor

    data = Path("assets/characters/barbarian/goliath_barbarian.glb").read_bytes()
    parsed = parse_glb(data)
    bin_buf = bytearray(parsed.bin_blob)
    pos_by: dict[int, list[tuple[float, float, float]]] = {}
    for mesh in parsed.gltf.get("meshes", []):
        for prim in mesh.get("primitives", []):
            ai = int(prim["attributes"]["POSITION"])
            pos_by.setdefault(ai, read_position_accessor(parsed.gltf, bytes(bin_buf), ai))
    _normalize_mesh_height(parsed.gltf, bin_buf, pos_by, 1.97)
    mins, maxs = _bounds(pos_by)
    all_pos = [p for pts in pos_by.values() for p in pts]
    g = joint_globals_from_mesh(all_pos, mins, maxs)
    w = maxs[0] - mins[0]
    cx = (mins[0] + maxs[0]) * 0.5
    hand_pts = [p for p in all_pos if p[0] > cx + w * 0.32 and mins[1] + 0.38 * (maxs[1] - mins[1]) < p[1]]
    mesh_hand = (
        statistics.mean(p[0] for p in hand_pts),
        statistics.mean(p[1] for p in hand_pts),
        statistics.mean(p[2] for p in hand_pts),
    )
    dist = math.sqrt(sum((g[10][i] - mesh_hand[i]) ** 2 for i in range(3)))
    assert dist < 0.12, f"hand_r joint too far from mesh hand cluster: {dist:.3f}"


def test_barbarian_canonical_rig_has_dual_weights():
    source = ROOT / "assets/characters/barbarian/goliath_barbarian.glb"
    assert source.is_file()
    out = rig_glb_canonical_bytes(source.read_bytes(), hero_id="barbarian")
    parsed = parse_glb(out)
    bin_buf = parsed.bin_blob
    w_acc = parsed.gltf["accessors"][
        parsed.gltf["meshes"][0]["primitives"][0]["attributes"]["WEIGHTS_0"]
    ]
    w_view = parsed.gltf["bufferViews"][w_acc["bufferView"]]
    off = w_view["byteOffset"]
    dual = 0
    for i in range(int(w_acc["count"])):
        w0, w1, w2, w3 = struct.unpack_from("<ffff", bin_buf, off + i * 16)
        if w1 > 0.01:
            dual += 1
    assert dual > 500, f"expected many dual-weight verts on goliath, got {dual}"
