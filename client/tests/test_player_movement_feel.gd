# Run via: godot --headless --path client --script res://tests/test_player_movement_feel.gd
extends SceneTree

const PlayerMovementFeelScript := preload("res://scripts/player_movement_feel.gd")
const MainConfigLoaderScript := preload("res://scripts/main_config_loader.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	MainConfigLoaderScript.ensure_loaded()
	_test_starts_at_min_speed()
	_test_reaches_full_speed_after_accel_window()
	_test_direction_change_resets_ramp()
	_test_small_correction_within_grace_does_not_reset()
	_test_sharp_turn_always_resets()
	if _fail_count == 0:
		print("[gdtest] PASS: test_player_movement_feel (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(0)
	else:
		print("[gdtest] FAIL: test_player_movement_feel (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)


func _test_starts_at_min_speed() -> void:
	var feel := PlayerMovementFeelScript.new()
	var speed := feel.effective_speed(Vector2.RIGHT, 0.1)
	var expected := MainConfigLoaderScript.base_movement_speed() * MainConfigLoaderScript.movement_min_speed_factor()
	if absf(speed - expected) > 0.0001:
		_fail("initial speed", speed, expected)
		return
	_pass("initial speed uses min factor")


func _test_reaches_full_speed_after_accel_window() -> void:
	var feel := PlayerMovementFeelScript.new()
	var dir := Vector2.RIGHT
	var accel := MainConfigLoaderScript.movement_acceleration_seconds()
	var speed := 0.0
	var step := accel / 10.0
	for _i in range(12):
		speed = feel.effective_speed(dir, step)
	var expected := MainConfigLoaderScript.base_movement_speed()
	if absf(speed - expected) > 0.0001:
		_fail("full speed after ramp", speed, expected)
		return
	_pass("full speed after ramp")


func _test_direction_change_resets_ramp() -> void:
	var feel := PlayerMovementFeelScript.new()
	var accel := MainConfigLoaderScript.movement_acceleration_seconds()
	for _i in range(10):
		feel.effective_speed(Vector2.RIGHT, accel / 10.0)
	var after_turn := feel.effective_speed(Vector2.UP, 0.1)
	var expected := MainConfigLoaderScript.base_movement_speed() * MainConfigLoaderScript.movement_min_speed_factor()
	if absf(after_turn - expected) > 0.0001:
		_fail("direction change reset", after_turn, expected)
		return
	_pass("direction change resets ramp")


func _test_small_correction_within_grace_does_not_reset() -> void:
	# After holding RIGHT past the grace window and above min speed, a small correction
	# (dot >= 0.5) should NOT reset the ramp.
	var feel := PlayerMovementFeelScript.new()
	var grace := MainConfigLoaderScript.movement_direction_grace_seconds()
	var accel := MainConfigLoaderScript.movement_acceleration_seconds()
	var min_factor := MainConfigLoaderScript.movement_min_speed_factor()
	# Ramp past min: hold for enough time so _hold_seconds / accel > min_factor.
	# Steps of 0.1s each; hold for (min_factor + 0.1) * accel to be clearly above min.
	var hold_target := (min_factor + 0.1) * accel
	var step := 0.1
	while feel.get_debug_state()["hold_seconds"] < hold_target:
		feel.effective_speed(Vector2.RIGHT, step)
	# Confirm we are past the grace window as well.
	assert(feel.get_debug_state()["hold_seconds"] >= grace)
	var speed_before := feel.effective_speed(Vector2.RIGHT, 0.0)
	# Small correction: slightly up-right (dot with RIGHT ≈ 0.866, well above 0.5)
	var slight_up_right := Vector2(0.866, 0.5).normalized()
	var speed_after := feel.effective_speed(slight_up_right, 0.01)
	# Speed should NOT have reset to min — should be above min
	var min_speed := MainConfigLoaderScript.base_movement_speed() * min_factor
	if speed_after <= min_speed + 0.001:
		_fail("small correction should not reset ramp", speed_after, min_speed)
		return
	_pass("small correction within grace does not reset")


func _test_sharp_turn_always_resets() -> void:
	# Even after holding long past grace, a sharp turn (dot < 0.5) resets.
	var feel := PlayerMovementFeelScript.new()
	var accel := MainConfigLoaderScript.movement_acceleration_seconds()
	# Hold right until fully ramped
	for _i in range(15):
		feel.effective_speed(Vector2.RIGHT, accel / 10.0)
	# Sharp turn: UP (90 degrees from RIGHT, dot = 0)
	var after_sharp := feel.effective_speed(Vector2.UP, 0.01)
	var expected_min := MainConfigLoaderScript.base_movement_speed() * MainConfigLoaderScript.movement_min_speed_factor()
	if absf(after_sharp - expected_min) > 0.001:
		_fail("sharp turn should reset ramp", after_sharp, expected_min)
		return
	_pass("sharp turn always resets ramp")


func _pass(label: String) -> void:
	_pass_count += 1


func _fail(label: String, got: float, want: float) -> void:
	_fail_count += 1
	push_error("%s: got=%s want=%s" % [label, str(got), str(want)])
