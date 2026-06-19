#!/usr/bin/env python3
"""Validate the runtime asset pipeline (ADR-0006 D5, fast Python layer).

Engine-free checks over the asset manifest and the shared visual metadata:

  1. The manifest is a valid JSON Schema instance.
  2. Every ``runtime_path`` exists on disk.
  3. Every ``asset_id`` referenced by ``item_visuals`` resolves in the manifest.
  4. Equipment entries declare a ``slot`` matching the visuals that point at
     them, and each visual's ``asset_id`` resolves to an equipment entry.
  5. Rigged character entries declare hand mount bones (``hand_r`` and
     ``hand_l``). Static character entries may declare no required nodes and
     rely on runtime fallback sockets.
  6. Parse GLB skins and hard-fail unless every declared ``required_nodes``
     name is an actual skin joint. Static entries with no required nodes skip
     this rigged-skin check.

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
MONSTER_VISUALS_REL = "shared/assets/monster_visuals.v0.json"


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


def parse_glb_skin_joint_names(path: Path) -> set[str] | None:
    """Return the set of node names referenced by any skin's `joints`, or None.

    A required bone must be an actual skin joint, not merely a named node — this
    is what proves the GLB is skinned (spec §6), not a v2 socket placeholder.
    """
    try:
        data = path.read_bytes()
        if len(data) < 20 or data[0:4] != b"glTF":
            return None
        chunk_len, chunk_type = struct.unpack_from("<II", data, 12)
        if chunk_type != 0x4E4F534A:  # 'JSON'
            return None
        gltf = json.loads(data[20 : 20 + chunk_len].decode("utf-8"))
        nodes = gltf.get("nodes", [])
        joint_idx: set[int] = set()
        for skin in gltf.get("skins", []):
            joint_idx.update(skin.get("joints", []))
        return {nodes[i]["name"] for i in joint_idx if i < len(nodes) and "name" in nodes[i]}
    except Exception:  # noqa: BLE001
        return None


def sha256_of(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def validate(root: Path, report: Report) -> None:
    manifest_path = root / MANIFEST_REL
    schema_path = root / MANIFEST_SCHEMA_REL
    visuals_path = root / ITEM_VISUALS_REL
    monster_visuals_path = root / MONSTER_VISUALS_REL

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
    monster_visuals = load(monster_visuals_path)["monster_visuals"] if monster_visuals_path.is_file() else {}

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
    for def_id, vis in sorted(visuals.items()):
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

    # [4b] monster_visuals -> manifest resolution. Monster presentation is
    # data-driven by monster_def_id, but the manifest remains the source of
    # truth for runtime bytes.
    print("[4b] monster_visuals -> manifest resolution")
    for def_id, vis in sorted(monster_visuals.items()):
        entry = assets.get(vis["asset_id"])
        if entry is None:
            report.fail("monster asset_id resolution", f"{def_id}: asset_id {vis['asset_id']} not in manifest")
            continue
        if entry["type"] != "monster":
            report.fail("monster asset_id type", f"{def_id}: asset {vis['asset_id']} is {entry['type']}, expected monster")
        else:
            report.ok(f"{def_id} -> {vis['asset_id']} resolves to monster asset")

    # [5] character mount-bone coverage (spec §4.3): item_visuals names runtime
    #     hand sockets. Rigged character assets satisfy that with hand bones;
    #     explicitly static character assets rely on root-relative fallback
    #     sockets created by the Godot character visual.
    print("[5] character mount-bone coverage")
    characters = {aid: e for aid, e in assets.items() if e["type"] == "character"}
    if not characters:
        report.fail("character coverage", "no character asset declared")
    HAND_MOUNT_BONES = {"hand_r", "hand_l"}
    for asset_id, entry in sorted(characters.items()):
        declared = set(entry.get("required_nodes", []))
        if not declared:
            report.ok(f"{asset_id} declares static character fallback sockets")
            continue
        missing = sorted(HAND_MOUNT_BONES - declared)
        if missing:
            report.fail(
                "mount bone",
                f"{asset_id}: required_nodes missing hand mount bones {missing}",
            )
        else:
            report.ok(f"{asset_id} declares hand mount bones {sorted(HAND_MOUNT_BONES)}")

    # [6] GLB skin-joint inspection: required_nodes must be SKIN JOINTS, proving
    #     the GLB is actually rigged (spec §6, §10). Characters/monsters are
    #     skinned; equipment (the sword) is static and declares no required_nodes.
    print("[6] GLB skin-joint inspection")
    for asset_id, entry in sorted(assets.items()):
        required = entry.get("required_nodes", [])
        if not required:
            continue
        rt = root / entry["runtime_path"]
        if not rt.is_file():
            continue  # already failed in [2]
        joints = parse_glb_skin_joint_names(rt)
        if joints is None:
            report.fail("glb skin", f"{asset_id}: could not parse GLB skin joints")
            continue
        absent = [n for n in required if n not in joints]
        if absent:
            report.fail("glb joint", f"{asset_id}: required_nodes not skin joints: {absent}")
        else:
            report.ok(f"{asset_id} GLB skin includes joints {required}")


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
