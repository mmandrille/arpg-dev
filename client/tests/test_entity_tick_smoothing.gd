# Unit tests for entity tick smoothing between authoritative snapshots (v349).
extends SceneTree

const EntityTickSmoothingScript := preload("res://scripts/entity_tick_smoothing.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_begin_segment_interpolates()
	_test_large_delta_snaps()
	_test_advance_settles()

	print("[gdtest] PASS: test_entity_tick_smoothing (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_begin_segment_interpolates() -> void:
	var smoothing := EntityTickSmoothingScript.new()
	smoothing.configure(0.1, 2.0)
	smoothing.begin_segment(Vector3(1.0, 0.0, 0.0), Vector3.ZERO)
	_assert_true("segment active", smoothing.is_active())
	var mid := smoothing.advance(0.05)
	_assert_true("midpoint between endpoints", mid.x > 0.0 and mid.x < 1.0)


func _test_large_delta_snaps() -> void:
	var smoothing := EntityTickSmoothingScript.new()
	smoothing.configure(0.1, 2.0)
	smoothing.begin_segment(Vector3(5.0, 0.0, 0.0), Vector3.ZERO)
	_assert_false("large delta snaps inactive", smoothing.is_active())
	var pos := smoothing.advance(0.05)
	_assert_approx("large delta snaps position", pos.x, 5.0, 0.001)


func _test_advance_settles() -> void:
	var smoothing := EntityTickSmoothingScript.new()
	smoothing.configure(0.1, 2.0)
	smoothing.begin_segment(Vector3(0.5, 0.0, 0.0), Vector3.ZERO)
	for i in range(20):
		smoothing.advance(0.01)
	_assert_false("settled inactive", smoothing.is_active())
	var settled := smoothing.advance(0.0)
	_assert_approx("settled at target", settled.x, 0.5, 0.001)


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
