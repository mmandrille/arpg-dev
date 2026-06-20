# Unit tests for local movement visual smoothing (v299).
# Run via: godot --headless --path client --script res://tests/test_movement_visual_smoothing.gd
extends SceneTree

const MovementVisualSmoothingScript := preload("res://scripts/movement_visual_smoothing.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_small_anchor_step_preserves_visual_world_position()
	_test_tick_eases_offset_to_zero()
	_test_large_anchor_step_resets_offset()

	print("[gdtest] PASS: test_movement_visual_smoothing (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_small_anchor_step_preserves_visual_world_position() -> void:
	var anchor := Node3D.new()
	var visual := Node3D.new()
	anchor.add_child(visual)
	var smoothing := MovementVisualSmoothingScript.new()
	smoothing.reset(anchor, visual)
	anchor.position = Vector3(0.28, 0.0, 0.0)
	smoothing.preserve_after_anchor_move(anchor, visual)
	_assert_approx("small step visual offset x", visual.position.x, -0.28, 0.001)
	_assert_true("small step smoothing active", bool(smoothing.get_debug_state(visual).get("active", false)))
	anchor.free()


func _test_tick_eases_offset_to_zero() -> void:
	var visual := Node3D.new()
	visual.position = Vector3(-0.28, 0.0, 0.0)
	var smoothing := MovementVisualSmoothingScript.new()
	smoothing.tick(0.10, visual)
	_assert_true("tick reduces offset", absf(visual.position.x) < 0.28)
	for i in range(20):
		smoothing.tick(0.10, visual)
	_assert_approx("tick settles offset x", visual.position.x, 0.0, 0.001)
	_assert_false("settled smoothing inactive", bool(smoothing.get_debug_state(visual).get("active", true)))
	visual.free()


func _test_large_anchor_step_resets_offset() -> void:
	var anchor := Node3D.new()
	var visual := Node3D.new()
	anchor.add_child(visual)
	var smoothing := MovementVisualSmoothingScript.new()
	smoothing.reset(anchor, visual)
	anchor.position = Vector3(4.0, 0.0, 0.0)
	smoothing.preserve_after_anchor_move(anchor, visual)
	_assert_approx("large step reset offset x", visual.position.x, 0.0, 0.001)
	_assert_false("large step smoothing inactive", bool(smoothing.get_debug_state(visual).get("active", true)))
	anchor.free()


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _assert_approx(label: String, got: float, want: float, tolerance: float) -> void:
	if absf(got - want) <= tolerance:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s (got %s want %s)" % [label, str(got), str(want)])
