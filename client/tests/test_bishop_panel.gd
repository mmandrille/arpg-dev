# Headless tests for bishop badge-cost panel state.
# Run via: godot --headless --path client --script res://tests/test_bishop_panel.gd
extends SceneTree

const BishopPanelScript := preload("res://scripts/bishop_panel.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var panel := BishopPanelScript.new()
	get_root().add_child(panel)
	await process_frame

	panel.show_bishop("9", "bishop", 0, false, 25, {})
	var state := panel.get_debug_state()
	_assert_false("respec disabled without badge", bool(state.get("respec_enabled", true)))
	_assert_false("revive disabled without badge", bool(state.get("revive_all_enabled", true)))
	_assert_eq("respec resource id", str(state.get("resource_item_def_id", "")), "respec_badge")
	_assert_eq("revive resource id", str(state.get("revive_resource_item_def_id", "")), "resurrection_badge")
	_assert_true("respec text names badge", str(state.get("respec_text", "")).contains("Respec x1"))
	_assert_true("revive text names badge", str(state.get("revive_all_text", "")).contains("Resurrect x1"))

	panel.set_resource_wallet({"respec_badge": 1})
	state = panel.get_debug_state()
	_assert_true("respec enabled with badge", bool(state.get("respec_enabled", false)))
	_assert_false("revive still disabled without badge", bool(state.get("revive_all_enabled", true)))
	_assert_eq("respec wallet count", int(state.get("resource_wallet_count", 0)), 1)

	panel.set_resource_wallet({"respec_badge": 1, "resurrection_badge": 1})
	state = panel.get_debug_state()
	_assert_true("respec enabled with both badges", bool(state.get("respec_enabled", false)))
	_assert_true("revive enabled with resurrection badge", bool(state.get("revive_all_enabled", false)))
	_assert_eq("revive wallet count", int(state.get("revive_resource_wallet_count", 0)), 1)

	if _fail_count == 0:
		print("[gdtest] PASS: test_bishop_panel (%d passed, %d failed)" % [_pass_count, _fail_count])
	else:
		push_error("[gdtest] FAIL: test_bishop_panel (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(_fail_count)


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("%s: expected true" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("%s: got %s want %s" % [label, str(got), str(want)])
