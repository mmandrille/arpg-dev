# Unit tests for force-stand directional attack helpers (v37).
# Run via: godot --headless --path client --script res://tests/test_directional_attack_input.gd
extends SceneTree

const DirectionalAttackInputScript := preload("res://scripts/directional_attack_input.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_direction_normalizes_mouse_aim()
	_test_direction_falls_back_to_facing()
	_test_direction_has_last_resort_default()
	_test_payload_uses_normalized_direction()
	_test_repeat_gate()

	print("[gdtest] PASS: test_directional_attack_input (%d passed, %d failed)" % [_pass_count, _fail_count])
	if _fail_count > 0:
		quit(1)
	else:
		quit(0)


func _test_direction_normalizes_mouse_aim() -> void:
	var got := DirectionalAttackInputScript.direction_or_fallback(Vector2(3.0, 4.0), Vector2.LEFT)
	_assert_float("aim x", got.x, 0.6)
	_assert_float("aim y", got.y, 0.8)


func _test_direction_falls_back_to_facing() -> void:
	var got := DirectionalAttackInputScript.direction_or_fallback(Vector2.ZERO, Vector2(0.0, -2.0))
	_assert_float("fallback x", got.x, 0.0)
	_assert_float("fallback y", got.y, -1.0)


func _test_direction_has_last_resort_default() -> void:
	var got := DirectionalAttackInputScript.direction_or_fallback(Vector2.ZERO, Vector2.ZERO)
	_assert_float("default x", got.x, 1.0)
	_assert_float("default y", got.y, 0.0)


func _test_payload_uses_normalized_direction() -> void:
	var payload: Dictionary = DirectionalAttackInputScript.payload(Vector2(0.0, 5.0))
	var direction: Dictionary = payload.get("direction", {})
	_assert_float("payload x", float(direction.get("x", 99.0)), 0.0)
	_assert_float("payload y", float(direction.get("y", 99.0)), 1.0)


func _test_repeat_gate() -> void:
	_assert_true("repeat allowed", DirectionalAttackInputScript.can_repeat(true, true, true, 10))
	_assert_false("repeat stops without shift", DirectionalAttackInputScript.can_repeat(false, true, true, 10))
	_assert_false("repeat stops without mouse", DirectionalAttackInputScript.can_repeat(true, false, true, 10))
	_assert_false("repeat stops blocked gameplay", DirectionalAttackInputScript.can_repeat(true, true, false, 10))
	_assert_false("repeat stops dead player", DirectionalAttackInputScript.can_repeat(true, true, true, 0))


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _assert_float(label: String, got: float, want: float) -> void:
	if absf(got - want) <= 0.0001:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s (got %s want %s)" % [label, str(got), str(want)])
