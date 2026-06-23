# Panel UI intents must stay dispatchable while gameplay panels are open.
# Run via: godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_panel_intent_input.gd
extends SceneTree

const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")
const MainScript := preload("res://scripts/main.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_open_panel_blocks_world_input_not_panel_intents()
	print("[gdtest] PASS: test_panel_intent_input (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_open_panel_blocks_world_input_not_panel_intents() -> void:
	var main := MainScript.new()
	var panel := InventoryPanelScript.new()
	panel.visible = true
	main.inventory_panel = panel
	_assert_true("open panel blocks world input", main._input_locked())
	_assert_false("open panel does not block panel intents", main._automation_input_locked())
	main.free()
	panel.free()


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)
