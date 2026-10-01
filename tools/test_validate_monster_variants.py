"""Focused coverage for monster variant catalog cross-checks (v511)."""
from __future__ import annotations

import copy
import json
from pathlib import Path

from tools.validate_monster_variants import validate_monster_variants

ROOT = Path(__file__).resolve().parents[1]


class Report:
    def __init__(self) -> None:
        self.failures: list[str] = []

    def ok(self, _label: str) -> None:
        pass

    def fail(self, _label: str, detail: str) -> None:
        self.failures.append(detail)


def _catalogs() -> tuple[dict, dict]:
    kit = json.loads((ROOT / "shared/assets/kit_monster_presentation.v0.json").read_text())
    dungeon = json.loads((ROOT / "shared/rules/dungeon_generation.v0.json").read_text())
    return kit, dungeon


def _failures(kit: dict, dungeon: dict) -> list[str]:
    report = Report()
    validate_monster_variants(report, kit, dungeon)
    return report.failures


def test_shipped_catalog_passes() -> None:
    assert _failures(*_catalogs()) == []


def test_unknown_rarity_key_fails() -> None:
    kit, dungeon = _catalogs()
    kit = copy.deepcopy(kit)
    kit["variants"]["rarities"]["mythic"] = copy.deepcopy(kit["variants"]["rarities"]["rare"])
    assert any("extra" in f for f in _failures(kit, dungeon))


def test_missing_palette_fails() -> None:
    kit, dungeon = _catalogs()
    kit = copy.deepcopy(kit)
    del kit["variants"]["depth"]["deep_vault"]
    assert any("deep_vault" in f for f in _failures(kit, dungeon))


def test_unknown_family_fails() -> None:
    kit, dungeon = _catalogs()
    kit = copy.deepcopy(kit)
    kit["variants"]["families"]["monster_nope"] = {"eye_mesh": "", "aura_radius_scale": 1.0}
    assert any("monster_nope" in f for f in _failures(kit, dungeon))


def test_duplicate_rarity_looks_fail() -> None:
    kit, dungeon = _catalogs()
    kit = copy.deepcopy(kit)
    kit["variants"]["rarities"]["unique"] = copy.deepcopy(kit["variants"]["rarities"]["rare"])
    assert any("duplicates" in f for f in _failures(kit, dungeon))


def test_non_neutral_common_fails() -> None:
    kit, dungeon = _catalogs()
    kit = copy.deepcopy(kit)
    kit["variants"]["rarities"]["common"]["tint"]["strength"] = 0.3
    assert any("common" in f for f in _failures(kit, dungeon))
