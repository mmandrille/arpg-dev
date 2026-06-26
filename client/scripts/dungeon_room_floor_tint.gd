class_name DungeonRoomFloorTint
extends RefCounted

const ROOT_NAME := "DungeonRoomFloorTint"
const ARCHETYPE_CYCLE := ["combat", "corridor", "rest"]


static func sync(
	ground_node: MeshInstance3D,
	factory: RefCounted,
	level: int,
	walls: Array,
	entities: Dictionary,
) -> void:
	if ground_node == null or factory == null or level >= 0:
		_clear(ground_node)
		return
	DungeonRoomPresentationLoader.ensure_loaded()
	_clear(ground_node)
	if walls.is_empty() or not factory.has_method("floor_size_for_level"):
		return
	var floor_size: Vector2 = factory.floor_size_for_level(level)
	if floor_size.x <= 0.0 or floor_size.y <= 0.0:
		return
	var x_lines := _axis_lines(floor_size.x, walls, true)
	var z_lines := _axis_lines(floor_size.y, walls, false)
	var treasure_points := _treasure_points(entities)
	var root := Node3D.new()
	root.name = ROOT_NAME
	ground_node.add_child(root)
	var cycle_index := 0
	for xi in range(x_lines.size() - 1):
		for zi in range(z_lines.size() - 1):
			var x0 := float(x_lines[xi])
			var x1 := float(x_lines[xi + 1])
			var z0 := float(z_lines[zi])
			var z1 := float(z_lines[zi + 1])
			var width := x1 - x0
			var depth := z1 - z0
			if width < 0.5 or depth < 0.5:
				continue
			var center := Vector2((x0 + x1) * 0.5, (z0 + z1) * 0.5)
			var archetype := _archetype_for_cell(center, width, depth, treasure_points, cycle_index)
			cycle_index += 1
			root.add_child(_make_overlay(width, depth, Vector3(center.x, 0.018, center.y), archetype))


static func _clear(ground_node: MeshInstance3D) -> void:
	if ground_node == null:
		return
	var existing := ground_node.get_node_or_null(ROOT_NAME)
	if existing != null:
		ground_node.remove_child(existing)
		existing.queue_free()


static func _axis_lines(span: float, walls: Array, horizontal_axis: bool) -> PackedFloat32Array:
	var values := PackedFloat32Array([0.0, span])
	for wall in walls:
		if typeof(wall) != TYPE_DICTIONARY:
			continue
		var rec: Dictionary = wall
		if str(rec.get("source", "")) != "room_divider":
			continue
		var pos: Dictionary = rec.get("position", {})
		var size: Dictionary = rec.get("size", {})
		var sx := float(size.get("x", 1.0))
		var sy := float(size.get("y", 1.0))
		if horizontal_axis:
			if sx >= sy:
				values.append(float(pos.get("y", 0.0)))
		else:
			if sy >= sx:
				values.append(float(pos.get("x", 0.0)))
	values.sort()
	var out := PackedFloat32Array()
	var last := -1.0
	for value in values:
		if absf(value - last) > 0.25:
			out.append(value)
			last = value
	return out


static func _treasure_points(entities: Dictionary) -> Array[Vector2]:
	var points: Array[Vector2] = []
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		if str(rec.get("type", "")) != "interactable":
			continue
		var def_id := str(rec.get("interactable_def_id", ""))
		if def_id != "treasure_chest" and def_id != "elite_objective_chest":
			continue
		var pos: Dictionary = rec.get("position", {})
		points.append(Vector2(float(pos.get("x", 0.0)), float(pos.get("y", 0.0))))
	return points


static func _archetype_for_cell(
	center: Vector2,
	width: float,
	depth: float,
	treasure_points: Array[Vector2],
	cycle_index: int,
) -> String:
	for point in treasure_points:
		if center.distance_to(point) <= DungeonRoomPresentationLoader.treasure_radius():
			return "treasure"
	var max_span := maxf(width, depth)
	if max_span <= DungeonRoomPresentationLoader.corridor_max_span():
		return "corridor"
	return ARCHETYPE_CYCLE[cycle_index % ARCHETYPE_CYCLE.size()]


static func _make_overlay(width: float, depth: float, position: Vector3, archetype: String) -> MeshInstance3D:
	var cfg: Dictionary = DungeonRoomPresentationLoader.archetype(archetype)
	var tint := Color(str(cfg.get("floor_tint", "#6b6058")))
	var alpha := float(cfg.get("alpha", 0.14))
	var node := MeshInstance3D.new()
	node.name = "RoomTint_%s" % archetype
	var mesh := PlaneMesh.new()
	mesh.size = Vector2(width * 0.96, depth * 0.96)
	node.mesh = mesh
	node.position = position
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(tint.r, tint.g, tint.b, alpha)
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	mat.cull_mode = BaseMaterial3D.CULL_DISABLED
	node.material_override = mat
	return node
