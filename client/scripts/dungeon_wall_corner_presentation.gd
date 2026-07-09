class_name DungeonWallCornerPresentation
extends RefCounted

const STYLE_BEVEL := "bevel_cap"
const STYLE_ROUNDED := "rounded_cap"
const STYLE_DEFAULT := STYLE_ROUNDED
const STYLE_ENV := "ARPG_DUNGEON_CORNER_STYLE"
const REVIEW_STYLES := [STYLE_BEVEL, STYLE_ROUNDED]
const ELIGIBLE_SOURCES := {
	"room_wall": true,
	"generated": true,
	"perimeter": true,
	"room_divider": true,
}
const FLOOR_OVERLAP := 0.08
const ENDPOINT_SCALE := 100.0
const BEVEL_SIZE := 1.22
const ROUNDED_RADIUS := 0.62
const HORIZONTAL_TOP_RADIUS := 0.38
const VERTICAL_TOP_RADIUS := 0.24

var _level: int = 0
var _wall_height: float = 1.0
var _ground_factory: RefCounted
var _style_override: String = ""


func _init(level: int, wall_height: float, ground_factory: RefCounted, style_override: String = "") -> void:
	_level = level
	_wall_height = wall_height
	_ground_factory = ground_factory
	_style_override = style_override


static func review_styles() -> Array:
	return REVIEW_STYLES.duplicate()


func active_style() -> String:
	if _style_override != "":
		return _validated_style(_style_override)
	var env_style := _validated_style(OS.get_environment(STYLE_ENV))
	if env_style != STYLE_BEVEL and env_style != STYLE_ROUNDED:
		return STYLE_DEFAULT
	return env_style


func detect_room_wall_corners(wall_layout: Array) -> Array:
	if _level >= 0:
		return []
	var by_point: Dictionary = {}
	for raw_wall in wall_layout:
		if typeof(raw_wall) != TYPE_DICTIONARY:
			continue
		var wall := raw_wall as Dictionary
		if not _eligible_dungeon_wall(wall):
			continue
		for endpoint in _endpoint_entries(wall):
			var key := str((endpoint as Dictionary).get("key", ""))
			var bucket: Array = by_point.get(key, [])
			bucket.append(endpoint)
			by_point[key] = bucket
	var out: Array = []
	for key in by_point.keys():
		var bucket: Array = by_point[key]
		if bucket.size() != 2:
			continue
		var first := bucket[0] as Dictionary
		var second := bucket[1] as Dictionary
		if str(first.get("orientation", "")) == str(second.get("orientation", "")):
			continue
		var horizontal := first if str(first.get("orientation", "")) == "horizontal" else second
		var vertical := first if str(first.get("orientation", "")) == "vertical" else second
		var point: Vector2 = horizontal.get("point", Vector2.ZERO)
		var h_center: Vector2 = horizontal.get("center", Vector2.ZERO)
		var v_center: Vector2 = vertical.get("center", Vector2.ZERO)
		var dir_x := _away_sign(h_center.x - point.x)
		var dir_y := _away_sign(v_center.y - point.y)
		if dir_x == 0.0 or dir_y == 0.0:
			continue
		out.append({
			"key": key,
			"point": point,
			"dir_x": dir_x,
			"dir_y": dir_y,
			"inner_face_x": float(vertical.get("thickness", 0.5)),
			"inner_face_y": float(horizontal.get("thickness", 0.5)),
			"wall_ids": [str(horizontal.get("wall_id", "")), str(vertical.get("wall_id", ""))],
		})
	return out


func detect_wall_endpoints(wall_layout: Array) -> Array:
	if _level >= 0:
		return []
	var by_point: Dictionary = {}
	for raw_wall in wall_layout:
		if typeof(raw_wall) != TYPE_DICTIONARY:
			continue
		var wall := raw_wall as Dictionary
		if not _eligible_dungeon_wall(wall):
			continue
		for endpoint in _endpoint_entries(wall):
			var key := str((endpoint as Dictionary).get("key", ""))
			if by_point.has(key):
				var existing := by_point[key] as Dictionary
				var wall_ids: Array = existing.get("wall_ids", [])
				wall_ids.append(str((endpoint as Dictionary).get("wall_id", "")))
				existing["wall_ids"] = wall_ids
				by_point[key] = existing
				continue
			by_point[key] = {
				"key": key,
				"point": (endpoint as Dictionary).get("point", Vector2.ZERO),
				"wall_ids": [str((endpoint as Dictionary).get("wall_id", ""))],
			}
	var out: Array = []
	for key in by_point.keys():
		out.append(by_point[key])
	return out


