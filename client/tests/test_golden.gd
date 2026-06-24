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
	if not bool(item_def["equippable"]) or str(item_def["slot"]) != "main_hand":
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
	if float(dungeon_generation["floor_size"]["width"]) != 100.0:
		_fail("dungeon_generation width mismatch")
		return
	if float(dungeon_generation["floor_size"]["height"]) != 50.0:
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
	var level3: Dictionary = dungeon_stairs["levels"]["-3"]
	var town: Dictionary = dungeon_stairs["levels"]["0"]
	if not _vec2_equals(town["stairs_down"], 11.0, 8.0):
		_fail("town stairs_down mismatch")
		return
	if not _vec2_equals(town["teleporter"], 2.0, 12.0):
		_fail("town teleporter mismatch")
		return
	if not _vec2_equals(level1["stairs_up"], 4.0, 10.0):
		_fail("dungeon level -1 stairs_up mismatch")
		return
	if not _vec2_equals(level1["stairs_down"], 27.0, 22.0):
		_fail("dungeon level -1 stairs_down mismatch")
		return
	if level1.has("teleporter"):
		_fail("dungeon level -1 should not have teleporter")
		return
	if not _vec2_equals(level2["stairs_up"], 24.0, 43.0) or not _vec2_equals(level2["stairs_down"], 30.0, 2.0):
		_fail("dungeon level -2 stairs mismatch")
		return
	if level2.has("teleporter"):
		_fail("dungeon level -2 should not have teleporter")
		return
	if not _vec2_equals(level3["stairs_up"], 14.0, 30.0) or not _vec2_equals(level3["stairs_down"], 69.0, 34.0):
		_fail("dungeon level -3 stairs mismatch")
		return
	if not _vec2_equals(level3["teleporter"], 60.0, 22.0):
		_fail("dungeon level -3 teleporter mismatch")
		return
	var dungeon_loot: Array = level2["loot"]
	if dungeon_loot.size() != 0:
		_fail("dungeon level -2 loot count mismatch")
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
	if int(tp_outcome["expected_level"]) != -3:
		_fail("dungeon teleporters expected level mismatch")
		return
	if not _vec2_equals(tp_outcome["expected_player_position"], 61.0, 22.0):
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
	if int(dungeon_monster_attack["attack_cooldown_ticks"]) != int(dungeon_mob["attack_cooldown_ticks"]):
		_fail("dungeon monster attack cooldown mismatch")
		return
	var attack_damage: Dictionary = dungeon_mob["attack_damage"]
	var pinned_damage := int(dungeon_monster_attack["damage"])
	var damage_matches_rarity := false
	for rarity in dungeon_generation["monster_rarities"]:
		var damage_mult := float(rarity["damage_multiplier"])
		var scaled_min := _round_positive(float(attack_damage["min"]) * damage_mult)
		var scaled_max := _round_positive(float(attack_damage["max"]) * damage_mult)
		if pinned_damage >= scaled_min and pinned_damage <= scaled_max:
			damage_matches_rarity = true
			break
	if not damage_matches_rarity:
		_fail("dungeon monster attack damage outside rules")
		return
	if int(dungeon_monster_attack["player_hp_after"]) != 10 - pinned_damage:
		_fail("dungeon monster attack hp mismatch")
		return

	# 14. Monster rarity golden mirrors shared dungeon-generation presentation/scaling data.
	var monster_rarity := _read(shared.path_join("golden/monster_rarity.json"))
	var expected_rarity_ids := ["common", "champion", "rare", "unique"]
	for i in range(expected_rarity_ids.size()):
		var rarity: Dictionary = dungeon_generation["monster_rarities"][i]
		if str(rarity["id"]) != expected_rarity_ids[i]:
			_fail("monster_rarity id order mismatch")
			return
		if int(rarity["weight"]) <= 0:
			_fail("monster_rarity weight must be positive")
			return
		if int(rarity["loot_depth_offset"]) < 0:
			_fail("monster_rarity loot offset must be non-negative")
			return
	for c in monster_rarity["effective_depth_cases"]:
		var rarity_offset := -1
		for rarity in dungeon_generation["monster_rarities"]:
			if str(rarity["id"]) == str(c["rarity"]):
				rarity_offset = int(rarity["loot_depth_offset"])
				break
		if rarity_offset < 0:
			_fail("monster_rarity effective depth unknown rarity")
			return
		var expected_depth := absi(int(c["level"])) + rarity_offset
		if expected_depth < 1:
			_fail("monster_rarity effective depth invalid")
			return
		if not loot["loot_tables"].has(str(c["expected_monster_loot_table"])):
			_fail("monster_rarity expected loot table missing")
			return

	# 15. Waypoint panel golden matches client layout constants.
	var waypoint_panel := _read(shared.path_join("golden/waypoint_panel.json"))
	const WaypointPanelConfig := preload("res://scripts/waypoint_panel_config.gd")
	if WaypointPanelConfig.SCROLL_MAX_VISIBLE_ROWS != int(waypoint_panel["scroll_max_visible_rows"]):
		_fail("waypoint panel scroll max rows mismatch")
		return
	if WaypointPanelConfig.SCROLL_VIEWPORT_UNIT_PX != int(waypoint_panel["scroll_viewport_unit_px"]):
		_fail("waypoint panel viewport unit mismatch")
		return

	# 16. Item roll golden references shared item template fields for tooltip display.
	var item_templates := _read(shared.path_join("rules/item_templates.v0.json"))
	var item_rolls := _read(shared.path_join("golden/item_rolls.json"))
	var unique_effects := _read(shared.path_join("rules/unique_effects.v0.json"))
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
		var effect_ids := expected["effect_ids"] as Array
		if str(expected["rarity"]) == "unique":
			if effect_ids.size() != 1:
				_fail("unique item_rolls must attach one effect_id")
				return
			if not (unique_effects["effects"] as Dictionary).has(str(effect_ids[0])):
				_fail("unique item_rolls references unknown effect_id")
				return
			var effect: Dictionary = unique_effects["effects"][str(effect_ids[0])]
			var item_type_name := str(template.get("item_type", "")).replace("_", " ").capitalize()
			var expected_unique_name := "%s of %s" % [item_type_name, str(effect.get("display_name", ""))]
			if str(expected["display_name"]) != expected_unique_name:
				_fail("unique item_rolls display_name mismatch")
				return
		elif effect_ids.size() != 0:
			_fail("non-unique item_rolls effect_ids must be empty")
			return
		elif not str(expected["display_name"]).ends_with(str(template["name"])):
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

	# 17. Treasure class golden references shared item/template reward sources.
	var treasure_classes := _read(shared.path_join("rules/treasure_classes.v0.json"))
	var treasure_class_rolls := _read(shared.path_join("golden/treasure_class_rolls.json"))
	var treasure_class_id := str(treasure_class_rolls["treasure_class_id"])
	if not treasure_classes["classes"].has(treasure_class_id):
		_fail("treasure_class_rolls references unknown treasure class")
		return
	for c in treasure_class_rolls["cases"]:
		for drop in c["expected_drops"]:
			var item_def_id := str(drop.get("item_def_id", ""))
			var item_template_id := str(drop.get("item_template_id", ""))
			if item_def_id == "" and item_template_id == "":
				_fail("treasure_class_rolls expected drop missing item source")
				return
			if item_def_id != "" and not items["items"].has(item_def_id):
				_fail("treasure_class_rolls references unknown item %s" % item_def_id)
				return
			if item_template_id != "" and not item_templates["templates"].has(item_template_id):
				_fail("treasure_class_rolls references unknown template %s" % item_template_id)
				return

	# 18. Guarded chest golden references shared dungeon/interactable/loot rules.
	var interactables := _read(shared.path_join("rules/interactables.v0.json"))
	var guarded_chest := _read(shared.path_join("golden/guarded_chest_generation.json"))
	var chest_def_id := str(dungeon_generation["chest_placement"]["interactable_def_id"])
	if not interactables["interactables"].has(chest_def_id):
		_fail("guarded_chest references unknown interactable")
		return
	if str(interactables["interactables"][chest_def_id]["initial_state"]) != "closed":
		_fail("guarded_chest interactable must start closed")
		return
	var guarded_depth: int = absi(int(guarded_chest["level"]))
	var chest_loot_table := ""
	for band in dungeon_generation["loot_bands"]:
		var min_depth := int(band["min_depth"])
		var max_depth: Variant = band["max_depth"]
		if guarded_depth >= min_depth and (max_depth == null or guarded_depth <= int(max_depth)):
			chest_loot_table = str(band["chest_loot_table"])
			break
	if chest_loot_table == "":
		_fail("guarded_chest missing depth loot band")
		return
	if not loot["loot_tables"].has(chest_loot_table):
		_fail("guarded_chest references unknown loot table")
		return
	if not loot["loot_tables"][chest_loot_table].has("treasure_class_id"):
		_fail("guarded_chest loot table must resolve treasure class")
		return
	for c in guarded_chest["cases"]:
		var expected_chest = c["expected_chest"]
		if expected_chest == null:
			continue
		if str(expected_chest["interactable_def_id"]) != chest_def_id:
			_fail("guarded_chest interactable mismatch")
			return
		if str(expected_chest["loot_table"]) != chest_loot_table:
			_fail("guarded_chest loot table mismatch")
			return

	# 19. Character progression golden mirrors display-side derived formulas.
	var progression_rules := _read(shared.path_join("rules/character_progression.v0.json"))
	var progression_golden := _read(shared.path_join("golden/character_progression.json"))
	var progression_combat_rules := _read(shared.path_join("rules/combat.v0.json"))
	var progression_main_config := _read(shared.path_join("rules/main_config.v0.json"))
	var progression_base_move_speed := float((progression_main_config.get("gameplay", {}) as Dictionary).get("base_movement_speed", 1.0))
	if progression_rules["base_stats"] != progression_golden["base_stats"]:
		_fail("character_progression base stats mismatch")
		return
	for c in progression_golden["cases"]:
		var stats: Dictionary = c["base_stats"].duplicate(true)
		if c.has("allocated_stat"):
			stats[str(c["allocated_stat"])] = int(stats[str(c["allocated_stat"])]) + int(c["allocated_points"])
		var expected: Dictionary = c["expected"]
		for key in progression_rules["derived_stats"].keys():
			var got := _eval_progression_formula(progression_rules["derived_stats"][key], stats)
			if key == "damage_min":
				got += float(progression_combat_rules["player_damage"]["min"])
			elif key == "damage_max":
				got += float(progression_combat_rules["player_damage"]["max"])
			elif key == "movement_speed":
				# movement_speed formula yields the DEX multiplier; scale by classBase
				# to match Go's playerEffectiveMovementSpeed = classBase * dex_mult * gear%
				var _case_class := str(c.get("character_class", ""))
				var _class_base := progression_base_move_speed
				if _case_class != "":
					var _classes := progression_rules.get("classes", {}) as Dictionary
					if _classes.has(_case_class):
						var _cms := float((_classes[_case_class] as Dictionary).get("base_movement_speed", 0.0))
						if _cms > 0.0:
							_class_base = _cms
				got *= _class_base
			var want := float(expected["derived_stats"][key])
			if not is_equal_approx(got, want):
				_fail("character_progression case %s %s got %.4f want %.4f" % [str(c["name"]), str(key), got, want])
				return
		if expected["derived_stats"].has("attack_interval_ticks"):
			var attack_speed := float(expected["derived_stats"]["attack_speed"])
			var got_interval := _attack_interval_ticks(progression_combat_rules, attack_speed)
			var want_interval := int(expected["derived_stats"]["attack_interval_ticks"])
			if got_interval != want_interval:
				_fail("character_progression case %s attack_interval_ticks got %d want %d" % [str(c["name"]), got_interval, want_interval])
				return

	# 20. Skill points and Magic Bolt golden mirrors v44 shared rules.
	var skill_golden := _read(shared.path_join("golden/skill_points_and_magic_bolt.json"))
	var skill_rules := _read(shared.path_join("rules/skills.v0.json"))
	if int(skill_golden["progression"]["points_per_level"]) != int(progression_rules["points_per_level"]):
		_fail("skill golden points_per_level mismatch")
		return
	if skill_golden["progression"]["skill_points"] != progression_rules["skill_points"]:
		_fail("skill golden skill point cadence mismatch")
		return
	for c in skill_golden["progression"]["level_cases"]:
		var level := int(c["level"])
		var expected_stats := (level - 1) * int(progression_rules["points_per_level"])
		var expected_skills := _skill_points_for_level(level, progression_rules["skill_points"])
		if expected_stats != int(c["expected_unspent_stat_points"]) or expected_skills != int(c["expected_unspent_skill_points"]):
			_fail("skill golden level %d cadence mismatch" % level)
			return
	if int(skill_golden["attack_speed"]["base_attack_interval_ticks"]) != int(progression_combat_rules["base_attack_interval_ticks"]):
		_fail("skill golden base attack interval mismatch")
		return
	if float(skill_golden["attack_speed"]["min_effective_attack_speed"]) != float(progression_combat_rules["min_effective_attack_speed"]):
		_fail("skill golden min attack speed mismatch")
		return
	if float(skill_golden["attack_speed"]["max_effective_attack_speed"]) != float(progression_combat_rules["max_effective_attack_speed"]):
		_fail("skill golden max attack speed mismatch")
		return
	var skill_id := str(skill_golden["skill"]["skill_id"])
	if not skill_rules["skills"].has(skill_id):
		_fail("skill golden references unknown skill %s" % skill_id)
		return
	var magic_bolt: Dictionary = skill_rules["skills"][skill_id]
	if int(skill_golden["skill"]["max_rank"]) != int(magic_bolt["max_rank"]):
		_fail("skill golden max rank mismatch")
		return
	if skill_golden["skill"].get("requirements", {}) != magic_bolt.get("requirements", {}):
		_fail("skill golden requirements mismatch")
		return
	if int(magic_bolt["requirements"]["stats"]["magic"]) != 5:
		_fail("magic bolt magic requirement mismatch")
		return
	if int(magic_bolt["requirements"].get("level_per_rank", 0)) != 1 or int(magic_bolt["requirements"].get("stats_per_rank", {}).get("magic", 0)) != 3:
		_fail("magic bolt per-rank requirement mismatch")
		return
	var cooldown_multiplier := float(magic_bolt["cooldown"]["multiplier"])
	for c in skill_golden["attack_speed"]["cases"]:
		var effective := float(c["dex_attack_speed"]) * float(c["weapon_attack_speed"]) * (1.0 + float(c["item_attack_speed_percent"]) / 100.0)
		effective = clampf(effective, float(progression_combat_rules["min_effective_attack_speed"]), float(progression_combat_rules["max_effective_attack_speed"]))
		if not is_equal_approx(effective, float(c["expected_effective_attack_speed"])):
			_fail("skill golden attack speed case %s effective mismatch" % str(c["name"]))
			return
		var interval := _attack_interval_ticks(progression_combat_rules, effective)
		var cooldown := maxi(1, int(ceil(float(interval) * cooldown_multiplier)))
		if interval != int(c["expected_attack_interval_ticks"]) or cooldown != int(c["expected_magic_bolt_cooldown_ticks"]):
			_fail("skill golden attack speed case %s interval/cooldown got %d/%d want %d/%d" % [str(c["name"]), interval, cooldown, int(c["expected_attack_interval_ticks"]), int(c["expected_magic_bolt_cooldown_ticks"])])
			return
	for c in skill_golden["skill"]["rank_requirement_cases"]:
		var req_rank := int(c["rank"])
		var expected_requirements := _skill_requirements_for_rank(magic_bolt["requirements"], req_rank)
		if int(c["level"]) != int(expected_requirements["level"]) or not _same_int_dictionary(c["stats"], expected_requirements["stats"]):
			_fail("skill golden rank %d requirement mismatch" % req_rank)
			return
	for c in skill_golden["skill"]["rank_cases"]:
		var rank := int(c["rank"])
		if _skill_mana_cost(magic_bolt, rank) != int(c["mana_cost"]):
			_fail("skill golden rank %d mana mismatch" % rank)
			return
		var expected_damage: Dictionary = c["damage"]
		if _skill_damage_min(magic_bolt, rank) != int(expected_damage["min"]) or _skill_damage_max(magic_bolt, rank) != int(expected_damage["max"]):
			_fail("skill golden rank %d damage mismatch" % rank)
			return

	# 21. Combat stat effects golden is available to the display client.
	var combat_rules := _read(shared.path_join("rules/combat.v0.json"))
	var combat_effects := _read(shared.path_join("golden/combat_stat_effects.json"))
	if int(combat_rules["minimum_damage"]) != int(combat_effects["combat"]["minimum_damage"]):
		_fail("combat_stat_effects minimum_damage mismatch")
		return
	if float(combat_rules["block_cap_percent"]) != float(combat_effects["combat"]["block_cap_percent"]):
		_fail("combat_stat_effects block_cap mismatch")
		return
	var outcomes := {}
	for c in combat_effects["cases"]:
		outcomes[str(c["outcome"])] = true
	for required in ["hit", "crit", "miss", "block"]:
		if not outcomes.has(required):
			_fail("combat_stat_effects missing outcome %s" % required)
			return
	var breakdowns: Array = combat_effects.get("stat_breakdowns", [])
	var has_armor := false
	var has_capped_block := false
	for row in breakdowns:
		if str(row.get("key", "")) == "armor" and float(row.get("value", 0.0)) >= 1.0:
			has_armor = true
		if str(row.get("key", "")) == "block_percent" and row.get("cap", null) != null and float(row.get("value", 0.0)) <= float(row.get("cap", 0.0)):
			has_capped_block = true
	if not has_armor or not has_capped_block:
		_fail("combat_stat_effects stat breakdown mismatch")
		return

	# 22. Shop pricing/offers goldens are available to the display client.
	var shops := _read(shared.path_join("rules/shops.v0.json"))
	var shop_pricing := _read(shared.path_join("golden/shop_pricing.json"))
	var shop_offers := _read(shared.path_join("golden/shop_offers.json"))
	var shop_stock_lifecycle := _read(shared.path_join("golden/shop_stock_lifecycle.json"))
	if not shops["shops"].has("town_vendor"):
		_fail("shops missing town_vendor")
		return
	var town_vendor: Dictionary = shops["shops"]["town_vendor"]
	if str(shop_pricing["shop_id"]) != "town_vendor" or str(shop_offers["shop_id"]) != "town_vendor":
		_fail("shop goldens must target town_vendor")
		return
	if (town_vendor["fixed_offers"] as Array).size() != 2:
		_fail("town_vendor fixed offer count mismatch")
		return
	for offer in town_vendor["fixed_offers"]:
		var item_id := str(offer["item_def_id"])
		if not items["items"].has(item_id):
			_fail("town_vendor fixed offer unknown item %s" % item_id)
			return
		if int(offer["buy_price"]) <= 0:
			_fail("town_vendor fixed offer has non-positive price")
			return
	var pricing_cases: Array = shop_pricing["cases"]
	if pricing_cases.size() < 5:
		_fail("shop_pricing must cover fixed/common/magic/rare cases")
		return
	for pricing_case in pricing_cases:
		if int(pricing_case["expected"]["buy_price"]) <= 0 or int(pricing_case["expected"]["sell_price"]) <= 0:
			_fail("shop_pricing case %s has non-positive price" % str(pricing_case["name"]))
			return
	for offer_case in shop_offers["cases"]:
		var expected: Array = offer_case["expected"]
		if expected.size() != int(offer_case["expected_offer_count"]):
			_fail("shop_offers case %s count mismatch" % str(offer_case["name"]))
			return
		for offer in expected:
			var offer_template_id := str(offer["item_template_id"])
			if not item_templates["templates"].has(offer_template_id):
				_fail("shop_offers references unknown template %s" % offer_template_id)
				return
			if not str(offer["offer_id"]).begins_with("generated:depth"):
				_fail("shop_offers offer_id must be generated depth id")
				return
			if int(offer["buy_price"]) <= 0:
				_fail("shop_offers buy_price must be positive")
				return
	if str(shop_stock_lifecycle["shop_id"]) != "town_vendor":
		_fail("shop_stock_lifecycle must target town_vendor")
		return
	var lifecycle_generated: Dictionary = shop_stock_lifecycle["generated_stock"]
	var generated_rules: Dictionary = town_vendor["generated_offers"]
	for key in ["offer_count", "source", "source_depth_policy", "refresh_on", "max_rarity"]:
		if lifecycle_generated[key] != generated_rules[key]:
			_fail("shop_stock_lifecycle generated %s mismatch" % str(key))
			return
	if str(lifecycle_generated["max_rarity"]) != "rare":
		_fail("shop_stock_lifecycle max rarity must be rare")
		return
	for source_case in lifecycle_generated["cases"]:
		if int(source_case["expected_min_source_depth"]) > int(source_case["expected_max_source_depth"]):
			_fail("shop_stock_lifecycle source-depth bounds inverted")
			return
	var finite_stock: Dictionary = shop_stock_lifecycle["finite_stock"]
	if int(finite_stock["initial_generated_count"]) != int(generated_rules["offer_count"]):
		_fail("shop_stock_lifecycle initial stock count mismatch")
		return
	if int(finite_stock["after_generated_purchase_count"]) != int(generated_rules["offer_count"]) - 1:
		_fail("shop_stock_lifecycle purchase count mismatch")
		return
	if int(finite_stock["fixed_offer_count"]) != (town_vendor["fixed_offers"] as Array).size():
		_fail("shop_stock_lifecycle fixed count mismatch")
		return
	var buyback: Dictionary = shop_stock_lifecycle["buyback"]
	if not bool(buyback["enabled"]) or str(buyback["scope"]) != "session_town_visit":
		_fail("shop_stock_lifecycle buyback scope mismatch")
		return
	if not str(buyback["offer_id"]).begins_with("buyback:"):
		_fail("shop_stock_lifecycle buyback offer id mismatch")
		return
	if int(buyback["buy_price"]) <= 0 or int(buyback["sell_price"]) <= 0:
		_fail("shop_stock_lifecycle buyback price must be positive")
		return
	if bool(buyback["persisted"]):
		_fail("shop_stock_lifecycle buyback must be session-local")
		return

	# 23. Equipment requirements golden mirrors shared item requirements and client-readable status rows.
	var equipment_requirements := _read(shared.path_join("golden/equipment_requirements.json"))
	var req_template_id := str(equipment_requirements["template_id"])
	if not item_templates["templates"].has(req_template_id):
		_fail("equipment_requirements references unknown template")
		return
	var req_template: Dictionary = item_templates["templates"][req_template_id]
	var expected_requirements: Dictionary = equipment_requirements["requirements"]
	if req_template["requirements"] != expected_requirements:
		_fail("equipment_requirements requirements mismatch")
		return
	for case_name in ["fresh_character", "after_allocation"]:
		var character_case: Dictionary = equipment_requirements[case_name]
		var status_rows: Array = character_case["status"]
		if status_rows.size() != expected_requirements.keys().size():
			_fail("equipment_requirements %s status count mismatch" % str(case_name))
			return
		var all_met := true
		for row in status_rows:
			var rec := row as Dictionary
			var stat := str(rec["stat"])
			if not expected_requirements.has(stat):
				_fail("equipment_requirements %s unknown stat %s" % [str(case_name), stat])
				return
			var required := int(expected_requirements[stat])
			var current := int(character_case["level"]) if stat == "level" else int(character_case["base_stats"][stat])
			var met := current >= required
			if int(rec["required"]) != required or int(rec["current"]) != current or bool(rec["met"]) != met:
				_fail("equipment_requirements %s status mismatch for %s" % [str(case_name), stat])
				return
			if not met:
				all_met = false
		if bool(character_case["requirements_met"]) != all_met:
			_fail("equipment_requirements %s aggregate mismatch" % str(case_name))
			return
	if str(equipment_requirements["expected_reject"]) != "requirements_not_met":
		_fail("equipment_requirements expected reject mismatch")
		return

	print("[gdtest] PASS: consumed shared/golden fixtures (damage_formula, retaliation_damage, equipped_weapon_damage, melee_reach, loot_roll, auto_path, ranged_projectile, inventory_drop, use_consumable, monster_chase, dungeon_stairs, dungeon_teleporters, dungeon_monster_attack, monster_rarity, waypoint_panel, item_rolls, treasure_class_rolls, guarded_chest_generation, character_progression, skill_points_and_magic_bolt, combat_stat_effects, shop_pricing, shop_offers, equipment_requirements)")
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


