"""Focused coverage for rarity cue key parity and non-color distinctions."""
from __future__ import annotations

import copy
import json
from pathlib import Path

from tools.validate_rarity_cues import validate_rarity_cues

ROOT = Path(__file__).resolve().parents[1]


class Report:
    def __init__(self) -> None:
        self.failures: list[str] = []

    def ok(self, _label: str) -> None:
        pass

    def fail(self, _label: str, detail: str) -> None:
        self.failures.append(detail)


def _catalogs() -> tuple[dict, dict]:
    cues = json.loads((ROOT / "shared/assets/rarity_cues.v0.json").read_text())
    templates = json.loads((ROOT / "shared/rules/item_templates.v0.json").read_text())
    return cues, templates


def _failures(cues: dict, templates: dict) -> list[str]:
    report = Report()
    validate_rarity_cues(report, cues, templates)
    return report.failures


def test_committed_catalog_matches_rarity_rules() -> None:
    cues, templates = _catalogs()
    assert _failures(cues, templates) == []


def test_missing_and_extra_rarity_fail() -> None:
    cues, templates = _catalogs()
    missing = copy.deepcopy(cues)
    missing["rarities"].pop(next(iter(templates["rarities"])))
    assert any("missing=" in failure for failure in _failures(missing, templates))
    extra = copy.deepcopy(cues)
    extra["rarities"]["hidden"] = copy.deepcopy(next(iter(cues["rarities"].values())))
    assert any("extra=" in failure for failure in _failures(extra, templates))


def test_duplicate_non_color_cues_fail() -> None:
    cues, templates = _catalogs()
    rarity_names = list(templates["rarities"])
    for field in ("short", "shape", "world_symbol"):
        duplicate = copy.deepcopy(cues)
        duplicate["rarities"][rarity_names[1]][field] = duplicate["rarities"][rarity_names[0]][field]
        assert any(f"duplicate {field}" in failure for failure in _failures(duplicate, templates))
