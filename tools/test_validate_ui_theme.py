"""Focused coverage for UI theme token resolution and rarity parity."""
from __future__ import annotations

import copy
import json
from pathlib import Path

from tools.validate_ui_theme import unresolved_client_tokens, validate_ui_theme

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


def test_missing_skill_tree_frame_fails() -> None:
    theme, templates = _catalogs()
    theme = copy.deepcopy(theme)
    del theme["frames"]["skill_node_available"]
    assert any("frames.skill_node_available" in f for f in _failures(theme, templates))


def test_missing_skill_state_color_fails() -> None:
    theme, templates = _catalogs()
    theme = copy.deepcopy(theme)
    del theme["colors"]["skill_node_locked_text"]
    assert any("colors.skill_node_locked_text" in f for f in _failures(theme, templates))


def test_shipped_client_call_sites_resolve() -> None:
    theme, _ = _catalogs()
    assert unresolved_client_tokens(theme, ROOT / "client/scripts") == []


def test_scan_finds_literal_call_sites() -> None:
    # Guard against the scanner silently matching nothing: a removed token must be reported.
    theme, _ = _catalogs()
    theme = copy.deepcopy(theme)
    del theme["colors"]["hud_bronze"]
    del theme["frames"]["hud_panel"]
    found = " ".join(unresolved_client_tokens(theme, ROOT / "client/scripts"))
    assert "hud_bronze" in found and "hud_panel" in found


def test_scan_ignores_dynamic_tokens_and_comments(tmp_path: Path) -> None:
    theme, _ = _catalogs()
    (tmp_path / "a.gd").write_text(
        '# UiTheme.color("nope_comment")\n'
        'var a = UiTheme.color("inventory_rarity_" + r)\n'
        'var b = UiTheme.color("panel_bg")\n'
        'var c = UiTheme.frame("no_such_frame", "hover")\n'
        'UiTheme.apply_font(label, "no_such_role")\n'
    )
    found = unresolved_client_tokens(theme, tmp_path)
    assert len(found) == 2
    assert any("no_such_frame" in f for f in found) and any("no_such_role" in f for f in found)
    assert not any("nope_comment" in f or "inventory_rarity_" in f for f in found)
