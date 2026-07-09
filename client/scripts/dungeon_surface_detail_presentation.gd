class_name DungeonSurfaceDetailPresentation
extends RefCounted

const DungeonRoomPresentationLoaderScript := preload("res://scripts/dungeon_room_presentation_loader.gd")
const DungeonSurfaceDetailLoaderScript := preload("res://scripts/dungeon_surface_detail_loader.gd")

const FLOOR_ROOT_NAME := "DungeonSurfaceFloorDetails"
const WALL_ROOT_NAME := "DungeonSurfaceWallDetails"
const ELIGIBLE_WALL_SOURCES := {
	"generated": true,
	"room_wall": true,
	"room_divider": true,
	"perimeter": true,
}
const ARCHETYPE_CYCLE := ["combat", "corridor", "rest"]


static func sync(
	ground_node: MeshInstance3D,
	walls_root: Node3D,
	factory: RefCounted,
	level: int,
	walls: Array,
	entities: Dictionary,
) -> void:
	_clear(ground_node, walls_root)
	if ground_node == null or walls_root == null or factory == null or level >= 0:
		return
	if not factory.has_method("floor_size_for_level"):
		return
	DungeonSurfaceDetailLoaderScript.ensure_loaded()
	DungeonRoomPresentationLoaderScript.ensure_loaded()
	var floor_size: Vector2 = factory.floor_size_for_level(level)
	if floor_size.x <= 0.0 or floor_size.y <= 0.0:
		return
	_attach_floor_details(ground_node, floor_size, walls, entities)
	_attach_wall_details(walls_root, factory, level, walls)


static func _clear(ground_node: MeshInstance3D, walls_root: Node3D) -> void:
	if ground_node != null:
		var floor_root := ground_node.get_node_or_null(FLOOR_ROOT_NAME)
		if floor_root != null:
			ground_node.remove_child(floor_root)
			floor_root.queue_free()
	if walls_root != null:
		var wall_root := walls_root.get_node_or_null(WALL_ROOT_NAME)
		if wall_root != null:
			walls_root.remove_child(wall_root)
			wall_root.queue_free()


static func _attach_floor_details(ground_node: MeshInstance3D, floor_size: Vector2, walls: Array, entities: Dictionary) -> void:
	var x_lines := _axis_lines(floor_size.x, walls, true)
	var z_lines := _axis_lines(floor_size.y, walls, false)
	var treasure_points := _treasure_points(entities)
	var motifs: Dictionary = DungeonSurfaceDetailLoaderScript.floor_motifs()
	var motif_names: Array = motifs.keys()
	if motif_names.is_empty():
		return
	var root := Node3D.new()
	root.name = FLOOR_ROOT_NAME
	ground_node.add_child(root)
	var cell_index := 0
	for xi in range(x_lines.size() - 1):
		for zi in range(z_lines.size() - 1):
			var x0 := float(x_lines[xi])
			var x1 := float(x_lines[xi + 1])
			var z0 := float(z_lines[zi])
			var z1 := float(z_lines[zi + 1])
			var width := x1 - x0
			var depth := z1 - z0
			if width < 2.0 or depth < 2.0:
				continue
			var center := Vector2((x0 + x1) * 0.5, (z0 + z1) * 0.5)
			var archetype := _archetype_for_cell(center, width, depth, treasure_points, cell_index)
			var motif_name := str(motif_names[cell_index % motif_names.size()])
			var motif: Dictionary = motifs.get(motif_name, {}) as Dictionary
			var large_every: int = DungeonSurfaceDetailLoaderScript.corridor_large_detail_every() if archetype == "corridor" else DungeonSurfaceDetailLoaderScript.room_large_detail_every()
			var large: bool = (cell_index % maxi(1, large_every)) == 0
			root.add_child(_make_floor_detail_node(width, depth, center, motif_name, motif, large, cell_index))
			cell_index += 1


