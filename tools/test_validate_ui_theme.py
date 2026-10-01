"""Focused coverage for UI theme token resolution and rarity parity."""
from __future__ import annotations

import copy
import json
from pathlib import Path

from tools.validate_ui_theme import validate_ui_theme

ROOT = Path(__file__).resolve().parents[1]


class Report:
    def __init__(self) -> None:
        self.failures: list[str] = []

    def ok(self, _label: str) -> None:
        pass

    def fail(self, _label: str, detail: str) -> None:
        self.failures.append(detail)


def _catalogs() -> tuple[dict, dict]:
    theme = json.loads((ROOT / "shared/assets/ui_theme.v0.json").read_text())
    templates = json.loads((ROOT / "shared/rules/item_templates.v0.json").read_text())
    return theme, templates


def _failures(theme: dict, templates: dict) -> list[str]:
    report = Report()
    validate_ui_theme(report, theme, templates)
    return report.failures


def test_shipped_catalog_is_valid() -> None:
    assert _failures(*_catalogs()) == []


def test_unknown_color_ref_fails() -> None:
    theme, templates = _catalogs()
    theme = copy.deepcopy(theme)
    theme["frames"]["panel"]["bg"] = "no_such_color"
    assert any("no_such_color" in f for f in _failures(theme, templates))


def test_unknown_spacing_and_state_ref_fail() -> None:
    theme, templates = _catalogs()
    theme = copy.deepcopy(theme)
    theme["frames"]["slot"]["margin"] = "no_such_spacing"
    theme["frames"]["slot"]["states"]["hover"]["border"] = "no_such_color"
    failures = " ".join(_failures(theme, templates))
    assert "no_such_spacing" in failures and "no_such_color" in failures


def test_missing_rarity_fails() -> None:
    theme, templates = _catalogs()
    theme = copy.deepcopy(theme)
    del theme["rarity_slot_backgrounds"]["set"]
    assert any("missing" in f for f in _failures(theme, templates))


def test_duplicate_rarity_background_fails() -> None:
    theme, templates = _catalogs()
    theme = copy.deepcopy(theme)
    theme["rarity_slot_backgrounds"]["magic"] = theme["rarity_slot_backgrounds"]["common"]
    assert any("duplicate" in f for f in _failures(theme, templates))


def test_missing_inventory_rarity_token_fails() -> None:
    theme, templates = _catalogs()
    theme = copy.deepcopy(theme)
    del theme["colors"]["inventory_rarity_rare"]
    assert any("inventory_rarity_rare" in f for f in _failures(theme, templates))
