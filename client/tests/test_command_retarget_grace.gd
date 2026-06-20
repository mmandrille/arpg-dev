# Unit tests for local command retarget grace (v300).
# Run via: godot --headless --path client --script res://tests/test_command_retarget_grace.gd
extends SceneTree

const CommandRetargetGraceScript := preload("res://scripts/command_retarget_grace.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_latest_floor_retarget_replaces_prior()
	_test_pop_ready_records_dispatched_target()
	_test_grace_expires_before_long_cooldown()

	print("[gdtest] PASS: test_command_retarget_grace (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_latest_floor_retarget_replaces_prior() -> void:
	var grace := CommandRetargetGraceScript.new()
	grace.queue_floor(Vector3(1.0, 0.0, 2.0))
	grace.queue_floor(Vector3(3.0, 0.0, 4.0))
	var state := grace.get_debug_state()
	_assert_true("retarget active", bool(state.get("active", false)))
	_assert_eq("queued count", int(state.get("queued_count", 0)), 2)
	_assert_eq("replaced count", int(state.get("replaced_count", 0)), 1)
	_assert_approx("latest ground x", float(state.get("ground_x", 0.0)), 3.0, 0.001)
	_assert_approx("latest ground z", float(state.get("ground_z", 0.0)), 4.0, 0.001)


func _test_pop_ready_records_dispatched_target() -> void:
	var grace := CommandRetargetGraceScript.new()
	grace.queue_floor(Vector3(-2.0, 0.0, 5.0))
	_assert_true("cooldown blocks pop", grace.pop_ready(0.01).is_empty())
	var command := grace.pop_ready(0.0)
	var command_ground: Vector3 = command.get("ground", Vector3.ZERO)
	var state := grace.get_debug_state()
	_assert_eq("pop kind", str(command.get("kind", "")), "floor")
	_assert_approx("pop ground x", command_ground.x, -2.0, 0.001)
	_assert_false("inactive after pop", bool(state.get("active", true)))
	_assert_eq("dispatched count", int(state.get("dispatched_count", 0)), 1)
	_assert_approx("last dispatched z", float(state.get("last_dispatched_ground_z", 0.0)), 5.0, 0.001)


func _test_grace_expires_before_long_cooldown() -> void:
	var grace := CommandRetargetGraceScript.new()
	grace.queue_floor(Vector3(1.0, 0.0, 1.0), 0.05)
	grace.tick(0.06)
	var state := grace.get_debug_state()
	_assert_false("expired inactive", bool(state.get("active", true)))
	_assert_eq("expired count", int(state.get("expired_count", 0)), 1)
	_assert_true("expired cannot pop", grace.pop_ready(0.0).is_empty())


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


func _assert_approx(label: String, got: float, want: float, tolerance: float) -> void:
	if absf(got - want) <= tolerance:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s (got %s want %s)" % [label, str(got), str(want)])
