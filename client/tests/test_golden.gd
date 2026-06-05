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

	# 3. Loot roll: single-entry table resolves to the expected item.
	var loot := _read(shared.path_join("rules/loot_tables.v0.json"))
	var loot_golden := _read(shared.path_join("golden/loot_roll.json"))
	var entries: Array = loot["loot_tables"][loot_golden["loot_table"]]["entries"]
	if entries.size() != 1 or str(entries[0]["item_def_id"]) != str(loot_golden["expected_item_def_id"]):
		_fail("loot golden mismatch")
		return

	print("[gdtest] PASS: consumed shared/golden fixtures (damage_formula, retaliation_damage, loot_roll)")
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
