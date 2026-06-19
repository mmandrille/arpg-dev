from __future__ import annotations

import json
from pathlib import Path

import pytest

from tools.assets import model_catalog


def _write(path: Path, data: dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data), encoding="utf-8")


def _repo(tmp_path: Path) -> Path:
    root = tmp_path / "repo"
    _write(root / model_catalog.MANIFEST_REL, {
        "version": 0,
        "assets": {
            "character_paladin_v0": {
                "type": "character",
                "runtime_path": "client/assets/characters/paladin/paladin.glb",
                "format": "glb",
                "required_nodes": [],
            },
            "monster_dummy_v0": {
                "type": "monster",
                "runtime_path": "client/assets/monsters/dummy/monster_dummy.glb",
                "format": "glb",
                "required_nodes": [],
            },
            "monster_tiny_flyer_v0": {
                "type": "monster",
                "runtime_path": "client/assets/monsters/tiny_flyer/monster_tiny_flyer.glb",
                "format": "glb",
                "required_nodes": [],
            },
            "weapon_rusty_sword_v0": {
                "type": "equipment",
                "slot": "main_hand",
                "runtime_path": "client/assets/equipment/weapons/rusty_sword/rusty_sword.glb",
                "format": "glb",
                "required_nodes": [],
            },
        },
    })
    _write(root / model_catalog.CLASS_PRESENTATIONS_REL, {
        "version": 0,
        "classes": {
            "paladin": {"model": {"asset_id": "character_paladin_v0", "scale": 10.0}},
        },
    })
    _write(root / model_catalog.MONSTER_VISUALS_REL, {
        "version": 0,
        "monster_visuals": {
            "training_dummy": {
                "asset_id": "monster_dummy_v0",
                "scene": "monster_dummy",
                "scale": 1.0,
                "height_offset": 0.0,
                "animation_profile": "ground_biped",
            },
            "mercenary_guard": {
                "asset_id": "monster_dummy_v0",
                "scene": "monster_dummy",
                "scale": 1.05,
                "height_offset": 0.0,
                "animation_profile": "ground_biped",
            },
            "dungeon_bat": {
                "asset_id": "monster_tiny_flyer_v0",
                "scene": "monster_tiny_flyer",
                "scale": 0.7,
                "height_offset": 0.9,
                "animation_profile": "hover_flyer",
            },
            "bad_item_reference": {
                "asset_id": "weapon_rusty_sword_v0",
                "scene": "not_previewable",
                "scale": 1.0,
                "height_offset": 0.0,
                "animation_profile": "ground_biped",
            },
        },
    })
    return root


def test_catalog_discovers_characters_and_monsters(tmp_path: Path) -> None:
    rows = model_catalog.load_catalog(_repo(tmp_path))

    assert [row.asset_id for row in rows] == [
        "character_paladin_v0",
        "monster_dummy_v0",
        "monster_tiny_flyer_v0",
    ]
    paladin = model_catalog.resolve("character_paladin_v0", _repo(tmp_path))
    assert paladin.asset_type == "character"
    assert paladin.runtime_path == "client/assets/characters/paladin/paladin.glb"
    assert paladin.used_by == ("paladin",)
    assert paladin.scale == 10.0


def test_generated_catalog_round_trips_discovered_rows(tmp_path: Path) -> None:
    root = _repo(tmp_path)
    path = model_catalog.write_generated_catalog(root)

    assert path == root / model_catalog.GENERATED_CATALOG_REL
    assert [row.asset_id for row in model_catalog.load_generated_catalog(root)] == [
        "character_paladin_v0",
        "monster_dummy_v0",
        "monster_tiny_flyer_v0",
    ]
    assert model_catalog.generated_catalog_mismatch(root) == ""


def test_repo_generated_catalog_matches_source_data() -> None:
    assert model_catalog.generated_catalog_mismatch(model_catalog.ROOT) == ""


def test_catalog_groups_multiple_used_by_labels(tmp_path: Path) -> None:
    dummy = model_catalog.resolve("monster_dummy_v0", _repo(tmp_path))

    assert dummy.used_by == ("mercenary_guard", "training_dummy")
    assert "used_by=mercenary_guard,training_dummy" in model_catalog.format_row(dummy)


def test_catalog_excludes_equipment_even_if_referenced(tmp_path: Path) -> None:
    rows = model_catalog.load_catalog(_repo(tmp_path))

    assert "weapon_rusty_sword_v0" not in {row.asset_id for row in rows}


def test_resolve_unknown_asset_raises(tmp_path: Path) -> None:
    with pytest.raises(KeyError):
        model_catalog.resolve("missing_model_v0", _repo(tmp_path))


def test_cli_unknown_asset_points_to_model_list(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    status = model_catalog.main(["--root", str(_repo(tmp_path)), "resolve", "missing_model_v0"])

    assert status == 2
    assert "make model-list" in capsys.readouterr().err
