"""Pytest for the asset validator (ADR-0006 D5).

Each test builds a throwaway repo-shaped root under tmp_path so the validator's
disk checks (runtime_path existence, sha256, GLB node names) run for real.
"""
from __future__ import annotations

import hashlib
import json
import shutil
import struct
from pathlib import Path

from tools.assets.validate_assets import Report, validate

REPO_ROOT = Path(__file__).resolve().parents[2]
REAL_SCHEMA = REPO_ROOT / "assets/manifests/assets.v0.schema.json"

CHAR_GLB = "client/assets/characters/base_humanoid/base_humanoid.glb"
STATIC_CHAR_GLB = "client/assets/characters/static_hero/static_hero.glb"
SWORD_GLB = "client/assets/equipment/weapons/rusty_sword/rusty_sword.glb"
MONSTER_GLB = "client/assets/monsters/dummy/monster_dummy.glb"


def make_glb(node_names: list[str]) -> bytes:
    gltf = {"asset": {"version": "2.0"}, "nodes": [{"name": n} for n in node_names]}
    payload = json.dumps(gltf).encode("utf-8")
    payload += b" " * ((4 - len(payload) % 4) % 4)
    chunk = struct.pack("<II", len(payload), 0x4E4F534A) + payload  # 'JSON'
    total = 12 + len(chunk)
    return b"glTF" + struct.pack("<II", 2, total) + chunk


def make_skinned_glb(joint_names: list[str]) -> bytes:
    """Emit a minimal valid skinned glTF whose `skins[0].joints` are joint_names.

    Enough of a skin (non-empty `skins`, JOINTS_0/WEIGHTS_0 attributes, inverse
    bind matrices) that ``parse_glb_skin_joint_names`` returns exactly the given
    names — no Godot import needed for headless tests.
    """
    n = len(joint_names)
    # Joint nodes (indices 0..n-1) + one mesh node referencing the skin.
    nodes = [{"name": name, "translation": [0.0, 0.0, 0.0]} for name in joint_names]
    nodes.append({"name": "Mesh", "mesh": 0, "skin": 0})

    bin_buf = bytearray()
    # POSITION: one vertex.
    pos_off = len(bin_buf)
    bin_buf += struct.pack("<fff", 0.0, 0.0, 0.0)
    # JOINTS_0 (VEC4 u16): weighted to joint 0.
    j_off = len(bin_buf)
    bin_buf += struct.pack("<HHHH", 0, 0, 0, 0)
    # WEIGHTS_0 (VEC4 f32).
    w_off = len(bin_buf)
    bin_buf += struct.pack("<ffff", 1.0, 0.0, 0.0, 0.0)
    # Inverse bind matrices: identity per joint.
    ibm_off = len(bin_buf)
    identity = [1.0, 0.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 0.0, 1.0]
    for _ in range(n):
        bin_buf += struct.pack("<16f", *identity)
    while len(bin_buf) % 4 != 0:
        bin_buf.append(0)

    gltf = {
        "asset": {"version": "2.0"},
        "scene": 0,
        "scenes": [{"nodes": [0, n]}],
        "nodes": nodes,
        "meshes": [{
            "primitives": [{
                "attributes": {"POSITION": 0, "JOINTS_0": 1, "WEIGHTS_0": 2},
                "mode": 4,
            }],
        }],
        "skins": [{"joints": list(range(n)), "inverseBindMatrices": 3}],
        "accessors": [
            {"bufferView": 0, "componentType": 5126, "count": 1, "type": "VEC3"},
            {"bufferView": 1, "componentType": 5123, "count": 1, "type": "VEC4"},
            {"bufferView": 2, "componentType": 5126, "count": 1, "type": "VEC4"},
            {"bufferView": 3, "componentType": 5126, "count": n, "type": "MAT4"},
        ],
        "bufferViews": [
            {"buffer": 0, "byteOffset": pos_off, "byteLength": j_off - pos_off},
            {"buffer": 0, "byteOffset": j_off, "byteLength": w_off - j_off},
            {"buffer": 0, "byteOffset": w_off, "byteLength": ibm_off - w_off},
            {"buffer": 0, "byteOffset": ibm_off, "byteLength": n * 64},
        ],
        "buffers": [{"byteLength": len(bin_buf)}],
    }
    json_bytes = bytearray(json.dumps(gltf).encode("utf-8"))
    while len(json_bytes) % 4 != 0:
        json_bytes.append(0x20)
    json_chunk = struct.pack("<II", len(json_bytes), 0x4E4F534A) + bytes(json_bytes)
    bin_chunk = struct.pack("<II", len(bin_buf), 0x004E4942) + bytes(bin_buf)
    total = 12 + len(json_chunk) + len(bin_chunk)
    return b"glTF" + struct.pack("<II", 2, total) + json_chunk + bin_chunk


