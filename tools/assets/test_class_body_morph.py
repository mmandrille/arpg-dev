from __future__ import annotations

import math
import statistics
from pathlib import Path

import pytest

from tools.assets.class_body_morph import CLASS_MORPHS, PLAYER_CLASSES, fork_class_mesh, morph_mesh_bytes
from tools.assets.canonical_skeleton import joint_globals_from_mesh
from tools.assets.rig_hero_glbs import ROOT, _bounds, _normalize_mesh_height, parse_glb, read_position_accessor


def test_class_morphs_cover_all_player_classes():
    assert set(CLASS_MORPHS) == set(PLAYER_CLASSES)


def test_morph_rejects_skinned_source():
    from tools.assets.test_rig_hero_glbs import _minimal_static_glb
    from tools.assets.rig_hero_glbs import rig_glb_bytes

    skinned = rig_glb_bytes(_minimal_static_glb())
    with pytest.raises(ValueError, match="already skinned"):
        morph_mesh_bytes(skinned, "barbarian")


def test_forked_meshes_are_static_and_distinct(tmp_path: Path, monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setattr("tools.assets.class_body_morph.ROOT", tmp_path)
    base = tmp_path / "assets/characters/base_human"
    base.mkdir(parents=True)
    source = ROOT / "assets/characters/base_human/base_human_mesh.glb"
    (base / "base_human_mesh.glb").write_bytes(source.read_bytes())

    outputs = [fork_class_mesh(class_id, root=tmp_path) for class_id in PLAYER_CLASSES]
    hashes = {path.name: path.read_bytes() for path in outputs}

    assert len(set(hashes.values())) == len(PLAYER_CLASSES)
    for path in outputs:
        parsed = parse_glb(path.read_bytes())
        assert not parsed.gltf.get("skins")


@pytest.mark.parametrize("class_id", PLAYER_CLASSES)
def test_class_landmark_hands_near_mesh_extents(class_id: str):
    source = ROOT / f"assets/characters/{class_id}/{class_id}_mesh.glb"
    if not source.is_file():
        fork_class_mesh(class_id)

    from tools.assets.rig_canonical_hero import rig_glb_canonical_bytes
    from tools.assets.rig_hero_glbs import HERO_TARGET_HEIGHTS

    mesh_data = source.read_bytes()
    out = rig_glb_canonical_bytes(mesh_data, hero_id=class_id, target_height=HERO_TARGET_HEIGHTS[class_id])
    parsed = parse_glb(out)
    bin_buf = bytearray(parsed.bin_blob)
    pos_by: dict[int, list[tuple[float, float, float]]] = {}
    for mesh in parsed.gltf.get("meshes", []):
        for prim in mesh.get("primitives", []):
            ai = int(prim["attributes"]["POSITION"])
            pos_by.setdefault(ai, read_position_accessor(parsed.gltf, bytes(bin_buf), ai))

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
    assert dist < 0.12, f"{class_id} hand_r joint too far from mesh hand cluster: {dist:.3f}"
