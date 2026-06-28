# Unit tests for loot tick smoothing runtime (v360).
extends SceneTree

const EntityTickSmoothingRuntimeScript := preload("res://scripts/entity_tick_smoothing_runtime.gd")
const MovementPresentationLoaderScript := preload("res://scripts/movement_presentation_loader.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	MovementPresentationLoaderScript.reset_for_tests()
	var runtime := EntityTickSmoothingRuntimeScript.new()
	var node := Node3D.new()
	var rec := {"type": "loot"}
	runtime.apply_loot_authoritative(rec, node, Vector3(1.0, 0.0, 0.0), true)
	runtime.apply_loot_authoritative(rec, node, Vector3(1.4, 0.0, 0.0), false)
	var smoothing = rec.get("tick_smoothing")
	_assert_true("loot segment active", smoothing.is_active())
	node.position = smoothing.advance(0.05)
	_assert_true("loot midpoint moved", node.position.x > 1.0 and node.position.x < 1.4)

	print("[gdtest] PASS: test_loot_tick_smoothing (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s" % label)