static func _make_floor_detail_node(
	width: float,
	depth: float,
	center: Vector2,
	motif_name: String,
	motif: Dictionary,
	large: bool,
	index: int,
) -> Node3D:
	var root := Node3D.new()
	root.name = "FloorDetail_%02d" % index
	root.position = Vector3(center.x, DungeonSurfaceDetailLoaderScript.floor_detail_y(), center.y)
	root.set_meta("motif", motif_name)
	var footprint := minf(width, depth)
	var scale_min := float(motif.get("scale_min", 0.18))
	var scale_max := float(motif.get("scale_max", 0.42))
	var detail_scale := footprint * (scale_max if large else (scale_min + scale_max) * 0.5)
	root.set_meta("detail_scale", detail_scale)
	var color := Color(str(motif.get("color", "#d2c0a0")))
	var alpha := float(motif.get("alpha", 0.18))
	var band_width := maxf(0.04, detail_scale * float(motif.get("band_width", 0.12)))
	match motif_name:
		"ritual_square":
			_add_floor_band(root, Vector2(detail_scale, band_width), Vector3(0.0, 0.0, -detail_scale * 0.34), color, alpha)
			_add_floor_band(root, Vector2(detail_scale, band_width), Vector3(0.0, 0.0, detail_scale * 0.34), color, alpha)
			_add_floor_band(root, Vector2(band_width, detail_scale), Vector3(-detail_scale * 0.34, 0.0, 0.0), color, alpha)
			_add_floor_band(root, Vector2(band_width, detail_scale), Vector3(detail_scale * 0.34, 0.0, 0.0), color, alpha)
			_add_floor_band(root, Vector2(detail_scale * 0.52, band_width * 0.9), Vector3(0.0, 0.0, 0.0), color.lightened(0.08), alpha * 0.92)
		"broken_tile":
			_add_floor_band(root, Vector2(detail_scale * 0.82, detail_scale * 0.56), Vector3(0.0, 0.0, 0.0), color, alpha * 0.8)
			_add_floor_band(root, Vector2(detail_scale * 0.74, band_width), Vector3(0.0, 0.0, 0.0), color.darkened(0.14), alpha)
			_add_floor_band(root, Vector2(band_width, detail_scale * 0.58), Vector3(detail_scale * 0.12, 0.0, 0.0), color.darkened(0.08), alpha * 0.9)
		_:
			_add_floor_band(root, Vector2(detail_scale, band_width), Vector3(0.0, 0.0, 0.0), color, alpha)
			_add_floor_band(root, Vector2(detail_scale * 0.82, band_width), Vector3(0.0, 0.0, 0.0), color, alpha * 0.9, 55.0)
			_add_floor_band(root, Vector2(detail_scale * 0.52, band_width * 0.9), Vector3(0.0, 0.0, -detail_scale * 0.18), color.darkened(0.1), alpha * 0.85, -25.0)
	return root


static func _add_floor_band(root: Node3D, size: Vector2, position: Vector3, color: Color, alpha: float, yaw_degrees: float = 0.0) -> void:
	var band := MeshInstance3D.new()
	var mesh := PlaneMesh.new()
	mesh.size = size
	band.mesh = mesh
	band.position = position
	band.rotation_degrees = Vector3(-90.0, yaw_degrees, 0.0)
	band.material_override = _overlay_material(color, alpha)
	root.add_child(band)


static func _attach_wall_details(walls_root: Node3D, factory: RefCounted, level: int, walls: Array) -> void:
	var motifs: Dictionary = DungeonSurfaceDetailLoaderScript.wall_motifs()
	var motif_names: Array = motifs.keys()
	if motif_names.is_empty():
		return
	var wall_height := 1.0
	if factory.has_method("dungeon_ceiling_height"):
		wall_height = float(factory.dungeon_ceiling_height())
	var root := Node3D.new()
	root.name = WALL_ROOT_NAME
	walls_root.add_child(root)
	var detail_index := 0
	for wall in walls:
		if typeof(wall) != TYPE_DICTIONARY:
			continue
		var rec: Dictionary = wall
		if not _eligible_wall(rec):
			continue
		var orientation := _orientation(rec)
		if orientation == "":
			continue
		var motif_name := str(motif_names[detail_index % motif_names.size()])
		var motif: Dictionary = motifs.get(motif_name, {}) as Dictionary
		root.add_child(_make_wall_detail_node(rec, wall_height, motif_name, motif, orientation, detail_index))
		detail_index += 1
	if detail_index == 0:
		walls_root.remove_child(root)
		root.queue_free()


