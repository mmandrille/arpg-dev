# Unit test for the elite objective minimap pin.
extends SceneTree

const EliteObjectiveMinimapScript := preload("res://scripts/elite_objective_minimap.gd")
const EliteObjectiveMinimapStateScript := preload("res://scripts/elite_objective_minimap_state.gd")

var _failures: int = 0


func _initialize() -> void:
	_test_state_helper()
	_test_widget_debug_state()
	if _failures > 0:
		quit(1)
	print("[gdtest] PASS: test_elite_objective_minimap")
	quit(0)


func _test_state_helper() -> void:
	var chest := Node3D.new()
	chest.position = Vector3(4.0, 0.0, -2.0)
	var state := EliteObjectiveMinimapStateScript.from_entities({
		"chest_1": {"elite_objective": true, "state": "closed", "node": chest}
	}, Vector3.ZERO)
	_assert_true("active visible", bool(state.get("visible", false)))
	_assert_true("active has pin", bool(state.get("has_pin", false)))
	_assert_eq("active status", str(state.get("status", "")), "active")
	_assert_range("pin x", float(state.get("pin_x", 0.0)), 0.5, 0.9)
	_assert_range("pin y", float(state.get("pin_y", 1.0)), 0.1, 0.5)
	chest.position = Vector3(100.0, 0.0, 100.0)
	var clamped := EliteObjectiveMinimapStateScript.from_entities({
		"chest_1": {"elite_objective": true, "state": "closed", "node": chest}
	}, Vector3.ZERO)
	_assert_eq("clamped x", float(clamped.get("pin_x", 0.0)), 0.9)
	_assert_eq("clamped y", float(clamped.get("pin_y", 0.0)), 0.9)
	var complete := EliteObjectiveMinimapStateScript.from_entities({
		"chest_1": {"elite_objective": true, "state": "open", "node": chest}
	}, Vector3.ZERO)
	_assert_true("complete hidden", not bool(complete.get("visible", true)))
	_assert_eq("complete status", str(complete.get("status", "")), "complete")
	chest.free()


func _test_widget_debug_state() -> void:
	var minimap: EliteObjectiveMinimap = EliteObjectiveMinimapScript.new()
	get_root().add_child(minimap)
	minimap._build()
	_assert_true("initial hidden", not bool(minimap.get_debug_state().get("visible", true)))
	minimap.set_state({"visible": true, "has_pin": true, "status": "active", "pin_x": 1.4, "pin_y": -0.2})
	var active := minimap.get_debug_state()
	_assert_true("widget visible", bool(active.get("visible", false)))
	_assert_true("widget has pin", bool(active.get("has_pin", false)))
	_assert_eq("widget status", str(active.get("status", "")), "active")
	_assert_eq("widget clamped x", float(active.get("pin_x", 0.0)), 1.0)
	_assert_eq("widget clamped y", float(active.get("pin_y", 0.0)), 0.0)
	minimap.set_state({"visible": false, "has_pin": false, "status": "hidden"})
	_assert_true("widget hidden", not bool(minimap.get_debug_state().get("visible", true)))
	minimap.free()


func _assert_true(label: String, value: bool) -> void:
	if not value:
		_failures += 1
		printerr("[gdtest] FAIL %s" % label)


func _assert_eq(label: String, got, want) -> void:
	if got != want:
		_failures += 1
		printerr("[gdtest] FAIL %s got=%s want=%s" % [label, str(got), str(want)])


func _assert_range(label: String, got: float, min_value: float, max_value: float) -> void:
	if got < min_value or got > max_value:
		_failures += 1
		printerr("[gdtest] FAIL %s got=%s range=%s..%s" % [label, str(got), str(min_value), str(max_value)])
