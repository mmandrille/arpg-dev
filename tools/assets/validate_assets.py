#!/usr/bin/env python3
"""Validate the runtime asset pipeline (ADR-0006 D5, fast Python layer).

Engine-free checks over the asset manifest and the shared visual metadata:

  1. The manifest is a valid JSON Schema instance.
  2. Every ``runtime_path`` exists on disk.
  3. Every ``asset_id`` referenced by ``item_visuals`` resolves in the manifest.
  4. Every character entry's ``required_nodes`` covers all mount sockets the
     item visuals reference, and equipment entries declare a ``slot`` matching
     the visuals that point at them.
  5. When ``provenance.sha256`` is present it matches the committed file.
  6. Best-effort: parse the GLB JSON chunk and confirm declared ``required_nodes``
     names exist (warn-only if the file can't be parsed; hard-fail on an
     explicit name mismatch).

Authoritative runtime socket/visibility truth lives in the Godot headless smoke,
not here. Exit code is non-zero if anything fails. Run via ``make validate-assets``.
"""
from __future__ import annotations

import hashlib
import json
import struct
import sys
from pathlib import Path

from jsonschema import Draft202012Validator

# tools/assets/validate_assets.py -> repo root is parents[2].
ROOT = Path(__file__).resolve().parents[2]

MANIFEST_REL = "assets/manifests/assets.v0.json"
MANIFEST_SCHEMA_REL = "assets/manifests/assets.v0.schema.json"
ITEM_VISUALS_REL = "shared/assets/item_visuals.v0.json"


def load(path: Path):
    with path.open(encoding="utf-8") as fh:
        return json.load(fh)


class Report:
    def __init__(self) -> None:
        self.passed = 0
        self.failures: list[str] = []
        self.warnings: list[str] = []

    def ok(self, label: str) -> None:
        self.passed += 1
        print(f"  ok   {label}")

    def warn(self, label: str, detail: str) -> None:
        self.warnings.append(f"{label}: {detail}")
        print(f"  warn {label}: {detail}")

    def fail(self, label: str, detail: str) -> None:
        self.failures.append(f"{label}: {detail}")
        print(f"  FAIL {label}: {detail}")


def parse_glb_node_names(path: Path) -> set[str] | None:
    """Return the set of node names in a .glb JSON chunk, or None if unparseable.

    glTF-binary layout: 12-byte header ('glTF', version, total length) followed
    by length-prefixed chunks; the first chunk is JSON.
    """
    try:
        data = path.read_bytes()
        if len(data) < 20 or data[0:4] != b"glTF":
            return None
        # Skip the 12-byte header (magic, version, total length); the first
        # chunk header (length, type) sits at offset 12, its payload at 20.
        chunk_len, chunk_type = struct.unpack_from("<II", data, 12)
        if chunk_type != 0x4E4F534A:  # 'JSON'
            return None
        chunk = data[20 : 20 + chunk_len]
        gltf = json.loads(chunk.decode("utf-8"))
        return {n["name"] for n in gltf.get("nodes", []) if "name" in n}
    except Exception:  # noqa: BLE001 - any read/parse error -> "unknown", warn upstream
        return None


