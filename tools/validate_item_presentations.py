"""Item-presentation cross checks for shared validation."""
from __future__ import annotations

from pathlib import Path
from typing import Any, Callable

HAND_SLOTS = frozenset({"main_hand", "off_hand"})


def validate_item_presentations(
    report: Any,
    *,
    assets_dir: Path,
    load_json: Callable[[Path], Any],
    items: dict[str, Any],
    item_templates: dict[str, Any],
    manifest_assets: dict[str, Any],
) -> None:
    # Every current item needs display metadata, and no presentation entry
    # should point at a missing item. Drift here causes silent client fallbacks.
    item_presentations = load_json(assets_dir / "item_presentations.v0.json")
    presentation_families = item_presentations["families"]
    presentations = item_presentations["items"]
    expected_families = {str(template.get("item_type", "")) for template in item_templates["templates"].values()}
    expected_families |= {"gold", "quest", "health_potion", "mana_potion"}
    missing_families = sorted(expected_families - set(presentation_families))
    if missing_families:
        report.fail("item presentation families", f"missing families: {missing_families}")
    else:
        report.ok("item presentation families cover every item family")

    # v487: hand items (main_hand/off_hand) take their ground model from item_visuals.v0.json, the
    # model the hero wields. A family 3d_model on a hand family would be a second, drifting mapping.
    item_visuals = load_json(assets_dir / "item_visuals.v0.json")["item_visuals"]
    hand_families = {
        str(presentation.get("family", ""))
        for def_id, presentation in presentations.items()
        if item_visuals.get(def_id, {}).get("slot") in HAND_SLOTS
    }
    for family_id, family in sorted(presentation_families.items()):
        model_id = family.get("3d_model")
        if model_id and family_id in hand_families:
            report.fail(
                "item presentation family 3d_model",
                f"{family_id}: hand-item family must not set 3d_model (ground loot uses item_visuals.v0.json)",
            )
        elif model_id and model_id not in manifest_assets:
            report.fail("item presentation family 3d_model", f"{family_id}: unknown asset {model_id}")
        elif model_id:
            report.ok(f"item presentation family {family_id} 3d_model resolves")
        _validate_ground_model_tiers(report, family_id, family.get("ground_model_tiers"), manifest_assets)

    for def_id in sorted(presentations):
        if def_id not in items["items"] and def_id not in item_templates["templates"]:
            report.fail("item_presentations key", f"{def_id} not in items.v0.json or item_templates.v0.json")
            continue
        family_id = str(presentations[def_id].get("family", ""))
        if family_id not in presentation_families:
            report.fail("item_presentations family", f"{def_id}: unknown family {family_id}")
        elif presentations[def_id].get("3d_model") and presentations[def_id]["3d_model"] not in manifest_assets:
            report.fail("item_presentations 3d_model", f"{def_id}: unknown asset {presentations[def_id]['3d_model']}")
        else:
            report.ok(f"item_presentations {def_id} resolves to item/template rules and family {family_id}")

    missing_presentations = sorted((set(items["items"]) | set(item_templates["templates"])) - set(presentations))
    if missing_presentations:
        report.fail("item_presentations coverage", f"missing entries: {missing_presentations}")
    else:
        report.ok("item_presentations covers all item rules")


def _validate_ground_model_tiers(report: Any, family_id: str, tiers: Any, manifest_assets: dict[str, Any]) -> None:
    """v490: amount tiers must resolve to manifest assets and ascend strictly by min_amount."""
    if not tiers:
        return
    previous = 0
    for tier in tiers:
        asset_id, min_amount = tier.get("asset_id"), int(tier.get("min_amount", 0))
        if asset_id not in manifest_assets:
            report.fail("item presentation ground_model_tiers", f"{family_id}: unknown asset {asset_id}")
        elif min_amount <= previous:
            report.fail("item presentation ground_model_tiers", f"{family_id}: min_amount {min_amount} not above {previous}")
        else:
            report.ok(f"item presentation family {family_id} tier {min_amount} resolves")
        previous = max(previous, min_amount)
