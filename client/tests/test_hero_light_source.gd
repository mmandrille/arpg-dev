extends SceneTree

const HeroLightSourceScript := preload("res://scripts/hero_light_source.gd")

var _failures: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	await _test_local_light_position_uses_mesh_bounds()
	await _test_world_height_scales_with_visual_scale()
	await _test_world_height_falls_back_to_anchor()
	if _failures > 0:
		quit(1)
	print("[gdtest] PASS: test_hero_light_source")
	quit(0)


func _test_local_light_position_uses_mesh_bounds() -> void:
	var visual := Node3D.new()
	get_root().add_child(visual)
	var mesh := MeshInstance3D.new()
	var box := BoxMesh.new()
	box.size = Vector3(1.0, 2.0, 1.0)
	mesh.mesh = box
	mesh.position = Vector3(0.0, 1.0, 0.0)
	visual.add_child(mesh)
	await process_frame
	var cfg := {"height_fraction": 0.5, "min_height": 0.25}
	var pos := HeroLightSourceScript.local_light_position(visual, cfg)
	_assert_true("local light y at mesh mid-height", absf(pos.y - 1.0) <= 0.001)
	visual.scale = Vector3.ONE * 1.5
	pos = HeroLightSourceScript.local_light_position(visual, cfg)
	_assert_true("local light y stays in unscaled local space", absf(pos.y - 1.0) <= 0.001)
	visual.free()


func _test_world_height_scales_with_visual_scale() -> void:
	var anchor := Node3D.new()
	var visual := Node3D.new()
	get_root().add_child(anchor)
	anchor.add_child(visual)
	var mesh := MeshInstance3D.new()
	var box := BoxMesh.new()
	box.size = Vector3(1.0, 2.0, 1.0)
	mesh.mesh = box
	mesh.position = Vector3(0.0, 1.0, 0.0)
	visual.add_child(mesh)
	visual.scale = Vector3.ONE * 2.0
	await process_frame
	var cfg := {"height_fraction": 0.5, "min_height": 0.25}
	var height := HeroLightSourceScript.estimate_world_height(visual, anchor, cfg)
	_assert_true("world height scales with visual", height >= 2.0)
	anchor.free()


func _test_world_height_falls_back_to_anchor() -> void:
	var anchor := Node3D.new()
	anchor.position = Vector3(0.0, 4.0, 0.0)
	get_root().add_child(anchor)
	await process_frame
	var cfg := {"height_fraction": 0.5, "min_height": 1.25}
	var height := HeroLightSourceScript.estimate_world_height(null, anchor, cfg)
	_assert_true("fallback uses anchor plus min height", absf(height - 5.25) <= 0.001)
	anchor.free()


func _assert_true(label: String, value: bool) -> void:
	if not value:
		_failures += 1
		printerr("[gdtest] FAIL %s" % label)
