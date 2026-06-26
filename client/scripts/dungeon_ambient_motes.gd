class_name DungeonAmbientMotes
extends RefCounted

const ROOT_NAME := "DungeonAmbientMotes"


static func sync(ground_node: MeshInstance3D, level: int, floor_size: Vector2) -> void:
	if ground_node == null:
		return
	_clear(ground_node)
	if level >= 0 or floor_size.x <= 0.0 or floor_size.y <= 0.0:
		return
	DungeonRoomPresentationLoader.ensure_loaded()
	var cfg: Dictionary = DungeonRoomPresentationLoader.ambient_motes()
	var depth := absi(level)
	var count_min := int(cfg.get("count_min", 6))
	var count_max := int(cfg.get("count_max", 18))
	var per_level := int(cfg.get("depth_per_level", 2))
	var count := clampi(count_min + depth * per_level, count_min, count_max)
	var color := Color(str(cfg.get("color", "#c8b8a0")))
	var alpha := float(cfg.get("alpha", 0.35))
	var height := float(cfg.get("height", 0.55))
	var radius := float(cfg.get("radius", 0.06))
	var root := Node3D.new()
	root.name = ROOT_NAME
	ground_node.add_child(root)
	for i in count:
		var px := _hash_unit(depth, i, 11) * floor_size.x
		var pz := _hash_unit(depth, i, 29) * floor_size.y
		root.add_child(_make_mote(Vector3(px, height, pz), color, alpha, radius))


static func _clear(ground_node: MeshInstance3D) -> void:
	if ground_node == null:
		return
	var existing := ground_node.get_node_or_null(ROOT_NAME)
	if existing != null:
		ground_node.remove_child(existing)
		existing.queue_free()


static func _hash_unit(depth: int, index: int, salt: int) -> float:
	var raw := int((depth * 131 + index * 977 + salt * 53) % 997)
	return float(raw) / 996.0


static func _make_mote(position: Vector3, color: Color, alpha: float, radius: float) -> MeshInstance3D:
	var node := MeshInstance3D.new()
	node.name = "AmbientMote"
	var mesh := SphereMesh.new()
	mesh.radius = radius
	mesh.height = radius * 2.0
	node.mesh = mesh
	node.position = position
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(color.r, color.g, color.b, alpha)
	mat.emission_enabled = true
	mat.emission = color
	mat.emission_energy_multiplier = 0.35
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	node.material_override = mat
	return node
