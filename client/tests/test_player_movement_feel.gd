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


func _pass(label: String) -> void:
	_pass_count += 1


func _fail(label: String, got: float, want: float) -> void:
	_fail_count += 1
	push_error("%s: got=%s want=%s" % [label, str(got), str(want)])
