#!/usr/bin/env python3
"""Validate all shared JSON contracts.

Covers everything under shared/protocol, shared/rules, and shared/golden:

  1. Every *.schema.json is itself a valid JSON Schema (draft 2020-12).
  2. Every data / example / golden instance validates against its schema.
  3. Cross-consistency drift guards (e.g. the golden damage params match
     combat.v0.json) so the two-language stack cannot silently diverge.

Exit code is non-zero if anything fails. Run via `make validate-shared`.
"""
from __future__ import annotations

import json
import sys
from pathlib import Path

from jsonschema import Draft202012Validator

ROOT = Path(__file__).resolve().parent.parent
SHARED = ROOT / "shared"
PROTOCOL = SHARED / "protocol"
RULES = SHARED / "rules"
GOLDEN = SHARED / "golden"
ASSETS = SHARED / "assets"


def load(path: Path):
    with path.open(encoding="utf-8") as fh:
        return json.load(fh)


class Report:
    def __init__(self) -> None:
        self.passed = 0
        self.failures: list[str] = []

    def ok(self, label: str) -> None:
        self.passed += 1
        print(f"  ok   {label}")

    def fail(self, label: str, detail: str) -> None:
        self.failures.append(f"{label}: {detail}")
        print(f"  FAIL {label}: {detail}")


def schema_for(instance_path: Path) -> Path:
    """Map an instance file to the schema that should validate it."""
    rel = instance_path.relative_to(SHARED)
    parts = rel.parts
    if parts[0] == "rules":
        # foo.v0.json -> foo.v0.schema.json
        return RULES / instance_path.name.replace(".v0.json", ".v0.schema.json")
    if parts[0] == "assets":
        # foo.v0.json -> foo.v0.schema.json
        return ASSETS / instance_path.name.replace(".v0.json", ".v0.schema.json")
    if parts[0] == "golden":
        # foo.json -> foo.v0.schema.json
        return GOLDEN / (instance_path.stem + ".v0.schema.json")
    if parts[0] == "protocol" and parts[1] == "examples":
        name = instance_path.name
        if name == "session_snapshot.json":
            return PROTOCOL / "session_snapshot.v0.schema.json"
        if name == "state_delta.json":
            return PROTOCOL / "state_delta.v0.schema.json"
        return PROTOCOL / "messages.v0.schema.json"
    raise ValueError(f"no schema mapping for {instance_path}")


def iter_schemas() -> list[Path]:
    return sorted(SHARED.rglob("*.schema.json"))


def iter_instances() -> list[Path]:
    instances: list[Path] = []
    instances += sorted(p for p in RULES.glob("*.v0.json") if not p.name.endswith(".schema.json"))
    instances += sorted(p for p in ASSETS.glob("*.v0.json") if not p.name.endswith(".schema.json"))
    instances += sorted(p for p in GOLDEN.glob("*.json") if not p.name.endswith(".schema.json"))
    instances += sorted(PROTOCOL.glob("examples/*.json"))
    return instances


def validate_schemas(report: Report) -> None:
    print("[1] meta-validating schemas (draft 2020-12)")
    for schema_path in iter_schemas():
        label = str(schema_path.relative_to(ROOT))
        try:
            Draft202012Validator.check_schema(load(schema_path))
            report.ok(label)
        except Exception as exc:  # noqa: BLE001 - surface any schema error
            report.fail(label, str(exc).splitlines()[0])


def validate_instances(report: Report) -> None:
    print("[2] validating instances against schemas")
    for instance_path in iter_instances():
        label = str(instance_path.relative_to(ROOT))
        try:
            schema = load(schema_for(instance_path))
            validator = Draft202012Validator(schema)
            errors = sorted(validator.iter_errors(load(instance_path)), key=lambda e: e.path)
            if errors:
                first = errors[0]
                loc = "/".join(str(p) for p in first.path) or "<root>"
                report.fail(label, f"at {loc}: {first.message}")
            else:
                report.ok(label)
        except Exception as exc:  # noqa: BLE001
            report.fail(label, str(exc).splitlines()[0])