func _round_positive(value: float) -> int:
	return maxi(1, int(floor(value + 0.5)))


func _skill_points_for_level(level: int, cadence: Dictionary) -> int:
	var first := int(cadence["first_grant_level"])
	if level < first:
		return 0
	var every := maxi(1, int(cadence["grant_every_levels"]))
	var grants := int(floor(float(level - first) / float(every))) + 1
	return grants * int(cadence["points_per_grant"])


func _attack_interval_ticks(combat: Dictionary, effective_attack_speed: float) -> int:
	var min_speed := float(combat["min_effective_attack_speed"])
	var max_speed := float(combat["max_effective_attack_speed"])
	var speed := clampf(effective_attack_speed, min_speed, max_speed)
	if speed <= 0.0:
		speed = 1.0
	return maxi(1, int(ceil(float(combat["base_attack_interval_ticks"]) / speed)))


func _skill_mana_cost(skill: Dictionary, rank: int) -> int:
	var mana: Dictionary = skill["cost"]["mana"]
	return maxi(0, int(mana["base"]) + int(mana["per_rank"]) * maxi(0, rank - 1))


func _skill_requirements_for_rank(requirements: Dictionary, rank: int) -> Dictionary:
	var rank_offset := maxi(0, rank - 1)
	var stats: Dictionary = {}
	var base_stats: Dictionary = requirements.get("stats", {})
	var stats_per_rank: Dictionary = requirements.get("stats_per_rank", {})
	for stat in ["str", "dex", "vit", "magic"]:
		var required := int(base_stats.get(stat, 0)) + int(stats_per_rank.get(stat, 0)) * rank_offset
		if required > 0:
			stats[stat] = required
	return {
		"level": int(requirements.get("level", 0)) + int(requirements.get("level_per_rank", 0)) * rank_offset,
		"stats": stats,
	}