static func _make_wall_detail_node(
	wall: Dictionary,
	wall_height: float,
	motif_name: String,
	motif: Dictionary,
	orientation: String,
	index: int,
) -> Node3D:
	var root := Node3D.new()
	root.name = "WallDetail_%02d" % index
	root.set_meta("motif", motif_name)
	root.set_meta("wall_id", str(wall.get("id", "")))
	var center := _center(wall)
	var size := _size(wall)
	var width := size.x if orientation == "horizontal" else size.y
	var face_width := maxf(0.3, width * float(motif.get("width_ratio", 0.34)))
	var face_height := maxf(0.24, wall_height * float(motif.get("height_ratio", 0.2)))
	root.set_meta("detail_width", face_width)
	root.position = Vector3(center.x, wall_height * DungeonSurfaceDetailLoaderScript.wall_detail_height_ratio(), center.y)
	var color := Color(str(motif.get("color", "#73624d")))
	var alpha := float(motif.get("alpha", 0.9))
	var depth := float(motif.get("depth", 0.08))
	var inset: float = DungeonSurfaceDetailLoaderScript.wall_detail_inset()
	var face_offset := _wall_face_offset(size, orientation, inset)
	match motif_name:
		"crack_cluster":
			_add_wall_panel(root, Vector3(face_width, face_height * 0.16, depth), face_offset, color, alpha)
			_add_wall_panel(root, Vector3(face_width * 0.72, face_height * 0.12, depth), face_offset + Vector3(0.0, face_height * 0.2, 0.0), color.lightened(0.04), alpha * 0.92, 28.0 if orientation == "horizontal" else 0.0, 28.0 if orientation == "vertical" else 0.0)
			_add_wall_panel(root, Vector3(face_width * 0.48, face_height * 0.1, depth), face_offset + Vector3(0.0, -face_height * 0.2, 0.0), color.lightened(0.06), alpha * 0.86, -22.0 if orientation == "horizontal" else 0.0, -22.0 if orientation == "vertical" else 0.0)
		"stone_inset":
			_add_wall_panel(root, Vector3(face_width, face_height, depth), face_offset, color, alpha * 0.9)
			_add_wall_panel(root, Vector3(face_width * 0.68, face_height * 0.56, depth * 0.7), face_offset + Vector3(0.0, 0.0, 0.0), color.darkened(0.12), alpha * 0.84)
		_:
			_add_wall_panel(root, Vector3(face_width, face_height, depth), face_offset, color, alpha)
			_add_wall_panel(root, Vector3(face_width * 0.78, face_height * 0.18, depth * 0.6), face_offset + Vector3(0.0, -face_height * 0.18, 0.0), color.lightened(0.08), alpha * 0.84)
	if orientation == "vertical":
		root.rotation_degrees.y = 90.0
	return root


static func _add_wall_panel(root: Node3D, size: Vector3, position: Vector3, color: Color, alpha: float, rot_z: float = 0.0, rot_x: float = 0.0) -> void:
	var mesh_node := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = size
	mesh_node.mesh = mesh
	mesh_node.position = position
	mesh_node.rotation_degrees = Vector3(rot_x, 0.0, rot_z)
	mesh_node.material_override = _solid_material(color, alpha)
	root.add_child(mesh_node)


static func _wall_face_offset(size: Vector2, orientation: String, inset: float) -> Vector3:
	if orientation == "horizontal":
		return Vector3(0.0, 0.0, size.y * 0.5 - inset)
	return Vector3(size.x * 0.5 - inset, 0.0, 0.0)


static func _eligible_wall(wall: Dictionary) -> bool:
	var kind := str(wall.get("kind", "wall"))
	if kind != "" and kind != "wall":
		return false
	return ELIGIBLE_WALL_SOURCES.has(str(wall.get("source", "")))


static func _axis_lines(span: float, walls: Array, horizontal_axis: bool) -> PackedFloat32Array:
	var values := PackedFloat32Array([0.0, span])
	for wall in walls:
		if typeof(wall) != TYPE_DICTIONARY:
			continue
		var rec: Dictionary = wall
		if str(rec.get("source", "")) != "room_divider":
			continue
		var size := _size(rec)
		if horizontal_axis:
			if size.x >= size.y:
				values.append(_center(rec).y)
		else:
			if size.y >= size.x:
				values.append(_center(rec).x)
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


static func _archetype_for_cell(center: Vector2, width: float, depth: float, treasure_points: Array[Vector2], cycle_index: int) -> String:
	for point in treasure_points:
		if center.distance_to(point) <= DungeonRoomPresentationLoaderScript.treasure_radius():
			return "treasure"
	if maxf(width, depth) <= DungeonRoomPresentationLoaderScript.corridor_max_span():
		return "corridor"
	return ARCHETYPE_CYCLE[cycle_index % ARCHETYPE_CYCLE.size()]


static func _center(wall: Dictionary) -> Vector2:
	var pos: Dictionary = wall.get("position", {})
	return Vector2(float(pos.get("x", 0.0)), float(pos.get("y", 0.0)))


static func _size(wall: Dictionary) -> Vector2:
	var size: Dictionary = wall.get("size", {})
	return Vector2(float(size.get("x", 1.0)), float(size.get("y", 1.0)))


static func _orientation(wall: Dictionary) -> String:
	var size := _size(wall)
	if size.x > size.y:
		return "horizontal"
	if size.y > size.x:
		return "vertical"
	return ""


static func _overlay_material(color: Color, alpha: float) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(color.r, color.g, color.b, alpha)
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	mat.cull_mode = BaseMaterial3D.CULL_DISABLED
	return mat


static func _solid_material(color: Color, alpha: float) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(color.r, color.g, color.b, alpha)
	mat.roughness = 0.95
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA if alpha < 1.0 else BaseMaterial3D.TRANSPARENCY_DISABLED
	return mat
