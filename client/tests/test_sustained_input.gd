# Unit tests for sustained left-click hold state (v27).
# Run via: godot --headless --path client --script res://tests/test_sustained_input.gd
extends SceneTree

const SustainedClickInputScript := preload("res://scripts/sustained_click_input.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_begin_monster_hold()
	_test_begin_floor_hold()
	_test_begin_oneshot_no_hold()
	_test_should_stop_missing_target()
	_test_should_stop_dead_monster()
	_test_can_repeat_move_epsilon()
	_test_clear_resets()

	print("[gdtest] PASS: test_sustained_input (%d passed, %d failed)" % [_pass_count, _fail_count])
	if _fail_count > 0:
		quit(1)
	else:
		quit(0)


func _test_begin_monster_hold() -> void:
	var hold := SustainedClickInputScript.new()
	_assert_true("monster hold starts", hold.begin_from_pick({"kind": "monster", "target_id": "1002"}))
	_assert_true("monster hold active", hold.active)
	_assert_eq("monster hold mode", hold.mode, "attack")
	_assert_eq("monster sticky target", hold.target_id, "1002")


func _test_begin_floor_hold() -> void:
	var hold := SustainedClickInputScript.new()
	var ground := Vector3(3.0, 0.0, 4.0)
	_assert_true("floor hold starts", hold.begin_from_pick({"kind": "floor", "ground": ground}))
	_assert_true("floor hold active", hold.active)
	_assert_eq("floor hold mode", hold.mode, "move")
	_assert_eq("floor last ground x", hold.last_ground.x, 3.0)
	_assert_eq("floor last ground y", hold.last_ground.y, 4.0)


func _test_begin_oneshot_no_hold() -> void:
	var hold := SustainedClickInputScript.new()
	hold.begin_from_pick({"kind": "monster", "target_id": "1002"})
	_assert_false("loot click does not hold", hold.begin_from_pick({"kind": "oneshot", "target_id": "2001"}))
	_assert_false("oneshot not active", hold.active)


func _test_should_stop_missing_target() -> void:
	var hold := SustainedClickInputScript.new()
	hold.begin_from_pick({"kind": "monster", "target_id": "1002"})
	_assert_true("missing target stops", hold.should_stop(10, {}))


func _test_should_stop_dead_monster() -> void:
	var hold := SustainedClickInputScript.new()
	hold.begin_from_pick({"kind": "monster", "target_id": "1002"})
	var entities := {
		"1002": {"type": "monster", "hp": 0},
	}
	_assert_true("dead monster stops", hold.should_stop(10, entities))


func _test_can_repeat_move_epsilon() -> void:
	var hold := SustainedClickInputScript.new()
	hold.begin_from_pick({"kind": "floor", "ground": Vector3(0.0, 0.0, 0.0)})
	_assert_false("small move below epsilon", hold.can_repeat_move(Vector3(0.1, 0.0, 0.1)))
	_assert_true("large move above epsilon", hold.can_repeat_move(Vector3(0.5, 0.0, 0.0)))


func _test_clear_resets() -> void:
	var hold := SustainedClickInputScript.new()
	hold.begin_from_pick({"kind": "monster", "target_id": "1002"})
	hold.clear()
	_assert_false("clear deactivates", hold.active)
	_assert_eq("clear mode", hold.mode, "")
	_assert_eq("clear target", hold.target_id, "")


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _assert_eq(label: String, got: Variant, want: Variant) -> void:
	if got == want:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s (got %s want %s)" % [label, str(got), str(want)])
