"""Showme path and script layout checks."""
from __future__ import annotations

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def test_showme_scripts_live_under_client() -> None:
    capture = ROOT / "client" / "scripts" / "showme" / "visual_capture.gd"
    icons = ROOT / "client" / "scripts" / "showme" / "showme_item_icons_capture.gd"
    assert capture.is_file(), f"missing {capture}"
    assert icons.is_file(), f"missing {icons}"
    text = capture.read_text(encoding="utf-8")
    assert "res://scripts/showme/showme_item_icons_capture.gd" in text
    assert "res://skills/showme" not in text
    assert not (ROOT / "skills" / "showme" / "scripts" / "visual_capture.gd").exists()


def test_render_focus_points_at_client_capture() -> None:
    script = (ROOT / "skills" / "showme" / "scripts" / "render_focus.py").read_text(encoding="utf-8")
    assert 'client" / "scripts" / "showme" / "visual_capture.gd"' in script
