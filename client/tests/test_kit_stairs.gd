extends SceneTree

## v489 kit stairs: node structure, scale and state tint derived from dungeon_kit_presentation.v0.json.

const LoaderScript := preload("res://scripts/dungeon_kit_presentation_loader.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_up_and_down_use_catalog_pieces()
	_test_down_has_pit_inside_hatch_footprint()
	_test_state_tint_keeps_texture()
	_test_disabled_catalog_falls_back_to_primitive()
	_finish()


func _cfg() -> Dictionary:
	return LoaderScript.stairs_config()


func _first_mesh(root: Node) -> MeshInstance3D:
	if root is MeshInstance3D:
		return root as MeshInstance3D
	for child in root.get_children():
		var found := _first_mesh(child)
		if found != null:
			return found
	return null


func _test_up_and_down_use_catalog_pieces() -> void:
	for def_id in ["stairs_up", "stairs_down"]:
		var node := TownNodeFactory.make_stair_node(def_id)
		var model := node.find_child(KitStairs.MODEL_NAME, true, false) as Node3D
		_assert_true("%s uses the kit model" % def_id, model != null and node.has_meta("kit_stairs"))
		if model != null:
			var key := "down_scale" if def_id == "stairs_down" else "up_scale"
			_assert_true("%s scale from catalog" % def_id, is_equal_approx(model.scale.x, float(_cfg()[key])))
		node.free()


func _test_down_has_pit_inside_hatch_footprint() -> void:
	var node := TownNodeFactory.make_stair_node("stairs_down")
	var pit := node.find_child(KitStairs.PIT_NAME, true, false) as MeshInstance3D
	_assert_true("stairs_down has a pit face", pit != null)
	if pit != null:
		var box := KitPieceLibrary.bounds(str(_cfg()["down_asset_id"]))
		var footprint := box.size.x * float(_cfg()["down_scale"])
		var size: Vector3 = (pit.mesh as BoxMesh).size
		_assert_true("pit is inset inside the hatch", size.x < footprint and is_equal_approx(size.x, footprint * float(_cfg()["down_pit_inset"])))
		_assert_true("pit colour from catalog", (pit.material_override as StandardMaterial3D).albedo_color.is_equal_approx(Color(str(_cfg()["down_pit_color"]))))
	_assert_true("stairs_up has no pit", TownNodeFactory.make_stair_node("stairs_up").find_child(KitStairs.PIT_NAME, true, false) == null)
	node.free()


func _test_state_tint_keeps_texture() -> void:
	var node := TownNodeFactory.make_stair_node("stairs_up")
	var mesh := _first_mesh(node.find_child(KitStairs.MODEL_NAME, true, false))
	var source := mesh.mesh.surface_get_material(0) as StandardMaterial3D if mesh != null and mesh.mesh != null else null
	KitStairs.apply_state(node, "stairs_up", "locked")
	var locked := mesh.material_override as StandardMaterial3D if mesh != null else null
	_assert_true("locked tint from catalog", locked != null and locked.albedo_color.is_equal_approx(Color(str(_cfg()["locked_tint"]))))
	_assert_true("tint keeps the kit texture", locked != null and source != null and locked.albedo_texture == source.albedo_texture and locked.albedo_texture != null)
	KitStairs.apply_state(node, "stairs_up", "ready")
	var ready := mesh.material_override as StandardMaterial3D if mesh != null else null
	_assert_true("ready tint restored", ready != null and ready.albedo_color.is_equal_approx(Color(str(_cfg()["ready_tint"]))))
	node.free()


func _test_disabled_catalog_falls_back_to_primitive() -> void:
	var previous: String = LoaderScript.enabled_override
	LoaderScript.enabled_override = "off"
	var node := TownNodeFactory.make_stair_node("stairs_down")
	_assert_true("disabled kit falls back to the procedural stairs", not node.has_meta("kit_stairs") and node.find_child("DownPitOpening", true, false) != null)
	node.free()
	LoaderScript.enabled_override = previous


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL %s" % label)


func _finish() -> void:
	if _fail_count > 0:
		printerr("[gdtest] FAIL: test_kit_stairs (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)
		return
	print("[gdtest] PASS: test_kit_stairs (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit()
