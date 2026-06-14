class_name ChestPresentation
extends RefCounted

const OBJECTIVE_MARKER_NAME := "EliteObjectiveMarker"


static func add_part(parent: Node3D, part_name: String, size: Vector3, position: Vector3, color: Color) -> MeshInstance3D:
	var part := MeshInstance3D.new()
	part.name = part_name
	var mesh := BoxMesh.new()
	mesh.size = size
	part.mesh = mesh
	part.position = position
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	part.material_override = mat
	parent.add_child(part)
	return part


static func sync_objective_marker(root: Node3D, enabled: bool, opened: bool) -> void:
	if root == null:
		return
	var marker := root.find_child(OBJECTIVE_MARKER_NAME, true, false) as MeshInstance3D
	if not enabled:
		if marker != null:
			marker.visible = false
		return
	if marker == null:
		marker = add_part(root, OBJECTIVE_MARKER_NAME, Vector3(1.38, 0.035, 1.02), Vector3(0.0, 0.035, 0.0), Color("#6ee68b"))
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color("#f4d481") if opened else Color("#6ee68b")
	mat.emission_enabled = true
	mat.emission = Color("#7d5b20") if opened else Color("#1f6f44")
	marker.material_override = mat
	marker.visible = true


static func has_objective_marker(root: Node3D) -> bool:
	if root == null:
		return false
	var marker := root.find_child(OBJECTIVE_MARKER_NAME, true, false) as MeshInstance3D
	return marker != null and marker.visible