func attach_room_wall_presentation(walls_root: Node3D, wall_layout: Array) -> int:
	if walls_root == null or _level >= 0:
		return 0
	var corners := detect_wall_endpoints(wall_layout) if active_style() == STYLE_ROUNDED else detect_room_wall_corners(wall_layout)
	var room_walls := eligible_room_walls(wall_layout)
	if corners.is_empty() and room_walls.is_empty():
		return 0
	var root := Node3D.new()
	root.name = "DungeonRoomWallPresentation"
	root.set_meta("style", active_style())
	var corner_root := Node3D.new()
	corner_root.name = "CornerCaps"
	for index in range(corners.size()):
		corner_root.add_child(_make_corner_node(corners[index] as Dictionary, index))
	if corner_root.get_child_count() > 0:
		root.add_child(corner_root)
	var top_root := Node3D.new()
	top_root.name = "WallTopRounding"
	for index in range(room_walls.size()):
		top_root.add_child(_make_wall_top_node(room_walls[index] as Dictionary, index))
	if top_root.get_child_count() > 0:
		root.add_child(top_root)
	if root.get_child_count() == 0:
		root.queue_free()
		return 0
	walls_root.add_child(root)
	return corner_root.get_child_count() + top_root.get_child_count()


func eligible_room_walls(wall_layout: Array) -> Array:
	var out: Array = []
	for raw_wall in wall_layout:
		if typeof(raw_wall) != TYPE_DICTIONARY:
			continue
		var wall := raw_wall as Dictionary
		if _eligible_dungeon_wall(wall):
			out.append(wall)
	return out


func _make_corner_node(corner: Dictionary, index: int) -> Node3D:
	var root := Node3D.new()
	root.name = "RoundedCorner_%02d" % index
	root.set_meta("style", active_style())
	root.set_meta("wall_ids", corner.get("wall_ids", []))
	var point: Vector2 = corner.get("point", Vector2.ZERO)
	root.position = Vector3(point.x, 0.0, point.y)
	var child := MeshInstance3D.new()
	child.name = "RoundedCap" if active_style() == STYLE_ROUNDED else "BevelCap"
	child.material_override = _corner_material()
	if active_style() == STYLE_ROUNDED:
		var mesh := CylinderMesh.new()
		mesh.top_radius = ROUNDED_RADIUS
		mesh.bottom_radius = ROUNDED_RADIUS
		mesh.height = _total_height()
		child.mesh = mesh
		child.position = Vector3(0.0, _total_height() * 0.5 - FLOOR_OVERLAP, 0.0)
	else:
		var mesh := BoxMesh.new()
		mesh.size = Vector3(BEVEL_SIZE, _total_height(), BEVEL_SIZE)
		child.mesh = mesh
		child.position = Vector3(0.0, _total_height() * 0.5 - FLOOR_OVERLAP, 0.0)
		var dir_x := float(corner.get("dir_x", 0.0))
		var dir_y := float(corner.get("dir_y", 0.0))
		child.rotation.y = PI * 0.25 if dir_x == dir_y else -PI * 0.25
	root.add_child(child)
	return root


