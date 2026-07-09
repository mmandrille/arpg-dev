# Unit tests for entity tick smoothing between authoritative snapshots (v349).
extends SceneTree

const EntityTickSmoothingScript := preload("res://scripts/entity_tick_smoothing.gd")
const EntityTickSmoothingRuntimeScript := preload("res://scripts/entity_tick_smoothing_runtime.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_begin_segment_interpolates()
	_test_large_delta_snaps()
	_test_advance_settles()
	_test_adaptive_segment_duration()
	_test_remote_adaptive_runtime_clamps_to_min()
	_test_remote_adaptive_runtime_clamps_to_max()

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


func _test_adaptive_segment_duration() -> void:
	var smoothing := EntityTickSmoothingScript.new()
	smoothing.configure(0.1, 2.0)
	smoothing.begin_segment(Vector3(1.0, 0.0, 0.0), Vector3.ZERO, 0.12)
	var debug := smoothing.get_debug_state()
	_assert_approx("adaptive duration applied", float(debug.get("segment_duration", 0.0)), 0.12, 0.001)


func _test_remote_adaptive_runtime_clamps_to_min() -> void:
	var runtime := EntityTickSmoothingRuntimeScript.new()
	var rec := {"type": "monster"}
	var node := Node3D.new()
	node.position = Vector3.ZERO
	runtime.apply_entity_authoritative(rec, node, Vector3(0.5, 0.0, 0.0), false, true)
	var smoothing = rec.get("tick_smoothing") as EntityTickSmoothing
	_assert_true("remote adaptive creates smoothing", smoothing != null)
	if smoothing != null:
		var debug := smoothing.get_debug_state()
		_assert_approx("remote adaptive min duration", float(debug.get("segment_duration", 0.0)), 0.11, 0.001)
	node.free()


func _test_remote_adaptive_runtime_clamps_to_max() -> void:
	var runtime := EntityTickSmoothingRuntimeScript.new()
	var rec := {"type": "monster"}
	var node := Node3D.new()
	node.position = Vector3.ZERO
	runtime.apply_entity_authoritative(rec, node, Vector3(1.2, 0.0, 0.0), false, true)
	var smoothing = rec.get("tick_smoothing") as EntityTickSmoothing
	_assert_true("remote adaptive creates smoothing at max distance", smoothing != null)
	if smoothing != null:
		var debug := smoothing.get_debug_state()
		_assert_approx("remote adaptive max duration", float(debug.get("segment_duration", 0.0)), 0.18, 0.001)
	node.free()


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
