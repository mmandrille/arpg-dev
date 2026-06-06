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
            return PROTOCOL / "session_snapshot.v1.schema.json"
        if name.startswith("state_delta"):
            return PROTOCOL / "state_delta.v1.schema.json"
        return PROTOCOL / "messages.v1.schema.json"
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
    interactables = load(RULES / "interactables.v0.json")
    navigation = load(RULES / "navigation.v0.json")
    worlds = load(RULES / "worlds.v0.json")
    dungeon_generation = load(RULES / "dungeon_generation.v0.json")
    damage_golden = load(GOLDEN / "damage_formula.json")
    retaliation_golden = load(GOLDEN / "retaliation_damage.json")
    equipped_weapon_golden = load(GOLDEN / "equipped_weapon_damage.json")
    loot_golden = load(GOLDEN / "loot_roll.json")
    slice_golden = load(GOLDEN / "slice_outcome.json")
    auto_path_golden = load(GOLDEN / "auto_path.json")
    ranged_projectile_golden = load(GOLDEN / "ranged_projectile.json")
    inventory_drop_golden = load(GOLDEN / "inventory_drop.json")
    use_consumable_golden = load(GOLDEN / "use_consumable.json")
    monster_chase_golden = load(GOLDEN / "monster_chase.json")

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

    # retaliation_damage golden must match the training dummy range and formula.
    dummy = monsters["monsters"].get("training_dummy")
    retaliation = dummy.get("retaliation_damage") if dummy else None
    if retaliation is None:
        report.fail("training_dummy retaliation", "missing retaliation_damage")
    elif retaliation["max"] < retaliation["min"]:
        report.fail("training_dummy retaliation", "max must be >= min")
    else:
        report.ok("training_dummy retaliation range is configured")

    for mid, mdef in monsters["monsters"].items():
        rd = mdef.get("retaliation_damage")
        if rd is not None and rd["max"] < rd["min"]:
            report.fail("monster retaliation range", f"{mid}: max must be >= min")
        behavior = mdef.get("behavior", "static")
        if behavior not in ("static", "chase"):
            report.fail("monster behavior", f"{mid}: invalid behavior {behavior}")
            continue
        if behavior == "static":
            if mdef.get("aggro_radius") or mdef.get("leash_radius") or mdef.get("move_speed"):
                report.fail("monster behavior fields", f"{mid}: aggro/leash/move_speed only valid for chase")
            else:
                report.ok(f"monster {mid} static behavior is valid")
            continue
        aggro = mdef.get("aggro_radius")
        if not isinstance(aggro, (int, float)) or aggro <= 0:
            report.fail("monster aggro_radius", f"{mid}: chase requires positive aggro_radius")
            continue
        leash = mdef.get("leash_radius")
        if leash is not None:
            if not isinstance(leash, (int, float)) or leash < aggro:
                report.fail("monster leash_radius", f"{mid}: leash_radius must be >= aggro_radius")
                continue
        move_speed = mdef.get("move_speed", navigation["cell_size"])
        if move_speed != navigation["cell_size"]:
            report.fail("monster move_speed", f"{mid}: move_speed must equal navigation.cell_size in v17")
        else:
            report.ok(f"monster {mid} chase behavior is valid")

    if retaliation is not None and retaliation_golden["retaliation_damage"] != retaliation:
        report.fail("retaliation_damage vs monster", "retaliation_damage mismatch")
    else:
        report.ok("retaliation_damage golden matches training_dummy")

    rmin = retaliation_golden["retaliation_damage"]["min"]
    rmax = retaliation_golden["retaliation_damage"]["max"]
    rspan = rmax - rmin + 1
    if rspan <= 0:
        report.fail("retaliation_damage golden", "max must be >= min")
    else:
        rbad = [c for c in retaliation_golden["cases"] if c["expected_damage"] != rmin + (c["draw"] % rspan)]
        if rbad:
            report.fail("retaliation_damage cases", f"{len(rbad)} case(s) violate min + (draw mod span)")
        else:
            report.ok("retaliation_damage cases satisfy min + (draw mod span)")

    # item damage is weapon-only and the equipped_weapon_damage golden must
    # mirror the referenced item definition.
    for item_id, item in items["items"].items():
        dmg = item.get("damage")
        reach = item.get("reach")
        if dmg is not None:
            if not item.get("equippable") or item.get("slot") != "weapon":
                report.fail("item damage eligibility", f"{item_id}: damage is only valid on equippable weapons")
                continue
            if dmg["max"] < dmg["min"]:
                report.fail("item damage range", f"{item_id}: max must be >= min")
            else:
                report.ok(f"item {item_id} weapon damage range is valid")
        if reach is not None:
            if not item.get("equippable") or item.get("slot") != "weapon":
                report.fail("item reach eligibility", f"{item_id}: reach is only valid on equippable weapons")
                continue
            if reach <= 0:
                report.fail("item reach", f"{item_id}: reach must be positive")
            else:
                report.ok(f"item {item_id} weapon reach is valid")
        attack_mode = item.get("attack_mode", "melee")
        projectile_speed = item.get("projectile_speed")
        if attack_mode == "ranged":
            if not item.get("equippable") or item.get("slot") != "weapon":
                report.fail("item ranged eligibility", f"{item_id}: ranged mode is only valid on equippable weapons")
            elif dmg is None or reach is None or projectile_speed is None:
                report.fail("item ranged fields", f"{item_id}: ranged weapon requires damage, reach, and projectile_speed")
            elif projectile_speed <= 0:
                report.fail("item projectile_speed", f"{item_id}: projectile_speed must be positive")
            else:
                report.ok(f"item {item_id} ranged weapon fields are valid")
        elif projectile_speed is not None:
            report.fail("item projectile_speed", f"{item_id}: projectile_speed is only valid on ranged weapons")

    if combat.get("unarmed_reach", 0) <= 0:
        report.fail("combat unarmed_reach", "must be positive")
    else:
        report.ok("combat unarmed_reach is positive")

    if navigation.get("cell_size") != 1.0:
        report.fail("navigation cell_size", "must be 1.0 for v11 moveSpeed parity")
    elif navigation.get("max_auto_steps", 0) <= 0:
        report.fail("navigation max_auto_steps", "must be positive")
    elif navigation.get("stop_distance", -1) < 0:
        report.fail("navigation stop_distance", "must be non-negative")
    else:
        report.ok("navigation rules are within v11 bounds")

    floor_size = dungeon_generation["floor_size"]
    if floor_size.get("width", 0) < 16 or floor_size.get("height", 0) < 10:
        report.fail("dungeon_generation floor_size", "must be at least 16x10")
    else:
        report.ok("dungeon_generation floor_size is at least legacy arena size")
    teleporter_placement = dungeon_generation.get("teleporter_placement", {})
    if teleporter_placement.get("margin_from_wall", -1) < 0:
        report.fail("dungeon_generation teleporter_placement", "margin_from_wall must be non-negative")
    elif teleporter_placement.get("min_stair_distance", 0) <= 0:
        report.fail("dungeon_generation teleporter_placement", "min_stair_distance must be positive")
    elif teleporter_placement.get("max_attempts", 0) <= 0:
        report.fail("dungeon_generation teleporter_placement", "max_attempts must be positive")
    else:
        report.ok("dungeon_generation teleporter placement is valid")
    for key in dungeon_generation["level_names"]:
        try:
            level_num = int(key)
        except ValueError:
            report.fail("dungeon_generation level_names", f"{key}: must be an integer string")
            continue
        if level_num >= 0:
            report.fail("dungeon_generation level_names", f"{key}: must be a negative level")
        else:
            report.ok(f"dungeon_generation level name {key} is negative")

    golden_item_id = equipped_weapon_golden["item_def_id"]
    golden_item = items["items"].get(golden_item_id)
    if golden_item is None:
        report.fail("equipped_weapon_damage golden", f"unknown item_def_id {golden_item_id}")
    elif not golden_item.get("equippable") or golden_item.get("slot") != "weapon":
        report.fail("equipped_weapon_damage golden", f"{golden_item_id} is not an equippable weapon")
    elif golden_item.get("damage") != equipped_weapon_golden["damage"]:
        report.fail("equipped_weapon_damage golden", "damage range mismatch with item rules")
    else:
        report.ok("equipped_weapon_damage golden matches weapon item rules")

    ew_min = equipped_weapon_golden["damage"]["min"]
    ew_max = equipped_weapon_golden["damage"]["max"]
    ew_span = ew_max - ew_min + 1
    if ew_span <= 0:
        report.fail("equipped_weapon_damage golden", "max must be >= min")
    else:
        ew_bad = [
            c for c in equipped_weapon_golden["cases"]
            if c["expected_damage"] != ew_min + (c["draw"] % ew_span)
        ]
        if ew_bad:
            report.fail("equipped_weapon_damage cases", f"{len(ew_bad)} case(s) violate min + (draw mod span)")
        else:
            report.ok("equipped_weapon_damage cases satisfy min + (draw mod span)")

    # monster -> loot table -> item references resolve.
    for mid, mdef in monsters["monsters"].items():
        table = mdef["loot_table"]
        if table not in loot["loot_tables"]:
            report.fail("monster loot_table", f"{mid} -> unknown table {table}")
            continue
        for entry in loot["loot_tables"][table].get("entries", []):
            if entry["item_def_id"] not in items["items"]:
                report.fail("loot entry item", f"{table} -> unknown item {entry['item_def_id']}")
                break
        else:
            for item_id in loot["loot_tables"][table].get("drops", []):
                if item_id not in items["items"]:
                    report.fail("loot drop item", f"{table} -> unknown item {item_id}")
                    break
            else:
                report.ok(f"monster {mid} loot table + items resolve")

    # world presets: entity references resolve and type-specific fields are present.
    for world_id, world in worlds["worlds"].items():
        mode = world.get("mode")
        if mode is None:
            report.ok(f"world {world_id} uses single-level mode")
        elif mode == "multi_level":
            report.ok(f"world {world_id} uses multi-level mode")
        else:
            report.fail("world mode", f"{world_id}: unsupported mode {mode}")

        pos = world["player"]["position"]
        if not isinstance(pos.get("x"), (int, float)) or not isinstance(pos.get("y"), (int, float)):
            report.fail("world player position", f"{world_id}: player.position must have numeric x/y")
        else:
            report.ok(f"world {world_id} player position is numeric")

        for idx, entity in enumerate(world["entities"]):
            etype = entity.get("type")
            label = f"world {world_id} entity {idx}"
            if etype == "monster":
                monster_id = entity.get("monster_def_id")
                if not monster_id:
                    report.fail("world monster entity", f"{label}: missing monster_def_id")
                elif monster_id not in monsters["monsters"]:
                    report.fail("world monster entity", f"{label}: unknown monster {monster_id}")
                else:
                    report.ok(f"{label} monster reference resolves")
            elif etype == "loot":
                item_id = entity.get("item_def_id")
                if not item_id:
                    report.fail("world loot entity", f"{label}: missing item_def_id")
                elif item_id not in items["items"]:
                    report.fail("world loot entity", f"{label}: unknown item {item_id}")
                else:
                    report.ok(f"{label} loot item reference resolves")
            elif etype == "wall":
                size = entity.get("size", {})
                if not isinstance(size.get("x"), (int, float)) or not isinstance(size.get("y"), (int, float)):
                    report.fail("world wall entity", f"{label}: wall size must have numeric x/y")
                elif size["x"] <= 0 or size["y"] <= 0:
                    report.fail("world wall entity", f"{label}: wall size must be positive")
                else:
                    report.ok(f"{label} wall size is positive")
            elif etype == "interactable":
                interactable_id = entity.get("interactable_def_id")
                if not interactable_id:
                    report.fail("world interactable entity", f"{label}: missing interactable_def_id")
                elif interactable_id not in interactables["interactables"]:
                    report.fail("world interactable entity", f"{label}: unknown interactable {interactable_id}")
                else:
                    report.ok(f"{label} interactable reference resolves")
            else:
                report.fail("world entity type", f"{label}: unknown type {etype}")

    if auto_path_golden["navigation"] != navigation:
        report.fail("auto_path navigation", "golden navigation block must match navigation.v0.json")
    else:
        report.ok("auto_path golden navigation matches navigation.v0.json")

    for case in auto_path_golden["cases"]:
        world_id = case["world_id"]
        if world_id not in worlds["worlds"]:
            report.fail("auto_path world", f"{case['name']}: unknown world_id {world_id}")
        elif case.get("goal_mode") != "melee_approach" or case.get("target_kind") != "monster":
            report.fail("auto_path goal", f"{case['name']}: must use melee_approach on monster")
        else:
            report.ok(f"auto_path {case['name']} references world and goal mode")

    constants = ranged_projectile_golden["constants"]
    if constants.get("projectile_radius") != 0.10:
        report.fail("ranged_projectile projectile_radius", "must match v12 projectileRadius 0.10")
    elif constants.get("tick_duration") != 0.05:
        report.fail("ranged_projectile tick_duration", "must match 20 Hz tick duration 0.05")
    elif constants.get("monster_radius") != 0.45:
        report.fail("ranged_projectile monster_radius", "must match server monsterRadius 0.45")
    else:
        report.ok("ranged_projectile constants match v12 sim constants")

    for case in ranged_projectile_golden["cases"]:
        name = case["name"]
        world_id = case["world_id"]
        weapon_id = case["equipped_weapon"]
        monster_id = case["target_monster_def_id"]
        if world_id not in worlds["worlds"]:
            report.fail("ranged_projectile world", f"{name}: unknown world_id {world_id}")
        elif weapon_id not in items["items"]:
            report.fail("ranged_projectile weapon", f"{name}: unknown equipped_weapon {weapon_id}")
        elif items["items"][weapon_id].get("attack_mode") != "ranged":
            report.fail("ranged_projectile weapon", f"{name}: {weapon_id} is not ranged")
        elif monster_id not in monsters["monsters"]:
            report.fail("ranged_projectile monster", f"{name}: unknown target_monster_def_id {monster_id}")
        elif case.get("expected_event") not in (None, "projectile_blocked", "projectile_expired", "attack_missed", "monster_killed"):
            report.fail("ranged_projectile event", f"{name}: unexpected expected_event {case.get('expected_event')}")
        else:
            report.ok(f"ranged_projectile {name} references valid rules")

    inventory_world_id = inventory_drop_golden["world_id"]
    inventory_item_id = inventory_drop_golden["item_def_id"]
    inventory_constants = inventory_drop_golden["constants"]
    if inventory_world_id not in worlds["worlds"]:
        report.fail("inventory_drop world", f"unknown world_id {inventory_world_id}")
    elif inventory_item_id not in items["items"]:
        report.fail("inventory_drop item", f"unknown item_def_id {inventory_item_id}")
    elif inventory_constants.get("loot_drop_radius") != 0.35:
        report.fail("inventory_drop loot_drop_radius", "must match v13 loot drop radius 0.35")
    elif inventory_constants.get("player_radius") != 0.45:
        report.fail("inventory_drop player_radius", "must match server playerRadius 0.45")
    elif inventory_constants.get("drop_step") != navigation.get("cell_size"):
        report.fail("inventory_drop drop_step", "must match navigation.cell_size")
    else:
        report.ok("inventory_drop golden references valid rules and constants")

    use_item_id = use_consumable_golden["item_def_id"]
    use_heal = use_consumable_golden["heal"]
    use_item = items["items"].get(use_item_id)
    if use_item is None:
        report.fail("use_consumable item", f"unknown item_def_id {use_item_id}")
    elif use_item.get("category") != "consumable":
        report.fail("use_consumable item", f"{use_item_id} is not consumable")
    elif use_item.get("heal") != use_heal:
        report.fail("use_consumable item", "heal range mismatch with item rules")
    else:
        report.ok("use_consumable golden matches consumable item rules")
    umin = int(use_heal["min"])
    umax = int(use_heal["max"])
    if umax < umin:
        report.fail("use_consumable heal", "max must be >= min")
    else:
        uspan = umax - umin + 1
        bad_use = []
        for case in use_consumable_golden["cases"]:
            rolled = umin + (int(case["draw"]) % uspan)
            capped = min(rolled, int(case["player_max_hp"]) - int(case["player_hp"]))
            if capped != int(case["expected_heal"]) or int(case["player_hp"]) + capped != int(case["expected_player_hp"]):
                bad_use.append(case["name"])
        if bad_use:
            report.fail("use_consumable cases", f"{len(bad_use)} case(s) violate heal cap formula: {bad_use}")
        else:
            report.ok("use_consumable cases satisfy heal roll + HP cap")

    if monster_chase_golden["navigation"] != navigation:
        report.fail("monster_chase navigation", "navigation block mismatch")
    else:
        report.ok("monster_chase.navigation matches navigation.v0.json")

    chase_worlds = {"chase_lab", "chase_maze", "leash_lab"}
    for case in monster_chase_golden["cases"]:
        world_id = case.get("world_id", monster_chase_golden.get("world_id"))
        if world_id not in worlds["worlds"]:
            report.fail("monster_chase world", f"{case['name']}: unknown world_id {world_id}")
            continue
        report.ok(f"monster_chase case {case['name']} references world {world_id}")
        for entity in worlds["worlds"][world_id]["entities"]:
            if entity.get("type") != "monster":
                continue
            monster_id = entity["monster_def_id"]
            if monsters["monsters"][monster_id].get("behavior") != "chase":
                report.fail("monster_chase world monster", f"{world_id}: {monster_id} must be chase")
            elif world_id in chase_worlds and monster_id != "training_dummy_chase":
                report.fail("monster_chase lab monster", f"{world_id}: expected training_dummy_chase")
            else:
                report.ok(f"monster_chase world {world_id} uses chase monster {monster_id}")

    for interactable_id, interactable in interactables["interactables"].items():
        initial_state = interactable.get("initial_state")
        if initial_state == "closed":
            report.ok(f"interactable {interactable_id} initial_state is closed")
            size = interactable.get("barrier_when_closed", {}).get("size", {})
            if not isinstance(size.get("x"), (int, float)) or not isinstance(size.get("y"), (int, float)):
                report.fail("interactable barrier", f"{interactable_id}: size must have numeric x/y")
            elif size["x"] <= 0 or size["y"] <= 0:
                report.fail("interactable barrier", f"{interactable_id}: size must be positive")
            else:
                report.ok(f"interactable {interactable_id} barrier size is positive")
            if "transition" in interactable:
                report.fail("interactable transition", f"{interactable_id}: closed blocker must not declare transition")
            continue
        if initial_state == "ready":
            transition = interactable.get("transition")
            if transition not in ("ascend", "descend", "waypoint"):
                report.fail("interactable transition", f"{interactable_id}: ready interactable needs ascend/descend/waypoint")
            elif "barrier_when_closed" in interactable:
                report.fail("interactable barrier", f"{interactable_id}: ready interactable must not block")
            else:
                report.ok(f"interactable {interactable_id} ready transition is {transition}")
            continue
        report.fail("interactable initial_state", f"{interactable_id}: unsupported state {initial_state}")

    for stair_id, expected_transition in {"stairs_down": "descend", "stairs_up": "ascend"}.items():
        stair = interactables["interactables"].get(stair_id)
        if stair is None:
            report.fail("stair interactable", f"missing {stair_id}")
        elif stair.get("initial_state") != "ready" or stair.get("transition") != expected_transition:
            report.fail("stair interactable", f"{stair_id}: expected ready/{expected_transition}")
        else:
            report.ok(f"stair interactable {stair_id} is ready/{expected_transition}")
    teleporter = interactables["interactables"].get("teleporter")
    if teleporter is None:
        report.fail("teleporter interactable", "missing teleporter")
    elif teleporter.get("initial_state") != "ready" or teleporter.get("transition") != "waypoint":
        report.fail("teleporter interactable", "expected ready/waypoint")
    else:
        report.ok("teleporter interactable is ready/waypoint")

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
    elif not slice_golden.get("pinned_seed"):
        report.fail("slice_outcome golden", "missing pinned_seed")
    elif slice_golden["monster_def_id"] == "training_dummy" and slice_golden["final_player_hp"] != 9:
        report.fail("slice_outcome golden", "training_dummy pinned_seed final_player_hp must be 9")
    else:
        report.ok("slice_outcome golden references valid defs and pinned_seed")

    # item_visuals: every keyed item_def_id exists in items.v0.json with a
    # matching slot, and the visual-resolution golden matches the metadata
    # (spec equip-and-see-it §4.9 #1; rendering contract, not gameplay stats).
    visuals = load(ASSETS / "item_visuals.v0.json")["item_visuals"]
    for def_id, vis in visuals.items():
        item = items["items"].get(def_id)
        if item is None:
            report.fail("item_visuals key", f"{def_id} not in items.v0.json")
        elif item.get("slot") != vis["slot"]:
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

    # item_presentations: every current item has display metadata, and no
    # presentation entry points at a missing item. This is client-only rendering
    # data, but drift would make loot/inventory presentation fall back silently.
    presentations = load(ASSETS / "item_presentations.v0.json")["items"]
    for def_id in sorted(presentations):
        if def_id not in items["items"]:
            report.fail("item_presentations key", f"{def_id} not in items.v0.json")
        else:
            report.ok(f"item_presentations {def_id} resolves to items.v0.json")
    missing_presentations = sorted(set(items["items"]) - set(presentations))
    if missing_presentations:
        report.fail("item_presentations coverage", f"missing entries: {missing_presentations}")
    else:
        report.ok("item_presentations covers all item rules")


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
