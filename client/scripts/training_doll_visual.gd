class_name TrainingDollVisual
extends RefCounted

const BODY_COLOR := Color("#8a93a8")
const HEAD_COLOR := Color("#9aa3b8")


static func make_node() -> Node3D:
	var root := Node3D.new()
	root.name = "TrainingDollSilhouette"
	var body := MeshInstance3D.new()
	var body_mesh := CapsuleMesh.new()
	body_mesh.radius = 0.22
	body_mesh.height = 0.95
	body.mesh = body_mesh
	body.position = Vector3(0.0, 0.48, 0.0)
	body.material_override = _material(BODY_COLOR)
	root.add_child(body)
	var head := MeshInstance3D.new()
	var head_mesh := SphereMesh.new()
	head_mesh.radius = 0.16
	head_mesh.height = 0.32
	head.mesh = head_mesh
	head.position = Vector3(0.0, 1.02, 0.0)
	head.material_override = _material(HEAD_COLOR)
	root.add_child(head)
	return root


static func _material(color: Color) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	return mat