def write(path: Path, data) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    if isinstance(data, bytes):
        path.write_bytes(data)
    else:
        path.write_text(json.dumps(data), encoding="utf-8")


def default_manifest() -> dict:
    return {
        "version": 0,
        "assets": {
            "character_base_humanoid_v0": {
                "type": "character",
                "runtime_path": CHAR_GLB,
                "format": "glb",
                "required_nodes": ["root", "spine", "arm_l", "hand_l", "arm_r", "hand_r", "leg_l", "leg_r"],
            },
            "weapon_rusty_sword_v0": {
                "type": "equipment",
                "slot": "main_hand",
                "runtime_path": SWORD_GLB,
                "format": "glb",
                "required_nodes": [],
            },
            "monster_dummy_v0": {
                "type": "monster",
                "runtime_path": MONSTER_GLB,
                "format": "glb",
                "required_nodes": ["root", "pivot"],
            },
        },
    }


def default_visuals() -> dict:
    return {
        "version": 0,
        "item_visuals": {
            "rusty_sword": {
                "asset_id": "weapon_rusty_sword_v0",
                "slot": "main_hand",
                "mount_socket": "right_hand_socket",
                "local_transform": {
                    "position": {"x": 0.0, "y": 0.0, "z": 0.0},
                    "rotation_degrees": {"x": 0.0, "y": 0.0, "z": 0.0},
                    "scale": {"x": 1.0, "y": 1.0, "z": 1.0},
                },
            }
        },
    }


def default_monster_visuals() -> dict:
    return {
        "version": 0,
        "monster_visuals": {
            "training_dummy": {
                "asset_id": "monster_dummy_v0",
                "scene": "monster_dummy",
                "scale": 1.0,
                "height_offset": 0.0,
                "animation_profile": "ground_biped",
            }
        },
    }


def build_root(
    tmp_path: Path,
    *,
    manifest: dict | None = None,
    visuals: dict | None = None,
    monster_visuals: dict | None = None,
    char_nodes: list[str] | None = None,
    write_sword: bool = True,
) -> Path:
    root = tmp_path / "repo"
    (root / "assets/manifests").mkdir(parents=True, exist_ok=True)
    shutil.copy(REAL_SCHEMA, root / "assets/manifests/assets.v0.schema.json")
    write(root / "assets/manifests/assets.v0.json", manifest or default_manifest())
    write(root / "shared/assets/item_visuals.v0.json", visuals or default_visuals())
    write(root / "shared/assets/monster_visuals.v0.json", monster_visuals or default_monster_visuals())
    char_joints = char_nodes if char_nodes is not None else [
        "root", "spine", "arm_l", "hand_l", "arm_r", "hand_r", "leg_l", "leg_r"
    ]
    write(root / CHAR_GLB, make_skinned_glb(char_joints))
    if write_sword:
        write(root / SWORD_GLB, make_glb([]))
    write(root / MONSTER_GLB, make_skinned_glb(["root", "pivot"]))
    return root


def run(root: Path) -> Report:
    report = Report()
    validate(root, report)
    return report


def test_happy_path(tmp_path):
    report = run(build_root(tmp_path))
    assert report.failures == []


def test_missing_runtime_file(tmp_path):
    report = run(build_root(tmp_path, write_sword=False))
    assert any("runtime_path" in f for f in report.failures)


def test_unknown_asset_id(tmp_path):
    visuals = default_visuals()
    visuals["item_visuals"]["rusty_sword"]["asset_id"] = "nope_v0"
    report = run(build_root(tmp_path, visuals=visuals))
    assert any("asset_id resolution" in f for f in report.failures)


def test_socket_coverage_failure(tmp_path):
    manifest = default_manifest()
    # Drop the weapon mount bone: the mount-bone coverage check must fail.
    manifest["assets"]["character_base_humanoid_v0"]["required_nodes"] = [
        "root", "spine", "arm_r", "hand_r", "leg_l", "leg_r"
    ]
    report = run(build_root(
        tmp_path, manifest=manifest,
        char_nodes=["root", "spine", "arm_r", "hand_r", "leg_l", "leg_r"],
    ))
    assert any("mount bone" in f for f in report.failures)


