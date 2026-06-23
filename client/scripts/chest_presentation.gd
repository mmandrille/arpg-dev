class_name ChestPresentation
extends RefCounted

const OBJECTIVE_MARKER_NAME := "EliteObjectiveMarker"
const QUEST_MARKER_NAME := "QuestRewardMarker"


const OPEN_BURST_NAME := "ChestOpenBurst"


static func sync_open_burst(root: Node3D, opened: bool) -> void:
	if root == null:
		return
	var burst := root.find_child(OPEN_BURST_NAME, true, false) as MeshInstance3D
	if not opened:
		if burst != null:
			burst.queue_free()
		return
	if burst != null:
		return
	burst = MeshInstance3D.new()
	burst.name = OPEN_BURST_NAME
	var mesh := TorusMesh.new()
	mesh.inner_radius = 0.42
	mesh.outer_radius = 0.62
	burst.mesh = mesh
	burst.position = Vector3(0.0, 0.72, 0.0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(1.0, 0.86, 0.34, 0.42)
	mat.emission_enabled = true
	mat.emission = Color("#f0cf72")
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	burst.material_override = mat
	root.add_child(burst)
	if root.get_tree() != null:
		var timer := root.get_tree().create_timer(0.42)
		timer.timeout.connect(func() -> void:
			if is_instance_valid(burst):
				burst.queue_free()
		)


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


static func sync_quest_marker(root: Node3D, enabled: bool, opened: bool) -> void:
	if root == null:
		return
	var marker := root.find_child(QUEST_MARKER_NAME, true, false) as MeshInstance3D
	if not enabled:
		if marker != null:
			marker.visible = false
		return
	if marker == null:
		marker = add_part(root, QUEST_MARKER_NAME, Vector3(0.58, 0.08, 0.58), Vector3(0.0, 0.98, 0.0), Color("#79b8ff"))
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color("#f7e27a") if opened else Color("#79b8ff")
	mat.emission_enabled = true
	mat.emission = Color("#8c6f18") if opened else Color("#164a8c")
	marker.material_override = mat
	marker.visible = true


static func has_objective_marker(root: Node3D) -> bool:
	if root == null:
		return false
	var marker := root.find_child(OBJECTIVE_MARKER_NAME, true, false) as MeshInstance3D
	return marker != null and marker.visible


static func has_quest_marker(root: Node3D) -> bool:
	if root == null:
		return false
	var marker := root.find_child(QUEST_MARKER_NAME, true, false) as MeshInstance3D
	return marker != null and marker.visible
