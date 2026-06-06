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

	print("[gdtest] PASS: consumed shared/golden fixtures (damage_formula, retaliation_damage, equipped_weapon_damage, melee_reach, loot_roll, auto_path, ranged_projectile, inventory_drop)")
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


func _fail(msg: String) -> void:
	printerr("[gdtest] FAIL: ", msg)
	quit(1)
