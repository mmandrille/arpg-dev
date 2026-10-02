class_name RarityCuePresenter
extends RefCounted

const Loader := preload("res://scripts/rarity_cue_loader.gd")
static var capture_baseline: bool = false
static var _marker_meshes: Dictionary = {}
static var _marker_material: StandardMaterial3D


static func draw_slot(canvas: Control, rect: Rect2, item: Dictionary) -> void:
	if capture_baseline:
		return
	var cue := Loader.cue_for_item(item)
	if cue.is_empty():
		return
	var badge := slot_badge_rect(rect)
	var config: Dictionary = Loader.catalog().get("slot", {})
	var background := Color(str(config.get("background", "#101010")))
	var foreground := Color(str(config.get("foreground", "#ffffff")))
	canvas.draw_rect(badge, background, true)
	var center := badge.get_center()
	var radius := badge.size.x * 0.42
	var points := _shape_points(str(cue.get("shape", "")), center, radius)
	if points.size() >= 3:
		canvas.draw_colored_polygon(points, foreground)


static func slot_badge_rect(rect: Rect2) -> Rect2:
	var config: Dictionary = Loader.catalog().get("slot", {})
	var side := maxf(float(config.get("minimum_size_px", 15)), minf(rect.size.x, rect.size.y) * float(config.get("badge_fraction", 0.28)))
	var margin := float(config.get("margin_px", 2))
	return Rect2(rect.position + Vector2(margin, margin), Vector2(side, side))


static func _shape_points(shape: String, center: Vector2, radius: float) -> PackedVector2Array:
	var points := PackedVector2Array()
	match shape:
		"square":
			points = PackedVector2Array([
				center + Vector2(-radius, -radius), center + Vector2(radius, -radius),
				center + Vector2(radius, radius), center + Vector2(-radius, radius),
			])
		"circle":
			for i in range(24):
				var angle := -PI * 0.5 + float(i) * TAU / 24.0
				points.append(center + Vector2(cos(angle), sin(angle)) * radius)
		"triangle":
			points = PackedVector2Array([center + Vector2(0, -radius), center + Vector2(radius, radius), center + Vector2(-radius, radius)])
		"diamond":
			points = PackedVector2Array([center + Vector2(0, -radius), center + Vector2(radius, 0), center + Vector2(0, radius), center + Vector2(-radius, 0)])
		"star":
			for i in range(10):
				var angle := -PI * 0.5 + float(i) * PI / 5.0
				var point_radius := radius if i % 2 == 0 else radius * 0.46
				points.append(center + Vector2(cos(angle), sin(angle)) * point_radius)
	return points


static func add_world_marker(parent: Node3D, item: Dictionary, scale: float) -> void:
	if capture_baseline:
		return
	var cue := Loader.cue_for_item(item)
	if cue.is_empty():
		return
	var config: Dictionary = Loader.catalog().get("world", {})
	var marker := MeshInstance3D.new()
	marker.name = "RarityCue"
	marker.mesh = _world_marker_mesh(str(cue.get("shape", "")))
	if marker.mesh == null:
		return
	marker.material_override = _world_marker_material()
	marker.position = Vector3(float(config.get("offset_x", 0.44)) * scale, float(config.get("height", 0.58)) * maxf(scale, 0.8), 0.0)
	marker.scale = Vector3.ONE * float(config.get("marker_size", 0.22)) * maxf(scale, 0.8)
	parent.add_child(marker)


static func _world_marker_mesh(shape: String) -> ArrayMesh:
	if _marker_meshes.has(shape):
		return _marker_meshes[shape]
	var points := _shape_points(shape, Vector2.ZERO, 0.5)
	if points.size() < 3:
		return null
	var vertices := PackedVector3Array()
	var colors := PackedColorArray()
	for pass_index in 2:
		var factor := 1.0 if pass_index == 0 else 0.78
		var color := Color(str(Loader.catalog().get("world", {}).get("outline", "#000000"))) if pass_index == 0 else Color(str(Loader.catalog().get("world", {}).get("foreground", "#ffffff")))
		for i in range(points.size()):
			vertices.append(Vector3.ZERO)
			vertices.append(Vector3(points[i].x * factor, points[i].y * factor, 0.0))
			var next := points[(i + 1) % points.size()]
			vertices.append(Vector3(next.x * factor, next.y * factor, 0.0))
			colors.append(color)
			colors.append(color)
			colors.append(color)
	var arrays := []
	arrays.resize(Mesh.ARRAY_MAX)
	arrays[Mesh.ARRAY_VERTEX] = vertices
	arrays[Mesh.ARRAY_COLOR] = colors
	var mesh := ArrayMesh.new()
	mesh.add_surface_from_arrays(Mesh.PRIMITIVE_TRIANGLES, arrays)
	_marker_meshes[shape] = mesh
	return mesh


static func _world_marker_material() -> StandardMaterial3D:
	if _marker_material == null:
		_marker_material = StandardMaterial3D.new()
		_marker_material.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
		_marker_material.vertex_color_use_as_albedo = true
		_marker_material.billboard_mode = BaseMaterial3D.BILLBOARD_ENABLED
		_marker_material.no_depth_test = true
		_marker_material.depth_draw_mode = BaseMaterial3D.DEPTH_DRAW_DISABLED
		_marker_material.cull_mode = BaseMaterial3D.CULL_DISABLED
	return _marker_material
