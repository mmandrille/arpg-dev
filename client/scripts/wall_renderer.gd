class_name WallRenderer
extends RefCounted

const ClientConstantsScript := preload("res://scripts/client_constants.gd")

var _walls_root: Node3D
var _ground_factory: RefCounted
var _current_level: int = 0
var _ceiling_node: MeshInstance3D

const TOWN_WALL_HEIGHT := 1.0
const WALL_FLOOR_SEAM_OVERLAP := 0.08

func _init(walls_root: Node3D, ground_factory: RefCounted) -> void:
	_walls_root = walls_root
	_ground_factory = ground_factory

func render_world_walls(world_id: String) -> Array:
	clear_wall_nodes()
	var rules_path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/worlds.v0.json")
	var parsed = _read_json(rules_path)
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("[wall_renderer] could not read world rules for walls: %s" % rules_path)
		return []
	var worlds: Dictionary = parsed.get("worlds", {})
	var world: Dictionary = worlds.get(world_id, {})
	if str(world.get("mode", "")) == "multi_level":
		return []
	var local_walls: Array = []
	var local_index := 0
	for entity in world.get("entities", []):
		if str(entity.get("type", "")) != "wall":
			continue
		var pos: Dictionary = entity.get("position", {})
		var size: Dictionary = entity.get("size", {})
		var wall := {
			"id": "%s_wall_%03d" % [world_id, local_index],
			"position": {"x": float(pos.get("x", 0.0)), "y": float(pos.get("y", 0.0))},
			"size": {"x": float(size.get("x", 1.0)), "y": float(size.get("y", 1.0))},
			"source": "preset",
		}
		var kind := str(entity.get("kind", "wall"))
		if kind != "" and kind != "wall":
			wall["kind"] = kind
		local_walls.append(wall)
		local_index += 1
	return render_wall_layout(local_walls)

func render_wall_layout(walls: Array) -> Array:
	clear_wall_nodes()
	var current_wall_layout: Array = []
	for wall in walls:
		if typeof(wall) != TYPE_DICTIONARY:
			continue
		var normalized := normalized_wall_view(wall as Dictionary, current_wall_layout.size())
		current_wall_layout.append(normalized)
		if _walls_root != null:
			_walls_root.add_child(make_wall_node(normalized))
	_sync_dungeon_ceiling()
	return current_wall_layout

func set_level(level: int) -> void:
	_current_level = level


func clear_wall_nodes() -> void:
	_ceiling_node = null
	if _walls_root == null:
		return
	for child in _walls_root.get_children():
		_walls_root.remove_child(child)
		child.queue_free()


func _dungeon_presentation_active() -> bool:
	return _current_level < 0


func _wall_height() -> float:
	if not _dungeon_presentation_active():
		return TOWN_WALL_HEIGHT
	if _ground_factory != null and _ground_factory.has_method("dungeon_ceiling_height"):
		return maxf(TOWN_WALL_HEIGHT, float(_ground_factory.dungeon_ceiling_height()))
	return 4.0


func set_ceiling_visible(visible: bool) -> void:
	if _ceiling_node != null:
		_ceiling_node.visible = visible


func _sync_dungeon_ceiling() -> void:
	if _walls_root == null:
		return
	if not _dungeon_presentation_active():
		return
	if _ground_factory == null or not _ground_factory.has_method("make_ceiling_node"):
		return
	_ceiling_node = _ground_factory.make_ceiling_node(_current_level) as MeshInstance3D
	_walls_root.add_child(_ceiling_node)

func normalized_wall_view(wall: Dictionary, index: int) -> Dictionary:
	var pos: Dictionary = {}
	var size: Dictionary = {}
	if typeof(wall.get("position", {})) == TYPE_DICTIONARY:
		pos = wall.get("position", {})
	if typeof(wall.get("size", {})) == TYPE_DICTIONARY:
		size = wall.get("size", {})
	var out := {
		"id": str(wall.get("id", "wall_%03d" % index)),
		"position": {"x": float(pos.get("x", 0.0)), "y": float(pos.get("y", 0.0))},
		"size": {"x": float(size.get("x", 1.0)), "y": float(size.get("y", 1.0))},
	}
	if wall.has("source"):
		out["source"] = str(wall.get("source", ""))
	var kind := str(wall.get("kind", "wall"))
	if kind != "" and kind != "wall":
		out["kind"] = kind
	if wall.has("blocks_line_of_sight"):
		out["blocks_line_of_sight"] = bool(wall.get("blocks_line_of_sight", false))
	return out

