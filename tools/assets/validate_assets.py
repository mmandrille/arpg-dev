#!/usr/bin/env python3
"""Validate the runtime asset pipeline (ADR-0006 D5, fast Python layer).

Engine-free checks over the asset manifest and the shared visual metadata:

  1. The manifest is a valid JSON Schema instance.
  2. Every ``runtime_path`` exists on disk.
  3. Every ``asset_id`` referenced by ``item_visuals`` resolves in the manifest.
  4. Equipment entries declare a ``slot`` matching the visuals that point at
     them, and each visual's ``asset_id`` resolves to an equipment entry.
  5. Rigged character entries declare the hand mount bones that
     ``gear_sockets`` names for the weapon sockets (kit ``handslot.r`` /
     ``handslot.l``). Static character entries may declare no required nodes and
     rely on runtime fallback sockets.
  6. Parse GLB skins and hard-fail unless every declared ``required_nodes``
     name is an actual skin joint. Static entries with no required nodes skip
     this rigged-skin check.
  7. No unmanifested GLB/texture/import files under ``client/assets``.
  8. Every asset fits its type's triangle/texture budget (ADR-0018 D8) or
     carries a named exemption in ``asset_budgets.v0.json``.

Authoritative runtime socket/visibility truth lives in the Godot headless smoke,
not here. Exit code is non-zero if anything fails. Run via ``make validate-assets``.
"""
from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path

from jsonschema import Draft202012Validator

# tools/assets/validate_assets.py -> repo root is parents[2].
ROOT = Path(__file__).resolve().parents[2]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from tools.assets import glb_reader  # noqa: E402
from tools.assets.asset_budgets import check_budgets  # noqa: E402

MANIFEST_REL = "assets/manifests/assets.v0.json"
MANIFEST_SCHEMA_REL = "assets/manifests/assets.v0.schema.json"
ITEM_VISUALS_REL = "shared/assets/item_visuals.v0.json"
MONSTER_VISUALS_REL = "shared/assets/monster_visuals.v0.json"
DUNGEON_KIT_REL = "shared/assets/dungeon_kit_presentation.v0.json"
KIT_MONSTER_REL = "shared/assets/kit_monster_presentation.v0.json"


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
        return glb_reader.skin_joint_name_set(glb_reader.load_gltf(path))
    except Exception:  # noqa: BLE001
        return None


def parse_glb_non_unit_node_scales(path: Path, tolerance: float = 0.01) -> list[tuple[str, list[float]]]:
    """Return [(node_name, scale)] for glTF nodes whose scale is not ~identity.

    An unreadable file reports one sentinel issue so the caller fails loudly.
    """
    try:
        return glb_reader.non_unit_node_scales(glb_reader.load_gltf(path), tolerance)
    except Exception:  # noqa: BLE001
        return [("?", [0.0])]



HAND_SOCKETS = ("right_hand_socket", "off_hand_socket")
GEAR_SOCKETS_REL = "shared/assets/gear_sockets.v0.json"


def hand_mount_bone_options(root: Path) -> dict[str, set[str]]:
    """Per weapon socket: the kit bone that carries it (ADR-0018 D4; P3c removed legacy fallbacks)."""
    path = root / GEAR_SOCKETS_REL
    if not path.is_file():
        return {"right_hand_socket": {"handslot.r"}, "off_hand_socket": {"handslot.l"}}
    sockets = load(path).get("default", {}).get("sockets", {})
    options: dict[str, set[str]] = {}
    for socket in HAND_SOCKETS:
        bone = sockets.get(socket, {}).get("bone", "")
        options[socket] = {bone} if bone else set()
    return options

