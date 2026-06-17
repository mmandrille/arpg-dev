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
import math
import sys
from pathlib import Path

from jsonschema import Draft202012Validator

try:
    from .content_manifest import (
        ManifestError,
        merge_catalog_files,
        skill_presentation_entries,
        skill_rule_entries,
    )
    from .validate_boss_patterns import validate_boss_patterns
    from .validate_i18n import validate_i18n_catalog, validate_locale_catalog
    from .validate_item_presentations import validate_item_presentations
    from .validate_main_config import validate_main_config_gameplay
    from .validate_skills import validate_skill_catalogs
    from .validate_unique_items import validate_unique_items_catalog
except ImportError:  # pragma: no cover - direct script execution
    from content_manifest import (  # type: ignore[no-redef]
        ManifestError,
        merge_catalog_files,
        skill_presentation_entries,
        skill_rule_entries,
    )
    from validate_boss_patterns import validate_boss_patterns  # type: ignore[no-redef]
    from validate_i18n import validate_i18n_catalog, validate_locale_catalog  # type: ignore[no-redef]
    from validate_item_presentations import validate_item_presentations  # type: ignore[no-redef]
    from validate_main_config import validate_main_config_gameplay  # type: ignore[no-redef]
    from validate_skills import validate_skill_catalogs  # type: ignore[no-redef]
    from validate_unique_items import validate_unique_items_catalog  # type: ignore[no-redef]

ROOT = Path(__file__).resolve().parent.parent
SHARED = ROOT / "shared"
PROTOCOL = SHARED / "protocol"
RULES = SHARED / "rules"
GOLDEN = SHARED / "golden"
ASSETS = SHARED / "assets"
CONTENT = SHARED / "content"
I18N = SHARED / "i18n"
ASSET_MANIFEST = ROOT / "assets" / "manifests" / "assets.v0.json"


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


def adjacent_to(pos: dict, marker: dict, *, cell_size: float = 1.0) -> bool:
    dist = math.hypot(float(pos["x"]) - float(marker["x"]), float(pos["y"]) - float(marker["y"]))
    return dist > 1e-9 and dist <= math.sqrt(2) * cell_size + 1e-9


def equipment_visual_slot_matches(rule_slot: str | None, visual_slot: str) -> bool:
    if rule_slot == "ring":
        return visual_slot in {"ring_left", "ring_right"}
    return rule_slot == visual_slot


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
    if parts[0] == "content":
        # foo.v0.json -> foo.v0.schema.json
        return CONTENT / instance_path.name.replace(".v0.json", ".v0.schema.json")
    if parts[0] == "i18n":
        return I18N / "i18n.v0.schema.json"
    if parts[0] == "golden":
        # foo.json -> foo.v0.schema.json
        return GOLDEN / (instance_path.stem + ".v0.schema.json")
    if parts[0] == "protocol" and parts[1] == "examples":
        name = instance_path.name
        if name == "session_snapshot.json":
            return PROTOCOL / "session_snapshot.v8.schema.json"
        if name.startswith("state_delta"):
            return PROTOCOL / "state_delta.v8.schema.json"
        return PROTOCOL / "messages.v8.schema.json"
    raise ValueError(f"no schema mapping for {instance_path}")


def iter_schemas() -> list[Path]:
    return sorted(SHARED.rglob("*.schema.json"))


def iter_instances() -> list[Path]:
    instances: list[Path] = []
    instances += sorted(p for p in RULES.glob("*.v0.json") if not p.name.endswith(".schema.json"))
    instances += sorted(p for p in ASSETS.glob("*.v0.json") if not p.name.endswith(".schema.json"))
    instances += sorted(p for p in CONTENT.glob("*.v0.json") if not p.name.endswith(".schema.json"))
    instances += sorted(p for p in I18N.glob("*.json") if not p.name.endswith(".schema.json"))
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


class ShopRNG:
    """Splitmix64 RNG matching the Go server implementation.

    Used by generated-offer validation to replay shop stock deterministically
    and compare against golden fixtures.
    """

    def __init__(self, state: int) -> None:
        self.state = state & ((1 << 64) - 1)

    def next(self) -> int:
        self.state = (self.state + 0x9E3779B97F4A7C15) & ((1 << 64) - 1)
        z = self.state
        z = ((z ^ (z >> 30)) * 0xBF58476D1CE4E5B9) & ((1 << 64) - 1)
        z = ((z ^ (z >> 27)) * 0x94D049BB133111EB) & ((1 << 64) - 1)
        return (z ^ (z >> 31)) & ((1 << 64) - 1)

    def intn(self, n: int) -> int:
        if n <= 0:
            return 0
        return int(self.next() % n)


def seed_to_uint64(seed: str) -> int:
    """Convert a hex or ASCII seed string to uint64, matching the Go server."""
    try:
        raw = bytes.fromhex(seed)
        if not raw:
            raw = seed.encode()
    except ValueError:
        raw = seed.encode()
    value = 1469598103934665603
    for byte in raw:
        value ^= byte
        value = (value * 1099511628211) & ((1 << 64) - 1)
    return value


