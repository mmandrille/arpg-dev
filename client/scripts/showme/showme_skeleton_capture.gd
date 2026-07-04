class_name ShowmeSkeletonCapture
extends RefCounted

const CharacterScene := preload("res://scenes/character.tscn")
const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")
const ClassIdleStanceScript := preload("res://scripts/class_idle_stance.gd")

const SOCKET_COLORS := {
	"right_hand_socket": Color("#4488ff"),
	"off_hand_socket": Color("#4488ff"),
	"head_socket": Color("#ffee44"),
	"chest_socket": Color("#44dddd"),
	"belt_socket": Color("#44dddd"),
	"amulet_socket": Color("#44dddd"),
	"boots_socket": Color("#ff8822"),
	"gloves_socket": Color("#ff8822"),
	"ring_left_socket": Color("#cc44ff"),
	"ring_right_socket": Color("#cc44ff"),
}

const SPREAD_BONES := {
	"arm_r": Vector3(0.0, 0.0, -1.0),
	"arm_l": Vector3(0.0, 0.0, 1.0),
	"leg_r": Vector3(0.0, 0.0, -1.0),
	"leg_l": Vector3(0.0, 0.0, 1.0),
}
const ARM_SPREAD_ANGLE := PI / 2.0
const LEG_SPREAD_ANGLE := PI / 5.0


static func setup(capture: SceneTree, class_id: String) -> Node3D:
	var root := Node3D.new()
	root.name = "VisualFeedbackSkeleton"
	capture.get_root().add_child(root)

	_add_light(root)
	_add_camera(root)
	_add_floor(root)

	var character := CharacterScene.instantiate() as Node3D
	character.name = "FocusedCharacter"
	root.add_child(character)

	var effective_class := class_id if class_id != "" else "paladin"
	_apply_class_model(character, effective_class)
	await capture.process_frame
	await capture.process_frame

	var skel := character.find_child("Skeleton3D", true, false) as Skeleton3D
	if skel != null:
		_spread_pose(skel)
		await capture.process_frame

		_place_bone_dots(skel, root)
		_place_socket_spheres(character, root)
	else:
		push_warning("[skeleton] no Skeleton3D found for class %s" % effective_class)
		_place_socket_spheres(character, root)

	return character


static func _spread_pose(skel: Skeleton3D) -> void:
	for bone_name in SPREAD_BONES.keys():
		var idx := skel.find_bone(str(bone_name))
		if idx < 0:
			continue
		var axis: Vector3 = SPREAD_BONES[bone_name]
		var angle := ARM_SPREAD_ANGLE if str(bone_name).begins_with("arm") else LEG_SPREAD_ANGLE
		skel.set_bone_pose_rotation(idx, Quaternion(axis.normalized(), angle))


static func _place_bone_dots(skel: Skeleton3D, root: Node3D) -> void:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color("#e03030")
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	for i in range(skel.get_bone_count()):
		var bone_transform: Transform3D = skel.global_transform * skel.get_bone_global_pose(i)
		var dot := MeshInstance3D.new()
		dot.name = "BoneDot_%s" % skel.get_bone_name(i)
		var mesh := SphereMesh.new()
		mesh.radius = 0.025
		mesh.height = 0.05
		dot.mesh = mesh
		dot.material_override = mat
		root.add_child(dot)
		dot.global_position = bone_transform.origin


static func _place_socket_spheres(character: Node3D, root: Node3D) -> void:
	for socket_name in SOCKET_COLORS.keys():
		var socket := character.find_child(str(socket_name), true, false) as Node3D
		if socket == null:
			continue
		var color: Color = SOCKET_COLORS[socket_name]
		var mat := StandardMaterial3D.new()
		mat.albedo_color = color
		mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
		var sphere := MeshInstance3D.new()
		sphere.name = "Socket_%s" % socket_name
		var mesh := SphereMesh.new()
		mesh.radius = 0.045
		mesh.height = 0.09
		sphere.mesh = mesh
		sphere.material_override = mat
		root.add_child(sphere)
		sphere.global_position = socket.global_position
		var label := Label3D.new()
		label.name = "Label_%s" % socket_name
		label.text = str(socket_name).replace("_socket", "")
		label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
		label.font_size = 28
		label.modulate = color
		root.add_child(label)
		label.global_position = socket.global_position + Vector3(0.0, 0.12, 0.0)


static func _apply_class_model(character: Node3D, class_id: String) -> void:
	var resolved := ClassPresentationsLoaderScript.resolve(class_id)
	var packed := ClassPresentationsLoaderScript.packed_scene_for_class(class_id)
	if packed == null:
		return
	var old_model := character.find_child("ModelRoot", false, false) as Node
	if old_model != null:
		character.remove_child(old_model)
		old_model.free()
	var model := packed.instantiate() as Node3D
	model.name = "ModelRoot"
	model.scale = Vector3.ONE * float(resolved.get("scale", 1.0))
	model.position.y = float(resolved.get("height_offset", 0.0))
	ClassIdleStanceScript.apply_to_model(model, class_id)
	character.add_child(model)
	character.move_child(model, 0)
	var ap := character.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		ap.root_node = NodePath("../ModelRoot")
	if "class_id" in character:
		character.set("class_id", class_id)
	if character.has_method("_ensure_weapon_socket"):
		character.call("_ensure_weapon_socket")


static func _add_light(root: Node3D) -> void:
	var light := DirectionalLight3D.new()
	light.name = "key_light"
	light.light_energy = 2.2
	light.rotation_degrees = Vector3(-55, -35, 0)
	root.add_child(light)


static func _add_camera(root: Node3D) -> void:
	var camera := Camera3D.new()
	camera.name = "capture_camera"
	camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	camera.size = 3.2
	root.add_child(camera)
	camera.look_at_from_position(Vector3(3.2, 2.2, 4.5), Vector3(0.0, 1.1, 0.0), Vector3.UP)
	camera.current = true


static func _add_floor(root: Node3D) -> void:
	var floor := MeshInstance3D.new()
	floor.name = "reference_floor"
	var mesh := BoxMesh.new()
	mesh.size = Vector3(4.0, 0.04, 4.0)
	floor.mesh = mesh
	floor.position = Vector3(0, -0.03, 0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color("#3f3f3c")
	floor.material_override = mat
	root.add_child(floor)