def sha256_of(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def manifest_client_asset_paths(root: Path, assets: dict) -> set[str]:
    """Paths under client/assets that are allowed: manifest GLBs plus Godot sidecars."""
    allowed: set[str] = set()
    for entry in assets.values():
        rel = entry.get("runtime_path", "")
        if not rel or not rel.startswith("client/assets/"):
            continue
        allowed.add(rel)
        allowed.add(rel + ".import")
        parent = (root / rel).parent
        stem = Path(rel).stem
        if not parent.is_dir():
            continue
        for child in parent.iterdir():
            name = child.name
            if name.startswith(stem + "_") and child.suffix in {".png", ".jpg"}:
                sidecar = str(child.relative_to(root)).replace("\\", "/")
                allowed.add(sidecar)
                sidecar_import = root / f"{sidecar}.import"
                if sidecar_import.is_file():
                    allowed.add(f"{sidecar}.import")

    return allowed


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

    # [3b] imported equipment GLBs must not keep cm-scale glTF node transforms.
    print("[3b] equipment GLB node scales")
    for asset_id, entry in sorted(assets.items()):
        if entry.get("type") != "equipment":
            continue
        runtime_rel = str(entry.get("runtime_path", ""))
        if "/equipment/" not in runtime_rel:
            continue
        origin = str(entry.get("provenance", {}).get("origin", ""))
        if "poly.pizza" not in origin.lower():
            continue
        rt = root / runtime_rel
        if not rt.is_file():
            continue
        issues = parse_glb_non_unit_node_scales(rt)
        if issues:
            report.fail("equipment node scale", f"{asset_id}: non-identity node scales {issues}")
        else:
            report.ok(f"{asset_id} equipment GLB node scales are identity")

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
    # Hand mount bones come from gear_sockets (ADR-0018 D4): each weapon socket's kit `bone` must be
    # declared, so a rig without it can never mount weapons.
    hand_options = hand_mount_bone_options(root)
    for asset_id, entry in sorted(characters.items()):
        declared = set(entry.get("required_nodes", []))
        if not declared:
            report.ok(f"{asset_id} declares static character fallback sockets")
            continue
        missing = [socket for socket, options in hand_options.items() if not (options & declared)]
        if missing:
            report.fail(
                "mount bone",
                f"{asset_id}: required_nodes cover no bone for {missing} (options {hand_options})",
            )
        else:
            report.ok(f"{asset_id} declares hand mount bones for {sorted(hand_options)}")

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

    # [7] Reject unmanifested client/assets GLB/texture/import sidecars so local
    # Godot imports cannot accumulate outside the asset pipeline.
    print("[7] client/assets orphan check")
    allowed_paths = manifest_client_asset_paths(root, assets)
    client_assets = root / "client" / "assets"
    orphans: list[str] = []
    if client_assets.is_dir():
        for path in sorted(client_assets.rglob("*")):
            if not path.is_file() or path.suffix not in {".glb", ".png", ".jpg", ".import"}:
                continue
            rel = str(path.relative_to(root)).replace("\\", "/")
            if rel not in allowed_paths:
                orphans.append(rel)
    if orphans:
        for rel in orphans:
            report.fail("orphan client asset", rel)
    else:
        report.ok("no orphan client/assets GLB or import sidecars")

    # [8] ADR-0018 D8 triangle/texture budgets per asset type (data: asset_budgets.v0.json).
    print("[8] asset budgets")
    check_budgets(root, assets, report)

    # [9] ADR-0018 P2 dungeon kit catalog ids resolve to environment assets.
    kit_path = root / DUNGEON_KIT_REL
    if kit_path.is_file():
        print("[9] dungeon kit asset ids")
        kit = load(kit_path)
        kit_ids = [kit["wall"]["full_asset_id"], kit["wall"]["half_asset_id"], kit["column"]["asset_id"]]
        kit_ids += [v["asset_id"] for v in kit["floor"]["variants"]]
        kit_ids += [kit["torch"]["asset_id"], kit["chest"]["asset_id"], kit["chest"]["elite_objective_asset_id"]]
        if "stairs" in kit:  # v489 kit stairs
            kit_ids += [kit["stairs"]["up_asset_id"], kit["stairs"]["down_asset_id"]]
        if "dressing" in kit:  # v494 room props
            kit_ids += [p["asset_id"] for p in kit["dressing"].get("props", [])]
        town_path = root / "shared/assets/town_presentation.v0.json"
        if town_path.is_file():  # v491 town dressing props
            dressing = load(town_path).get("dressing", {})
            kit_ids += sorted({p["asset_id"] for p in dressing.get("props", [])})
            plaza = dressing.get("plaza", {})
            kit_ids += [v["asset_id"] for v in plaza.get("tile_variants", [])]
            kit_ids += [v["asset_id"] for v in plaza.get("rim", {}).get("tile_variants", [])]  # v493
            kit_ids += [v["asset_id"] for v in dressing.get("edge", {}).get("tile_variants", [])]
            scatter = dressing.get("scatter", {})
            kit_ids += [v["asset_id"] for v in scatter.get("patches", []) + scatter.get("rocks", [])]
        for asset_id in kit_ids:
            entry = assets.get(asset_id)
            if entry is None or entry.get("type") != "environment":
                report.fail("dungeon kit asset", f"{asset_id}: not an environment asset in the manifest")
            else:
                report.ok(f"dungeon kit {asset_id} resolves")

    # [10] ADR-0018 P4a kit monster scenes and attachments resolve to monster assets.
    kit_monster_path = root / KIT_MONSTER_REL
    if kit_monster_path.is_file():
        print("[10] kit monster asset ids")
        kit_monsters = load(kit_monster_path)
        for scene_key, spec in sorted(kit_monsters["monsters"].items()):
            if spec["clip_profile"] not in kit_monsters["clip_profiles"]:
                report.fail("kit monster clip profile", f"{scene_key}: unknown profile {spec['clip_profile']}")
            for asset_id in [spec["asset_id"]] + [a["asset_id"] for a in spec["attachments"]]:
                entry = assets.get(asset_id)
                if entry is None or entry.get("type") != "monster":
                    report.fail("kit monster asset", f"{scene_key}: {asset_id} is not a monster asset")
                else:
                    report.ok(f"kit monster {scene_key} -> {asset_id} resolves")
        # v513: every clip id a profile names must exist among the animations embedded in the
        # GLB of each monster using that profile (runtime only warns on a missing clip).
        for problem in kit_clip_problems(root, kit_monsters, assets):
            report.fail("kit monster clip", problem)
        report.ok("kit monster clip profiles resolve against embedded GLB clips")


def profile_clip_ids(profile: dict) -> list[tuple[str, str]]:
    """(logical, kit_clip) pairs a clip profile asks the rig for (v513)."""
    return sorted((str(k), str(v)) for k, v in profile.get("clips", {}).items())


def kit_clip_problems(root: Path, kit_monsters: dict, assets: dict) -> list[str]:
    problems: list[str] = []
    profiles = kit_monsters.get("clip_profiles", {})
    clip_cache: dict[str, set[str]] = {}
    for scene_key, spec in sorted(kit_monsters.get("monsters", {}).items()):
        profile = profiles.get(spec.get("clip_profile", ""))
        entry = assets.get(spec.get("asset_id", ""))
        if profile is None or entry is None or "runtime_path" not in entry:
            continue  # reported by the checks above
        rel = entry["runtime_path"]
        if rel not in clip_cache:
            glb = root / rel
            clip_cache[rel] = (
                {a["name"] for a in glb_reader.animation_summaries(glb_reader.load_gltf(glb))}
                if glb.is_file() else set()
            )
        embedded = clip_cache[rel]
        logicals = {k for k, _ in profile_clip_ids(profile)}
        for logical, clip in profile_clip_ids(profile):
            if clip not in embedded:
                problems.append(f"{scene_key}: logical clip {logical} -> {clip} not embedded in {rel}")
        for group, members in sorted(profile.get("variants", {}).items()):
            for member in members:
                if member not in logicals:
                    problems.append(f"{scene_key}: variants.{group} member {member} is not a profile clip")
        refs = list(profile.get("hit_directional", {}).values())
        refs += list(profile.get("attack_contact", {}).keys())
        if "clip" in profile.get("spawn", {}):
            refs.append(profile["spawn"]["clip"])
        for ref in refs:
            if ref not in logicals:
                problems.append(f"{scene_key}: {ref} is referenced but not a profile clip")
    return problems


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
