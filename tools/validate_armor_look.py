"""Cross-checks for shared/assets/armor_look.v0.json (ADR-0018 D5 / P3b, v483).

Equipped armor no longer mounts meshes: tint/headgear slots recolour the kit hero (armor_look),
jewelry has no world visual, and only weapon slots stay in item_visuals. Together the two catalogs
must cover every equippable item.
"""
from __future__ import annotations

from typing import Any, Callable, Iterable, Mapping


def _slot_in(rule_slot: str | None, catalog_slots: Iterable[str], slot_matches: Callable[[str | None, str], bool]) -> bool:
    return any(slot_matches(rule_slot, slot) for slot in catalog_slots)


def validate_armor_look(
    report: Any,
    look: dict,
    visuals: dict,
    equippables: Mapping[str, str | None],
    slot_matches: Callable[[str | None, str], bool],
) -> None:
    """`equippables`: equippable item_def_id -> rule slot; `visuals`: item_visuals entries."""
    slots: dict = look.get("slots", {})
    regions: dict = look.get("regions", {})
    no_visual = look.get("no_world_visual_slots", [])
    for slot, entry in sorted(slots.items()):
        if entry.get("region") not in regions:
            report.fail("armor_look region", f"slot {slot} uses unknown region {entry.get('region')}")
        else:
            report.ok(f"armor_look slot {slot} -> region {entry['region']}")
    for slot in look.get("body_precedence", []):
        if slot not in slots:
            report.fail("armor_look precedence", f"body_precedence lists {slot}, which is not a look slot")
    overlap = sorted(set(slots) & set(no_visual))
    if overlap:
        report.fail("armor_look slots", f"slots both tinted and without world visual: {overlap}")

    for def_id, entry in sorted(look.get("items", {}).items()):
        rule_slot = equippables.get(def_id, "<missing>")
        if rule_slot == "<missing>":
            report.fail("armor_look item", f"{def_id} is not an equippable item/template")
        elif not _slot_in(rule_slot, slots, slot_matches):
            report.fail("armor_look item", f"{def_id}: slot {rule_slot} is not a tint/headgear slot")
        else:
            report.ok(f"armor_look {def_id} colours slot {rule_slot}")

    missing = []
    for def_id, rule_slot in sorted(equippables.items()):
        if def_id in visuals:
            continue
        if _slot_in(rule_slot, no_visual, slot_matches):
            continue
        if _slot_in(rule_slot, slots, slot_matches) and def_id in look.get("items", {}):
            continue
        missing.append(def_id)
    if missing:
        report.fail("equipment visual coverage", f"not covered by item_visuals or armor_look: {missing}")
    else:
        report.ok("item_visuals + armor_look cover all equippable equipment")
