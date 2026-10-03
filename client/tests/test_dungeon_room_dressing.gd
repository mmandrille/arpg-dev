extends SceneTree

const DressingScript := preload("res://scripts/dungeon_room_dressing.gd")
const LoaderScript := preload("res://scripts/dungeon_kit_presentation_loader.gd")
const RendererScript := preload("res://scripts/wall_renderer.gd")
const FactoryScript := preload("res://scripts/ground_wall_factory.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")

var _passes := 0
var _failures := 0


func _initialize() -> void:
	LoaderScript.enabled_override = "on"
	_test_placements_follow_server_walls()
	_test_catalog_matches_server_catalog()
	_test_renderer_batches_props_and_leaves_floor_alone()
	if _failures > 0:
		print("[gdtest] FAIL: test_dungeon_room_dressing (%d failures)" % _failures)
		quit(1)
	else:
		print("[gdtest] PASS: test_dungeon_room_dressing (%d checks)" % _passes)
		quit(0)


func _structure(span: float = 28.0) -> Array:
	return [
		{"id": "north", "position": {"x": span * 0.5, "y": -0.5}, "size": {"x": span + 1.0, "y": 1.0}, "source": "perimeter"},
		{"id": "south", "position": {"x": span * 0.5, "y": span + 0.5}, "size": {"x": span + 1.0, "y": 1.0}, "source": "perimeter"},
		{"id": "west", "position": {"x": -0.5, "y": span * 0.5}, "size": {"x": 1.0, "y": span}, "source": "perimeter"},
		{"id": "east", "position": {"x": span + 0.5, "y": span * 0.5}, "size": {"x": 1.0, "y": span}, "source": "perimeter"},
	]


func _prop(id: String, prop_id: String, x: float, y: float) -> Dictionary:
	return {"id": id, "position": {"x": x, "y": y}, "size": {"x": 1.0, "y": 1.0}, "source": "generated", "kind": "prop", "prop_id": prop_id}


func _catalog_ids() -> Array:
	return (LoaderScript.dressing_config().get("props", {}) as Dictionary).keys()


func _test_placements_follow_server_walls() -> void:
	var cfg := LoaderScript.dressing_config()
	var ids := _catalog_ids()
	if ids.is_empty():
		_fail("presentation catalog must define props")
		return
	var walls := _structure()
	walls.append(_prop("prop_a", ids[0], 6.5, 9.0))
	walls.append(_prop("prop_b", ids[ids.size() - 1], 14.0, 3.5))
	walls.append(_prop("prop_unknown", "not_in_catalog", 20.0, 20.0))
	walls.append({"id": "rock", "position": {"x": 3.0, "y": 3.0}, "size": {"x": 2.0, "y": 2.0}, "source": "generated", "kind": "rock"})
	var placements := DressingScript.placements_from_walls(walls, cfg)
	if placements.size() != 2:
		_fail("only catalog props become placements (got %d)" % placements.size())
		return
	if placements[0]["point"] != Vector2(6.5, 9.0) or placements[1]["point"] != Vector2(14.0, 3.5):
		_fail("placement points must be the server positions")
		return
	var again := DressingScript.placements_from_walls(walls, cfg)
	if placements != again:
		_fail("placements must be stable for the same walls")
		return
	for placement in placements:
		var entry: Dictionary = (cfg["props"] as Dictionary)[ids[0] if placement == placements[0] else ids[ids.size() - 1]]
		if str(placement["asset_id"]) != str(entry["asset_id"]) or not (entry["yaw_degrees"] as Array).has(placement["yaw_degrees"]):
			_fail("placement model and yaw must come from the catalog entry")
			return
	var disabled := cfg.duplicate(true)
	disabled["enabled"] = false
	if not DressingScript.placements_from_walls(walls, disabled).is_empty():
		_fail("a disabled catalog places nothing")
		return
	_pass("placements follow server prop walls")


## The server catalog is the source of prop ids; the client must present every one of them.
func _test_catalog_matches_server_catalog() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/dungeon_generation.v0.json")
	var rules = JSON.parse_string(FileAccess.get_file_as_string(path))
	var server_ids: Array = []
	for entry in rules["obstacle_generation"]["props"]["catalog"]:
		server_ids.append(str(entry["prop_id"]))
	var client_ids := _catalog_ids()
	for id in server_ids:
		if not client_ids.has(id):
			_fail("client presentation is missing server prop %s" % id)
			return
	# The collision footprint should cover what the player sees: the rendered model may overhang the
	# server footprint by at most a sliver, and the footprint should not dwarf the model.
	var props: Dictionary = LoaderScript.dressing_config().get("props", {})
	for entry in rules["obstacle_generation"]["props"]["catalog"]:
		var shown: Dictionary = props[str(entry["prop_id"])]
		var box: AABB = LibraryScript.bounds(str(shown["asset_id"]))
		var model := maxf(box.size.x, box.size.z) * float(shown["scale"])
		var footprint := maxf(float(entry["footprint"]["x"]), float(entry["footprint"]["y"]))
		if model > footprint + 0.3 or footprint > model + 0.5:
			_fail("prop %s model extent %.2f does not match its collision footprint %.2f" % [str(entry["prop_id"]), model, footprint])
			return
	_pass("client presents every server prop id, with footprints matching the models")


func _test_renderer_batches_props_and_leaves_floor_alone() -> void:
	var cfg := LoaderScript.dressing_config()
	var ids := _catalog_ids()
	var plain := _structure()
	var with_props := _structure()
	for i in 6:
		with_props.append(_prop("prop_%d" % i, ids[i % ids.size()], 6.0 + float(i) * 3.0, 10.0))
	var root := Node3D.new()
	get_root().add_child(root)
	var renderer = RendererScript.new(root, FactoryScript.new())
	renderer.set_level(-1)
	renderer.render_wall_layout(with_props)
	var node := root.get_node_or_null(DressingScript.ROOT_NAME)
	if node == null:
		_fail("renderer must attach the dungeon props batch")
		root.free()
		return
	var count := 0
	for child in node.get_children():
		if not child is MultiMeshInstance3D:
			_fail("props must be MultiMeshes without client collision")
			root.free()
			return
		count += (child as MultiMeshInstance3D).multimesh.instance_count
	if count != 6 or node.get_child_count() > ids.size():
		_fail("batched instances must equal the prop walls with at most one draw node per asset")
		root.free()
		return
	var floor_with_props := _floor_tile_count(root)
	renderer.render_wall_layout(plain)
	if floor_with_props <= 0 or floor_with_props != _floor_tile_count(root):
		_fail("props must not remove floor tiles (%d with props)" % floor_with_props)
		root.free()
		return
	renderer.set_level(0)
	renderer.render_wall_layout([])
	var cleared := root.get_node_or_null(DressingScript.ROOT_NAME) == null
	root.free()
	if not cleared:
		_fail("props batch must be removed on town transition")
		return
	_pass("MultiMesh batch, structure-only floor, and level teardown")


func _floor_tile_count(root: Node3D) -> int:
	var count := 0
	var kit_floor := root.get_node_or_null("KitFloor")
	if kit_floor == null:
		return 0
	for child in kit_floor.get_children():
		if child is MultiMeshInstance3D:
			count += (child as MultiMeshInstance3D).multimesh.instance_count
	return count


func _pass(message: String) -> void:
	_passes += 1
	print("  ok   %s" % message)


func _fail(message: String) -> void:
	_failures += 1
	push_error(message)
	print("  FAIL %s" % message)
