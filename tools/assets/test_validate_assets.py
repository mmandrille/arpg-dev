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
SWORD_GLB = "client/assets/equipment/weapons/rusty_sword/rusty_sword.glb"


def make_glb(node_names: list[str]) -> bytes:
    gltf = {"asset": {"version": "2.0"}, "nodes": [{"name": n} for n in node_names]}
    payload = json.dumps(gltf).encode("utf-8")
    payload += b" " * ((4 - len(payload) % 4) % 4)
    chunk = struct.pack("<II", len(payload), 0x4E4F534A) + payload  # 'JSON'
    total = 12 + len(chunk)
    return b"glTF" + struct.pack("<II", 2, total) + chunk


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
                "required_nodes": ["right_hand_socket"],
            },
            "weapon_rusty_sword_v0": {
                "type": "equipment",
                "slot": "weapon",
                "runtime_path": SWORD_GLB,
                "format": "glb",
                "required_nodes": [],
            },
        },
    }


def default_visuals() -> dict:
    return {
        "version": 0,
        "item_visuals": {
            "rusty_sword": {
                "asset_id": "weapon_rusty_sword_v0",
                "slot": "weapon",
                "mount_socket": "right_hand_socket",
                "local_transform": {
                    "position": {"x": 0.0, "y": 0.0, "z": 0.0},
                    "rotation_degrees": {"x": 0.0, "y": 0.0, "z": 0.0},
                    "scale": {"x": 1.0, "y": 1.0, "z": 1.0},
                },
            }
        },
    }


def build_root(
    tmp_path: Path,
    *,
    manifest: dict | None = None,
    visuals: dict | None = None,
    char_nodes: list[str] | None = None,
    write_sword: bool = True,
) -> Path:
    root = tmp_path / "repo"
    (root / "assets/manifests").mkdir(parents=True, exist_ok=True)
    shutil.copy(REAL_SCHEMA, root / "assets/manifests/assets.v0.schema.json")
    write(root / "assets/manifests/assets.v0.json", manifest or default_manifest())
    write(root / "shared/assets/item_visuals.v0.json", visuals or default_visuals())
    write(root / CHAR_GLB, make_glb(char_nodes if char_nodes is not None else ["right_hand_socket"]))
    if write_sword:
        write(root / SWORD_GLB, make_glb([]))
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
    manifest["assets"]["character_base_humanoid_v0"]["required_nodes"] = []
    # GLB still has the socket node, but the manifest no longer declares it.
    report = run(build_root(tmp_path, manifest=manifest))
    assert any("required_nodes" in f for f in report.failures)


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


def test_glb_missing_required_node(tmp_path):
    # Manifest declares the socket, but the GLB body lacks that node.
    report = run(build_root(tmp_path, char_nodes=["some_other_node"]))
    assert any("glb node" in f for f in report.failures)


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