func make_wall_node(wall: Dictionary) -> Node3D:
	var pos: Dictionary = wall.get("position", {})
	var size: Dictionary = wall.get("size", {})
	match str(wall.get("kind", "wall")):
		"water":
			return _make_water_node(wall)
		"hole":
			return _make_hole_node(wall)
		"rock":
			return _make_rock_node(wall)
		"column":
			return _make_column_node(wall)
		"rubble":
			return _make_rubble_node(wall)
	var wall_height := _wall_height()
	var floor_overlap := WALL_FLOOR_SEAM_OVERLAP if _dungeon_presentation_active() else 0.0
	var total_height := wall_height + floor_overlap
	var sx := float(size.get("x", 1.0))
	var sy := float(size.get("y", 1.0))
	var body := StaticBody3D.new()
	body.name = "Wall_%s" % str(wall.get("id", ""))
	body.set_meta("wall_id", str(wall.get("id", "")))
	body.set_meta("source", str(wall.get("source", "")))
	body.set_meta("kind", "wall")
	body.position = Vector3(
		float(pos.get("x", 0.0)),
		total_height * 0.5 - floor_overlap,
		float(pos.get("y", 0.0)),
	)
	var shape_node := CollisionShape3D.new()
	var box := BoxShape3D.new()
	box.size = Vector3(sx, total_height, sy)
	shape_node.shape = box
	body.add_child(shape_node)
	var node := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = Vector3(sx, total_height, sy)
	node.mesh = mesh
	var mat := _make_wall_material(wall, wall_height)
	node.material_override = mat
	body.add_child(node)
	return body

func _make_wall_material(wall: Dictionary, wall_height: float = TOWN_WALL_HEIGHT) -> StandardMaterial3D:
	var size: Dictionary = wall.get("size", {})
	var source := str(wall.get("source", ""))
	if _ground_factory != null and _ground_factory.has_method("wall_material_for_level"):
		return _ground_factory.wall_material_for_level(_current_level, source, size, wall_height)
	var mat := StandardMaterial3D.new()
	var palette: Dictionary = _ground_factory.biome_palette_for_level(_current_level) if _ground_factory != null and _ground_factory.has_method("biome_palette_for_level") else {}
	if _ground_factory != null and _ground_factory.has_method("make_wall_texture"):
		mat.albedo_texture = _ground_factory.make_wall_texture(ClientConstantsScript.WALL_TEXTURE_CAVE, palette)
	mat.texture_filter = BaseMaterial3D.TEXTURE_FILTER_NEAREST
	mat.roughness = 0.96
	mat.uv1_scale = Vector3(max(1.0, float(size.get("x", 1.0)) / 2.0), max(1.0, wall_height / 2.0), 1.0)
	match source:
		"generated":
			mat.albedo_color = Color(0.92, 0.86, 0.76)
		"perimeter":
			mat.albedo_color = Color(0.62, 0.64, 0.68)
		_:
			mat.albedo_color = Color(0.78, 0.80, 0.82)
	return mat

func _make_obstacle_root(wall: Dictionary, prefix: String, kind: String) -> Node3D:
	var pos: Dictionary = wall.get("position", {})
	var root := Node3D.new()
	root.name = "%s_%s" % [prefix, str(wall.get("id", ""))]
	root.set_meta("wall_id", str(wall.get("id", "")))
	root.set_meta("source", str(wall.get("source", "")))
	root.set_meta("kind", kind)
	root.position = Vector3(float(pos.get("x", 0.0)), 0.0, float(pos.get("y", 0.0)))
	return root

func _make_obstacle_material(wall: Dictionary, kind: String) -> StandardMaterial3D:
	var size: Dictionary = wall.get("size", {})
	if _ground_factory != null and _ground_factory.has_method("obstacle_material_for_level"):
		return _ground_factory.obstacle_material_for_level(_current_level, kind, size)
	var mat := StandardMaterial3D.new()
	var palette: Dictionary = _ground_factory.biome_palette_for_level(_current_level) if _ground_factory != null and _ground_factory.has_method("biome_palette_for_level") else {}
	if _ground_factory != null and _ground_factory.has_method("make_wall_texture"):
		mat.albedo_texture = _ground_factory.make_wall_texture(ClientConstantsScript.WALL_TEXTURE_CAVE, palette)
	mat.texture_filter = BaseMaterial3D.TEXTURE_FILTER_NEAREST
	mat.roughness = 0.98
	mat.uv1_scale = Vector3(max(1.0, float(size.get("x", 1.0)) / 2.0), max(1.0, float(size.get("y", 1.0)) / 2.0), 1.0)
	match kind:
		"rock":
			mat.albedo_color = Color(0.56, 0.58, 0.55)
		"column":
			mat.albedo_color = Color(0.70, 0.68, 0.61)
		"rubble":
			mat.albedo_color = Color(0.50, 0.47, 0.42)
		_:
			mat.albedo_color = Color(0.72, 0.72, 0.68)
	return mat

