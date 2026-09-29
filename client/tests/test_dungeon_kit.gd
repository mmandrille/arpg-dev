extends SceneTree
## ADR-0018 P2 (v471): KayKit dungeon walls, columns and floors.
## Dimensions are derived from the measured kit piece bounds, never duplicated literals.

const LoaderScript := preload("res://scripts/dungeon_kit_presentation_loader.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")
const WallBuilderScript := preload("res://scripts/dungeon_kit_wall_builder.gd")
const FloorScript := preload("res://scripts/dungeon_kit_floor.gd")
const WallRendererScript := preload("res://scripts/wall_renderer.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")

const EPS := 0.01

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_catalog_assets_resolve()
	_test_plan_segments_cover_length()
	_test_wall_visual_fills_rectangle(true)
	_test_wall_visual_fills_rectangle(false)
	_test_thick_wall_uses_rows()
	_test_column_matches_footprint()
	_test_floor_cells_skip_rectangles_and_are_deterministic()
	_test_renderer_kit_occlusion_fades_all_meshes()
	_test_disabled_kit_falls_back_to_box_walls()
	_finish()


func _full_len() -> float:
	return LibraryScript.bounds(str(LoaderScript.wall_asset_ids()["full"])).size.x


func _half_len() -> float:
	return LibraryScript.bounds(str(LoaderScript.wall_asset_ids()["half"])).size.x


func _depth() -> float:
	return LibraryScript.bounds(str(LoaderScript.wall_asset_ids()["full"])).size.z


func _test_catalog_assets_resolve() -> void:
	var ids := LoaderScript.wall_asset_ids()
	for asset_id in [ids["full"], ids["half"], LoaderScript.column_asset_id()]:
		if LibraryScript.scene(str(asset_id)) == null:
			_fail("kit asset %s must resolve to an imported scene" % asset_id)
			return
	for v in LoaderScript.floor_config().get("variants", []):
		if LibraryScript.mesh(str((v as Dictionary)["asset_id"])) == null:
			_fail("floor variant %s must resolve to a mesh" % str(v))
			return
	if _full_len() <= _half_len() or _half_len() <= 0.0:
		_fail("full wall piece must be longer than the half piece")
		return
	_pass("catalog asset ids resolve through the manifest")


func _test_plan_segments_cover_length() -> void:
	var full := _full_len()
	var half := _half_len()
	for length in [half * 0.4, half, full, full * 2.0 + half, full * 3.0 + half * 0.1, full + half * 0.6]:
		var segs := WallBuilderScript.plan_segments(length, full, half)
		var cursor := 0.0
		for seg in segs:
			if absf(float(seg["start"]) - cursor) > EPS:
				_fail("segments must be contiguous for length %.2f" % length)
				return
			cursor += float(seg["length"])
		if absf(cursor - length) > EPS:
			_fail("segments must cover length %.2f (got %.2f)" % [length, cursor])
			return
	_pass("segment plans cover wall lengths without gaps")


func _visual_aabb(root: Node3D) -> AABB:
	var box := AABB()
	var first := true
	for mi in LibraryScript.mesh_instances(root):
		var local := LibraryScript._transform_to(root, mi) * (mi as MeshInstance3D).get_aabb()
		box = local if first else box.merge(local)
		first = false
	return box


func _test_wall_visual_fills_rectangle(along_x: bool) -> void:
	var length := _full_len() * 2.0 + _half_len() * 0.5
	var thickness := _depth()
	var size := {"x": length, "y": thickness} if along_x else {"x": thickness, "y": length}
	var wall_height := 4.0
	var visual := WallBuilderScript.build_wall_visual({"id": "w", "position": {"x": 0.0, "y": 0.0}, "size": size}, wall_height)
	if visual == null:
		_fail("wall visual must build")
		return
	var box := _visual_aabb(visual)
	var want := Vector3(float(size["x"]), wall_height, float(size["y"]))
	if not box.size.is_equal_approx(want) and (box.size - want).length() > EPS * 4.0:
		_fail("wall visual %s must fill %s (got %s)" % ["x" if along_x else "z", str(want), str(box.size)])
		visual.free()
		return
	if absf(box.position.y) > EPS or box.get_center().distance_to(Vector3(0.0, wall_height * 0.5, 0.0)) > EPS * 4.0:
		_fail("wall visual must sit on the floor centered on the rectangle (got %s)" % str(box))
		visual.free()
		return
	visual.free()
	_pass("wall visual fills its rectangle (%s axis)" % ("x" if along_x else "z"))


func _test_thick_wall_uses_rows() -> void:
	var thin := WallBuilderScript.build_wall_visual({"size": {"x": _full_len(), "y": _depth()}}, 4.0)
	var thick := WallBuilderScript.build_wall_visual({"size": {"x": _full_len(), "y": _depth() * 2.0}}, 4.0)
	var ok := thick.get_child_count() == thin.get_child_count() * 2
	thin.free()
	thick.free()
	if not ok:
		_fail("a wall twice the kit depth must use two rows")
		return
	_pass("thick walls use parallel rows")


func _test_column_matches_footprint() -> void:
	var size := {"x": 1.2, "y": 3.4}
	var visual := WallBuilderScript.build_column_visual({"size": size}, 4.0)
	var box := _visual_aabb(visual)
	visual.free()
	if absf(box.size.x - float(size["x"])) > EPS or absf(box.size.z - float(size["y"])) > EPS:
		_fail("pillar must match column footprint (got %s)" % str(box.size))
		return
	_pass("kit pillar matches column footprint")


