from __future__ import annotations

import json
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[2]
_UNIQUE_EFFECT_IDS: list[str] | None = None


def enabled_unique_effect_ids() -> list[str]:
    global _UNIQUE_EFFECT_IDS
    if _UNIQUE_EFFECT_IDS is None:
        data = json.loads((ROOT / "shared" / "rules" / "unique_effects.v0.json").read_text(encoding="utf-8"))
        _UNIQUE_EFFECT_IDS = sorted(
            str(effect_id)
            for effect_id, effect in data.get("effects", {}).items()
            if bool(effect.get("enabled")) and str(effect.get("status", "")) == "ready"
        )
    return list(_UNIQUE_EFFECT_IDS)


def assert_inventory_unique_effect_coverage(inventory: list[dict], assertion: dict[str, Any], where: str) -> None:
    expected = enabled_unique_effect_ids()
    required_named = list(assertion.get("required_named_uniques", []))
    required_names = {str(unique.get("display_name", "")) for unique in required_named}
    seen: dict[str, list[dict[str, Any]]] = {effect_id: [] for effect_id in expected}
    unique_rows = []
    for item in inventory:
        if str(item.get("rarity", "")) == "unique":
            unique_rows.append(item)
    coverage_rows = [item for item in unique_rows if str(item.get("display_name", "")) not in required_names]
    for item in coverage_rows:
        for effect_id in item.get("effect_ids", []):
            if str(effect_id) in seen:
                seen[str(effect_id)].append(item)
    missing = [effect_id for effect_id in expected if not seen[effect_id]]
    duplicate = [effect_id for effect_id, rows in seen.items() if len(rows) > 1]
    if missing or duplicate:
        raise AssertionError(f"{where}: unique effect coverage missing={missing} duplicate={duplicate} inventory={inventory}")
    if assertion.get("equals_enabled", True) and len(coverage_rows) != len(expected):
        raise AssertionError(f"{where}: unique coverage rows={len(coverage_rows)} want enabled effects={len(expected)} rows={coverage_rows}")
    bad = [item for item in unique_rows if len(item.get("effect_ids", [])) != 1]
    if assertion.get("requires_single_effect", True) and bad:
        raise AssertionError(f"{where}: unique rows must have exactly one effect: {bad}")
    for unique in required_named:
        display_name = str(unique.get("display_name", ""))
        matches = [item for item in unique_rows if str(item.get("display_name", "")) == display_name]
        if len(matches) != 1:
            raise AssertionError(f"{where}: named unique {display_name!r} count={len(matches)} want 1 rows={unique_rows}")
        expected_effects = [str(effect_id) for effect_id in unique.get("effect_ids", [])]
        if expected_effects and matches[0].get("effect_ids", []) != expected_effects:
            raise AssertionError(f"{where}: named unique {display_name!r} effect_ids {matches[0].get('effect_ids', [])} != {expected_effects}")
