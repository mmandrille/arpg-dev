extends SceneTree

## v491 town dressing: live-ground sync (world-aligned, removed below town), plaza coverage and
## prop placement, all derived from town_presentation.v0.json.

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_sync_attaches_world_aligned_in_town_only()
	_test_plaza_cells_inside_disc_or_road()
	_finish()


func _test_sync_attaches_world_aligned_in_town_only() -> void:
	var factory := GroundWallFactory.new()
	var ground := factory.make_ground_node(0)
	factory.update_ground_material(ground, 0)
	var root := ground.find_child(TownDressing.ROOT_NAME, false, false) as Node3D
	_assert_true("town ground has the dressing root", root != null)
	if root != null:
		_assert_true("dressing root holds the ground detail", root.find_child(TownGroundDetail.ROOT_NAME, false, false) != null)
		_assert_true("dressing root cancels the ground offset", (ground.position + root.position).is_equal_approx(Vector3.ZERO))
		var props: Array = TownPresentationLoader.dressing().get("props", [])
		var first: Dictionary = props[0]
		var node := root.find_child(TownDressing.prop_node_name(0, str(first["asset_id"])), false, false) as Node3D
		var want := Vector3(float(first["position"]["x"]), 0.0, float(first["position"]["y"]))
		_assert_true("first prop at its catalog world position", node != null and (ground.position + root.position + node.position).is_equal_approx(want))
	factory.update_ground_material(ground, -1)
	_assert_true("dressing removed on a dungeon level", ground.find_child(TownDressing.ROOT_NAME, false, false) == null or ground.find_child(TownDressing.ROOT_NAME, false, false).is_queued_for_deletion())
	ground.free()


func _test_plaza_cells_inside_disc_or_road() -> void:
	var plaza: Dictionary = TownPresentationLoader.dressing().get("plaza", {})
	var center := TownPresentationLoader.center()
	var gate := TownPresentationLoader.gate_position()
	var cells := TownDressing.plaza_cells(plaza, 2.0)
	_assert_true("plaza has cells", cells.size() > 0)
	var radius := float(plaza["radius_m"])
	var half := float(plaza["path_width_m"]) * 0.5
	var bad := 0
	var reaches_gate := false
	for cell in cells:
		if not TownDressing.in_plaza(cell, center, gate, radius, half):
			bad += 1
		if (cell as Vector2).distance_to(gate) <= 2.0:
			reaches_gate = true
	_assert_true("every cell is inside the disc or road", bad == 0)
	_assert_true("the road reaches the gate", reaches_gate)
	_assert_true("the centre is paved", cells.has(center))


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL %s" % label)


func _finish() -> void:
	if _fail_count > 0:
		printerr("[gdtest] FAIL: test_town_dressing (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)
		return
	print("[gdtest] PASS: test_town_dressing (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit()
