extends SceneTree

const SetCollectionPanelScript := preload("res://scripts/set_collection_panel.gd")


func _init() -> void:
	var root := Control.new()
	get_root().add_child(root)
	var panel := SetCollectionPanelScript.new()
	root.add_child(panel)
	_test_empty_catalog_state(panel)
	_test_owned_and_equipped_progress(panel)
	_test_refresh_resets_removed_piece(panel)
	panel.queue_free()
	root.queue_free()
	print("[gdtest] PASS: test_set_collection_panel")
	quit()


func _test_empty_catalog_state(panel) -> void:
	panel.set_items([], {})
	var state: Dictionary = panel.get_debug_state()
	var sets: Array = state.get("sets", [])
	_assert_true("catalog exposes enabled sets", sets.size() >= 2)
	var verdant := _set_by_name(sets, "Verdant Vanguard")
	_assert_eq("empty owned", int(verdant.get("owned_count", -1)), 0)
	_assert_eq("empty equipped", int(verdant.get("equipped_count", -1)), 0)


func _test_owned_and_equipped_progress(panel) -> void:
	var items := [
		_set_item("1001", "Verdant Vanguard Blade"),
		_set_item("1002", "Verdant Vanguard Helm"),
		_set_item("2001", "Stormrunner Covenant Mask"),
	]
	panel.set_items(items, {"main_hand": "1001", "head": "1002"})
	var sets: Array = panel.get_debug_state().get("sets", [])
	var verdant := _set_by_name(sets, "Verdant Vanguard")
	_assert_eq("verdant owned", int(verdant.get("owned_count", -1)), 2)
	_assert_eq("verdant equipped", int(verdant.get("equipped_count", -1)), 2)
	_assert_eq("verdant blade state", _piece_state(verdant, "Verdant Vanguard Blade"), "equipped")
	_assert_eq("verdant mail missing", _piece_state(verdant, "Verdant Vanguard Mail"), "missing")
	_assert_true("two-piece bonus active", _bonus_active(verdant, 2))
	_assert_false("three-piece bonus inactive", _bonus_active(verdant, 3))
	var stormrunner := _set_by_name(sets, "Stormrunner Covenant")
	_assert_eq("stormrunner owned", int(stormrunner.get("owned_count", -1)), 1)
	_assert_eq("stormrunner equipped", int(stormrunner.get("equipped_count", -1)), 0)
	_assert_eq("stormrunner mask state", _piece_state(stormrunner, "Stormrunner Covenant Mask"), "owned")


func _test_refresh_resets_removed_piece(panel) -> void:
	panel.set_items([_set_item("1001", "Verdant Vanguard Blade")], {})
	var verdant := _set_by_name(panel.get_debug_state().get("sets", []), "Verdant Vanguard")
	_assert_eq("pre-refresh owned", int(verdant.get("owned_count", -1)), 1)
	panel.set_items([], {})
	verdant = _set_by_name(panel.get_debug_state().get("sets", []), "Verdant Vanguard")
	_assert_eq("post-refresh owned", int(verdant.get("owned_count", -1)), 0)
	_assert_eq("post-refresh blade missing", _piece_state(verdant, "Verdant Vanguard Blade"), "missing")


func _set_item(instance_id: String, display_name: String) -> Dictionary:
	return {
		"item_instance_id": instance_id,
		"display_name": display_name,
		"rarity": "set",
		"summary_lines": ["Set: %s (0/5 equipped)" % display_name],
	}


func _set_by_name(sets: Array, name: String) -> Dictionary:
	for row in sets:
		var rec := row as Dictionary
		if str(rec.get("set_name", "")) == name:
			return rec
	_fail("missing set %s in %s" % [name, str(sets)])
	return {}


func _piece_state(set_state: Dictionary, piece_name: String) -> String:
	for piece in set_state.get("pieces", []):
		var rec := piece as Dictionary
		if str(rec.get("name", "")) == piece_name:
			return str(rec.get("state", ""))
	_fail("missing piece %s in %s" % [piece_name, str(set_state)])
	return ""


func _bonus_active(set_state: Dictionary, required: int) -> bool:
	for bonus in set_state.get("bonuses", []):
		var rec := bonus as Dictionary
		if int(rec.get("required_pieces", 0)) == required:
			return bool(rec.get("active", false))
	_fail("missing bonus %d in %s" % [required, str(set_state)])
	return false


func _assert_eq(label: String, got, want) -> void:
	if got != want:
		_fail("%s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_true(label: String, value: bool) -> void:
	if not value:
		_fail("%s: expected true" % label)


func _assert_false(label: String, value: bool) -> void:
	if value:
		_fail("%s: expected false" % label)


func _fail(message: String) -> void:
	push_error(message)
	quit(1)