def test_static_character_asset_allowed(tmp_path):
    manifest = default_manifest()
    manifest["assets"]["character_static_hero_v0"] = {
        "type": "character",
        "runtime_path": STATIC_CHAR_GLB,
        "format": "glb",
        "required_nodes": [],
    }
    root = build_root(tmp_path, manifest=manifest)
    write(root / STATIC_CHAR_GLB, make_glb(["world", "Static Hero"]))
    report = run(root)
    assert report.failures == []


def test_sha256_mismatch(tmp_path):
    manifest = default_manifest()
    manifest["assets"]["weapon_rusty_sword_v0"]["provenance"] = {
        "license": "CC0",
        "sha256": "0" * 64,
    }
    report = run(build_root(tmp_path, manifest=manifest))
    assert any("provenance.sha256" in f for f in report.failures)


def test_sha256_match(tmp_path):
    # Compute the real digest of the sword GLB and assert it passes.
    sword_bytes = make_glb([])
    digest = hashlib.sha256(sword_bytes).hexdigest()
    manifest = default_manifest()
    manifest["assets"]["weapon_rusty_sword_v0"]["provenance"] = {
        "license": "CC0",
        "sha256": digest,
    }
    report = run(build_root(tmp_path, manifest=manifest))
    assert report.failures == []


def test_glb_required_node_not_a_skin_joint(tmp_path):
    # Manifest declares the rig joints, but the GLB skin omits one of them
    # (here `spine`) -> [6] must hard-fail because it is not an actual joint.
    report = run(build_root(
        tmp_path,
        char_nodes=["root", "arm_l", "hand_l", "arm_r", "hand_r", "leg_l", "leg_r"],
    ))
    assert any("glb joint" in f or "not skin joints" in f for f in report.failures)


def test_asset_id_wrong_type(tmp_path):
    # Point the weapon visual at the character asset -> type mismatch.
    visuals = default_visuals()
    visuals["item_visuals"]["rusty_sword"]["asset_id"] = "character_base_humanoid_v0"
    report = run(build_root(tmp_path, visuals=visuals))
    assert any("asset_id type" in f for f in report.failures)


def test_schema_invalid_manifest(tmp_path):
    manifest = default_manifest()
    manifest["version"] = 99  # violates const 0
    report = run(build_root(tmp_path, manifest=manifest))
    assert any("manifest schema" in f for f in report.failures)


def test_monster_entry_passes(tmp_path):
    manifest = default_manifest()
    manifest["assets"]["monster_dummy_v0"] = {
        "type": "monster",
        "runtime_path": MONSTER_GLB,
        "format": "glb",
        "required_nodes": ["root", "pivot"],
    }
    root = build_root(tmp_path, manifest=manifest)
    write(root / MONSTER_GLB, make_skinned_glb(["root", "pivot"]))
    report = run(root)
    assert report.failures == []


def test_monster_visual_unknown_asset_id(tmp_path):
    monster_visuals = default_monster_visuals()
    monster_visuals["monster_visuals"]["training_dummy"]["asset_id"] = "missing_monster_v0"
    report = run(build_root(tmp_path, monster_visuals=monster_visuals))
    assert any("monster asset_id resolution" in f for f in report.failures)


def test_monster_visual_wrong_asset_type(tmp_path):
    monster_visuals = default_monster_visuals()
    monster_visuals["monster_visuals"]["training_dummy"]["asset_id"] = "weapon_rusty_sword_v0"
    report = run(build_root(tmp_path, monster_visuals=monster_visuals))
    assert any("monster asset_id type" in f for f in report.failures)


def test_required_node_not_a_skin_joint_fails(tmp_path):
    manifest = default_manifest()
    manifest["assets"]["character_base_humanoid_v0"]["required_nodes"] = ["not_a_joint"]
    root = build_root(tmp_path, manifest=manifest, char_nodes=["root", "hand_r"])
    report = run(root)
    assert any("not skin joints" in f or "not_a_joint" in f for f in report.failures)


def test_orphan_client_asset_fails(tmp_path):
    root = build_root(tmp_path)
    write(root / "client/assets/monsters/bat.glb", make_glb(["root"]))
    report = run(root)
    assert any("orphan client asset" in f for f in report.failures)
