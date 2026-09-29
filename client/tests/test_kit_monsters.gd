extends SceneTree
## ADR-0018 P4a (v474): texture-preserving model tint + KayKit monster visuals.

const ModelTintScript := preload("res://scripts/model_tint.gd")
const LoaderScript := preload("res://scripts/kit_monster_presentation_loader.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_tint_preserves_texture()
	_test_tint_untextured_mesh()
	await _test_every_catalog_scene_aliases_profile_clips()
	await _test_alias_does_not_mutate_kit_clip_loop()
	_finish()


func _test_tint_preserves_texture() -> void:
	var mesh := MeshInstance3D.new()
	var box := BoxMesh.new()
	var source := StandardMaterial3D.new()
	source.albedo_texture = ImageTexture.create_from_image(Image.create(2, 2, false, Image.FORMAT_RGB8))
	box.material = source
	mesh.mesh = box
	var tint := Color(0.4, 0.6, 0.8)
	var mat := ModelTintScript.tinted_material(mesh, tint)
	var ok := mat.albedo_texture == source.albedo_texture and mat.albedo_color.is_equal_approx(tint) and mat != source
	mesh.free()
	if not ok:
		_fail("tint must keep the source albedo texture, set the tint, and not mutate the source")
		return
	_pass("tint preserves the mesh texture")


func _test_tint_untextured_mesh() -> void:
	var mesh := MeshInstance3D.new()
	mesh.mesh = BoxMesh.new()
	var mat := ModelTintScript.tinted_material(mesh, Color.RED)
	mesh.free()
	if mat == null or not mat.albedo_color.is_equal_approx(Color.RED):
		_fail("untextured meshes still get a flat tint")
		return
	_pass("untextured meshes keep the flat tint")


func _scene_for(visual_key: String) -> Node3D:
	var packed := load("res://scenes/%s.tscn" % visual_key) as PackedScene
	return packed.instantiate() as Node3D if packed != null else null


func _test_every_catalog_scene_aliases_profile_clips() -> void:
	var keys := LoaderScript.visual_keys()
	if keys.is_empty():
		_fail("kit monster catalog must list scenes")
		return
	for visual_key in keys:
		var node := _scene_for(str(visual_key))
		if node == null:
			_fail("missing scene for %s" % visual_key)
			return
		get_root().add_child(node)
		await process_frame
		var cfg := LoaderScript.monster(str(visual_key))
		var profile := LoaderScript.clip_profile(str(cfg["clip_profile"]))
		var ap := node.find_child("AnimationPlayer", true, false) as AnimationPlayer
		for logical in (profile["clips"] as Dictionary):
			if ap == null or not ap.has_animation(logical):
				_fail("%s must alias logical clip %s" % [visual_key, logical])
				node.queue_free()
				return
		var mounts := node.find_children("Attach_*", "BoneAttachment3D", true, false)
		if mounts.size() != (cfg["attachments"] as Array).size():
			_fail("%s must mount every catalog attachment (%d/%d)" % [visual_key, mounts.size(), (cfg["attachments"] as Array).size()])
			node.queue_free()
			return
		node.queue_free()
		await process_frame
	_pass("every kit monster scene aliases its profile clips and mounts its attachments")


func _test_alias_does_not_mutate_kit_clip_loop() -> void:
	var visual_key := str(LoaderScript.visual_keys()[0])
	var node := _scene_for(visual_key)
	get_root().add_child(node)
	await process_frame
	var profile := LoaderScript.clip_profile(str(LoaderScript.monster(visual_key)["clip_profile"]))
	var ap := node.find_child("AnimationPlayer", true, false) as AnimationPlayer
	var kit_walk := str((profile["clips"] as Dictionary)["walk"])
	var ok := ap.get_animation("walk").loop_mode == Animation.LOOP_LINEAR and ap.get_animation(kit_walk) != ap.get_animation("walk")
	node.queue_free()
	await process_frame
	if not ok:
		_fail("walk alias must be a looping copy, not the kit clip itself")
		return
	_pass("aliases are looping copies of kit clips")


func _pass(msg: String) -> void:
	_pass_count += 1
	print("  ok   %s" % msg)


func _fail(msg: String) -> void:
	_fail_count += 1
	push_error("FAIL: %s" % msg)
	print("  FAIL %s" % msg)


func _finish() -> void:
	if _fail_count > 0:
		print("[gdtest] FAIL: test_kit_monsters (%d failed)" % _fail_count)
		quit(1)
		return
	print("[gdtest] PASS: test_kit_monsters (%d checks)" % _pass_count)
	quit(0)