def cross_checks(report: Report) -> None:
    print("[3] cross-consistency drift guards")
    main_config = load(RULES / "main_config.v0.json")
    combat = load(RULES / "combat.v0.json")
    character_progression = load(RULES / "character_progression.v0.json")
    content_manifest_path = CONTENT / "content_libraries.v0.json"
    content_manifest = load(content_manifest_path)
    english_text = load(I18N / "en.json")
    locale_texts = [
        load(path)
        for path in sorted(I18N.glob("*.json"))
        if path.name != "en.json" and not path.name.endswith(".schema.json")
    ]
    skills = load(RULES / "skills.v0.json")
    skill_presentations = load(ASSETS / "skill_presentations.v0.json")
    class_presentations = load(ASSETS / "class_presentations.v0.json")
    items = load(RULES / "items.v0.json")
    item_templates = load(RULES / "item_templates.v0.json")
    unique_items = load(RULES / "unique_items.v0.json")
    unique_effects = load(RULES / "unique_effects.v0.json")
    unique_effect_defs = unique_effects.get("effects", {})
    set_item_ids = {piece.get("id") for set_def in load(RULES / "set_items.v0.json")["sets"].values() for piece in set_def.get("items", [])}
    treasure_classes = load(RULES / "treasure_classes.v0.json")
    monsters = load(RULES / "monsters.v0.json")
    loot = load(RULES / "loot_tables.v0.json")
    shops = load(RULES / "shops.v0.json")
    interactables = load(RULES / "interactables.v0.json")
    navigation = load(RULES / "navigation.v0.json")
    worlds = load(RULES / "worlds.v0.json")
    dungeon_generation = load(RULES / "dungeon_generation.v0.json")
    boss_templates = load(RULES / "boss_templates.v0.json")
    boss_patterns = load(RULES / "boss_patterns.v0.json")
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
    dungeon_stairs_golden = load(GOLDEN / "dungeon_stairs.json")
    dungeon_teleporters_golden = load(GOLDEN / "dungeon_teleporters.json")
    dungeon_monster_attack_golden = load(GOLDEN / "dungeon_monster_attack.json")
    item_rolls_golden = load(GOLDEN / "item_rolls.json")
    treasure_class_rolls_golden = load(GOLDEN / "treasure_class_rolls.json")
    dungeon_equipment_drops_golden = load(GOLDEN / "dungeon_equipment_drops.json")
    monster_rarity_golden = load(GOLDEN / "monster_rarity.json")
    guarded_chest_generation_golden = load(GOLDEN / "guarded_chest_generation.json")
    character_progression_golden = load(GOLDEN / "character_progression.json")
    skill_magic_golden = load(GOLDEN / "skill_points_and_magic_bolt.json")
    combat_stat_effects_golden = load(GOLDEN / "combat_stat_effects.json")
    boss_floor_golden = load(GOLDEN / "boss_floor_-5.json")
    boss_pattern_golden = load(GOLDEN / "boss_pattern_timeline.json")
    inventory_capacity_golden = load(GOLDEN / "inventory_capacity.json")
    dungeon_obstacles_golden = load(GOLDEN / "dungeon_obstacles.json")
    shop_pricing_golden = load(GOLDEN / "shop_pricing.json")
    shop_offers_golden = load(GOLDEN / "shop_offers.json")
    shop_appraisals_golden = load(GOLDEN / "shop_appraisals.json")
    shop_stock_lifecycle_golden = load(GOLDEN / "shop_stock_lifecycle.json")
    equipment_requirements_golden = load(GOLDEN / "equipment_requirements.json")
    main_gameplay = main_config.get("gameplay", {})

    validate_i18n_catalog(report, english_text, skills, skill_presentations, monsters)
    for locale_text in locale_texts:
        validate_locale_catalog(report, locale_text, english_text)

    try:
        manifest_skills = merge_catalog_files(content_manifest_path, skill_rule_entries(content_manifest), "skills")
    except ManifestError as exc:
        manifest_skills = {}
        report.fail("content manifest skill rules", str(exc))
    else:
        if set(manifest_skills) != set(skills.get("skills", {})):
            report.fail(
                "content manifest skill rules",
                f"merged ids {sorted(manifest_skills)} != skills.v0.json ids {sorted(skills.get('skills', {}))}",
            )
        else:
            report.ok("content manifest skill rules merge to canonical skills")

    try:
        manifest_skill_presentations = merge_catalog_files(
            content_manifest_path,
            skill_presentation_entries(content_manifest),
            "skills",
        )
    except ManifestError as exc:
        manifest_skill_presentations = {}
        report.fail("content manifest skill presentations", str(exc))
    else:
        if set(manifest_skill_presentations) != set(skill_presentations.get("skills", {})):
            report.fail(
                "content manifest skill presentations",
                "merged ids "
                f"{sorted(manifest_skill_presentations)} != skill_presentations.v0.json ids "
                f"{sorted(skill_presentations.get('skills', {}))}",
            )
        else:
            report.ok("content manifest skill presentations merge to canonical presentations")

    v4_protocol_files = [
        PROTOCOL / "envelope.v4.schema.json",
        PROTOCOL / "messages.v4.schema.json",
        PROTOCOL / "session_snapshot.v4.schema.json",
        PROTOCOL / "state_delta.v4.schema.json",
    ]
    v5_protocol_files = [
        PROTOCOL / "envelope.v5.schema.json",
        PROTOCOL / "messages.v5.schema.json",
        PROTOCOL / "session_snapshot.v5.schema.json",
        PROTOCOL / "state_delta.v5.schema.json",
    ]
    v6_protocol_files = [
        PROTOCOL / "envelope.v6.schema.json",
        PROTOCOL / "messages.v6.schema.json",
        PROTOCOL / "session_snapshot.v6.schema.json",
        PROTOCOL / "state_delta.v6.schema.json",
    ]
    v7_protocol_files = [
        PROTOCOL / "envelope.v7.schema.json",
        PROTOCOL / "messages.v7.schema.json",
        PROTOCOL / "session_snapshot.v7.schema.json",
        PROTOCOL / "state_delta.v7.schema.json",
    ]
    v8_protocol_files = [
        PROTOCOL / "envelope.v8.schema.json",
        PROTOCOL / "messages.v8.schema.json",
        PROTOCOL / "session_snapshot.v8.schema.json",
        PROTOCOL / "state_delta.v8.schema.json",
    ]
    missing_v4 = [str(path.relative_to(ROOT)) for path in v4_protocol_files if not path.exists()]
    if missing_v4:
        report.fail("protocol v4 schema set", f"missing {', '.join(missing_v4)}")
    else:
        report.ok("protocol v4 schema set is present")
    missing_v5 = [str(path.relative_to(ROOT)) for path in v5_protocol_files if not path.exists()]
    if missing_v5:
        report.fail("protocol v5 schema set", f"missing {', '.join(missing_v5)}")
    else:
        report.ok("protocol v5 schema set is present")
    missing_v6 = [str(path.relative_to(ROOT)) for path in v6_protocol_files if not path.exists()]
    if missing_v6:
        report.fail("protocol v6 schema set", f"missing {', '.join(missing_v6)}")
    else:
        report.ok("protocol v6 schema set is present")
    missing_v7 = [str(path.relative_to(ROOT)) for path in v7_protocol_files if not path.exists()]
    if missing_v7:
        report.fail("protocol v7 schema set", f"missing {', '.join(missing_v7)}")
    else:
        report.ok("protocol v7 schema set is present")
    missing_v8 = [str(path.relative_to(ROOT)) for path in v8_protocol_files if not path.exists()]
    if missing_v8:
        report.fail("protocol v8 schema set", f"missing {', '.join(missing_v8)}")
    else:
        report.ok("protocol v8 schema set is present")

    messages_v4 = load(PROTOCOL / "messages.v4.schema.json")
    messages_v5 = load(PROTOCOL / "messages.v5.schema.json")
    messages_v6 = load(PROTOCOL / "messages.v6.schema.json")
    messages_v7 = load(PROTOCOL / "messages.v7.schema.json")
    messages_v8 = load(PROTOCOL / "messages.v8.schema.json")
    actor_fields = {"player_id", "account_id", "character_id"}
    for protocol_version, messages_schema in (("v4", messages_v4), ("v5", messages_v5), ("v6", messages_v6), ("v7", messages_v7), ("v8", messages_v8)):
        intent_names = [name for name in messages_schema["$defs"] if name.endswith("_intent") or name == "client_ready"]
        actor_leaks: list[str] = []
        for name in sorted(intent_names):
            intent_schema = messages_schema["$defs"][name]
            if intent_schema.get("additionalProperties") is not False:
                actor_leaks.append(f"{name}: additionalProperties must be false")
            leaked = actor_fields.intersection(intent_schema.get("properties", {}))
            if leaked:
                actor_leaks.append(f"{name}: actor fields {sorted(leaked)}")
        if actor_leaks:
            report.fail(f"protocol {protocol_version} actor-free intents", "; ".join(actor_leaks))
        else:
            report.ok(f"protocol {protocol_version} intents are actor-free")

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
        attack_mode = str(mdef.get("attack_mode", "melee"))
        if attack_mode not in ("melee", "ranged"):
            report.fail("monster attack_mode", f"{mid}: invalid attack_mode {attack_mode}")
            continue
        if attack_mode == "melee":
            if any(mdef.get(key) is not None for key in ("attack_range", "projectile_speed", "projectile_def_id")):
                report.fail("monster attack fields", f"{mid}: projectile fields only valid for ranged attacks")
                continue
        else:
            attack_damage = mdef.get("attack_damage")
            if attack_damage is None or int(mdef.get("attack_cooldown_ticks", 0)) <= 0:
                report.fail("monster ranged attack", f"{mid}: ranged attacks require damage and cooldown")
                continue
            if float(mdef.get("attack_range", 0)) <= float(combat["unarmed_reach"]):
                report.fail("monster ranged attack", f"{mid}: attack_range must exceed unarmed reach")
                continue
            if float(mdef.get("projectile_speed", 0)) <= 0:
                report.fail("monster ranged attack", f"{mid}: projectile_speed must be positive")
                continue
            if not str(mdef.get("projectile_def_id", "")):
                report.fail("monster ranged attack", f"{mid}: projectile_def_id is required")
                continue
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
        if attack_mode == "ranged":
            report.ok(f"monster {mid} ranged attack is valid")
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
        if move_speed <= 0 or move_speed > navigation["cell_size"]:
            report.fail("monster move_speed", f"{mid}: move_speed must be > 0 and <= navigation.cell_size")
        else:
            report.ok(f"monster {mid} chase behavior is valid")

    monster_placement = dungeon_generation["monster_placement"]
    monster_count = int(monster_placement["count"])
    pool = monster_placement.get("monster_pool", [])
    minimums = monster_placement.get("minimum_monsters", [])
    if monster_count > 0:
        pool_ids = {str(entry["monster_def_id"]) for entry in pool}
        total_weight = 0
        failed_pool = False
        for entry in pool:
            mid = str(entry["monster_def_id"])
            total_weight += int(entry["weight"])
            mdef = monsters["monsters"].get(mid)
            if mdef is None:
                report.fail("dungeon monster_pool", f"unknown monster {mid}")
                failed_pool = True
            elif mdef.get("behavior", "static") != "chase":
                report.fail("dungeon monster_pool", f"{mid} must use chase behavior")
                failed_pool = True
        min_total = 0
        for entry in minimums:
            mid = str(entry["monster_def_id"])
            min_total += int(entry["count"])
            if mid not in pool_ids:
                report.fail("dungeon minimum_monsters", f"{mid} must also be in monster_pool")
                failed_pool = True
        if pool and total_weight <= 0:
            report.fail("dungeon monster_pool", "total weight must be positive")
            failed_pool = True
        if min_total > monster_count:
            report.fail("dungeon minimum_monsters", f"total {min_total} exceeds count {monster_count}")
            failed_pool = True
        archer = monsters["monsters"].get("dungeon_archer")
        if archer is None or archer.get("attack_mode") != "ranged":
            report.fail("dungeon_archer", "must exist and use ranged attack_mode")
            failed_pool = True
        if "dungeon_archer" not in pool_ids:
            report.fail("dungeon monster_pool", "must include dungeon_archer")
            failed_pool = True
        if not any(str(entry["monster_def_id"]) == "dungeon_archer" and int(entry["count"]) > 0 for entry in minimums):
            report.fail("dungeon minimum_monsters", "must guarantee at least one dungeon_archer")
            failed_pool = True
        if not failed_pool:
            report.ok("dungeon monster pool includes guaranteed ranged archer")

    if retaliation is not None and retaliation_golden["retaliation_damage"] != retaliation:
        report.fail("retaliation_damage vs monster", "retaliation_damage mismatch")
    else:
        report.ok("retaliation_damage golden matches training_dummy")

    progression_stats = {"str", "dex", "vit", "magic"}
    derived_keys = {
        "damage_min",
        "damage_max",
        "armor",
        "attack_speed",
        "hit_chance",
        "crit_chance",
        "crit_damage",
        "movement_speed",
        "max_hp",
        "max_mana",
        "health_regen_per_second",
        "mana_regen_per_second",
        "light_radius",
    }
    progression_base_stats = character_progression["base_stats"]
    if set(progression_base_stats) != progression_stats:
        report.fail("character_progression base_stats", "must define str/dex/vit/magic exactly")
    elif any(value < 1 for value in progression_base_stats.values()):
        report.fail("character_progression base_stats", "all base stats must be >= 1")
    else:
        report.ok("character_progression base_stats are valid")

    class_defs = character_progression.get("classes", {})
    expected_classes = {"barbarian", "sorcerer", "paladin", "rogue", "ranger"}
    if set(class_defs) != expected_classes:
        report.fail("character_progression classes", f"must define exactly {sorted(expected_classes)}")
    else:
        invalid_classes = []
        duplicate_stat_shapes = set()
        for class_id, class_def in class_defs.items():
            stats = class_def.get("base_stats", {})
            if not class_def.get("name") or set(stats) != progression_stats or any(value < 1 for value in stats.values()):
                invalid_classes.append(class_id)
            duplicate_stat_shapes.add(tuple(stats.get(stat, 0) for stat in sorted(progression_stats)))
        if invalid_classes:
            report.fail("character_progression classes", f"invalid class stat maps: {invalid_classes}")
        elif len(duplicate_stat_shapes) != len(expected_classes):
            report.fail("character_progression classes", "each class must have distinct starting stats")
        else:
            report.ok("character_progression classes are valid")
    class_presentation_defs = class_presentations.get("classes", {})
    missing_class_presentations = sorted(set(class_defs) - set(class_presentation_defs))
    extra_class_presentations = sorted(set(class_presentation_defs) - set(class_defs))
    if missing_class_presentations:
        report.fail("class_presentations coverage", f"missing presentations for {missing_class_presentations}")
    elif extra_class_presentations:
        report.fail("class_presentations keys", f"unknown classes {extra_class_presentations}")
    else:
        report.ok("class presentations cover character classes")
    manifest_assets = load(ASSET_MANIFEST)["assets"]
    bad_class_models = []
    for class_id, presentation in sorted(class_presentation_defs.items()):
        model = presentation.get("model", {})
        asset_id = model.get("asset_id") if isinstance(model, dict) else None
        asset = manifest_assets.get(asset_id or "")
        if not asset_id or asset is None or asset.get("type") != "character":
            bad_class_models.append(f"{class_id}:{asset_id}")
    if bad_class_models:
        report.fail("class_presentations model assets", f"must resolve to character assets: {bad_class_models}")
    else:
        report.ok("class presentation model assets resolve")

    if character_progression["points_per_level"] <= 0:
        report.fail("character_progression points_per_level", "must be positive")
    else:
        report.ok("character_progression points_per_level is positive")

    skill_point_rules = character_progression.get("skill_points", {})
    if not skill_point_rules:
        report.fail("character_progression skill_points", "must define v44 skill-point cadence")
    elif int(skill_point_rules.get("points_per_grant", 0)) <= 0:
        report.fail("character_progression skill_points.points_per_grant", "must be positive")
    elif int(skill_point_rules.get("grant_every_levels", 0)) <= 0:
        report.fail("character_progression skill_points.grant_every_levels", "must be positive")
    elif int(skill_point_rules.get("first_grant_level", 0)) < 1:
        report.fail("character_progression skill_points.first_grant_level", "must be at least 1")
    elif int(skill_point_rules["first_grant_level"]) > int(character_progression["level_cap"]):
        report.fail("character_progression skill_points.first_grant_level", "must be within level cap")
    else:
        report.ok("character_progression skill-point cadence is valid")

    curve = character_progression["experience_curve"]
    levels = curve["levels"]
    if curve.get("type") != "table":
        report.fail("character_progression experience_curve", "only table curves are supported")
    elif len(levels) != character_progression["level_cap"] - 1:
        report.fail("character_progression experience_curve", "must define one threshold for each level below cap")
    else:
        prev_xp = 0
        failed_curve = False
        for idx, entry in enumerate(levels, start=1):
            if entry["level"] != idx:
                report.fail("character_progression experience_curve", f"expected level {idx}, got {entry['level']}")
                failed_curve = True
                break
            if entry["next_level_total_xp"] <= prev_xp:
                report.fail("character_progression experience_curve", f"level {idx}: threshold must increase")
                failed_curve = True
                break
            prev_xp = entry["next_level_total_xp"]
        if not failed_curve:
            report.ok("character_progression experience_curve is monotonic and complete")

    formula_failed = False
    for stat_id, formula in character_progression["derived_stats"].items():
        if stat_id not in derived_keys:
            report.fail("character_progression derived stat", f"unsupported derived stat {stat_id}")
            formula_failed = True
            break
        formula_type = formula.get("type")
        if formula_type not in {"linear", "logarithmic"}:
            report.fail("character_progression formula", f"{stat_id}: unsupported formula type {formula_type}")
            formula_failed = True
            break
        if formula_type == "logarithmic":
            if formula.get("stat") not in progression_stats:
                report.fail("character_progression formula", f"{stat_id}: unsupported logarithmic stat {formula.get('stat')}")
                formula_failed = True
                break
            if float(formula.get("denominator", 0)) <= 0:
                report.fail("character_progression formula", f"{stat_id}: logarithmic denominator must be positive")
                formula_failed = True
                break
        for key in formula:
            if not key.startswith("per_"):
                continue
            if key.removeprefix("per_") not in progression_stats:
                report.fail("character_progression formula", f"{stat_id}: unsupported coefficient {key}")
                formula_failed = True
                break
        if formula_failed:
            break
        if "min" in formula and "max" in formula and formula["max"] < formula["min"]:
            report.fail("character_progression formula", f"{stat_id}: max must be >= min")
            formula_failed = True
            break
    if not formula_failed:
        report.ok("character_progression derived stat formulas are bounded and supported")

    def progression_level(experience: int) -> int:
        level = 1
        for entry in levels:
            if experience >= entry["next_level_total_xp"]:
                level = entry["level"] + 1
        return min(level, character_progression["level_cap"])

    def previous_threshold(level: int) -> int:
        if level <= 1:
            return 0
        return levels[level - 2]["next_level_total_xp"]

    def next_threshold(level: int) -> int:
        if level >= character_progression["level_cap"]:
            return previous_threshold(level)
        return levels[level - 1]["next_level_total_xp"]

    def evaluate_formula(formula: dict, stats: dict) -> float:
        value = float(formula["base"])
        if formula.get("type") == "logarithmic":
            raw = max(0.0, float(stats[str(formula["stat"])]) - float(formula.get("offset", 0.0)))
            value += float(formula["scale"]) * (math.log1p(raw) / math.log1p(float(formula["denominator"])))
        else:
            for stat in progression_stats:
                value += float(formula.get(f"per_{stat}", 0.0)) * float(stats[stat])
        if "min" in formula:
            value = max(value, float(formula["min"]))
        if "max" in formula:
            value = min(value, float(formula["max"]))
        return round(value, 6)

    def derived_stats_for(stats: dict) -> dict[str, float]:
        derived = {
            stat_id: evaluate_formula(character_progression["derived_stats"][stat_id], stats)
            for stat_id in sorted(character_progression["derived_stats"])
        }
        derived["damage_min"] += combat["player_damage"]["min"]
        derived["damage_max"] += combat["player_damage"]["max"]
        attack_speed = max(float(combat["min_effective_attack_speed"]), min(float(combat["max_effective_attack_speed"]), float(derived["attack_speed"])))
        derived["attack_speed"] = round(attack_speed, 6)
        derived["attack_interval_ticks"] = int(math.ceil(int(combat["base_attack_interval_ticks"]) / attack_speed))
        return derived

    def expected_skill_points_for_level(level: int) -> int:
        cadence = character_progression["skill_points"]
        if level < int(cadence["first_grant_level"]):
            return 0
        grants = ((level - int(cadence["first_grant_level"])) // int(cadence["grant_every_levels"])) + 1
        return grants * int(cadence["points_per_grant"])

    if character_progression_golden["base_stats"] != progression_base_stats:
        report.fail("character_progression golden", "base_stats must match character_progression.v0.json")
    elif character_progression_golden["points_per_level"] != character_progression["points_per_level"]:
        report.fail("character_progression golden", "points_per_level must match character_progression.v0.json")
    else:
        failed_progression_golden = False
        for case in character_progression_golden["cases"]:
            stats = dict(case["base_stats"])
            unspent = int(case["starting_unspent_stat_points"])
            level = progression_level(int(case["experience"]))
            gained_levels = max(0, level - 1)
            unspent += gained_levels * int(character_progression["points_per_level"])
            if "allocated_stat" in case:
                stat = case["allocated_stat"]
                points = int(case.get("allocated_points", 0))
                if points <= 0 or unspent < points:
                    report.fail("character_progression golden", f"{case['name']}: invalid allocation")
                    failed_progression_golden = True
                    break
                stats[stat] += points
                unspent -= points
            expected = case["expected"]
            prev = previous_threshold(level)
            next_total = next_threshold(level)
            expected_current = int(case["experience"]) - prev
            expected_next = max(0, next_total - prev)
            if expected["level"] != level:
                report.fail("character_progression golden", f"{case['name']}: level mismatch")
                failed_progression_golden = True
                break
            if expected["current_level_xp"] != expected_current or expected["next_level_xp"] != expected_next:
                report.fail("character_progression golden", f"{case['name']}: XP progress mismatch")
                failed_progression_golden = True
                break
            if expected["unspent_stat_points"] != unspent:
                report.fail("character_progression golden", f"{case['name']}: unspent points mismatch")
                failed_progression_golden = True
                break
            if expected["unspent_skill_points"] != expected_skill_points_for_level(level):
                report.fail("character_progression golden", f"{case['name']}: unspent skill points mismatch")
                failed_progression_golden = True
                break
            if expected["base_stats"] != stats:
                report.fail("character_progression golden", f"{case['name']}: base stats mismatch")
                failed_progression_golden = True
                break
            got_derived = derived_stats_for(stats)
            for stat_id, want in expected["derived_stats"].items():
                got = got_derived[stat_id]
                if not math.isclose(float(want), got, rel_tol=0, abs_tol=0.000001):
                    report.fail("character_progression golden", f"{case['name']}.{stat_id}: got {got}, want {want}")
                    failed_progression_golden = True
                    break
            if failed_progression_golden:
                break
        if not failed_progression_golden:
            report.ok("character_progression golden matches rules and formulas")

    if skill_magic_golden["progression"]["points_per_level"] != character_progression["points_per_level"]:
        report.fail("skill_points golden progression", "points_per_level must match character_progression.v0.json")
    elif skill_magic_golden["progression"]["skill_points"] != skill_point_rules:
        report.fail("skill_points golden progression", "skill point cadence must match character_progression.v0.json")
    else:
        failed_skill_progression = False

        def skill_points_for_level(level: int) -> int:
            first = int(skill_point_rules["first_grant_level"])
            every = int(skill_point_rules["grant_every_levels"])
            points = int(skill_point_rules["points_per_grant"])
            if level < first:
                return 0
            return ((level - first) // every + 1) * points

        for case in skill_magic_golden["progression"]["level_cases"]:
            level = int(case["level"])
            expected_stat_points = max(0, level - 1) * int(character_progression["points_per_level"])
            expected_skill_points = skill_points_for_level(level)
            if int(case["expected_unspent_stat_points"]) != expected_stat_points:
                report.fail("skill_points golden progression", f"level {level}: stat points mismatch")
                failed_skill_progression = True
                break
            if int(case["expected_unspent_skill_points"]) != expected_skill_points:
                report.fail("skill_points golden progression", f"level {level}: skill points mismatch")
                failed_skill_progression = True
                break
        if not failed_skill_progression:
            report.ok("skill_points golden progression cadence matches rules")

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

    equipment_slots = {
        "head", "amulet", "chest", "gloves", "belt", "boots",
        "ring_left", "ring_right", "main_hand", "off_hand",
    }
    hand_slots = {"main_hand", "off_hand"}
    required_templates = {
        "cave_blade", "cave_greatsword", "cave_bow", "cave_shield",
        "cave_helm", "cave_mail", "cave_gloves", "cave_belt",
        "cave_boots", "cave_ring", "cave_amulet",
    }

    def treasure_class_id_for_table(table_id: str) -> str | None:
        table = loot["loot_tables"].get(table_id)
        if not table:
            return None
        treasure_class_id = table.get("treasure_class_id")
        if not treasure_class_id or treasure_class_id not in treasure_class_defs:
            return None
        return treasure_class_id

    def templates_reachable_from_table(table_id: str) -> set[str]:
        treasure_class_id = treasure_class_id_for_table(table_id)
        if not treasure_class_id:
            return set()
        return {
            entry["item_template_id"]
            for attempt in treasure_class_defs[treasure_class_id].get("attempts", [])
            for entry in attempt.get("entries", [])
            if "item_template_id" in entry
        }

    def success_weight_for_table(table_id: str) -> int:
        treasure_class_id = treasure_class_id_for_table(table_id)
        if not treasure_class_id:
            return 0
        return sum(int(attempt.get("success_weight", 0)) for attempt in treasure_class_defs[treasure_class_id].get("attempts", []))

    # Item damage is hand-equipment only and the equipped_weapon_damage golden
    # must mirror the referenced item definition.
    for item_id, item in items["items"].items():
        dmg = item.get("damage")
        reach = item.get("reach")
        attack_speed = item.get("attack_speed")
        is_weapon_item = item.get("equippable") and item.get("slot") in hand_slots and (dmg is not None or reach is not None or item.get("attack_mode"))
        if is_weapon_item:
            if not isinstance(attack_speed, (int, float)):
                report.fail("item weapon attack_speed", f"{item_id}: weapon must declare attack_speed")
            elif attack_speed <= 0:
                report.fail("item weapon attack_speed", f"{item_id}: attack_speed must be positive")
            else:
                report.ok(f"item {item_id} weapon attack_speed is valid")
        elif attack_speed is not None:
            report.fail("item attack_speed eligibility", f"{item_id}: attack_speed is only valid on weapons")
        if dmg is not None:
            if not item.get("equippable") or item.get("slot") not in hand_slots:
                report.fail("item damage eligibility", f"{item_id}: damage is only valid on equippable hand items")
                continue
            if dmg["max"] < dmg["min"]:
                report.fail("item damage range", f"{item_id}: max must be >= min")
            else:
                report.ok(f"item {item_id} weapon damage range is valid")
        if reach is not None:
            if not item.get("equippable") or item.get("slot") not in hand_slots:
                report.fail("item reach eligibility", f"{item_id}: reach is only valid on equippable hand items")
                continue
            if reach <= 0:
                report.fail("item reach", f"{item_id}: reach must be positive")
            else:
                report.ok(f"item {item_id} weapon reach is valid")
        attack_mode = item.get("attack_mode", "melee")
        projectile_speed = item.get("projectile_speed")
        if attack_mode == "ranged":
            if not item.get("equippable") or item.get("slot") not in hand_slots:
                report.fail("item ranged eligibility", f"{item_id}: ranged mode is only valid on equippable hand items")
            elif dmg is None or reach is None or projectile_speed is None:
                report.fail("item ranged fields", f"{item_id}: ranged weapon requires damage, reach, and projectile_speed")
            elif projectile_speed <= 0:
                report.fail("item projectile_speed", f"{item_id}: projectile_speed must be positive")
            else:
                report.ok(f"item {item_id} ranged weapon fields are valid")
        elif projectile_speed is not None:
            report.fail("item projectile_speed", f"{item_id}: projectile_speed is only valid on ranged weapons")
        class_required = item.get("class_required")
        if class_required is not None:
            if class_required not in class_defs:
                report.fail("item class_required", f"{item_id}: unknown class {class_required}")
            elif not item.get("equippable"):
                report.fail("item class_required", f"{item_id}: only equippable items can require a class")
            else:
                report.ok(f"item {item_id} class requirement resolves")
    class_weapons = {
        "barbarian_axe": "barbarian",
        "sorcerer_staff": "sorcerer",
        "paladin_mace": "paladin",
        "ranger_shortbow": "ranger",
    }
    rusty_damage = items["items"].get("rusty_sword", {}).get("damage", {})
    for item_id, class_id in class_weapons.items():
        item = items["items"].get(item_id)
        if item is None:
            report.fail("class weapon", f"missing {item_id}")
        elif item.get("class_required") != class_id:
            report.fail("class weapon", f"{item_id}: must require {class_id}")
        elif item.get("damage", {}).get("max", 0) <= rusty_damage.get("max", 0):
            report.fail("class weapon", f"{item_id}: must be stronger than rusty_sword")
        else:
            report.ok(f"class weapon {item_id} is valid")

    valid_combat_roll_stats = {"damage_min", "damage_max", "str", "dex", "vit", "magic", "all_skills", "max_hp", "max_mana", "armor", "block_percent", "attack_speed_percent", "hit_chance", "crit_chance", "evade_chance", "health_regen_per_10_seconds", "mana_regen_per_10_seconds", "skill_damage_percent", "skill_cooldown_reduction_percent", "skill_mana_cost_reduction", "magic_find_percent", "light_radius"}
    valid_roll_stats = valid_combat_roll_stats | {"hotbar_slots", "inventory_rows"}
    rarities = item_templates["rarities"]
    for rarity_id, rarity in rarities.items():
        if rarity["weight"] <= 0:
            report.fail("item template rarity", f"{rarity_id}: weight must be positive")
        elif rarity["stat_rolls_min"] <= 0:
            report.fail("item template rarity", f"{rarity_id}: stat_rolls_min must be positive")
        elif rarity["stat_rolls_max"] < rarity["stat_rolls_min"]:
            report.fail("item template rarity", f"{rarity_id}: stat_rolls_max must be >= stat_rolls_min")
        elif rarity_id == "set" and rarity.get("random_rollable", True):
            report.fail("item template rarity", "set: random_rollable must stay false until random set drops are designed")
        else:
            report.ok(f"item template rarity {rarity_id} is valid")
    for template_id, template in item_templates["templates"].items():
        slot = template.get("slot")
        item_type = template.get("item_type")
        if template.get("category") != "equipment" or not template.get("equippable"):
            report.fail("item template equipment", f"{template_id}: templates must be equippable equipment")
            continue
        if slot not in equipment_slots and slot != "ring":
            report.fail("item template slot", f"{template_id}: unsupported slot {slot}")
            continue
        attack_mode = template.get("attack_mode")
        if slot in hand_slots and item_type != "shield":
            if attack_mode not in {"melee", "ranged"}:
                report.fail("item template attack_mode", f"{template_id}: hand weapons need attack_mode")
                continue
            if not isinstance(template.get("attack_speed"), (int, float)) or template.get("attack_speed", 0) <= 0:
                report.fail("item template attack_speed", f"{template_id}: weapons must declare positive attack_speed")
                continue
            if template.get("reach", 0) <= 0:
                report.fail("item template reach", f"{template_id}: weapon reach must be positive")
                continue
            if attack_mode == "ranged" and template.get("projectile_speed", 0) <= 0:
                report.fail("item template projectile_speed", f"{template_id}: ranged weapons need projectile_speed")
                continue
            if template.get("handedness") not in {"one_handed", "two_handed"}:
                report.fail("item template handedness", f"{template_id}: hand weapons need handedness")
                continue
            occupies = set(template.get("occupies_hands", []))
            if template["handedness"] == "one_handed" and not occupies:
                report.fail("item template occupies_hands", f"{template_id}: one-handed item must occupy a hand")
                continue
            if template["handedness"] == "two_handed" and occupies != {"main_hand", "off_hand"}:
                report.fail("item template occupies_hands", f"{template_id}: two-handed item must occupy both hands")
                continue
        elif attack_mode is not None or "reach" in template or "projectile_speed" in template:
            report.fail("item template combat fields", f"{template_id}: non-weapon equipment must not define attack fields")
            continue
        requirements = template.get("requirements", {})
        invalid_requirements = sorted(set(requirements) - (progression_stats | {"level"}))
        if invalid_requirements:
            report.fail("item template requirements", f"{template_id}: unsupported requirement(s) {invalid_requirements}")
            continue
        if any(int(value) < 1 for value in requirements.values()):
            report.fail("item template requirements", f"{template_id}: requirement values must be >= 1")
            continue
        base_stats = template["base_stats"]
        invalid_base_stats = sorted(set(base_stats) - valid_roll_stats)
        if invalid_base_stats:
            report.fail("item template base_stats", f"{template_id}: unsupported stat(s) {invalid_base_stats}")
            continue
        invalid_base_values = [
            stat for stat, value in base_stats.items()
            if (stat == "attack_speed_percent" and not -75 <= int(value) <= 100)
            or (stat in {"hit_chance", "crit_chance", "evade_chance"} and not 0 <= int(value) <= 100)
            or (stat not in {"attack_speed_percent", "hit_chance", "crit_chance", "evade_chance"} and int(value) < 0)
        ]
        if invalid_base_values:
            report.fail("item template base_stats", f"{template_id}: invalid value(s) for {invalid_base_values}")
            continue
        if "damage_min" in base_stats and "damage_max" in base_stats and base_stats["damage_max"] < base_stats["damage_min"]:
            report.fail("item template base_stats", f"{template_id}: damage_max must be >= damage_min")
            continue
        seen_roll_stats = set()
        failed_roll = False
        bounded_roll_stats = {"hit_chance": 100, "crit_chance": 100, "evade_chance": 100, "skill_cooldown_reduction_percent": 75, "skill_mana_cost_reduction": 20, "magic_find_percent": 500, "light_radius": 20}
        for roll in template["rollable_stats"]:
            stat = roll["stat"]
            seen_roll_stats.add(stat)
            if stat not in valid_roll_stats:
                report.fail("item template rollable stat", f"{template_id}: unsupported stat {stat}")
                failed_roll = True
                break
            if stat == "attack_speed_percent" and (roll["min"] < -50 or roll["max"] > 50):
                report.fail("item template rollable stat", f"{template_id}.{stat}: min/max must be within -50..50")
                failed_roll = True
                break
            if stat in bounded_roll_stats and (roll["min"] < 0 or roll["max"] > bounded_roll_stats[stat]):
                report.fail("item template rollable stat", f"{template_id}.{stat}: min/max must be within 0..{bounded_roll_stats[stat]}")
                failed_roll = True
                break
            if stat != "attack_speed_percent" and roll["min"] < 0:
                report.fail("item template rollable stat", f"{template_id}.{stat}: min must be non-negative")
                failed_roll = True
                break
            if roll["max"] < roll["min"]:
                report.fail("item template rollable stat", f"{template_id}.{stat}: max must be >= min")
                failed_roll = True
                break
            if roll["weight"] <= 0:
                report.fail("item template rollable stat", f"{template_id}.{stat}: weight must be positive")
                failed_roll = True
                break
        if failed_roll:
            continue
        if slot in hand_slots and item_type != "shield":
            if "damage_min" not in seen_roll_stats or "damage_max" not in seen_roll_stats:
                report.fail("item template rollable stats", f"{template_id}: weapons must include damage_min and damage_max")
            else:
                report.ok(f"item template {template_id} weapon rolls are valid")
        else:
            report.ok(f"item template {template_id} roll ranges are valid")
    report.ok("item template stat keys are restricted to supported rolls")

    blade_speed = item_templates["templates"].get("cave_blade", {}).get("attack_speed")
    greatsword_speed = item_templates["templates"].get("cave_greatsword", {}).get("attack_speed")
    training_bow_speed = items["items"].get("training_bow", {}).get("attack_speed")
    cave_bow_speed = item_templates["templates"].get("cave_bow", {}).get("attack_speed")
    if not isinstance(blade_speed, (int, float)) or not isinstance(greatsword_speed, (int, float)):
        report.fail("weapon attack_speed relationships", "cave_blade and cave_greatsword speeds are required")
    elif float(greatsword_speed) > float(blade_speed) * 0.70:
        report.fail("weapon attack_speed relationships", "cave_greatsword must be at least 30% slower than cave_blade")
    elif not isinstance(training_bow_speed, (int, float)) or not isinstance(cave_bow_speed, (int, float)):
        report.fail("weapon attack_speed relationships", "training_bow and cave_bow speeds are required")
    elif float(training_bow_speed) <= float(cave_bow_speed):
        report.fail("weapon attack_speed relationships", "training_bow short-bow proof must be faster than cave_bow")
    else:
        report.ok("weapon attack_speed relationships match v44")

    req_template_id = str(equipment_requirements_golden["template_id"])
    req_template = item_templates["templates"].get(req_template_id)
    if not req_template:
        report.fail("equipment_requirements golden", f"unknown template {req_template_id}")
    elif equipment_requirements_golden["requirements"] != req_template.get("requirements", {}):
        report.fail("equipment_requirements golden", "requirements must match item template")
    else:
        failed_equipment_requirements = False
        for case_key in ("fresh_character", "after_allocation"):
            case = equipment_requirements_golden[case_key]
            stats = case["base_stats"]
            if set(stats) != progression_stats:
                report.fail("equipment_requirements golden", f"{case_key}: base stats must define str/dex/vit/magic")
                failed_equipment_requirements = True
                break
            status = case["status"]
            expected_status = []
            for stat in ("level", "str", "dex", "vit", "magic"):
                if stat not in req_template["requirements"]:
                    continue
                current = int(case["level"]) if stat == "level" else int(stats[stat])
                required = int(req_template["requirements"][stat])
                expected_status.append({
                    "stat": stat,
                    "required": required,
                    "current": current,
                    "met": current >= required,
                })
            if status != expected_status:
                report.fail("equipment_requirements golden", f"{case_key}: status mismatch")
                failed_equipment_requirements = True
                break
            if bool(case["requirements_met"]) != all(row["met"] for row in expected_status):
                report.fail("equipment_requirements golden", f"{case_key}: requirements_met mismatch")
                failed_equipment_requirements = True
                break
        if not failed_equipment_requirements:
            report.ok("equipment_requirements golden matches template requirements")

    inventory_columns = int(inventory_capacity_golden["columns"])
    base_inventory_rows = int(inventory_capacity_golden["base_inventory_rows"])
    row_item_id = str(inventory_capacity_golden["row_granting_item"])
    row_item = item_templates["templates"].get(row_item_id)
    row_item_bonus = int(inventory_capacity_golden["row_item_bonus"])
    if inventory_columns != 5:
        report.fail("inventory_capacity golden", "columns must be 5")
    elif base_inventory_rows != 3:
        report.fail("inventory_capacity golden", "base_inventory_rows must be 3")
    elif inventory_capacity_golden["base_capacity"] != base_inventory_rows * inventory_columns:
        report.fail("inventory_capacity golden", "base_capacity must be base_inventory_rows * columns")
    elif not row_item:
        report.fail("inventory_capacity golden", f"row_granting_item {row_item_id} is missing")
    elif int(row_item.get("base_stats", {}).get("inventory_rows", 0)) != row_item_bonus:
        report.fail("inventory_capacity golden", f"{row_item_id}.base_stats.inventory_rows must match row_item_bonus")
    elif inventory_capacity_golden["row_item_capacity"] != (base_inventory_rows + row_item_bonus) * inventory_columns:
        report.fail("inventory_capacity golden", "row_item_capacity must include the row item bonus")
    elif inventory_capacity_golden["bag_occupancy"] != {
        "equipped_counts": False,
        "hotbar_assigned_counts": False,
        "bag_item_counts": True,
    }:
        report.fail("inventory_capacity golden", "bag occupancy flags must match v36 contract")
    elif inventory_capacity_golden["rejections"] != {
        "pickup_full_bag": "inventory_full",
        "capacity_shrink_overflow": "capacity_would_overflow",
    }:
        report.fail("inventory_capacity golden", "rejection reasons must match v36 contract")
    else:
        report.ok("inventory_capacity golden matches item templates and v36 contract")

    treasure_class_defs = treasure_classes["classes"]
    for class_id, treasure_class in treasure_class_defs.items():
        attempts = treasure_class.get("attempts", [])
        if not attempts:
            report.fail("treasure class attempts", f"{class_id}: must define at least one attempt")
            continue
        seen_attempt_ids = set()
        failed_class = False
        for attempt in attempts:
            attempt_id = attempt.get("attempt_id")
            if attempt_id in seen_attempt_ids:
                report.fail("treasure class attempt_id", f"{class_id}: duplicate attempt_id {attempt_id}")
                failed_class = True
                break
            seen_attempt_ids.add(attempt_id)
            success_weight = attempt.get("success_weight", -1)
            no_drop_weight = attempt.get("no_drop_weight", -1)
            if success_weight < 0 or no_drop_weight < 0 or success_weight + no_drop_weight <= 0:
                report.fail("treasure class weights", f"{class_id}.{attempt_id}: success/no_drop total must be positive")
                failed_class = True
                break
            if success_weight > 0 and not attempt.get("entries"):
                report.fail("treasure class entries", f"{class_id}.{attempt_id}: success requires entries")
                failed_class = True
                break
            for entry in attempt.get("entries", []):
                item_def_id = entry.get("item_def_id")
                item_template_id = entry.get("item_template_id")
                unique_item_id, set_item_id = entry.get("unique_item_id"), entry.get("set_item_id")
                if sum(1 for ref in [item_def_id, item_template_id, unique_item_id, set_item_id] if ref) != 1:
                    report.fail("treasure class entry", f"{class_id}.{attempt_id}: exactly one drop reference"); failed_class = True; break
                if entry.get("weight", 0) <= 0:
                    report.fail("treasure class entry", f"{class_id}.{attempt_id}: entry weight must be positive")
                    failed_class = True
                    break
                if item_def_id and item_def_id not in items["items"]:
                    report.fail("treasure class item", f"{class_id}.{attempt_id}: unknown item {item_def_id}"); failed_class = True; break
                if unique_item_id and unique_item_id not in unique_items["uniques"]:
                    report.fail("treasure class unique item", f"{class_id}.{attempt_id}: unknown unique item {unique_item_id}"); failed_class = True; break
                if set_item_id and set_item_id not in set_item_ids:
                    report.fail("treasure class set item", f"{class_id}.{attempt_id}: unknown set item {set_item_id}"); failed_class = True; break
                if item_template_id and item_template_id not in item_templates["templates"]:
                    report.fail("treasure class template", f"{class_id}.{attempt_id}: unknown template {item_template_id}"); failed_class = True; break
            if failed_class:
                break
        if not failed_class:
            report.ok(f"treasure class {class_id} attempts are valid")

    equipment_tc = treasure_class_defs.get("equipment_lab_tc_1")
    if not equipment_tc:
        report.fail("equipment_lab treasure class", "missing equipment_lab_tc_1")
    else:
        tc_templates = {
            entry["item_template_id"]
            for attempt in equipment_tc.get("attempts", [])
            for entry in attempt.get("entries", [])
            if "item_template_id" in entry
        }
        missing_tc_templates = sorted(required_templates - tc_templates)
        if missing_tc_templates:
            report.fail("equipment_lab treasure class", f"missing templates: {missing_tc_templates}")
        else:
            report.ok("equipment_lab treasure class covers every v28 template")

    validate_main_config_gameplay(
        report,
        main_gameplay,
        dungeon_generation,
        treasure_class_defs,
        treasure_class_id_for_table,
    )
    expected_base_drop_rate = int(main_gameplay.get("base_drop_rate_percent", -1))

    if combat.get("unarmed_reach", 0) <= 0:
        report.fail("combat unarmed_reach", "must be positive")
    else:
        report.ok("combat unarmed_reach is positive")
    if not 0 <= float(combat.get("base_hit_chance", -1)) <= 1:
        report.fail("combat base_hit_chance", "must be within [0, 1]")
    else:
        report.ok("combat base_hit_chance is bounded")
    if not 0 <= float(combat.get("base_crit_chance", -1)) <= 1:
        report.fail("combat base_crit_chance", "must be within [0, 1]")
    else:
        report.ok("combat base_crit_chance is bounded")
    if float(combat.get("base_crit_damage", 0)) < 1:
        report.fail("combat base_crit_damage", "must be >= 1")
    else:
        report.ok("combat base_crit_damage is valid")
    if int(combat.get("minimum_damage", 0)) < 1:
        report.fail("combat minimum_damage", "must be >= 1")
    else:
        report.ok("combat minimum_damage is valid")
    block_cap_percent = int(combat.get("block_cap_percent", -1))
    if block_cap_percent < 0:
        report.fail("combat block_cap_percent", "must be non-negative")
    elif block_cap_percent > 75:
        report.fail("combat block_cap_percent", "must not exceed 75")
    else:
        report.ok("combat block_cap_percent is within cap")
    base_attack_interval = int(combat.get("base_attack_interval_ticks", 0))
    min_attack_speed = float(combat.get("min_effective_attack_speed", 0))
    max_attack_speed = float(combat.get("max_effective_attack_speed", 0))
    if base_attack_interval <= 0:
        report.fail("combat base_attack_interval_ticks", "must be positive")
    elif min_attack_speed <= 0 or max_attack_speed <= 0 or max_attack_speed < min_attack_speed:
        report.fail("combat attack speed clamps", "min/max effective attack speed are invalid")
    else:
        report.ok("combat attack interval and speed clamps are valid")

    validate_skill_catalogs(
        report,
        skills,
        skill_presentations,
        class_defs,
        skill_magic_golden,
        base_attack_interval=base_attack_interval,
        min_attack_speed=min_attack_speed,
        max_attack_speed=max_attack_speed,
    )

    if navigation.get("cell_size", 0) <= 0:
        report.fail("navigation cell_size", "must be positive")
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
    monster_placement = dungeon_generation.get("monster_placement", {})
    monster_id = monster_placement.get("monster_def_id")
    monster_def = monsters["monsters"].get(monster_id)
    if monster_placement.get("count", -1) < 0:
        report.fail("dungeon_generation monster_placement", "count must be non-negative")
    elif monster_def is None:
        report.fail("dungeon_generation monster_placement", f"unknown monster_def_id {monster_id}")
    elif monster_def.get("behavior", "static") != "chase":
        report.fail("dungeon_generation monster_placement", f"{monster_id} must be a chase monster")
    elif monster_placement.get("margin_from_wall", -1) < 0:
        report.fail("dungeon_generation monster_placement", "margin_from_wall must be non-negative")
    elif monster_placement.get("min_spawn_distance", 0) <= 0:
        report.fail("dungeon_generation monster_placement", "min_spawn_distance must be positive")
    elif monster_placement.get("max_attempts", 0) <= 0:
        report.fail("dungeon_generation monster_placement", "max_attempts must be positive")
    else:
        report.ok("dungeon_generation monster placement is valid")
    chest_placement = dungeon_generation.get("chest_placement", {})
    chest_interactable_id = chest_placement.get("interactable_def_id")
    chest_loot_table = chest_placement.get("loot_table")
    chest_loot_def = loot["loot_tables"].get(chest_loot_table)
    if chest_placement.get("enabled") and chest_placement.get("chance_weight", 0) + chest_placement.get("no_chest_weight", 0) <= 0:
        report.fail("dungeon_generation chest_placement", "chance/no_chest total must be positive")
    elif chest_interactable_id not in interactables["interactables"]:
        report.fail("dungeon_generation chest_placement", f"unknown interactable_def_id {chest_interactable_id}")
    elif chest_interactable_id != "treasure_chest":
        report.fail("dungeon_generation chest_placement", "interactable_def_id must be treasure_chest in v25")
    elif chest_loot_def is None:
        report.fail("dungeon_generation chest_placement", f"unknown loot_table {chest_loot_table}")
    elif not chest_loot_def.get("treasure_class_id"):
        report.fail("dungeon_generation chest_placement", "loot_table must resolve to a treasure class")
    elif chest_loot_def["treasure_class_id"] not in treasure_class_defs:
        report.fail("dungeon_generation chest_placement", f"unknown treasure class {chest_loot_def['treasure_class_id']}")
    elif chest_placement.get("monster_count_bonus", -1) < 0:
        report.fail("dungeon_generation chest_placement", "monster_count_bonus must be non-negative")
    elif chest_placement.get("min_stair_distance", 0) <= 0:
        report.fail("dungeon_generation chest_placement", "min_stair_distance must be positive")
    elif chest_placement.get("max_attempts", 0) <= 0:
        report.fail("dungeon_generation chest_placement", "max_attempts must be positive")
    else:
        report.ok("dungeon_generation chest placement is valid")
    elite_objective = dungeon_generation.get("elite_objective", {})
    objective_interactable_id = elite_objective.get("interactable_def_id")
    objective_loot_table = elite_objective.get("loot_table")
    objective_loot_def = loot["loot_tables"].get(objective_loot_table)
    if elite_objective.get("enabled") and objective_interactable_id not in interactables["interactables"]:
        report.fail("dungeon_generation elite_objective", f"unknown interactable_def_id {objective_interactable_id}")
    elif elite_objective.get("enabled") and objective_interactable_id != "treasure_chest":
        report.fail("dungeon_generation elite_objective", "interactable_def_id must be treasure_chest in v158")
    elif elite_objective.get("enabled") and objective_loot_def is None:
        report.fail("dungeon_generation elite_objective", f"unknown loot_table {objective_loot_table}")
    elif elite_objective.get("enabled") and not objective_loot_def.get("treasure_class_id"):
        report.fail("dungeon_generation elite_objective", "loot_table must resolve to a treasure class")
    elif elite_objective.get("enabled") and objective_loot_def["treasure_class_id"] not in treasure_class_defs:
        report.fail("dungeon_generation elite_objective", f"unknown treasure class {objective_loot_def['treasure_class_id']}")
    elif elite_objective.get("enabled") and elite_objective.get("min_stair_distance", 0) <= 0:
        report.fail("dungeon_generation elite_objective", "min_stair_distance must be positive")
    elif elite_objective.get("enabled") and elite_objective.get("max_attempts", 0) <= 0:
        report.fail("dungeon_generation elite_objective", "max_attempts must be positive")
    else:
        report.ok("dungeon_generation elite objective is valid")
    obstacle_generation = dungeon_generation.get("obstacle_generation", {})
    target_group_count = obstacle_generation.get("target_group_count", {})
    wall_segment = obstacle_generation.get("wall_segment", {})
    solid_block = obstacle_generation.get("solid_block", {})
    shape_weights = obstacle_generation.get("shape_weights", {})
    clearance = obstacle_generation.get("clearance", {})
    obstacle_failed = False
    if obstacle_generation.get("max_attempts", 0) <= 0:
        report.fail("dungeon_generation obstacle_generation", "max_attempts must be positive")
        obstacle_failed = True
    elif target_group_count.get("min", -1) < 0 or target_group_count.get("max", -1) < target_group_count.get("min", 0):
        report.fail("dungeon_generation obstacle_generation.target_group_count", "min/max range is invalid")
        obstacle_failed = True
    elif wall_segment.get("min_length", 0) <= 0 or wall_segment.get("max_length", 0) < wall_segment.get("min_length", 0):
        report.fail("dungeon_generation obstacle_generation.wall_segment", "min/max length range is invalid")
        obstacle_failed = True
    elif wall_segment.get("thickness", 0) <= 0:
        report.fail("dungeon_generation obstacle_generation.wall_segment", "thickness must be positive")
        obstacle_failed = True
    else:
        min_size = solid_block.get("min_size", {})
        max_size = solid_block.get("max_size", {})
        for axis in ("x", "y"):
            if min_size.get(axis, 0) <= 0 or max_size.get(axis, 0) < min_size.get(axis, 0):
                report.fail("dungeon_generation obstacle_generation.solid_block", f"{axis} min/max range is invalid")
                obstacle_failed = True
                break
    if not obstacle_failed:
        weight_total = sum(int(shape_weights.get(name, 0)) for name in ("line", "l", "t", "block"))
        if weight_total <= 0:
            report.fail("dungeon_generation obstacle_generation.shape_weights", "at least one shape must have positive weight")
            obstacle_failed = True
        elif any(float(clearance.get(name, -1)) < 0 for name in ("player_spawn", "stairs", "teleporter", "chest", "monster", "loot")):
            report.fail("dungeon_generation obstacle_generation.clearance", "clearances must be non-negative")
            obstacle_failed = True
    if not obstacle_failed:
        max_shape_span = max(
            float(wall_segment.get("max_length", 0)),
            float(solid_block.get("max_size", {}).get("x", 0)),
            float(solid_block.get("max_size", {}).get("y", 0)),
        )
        floor_min_axis = min(float(floor_size.get("width", 0)), float(floor_size.get("height", 0)))
        if max_shape_span >= floor_min_axis:
            report.fail("dungeon_generation obstacle_generation", "largest obstacle span must fit inside floor")
        else:
            report.ok("dungeon_generation obstacle generation tuning is valid")
    monster_rarities = dungeon_generation.get("monster_rarities", [])
    rarity_by_id = {r["id"]: r for r in monster_rarities}
    expected_rarity_order = ["common", "champion", "rare", "unique"]
    if [r.get("id") for r in monster_rarities] != expected_rarity_order:
        report.fail("dungeon_generation monster_rarities", f"must be ordered {expected_rarity_order}")
    else:
        failed_rarity = False
        seen_rarity_ids = set()
        for rarity in monster_rarities:
            rarity_id = rarity["id"]
            seen_rarity_ids.add(rarity_id)
            if rarity_id.lower() != rarity_id or not rarity_id.replace("_", "").isalnum():
                report.fail("dungeon_generation monster_rarities", f"{rarity_id}: id must be stable lowercase")
                failed_rarity = True
                break
            if not isinstance(rarity.get("weight"), int) or rarity["weight"] <= 0:
                report.fail("dungeon_generation monster_rarities", f"{rarity_id}: weight must be positive")
                failed_rarity = True
                break
            color = rarity.get("color")
            if not isinstance(color, str) or len(color) != 7 or not color.startswith("#"):
                report.fail("dungeon_generation monster_rarities", f"{rarity_id}: color must be #RRGGBB")
                failed_rarity = True
                break
            try:
                int(color[1:], 16)
            except ValueError:
                report.fail("dungeon_generation monster_rarities", f"{rarity_id}: color must be hex")
                failed_rarity = True
                break
            for field in ("hp_multiplier", "damage_multiplier", "xp_multiplier"):
                if not isinstance(rarity.get(field), (int, float)) or rarity[field] <= 0:
                    report.fail("dungeon_generation monster_rarities", f"{rarity_id}.{field}: must be positive")
                    failed_rarity = True
                    break
            if failed_rarity:
                break
            if not isinstance(rarity.get("loot_depth_offset"), int) or rarity["loot_depth_offset"] < 0:
                report.fail("dungeon_generation monster_rarities", f"{rarity_id}: loot_depth_offset must be non-negative")
                failed_rarity = True
                break
        if not failed_rarity:
            if seen_rarity_ids != set(expected_rarity_order):
                report.fail("dungeon_generation monster_rarities", "must define common/champion/rare/unique exactly")
            else:
                report.ok("dungeon_generation monster rarities are structurally valid")
                report.ok("dungeon_generation monster rarities leave tuning values to monster_rarity golden")
    boss_floor = dungeon_generation.get("boss_floor", {})
    if boss_floor.get("cadence") != 5 or boss_floor.get("first_level") != -5:
        report.fail("boss_floor cadence", "cadence must be 5 and first_level must be -5")
    elif boss_floor.get("floor_size") != {"width": 30, "height": 30}:
        report.fail("boss_floor floor_size", "v35 boss floor must be 30 x 30")
    elif boss_floor.get("locked_exit_reason") != "boss_alive":
        report.fail("boss_floor locked_exit_reason", "must be boss_alive")
    else:
        report.ok("boss_floor cadence, size, and lock reason match v35")

    boss_floor_failed = False
    boss_floor_size = boss_floor.get("floor_size", {})
    width = boss_floor_size.get("width", 0)
    height = boss_floor_size.get("height", 0)
    for label in ("boss_spawn", "chest_position", "stairs_up_position", "stairs_down_position", "teleporter_position"):
        point = boss_floor.get(label, {})
        if not (0 <= point.get("x", -1) <= width and 0 <= point.get("y", -1) <= height):
            report.fail("boss_floor placement", f"{label} outside 30 x 30 floor")
            boss_floor_failed = True
            break
    if not boss_floor_failed:
        report.ok("boss_floor fixed placements fit 30 x 30 floor")

    obstacle_expected = dungeon_obstacles_golden.get("expected", {})
    obstacle_floor = obstacle_expected.get("floor_size", {})
    obstacle_shapes = set(obstacle_expected.get("shape_families", []))
    obstacle_walls = obstacle_expected.get("walls", [])
    generated_walls = [wall for wall in obstacle_walls if wall.get("source") == "generated"]
    if dungeon_obstacles_golden.get("level", 0) >= 0:
        report.fail("dungeon_obstacles golden", "level must be a generated dungeon floor")
    elif obstacle_floor != floor_size:
        report.fail("dungeon_obstacles golden", "floor_size must match dungeon_generation floor_size")
    elif len(obstacle_shapes) < 2:
        report.fail("dungeon_obstacles golden", "must name at least two shape families")
    elif obstacle_expected.get("minimum_generated_wall_count", 0) <= 0:
        report.fail("dungeon_obstacles golden", "minimum_generated_wall_count must be positive")
    elif len(generated_walls) > 0 and not obstacle_shapes.intersection({wall.get("shape_family") for wall in generated_walls}):
        report.fail("dungeon_obstacles golden", "generated wall shape_family must be represented in shape_families")
    else:
        report.ok("dungeon_obstacles golden declares v40 wall-layout contract")

    boss_chest_id = boss_floor.get("chest_interactable_def_id")
    boss_chest_table = boss_floor.get("chest_loot_table")
    if boss_chest_id != "treasure_chest" or boss_chest_id not in interactables["interactables"]:
        report.fail("boss_floor chest", "chest_interactable_def_id must resolve to treasure_chest")
    elif boss_chest_table not in loot["loot_tables"]:
        report.fail("boss_floor chest", f"unknown chest_loot_table {boss_chest_table}")
    else:
        report.ok("boss_floor chest rules resolve")

    missing_pool_templates = [
        template_id for template_id in boss_floor.get("boss_template_pool", [])
        if template_id not in boss_templates["bosses"]
    ]
    if missing_pool_templates:
        report.fail("boss_floor template pool", f"missing {missing_pool_templates}")
    else:
        report.ok("boss_floor template pool resolves")

    for template_id, template in boss_templates["bosses"].items():
        template_failed = False
        if template["base_monster_def_id"] not in monsters["monsters"]:
            report.fail("boss template base monster", f"{template_id}: {template['base_monster_def_id']} missing")
            template_failed = True
        elif template["loot_table"] not in loot["loot_tables"]:
            report.fail("boss template loot table", f"{template_id}: {template['loot_table']} missing")
            template_failed = True
        else:
            missing_patterns = [
                pattern_id for pattern_id in template["pattern_deck"]
                if pattern_id not in boss_patterns["patterns"]
            ]
            if missing_patterns:
                report.fail("boss template pattern deck", f"{template_id}: missing {missing_patterns}")
                template_failed = True
        visual = template.get("visual", {})
        if not template_failed and (visual.get("model") != "current_humanoid_player" or visual.get("scale") != 2.0):
            report.fail("boss template visual", f"{template_id}: v35 requires current_humanoid_player at 2.0 scale")
            template_failed = True
        if not template_failed:
            report.ok(f"boss template {template_id} references valid rules")

    validate_boss_patterns(report, boss_patterns, boss_pattern_golden, boss_floor, boss_floor_golden, boss_templates)
    loot_bands = dungeon_generation.get("loot_bands", [])
    if not loot_bands:
        report.fail("dungeon_generation loot_bands", "must define depth bands")
    else:
        failed_bands = False
        coverage = set()
        open_ended_bands = 0
        for idx, band in enumerate(loot_bands):
            min_depth = band.get("min_depth")
            max_depth = band.get("max_depth")
            label = f"band {idx}"
            if not isinstance(min_depth, int) or min_depth <= 0:
                report.fail("dungeon_generation loot_bands", f"{label}: min_depth must be positive")
                failed_bands = True
                break
            if max_depth is not None and (not isinstance(max_depth, int) or max_depth < min_depth):
                report.fail("dungeon_generation loot_bands", f"{label}: max_depth must be null or >= min_depth")
                failed_bands = True
                break
            if max_depth is None:
                open_ended_bands += 1
                covered_depths = {min_depth}
            else:
                covered_depths = set(range(min_depth, max_depth + 1))
            overlap = coverage & covered_depths
            if overlap:
                report.fail("dungeon_generation loot_bands", f"{label}: overlapping depths {sorted(overlap)}")
                failed_bands = True
                break
            coverage |= covered_depths
            for table_key in ("monster_loot_table", "chest_loot_table"):
                table_id = band.get(table_key)
                treasure_class_id = treasure_class_id_for_table(table_id)
                if treasure_class_id is None:
                    report.fail("dungeon_generation loot_bands", f"{label}.{table_key}: unknown table or treasure class {table_id}")
                    failed_bands = True
                    break
                report.ok(f"dungeon_generation {label} {table_key} resolves {treasure_class_id}")
            if failed_bands:
                break
            monster_success = success_weight_for_table(band["monster_loot_table"])
            chest_success = success_weight_for_table(band["chest_loot_table"])
            if chest_success <= monster_success:
                report.fail("dungeon_generation loot_bands", f"{label}: chest equipment odds must exceed monster odds")
                failed_bands = True
                break
        if not failed_bands:
            if coverage != {1, 2, 3} or open_ended_bands != 1:
                report.fail("dungeon_generation loot_bands", "must cover exactly depths 1, 2, and open-ended 3+")
            elif loot_bands[-1].get("min_depth") != 3 or loot_bands[-1].get("max_depth") is not None:
                report.fail("dungeon_generation loot_bands", "final band must be open-ended 3+")
            else:
                report.ok("dungeon_generation loot_bands cover 1, 2, and 3+ without overlap")
                depth3_band = loot_bands[-1]
                reachable_depth3 = (
                    templates_reachable_from_table(depth3_band["monster_loot_table"])
                    | templates_reachable_from_table(depth3_band["chest_loot_table"])
                )
                missing_depth3 = sorted(required_templates - reachable_depth3)
                if missing_depth3:
                    report.fail("dungeon_generation loot_bands 3+ reachability", f"missing templates: {missing_depth3}")
                else:
                    report.ok("dungeon_generation 3+ loot sources reach every v28 template")
    if monster_rarity_golden["monster_def_id"] != dungeon_generation["monster_placement"]["monster_def_id"]:
        report.fail("monster_rarity golden", "monster_def_id must match dungeon_generation monster placement")
    elif monster_rarity_golden["monster_def_id"] not in monsters["monsters"]:
        report.fail("monster_rarity golden", f"unknown monster_def_id {monster_rarity_golden['monster_def_id']}")
    else:
        failed_monster_rarity_golden = False
        base_monster = monsters["monsters"][monster_rarity_golden["monster_def_id"]]

        def rounded_positive(value: float) -> int:
            return max(1, int(math.floor(value + 0.5)))

        def rounded_non_negative(value: float) -> int:
            return max(0, int(math.floor(value + 0.5)))

        for golden_rarity in monster_rarity_golden["rarities"]:
            rarity_id = golden_rarity["id"]
            rule_rarity = rarity_by_id.get(rarity_id)
            if rule_rarity is None:
                report.fail("monster_rarity golden", f"{rarity_id}: missing from dungeon_generation monster_rarities")
                failed_monster_rarity_golden = True
                break
            for field in (
                "weight",
                "color",
                "hp_multiplier",
                "damage_multiplier",
                "xp_multiplier",
                "armor_multiplier",
                "armor_bonus",
                "hit_chance_bonus",
                "crit_chance_bonus",
                "block_percent_bonus",
                "attack_cooldown_multiplier",
                "loot_depth_offset",
                "visual_scale",
            ):
                if golden_rarity[field] != rule_rarity[field]:
                    report.fail("monster_rarity golden", f"{rarity_id}.{field}: mismatch with dungeon_generation")
                    failed_monster_rarity_golden = True
                    break
            if failed_monster_rarity_golden:
                break
            def clamp(value: float, minimum: float, maximum: float) -> float:
                return max(minimum, min(maximum, value))

            expected = golden_rarity["expected"]
            scaling = dungeon_generation["monster_depth_scaling"]
            depth_index = rule_rarity["loot_depth_offset"]
            hp_depth = 1 + scaling["hp_per_depth"] * depth_index
            damage_depth = 1 + scaling["damage_per_depth"] * depth_index
            expected_hp = rounded_positive(base_monster["max_hp"] * hp_depth * rule_rarity["hp_multiplier"])
            base_attack = base_monster["attack_damage"]
            expected_attack = {
                "min": rounded_positive(base_attack["min"] * damage_depth * rule_rarity["damage_multiplier"]),
                "max": rounded_positive(base_attack["max"] * damage_depth * rule_rarity["damage_multiplier"]),
            }
            expected_armor = rounded_non_negative(
                (base_monster.get("armor", 0) + scaling["armor_per_depth"] * depth_index)
                * rule_rarity["armor_multiplier"]
                + rule_rarity["armor_bonus"]
            )
            expected_hit = clamp(
                base_monster.get("hit_chance", combat["base_hit_chance"])
                + scaling["hit_chance_per_depth"] * depth_index
                + rule_rarity["hit_chance_bonus"],
                0,
                scaling["max_hit_chance"],
            )
            expected_crit = clamp(
                base_monster.get("crit_chance", combat["base_crit_chance"])
                + scaling["crit_chance_per_depth"] * depth_index
                + rule_rarity["crit_chance_bonus"],
                0,
                scaling["max_crit_chance"],
            )
            expected_block = clamp(
                base_monster.get("block_percent", 0)
                + scaling["block_percent_per_depth"] * depth_index
                + rule_rarity["block_percent_bonus"],
                0,
                scaling["max_block_percent"],
            )
            expected_cooldown = max(
                scaling["min_attack_cooldown_ticks"],
                rounded_positive(
                    base_monster["attack_cooldown_ticks"]
                    * (scaling["attack_cooldown_multiplier_per_depth"] ** depth_index)
                    * rule_rarity["attack_cooldown_multiplier"]
                ),
            )
            expected_xp = rounded_positive(base_monster["xp_reward"] * rule_rarity["xp_multiplier"])
            if expected["max_hp"] != expected_hp:
                report.fail("monster_rarity golden", f"{rarity_id}: max_hp mismatch")
                failed_monster_rarity_golden = True
                break
            if expected["attack_damage"] != expected_attack:
                report.fail("monster_rarity golden", f"{rarity_id}: attack_damage mismatch")
                failed_monster_rarity_golden = True
                break
            if (
                expected["armor"] != expected_armor
                or abs(expected["hit_chance"] - expected_hit) > 1e-12
                or abs(expected["crit_chance"] - expected_crit) > 1e-12
                or expected["block_percent"] != expected_block
                or expected["attack_cooldown_ticks"] != expected_cooldown
            ):
                report.fail("monster_rarity golden", f"{rarity_id}: derived stat mismatch")
                failed_monster_rarity_golden = True
                break
            if expected["xp_reward"] != expected_xp:
                report.fail("monster_rarity golden", f"{rarity_id}: xp_reward mismatch")
                failed_monster_rarity_golden = True
                break
        for case in monster_rarity_golden["effective_depth_cases"]:
            rarity = rarity_by_id.get(case["rarity"])
            if rarity is None:
                report.fail("monster_rarity effective depth", f"{case['rarity']}: unknown rarity")
                failed_monster_rarity_golden = True
                break
            effective_depth = abs(int(case["level"])) + int(rarity["loot_depth_offset"])
            matching_band = next(
                (
                    band for band in loot_bands
                    if effective_depth >= band["min_depth"]
                    and (band["max_depth"] is None or effective_depth <= band["max_depth"])
                ),
                None,
            )
            if matching_band is None:
                report.fail("monster_rarity effective depth", f"{case['level']} {case['rarity']}: no loot band")
                failed_monster_rarity_golden = True
                break
            if case["expected_monster_loot_table"] not in loot["loot_tables"]:
                report.fail("monster_rarity effective depth", f"{case['level']} {case['rarity']}: unknown expected loot table")
                failed_monster_rarity_golden = True
                break
        for case in monster_rarity_golden["generated_cases"]:
            for expected_monster in case["expected_monsters"]:
                rarity = rarity_by_id.get(expected_monster["rarity"])
                if rarity is None:
                    report.fail("monster_rarity generated case", f"{case['name']}: unknown rarity {expected_monster['rarity']}")
                    failed_monster_rarity_golden = True
                    break
                if expected_monster["loot_table"] not in loot["loot_tables"]:
                    report.fail("monster_rarity generated case", f"{case['name']}: unknown loot table {expected_monster['loot_table']}")
                    failed_monster_rarity_golden = True
                    break
            if failed_monster_rarity_golden:
                break
        if not failed_monster_rarity_golden:
            report.ok("monster_rarity golden matches rarity rules and depth bands")
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
    elif not golden_item.get("equippable") or golden_item.get("slot") not in hand_slots:
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
        for chance_key in ("hit_chance", "crit_chance"):
            chance = mdef.get(chance_key, combat["base_hit_chance"] if chance_key == "hit_chance" else combat["base_crit_chance"])
            if not isinstance(chance, (int, float)) or not 0 <= float(chance) <= 1:
                report.fail("monster combat chance", f"{mid}.{chance_key}: must be within [0, 1]")
                break
        else:
            crit_damage = mdef.get("crit_damage", combat["base_crit_damage"])
            armor = mdef.get("armor", 0)
            block_percent = mdef.get("block_percent", 0)
            if not isinstance(crit_damage, (int, float)) or float(crit_damage) < 1:
                report.fail("monster combat crit_damage", f"{mid}: must be >= 1")
            elif not isinstance(armor, int) or armor < 0:
                report.fail("monster combat armor", f"{mid}: must be a non-negative integer")
            elif not isinstance(block_percent, int) or block_percent < 0 or block_percent > 100:
                report.fail("monster combat block_percent", f"{mid}: raw stat must be within 0..100")
            else:
                report.ok(f"monster {mid} combat stats are bounded")
        xp_reward = mdef.get("xp_reward", 0)
        if not isinstance(xp_reward, int) or xp_reward < 0:
            report.fail("monster xp_reward", f"{mid}: xp_reward must be a non-negative integer")
        elif mid == dungeon_generation["monster_placement"]["monster_def_id"] and xp_reward <= 0:
            report.fail("monster xp_reward", f"{mid}: default dungeon monster must award XP")
        else:
            report.ok(f"monster {mid} xp_reward is valid")
        behavior = mdef.get("behavior", "static")
        if "attack_damage" in mdef:
            attack_damage = mdef["attack_damage"]
            if attack_damage["max"] < attack_damage["min"]:
                report.fail("monster attack_damage", f"{mid}: max must be >= min")
            elif behavior != "chase":
                report.fail("monster attack_damage", f"{mid}: proactive attack requires chase behavior")
            else:
                report.ok(f"monster {mid} attack damage is valid")
        if "attack_cooldown_ticks" in mdef:
            if "attack_damage" not in mdef:
                report.fail("monster attack cooldown", f"{mid}: cooldown requires attack_damage")
            elif behavior != "chase":
                report.fail("monster attack cooldown", f"{mid}: proactive attack requires chase behavior")
            else:
                report.ok(f"monster {mid} attack cooldown is valid")
        table = mdef["loot_table"]
        if table not in loot["loot_tables"]:
            report.fail("monster loot_table", f"{mid} -> unknown table {table}")
            continue
        loot_table = loot["loot_tables"][table]
        treasure_class_id = loot_table.get("treasure_class_id")
        if treasure_class_id:
            if treasure_class_id not in treasure_class_defs:
                report.fail("loot treasure class", f"{table} -> unknown treasure class {treasure_class_id}")
            else:
                report.ok(f"monster {mid} loot table resolves treasure class {treasure_class_id}")
            continue
        for entry in loot_table.get("entries", []):
            item_def_id = entry.get("item_def_id")
            item_template_id = entry.get("item_template_id")
            unique_item_id, set_item_id = entry.get("unique_item_id"), entry.get("set_item_id")
            if sum(1 for ref in [item_def_id, item_template_id, unique_item_id, set_item_id] if ref) != 1:
                report.fail("loot entry item", f"{table}: exactly one drop reference is required"); break
            if item_def_id and item_def_id not in items["items"]:
                report.fail("loot entry item", f"{table} -> unknown item {item_def_id}"); break
            if item_template_id and item_template_id not in item_templates["templates"]:
                report.fail("loot entry template", f"{table} -> unknown template {item_template_id}"); break
        else:
            for item_id in loot_table.get("drops", []):
                if item_id not in items["items"]:
                    report.fail("loot drop item", f"{table} -> unknown item {item_id}")
                    break
            else:
                report.ok(f"monster {mid} loot table + items resolve")

    no_drop = loot["loot_tables"].get("no_drop")
    if no_drop is None:
        report.fail("no_drop loot table", "missing table")
    elif no_drop.get("drops") or no_drop.get("entries"):
        report.fail("no_drop loot table", "must not define drops or entries")
    else:
        report.ok("no_drop loot table is empty")

    for table_id, loot_table in loot["loot_tables"].items():
        if "treasure_class_id" in loot_table and ("drops" in loot_table or "entries" in loot_table):
            report.fail("loot table shape", f"{table_id}: treasure_class_id cannot mix with drops/entries")
            continue
        treasure_class_id = loot_table.get("treasure_class_id")
        if treasure_class_id:
            if treasure_class_id not in treasure_class_defs:
                report.fail("loot table treasure_class_id", f"{table_id}: unknown {treasure_class_id}")
            else:
                report.ok(f"loot table {table_id} treasure_class_id resolves")

    shop_defs = shops["shops"]
    town_shop = shop_defs.get("town_vendor")
    if town_shop is None:
        report.fail("shop town_vendor", "missing town_vendor")
    else:
        pricing = town_shop["pricing"]
        fixed_offer_ids = [offer["offer_id"] for offer in town_shop["fixed_offers"]]
        if len(set(fixed_offer_ids)) != len(fixed_offer_ids):
            report.fail("shop fixed offers", "duplicate offer_id")
        else:
            report.ok("shop fixed offer ids are unique")
        expected_fixed = {"fixed:red_potion": "red_potion", "fixed:blue_potion": "blue_potion"}
        got_fixed = {offer["offer_id"]: offer["item_def_id"] for offer in town_shop["fixed_offers"]}
        if got_fixed != expected_fixed:
            report.fail("shop fixed offers", f"expected {expected_fixed}, got {got_fixed}")
        else:
            report.ok("shop fixed offers are red and blue potion")
        for offer in town_shop["fixed_offers"]:
            item = items["items"].get(offer["item_def_id"])
            if item is None:
                report.fail("shop fixed item", f"{offer['item_def_id']}: unknown item")
            elif item.get("category") in {"currency", "quest"}:
                report.fail("shop fixed item", f"{offer['item_def_id']}: currency/quest items cannot be fixed offers")
            elif offer["buy_price"] <= 0:
                report.fail("shop fixed price", f"{offer['offer_id']}: buy_price must be positive")
            else:
                report.ok(f"shop fixed offer {offer['offer_id']} resolves")

        generated = town_shop["generated_offers"]
        if generated.get("source") != "common_dungeon_mob":
            report.fail("shop generated source", "v41 source must be common_dungeon_mob")
        elif generated.get("offer_count") != 5:
            report.fail("shop generated count", "v41 offer_count must be 5")
        elif generated.get("min_depth") != 1:
            report.fail("shop generated min_depth", "v41 min_depth must be 1")
        elif generated.get("source_depth_policy") != "character_level_plus_one_to_deepest_else_any_achieved":
            report.fail("shop generated source_depth_policy", "v47 source-depth policy mismatch")
        elif generated.get("max_rarity") != "rare":
            report.fail("shop generated max_rarity", "v47 max rarity must be rare")
        elif generated.get("refresh_on") != "new_non_town_waypoint":
            report.fail("shop generated refresh_on", "v47 refresh trigger mismatch")
        elif generated.get("max_roll_attempts", 0) < generated.get("offer_count", 0):
            report.fail("shop generated max_roll_attempts", "must be >= offer_count")
        else:
            report.ok("shop generated offer config matches v47")

        buyback = town_shop.get("buyback", {})
        if buyback.get("enabled") is not True:
            report.fail("shop buyback", "must be enabled")
        elif buyback.get("scope") != "session_town_visit":
            report.fail("shop buyback scope", "must be session_town_visit")
        elif float(buyback.get("buy_price_multiplier", 0)) <= 0:
            report.fail("shop buyback multiplier", "must be positive")
        elif buyback.get("clear_on_leave_town") is not True:
            report.fail("shop buyback clear_on_leave_town", "must be true")
        else:
            report.ok("shop buyback config matches v47")

        random_rarities = {rarity_id for rarity_id, rarity in rarities.items() if rarity.get("random_rollable", True)}
        rarity_missing = sorted(random_rarities - set(pricing["rarity_multipliers"]))
        if rarity_missing:
            report.fail("shop rarity multipliers", f"missing current rarities {rarity_missing}")
        elif "unique" not in pricing["rarity_multipliers"]:
            report.fail("shop rarity multipliers", "must allow future unique multiplier")
        else:
            report.ok("shop rarity multipliers cover current rarities plus unique")
        template_slots = {template["slot"] for template in item_templates["templates"].values()}
        missing_slots = sorted(template_slots - set(pricing["slot_base"]))
        if missing_slots:
            report.fail("shop slot_base", f"missing template slots {missing_slots}")
        else:
            report.ok("shop slot_base covers current template slots")
        template_stats = {"str", "dex", "vit", "magic", "all_skills"}
        for template in item_templates["templates"].values():
            template_stats.update(template.get("base_stats", {}))
            template_stats.update(roll["stat"] for roll in template.get("rollable_stats", []))
        missing_weights = sorted(template_stats - set(pricing["stat_weights"]))
        if missing_weights:
            report.fail("shop stat_weights", f"missing template stats {missing_weights}")
        else:
            report.ok("shop stat_weights cover current base and rollable stats")

        def ceil_to_multiple(value: float, multiple: int) -> int:
            return int(math.ceil(value / multiple) * multiple)

        def fixed_buy_price(item_def_id: str) -> int | None:
            for offer in town_shop["fixed_offers"]:
                if offer["item_def_id"] == item_def_id:
                    return int(offer["buy_price"])
            return None

        def generated_buy_price(item_template_id: str, rarity: str, final_stats: dict) -> int:
            template = item_templates["templates"][item_template_id]
            base_stats = template.get("base_stats", {})
            base_score = int(pricing["slot_base"][template["slot"]])
            for stat, weight in pricing["stat_weights"].items():
                base_score += int(base_stats.get(stat, 0)) * int(weight)
            roll_score = 0
            for stat, weight in pricing["stat_weights"].items():
                roll_score += max(0, int(final_stats.get(stat, 0)) - int(base_stats.get(stat, 0))) * int(weight)
            raw_buy = (base_score + roll_score) * float(pricing["rarity_multipliers"][rarity])
            return ceil_to_multiple(max(1, raw_buy), int(pricing["round_buy_to"]))

        def sell_price(buy_price: int) -> int:
            return max(1, int(math.floor(buy_price * float(pricing["sell_multiplier"]))))

        failed_pricing = False
        if shop_pricing_golden["shop_id"] != "town_vendor":
            report.fail("shop_pricing golden", "shop_id must be town_vendor")
            failed_pricing = True
        for case in shop_pricing_golden["cases"]:
            if failed_pricing:
                break
            inp = case["input"]
            if inp.get("item_template_id"):
                buy = generated_buy_price(inp["item_template_id"], inp["rarity"], inp.get("rolled_stats", {}))
            else:
                fixed = fixed_buy_price(inp["item_def_id"])
                if fixed is None:
                    report.fail("shop_pricing golden", f"{case['name']}: item is not a fixed offer")
                    failed_pricing = True
                    break
                buy = fixed
            expected = case["expected"]
            got = {"buy_price": buy, "sell_price": sell_price(buy)}
            if got != expected:
                report.fail("shop_pricing golden", f"{case['name']}: got {got}, want {expected}")
                failed_pricing = True
                break
        if not failed_pricing:
            report.ok("shop_pricing golden matches v41 formula")

        def roll_treasure_class(class_id: str, rng: ShopRNG, drop_rate_override: int | None = None) -> list[dict]:
            out = []
            for attempt in treasure_class_defs[class_id].get("attempts", []):
                success_weight = int(attempt.get("success_weight", 0))
                no_drop_weight = int(attempt.get("no_drop_weight", 0))
                if drop_rate_override is not None and attempt.get("attempt_id") == "primary":
                    success_weight = drop_rate_override
                    no_drop_weight = 100 - drop_rate_override
                total = success_weight + no_drop_weight
                if total <= 0 or rng.intn(total) >= success_weight:
                    continue
                total_entries = sum(int(entry["weight"]) for entry in attempt.get("entries", []))
                if total_entries <= 0:
                    continue
                roll = rng.intn(total_entries)
                for entry in attempt.get("entries", []):
                    roll -= int(entry["weight"])
                    if roll < 0:
                        out.append(entry)
                        break
            return out

        rarity_order = sorted(rarities)
        def item_rarity_rank(rarity_id: str) -> int:
            return {"common": 0, "magic": 1, "rare": 2, "unique": 3, "set": 3}.get(rarity_id, -1)
        def scaled_attribute_roll_range(source_depth: int) -> tuple[int, int]:
            if source_depth < 1:
                source_depth = 1
            if source_depth <= 1:
                return 1, 3
            progress = min(1.0, max(0.0, float(source_depth - 1) / 99.0))
            max_value = round(3.0 + 47.0 * math.pow(progress, 1.15))
            if max_value < 3:
                max_value = 3
            min_value = max(1, math.floor(float(max_value) * 0.35))
            return int(min_value), int(max_value)
        def scaled_all_skills_roll_range(source_depth: int) -> tuple[int, int] | None:
            if source_depth < 10:
                return None
            return 1, max(1, source_depth // 10)
        def rollable_stats_for_rarity(template: dict, rarity_id: str, source_depth: int) -> list[dict]:
            stats = []
            for stat in template.get("rollable_stats", []):
                min_rarity = stat.get("min_rarity", "common")
                if item_rarity_rank(rarity_id) >= item_rarity_rank(min_rarity):
                    stats.append(stat)
            if item_rarity_rank(rarity_id) >= item_rarity_rank("magic"):
                min_value, max_value = scaled_attribute_roll_range(source_depth)
                for stat in ("str", "dex", "vit", "magic"):
                    stats.append({"stat": stat, "min_rarity": "magic", "min": min_value, "max": max_value, "weight": 2})
            if item_rarity_rank(rarity_id) >= item_rarity_rank("rare"):
                all_skills_range = scaled_all_skills_roll_range(source_depth)
                if all_skills_range is not None:
                    min_value, max_value = all_skills_range
                    stats.append({"stat": "all_skills", "min_rarity": "rare", "min": min_value, "max": max_value, "weight": 1})
            return stats
        def weighted_rollable_stat(stats: list[dict], rng: ShopRNG) -> dict | None:
            total = sum(int(stat["weight"]) for stat in stats)
            if total <= 0:
                return None
            roll = rng.intn(total)
            for stat in stats:
                roll -= int(stat["weight"])
                if roll < 0:
                    return stat
            return stats[-1]

        affix_words = {stat: (word, priority) for word, priority, group in [("Arcane", 90, "all_skills skill_damage_percent"), ("Focused", 85, "skill_cooldown_reduction_percent skill_mana_cost_reduction"), ("Keen", 80, "crit_chance hit_chance attack_speed_percent"), ("Savage", 70, "damage_min damage_max"), ("Stalwart", 65, "evade_chance block_percent armor"), ("Vigorous", 60, "max_hp health_regen_per_10_seconds vit"), ("Mystic", 55, "max_mana mana_regen_per_10_seconds magic"), ("Mighty", 50, "str"), ("Nimble", 50, "dex"), ("Fortunate", 48, "magic_find_percent"), ("Traveler's", 45, "inventory_rows hotbar_slots")] for stat in group.split()}

        def roll_template(template_id: str, rng: ShopRNG, source_depth: int = 1) -> dict:
            template = item_templates["templates"][template_id]
            total = sum(int(rarities[rarity_id]["weight"]) for rarity_id in rarity_order if rarities[rarity_id].get("random_rollable", True))
            roll = rng.intn(total)
            rarity_id = ""
            for candidate in rarity_order:
                if not rarities[candidate].get("random_rollable", True):
                    continue
                roll -= int(rarities[candidate]["weight"])
                if roll < 0:
                    rarity_id = candidate
                    break
            if not rarity_id:
                raise AssertionError("no random item rarity selected")
            stats = dict(template.get("base_stats", {}))
            rollable_stats = rollable_stats_for_rarity(template, rarity_id, source_depth)
            min_rolls = int(rarities[rarity_id]["stat_rolls_min"])
            max_rolls = int(rarities[rarity_id]["stat_rolls_max"])
            roll_count = min_rolls
            if max_rolls > min_rolls:
                roll_count += rng.intn(max_rolls - min_rolls + 1)
            for _ in range(roll_count):
                stat = weighted_rollable_stat(rollable_stats, rng)
                if stat is None:
                    continue
                stats[stat["stat"]] = int(stats.get(stat["stat"], 0)) + int(stat["min"]) + rng.intn(int(stat["max"]) - int(stat["min"]) + 1)
            display_name = f"{rarities[rarity_id]['name_prefix']} {template['name']}"
            gains = [(affix_words[s][1], int(v) - int(template.get("base_stats", {}).get(s, 0)), s, affix_words[s][0]) for s, v in stats.items() if s in affix_words and int(v) > int(template.get("base_stats", {}).get(s, 0))]
            if item_rarity_rank(rarity_id) >= item_rarity_rank("magic") and gains:
                display_name = f"{max(gains, key=lambda row: (row[0], row[1], ''.join(chr(255 - ord(ch)) for ch in row[2])))[3]} {display_name}"
            return {
                "item_template_id": template_id,
                "display_name": display_name,
                "rarity": rarity_id,
                "rolled_stats": stats,
            }

        def loot_band_for_depth(depth: int) -> dict | None:
            for band in dungeon_generation.get("loot_bands", []):
                if depth < int(band["min_depth"]):
                    continue
                max_depth = band.get("max_depth")
                if max_depth is not None and depth > int(max_depth):
                    continue
                return band
            return None

        def generated_shop_offers(seed: str, character_id: str, deepest_depth: int) -> list[dict]:
            depth = max(int(generated["min_depth"]), int(deepest_depth))
            label = f"{seed}|shop|town_vendor|{character_id}|{depth}|offers"
            rng = ShopRNG(seed_to_uint64(label))
            band = loot_band_for_depth(depth)
            if band is None:
                return []
            table = loot["loot_tables"].get(band["monster_loot_table"], {})
            class_id = table.get("treasure_class_id")
            if not class_id:
                return []
            out = []
            attempts = 0
            while len(out) < int(generated["offer_count"]) and attempts < int(generated["max_roll_attempts"]):
                attempts += 1
                for drop in roll_treasure_class(class_id, rng, expected_base_drop_rate):
                    template_id = drop.get("item_template_id")
                    if not template_id:
                        continue
                    template = item_templates["templates"].get(template_id)
                    if not template or template.get("category") != "equipment" or not template.get("equippable"):
                        continue
                    payload = roll_template(template_id, rng, depth)
                    buy = generated_buy_price(template_id, payload["rarity"], payload["rolled_stats"])
                    offer_index = len(out)
                    out.append({
                        "offer_id": f"generated:depth{depth}:{offer_index:03d}",
                        "kind": "generated",
                        "item_template_id": template_id,
                        "display_name": payload["display_name"],
                        "rarity": payload["rarity"],
                        "rolled_stats": payload["rolled_stats"],
                        "buy_price": buy,
                        "source": "common_dungeon_mob",
                        "depth": depth,
                    })
                    if len(out) >= int(generated["offer_count"]):
                        break
            return out

        failed_offers = False
        if shop_offers_golden["shop_id"] != "town_vendor":
            report.fail("shop_offers golden", "shop_id must be town_vendor")
            failed_offers = True
        for case in shop_offers_golden["cases"]:
            if failed_offers:
                break
            got = generated_shop_offers(shop_offers_golden["seed"], shop_offers_golden["character_id"], int(case["deepest_dungeon_depth"]))
            if len(got) != int(case["expected_offer_count"]):
                report.fail("shop_offers golden", f"{case['name']}: got {len(got)} offers")
                failed_offers = True
                break
            if got != case["expected"]:
                report.fail("shop_offers golden", f"{case['name']}: generated catalog drift")
                failed_offers = True
                break
        if not failed_offers:
            report.ok("shop_offers golden matches deterministic catalog")

        stat_order = ["damage_min", "damage_max", "str", "dex", "vit", "magic", "all_skills", "armor", "block_percent", "attack_speed_percent", "hit_chance", "crit_chance", "evade_chance", "max_hp", "max_mana", "health_regen_per_10_seconds", "mana_regen_per_10_seconds", "skill_damage_percent", "skill_cooldown_reduction_percent", "skill_mana_cost_reduction", "magic_find_percent", "hotbar_slots", "inventory_rows"]

        def comparison_deltas(offered: dict, equipped: dict) -> list[dict]:
            out = []
            for stat in stat_order:
                offered_value = int(offered.get(stat, 0))
                equipped_value = int(equipped.get(stat, 0))
                if offered_value == 0 and equipped_value == 0:
                    continue
                out.append({
                    "stat": stat,
                    "offered": offered_value,
                    "equipped": equipped_value,
                    "delta": offered_value - equipped_value,
                })
            return out

        failed_appraisals = False
        if shop_appraisals_golden["shop_id"] != "town_vendor":
            report.fail("shop_appraisals golden", "shop_id must be town_vendor")
            failed_appraisals = True
        if not failed_appraisals:
            fixed_case = shop_appraisals_golden["fixed_offer"]
            item = items["items"].get(fixed_case["item_def_id"])
            expected = fixed_case["expected"]
            if item is None:
                report.fail("shop_appraisals golden", f"unknown fixed item {fixed_case['item_def_id']}")
                failed_appraisals = True
            elif fixed_buy_price(fixed_case["item_def_id"]) != int(expected["buy_price"]):
                report.fail("shop_appraisals golden", f"fixed buy price mismatch for {fixed_case['item_def_id']}")
                failed_appraisals = True
            elif item.get("name") != expected["display_name"] or item.get("category") != expected["category"]:
                report.fail("shop_appraisals golden", "fixed display/category mismatch")
                failed_appraisals = True
            elif item.get("heal") and f"Restores {item['heal']['min']} HP" not in expected["summary_lines"]:
                report.fail("shop_appraisals golden", "fixed heal summary missing")
                failed_appraisals = True
        if not failed_appraisals:
            generated_case = shop_appraisals_golden["generated_offer"]
            template_id = generated_case["item_template_id"]
            template = item_templates["templates"].get(template_id)
            rarity = generated_case["rarity"]
            expected = generated_case["expected"]
            if template is None:
                report.fail("shop_appraisals golden", f"unknown generated template {template_id}")
                failed_appraisals = True
            elif rarity not in rarities:
                report.fail("shop_appraisals golden", f"unknown generated rarity {rarity}")
                failed_appraisals = True
            else:
                got_buy = generated_buy_price(template_id, rarity, generated_case["rolled_stats"])
                got_name = f"{rarities[rarity]['name_prefix']} {template['name']}"
                got_comparison = comparison_deltas(generated_case["rolled_stats"], generated_case["equipped_stats"])
                if got_buy != int(expected["buy_price"]):
                    report.fail("shop_appraisals golden", f"generated buy price {got_buy} != {expected['buy_price']}")
                    failed_appraisals = True
                elif got_name != expected["display_name"] or template["slot"] != expected["slot"] or template["category"] != expected["category"]:
                    report.fail("shop_appraisals golden", "generated display/slot/category mismatch")
                    failed_appraisals = True
                elif got_comparison != expected.get("comparison", []):
                    report.fail("shop_appraisals golden", f"comparison mismatch: {got_comparison} != {expected.get('comparison', [])}")
                    failed_appraisals = True
        if not failed_appraisals:
            sell_case = shop_appraisals_golden["sell_appraisal"]
            template_id = sell_case["item_template_id"]
            template = item_templates["templates"].get(template_id)
            rarity = sell_case["rarity"]
            expected = sell_case["expected"]
            if template is None:
                report.fail("shop_appraisals golden", f"unknown sell template {template_id}")
                failed_appraisals = True
            elif rarity not in rarities:
                report.fail("shop_appraisals golden", f"unknown sell rarity {rarity}")
                failed_appraisals = True
            else:
                got_sell = sell_price(generated_buy_price(template_id, rarity, sell_case["rolled_stats"]))
                got_name = f"{rarities[rarity]['name_prefix']} {template['name']}"
                got_comparison = comparison_deltas(sell_case["rolled_stats"], {})
                if got_sell != int(expected["sell_price"]):
                    report.fail("shop_appraisals golden", f"sell price {got_sell} != {expected['sell_price']}")
                    failed_appraisals = True
                elif got_name != expected["display_name"] or template["slot"] != expected["slot"] or template["category"] != expected["category"]:
                    report.fail("shop_appraisals golden", "sell display/slot/category mismatch")
                    failed_appraisals = True
                elif got_comparison != expected.get("comparison", []):
                    report.fail("shop_appraisals golden", f"sell comparison mismatch: {got_comparison} != {expected.get('comparison', [])}")
                    failed_appraisals = True
        if not failed_appraisals:
            if not shop_appraisals_golden["equipped_item_exclusion"].get("expected_excluded"):
                report.fail("shop_appraisals golden", "equipped_item_exclusion must expect exclusion")
            else:
                report.ok("shop_appraisals golden matches v42 appraisal contract")

        failed_lifecycle = False
        if shop_stock_lifecycle_golden["shop_id"] != "town_vendor":
            report.fail("shop_stock_lifecycle golden", "shop_id must be town_vendor")
            failed_lifecycle = True
        if not failed_lifecycle:
            golden_generated = shop_stock_lifecycle_golden["generated_stock"]
            for key in ("offer_count", "source", "source_depth_policy", "refresh_on", "max_rarity"):
                if golden_generated.get(key) != generated.get(key):
                    report.fail("shop_stock_lifecycle golden", f"generated_stock.{key} drift")
                    failed_lifecycle = True
                    break
        rarity_rank = {"common": 0, "magic": 1, "rare": 2}
        if not failed_lifecycle:
            max_rarity = str(generated["max_rarity"])
            if max_rarity not in rarities:
                report.fail("shop_stock_lifecycle max_rarity", f"unknown item rarity {max_rarity}")
                failed_lifecycle = True
            elif rarity_rank.get(max_rarity, 99) > rarity_rank["rare"]:
                report.fail("shop_stock_lifecycle max_rarity", "must not exceed rare")
                failed_lifecycle = True
        if not failed_lifecycle:
            def shop_source_depth_bounds(character_level: int, deepest_depth: int) -> tuple[int, int]:
                max_depth = max(int(generated["min_depth"]), int(deepest_depth))
                level_floor = int(character_level) + 1
                min_depth = level_floor if level_floor <= max_depth else int(generated["min_depth"])
                return min_depth, max_depth

            for case in shop_stock_lifecycle_golden["generated_stock"]["cases"]:
                got_min, got_max = shop_source_depth_bounds(int(case["character_level"]), int(case["deepest_dungeon_depth"]))
                if got_min != int(case["expected_min_source_depth"]) or got_max != int(case["expected_max_source_depth"]):
                    report.fail("shop_stock_lifecycle source-depth", f"{case['name']}: got {got_min}..{got_max}")
                    failed_lifecycle = True
                    break
        if not failed_lifecycle:
            finite = shop_stock_lifecycle_golden["finite_stock"]
            if int(finite["initial_generated_count"]) != int(generated["offer_count"]):
                report.fail("shop_stock_lifecycle finite_stock", "initial generated count mismatch")
                failed_lifecycle = True
            elif int(finite["after_generated_purchase_count"]) != int(generated["offer_count"]) - 1:
                report.fail("shop_stock_lifecycle finite_stock", "post-purchase generated count mismatch")
                failed_lifecycle = True
            elif int(finite["fixed_offer_count"]) != len(town_shop["fixed_offers"]):
                report.fail("shop_stock_lifecycle finite_stock", "fixed offer count mismatch")
                failed_lifecycle = True
        if not failed_lifecycle:
            golden_buyback = shop_stock_lifecycle_golden["buyback"]
            expected_buyback_price = max(1, int(math.ceil(int(golden_buyback["sell_price"]) * float(buyback["buy_price_multiplier"]))))
            if golden_buyback.get("enabled") != buyback.get("enabled"):
                report.fail("shop_stock_lifecycle buyback", "enabled mismatch")
                failed_lifecycle = True
            elif golden_buyback.get("scope") != buyback.get("scope"):
                report.fail("shop_stock_lifecycle buyback", "scope mismatch")
                failed_lifecycle = True
            elif float(golden_buyback["buy_price_multiplier"]) != float(buyback["buy_price_multiplier"]):
                report.fail("shop_stock_lifecycle buyback", "multiplier mismatch")
                failed_lifecycle = True
            elif int(golden_buyback["buy_price"]) != expected_buyback_price:
                report.fail("shop_stock_lifecycle buyback", f"buy price {golden_buyback['buy_price']} != {expected_buyback_price}")
                failed_lifecycle = True
            elif golden_buyback.get("clear_on_leave_town") != buyback.get("clear_on_leave_town"):
                report.fail("shop_stock_lifecycle buyback", "clear_on_leave_town mismatch")
                failed_lifecycle = True
            elif golden_buyback.get("persisted") is not False:
                report.fail("shop_stock_lifecycle buyback", "must not be persisted")
                failed_lifecycle = True
        if not failed_lifecycle:
            report.ok("shop_stock_lifecycle golden matches v47 lifecycle rules")

    town_vendor = interactables["interactables"].get("town_vendor")
    if town_vendor is None:
        report.fail("town_vendor interactable", "missing town_vendor")
    elif town_vendor.get("shop_id") != "town_vendor" or town_vendor.get("initial_state") != "ready":
        report.fail("town_vendor interactable", "must be ready and reference shop_id town_vendor")
    else:
        report.ok("town_vendor interactable references town_vendor shop")

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
            if etype in ("monster", "companion"):
                monster_id = entity.get("monster_def_id")
                if not monster_id:
                    report.fail("world monster entity", f"{label}: missing monster_def_id")
                elif monster_id not in monsters["monsters"]:
                    report.fail("world monster entity", f"{label}: unknown monster {monster_id}")
                else:
                    report.ok(f"{label} {etype} reference resolves")
            elif etype == "loot":
                item_id = entity.get("item_def_id")
                template_id = entity.get("item_template_id")
                if bool(item_id) == bool(template_id):
                    report.fail("world loot entity", f"{label}: exactly one of item_def_id/item_template_id")
                elif item_id and item_id not in items["items"]:
                    report.fail("world loot entity", f"{label}: unknown item {item_id}")
                elif template_id and template_id not in item_templates["templates"]:
                    report.fail("world loot entity", f"{label}: unknown item template {template_id}")
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

    combat_lab = worlds["worlds"].get("combat_stat_lab")
    required_combat_lab_monsters = {
        "combat_lab_soft_target",
        "combat_lab_armored_target",
        "combat_lab_blocking_target",
        "combat_lab_crit_attacker",
        "combat_lab_miss_attacker",
    }
    required_combat_lab_templates = {"cave_blade", "cave_shield", "cave_bow"}
    if combat_lab is None:
        report.fail("combat_stat_lab world", "missing world")
    else:
        lab_monsters = {
            entity.get("monster_def_id")
            for entity in combat_lab["entities"]
            if entity.get("type") == "monster"
        }
        lab_templates = {
            entity.get("item_template_id")
            for entity in combat_lab["entities"]
            if entity.get("type") == "loot" and entity.get("item_template_id")
        }
        missing_monsters = sorted(required_combat_lab_monsters - lab_monsters)
        missing_templates = sorted(required_combat_lab_templates - lab_templates)
        if missing_monsters:
            report.fail("combat_stat_lab monsters", f"missing {missing_monsters}")
        elif missing_templates:
            report.fail("combat_stat_lab equipment", f"missing {missing_templates}")
        else:
            report.ok("combat_stat_lab includes v31 proof monsters and equipment")

    if auto_path_golden["navigation"] != navigation:
        report.fail("auto_path navigation", "golden navigation block must match navigation.v0.json")
    else:
        report.ok("auto_path golden navigation matches navigation.v0.json")

    golden_combat = combat_stat_effects_golden["combat"]
    if golden_combat["minimum_damage"] != combat["minimum_damage"]:
        report.fail("combat_stat_effects combat", "minimum_damage must match combat.v0.json")
    elif golden_combat["block_cap_percent"] != combat["block_cap_percent"]:
        report.fail("combat_stat_effects combat", "block_cap_percent must match combat.v0.json")
    elif not math.isclose(float(golden_combat["base_crit_damage"]), float(combat["base_crit_damage"]), rel_tol=0, abs_tol=0.000001):
        report.fail("combat_stat_effects combat", "base_crit_damage must match combat.v0.json")
    elif combat_stat_effects_golden["world_id"] not in worlds["worlds"]:
        report.fail("combat_stat_effects world", f"unknown world_id {combat_stat_effects_golden['world_id']}")
    else:
        report.ok("combat_stat_effects golden matches combat constants and world")
    required_combat_case_names = {
        "player_miss",
        "player_crit",
        "monster_armor_minimum_damage",
        "player_armor_minimum_damage",
        "player_block",
        "block_cap_75",
        "monster_crit",
        "monster_block",
        "projectile_impact",
    }
    case_names = {case["name"] for case in combat_stat_effects_golden["cases"]}
    missing_case_names = sorted(required_combat_case_names - case_names)
    if missing_case_names:
        report.fail("combat_stat_effects cases", f"missing {missing_case_names}")
    else:
        failed_combat_case = False
        for case in combat_stat_effects_golden["cases"]:
            if case["outcome"] == "miss" and case["final_damage"] != 0:
                report.fail("combat_stat_effects cases", f"{case['name']}: miss must deal 0")
                failed_combat_case = True
                break
            if case["outcome"] == "block" and (not case["blocked"] or case["final_damage"] != 0):
                report.fail("combat_stat_effects cases", f"{case['name']}: block must set blocked and deal 0")
                failed_combat_case = True
                break
            if case["outcome"] == "crit" and not case["critical"]:
                report.fail("combat_stat_effects cases", f"{case['name']}: crit must set critical")
                failed_combat_case = True
                break
            if case["outcome"] not in {"miss", "block"} and case["final_damage"] < combat["minimum_damage"]:
                report.fail("combat_stat_effects cases", f"{case['name']}: non-blocked hit must respect minimum_damage")
                failed_combat_case = True
                break
        if not failed_combat_case:
            report.ok("combat_stat_effects cases cover v31 combat outcomes")
    breakdowns_by_key = {row["key"]: row for row in combat_stat_effects_golden["stat_breakdowns"]}
    block_breakdown = breakdowns_by_key.get("block_percent")
    if block_breakdown is None:
        report.fail("combat_stat_effects breakdowns", "missing block_percent breakdown")
    elif block_breakdown.get("cap") != combat["block_cap_percent"]:
        report.fail("combat_stat_effects breakdowns", "block cap must match combat.v0.json")
    elif block_breakdown["uncapped_value"] <= block_breakdown["value"]:
        report.fail("combat_stat_effects breakdowns", "block cap case must have uncapped_value > value")
    else:
        report.ok("combat_stat_effects breakdowns include capped block proof")

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
            if "barrier_when_closed" in interactable:
                size = interactable.get("barrier_when_closed", {}).get("size", {})
                if not isinstance(size.get("x"), (int, float)) or not isinstance(size.get("y"), (int, float)):
                    report.fail("interactable barrier", f"{interactable_id}: size must have numeric x/y")
                elif size["x"] <= 0 or size["y"] <= 0:
                    report.fail("interactable barrier", f"{interactable_id}: size must be positive")
                else:
                    report.ok(f"interactable {interactable_id} barrier size is positive")
            if "transition" in interactable:
                report.fail("interactable transition", f"{interactable_id}: closed blocker must not declare transition")
            if "shop_id" in interactable:
                report.fail("interactable shop", f"{interactable_id}: closed blocker must not declare shop_id")
            if "stash_id" in interactable:
                report.fail("interactable stash", f"{interactable_id}: closed blocker must not declare stash_id")
            continue
        if initial_state == "ready":
            transition = interactable.get("transition")
            shop_id = interactable.get("shop_id")
            stash_id = interactable.get("stash_id")
            service = interactable.get("service")
            if sum([bool(transition), bool(shop_id), bool(stash_id), bool(service)]) != 1:
                report.fail("interactable action", f"{interactable_id}: ready interactable needs exactly one transition, shop_id, stash_id, or service")
            elif transition and transition not in ("ascend", "descend", "waypoint"):
                report.fail("interactable transition", f"{interactable_id}: unsupported transition {transition}")
            elif shop_id and shop_id not in shop_defs:
                report.fail("interactable shop", f"{interactable_id}: unknown shop_id {shop_id}")
            elif stash_id and stash_id != "account_stash":
                report.fail("interactable stash", f"{interactable_id}: unknown stash_id {stash_id}")
            elif service and service not in {"bishop", "market", "mercenary", "blacksmith", "unique_test_chest"}:
                report.fail("interactable service", f"{interactable_id}: unsupported service {service}")
            elif "barrier_when_closed" in interactable:
                report.fail("interactable barrier", f"{interactable_id}: ready interactable must not block")
            elif shop_id:
                report.ok(f"interactable {interactable_id} ready shop is {shop_id}")
            elif stash_id:
                report.ok(f"interactable {interactable_id} ready stash is {stash_id}")
            elif service:
                report.ok(f"interactable {interactable_id} ready service is {service}")
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

    if dungeon_teleporters_golden["seed"] != dungeon_stairs_golden["seed"]:
        report.fail("dungeon_teleporters golden", "seed must match dungeon_stairs.json")
    else:
        tp_outcome = dungeon_teleporters_golden["discover_descend_teleport"]
        if tp_outcome["expected_level"] != -3:
            report.fail("dungeon_teleporters golden", "discover_descend_teleport.expected_level must be -3")
        elif not adjacent_to(tp_outcome["expected_player_position"], dungeon_stairs_golden["levels"]["-3"]["teleporter"]):
            report.fail("dungeon_teleporters golden", "expected_player_position must be adjacent to level -3 teleporter")
        else:
            report.ok("dungeon_teleporters golden matches stairs seed and travel outcome")

    # dungeon_monster_attack golden references proactive dungeon monster rules.
    golden_monster_id = dungeon_monster_attack_golden["monster_def_id"]
    golden_monster = monsters["monsters"].get(golden_monster_id)
    if golden_monster is None:
        report.fail("dungeon_monster_attack golden", f"unknown monster_def_id {golden_monster_id}")
    elif golden_monster_id != dungeon_generation["monster_placement"]["monster_def_id"]:
        report.fail("dungeon_monster_attack golden", "monster_def_id must match dungeon monster placement")
    elif dungeon_monster_attack_golden["level"] >= 0:
        report.fail("dungeon_monster_attack golden", "level must be a dungeon level")
    elif "attack_damage" not in golden_monster or "attack_cooldown_ticks" not in golden_monster:
        report.fail("dungeon_monster_attack golden", f"{golden_monster_id} is missing proactive attack fields")
    elif not any(
        rounded_positive(golden_monster["attack_damage"]["min"] * rarity["damage_multiplier"])
        <= dungeon_monster_attack_golden["damage"]
        <= rounded_positive(golden_monster["attack_damage"]["max"] * rarity["damage_multiplier"])
        for rarity in monster_rarities
    ):
        report.fail("dungeon_monster_attack golden", "damage is outside configured rarity-scaled monster attack_damage")
    elif dungeon_monster_attack_golden["player_hp_after"] != 10 - dungeon_monster_attack_golden["damage"]:
        report.fail("dungeon_monster_attack golden", "player_hp_after must reflect one hit from 10 HP")
    else:
        report.ok("dungeon_monster_attack golden matches proactive monster rules")

    template_id = item_rolls_golden["template_id"]
    template = item_templates["templates"].get(template_id)
    if template is None:
        report.fail("item_rolls golden", f"unknown template_id {template_id}")
    else:
        golden_failed = False
        for case in item_rolls_golden["cases"]:
            expected = case["expected"]
            rarity = expected["rarity"]
            stats = expected["stats"]
            if expected["item_template_id"] != template_id:
                report.fail("item_rolls golden", f"{case['name']}: item_template_id mismatch")
                golden_failed = True
                break
            if rarity not in rarities:
                report.fail("item_rolls golden", f"{case['name']}: unknown rarity {rarity}")
                golden_failed = True
                break
            effect_ids = expected.get("effect_ids", [])
            if expected.get("rarity") == "unique":
                if len(effect_ids) != 1:
                    report.fail("item_rolls golden", f"{case['name']}: unique rolls must attach exactly one effect id")
                    golden_failed = True
                    break
                elif effect_ids[0] not in unique_effect_defs:
                    report.fail("item_rolls golden", f"{case['name']}: unknown unique effect id {effect_ids[0]}")
                    golden_failed = True
                    break
                effect_name = unique_effect_defs[effect_ids[0]].get("display_name", "")
                item_type_name = str(template.get("item_type", "")).replace("_", " ").title()
                expected_unique_name = f"{item_type_name} of {effect_name}"
                if expected["display_name"] != expected_unique_name:
                    report.fail("item_rolls golden", f"{case['name']}: unique display_name must be {expected_unique_name!r}")
                    golden_failed = True
                    break
            elif effect_ids != []:
                report.fail("item_rolls golden", f"{case['name']}: non-unique effect_ids must be empty")
                golden_failed = True
                break
            elif not expected["display_name"].endswith(template["name"]):
                report.fail("item_rolls golden", f"{case['name']}: non-unique display_name must include template name")
                golden_failed = True
                break
            if stats["damage_max"] < stats["damage_min"]:
                report.fail("item_rolls golden", f"{case['name']}: damage_max must be >= damage_min")
                golden_failed = True
                break
            if expected.get("requirements", {}) != template.get("requirements", {}):
                report.fail("item_rolls golden", f"{case['name']}: requirements mismatch template")
                golden_failed = True
                break
        if not golden_failed:
            report.ok("item_rolls golden references valid template results")

    tc_id = treasure_class_rolls_golden["treasure_class_id"]
    if tc_id not in treasure_class_defs:
        report.fail("treasure_class_rolls golden", f"unknown treasure_class_id {tc_id}")
    else:
        failed_tc_golden = False
        for case in treasure_class_rolls_golden["cases"]:
            for drop in case["expected_drops"]:
                item_def_id = drop.get("item_def_id")
                item_template_id = drop.get("item_template_id")
                if bool(item_def_id) == bool(item_template_id):
                    report.fail("treasure_class_rolls golden", f"{case['name']}: expected drop must declare one item source")
                    failed_tc_golden = True
                    break
                if item_def_id and item_def_id not in items["items"]:
                    report.fail("treasure_class_rolls golden", f"{case['name']}: unknown item {item_def_id}")
                    failed_tc_golden = True
                    break
                if item_template_id and item_template_id not in item_templates["templates"]:
                    report.fail("treasure_class_rolls golden", f"{case['name']}: unknown template {item_template_id}")
                    failed_tc_golden = True
                    break
            if failed_tc_golden:
                break
        if not failed_tc_golden:
            report.ok("treasure_class_rolls golden references valid drop sources")

    if dungeon_equipment_drops_golden.get("world_id") not in worlds["worlds"]:
        report.fail("dungeon_equipment_drops golden", f"unknown world_id {dungeon_equipment_drops_golden.get('world_id')}")
    elif worlds["worlds"][dungeon_equipment_drops_golden["world_id"]].get("mode") != "multi_level":
        report.fail("dungeon_equipment_drops golden", "world_id must be a multi-level dungeon world")
    else:
        fixture_templates = set(dungeon_equipment_drops_golden.get("required_templates", []))
        if fixture_templates != required_templates:
            report.fail("dungeon_equipment_drops golden", f"required_templates mismatch: {sorted(fixture_templates)}")
        else:
            failed_dungeon_drop = False
            bands_by_depth = {band["min_depth"]: band for band in loot_bands}
            for band in dungeon_equipment_drops_golden["bands"]:
                depth = abs(int(band["level"]))
                rules_band = None
                for candidate in loot_bands:
                    max_depth = candidate.get("max_depth")
                    if depth >= candidate["min_depth"] and (max_depth is None or depth <= max_depth):
                        rules_band = candidate
                        break
                if rules_band is None:
                    report.fail("dungeon_equipment_drops golden", f"{band['level']}: no matching rules band")
                    failed_dungeon_drop = True
                    break
                if depth not in bands_by_depth and depth >= 3:
                    expected_monster = loot_bands[-1]["monster_loot_table"]
                    expected_chest = loot_bands[-1]["chest_loot_table"]
                else:
                    expected_monster = rules_band["monster_loot_table"]
                    expected_chest = rules_band["chest_loot_table"]
                if band["monster_loot_table"] != expected_monster or band["chest_loot_table"] != expected_chest:
                    report.fail("dungeon_equipment_drops golden", f"{band['level']}: loot table mismatch")
                    failed_dungeon_drop = True
                    break
            if not failed_dungeon_drop:
                for case in dungeon_equipment_drops_golden["cases"]:
                    table_id = case["loot_table"]
                    treasure_class_id = treasure_class_id_for_table(table_id)
                    if treasure_class_id != case["treasure_class_id"]:
                        report.fail("dungeon_equipment_drops golden", f"{case['name']}: treasure class mismatch")
                        failed_dungeon_drop = True
                        break
                    depth = abs(int(case["level"]))
                    rules_band = None
                    for candidate in loot_bands:
                        max_depth = candidate.get("max_depth")
                        if depth >= candidate["min_depth"] and (max_depth is None or depth <= max_depth):
                            rules_band = candidate
                            break
                    table_key = "monster_loot_table" if case["source"] == "monster" else "chest_loot_table"
                    if rules_band is None or case["loot_table"] != rules_band[table_key]:
                        report.fail("dungeon_equipment_drops golden", f"{case['name']}: source table mismatch")
                        failed_dungeon_drop = True
                        break
                    for drop in case["expected_drops"]:
                        item_def_id = drop.get("item_def_id")
                        item_template_id = drop.get("item_template_id")
                        if bool(item_def_id) == bool(item_template_id):
                            report.fail("dungeon_equipment_drops golden", f"{case['name']}: expected drop must declare one item source")
                            failed_dungeon_drop = True
                            break
                        if item_def_id and item_def_id not in items["items"]:
                            report.fail("dungeon_equipment_drops golden", f"{case['name']}: unknown item {item_def_id}")
                            failed_dungeon_drop = True
                            break
                        if item_template_id and item_template_id not in item_templates["templates"]:
                            report.fail("dungeon_equipment_drops golden", f"{case['name']}: unknown template {item_template_id}")
                            failed_dungeon_drop = True
                            break
                    if failed_dungeon_drop:
                        break
            if not failed_dungeon_drop:
                report.ok("dungeon_equipment_drops golden references valid depth bands and drop sources")

    if guarded_chest_generation_golden["level"] >= 0:
        report.fail("guarded_chest_generation golden", "level must be a dungeon level")
    else:
        guarded_depth = abs(int(guarded_chest_generation_golden["level"]))
        guarded_band = None
        for candidate in loot_bands:
            max_depth = candidate.get("max_depth")
            if guarded_depth >= candidate["min_depth"] and (max_depth is None or guarded_depth <= max_depth):
                guarded_band = candidate
                break
        failed_chest_golden = guarded_band is None
        if failed_chest_golden:
            report.fail("guarded_chest_generation golden", "level must resolve to a loot band")
        else:
            expected_guarded_loot_table = guarded_band["chest_loot_table"]
            for case in guarded_chest_generation_golden["cases"]:
                expected_chest = case["expected_chest"]
                if expected_chest is None:
                    continue
                if expected_chest["interactable_def_id"] != dungeon_generation["chest_placement"]["interactable_def_id"]:
                    report.fail("guarded_chest_generation golden", f"{case['name']}: interactable_def_id mismatch")
                    failed_chest_golden = True
                    break
                if expected_chest["loot_table"] != expected_guarded_loot_table:
                    report.fail("guarded_chest_generation golden", f"{case['name']}: loot_table mismatch")
                    failed_chest_golden = True
                    break
        if not failed_chest_golden:
            report.ok("guarded_chest_generation golden references valid chest rules")

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
        template = item_templates["templates"].get(def_id)
        slot = item.get("slot") if item is not None else (template or {}).get("slot")
        if item is None and template is None:
            report.fail("item_visuals key", f"{def_id} not in items.v0.json or item_templates.v0.json")
        elif not equipment_visual_slot_matches(slot, vis["slot"]):
            report.fail("item_visuals slot", f"{def_id}: {vis['slot']} != item/template slot {slot}")
        else:
            report.ok(f"item_visuals {def_id} resolves to item/template rules with matching slot")
    visual_required = {
        def_id
        for def_id, item in items["items"].items()
        if item.get("category") == "equipment" and item.get("equippable") and item.get("slot")
    } | {
        def_id
        for def_id, template in item_templates["templates"].items()
        if template.get("category") == "equipment" and template.get("equippable") and template.get("slot")
    }
    missing_visuals = sorted(visual_required - set(visuals))
    if missing_visuals:
        report.fail("item_visuals equipment coverage", f"missing equipment visual mappings: {missing_visuals}")
    else:
        report.ok("item_visuals covers all equippable equipment")

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

    full_equipment = load(GOLDEN / "full_equipment.json")
    fixture_slots = set(full_equipment.get("equipment_slots", []))
    if fixture_slots != equipment_slots:
        report.fail("full_equipment golden", f"equipment_slots mismatch: {sorted(fixture_slots)}")
    elif full_equipment.get("world_id") != "equipment_lab":
        report.fail("full_equipment golden", "world_id must be equipment_lab")
    elif not full_equipment.get("pinned_seed"):
        report.fail("full_equipment golden", "missing pinned_seed")
    else:
        report.ok("full_equipment golden declares v28 slot set and pinned lab")

    missing_templates = sorted(required_templates - set(item_templates["templates"]))
    if missing_templates:
        report.fail("full_equipment templates", f"missing templates: {missing_templates}")
    else:
        report.ok("full_equipment templates cover every equipment category")

    # health_regen naming contract: item modifiers use _per_10_seconds (integer)
    # and the sim converts them to _per_second (float) for derived stats and the
    # protocol. Both names must stay consistent across rules, schemas, and goldens;
    # a rename on either side silently breaks the conversion in sim.go.
    item_rollable_stat_keys: set[str] = set()
    for template in item_templates.get("templates", {}).values():
        for roll in template.get("rollable_stats", []):
            item_rollable_stat_keys.add(roll.get("stat", ""))
    shop_stat_weight_keys: set[str] = set()
    for shop_entry in shops.get("shops", {}).values():
        gen = shop_entry.get("generated_offers", {})
        shop_stat_weight_keys.update(gen.get("stat_weights", {}).keys())
    if "health_regen_per_10_seconds" not in item_rollable_stat_keys | shop_stat_weight_keys:
        report.fail("health_regen stat name contract", "health_regen_per_10_seconds not found in item_templates rollable_stats or shop stat_weights — was it renamed?")
    else:
        report.ok("health_regen_per_10_seconds present in item/shop rules")
    progression_formulae = character_progression.get("derived_stats", {})
    if "health_regen_per_second" not in progression_formulae:
        report.fail("health_regen stat name contract", "health_regen_per_second not found in character_progression derived_stats — was it renamed?")
    else:
        report.ok("health_regen_per_second present in character_progression derived_stats")

    validate_item_presentations(
        report,
        assets_dir=ASSETS,
        load_json=load,
        items=items,
        item_templates=item_templates,
        manifest_assets=manifest_assets,
    )

    validate_unique_items_catalog(report, unique_items, item_templates, unique_effects)

    template_item_types = {template.get("item_type") for template in item_templates["templates"].values()}
    if not unique_effect_defs:
        report.fail("unique_effects catalog", "must define at least one unique effect")
    else:
        failed_unique_effects = False
        supported_hooks = {
            "on_hero_damage_dealt",
            "on_basic_attack_hit",
            "on_equip_passive",
            "on_lethal_damage_taken",
            "on_block_or_evade",
            "on_enemy_killed",
            "on_skill_resource_shortfall",
            "on_large_hit_taken",
            "on_continuous_movement_attack",
            "on_projectile_hit_taken",
            "on_repeated_same_target_hit", "on_skill_damage_roll",
        }
        for effect_id, effect in unique_effect_defs.items():
            if effect.get("id") != effect_id:
                report.fail("unique_effects id", f"{effect_id}: id field must match key")
                failed_unique_effects = True
            if effect.get("enabled") is not True or effect.get("status") != "ready":
                report.fail("unique_effects status", f"{effect_id}: v103 effects must be ready and enabled")
                failed_unique_effects = True
            hook = effect.get("hook")
            if hook not in supported_hooks:
                report.fail("unique_effects hook", f"{effect_id}: unsupported hook {hook}")
                failed_unique_effects = True
            compatible_types = set(effect.get("compatible_item_types", []))
            unknown_types = sorted(compatible_types - template_item_types)
            if unknown_types:
                report.fail("unique_effects compatibility", f"{effect_id}: unknown item types {unknown_types}")
                failed_unique_effects = True
            params = effect.get("params", {})
            if effect_id == "everburning_wound":
                required_params = {
                    "status_id",
                    "damage_type",
                    "tick_damage_percent_of_original_hit",
                    "duration_seconds",
                    "tick_interval_seconds",
                }
                missing = sorted(required_params - set(params))
                if missing:
                    report.fail("unique_effects burn params", f"{effect_id}: missing {missing}")
                    failed_unique_effects = True
                elif params.get("status_id") != "burning":
                    report.fail("unique_effects burn params", f"{effect_id}: status_id must be burning")
                    failed_unique_effects = True
                elif params.get("damage_type") != "fire":
                    report.fail("unique_effects burn params", f"{effect_id}: damage_type must be fire")
                    failed_unique_effects = True
                elif params.get("tick_damage_percent_of_original_hit") != 10:
                    report.fail("unique_effects burn params", f"{effect_id}: burn tick percent must be 10")
                    failed_unique_effects = True
                elif params.get("duration_seconds") != 10 or params.get("tick_interval_seconds") != 1:
                    report.fail("unique_effects burn params", f"{effect_id}: burn must tick every second for 10 seconds")
                    failed_unique_effects = True
        if not failed_unique_effects:
            report.ok("unique_effects ready effects define hooks, params, and valid item-type compatibility")


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
