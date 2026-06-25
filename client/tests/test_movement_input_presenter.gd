# Unit tests for movement intent presentation helpers (v335).
# Run via: godot --headless --path client --script res://tests/test_movement_input_presenter.gd
extends SceneTree

const MovementInputPresenterScript := preload("res://scripts/movement_input_presenter.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_move_to_intent_starts_motion()
	_test_zero_move_intent_does_not_start_motion()
	_test_nonzero_move_intent_starts_motion()
	_test_mark_walking_sets_linger()

	print("[gdtest] PASS: test_movement_input_presenter (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_move_to_intent_starts_motion() -> void:
	_assert_true(
		"move_to starts motion",
		MovementInputPresenterScript.intent_starts_motion("move_to_intent", {"position": {"x": 1.0, "y": 2.0}}),
	)


func _test_zero_move_intent_does_not_start_motion() -> void:
	_assert_false(
		"zero move_intent",
		MovementInputPresenterScript.intent_starts_motion("move_intent", {"direction": {"x": 0.0, "y": 0.0}}),
	)


func _test_nonzero_move_intent_starts_motion() -> void:
	_assert_true(
		"nonzero move_intent",
		MovementInputPresenterScript.intent_starts_motion("move_intent", {"direction": {"x": 0.5, "y": 0.0}}),
	)


func _test_mark_walking_sets_linger() -> void:
	var presenter := MovementInputPresenterScript.new()
	presenter.mark_walking()
	_assert_gt("walk linger positive", presenter.walk_linger, 0.0)


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		printerr("[gdtest] FAIL: %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _assert_gt(label: String, value: float, threshold: float) -> void:
	if value > threshold:
		_pass_count += 1
	else:
		_fail_count += 1
		printerr("[gdtest] FAIL: %s (got %s want > %s)" % [label, value, threshold])
