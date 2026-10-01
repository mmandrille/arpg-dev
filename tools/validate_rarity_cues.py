"""Cross-catalog guard for the client-only equipment rarity vocabulary."""
from __future__ import annotations

from typing import Any


def validate_rarity_cues(report: Any, cues: dict, item_templates: dict) -> None:
    entries = cues["rarities"]
    expected = set(item_templates["rarities"])
    actual = set(entries)
    if actual != expected:
        report.fail("rarity cues coverage", f"missing={sorted(expected - actual)}, extra={sorted(actual - expected)}")
    else:
        report.ok("rarity cues cover exactly the gameplay rarity keys")
    for field in ("name", "short", "shape", "world_symbol"):
        values = [str(entry[field]).casefold() for entry in entries.values()]
        if len(values) != len(set(values)):
            report.fail("rarity cues distinct", f"duplicate {field}")
        else:
            report.ok(f"rarity cues distinct {field}")
