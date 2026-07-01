"""Tests for tools/content/build_codex.py."""

from __future__ import annotations

import json
from pathlib import Path

from tools.content.build_codex import build_index, max_item_level_for_depth

ROOT = Path(__file__).resolve().parents[1]


def test_build_codex_includes_barbarian_class_page() -> None:
    payload = build_index(["classes"])
    classes = next(ch for ch in payload["chapters"] if ch["id"] == "classes")
    barbarian = next(page for page in classes["pages"] if page["id"] == "class:barbarian")
    body = json.dumps(barbarian)
    assert "Barbarian" in body
    assert "Rage" in body or "Tier" in body


def test_build_codex_dagger_family_includes_rogue_affinity() -> None:
    payload = build_index(["item_families"])
    families = next(ch for ch in payload["chapters"] if ch["id"] == "item_families")
    dagger = next(page for page in families["pages"] if page["id"] == "family:dagger")
    body = json.dumps(dagger)
    assert "Rogue" in body
    assert "attack speed" in body


def test_build_codex_full_index_has_six_chapters() -> None:
    payload = build_index()
    assert len(payload["chapters"]) == 6
    chapter_ids = {chapter["id"] for chapter in payload["chapters"]}
    assert chapter_ids == {"concepts", "classes", "skills", "item_families", "resources", "loot"}


def test_max_item_level_formula_matches_v389() -> None:
    assert max_item_level_for_depth(25, 10) == 2
    assert max_item_level_for_depth(9, 10) == 1
