# Unit tests for projectile tick smoothing (v352).
extends SceneTree

const EntityTickSmoothingRuntimeScript := preload("res://scripts/entity_tick_smoothing_runtime.gd")
const MovementPresentationLoaderScript := preload("res://scripts/movement_presentation_loader.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_apply_projectile_authoritative_interpolates()
	_test_apply_projectile_disabled_snaps()
	_test_tick_entities_faces_motion()
	_test_active_projectile_debug_state()
	print("[gdtest] PASS: test_projectile_tick_smoothing (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_apply_projectile_authoritative_interpolates() -> void:
	MovementPresentationLoaderScript.reset_for_tests()
	var runtime := EntityTickSmoothingRuntimeScript.new()
	var node := Node3D.new()
	get_root().add_child(node)
	var rec := {"type": "projectile"}
	runtime.apply_projectile_authoritative(rec, node, Vector3.ZERO, true)
	runtime.apply_projectile_authoritative(rec, node, Vector3(0.5, 0.0, 0.0), false)
	var smoothing = rec["tick_smoothing"]
	_assert_true("projectile segment active", smoothing.is_active())
	var mid: Vector3 = smoothing.advance(0.05)
	_assert_true("projectile midpoint between endpoints", mid.x > 0.0 and mid.x < 0.5)
	node.queue_free()
	MovementPresentationLoaderScript.reset_for_tests()


func _test_apply_projectile_disabled_snaps() -> void:
	MovementPresentationLoaderScript.reset_for_tests()
	var runtime := EntityTickSmoothingRuntimeScript.new()
	runtime.ensure_config()
	runtime._projectiles_enabled = false
	var node := Node3D.new()
	get_root().add_child(node)
	node.position = Vector3.ZERO
	var rec := {"type": "projectile"}
	runtime.apply_projectile_authoritative(rec, node, Vector3(1.0, 0.0, 0.0), false)
	_assert_approx("disabled projectile snaps", node.position.x, 1.0, 0.001)
	node.queue_free()
	MovementPresentationLoaderScript.reset_for_tests()


func _test_tick_entities_faces_motion() -> void:
	MovementPresentationLoaderScript.reset_for_tests()
	var runtime := EntityTickSmoothingRuntimeScript.new()
	var node := Node3D.new()
	get_root().add_child(node)
	node.position = Vector3.ZERO
	var rec := {"type": "projectile", "node": node}
	runtime.apply_projectile_authoritative(rec, node, Vector3.ZERO, true)
	runtime.apply_projectile_authoritative(rec, node, Vector3(0.0, 0.0, 1.0), false)
	runtime.tick_entities({"p1": rec}, 0.05)
	_assert_true("projectile faces travel direction", absf(node.rotation.y) > 0.001)
	node.queue_free()
	MovementPresentationLoaderScript.reset_for_tests()


func _test_active_projectile_debug_state() -> void:
	MovementPresentationLoaderScript.reset_for_tests()
	var runtime := EntityTickSmoothingRuntimeScript.new()
	var node := Node3D.new()
	get_root().add_child(node)
	var rec := {"type": "projectile", "node": node}
	runtime.apply_projectile_authoritative(rec, node, Vector3.ZERO, true)
	runtime.apply_projectile_authoritative(rec, node, Vector3(0.4, 0.0, 0.0), false)
	var debug := runtime.get_active_projectile_debug_state({"p1": rec})
	_assert_true("debug reports active projectile smoothing", bool(debug.get("active", false)))
	node.queue_free()
	MovementPresentationLoaderScript.reset_for_tests()


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s" % label)


func _assert_approx(label: String, got: float, want: float, tolerance: float) -> void:
	if absf(got - want) <= tolerance:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s (got %s want %s)" % [label, str(got), str(want)])
