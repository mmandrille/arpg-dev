"""Focused tests for tools/validate_item_presentations.py (v487 hand-family 3d_model rule)."""
from __future__ import annotations

import copy
import json
from pathlib import Path
from typing import Any

from tools.validate_item_presentations import validate_item_presentations

ROOT = Path(__file__).resolve().parents[1]
ASSETS = ROOT / "shared" / "assets"


class Report:
    def __init__(self) -> None:
        self.failures: list[tuple[str, str]] = []

    def ok(self, _label: str) -> None:
        pass

    def fail(self, label: str, detail: str) -> None:
        self.failures.append((label, detail))


def _load(path: Path) -> Any:
    return json.loads(path.read_text(encoding="utf-8"))


def _run(overrides: dict[str, Any] | None = None) -> Report:
    overrides = overrides or {}

    def load_json(path: Path) -> Any:
        return copy.deepcopy(overrides[path.name]) if path.name in overrides else _load(path)

    report = Report()
    validate_item_presentations(
        report,
        assets_dir=ASSETS,
        load_json=load_json,
        items=_load(ROOT / "shared" / "rules" / "items.v0.json"),
        item_templates=_load(ROOT / "shared" / "rules" / "item_templates.v0.json"),
        manifest_assets=_load(ROOT / "assets" / "manifests" / "assets.v0.json")["assets"],
    )
    return report


def _presentations_with(family_id: str, model_id: str) -> dict[str, Any]:
    data = _load(ASSETS / "item_presentations.v0.json")
    data["families"][family_id]["3d_model"] = model_id
    return {"item_presentations.v0.json": data}


def test_committed_presentations_pass() -> None:
    assert _run().failures == []


def test_hand_family_with_3d_model_fails() -> None:
    visuals = _load(ASSETS / "item_visuals.v0.json")["item_visuals"]
    presentations = _load(ASSETS / "item_presentations.v0.json")["items"]
    hand_family = next(
        presentations[def_id]["family"]
        for def_id, visual in sorted(visuals.items())
        if visual["slot"] == "main_hand" and def_id in presentations
    )
    model_id = visuals[next(iter(sorted(visuals)))]["asset_id"]

    failures = _run(_presentations_with(hand_family, model_id)).failures

    assert any(hand_family in detail and "item_visuals" in detail for _label, detail in failures), failures


def test_armor_family_keeps_its_fallback_model() -> None:
    families = _load(ASSETS / "item_presentations.v0.json")["families"]
    armor_family, family = next((fid, fam) for fid, fam in sorted(families.items()) if fam.get("3d_model"))

    assert _run(_presentations_with(armor_family, family["3d_model"])).failures == []


def _gold_tiers_with(mutate) -> dict[str, Any]:
    data = _load(ASSETS / "item_presentations.v0.json")
    mutate(data["families"]["gold"]["ground_model_tiers"])
    return {"item_presentations.v0.json": data}


def test_gold_tier_with_unknown_asset_fails() -> None:
    failures = _run(_gold_tiers_with(lambda tiers: tiers[0].update(asset_id="not_an_asset"))).failures
    assert any("not_an_asset" in detail for _label, detail in failures), failures


def test_gold_tiers_must_ascend() -> None:
    failures = _run(_gold_tiers_with(lambda tiers: tiers.reverse())).failures
    assert any("min_amount" in detail for _label, detail in failures), failures