func _add_mesh_child(root: Node3D, name: String, mesh: Mesh, material: Material, local_pos: Vector3, rotation_y: float = 0.0) -> MeshInstance3D:
	var child := MeshInstance3D.new()
	child.name = name
	child.mesh = mesh
	child.material_override = material
	child.position = local_pos
	child.rotation.y = rotation_y
	root.add_child(child)
	return child

func _make_rock_node(wall: Dictionary) -> Node3D:
	var size: Dictionary = wall.get("size", {})
	var sx: float = maxf(0.5, float(size.get("x", 1.0)))
	var sz: float = maxf(0.5, float(size.get("y", 1.0)))
	var root: Node3D = _make_obstacle_root(wall, "Rock", "rock")
	var mat: StandardMaterial3D = _make_obstacle_material(wall, "rock")
	var offsets: Array[Vector3] = [
		Vector3(-sx * 0.22, 0.26, -sz * 0.14),
		Vector3(sx * 0.18, 0.22, sz * 0.08),
		Vector3(0.0, 0.34, sz * 0.20),
	]
	var scales: Array[Vector3] = [
		Vector3(maxf(0.32, sx * 0.48), 0.52, maxf(0.32, sz * 0.34)),
		Vector3(maxf(0.28, sx * 0.38), 0.44, maxf(0.28, sz * 0.30)),
		Vector3(maxf(0.24, sx * 0.30), 0.68, maxf(0.24, sz * 0.24)),
	]
	for i in offsets.size():
		var mesh: BoxMesh = BoxMesh.new()
		mesh.size = scales[i]
		_add_mesh_child(root, "RockChunk_%d" % i, mesh, mat, offsets[i], float(i) * 0.63)
	return root

func _make_column_node(wall: Dictionary) -> Node3D:
	var size: Dictionary = wall.get("size", {})
	var sx: float = maxf(0.5, float(size.get("x", 1.0)))
	var sz: float = maxf(0.5, float(size.get("y", 1.0)))
	var root: Node3D = _make_obstacle_root(wall, "Column", "column")
	var mat: StandardMaterial3D = _make_obstacle_material(wall, "column")
	var horizontal: bool = sx >= sz
	var long_extent: float = sx if horizontal else sz
	var short_extent: float = sz if horizontal else sx
	var count: int = max(1, int(floor(long_extent / 2.1)) + 1)
	var step: float = long_extent / float(count)
	var radius: float = maxf(0.18, minf(0.42, short_extent * 0.32))
	var pillar_height := maxf(1.15, _wall_height() - 0.12)
	for i in count:
		var mesh: CylinderMesh = CylinderMesh.new()
		mesh.top_radius = radius
		mesh.bottom_radius = radius
		mesh.height = pillar_height
		var along: float = -long_extent / 2.0 + step * (float(i) + 0.5)
		var local_pos: Vector3 = Vector3(along, pillar_height * 0.5, 0.0) if horizontal else Vector3(0.0, pillar_height * 0.5, along)
		_add_mesh_child(root, "ColumnPillar_%d" % i, mesh, mat, local_pos)
	return root

func _make_rubble_node(wall: Dictionary) -> Node3D:
	var size: Dictionary = wall.get("size", {})
	var sx: float = maxf(0.5, float(size.get("x", 1.0)))
	var sz: float = maxf(0.5, float(size.get("y", 1.0)))
	var root: Node3D = _make_obstacle_root(wall, "Rubble", "rubble")
	var mat: StandardMaterial3D = _make_obstacle_material(wall, "rubble")
	var offsets: Array[Vector3] = [
		Vector3(-sx * 0.26, 0.12, -sz * 0.18),
		Vector3(sx * 0.22, 0.10, -sz * 0.10),
		Vector3(-sx * 0.06, 0.16, sz * 0.08),
		Vector3(sx * 0.28, 0.11, sz * 0.22),
		Vector3(-sx * 0.30, 0.09, sz * 0.20),
	]
	for i in offsets.size():
		var mesh: BoxMesh = BoxMesh.new()
		var mesh_size := Vector3(
			maxf(0.20, sx * (0.18 + float(i % 2) * 0.06)),
			0.18 + float(i % 3) * 0.05,
			maxf(0.20, sz * (0.14 + float(i % 2) * 0.05))
		)
		mesh.size = mesh_size
		_add_mesh_child(root, "RubbleChunk_%d" % i, mesh, mat, offsets[i], float(i) * 0.48)
	return root