def sha256_of(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def validate(root: Path, report: Report) -> None:
    manifest_path = root / MANIFEST_REL
    schema_path = root / MANIFEST_SCHEMA_REL
    visuals_path = root / ITEM_VISUALS_REL

    # [1] schema-validate the manifest.
    print("[1] manifest schema validation")
    schema = load(schema_path)
    Draft202012Validator.check_schema(schema)
    manifest = load(manifest_path)
    errors = sorted(Draft202012Validator(schema).iter_errors(manifest), key=lambda e: list(e.path))
    if errors:
        first = errors[0]
        loc = "/".join(str(p) for p in first.path) or "<root>"
        report.fail("manifest schema", f"at {loc}: {first.message}")
        return  # downstream checks assume a schema-valid manifest
    report.ok("assets.v0.json validates against schema")

    assets = manifest["assets"]
    visuals = load(visuals_path)["item_visuals"]

    # [2] runtime_path existence + provenance sha256.
    print("[2] runtime files + provenance")
    for asset_id, entry in sorted(assets.items()):
        rt = root / entry["runtime_path"]
        if not rt.is_file():
            report.fail("runtime_path", f"{asset_id}: missing {entry['runtime_path']}")
            continue
        report.ok(f"{asset_id} runtime file exists")
        prov = entry.get("provenance")
        if prov and "sha256" in prov:
            actual = sha256_of(rt)
            if actual != prov["sha256"]:
                report.fail("provenance.sha256", f"{asset_id}: {actual} != manifest {prov['sha256']}")
            else:
                report.ok(f"{asset_id} sha256 matches provenance")

    # [3] equipment entries declare a slot.
    print("[3] equipment slot declarations")
    for asset_id, entry in sorted(assets.items()):
        if entry["type"] == "equipment" and "slot" not in entry:
            report.fail("equipment slot", f"{asset_id}: equipment entry missing slot")
        elif entry["type"] == "equipment":
            report.ok(f"{asset_id} declares slot {entry['slot']}")

    # [4] visual->manifest resolution + slot agreement (spec §4.9 #2, #4).
    print("[4] item_visuals -> manifest resolution")
    referenced_sockets: set[str] = set()
    for def_id, vis in sorted(visuals.items()):
        referenced_sockets.add(vis["mount_socket"])
        entry = assets.get(vis["asset_id"])
        if entry is None:
            report.fail("asset_id resolution", f"{def_id}: asset_id {vis['asset_id']} not in manifest")
            continue
        if entry["type"] != "equipment":
            report.fail("asset_id type", f"{def_id}: asset {vis['asset_id']} is {entry['type']}, expected equipment")
        elif entry.get("slot") != vis["slot"]:
            report.fail("slot agreement", f"{def_id}: visual slot {vis['slot']} != asset slot {entry.get('slot')}")
        else:
            report.ok(f"{def_id} -> {vis['asset_id']} resolves with matching slot")

    # [5] character required_nodes cover every referenced mount socket (spec §4.9 #3).
    print("[5] character socket coverage")
    characters = {aid: e for aid, e in assets.items() if e["type"] == "character"}
    if not characters:
        report.fail("character coverage", "no character asset declared")
    for asset_id, entry in sorted(characters.items()):
        declared = set(entry.get("required_nodes", []))
        missing = referenced_sockets - declared
        if missing:
            report.fail("required_nodes", f"{asset_id}: missing socket(s) {sorted(missing)}")
        else:
            report.ok(f"{asset_id} declares all referenced sockets {sorted(referenced_sockets)}")

    # [6] best-effort GLB node-name inspection for declared required_nodes.
    print("[6] GLB node-name inspection (best-effort)")
    for asset_id, entry in sorted(assets.items()):
        required = entry.get("required_nodes", [])
        if not required:
            continue
        rt = root / entry["runtime_path"]
        if not rt.is_file():
            continue  # already failed in [2]
        names = parse_glb_node_names(rt)
        if names is None:
            report.warn("glb inspect", f"{asset_id}: could not parse GLB nodes (skipped name check)")
            continue
        absent = [n for n in required if n not in names]
        if absent:
            report.fail("glb node", f"{asset_id}: required_nodes absent in GLB: {absent}")
        else:
            report.ok(f"{asset_id} GLB contains required nodes {required}")


def main() -> int:
    report = Report()
    validate(ROOT, report)
    print()
    if report.warnings:
        print(f"({len(report.warnings)} warning(s))")
    if report.failures:
        print(f"ASSET VALIDATION FAILED: {len(report.failures)} problem(s), {report.passed} ok")
        return 1
    print(f"ASSET VALIDATION OK: {report.passed} checks passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
