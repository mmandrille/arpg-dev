class_name LootQuestBadgeShapes
extends RefCounted

## Fixed low-poly topology for the trusted quest/badge ground-shape catalog.
## The catalog owns each item's shape choice, palette, and overall scale.

const SHAPES := ["leaf", "heart", "wing", "hooded_head", "skull", "badge", "shard", "stone"]


static func supports(shape: String) -> bool:
	return shape in SHAPES


static func add_shape(root: Node3D, shape: String, color: Color, accent: Color, scale: float) -> void:
	if not supports(shape):
		return
	var group := Node3D.new()
	group.name = "QuestBadgeShape_%s" % shape
	if shape in ["hooded_head", "skull"]:
		group.rotation_degrees.y = 180.0
	root.add_child(group)
	match shape:
		"leaf":
			_leaf(group, color, accent, scale)
		"heart":
			_heart(group, color, accent, scale)
		"wing":
			_wing(group, color, accent, scale)
		"hooded_head":
			_hooded_head(group, color, accent, scale)
		"skull":
			_skull(group, color, accent, scale)
		"badge":
			_badge(group, color, accent, scale)
		"shard":
			_shard(group, color, accent, scale)
		"stone":
			_stone(group, color, accent, scale)


static func _leaf(root: Node3D, color: Color, accent: Color, scale: float) -> void:
	_sphere(root, "LeafBlade", Vector3(0.50, 0.09, 0.28), Vector3(0.0, 0.17, 0.0), color, scale)
	_box(root, "LeafVein", Vector3(0.39, 0.035, 0.035), Vector3(0.0, 0.225, 0.0), accent, scale)
	_box(root, "LeafStem", Vector3(0.20, 0.055, 0.055), Vector3(0.28, 0.17, 0.0), accent, scale)


static func _heart(root: Node3D, color: Color, accent: Color, scale: float) -> void:
	_cylinder(root, "HeartPoint", 0.015, 0.17, 0.25, Vector3(0.0, 0.205, 0.0), color, scale, 6)
	_sphere(root, "HeartLobeLeft", Vector3(0.25, 0.20, 0.23), Vector3(-0.105, 0.30, 0.0), color, scale)
	_sphere(root, "HeartLobeRight", Vector3(0.25, 0.20, 0.23), Vector3(0.105, 0.30, 0.0), color, scale)
	_box(root, "HeartHighlight", Vector3(0.08, 0.025, 0.11), Vector3(-0.11, 0.397, -0.035), accent, scale, -20.0)


static func _wing(root: Node3D, color: Color, accent: Color, scale: float) -> void:
	_sphere(root, "WingOuter", Vector3(0.19, 0.07, 0.48), Vector3(-0.17, 0.17, 0.03), color, scale, -35.0)
	_sphere(root, "WingMiddle", Vector3(0.20, 0.075, 0.53), Vector3(0.0, 0.18, -0.02), color, scale)
	_sphere(root, "WingInner", Vector3(0.19, 0.07, 0.43), Vector3(0.17, 0.17, 0.07), color, scale, 35.0)
	_box(root, "WingBone", Vector3(0.055, 0.055, 0.52), Vector3(0.0, 0.23, 0.0), accent, scale)


static func _hooded_head(root: Node3D, color: Color, accent: Color, scale: float) -> void:
	_sphere(root, "Hood", Vector3(0.43, 0.39, 0.36), Vector3(0.0, 0.25, 0.0), color, scale)
	_sphere(root, "Face", Vector3(0.26, 0.25, 0.10), Vector3(0.0, 0.235, -0.16), accent, scale)
	_box(root, "HoodBrow", Vector3(0.37, 0.07, 0.13), Vector3(0.0, 0.39, -0.14), color, scale)
	_box(root, "HoodChin", Vector3(0.27, 0.09, 0.13), Vector3(0.0, 0.085, -0.12), color, scale)
	_cylinder(root, "HoodPeak", 0.12, 0.015, 0.10, Vector3(0.0, 0.44, 0.02), color, scale, 5)