def cross_checks(report: Report) -> None:
    print("[3] cross-consistency drift guards")
    combat = load(RULES / "combat.v0.json")
    items = load(RULES / "items.v0.json")
    monsters = load(RULES / "monsters.v0.json")
    loot = load(RULES / "loot_tables.v0.json")
    damage_golden = load(GOLDEN / "damage_formula.json")
    loot_golden = load(GOLDEN / "loot_roll.json")
    slice_golden = load(GOLDEN / "slice_outcome.json")

    # damage_formula golden must match combat rules and the pinned formula.
    if damage_golden["player_damage"] != combat["player_damage"]:
        report.fail("damage_formula vs combat", "player_damage mismatch")
    else:
        report.ok("damage_formula.player_damage matches combat.v0.json")

    pmin = combat["player_damage"]["min"]
    pmax = combat["player_damage"]["max"]
    span = pmax - pmin + 1
    bad = [c for c in damage_golden["cases"] if c["expected_damage"] != pmin + (c["draw"] % span)]
    if bad:
        report.fail("damage_formula cases", f"{len(bad)} case(s) violate min + (draw mod span)")
    else:
        report.ok("damage_formula cases satisfy min + (draw mod span)")

    # monster -> loot table -> item references resolve.
    for mid, mdef in monsters["monsters"].items():
        table = mdef["loot_table"]
        if table not in loot["loot_tables"]:
            report.fail("monster loot_table", f"{mid} -> unknown table {table}")
            continue
        for entry in loot["loot_tables"][table]["entries"]:
            if entry["item_def_id"] not in items["items"]:
                report.fail("loot entry item", f"{table} -> unknown item {entry['item_def_id']}")
                break
        else:
            report.ok(f"monster {mid} loot table + items resolve")

    # loot_roll golden: single-entry table resolves to the expected item.
    table = loot["loot_tables"].get(loot_golden["loot_table"])
    if not table:
        report.fail("loot_roll golden", f"unknown table {loot_golden['loot_table']}")
    elif len(table["entries"]) != 1:
        report.fail("loot_roll golden", "table is not single-entry; golden assumes determinism")
    elif table["entries"][0]["item_def_id"] != loot_golden["expected_item_def_id"]:
        report.fail("loot_roll golden", "expected_item_def_id does not match table entry")
    else:
        report.ok("loot_roll golden matches single-entry table")

    # slice_outcome golden references valid defs.
    if slice_golden["monster_def_id"] not in monsters["monsters"]:
        report.fail("slice_outcome golden", "unknown monster_def_id")
    elif slice_golden["dropped_item_def_id"] not in items["items"]:
        report.fail("slice_outcome golden", "unknown dropped_item_def_id")
    else:
        report.ok("slice_outcome golden references valid defs")

    # item_visuals: every keyed item_def_id exists in items.v0.json with a
    # matching slot, and the visual-resolution golden matches the metadata
    # (spec equip-and-see-it §4.9 #1; rendering contract, not gameplay stats).
    visuals = load(ASSETS / "item_visuals.v0.json")["item_visuals"]
    for def_id, vis in visuals.items():
        item = items["items"].get(def_id)
        if item is None:
            report.fail("item_visuals key", f"{def_id} not in items.v0.json")
        elif item["slot"] != vis["slot"]:
            report.fail("item_visuals slot", f"{def_id}: {vis['slot']} != items slot {item['slot']}")
        else:
            report.ok(f"item_visuals {def_id} resolves to items.v0.json with matching slot")

    visual_golden = load(GOLDEN / "item_visual_resolution.json")
    gdef = visual_golden["item_def_id"]
    gvis = visuals.get(gdef)
    if gvis is None:
        report.fail("item_visual_resolution golden", f"unmapped item_def_id {gdef}")
    elif (visual_golden["expected_asset_id"] != gvis["asset_id"]
          or visual_golden["expected_mount_socket"] != gvis["mount_socket"]
          or visual_golden["expected_slot"] != gvis["slot"]):
        report.fail("item_visual_resolution golden", "expected_* fields disagree with item_visuals metadata")
    else:
        report.ok("item_visual_resolution golden matches item_visuals metadata")


def main() -> int:
    report = Report()
    validate_schemas(report)
    validate_instances(report)
    cross_checks(report)
    print()
    if report.failures:
        print(f"VALIDATION FAILED: {len(report.failures)} problem(s), {report.passed} ok")
        return 1
    print(f"VALIDATION OK: {report.passed} checks passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