func _sample_layout() -> Array:
	return [
		{"id": "n", "position": {"x": 10.0, "y": -0.5}, "size": {"x": 22.0, "y": 1.0}, "source": "perimeter"},
		{"id": "s", "position": {"x": 10.0, "y": 12.5}, "size": {"x": 22.0, "y": 1.0}, "source": "perimeter"},
		{"id": "w", "position": {"x": -0.5, "y": 6.0}, "size": {"x": 1.0, "y": 12.0}, "source": "perimeter"},
		{"id": "e", "position": {"x": 20.5, "y": 6.0}, "size": {"x": 1.0, "y": 12.0}, "source": "perimeter"},
		{"id": "inner", "position": {"x": 8.0, "y": 6.0}, "size": {"x": 1.0, "y": 8.0}, "source": "generated"},
		{"id": "pit", "position": {"x": 15.0, "y": 4.0}, "size": {"x": 4.0, "y": 4.0}, "source": "generated", "kind": "hole"},
		{"id": "pool", "position": {"x": 4.0, "y": 9.0}, "size": {"x": 3.0, "y": 2.0}, "source": "generated", "kind": "water"},
	]


func _test_floor_cells_skip_rectangles_and_are_deterministic() -> void:
	var layout := _sample_layout()
	var tile := LibraryScript.bounds(str((LoaderScript.floor_config()["variants"] as Array)[0]["asset_id"])).size.x
	var cells := FloorScript.plan_cells(layout, tile)
	if cells.is_empty():
		_fail("floor must plan cells inside the perimeter")
		return
	var bounds := FloorScript.floor_bounds(layout)
	for cell in cells:
		if not bounds.has_point(cell):
			_fail("floor cell %s outside floor bounds" % str(cell))
			return
		for wall in layout:
			if FloorScript._rect(wall).has_point(cell):
				_fail("floor cell %s lies inside %s" % [str(cell), str(wall["id"])])
				return
	if FloorScript.plan_cells(layout, tile) != cells:
		_fail("floor cell plan must be deterministic")
		return
	var node := FloorScript.build(layout, -2)
	var instances := 0
	for child in node.get_children():
		instances += (child as MultiMeshInstance3D).multimesh.instance_count
	var picks_a: Array = []
	var picks_b: Array = []
	for cell in cells:
		picks_a.append(FloorScript.pick(cell, -2, [5, 1, 1]))
		picks_b.append(FloorScript.pick(cell, -2, [5, 1, 1]))
	node.free()
	if instances != cells.size():
		_fail("floor instances (%d) must equal planned cells (%d)" % [instances, cells.size()])
		return
	if picks_a != picks_b:
		_fail("floor variant picks must be deterministic")
		return
	_pass("floor tiles skip wall/hole/water rectangles deterministically")


func _test_renderer_kit_occlusion_fades_all_meshes() -> void:
	var root := Node3D.new()
	get_root().add_child(root)
	var renderer = WallRendererScript.new(root, GroundWallFactoryScript.new())
	renderer.set_level(-2)
	renderer.render_wall_layout([{"id": "fade_wall", "position": {"x": 4.0, "y": 4.0}, "size": {"x": _full_len() * 2.0, "y": 1.0}, "source": "generated"}])
	if not renderer.kit_active():
		_fail("kit must be active on dungeon level -2")
		root.queue_free()
		return
	var body := root.get_child(0) as StaticBody3D
	var meshes := LibraryScript.mesh_instances(body)
	if meshes.size() < 2 or not (body.get_child(0) is CollisionShape3D):
		_fail("kit wall body keeps collision and carries kit meshes (got %d meshes)" % meshes.size())
		root.queue_free()
		return
	renderer.apply_occlusion_fades({"fade_wall": 0.3})
	for mi in meshes:
		var mat := (mi as MeshInstance3D).material_override as StandardMaterial3D
		if mat == null or mat.albedo_color.a > 0.5:
			_fail("every kit mesh of a faded wall must fade")
			root.queue_free()
			return
	renderer.apply_occlusion_fades({})
	for mi in meshes:
		var mat := (mi as MeshInstance3D).material_override as StandardMaterial3D
		if mat != null and mat.albedo_color.a < 0.999:
			_fail("kit meshes must restore to opaque")
			root.queue_free()
			return
	root.queue_free()
	_pass("occlusion fade reaches every kit mesh and restores")


func _test_disabled_kit_falls_back_to_box_walls() -> void:
	LoaderScript.enabled_override = "off"
	var renderer = WallRendererScript.new(null, GroundWallFactoryScript.new())
	renderer.set_level(-2)
	var body: Node3D = renderer.make_wall_node({"id": "legacy", "position": {"x": 0.0, "y": 0.0}, "size": {"x": 3.0, "y": 1.0}, "source": "generated"})
	LoaderScript.enabled_override = ""
	var mesh := body.get_child(1) as MeshInstance3D
	var ok := mesh != null and mesh.mesh is BoxMesh
	body.free()
	if not ok:
		_fail("disabled kit must fall back to the legacy BoxMesh wall")
		return
	_pass("disabled kit falls back to legacy walls")


func _pass(msg: String) -> void:
	_pass_count += 1
	print("  ok   %s" % msg)


func _fail(msg: String) -> void:
	_fail_count += 1
	push_error("FAIL: %s" % msg)
	print("  FAIL %s" % msg)


func _finish() -> void:
	if _fail_count > 0:
		print("[gdtest] FAIL: test_dungeon_kit (%d failed)" % _fail_count)
		quit(1)
		return
	print("[gdtest] PASS: test_dungeon_kit (%d checks)" % _pass_count)
	quit(0)