static func _skull(root: Node3D, color: Color, accent: Color, scale: float) -> void:
	_sphere(root, "Cranium", Vector3(0.37, 0.33, 0.31), Vector3(0.0, 0.265, 0.02), color, scale)
	_box(root, "Jaw", Vector3(0.23, 0.12, 0.20), Vector3(0.0, 0.10, -0.09), color, scale)
	_box(root, "EyeSocketLeft", Vector3(0.10, 0.09, 0.035), Vector3(-0.085, 0.26, -0.16), accent, scale)
	_box(root, "EyeSocketRight", Vector3(0.10, 0.09, 0.035), Vector3(0.085, 0.26, -0.16), accent, scale)
	_box(root, "NoseSocket", Vector3(0.045, 0.06, 0.03), Vector3(0.0, 0.185, -0.15), accent, scale)


static func _badge(root: Node3D, color: Color, accent: Color, scale: float) -> void:
	_cylinder(root, "Medallion", 0.25, 0.25, 0.08, Vector3(0.0, 0.15, 0.0), color, scale, 12)
	var rim := TorusMesh.new()
	rim.inner_radius = 0.17 * scale
	rim.outer_radius = 0.25 * scale
	rim.ring_segments = 12
	_mesh(root, "MedallionRim", rim, Vector3(0.0, 0.205, 0.0) * scale, accent)
	_box(root, "MedallionMarkA", Vector3(0.07, 0.025, 0.25), Vector3(0.0, 0.213, 0.0), accent, scale)
	_box(root, "MedallionMarkB", Vector3(0.25, 0.025, 0.07), Vector3(0.0, 0.213, 0.0), accent, scale)


static func _shard(root: Node3D, color: Color, accent: Color, scale: float) -> void:
	_cylinder(root, "ShardCore", 0.12, 0.015, 0.36, Vector3(0.0, 0.24, 0.0), color, scale, 5)
	_cylinder(root, "ShardLeft", 0.075, 0.01, 0.23, Vector3(-0.15, 0.17, 0.02), accent, scale, 5, -18.0)
	_cylinder(root, "ShardRight", 0.065, 0.01, 0.20, Vector3(0.15, 0.16, 0.03), color, scale, 5, 18.0)


static func _stone(root: Node3D, color: Color, accent: Color, scale: float) -> void:
	_sphere(root, "RenewStone", Vector3(0.46, 0.22, 0.38), Vector3(0.0, 0.17, 0.0), color, scale, 20.0)
	_box(root, "StoneRuneA", Vector3(0.055, 0.025, 0.22), Vector3(0.0, 0.29, 0.0), accent, scale, 20.0)
	_box(root, "StoneRuneB", Vector3(0.19, 0.025, 0.055), Vector3(0.0, 0.29, 0.0), accent, scale, 20.0)


static func _box(root: Node3D, name: String, size: Vector3, position: Vector3, color: Color, scale: float, yaw: float = 0.0) -> void:
	var mesh := BoxMesh.new()
	mesh.size = size * scale
	_mesh(root, name, mesh, position * scale, color, yaw)


static func _sphere(root: Node3D, name: String, size: Vector3, position: Vector3, color: Color, scale: float, yaw: float = 0.0) -> void:
	var mesh := SphereMesh.new()
	mesh.radius = 0.5
	mesh.height = 1.0
	mesh.radial_segments = 8
	mesh.rings = 4
	var node := _mesh(root, name, mesh, position * scale, color, yaw)
	node.scale = size * scale


static func _cylinder(root: Node3D, name: String, bottom: float, top: float, height: float, position: Vector3, color: Color, scale: float, sides: int, roll: float = 0.0) -> void:
	var mesh := CylinderMesh.new()
	mesh.bottom_radius = bottom * scale
	mesh.top_radius = top * scale
	mesh.height = height * scale
	mesh.radial_segments = sides
	var node := _mesh(root, name, mesh, position * scale, color)
	node.rotation_degrees.z = roll


static func _mesh(root: Node3D, name: String, mesh: Mesh, position: Vector3, color: Color, yaw: float = 0.0) -> MeshInstance3D:
	var node := MeshInstance3D.new()
	node.name = name
	node.mesh = mesh
	node.position = position
	node.rotation_degrees.y = yaw
	var material := StandardMaterial3D.new()
	material.albedo_color = color
	material.roughness = 0.9
	node.material_override = material
	root.add_child(node)
	return node