func _make_hole_node(wall: Dictionary) -> MeshInstance3D:
	var pos: Dictionary = wall.get("position", {})
	var size: Dictionary = wall.get("size", {})
	var node := MeshInstance3D.new()
	node.name = "Hole_%s" % str(wall.get("id", ""))
	node.set_meta("wall_id", str(wall.get("id", "")))
	node.set_meta("source", str(wall.get("source", "")))
	node.set_meta("kind", "hole")
	var mesh := PlaneMesh.new()
	mesh.size = Vector2(max(0.25, float(size.get("x", 1.0))), max(0.25, float(size.get("y", 1.0))))
	node.mesh = mesh
	node.position = Vector3(float(pos.get("x", 0.0)), 0.012, float(pos.get("y", 0.0)))
	var mat := StandardMaterial3D.new()
	var palette: Dictionary = _ground_factory.biome_palette_for_level(_current_level) if _ground_factory != null and _ground_factory.has_method("biome_palette_for_level") else {}
	if _ground_factory != null and _ground_factory.has_method("make_hole_texture"):
		mat.albedo_texture = _ground_factory.make_hole_texture(palette)
	mat.albedo_color = Color(0.74, 0.70, 0.65)
	mat.texture_filter = BaseMaterial3D.TEXTURE_FILTER_NEAREST
	mat.roughness = 1.0
	mat.uv1_scale = Vector3(max(1.0, float(size.get("x", 1.0)) / 2.5), max(1.0, float(size.get("y", 1.0)) / 2.5), 1.0)
	node.material_override = mat
	_add_surface_overlay(node, "HoleParallaxBands", Vector2(mesh.size.x, mesh.size.y), Color("#353947"), 0.20, 0.018, 4)
	return node

func _make_water_node(wall: Dictionary) -> MeshInstance3D:
	var pos: Dictionary = wall.get("position", {})
	var size: Dictionary = wall.get("size", {})
	var node := MeshInstance3D.new()
	node.name = "Water_%s" % str(wall.get("id", ""))
	node.set_meta("wall_id", str(wall.get("id", "")))
	node.set_meta("source", str(wall.get("source", "")))
	node.set_meta("kind", "water")
	var mesh := PlaneMesh.new()
	mesh.size = Vector2(max(0.25, float(size.get("x", 1.0))), max(0.25, float(size.get("y", 1.0))))
	node.mesh = mesh
	node.position = Vector3(float(pos.get("x", 0.0)), 0.018, float(pos.get("y", 0.0)))
	var mat := StandardMaterial3D.new()
	var palette: Dictionary = _ground_factory.biome_palette_for_level(_current_level) if _ground_factory != null and _ground_factory.has_method("biome_palette_for_level") else {}
	if _ground_factory != null and _ground_factory.has_method("make_water_texture"):
		mat.albedo_texture = _ground_factory.make_water_texture(palette)
	mat.albedo_color = Color(0.85, 0.96, 1.0)
	mat.texture_filter = BaseMaterial3D.TEXTURE_FILTER_NEAREST
	mat.roughness = 0.82
	mat.uv1_scale = Vector3(max(1.0, float(size.get("x", 1.0)) / 3.0), max(1.0, float(size.get("y", 1.0)) / 3.0), 1.0)
	node.material_override = mat
	_add_surface_overlay(node, "WaterMotionBands", Vector2(mesh.size.x, mesh.size.y), Color("#b8eef5"), 0.28, 0.026, 5)
	return node

func _add_surface_overlay(parent: Node3D, overlay_name: String, size: Vector2, color: Color, alpha: float, y_offset: float, band_count: int) -> void:
	var root := Node3D.new()
	root.name = overlay_name
	root.position.y = y_offset
	for i in range(band_count):
		var band := MeshInstance3D.new()
		band.name = "Band_%d" % i
		var mesh := PlaneMesh.new()
		mesh.size = Vector2(maxf(0.12, size.x * 0.76), maxf(0.04, size.y * 0.08))
		band.mesh = mesh
		band.position = Vector3(0.0, 0.0, -size.y * 0.34 + (size.y * 0.68 * float(i) / maxf(1.0, float(band_count - 1))))
		band.rotation_degrees.y = -10.0 + float(i % 3) * 10.0
		band.material_override = _surface_overlay_material(color, alpha * (1.0 - float(i) * 0.08))
		root.add_child(band)
	parent.add_child(root)

func _surface_overlay_material(color: Color, alpha: float) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(color.r, color.g, color.b, alpha)
	mat.emission_enabled = true
	mat.emission = color
	mat.emission_energy_multiplier = 0.22
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.no_depth_test = true
	return mat

func _read_json(path: String):
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return null
	var parsed = JSON.parse_string(f.get_as_text())
	return parsed
