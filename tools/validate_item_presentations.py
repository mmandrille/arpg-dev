"""Item-presentation cross checks for shared validation."""
from __future__ import annotations

from pathlib import Path
from typing import Any, Callable


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

    for family_id, family in sorted(presentation_families.items()):
        model_id = family.get("3d_model")
        if model_id and model_id not in manifest_assets:
            report.fail("item presentation family 3d_model", f"{family_id}: unknown asset {model_id}")
        elif model_id:
            report.ok(f"item presentation family {family_id} 3d_model resolves")

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