func _same_int_dictionary(a: Dictionary, b: Dictionary) -> bool:
	if a.keys().size() != b.keys().size():
		return false
	for key in b.keys():
		if not a.has(key) or int(a[key]) != int(b[key]):
			return false
	return true


func _skill_damage_min(skill: Dictionary, rank: int) -> int:
	var damage: Dictionary = skill["damage"]
	return maxi(0, int(damage["min_base"]) + int(damage["min_per_rank"]) * maxi(0, rank - 1))


func _skill_damage_max(skill: Dictionary, rank: int) -> int:
	var damage: Dictionary = skill["damage"]
	var min_damage := _skill_damage_min(skill, rank)
	return maxi(min_damage, int(damage["max_base"]) + int(damage["max_per_rank"]) * maxi(0, rank - 1))


func _eval_progression_formula(formula: Dictionary, stats: Dictionary) -> float:
	var value := float(formula.get("base", 0.0))
	if str(formula.get("type", "")) == "logarithmic":
		var stat := str(formula.get("stat", ""))
		var raw := maxf(0.0, float(stats.get(stat, 0)) - float(formula.get("offset", 0.0)))
		value += float(formula.get("scale", 0.0)) * (log(1.0 + raw) / log(1.0 + float(formula.get("denominator", 1.0))))
	else:
		value += float(formula.get("per_str", 0.0)) * float(stats.get("str", 0))
		value += float(formula.get("per_dex", 0.0)) * float(stats.get("dex", 0))
		value += float(formula.get("per_vit", 0.0)) * float(stats.get("vit", 0))
		value += float(formula.get("per_magic", 0.0)) * float(stats.get("magic", 0))
	if formula.has("min"):
		value = maxf(value, float(formula["min"]))
	if formula.has("max"):
		value = minf(value, float(formula["max"]))
	return value


func _fail(msg: String) -> void:
	printerr("[gdtest] FAIL: ", msg)
	quit(1)