func _make_wall_top_node(wall: Dictionary, index: int) -> Node3D:
	var root := Node3D.new()
	root.name = "WallTop_%02d" % index
	root.set_meta("wall_id", str(wall.get("id", "")))
	var size := _size(wall)
	var orientation := _orientation(wall)
	root.position = Vector3(_center(wall).x, 0.0, _center(wall).y)
	var top := MeshInstance3D.new()
	top.name = "WallTopRounding"
	var mesh := CylinderMesh.new()
	var radius := HORIZONTAL_TOP_RADIUS if orientation == "horizontal" else VERTICAL_TOP_RADIUS
	mesh.top_radius = radius
	mesh.bottom_radius = radius
	mesh.height = size.x if orientation == "horizontal" else size.y
	top.mesh = mesh
	top.material_override = _wall_material(wall)
	var burial := 0.35 if orientation == "horizontal" else 0.60
	top.position = Vector3(0.0, _total_height() * 0.5 - radius * burial, 0.0)
	if orientation == "horizontal":
		top.rotation_degrees = Vector3(0.0, 0.0, 90.0)
	else:
		top.rotation_degrees = Vector3(90.0, 0.0, 0.0)
	root.add_child(top)
	return root


func _corner_material() -> Material:
	return _wall_material({
		"source": "generated",
		"size": {"x": 1.0, "y": 1.0},
	})


func _wall_material(wall: Dictionary) -> Material:
	var size: Dictionary = wall.get("size", {"x": 1.0, "y": 1.0})
	var source := str(wall.get("source", "room_wall"))
	if _ground_factory != null and _ground_factory.has_method("wall_material_for_level"):
		var mat: Variant = _ground_factory.wall_material_for_level(_level, source, size, _wall_height)
		if mat is StandardMaterial3D:
			return (mat as StandardMaterial3D).duplicate() as StandardMaterial3D
	var fallback := StandardMaterial3D.new()
	fallback.albedo_color = Color("#6b6255")
	fallback.roughness = 0.94
	fallback.texture_filter = BaseMaterial3D.TEXTURE_FILTER_NEAREST
	return fallback


func _eligible_dungeon_wall(wall: Dictionary) -> bool:
	var kind := str(wall.get("kind", "wall"))
	if kind != "" and kind != "wall":
		return false
	var source := str(wall.get("source", ""))
	if not ELIGIBLE_SOURCES.has(source):
		return false
	return _orientation(wall) != ""


func _endpoint_entries(wall: Dictionary) -> Array:
	var center := _center(wall)
	var size := _size(wall)
	var orientation := _orientation(wall)
	var endpoints: Array = []
	if orientation == "horizontal":
		endpoints = [
			Vector2(center.x - size.x * 0.5, center.y),
			Vector2(center.x + size.x * 0.5, center.y),
		]
	else:
		endpoints = [
			Vector2(center.x, center.y - size.y * 0.5),
			Vector2(center.x, center.y + size.y * 0.5),
		]
	var out: Array = []
	for point in endpoints:
		out.append({
			"key": _point_key(point),
			"point": point,
			"center": center,
			"orientation": orientation,
			"thickness": size.y * 0.5 if orientation == "horizontal" else size.x * 0.5,
			"wall_id": str(wall.get("id", "")),
		})
	return out


func _orientation(wall: Dictionary) -> String:
	var size := _size(wall)
	if absf(size.x - size.y) <= 0.01:
		return ""
	return "horizontal" if size.x > size.y else "vertical"


func _center(wall: Dictionary) -> Vector2:
	var pos: Dictionary = wall.get("position", {})
	return Vector2(float(pos.get("x", 0.0)), float(pos.get("y", 0.0)))


func _size(wall: Dictionary) -> Vector2:
	var size: Dictionary = wall.get("size", {})
	return Vector2(float(size.get("x", 1.0)), float(size.get("y", 1.0)))


func _point_key(point: Vector2) -> String:
	return "%d:%d" % [int(round(point.x * ENDPOINT_SCALE)), int(round(point.y * ENDPOINT_SCALE))]


func _away_sign(value: float) -> float:
	if value > 0.001:
		return 1.0
	if value < -0.001:
		return -1.0
	return 0.0


func _total_height() -> float:
	return _wall_height + FLOOR_OVERLAP


static func _validated_style(style: String) -> String:
	match style:
		STYLE_ROUNDED:
			return STYLE_ROUNDED
		STYLE_BEVEL:
			return STYLE_BEVEL
		_:
			return STYLE_DEFAULT
