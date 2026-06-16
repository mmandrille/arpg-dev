# Unit test for focused mercenary roster panel state.
# Run via: godot --headless --path client --script res://tests/test_mercenary_panel.gd
extends SceneTree

const MercenaryPanelScript := preload("res://scripts/mercenary_panel.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var panel := MercenaryPanelScript.new()
	root.add_child(panel)
	await process_frame

	panel.show_board("board-1", "mercenary", "fixed:mercenary_guard", "mercenary_guard", 75, true, 125)
	var state := panel.get_debug_state()
	_assert_true("panel visible after board event", bool(state.get("visible", false)))
	_assert_eq("service id", str(state.get("service_id", "")), "mercenary")
	_assert_eq("offer id", str(state.get("offer_id", "")), "fixed:mercenary_guard")
	_assert_eq("price", int(state.get("price", 0)), 75)
	_assert_eq("gold", int(state.get("gold", 0)), 125)
	_assert_true("affordable", bool(state.get("affordable", false)))
	_assert_eq("empty roster", int(state.get("hired_count", -1)), 0)

	panel.apply_hired_event({
		"target_entity_id": "2001",
		"monster_def_id": "mercenary_guard",
		"price": 75,
		"total_gold": 50,
	})
	panel.set_companions([
		{"id": "2001", "monster_def_id": "mercenary_guard", "hp": 24, "max_hp": 30},
		{"id": "wolf-1", "monster_def_id": "ranger_wolf", "hp": 10, "max_hp": 10},
	])
	state = panel.get_debug_state()
	_assert_eq("hired entity id", str(state.get("hired_entity_id", "")), "2001")
	_assert_eq("gold after hire", int(state.get("gold", 0)), 50)
	_assert_eq("mercenary roster filters non mercenary companions", int(state.get("hired_count", -1)), 1)
	var rows: Array = state.get("hired_rows", [])
	_assert_eq("roster monster id", str((rows[0] as Dictionary).get("monster_def_id", "")), "mercenary_guard")
	_assert_true("status names hire", str(state.get("status", "")).contains("Mercenary Guard"))

	panel.set_gold(30)
	state = panel.get_debug_state()
	_assert_false("affordability follows gold", bool(state.get("affordable", true)))

	panel.queue_free()
	print("[gdtest] PASS: test_mercenary_panel (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)
