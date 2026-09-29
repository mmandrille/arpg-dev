"""Tests for the ADR-0018 P0 kit inspector."""
from __future__ import annotations

import json
from pathlib import Path

from tools.assets import gen_glb, inspect_kit
from tools.assets.test_glb_reader import write_gltf_fixture


def _kit(tmp_path: Path) -> Path:
    kit = tmp_path / "kit"
    (kit / "chars").mkdir(parents=True)
    (kit / "chars" / "dummy.glb").write_bytes(gen_glb.monster_dummy_glb())
    (kit / "chars" / "skeleton.glb").write_bytes(gen_glb.monster_skeleton_glb())
    (kit / "chars" / "skeleton_copy.glb").write_bytes(gen_glb.monster_skeleton_glb())
    write_gltf_fixture(kit)
    (kit / "legacy.fbx").write_bytes(b"fbx")
    return kit


def test_report_covers_files_rigs_and_gaps(tmp_path: Path) -> None:
    report = inspect_kit.build_report([_kit(tmp_path)])
    paths = [entry["path"] for entry in report["files"]]
    assert paths == ["chars/dummy.glb", "chars/skeleton.glb", "chars/skeleton_copy.glb", "fixture.gltf"]
    assert report["unsupported"] == ["legacy.fbx"]
    by_path = {entry["path"]: entry for entry in report["files"]}
    for key in ("sha256", "node_tree", "meshes", "total_triangles", "materials", "images", "skins", "animations", "extents"):
        assert key in by_path["chars/dummy.glb"]
    assert by_path["chars/dummy.glb"]["skins"][0]["joints"] == ["root", "pivot"]
    assert by_path["fixture.gltf"]["total_triangles"] == 1
    assert by_path["fixture.gltf"]["animations"][0]["name"] == "Wave"
    # Two identical skeleton rigs group together; the dummy rig is separate.
    groups = report["rig_groups"]
    assert [len(g["files"]) for g in groups] == [2, 1]
    assert groups[0]["files"] == ["chars/skeleton.glb", "chars/skeleton_copy.glb"]


def test_cli_writes_json_and_markdown(tmp_path: Path) -> None:
    out = tmp_path / "report"
    assert inspect_kit.main([str(_kit(tmp_path)), "--out", str(out)]) == 0
    data = json.loads((out / "report.json").read_text(encoding="utf-8"))
    assert len(data["files"]) == 4
    markdown = (out / "report.md").read_text(encoding="utf-8")
    assert "## Rig comparison" in markdown
    assert "Missing vs group 1" in markdown
    assert "legacy.fbx" in markdown


def test_cli_rejects_missing_input(tmp_path: Path) -> None:
    assert inspect_kit.main([str(tmp_path / "nope")]) == 2


def test_malformed_file_is_reported_not_fatal(tmp_path: Path) -> None:
    bad = tmp_path / "broken.glb"
    bad.write_bytes(b"glTF-but-truncated")
    entry = inspect_kit.inspect_file(bad)
    assert "error" in entry and entry["sha256"]
