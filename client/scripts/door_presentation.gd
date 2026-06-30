## Code-native wooden door mesh + open burst for dungeon interactables.
class_name DoorPresentation
extends RefCounted

const InteractableRulesLoaderScript := preload("res://scripts/interactable_rules_loader.gd")

const PANEL_HEIGHT := 2.85
const OPEN_BURST_NAME := "DoorOpenBurst"


static func make_door_node() -> Node3D:
	var barrier := InteractableRulesLoaderScript.barrier_size("wooden_door")
	var width := float(barrier.get("x", 1.6))
	var depth := maxf(float(barrier.get("y", 0.25)), 0.25)
	var root := Node3D.new()
	root.name = "InteractableDoor"

	root.add_child(_part("DoorFrameLeft", Vector3(0.12, PANEL_HEIGHT, depth + 0.08), Vector3(-width * 0.5 - 0.04, PANEL_HEIGHT * 0.5, 0.0), Color("#3a2418")))
	root.add_child(_part("DoorFrameRight", Vector3(0.12, PANEL_HEIGHT, depth + 0.08), Vector3(width * 0.5 + 0.04, PANEL_HEIGHT * 0.5, 0.0), Color("#3a2418")))
	root.add_child(_part("DoorFrameTop", Vector3(width + 0.24, 0.14, depth + 0.10), Vector3(0.0, PANEL_HEIGHT + 0.02, 0.0), Color("#2f1c12")))

	var pivot := Node3D.new()
	pivot.name = "DoorPivot"
	pivot.position = Vector3(-width * 0.5, 0.0, 0.0)
	root.add_child(pivot)

	var panel := _part("DoorPanel", Vector3(width * 0.92, PANEL_HEIGHT * 0.94, depth), Vector3(width * 0.5, PANEL_HEIGHT * 0.5, 0.0), Color("#6b4226"))
	pivot.add_child(panel)

	for i in range(3):
		var plank_x := width * (0.22 + float(i) * 0.28)
		pivot.add_child(_part("DoorPlank%d" % i, Vector3(0.06, PANEL_HEIGHT * 0.82, depth * 0.55), Vector3(plank_x, PANEL_HEIGHT * 0.5, depth * 0.18), Color("#523018")))

	pivot.add_child(_part("DoorBand", Vector3(width * 0.88, 0.10, depth + 0.04), Vector3(width * 0.5, PANEL_HEIGHT * 0.58, 0.0), Color("#8a8d90")))
	pivot.add_child(_part("DoorHandle", Vector3(0.10, 0.22, 0.14), Vector3(width * 0.78, PANEL_HEIGHT * 0.48, depth * 0.55), Color("#caa85a")))

	return root


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
	mesh.inner_radius = 0.55
	mesh.outer_radius = 0.82
	burst.mesh = mesh
	burst.position = Vector3(0.0, PANEL_HEIGHT * 0.55, 0.0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.72, 0.58, 0.28, 0.38)
	mat.emission_enabled = true
	mat.emission = Color("#d4a843")
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	burst.material_override = mat
	root.add_child(burst)
	if root.get_tree() != null:
		var timer := root.get_tree().create_timer(0.45)
		timer.timeout.connect(func() -> void:
			if is_instance_valid(burst):
				burst.queue_free()
		)


static func _part(part_name: String, size: Vector3, position: Vector3, color: Color) -> MeshInstance3D:
	var part := MeshInstance3D.new()
	part.name = part_name
	var mesh := BoxMesh.new()
	mesh.size = size
	part.mesh = mesh
	part.position = position
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	part.material_override = mat
	return part
