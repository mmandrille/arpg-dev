extends SceneTree

const DressingScript := preload("res://scripts/dungeon_room_dressing.gd")
const LoaderScript := preload("res://scripts/dungeon_kit_presentation_loader.gd")
const FloorScript := preload("res://scripts/dungeon_kit_floor.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")
const RendererScript := preload("res://scripts/wall_renderer.gd")
const FactoryScript := preload("res://scripts/ground_wall_factory.gd")

var _passes := 0
var _failures := 0


func _initialize() -> void:
	_test_repeatability_and_floor_variation()
	_test_catalog_controls()
	_test_clearances()
	_test_no_safe_area()
	_test_batched_build_and_teardown()
	if _failures > 0:
		print("[gdtest] FAIL: test_dungeon_room_dressing (%d failures)" % _failures)
		quit(1)
	else:
		print("[gdtest] PASS: test_dungeon_room_dressing (%d checks)" % _passes)
		quit(0)


func _walls(span: float = 28.0) -> Array:
	return [
		{"id": "north", "position": {"x": span * 0.5, "y": -0.5}, "size": {"x": span + 1.0, "y": 1.0}, "source": "perimeter"},
		{"id": "south", "position": {"x": span * 0.5, "y": span + 0.5}, "size": {"x": span + 1.0, "y": 1.0}, "source": "perimeter"},
		{"id": "west", "position": {"x": -0.5, "y": span * 0.5}, "size": {"x": 1.0, "y": span}, "source": "perimeter"},
		{"id": "east", "position": {"x": span + 0.5, "y": span * 0.5}, "size": {"x": 1.0, "y": span}, "source": "perimeter"},
		{"id": "column", "position": {"x": span * 0.5, "y": span * 0.5}, "size": {"x": 2.0, "y": 2.0}, "kind": "column", "source": "generated"},
	]


func _anchors() -> Array:
	return [
		{"id": "up", "type": "interactable", "interactable_def_id": "stairs_up", "position": {"x": 4.0, "y": 4.0}},
		{"id": "down", "type": "interactable", "interactable_def_id": "stairs_down", "position": {"x": 24.0, "y": 24.0}},
		{"id": "chest", "type": "interactable", "interactable_def_id": "treasure_chest", "position": {"x": 23.0, "y": 4.0}},
		{"id": "door", "type": "interactable", "interactable_def_id": "wooden_door", "position": {"x": 12.0, "y": 4.0}},
	]


func _cfg() -> Dictionary:
	return LoaderScript.dressing_config()


func _test_repeatability_and_floor_variation() -> void:
	var cfg := _cfg()
	var a := DressingScript.plan(_walls(), "seed-a|-1", -1, _anchors(), cfg)
	var again := DressingScript.plan(_walls(), "seed-a|-1", -1, _anchors(), cfg)
	var b := DressingScript.plan(_walls(), "seed-b|-1", -1, _anchors(), cfg)
	if a != again or a["placements"].is_empty() or a["placements"] == b["placements"]:
		_fail("same floor must repeat; a different floor key must vary")
		return
	var deep := DressingScript.plan(_walls(), "seed-a|-4", -4, _anchors(), cfg)
	if deep["placements"].size() < a["placements"].size():
		_fail("dense depth band must place at least as many props as sparse band")
		return
	_pass("stable floor key and depth bands")


func _test_catalog_controls() -> void:
	var cfg := _cfg()
	cfg["enabled"] = false
	if DressingScript.plan(_walls(), "key", -1, _anchors(), cfg)["reason"] != "disabled":
		_fail("disabled catalog must place nothing")
		return
	cfg["enabled"] = true
	cfg["max_instances"] = 2
	cfg["props"] = [cfg["props"][0]]
	var plan := DressingScript.plan(_walls(), "key", -4, _anchors(), cfg)
	if plan["placements"].size() > 2 or plan["placements"].is_empty():
		_fail("global cap must limit a nonempty plan")
		return
	for placed in plan["placements"]:
		if placed["asset_id"] != cfg["props"][0]["asset_id"]:
			_fail("single weighted asset must own every placement")
			return
	_pass("catalog toggle, weights and cap")


func _test_clearances() -> void:
	var cfg := _cfg()
	var walls := _walls()
	var anchors := DressingScript._anchor_points(_anchors())
	var radius := DressingScript._max_prop_radius(cfg["props"])
	if DressingScript._safe_cell(Vector2(12.0, 4.0), walls, anchors, cfg, radius) or DressingScript._safe_cell(Vector2(6.0, 6.0), walls, anchors, cfg, radius) or DressingScript._safe_cell(Vector2(14.0, 14.0), walls, anchors, cfg, radius):
		_fail("door, stair route and column must exclude candidate cells")
		return
	var plan := DressingScript.plan(walls, "clearance|-4", -4, _anchors(), cfg)
	if plan["placements"].is_empty():
		_fail("clearance fixture needs placements")
		return
	for placed in plan["placements"]:
		var point: Vector2 = placed["point"]
		if not DressingScript._safe_cell(point, walls, anchors, cfg, radius):
			_fail("planned prop violates wall, anchor or route clearance")
			return
		for wall in walls:
			if FloorScript._rect(wall).has_point(point):
				_fail("planned prop overlaps a wall or column")
				return
	_pass("wall, column, anchor and route clearances")


func _test_no_safe_area() -> void:
	var cfg := _cfg()
	cfg["wall_clearance_m"] = 100.0
	var plan := DressingScript.plan(_walls(8.0), "empty", -1, [], cfg)
	if plan["reason"] != "no_safe_candidate" or plan["safe_candidates"] != 0 or not plan["placements"].is_empty():
		_fail("unsafe floor must report no safe candidate and remain clear")
		return
	_pass("no-safe-candidate case reported")


func _test_batched_build_and_teardown() -> void:
	var cfg := _cfg()
	var placements: Array = DressingScript.plan(_walls(), "build|-4", -4, _anchors(), cfg)["placements"]
	var node := DressingScript.build(placements, float(cfg["surface_y"]))
	var count := 0
	for child in node.get_children():
		if not child is MultiMeshInstance3D:
			_fail("dressing child must be a MultiMesh without collision")
			node.free()
			return
		count += (child as MultiMeshInstance3D).multimesh.instance_count
	if count != placements.size() or node.get_child_count() > cfg["props"].size():
		_fail("batched instances must equal placements with at most one draw node per asset")
		node.free()
		return
	node.free()
	var root := Node3D.new()
	get_root().add_child(root)
	var renderer = RendererScript.new(root, FactoryScript.new())
	renderer.set_level(-1)
	renderer.render_wall_layout(_walls(), "build|-1", _anchors())
	var present := root.get_node_or_null(DressingScript.ROOT_NAME) != null
	renderer.set_level(0)
	renderer.render_wall_layout([])
	var cleared := root.get_node_or_null(DressingScript.ROOT_NAME) == null
	root.free()
	if not present or not cleared:
		_fail("renderer must attach dungeon dressing and remove it on town transition")
		return
	_pass("MultiMesh batch and level teardown")


func _pass(message: String) -> void:
	_passes += 1
	print("  ok   %s" % message)


func _fail(message: String) -> void:
	_failures += 1
	push_error(message)
	print("  FAIL %s" % message)
