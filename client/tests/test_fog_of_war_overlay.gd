extends SceneTree

const FogOfWarOverlayScript := preload("res://scripts/fog_of_war_overlay.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	await _test_progression_sets_light_and_gloom_radius()
	await _test_wall_layout_generates_shadow()
	await _test_out_of_range_wall_skips_shadow()
	await _test_multiple_walls_generate_multiple_shadows()
	await _test_zero_radius_disables_overlay()
	print("[gdtest] PASS: test_fog_of_war_overlay (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_progression_sets_light_and_gloom_radius() -> void:
	var overlay = FogOfWarOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	overlay.set_progression({"derived_stats": {"light_radius": 12}})
	await process_frame
	var state := overlay.get_debug_state()
	_assert_true("overlay enabled", bool(state.get("enabled", false)))
	_assert_eq("light radius", float(state.get("light_radius", 0.0)), 12.0)
	_assert_eq("gloom radius", float(state.get("gloom_radius", 0.0)), 15.0)
	_assert_true("screen light radius positive", float(state.get("light_radius_px", 0.0)) > 0.0)
	_assert_eq("darkness alpha", float(state.get("darkness_alpha", 0.0)), 1.0)
	_assert_eq("no wall shadows", int(state.get("shadow_count", -1)), 0)
	overlay.free()


func _test_wall_layout_generates_shadow() -> void:
	var overlay = FogOfWarOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	overlay.set_progression({"derived_stats": {"light_radius": 9}})
	overlay.set_wall_layout([{"position": {"x": 3.0, "y": 0.0}, "size": {"x": 1.0, "y": 3.0}}])
	await process_frame
	var state := overlay.get_debug_state()
	_assert_eq("wall count", int(state.get("wall_count", 0)), 1)
	_assert_eq("occluder count", int(state.get("occluder_count", 0)), 1)
	_assert_eq("shadow count", int(state.get("shadow_count", 0)), 1)
	var shadows: Array = state.get("shadow_polygons", [])
	var first: Dictionary = shadows[0] if shadows.size() > 0 else {}
	_assert_true("shadow has polygon points", (first.get("points", []) as Array).size() >= 4)
	overlay.free()


func _test_out_of_range_wall_skips_shadow() -> void:
	var overlay = FogOfWarOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	overlay.set_progression({"derived_stats": {"light_radius": 9}})
	overlay.set_wall_layout([{"position": {"x": 99.0, "y": 0.0}, "size": {"x": 1.0, "y": 3.0}}])
	await process_frame
	var state := overlay.get_debug_state()
	_assert_eq("far wall retained", int(state.get("wall_count", 0)), 1)
	_assert_eq("far wall no occluder", int(state.get("occluder_count", -1)), 0)
	_assert_eq("far wall no shadow", int(state.get("shadow_count", -1)), 0)
	overlay.free()


func _test_multiple_walls_generate_multiple_shadows() -> void:
	var overlay = FogOfWarOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	overlay.set_progression({"derived_stats": {"light_radius": 9}})
	overlay.set_wall_layout([
		{"position": {"x": 3.0, "y": 0.0}, "size": {"x": 1.0, "y": 2.0}},
		{"position": {"x": 0.0, "y": 3.0}, "size": {"x": 2.0, "y": 1.0}},
	])
	await process_frame
	var state := overlay.get_debug_state()
	_assert_eq("multiple wall count", int(state.get("wall_count", 0)), 2)
	_assert_eq("multiple occluders", int(state.get("occluder_count", 0)), 2)
	_assert_eq("multiple shadows", int(state.get("shadow_count", 0)), 2)
	overlay.free()


func _test_zero_radius_disables_overlay() -> void:
	var overlay = FogOfWarOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	overlay.set_progression({"derived_stats": {"light_radius": 0}})
	var state := overlay.get_debug_state()
	_assert_false("zero radius disabled", bool(state.get("enabled", true)))
	_assert_eq("zero radius shadows", int(state.get("shadow_count", -1)), 0)
	overlay.free()


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)
