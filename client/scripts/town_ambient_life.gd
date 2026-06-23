class_name TownAmbientLife
extends RefCounted

const PROP_ROOT_NAME := "TownAmbientLife"


static func attach_to_town(root: Node3D) -> void:
	if root == null or root.find_child(PROP_ROOT_NAME, false, false) != null:
		return
	var props := Node3D.new()
	props.name = PROP_ROOT_NAME
	root.add_child(props)
	_add_silhouette(props, Vector3(10.5, 0.0, 12.2), Color("#6d5a48"))
	_add_silhouette(props, Vector3(13.8, 0.0, 14.6), Color("#4f6678"))
	_add_silhouette(props, Vector3(8.2, 0.0, 15.1), Color("#5f6b4a"))


static func _add_silhouette(parent: Node3D, position: Vector3, color: Color) -> void:
	var body := MeshInstance3D.new()
	body.name = "AmbientSilhouette"
	var mesh := CapsuleMesh.new()
	mesh.radius = 0.18
	mesh.height = 0.72
	body.mesh = mesh
	body.position = position + Vector3(0.0, 0.36, 0.0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	body.material_override = mat
	parent.add_child(body)
