extends SceneTree

const FogOfWarOverlayScript := preload("res://scripts/fog_of_war_overlay.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	await _test_progression_sets_light_and_gloom_radius()
	await _test_organic_edge_debug_state()
	await _test_organic_edge_rotates_only_while_target_moves()
	await _test_wall_layout_generates_shadow()
	await _test_tall_obstacle_layout_generates_shadow()
	await _test_water_layout_skips_shadow()
	await _test_hole_layout_skips_shadow()
	await _test_rubble_layout_skips_shadow()
	await _test_explicit_low_wall_skips_shadow()
	await _test_supplied_door_occluder_generates_shadow()
	await _test_diagonal_wall_shadow_starts_near_visible_edge()
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


func _test_organic_edge_debug_state() -> void:
	var overlay = FogOfWarOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	overlay.set_progression({"derived_stats": {"light_radius": 9}})
	await process_frame
	var state := overlay.get_debug_state()
	_assert_true("organic edge enabled", bool(state.get("organic_edge_enabled", false)))
	_assert_true("organic edge pixels positive", float(state.get("organic_edge_px", 0.0)) >= 5.0)
	_assert_true("organic edge stays modest", float(state.get("organic_edge_px", 9999.0)) <= float(state.get("gloom_radius_px", 1.0)) * 0.10)
	_assert_true("darkness feather pixels positive", float(state.get("darkness_feather_px", 0.0)) >= 8.0)
	_assert_true("darkness feather stays modest", float(state.get("darkness_feather_px", 9999.0)) <= float(state.get("gloom_radius_px", 1.0)) * 0.18)
	_assert_eq("organic edge segments", int(state.get("organic_edge_segments", 0)), 18)
	overlay.free()


func _test_organic_edge_rotates_only_while_target_moves() -> void:
	var target := Node3D.new()
	get_root().add_child(target)
	var overlay = FogOfWarOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	overlay.bind(null, target)
	overlay.set_progression({"derived_stats": {"light_radius": 9}})
	await process_frame
	var initial_state := overlay.get_debug_state()
	var initial_rotation := float(initial_state.get("organic_edge_rotation", -1.0))
	_assert_false("edge rotation initially idle", bool(initial_state.get("organic_edge_rotation_active", true)))
	target.position = Vector3(0.25, 0.0, 0.0)
	await process_frame
	var moving_state := overlay.get_debug_state()
	var moving_rotation := float(moving_state.get("organic_edge_rotation", -1.0))
	_assert_true("edge rotation active while target moves", bool(moving_state.get("organic_edge_rotation_active", false)))
	_assert_true("edge rotation advances while target moves", moving_rotation > initial_rotation)
	await process_frame
	var idle_state := overlay.get_debug_state()
	_assert_false("edge rotation idles after target stops", bool(idle_state.get("organic_edge_rotation_active", true)))
	_assert_eq("edge rotation holds when target stops", float(idle_state.get("organic_edge_rotation", -1.0)), moving_rotation)
	overlay.free()
	target.free()


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
	_assert_true("shadow core is not full black", float(state.get("shadow_core_alpha", 1.0)) < 1.0)
	_assert_true("shadow gloom underlay present", float(state.get("shadow_gloom_alpha", 0.0)) > 0.0)
	var shadows: Array = state.get("shadow_polygons", [])
	var first: Dictionary = shadows[0] if shadows.size() > 0 else {}
	_assert_true("shadow has polygon points", (first.get("points", []) as Array).size() >= 4)
	overlay.free()


func _test_tall_obstacle_layout_generates_shadow() -> void:
	for kind in ["rock", "column"]:
		var overlay = FogOfWarOverlayScript.new()
		get_root().add_child(overlay)
		await process_frame
		overlay.set_progression({"derived_stats": {"light_radius": 9}})
		overlay.set_wall_layout([{"kind": kind, "blocks_line_of_sight": true, "position": {"x": 3.0, "y": 0.0}, "size": {"x": 1.0, "y": 3.0}}])
		await process_frame
		var state := overlay.get_debug_state()
		_assert_eq("%s tall wall count" % kind, int(state.get("wall_count", 0)), 1)
		_assert_eq("%s tall occluder count" % kind, int(state.get("occluder_count", 0)), 1)
		_assert_eq("%s tall shadow count" % kind, int(state.get("shadow_count", 0)), 1)
		overlay.free()


func _test_water_layout_skips_shadow() -> void:
	var overlay = FogOfWarOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	overlay.set_progression({"derived_stats": {"light_radius": 9}})
	overlay.set_wall_layout([{"kind": "water", "position": {"x": 3.0, "y": 0.0}, "size": {"x": 3.0, "y": 3.0}}])
	await process_frame
	var state := overlay.get_debug_state()
	_assert_eq("water wall count", int(state.get("wall_count", -1)), 0)
	_assert_eq("water occluder count", int(state.get("occluder_count", -1)), 0)
	_assert_eq("water shadow count", int(state.get("shadow_count", -1)), 0)
	overlay.free()


func _test_hole_layout_skips_shadow() -> void:
	var overlay = FogOfWarOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	overlay.set_progression({"derived_stats": {"light_radius": 9}})
	overlay.set_wall_layout([{"kind": "hole", "position": {"x": 3.0, "y": 0.0}, "size": {"x": 3.0, "y": 3.0}}])
	await process_frame
	var state := overlay.get_debug_state()
	_assert_eq("hole wall count", int(state.get("wall_count", -1)), 0)
	_assert_eq("hole occluder count", int(state.get("occluder_count", -1)), 0)
	_assert_eq("hole shadow count", int(state.get("shadow_count", -1)), 0)
	overlay.free()


func _test_rubble_layout_skips_shadow() -> void:
	var overlay = FogOfWarOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	overlay.set_progression({"derived_stats": {"light_radius": 9}})
	overlay.set_wall_layout([{"kind": "rubble", "position": {"x": 3.0, "y": 0.0}, "size": {"x": 3.0, "y": 3.0}}])
	await process_frame
	var state := overlay.get_debug_state()
	_assert_eq("rubble wall count", int(state.get("wall_count", -1)), 0)
	_assert_eq("rubble occluder count", int(state.get("occluder_count", -1)), 0)
	_assert_eq("rubble shadow count", int(state.get("shadow_count", -1)), 0)
	overlay.free()


func _test_explicit_low_wall_skips_shadow() -> void:
	var overlay = FogOfWarOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	overlay.set_progression({"derived_stats": {"light_radius": 9}})
	overlay.set_wall_layout([{"blocks_line_of_sight": false, "position": {"x": 3.0, "y": 0.0}, "size": {"x": 1.0, "y": 3.0}}])
	await process_frame
	var state := overlay.get_debug_state()
	_assert_eq("explicit low wall count", int(state.get("wall_count", -1)), 0)
	_assert_eq("explicit low wall occluder count", int(state.get("occluder_count", -1)), 0)
	_assert_eq("explicit low wall shadow count", int(state.get("shadow_count", -1)), 0)
	overlay.free()


func _test_supplied_door_occluder_generates_shadow() -> void:
	var overlay = FogOfWarOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	overlay.set_progression({"derived_stats": {"light_radius": 9}})
	overlay.set_occluder_layout([{"position": {"x": 3.0, "y": 0.0}, "size": {"x": 1.0, "y": 0.25}}])
	await process_frame
	var state := overlay.get_debug_state()
	_assert_eq("door occluder leaves wall count", int(state.get("wall_count", -1)), 0)
	_assert_eq("door extra occluder count", int(state.get("extra_occluder_count", 0)), 1)
	_assert_eq("door occluder count", int(state.get("occluder_count", 0)), 1)
	_assert_eq("door shadow count", int(state.get("shadow_count", 0)), 1)
	overlay.free()


func _test_diagonal_wall_shadow_starts_near_visible_edge() -> void:
	var target := Node3D.new()
	target.position = Vector3(2.0, 0.0, 2.0)
	get_root().add_child(target)
	var camera := Camera3D.new()
	camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	camera.size = 20.0
	get_root().add_child(camera)
	camera.current = true
	camera.global_position = target.global_position + Vector3(9.0, 20.0, 15.0)
	camera.look_at(target.global_position, Vector3.UP)
	var overlay = FogOfWarOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	overlay.bind(camera, target)
	overlay.set_progression({"derived_stats": {"light_radius": 9}})
	overlay.set_wall_layout([{"position": {"x": 4.0, "y": 6.0}, "size": {"x": 1.0, "y": 6.0}}])
	await process_frame
	var state := overlay.get_debug_state()
	_assert_eq("diagonal shadow count", int(state.get("shadow_count", 0)), 1)
	_assert_true("diagonal shadow small start offset", float(state.get("shadow_start_offset", 1.0)) <= 0.25)
	_assert_eq("diagonal shadow wall height", float(state.get("shadow_wall_height", 0.0)), 1.0)
	var shadows: Array = state.get("shadow_polygons", [])
	var first: Dictionary = shadows[0] if shadows.size() > 0 else {}
	var points: Array = first.get("points", [])
	_assert_true("diagonal shadow has polygon points", points.size() >= 4)
	if points.size() >= 4:
		var near_a := _point_from_debug(points[0] as Dictionary)
		var top_a := camera.unproject_position(Vector3(4.5, 1.0, 3.0))
		var ground_a := camera.unproject_position(Vector3(4.5, 0.0, 3.0))
		_assert_true("diagonal shadow starts from wall top silhouette", near_a.distance_to(top_a) < near_a.distance_to(ground_a))
	overlay.free()
	camera.free()
	target.free()


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
	_assert_false("zero radius organic edge disabled", bool(state.get("organic_edge_enabled", true)))
	_assert_eq("zero radius organic edge px", float(state.get("organic_edge_px", -1.0)), 0.0)
	_assert_eq("zero radius darkness feather px", float(state.get("darkness_feather_px", -1.0)), 0.0)
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


func _point_from_debug(point: Dictionary) -> Vector2:
	return Vector2(float(point.get("x", 0.0)), float(point.get("y", 0.0)))
