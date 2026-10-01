"""Cross-catalog guard for the client-only monster variant looks (v511)."""
from __future__ import annotations

from typing import Any


def validate_monster_variants(report: Any, kit_monsters: dict, dungeon_generation: dict) -> None:
    variants = kit_monsters.get("variants")
    if variants is None:
        report.fail("monster variants", "kit_monster_presentation.variants missing")
        return
    rarity_ids = {str(r["id"]) for r in dungeon_generation["monster_rarities"]}
    palette_ids = {str(p["id"]) for p in dungeon_generation["biome_palettes"]}
    _same_keys(report, "monster variant rarities", set(variants["rarities"]), rarity_ids)
    _same_keys(report, "monster variant depth palettes", set(variants["depth"]), palette_ids)

    unknown = set(variants["families"]) - set(kit_monsters["monsters"])
    if unknown:
        report.fail("monster variant families", f"not kit monster scenes: {sorted(unknown)}")
    else:
        report.ok("monster variant families are kit monster scenes")

    looks = variants["rarities"]
    neutral = looks.get("common")
    if neutral is not None:
        clean = (
            float(neutral["tint"]["strength"]) == 0.0
            and float(neutral["scale_multiplier"]) == 1.0
            and "aura" not in neutral
        )
        if clean:
            report.ok("common monster variant is neutral")
        else:
            report.fail("monster variant common", "common must have zero tint, scale 1.0 and no aura")
    signatures: dict[tuple, str] = {}
    for rarity_id, look in sorted(looks.items()):
        if rarity_id == "common":
            continue
        if "eye" not in look and "aura" not in look:
            report.fail("monster variant cue", f"{rarity_id}: needs an eye or aura cue besides tint")
        sig = (
            round(float(look["tint"]["strength"]), 3),
            look["tint"]["color"].lower(),
            round(float(look["scale_multiplier"]), 3),
            round(float(look.get("aura", {}).get("alpha", 0.0)), 3),
        )
        if sig in signatures:
            report.fail("monster variant distinct", f"{rarity_id} duplicates {signatures[sig]}")
        signatures[sig] = rarity_id
    report.ok("monster variant rarity looks checked")


def _same_keys(report: Any, label: str, actual: set, expected: set) -> None:
    if actual != expected:
        report.fail(label, f"missing={sorted(expected - actual)}, extra={sorted(actual - expected)}")
    else:
        report.ok(f"{label} match the gameplay catalog")
