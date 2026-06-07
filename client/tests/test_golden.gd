# GDScript golden-fixture test (ADR-0001 D6, spec acceptance #7).
#
# Consumes the SAME shared/golden fixtures the Go tests consume, proving the
# cross-language rules contract holds. Run with:
#   godot --headless --path client --script res://tests/test_golden.gd
# Exits 0 on success, 1 on failure. Requires no server.
extends SceneTree


func _initialize() -> void:
	var shared := ProjectSettings.globalize_path("res://").path_join("../shared")

	# 1. Damage formula: damage = min + (draw mod span), against combat rules.
	var combat := _read(shared.path_join("rules/combat.v0.json"))
	var dmg := _read(shared.path_join("golden/damage_formula.json"))
	if dmg["player_damage"] != combat["player_damage"]:
		_fail("damage golden player_damage != combat rules")
		return
	var pmin: int = int(combat["player_damage"]["min"])
	var span: int = int(combat["player_damage"]["max"]) - pmin + 1
	for c in dmg["cases"]:
		var got: int = pmin + (int(c["draw"]) % span)
		if got != int(c["expected_damage"]):
			_fail("damage case draw=%d: got %d want %d" % [int(c["draw"]), got, int(c["expected_damage"])])
			return

	# 2. Retaliation formula: same range rule, against the training dummy.
	var monsters := _read(shared.path_join("rules/monsters.v0.json"))
	var retaliation := _read(shared.path_join("golden/retaliation_damage.json"))
	var dummy_retaliation: Dictionary = monsters["monsters"]["training_dummy"]["retaliation_damage"]
	if retaliation["retaliation_damage"] != dummy_retaliation:
		_fail("retaliation golden range != training_dummy rules")
		return
	var rmin: int = int(dummy_retaliation["min"])
	var rspan: int = int(dummy_retaliation["max"]) - rmin + 1
	for c in retaliation["cases"]:
		var got: int = rmin + (int(c["draw"]) % rspan)
		if got != int(c["expected_damage"]):
			_fail("retaliation case draw=%d: got %d want %d" % [int(c["draw"]), got, int(c["expected_damage"])])
			return

	# 3. Equipped weapon damage: same range rule, against item rules.
	var items := _read(shared.path_join("rules/items.v0.json"))
	var weapon_golden := _read(shared.path_join("golden/equipped_weapon_damage.json"))
	var item_def: Dictionary = items["items"][weapon_golden["item_def_id"]]
	if not bool(item_def["equippable"]) or str(item_def["slot"]) != "weapon":
		_fail("equipped weapon golden item is not an equippable weapon")
		return
	if item_def["damage"] != weapon_golden["damage"]:
		_fail("equipped weapon golden range != item rules")
		return
	var wmin: int = int(item_def["damage"]["min"])
	var wspan: int = int(item_def["damage"]["max"]) - wmin + 1
	for c in weapon_golden["cases"]:
		var got: int = wmin + (int(c["draw"]) % wspan)
		if got != int(c["expected_damage"]):
			_fail("equipped weapon case draw=%d: got %d want %d" % [int(c["draw"]), got, int(c["expected_damage"])])
			return

	# 4. Melee reach: same distance formula used by the authoritative sim.
	var reach_golden := _read(shared.path_join("golden/melee_reach.json"))
	for c in reach_golden["cases"]:
		var dist := float(c["distance"])
		var reach := float(c["reach"])
		var target_radius := float(c["target_radius"])
		var got := dist <= reach + target_radius + 0.000001
		if got != bool(c["in_range"]):
			_fail("melee reach case %s: got %s want %s" % [str(c["name"]), str(got), str(c["in_range"])])
			return

	# 5. Loot roll: single-entry table resolves to the expected item.
	var loot := _read(shared.path_join("rules/loot_tables.v0.json"))
	var loot_golden := _read(shared.path_join("golden/loot_roll.json"))
	var entries: Array = loot["loot_tables"][loot_golden["loot_table"]]["entries"]
	if entries.size() != 1 or str(entries[0]["item_def_id"]) != str(loot_golden["expected_item_def_id"]):
		_fail("loot golden mismatch")
		return

	# 6. Auto-path golden references shared navigation/world rules.
	var navigation := _read(shared.path_join("rules/navigation.v0.json"))
	var worlds := _read(shared.path_join("rules/worlds.v0.json"))
	var auto_path := _read(shared.path_join("golden/auto_path.json"))
	if auto_path["navigation"] != navigation:
		_fail("auto_path navigation != navigation rules")
		return
	if float(navigation["cell_size"]) != 1.0:
		_fail("navigation cell_size must be 1.0 for v11 client fixture")
		return
	for c in auto_path["cases"]:
		var world_id := str(c["world_id"])
		if not worlds["worlds"].has(world_id):
			_fail("auto_path references unknown world_id %s" % world_id)
			return
		if str(c["goal_mode"]) != "melee_approach":
			_fail("auto_path case %s must use melee_approach goal_mode" % str(c["name"]))
			return
		if str(c["target_kind"]) != "monster":
			_fail("auto_path case %s must target a monster" % str(c["name"]))
			return

	# 7. Ranged projectile golden references shared item/world/monster rules.
	var ranged_projectile := _read(shared.path_join("golden/ranged_projectile.json"))
	if float(ranged_projectile["constants"]["projectile_radius"]) != 0.10:
		_fail("ranged projectile radius constant mismatch")
		return
	if float(ranged_projectile["constants"]["tick_duration"]) != 0.05:
		_fail("ranged projectile tick duration mismatch")
		return
	if float(ranged_projectile["constants"]["monster_radius"]) != 0.45:
		_fail("ranged projectile monster radius mismatch")
		return
	for c in ranged_projectile["cases"]:
		var world_id := str(c["world_id"])
		var weapon_id := str(c["equipped_weapon"])
		var monster_id := str(c["target_monster_def_id"])
		if not worlds["worlds"].has(world_id):
			_fail("ranged projectile references unknown world_id %s" % world_id)
			return
		if not items["items"].has(weapon_id):
			_fail("ranged projectile references unknown weapon %s" % weapon_id)
			return
		if str(items["items"][weapon_id].get("attack_mode", "melee")) != "ranged":
			_fail("ranged projectile weapon %s is not ranged" % weapon_id)
			return
		if not monsters["monsters"].has(monster_id):
			_fail("ranged projectile references unknown monster %s" % monster_id)
			return

	# 8. Inventory drop golden references shared item/world/navigation rules.
	var inventory_drop := _read(shared.path_join("golden/inventory_drop.json"))
	var drop_world_id := str(inventory_drop["world_id"])
	var drop_item_id := str(inventory_drop["item_def_id"])
	if not worlds["worlds"].has(drop_world_id):
		_fail("inventory_drop references unknown world_id %s" % drop_world_id)
		return
	if not items["items"].has(drop_item_id):
		_fail("inventory_drop references unknown item %s" % drop_item_id)
		return
	if float(inventory_drop["constants"]["loot_drop_radius"]) != 0.35:
		_fail("inventory_drop loot_drop_radius mismatch")
		return
	if float(inventory_drop["constants"]["player_radius"]) != 0.45:
		_fail("inventory_drop player_radius mismatch")
		return
	if float(inventory_drop["constants"]["drop_step"]) != float(navigation["cell_size"]):
		_fail("inventory_drop drop_step != navigation.cell_size")
		return

	# 9. Use consumable golden references shared consumable heal rules.
	var use_consumable := _read(shared.path_join("golden/use_consumable.json"))
	var use_item_def: Dictionary = items["items"][use_consumable["item_def_id"]]
	if str(use_item_def.get("category", "")) != "consumable":
		_fail("use_consumable golden item is not consumable")
		return
	if use_item_def["heal"] != use_consumable["heal"]:
		_fail("use_consumable golden heal != item rules")
		return
	var hmin: int = int(use_consumable["heal"]["min"])
	var hspan: int = int(use_consumable["heal"]["max"]) - hmin + 1
	for c in use_consumable["cases"]:
		var rolled: int = hmin + (int(c["draw"]) % hspan)
		var capped: int = mini(rolled, int(c["player_max_hp"]) - int(c["player_hp"]))
		if capped != int(c["expected_heal"]) or int(c["player_hp"]) + capped != int(c["expected_player_hp"]):
			_fail("use_consumable case %s heal cap mismatch" % str(c["name"]))
			return

	# 10. Monster chase golden references shared navigation/world/monster rules.
	var monster_chase := _read(shared.path_join("golden/monster_chase.json"))
	if monster_chase["navigation"] != navigation:
		_fail("monster_chase navigation != navigation rules")
		return
	for c in monster_chase["cases"]:
		var world_id := str(c.get("world_id", monster_chase.get("world_id", "")))
		if not worlds["worlds"].has(world_id):
			_fail("monster_chase references unknown world_id %s" % world_id)
			return
		for entity in worlds["worlds"][world_id]["entities"]:
			if str(entity.get("type", "")) != "monster":
				continue
			var monster_id := str(entity["monster_def_id"])
			if str(monsters["monsters"][monster_id].get("behavior", "static")) != "chase":
				_fail("monster_chase world %s uses non-chase monster %s" % [world_id, monster_id])
				return

	# 11. Dungeon stairs golden references generated floor rules and level labels.
	var dungeon_generation := _read(shared.path_join("rules/dungeon_generation.v0.json"))
	var dungeon_stairs := _read(shared.path_join("golden/dungeon_stairs.json"))
	if float(dungeon_generation["floor_size"]["width"]) != 32.0:
		_fail("dungeon_generation width mismatch")
		return
	if float(dungeon_generation["floor_size"]["height"]) != 20.0:
		_fail("dungeon_generation height mismatch")
		return
	if str(dungeon_generation["level_names"]["-1"]) != "Entry Hall":
		_fail("dungeon level -1 name mismatch")
		return
	if float(dungeon_generation["teleporter_placement"]["min_stair_distance"]) != 4.0:
		_fail("dungeon teleporter placement mismatch")
		return
	var level1: Dictionary = dungeon_stairs["levels"]["-1"]
	var level2: Dictionary = dungeon_stairs["levels"]["-2"]
	var town: Dictionary = dungeon_stairs["levels"]["0"]
	if not _vec2_equals(town["stairs_down"], 8.0, 10.0):
		_fail("town stairs_down mismatch")
		return
	if not _vec2_equals(town["teleporter"], 4.0, 13.0):
		_fail("town teleporter mismatch")
		return
	if not _vec2_equals(level1["stairs_up"], 4.0, 10.0):
		_fail("dungeon level -1 stairs_up mismatch")
		return
	if not _vec2_equals(level1["stairs_down"], 14.0, 18.0):
		_fail("dungeon level -1 stairs_down mismatch")
		return
	if not _vec2_equals(level1["teleporter"], 14.0, 12.0):
		_fail("dungeon level -1 teleporter mismatch")
		return
	if not _vec2_equals(level2["stairs_up"], 9.0, 11.0) or not _vec2_equals(level2["stairs_down"], 28.0, 14.0):
		_fail("dungeon level -2 stairs mismatch")
		return
	if not _vec2_equals(level2["teleporter"], 10.0, 2.0):
		_fail("dungeon level -2 teleporter mismatch")
		return
	var dungeon_loot: Array = level2["loot"]
	if dungeon_loot.size() != 1:
		_fail("dungeon level -2 loot count mismatch")
		return
	if str(dungeon_loot[0]["item_def_id"]) != "training_badge" or not _vec2_equals(dungeon_loot[0]["position"], 16.0, 11.0):
		_fail("dungeon level -2 coin loot mismatch")
		return
	var fallback := str(dungeon_generation["default_level_name_template"]).replace("{n}", str(abs(-9)))
	if fallback != "Depth 9":
		_fail("dungeon fallback level name mismatch")
		return

	# 12. Dungeon teleporters golden matches stairs seed and pinned travel outcome.
	var dungeon_teleporters := _read(shared.path_join("golden/dungeon_teleporters.json"))
	if str(dungeon_teleporters["seed"]) != str(dungeon_stairs["seed"]):
		_fail("dungeon_teleporters seed mismatch")
		return
	var tp_outcome: Dictionary = dungeon_teleporters["discover_descend_teleport"]
	if int(tp_outcome["expected_level"]) != -1:
		_fail("dungeon teleporters expected level mismatch")
		return
	if not _vec2_equals(tp_outcome["expected_player_position"], 14.0, 12.0):
		_fail("dungeon teleporters expected player position mismatch")
		return

	# 13. Dungeon monster attack golden references shared proactive attack rules.
	var dungeon_monster_attack := _read(shared.path_join("golden/dungeon_monster_attack.json"))
	var dungeon_mob_id := str(dungeon_monster_attack["monster_def_id"])
	if dungeon_mob_id != str(dungeon_generation["monster_placement"]["monster_def_id"]):
		_fail("dungeon monster attack monster_def_id mismatch")
		return
	var dungeon_mob: Dictionary = monsters["monsters"][dungeon_mob_id]
	if not dungeon_mob.has("attack_damage") or not dungeon_mob.has("attack_cooldown_ticks"):
		_fail("dungeon monster missing proactive attack fields")
		return
	var attack_damage: Dictionary = dungeon_mob["attack_damage"]
	var pinned_damage := int(dungeon_monster_attack["damage"])
	if pinned_damage < int(attack_damage["min"]) or pinned_damage > int(attack_damage["max"]):
		_fail("dungeon monster attack damage outside rules")
		return
	if int(dungeon_monster_attack["player_hp_after"]) != 10 - pinned_damage:
		_fail("dungeon monster attack hp mismatch")
		return

	# 14. Waypoint panel golden matches client layout constants.
	var waypoint_panel := _read(shared.path_join("golden/waypoint_panel.json"))
	const WaypointPanelConfig := preload("res://scripts/waypoint_panel_config.gd")
	if WaypointPanelConfig.SCROLL_MAX_VISIBLE_ROWS != int(waypoint_panel["scroll_max_visible_rows"]):
		_fail("waypoint panel scroll max rows mismatch")
		return
	if WaypointPanelConfig.SCROLL_VIEWPORT_UNIT_PX != int(waypoint_panel["scroll_viewport_unit_px"]):
		_fail("waypoint panel viewport unit mismatch")
		return

	# 15. Item roll golden references shared item template fields for tooltip display.
	var item_templates := _read(shared.path_join("rules/item_templates.v0.json"))
	var item_rolls := _read(shared.path_join("golden/item_rolls.json"))
	var template_id := str(item_rolls["template_id"])
	if not item_templates["templates"].has(template_id):
		_fail("item_rolls references unknown template")
		return
	var template: Dictionary = item_templates["templates"][template_id]
	for c in item_rolls["cases"]:
		var expected: Dictionary = c["expected"]
		if str(expected["item_template_id"]) != template_id:
			_fail("item_rolls item_template_id mismatch")
			return
		if not str(expected["display_name"]).ends_with(str(template["name"])):
			_fail("item_rolls display_name missing template name")
			return
		var stats: Dictionary = expected["stats"]
		if not stats.has("damage_min") or not stats.has("damage_max"):
			_fail("item_rolls missing damage stats")
			return
		if int(stats["damage_max"]) < int(stats["damage_min"]):
			_fail("item_rolls damage range invalid")
			return
		if expected["requirements"] != template["requirements"]:
			_fail("item_rolls requirements mismatch")
			return
		if (expected["effect_ids"] as Array).size() != 0:
			_fail("item_rolls effect_ids should be empty in v23")
			return

	print("[gdtest] PASS: consumed shared/golden fixtures (damage_formula, retaliation_damage, equipped_weapon_damage, melee_reach, loot_roll, auto_path, ranged_projectile, inventory_drop, use_consumable, monster_chase, dungeon_stairs, dungeon_teleporters, dungeon_monster_attack, waypoint_panel, item_rolls)")
	quit(0)


func _read(path: String) -> Dictionary:
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		_fail("cannot open %s" % path)
		return {}
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		_fail("invalid JSON in %s" % path)
		return {}
	return parsed


func _vec2_equals(value: Dictionary, x: float, y: float) -> bool:
	return is_equal_approx(float(value.get("x", NAN)), x) and is_equal_approx(float(value.get("y", NAN)), y)


func _fail(msg: String) -> void:
	printerr("[gdtest] FAIL: ", msg)
	quit(1)
