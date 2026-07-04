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

const SOCKET_LABEL_OFFSETS := {
	"head_socket":       Vector3( 0.0,   0.22,  0.0),
	"amulet_socket":     Vector3(-0.28,  0.12,  0.0),
	"chest_socket":      Vector3(-0.30,  0.08,  0.0),
	"belt_socket":       Vector3( 0.28,  0.08,  0.0),
	"gloves_socket":     Vector3( 0.30,  0.0,   0.0),
	"boots_socket":      Vector3(-0.14,  0.14,  0.0),
	"ring_left_socket":  Vector3(-0.22,  0.0,   0.0),
	"ring_right_socket": Vector3( 0.22,  0.0,   0.0),
	"right_hand_socket": Vector3( 0.14,  0.14,  0.0),
	"off_hand_socket":   Vector3(-0.14,  0.14,  0.0),
}

const SPREAD_BONES := {
	"leg_r": Vector3(0.0, 0.0, 1.0),
	"leg_l": Vector3(0.0, 0.0, -1.0),
}
const LEG_SPREAD_ANGLE := PI / 7.0
const CLASS_LEG_SPREAD_OVERRIDE := {
	"sorcerer": PI / 10.0,
}


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
	_dim_mesh(character)

	var ap_stop := character.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap_stop != null:
		ap_stop.stop()
	await capture.process_frame

	var skel := character.find_child("Skeleton3D", true, false) as Skeleton3D
	if skel != null:
		skel.reset_bone_poses()
		_spread_pose(skel, effective_class)
		await capture.process_frame

		_place_bone_dots(skel, root)
		_place_socket_spheres(character, root)
	else:
		push_warning("[skeleton] no Skeleton3D found for class %s" % effective_class)
		_place_socket_spheres(character, root)

	return character


static func _spread_pose(skel: Skeleton3D, class_id: String) -> void:
	for bone_name in SPREAD_BONES.keys():
		var idx := skel.find_bone(str(bone_name))
		if idx < 0:
			continue
		var axis: Vector3 = SPREAD_BONES[bone_name]
		var leg_angle: float = CLASS_LEG_SPREAD_OVERRIDE.get(class_id, LEG_SPREAD_ANGLE)
		var angle: float = leg_angle
		skel.set_bone_pose_rotation(idx, Quaternion(axis.normalized(), angle))
	skel.force_update_all_bone_transforms()


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
		var label_offset: Vector3 = SOCKET_LABEL_OFFSETS.get(str(socket_name), Vector3(0.0, 0.12, 0.0))
		label.global_position = socket.global_position + label_offset
	# Also mark foot_r bone position with boots color (no socket node exists for it)
	var skel_node := character.find_child("Skeleton3D", true, false) as Skeleton3D
	if skel_node != null:
		var foot_r_idx := skel_node.find_bone("foot_r")
		if foot_r_idx >= 0:
			var foot_transform: Transform3D = skel_node.global_transform * skel_node.get_bone_global_pose(foot_r_idx)
			var mat := StandardMaterial3D.new()
			mat.albedo_color = Color("#ff8822")
			mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
			var sphere := MeshInstance3D.new()
			sphere.name = "Socket_boots_r"
			var mesh := SphereMesh.new()
			mesh.radius = 0.045
			mesh.height = 0.09
			sphere.mesh = mesh
			sphere.material_override = mat
			root.add_child(sphere)
			sphere.global_position = foot_transform.origin
			var label := Label3D.new()
			label.name = "Label_boots_r"
			label.text = "boots_r"
			label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
			label.font_size = 28
			label.modulate = Color("#ff8822")
			root.add_child(label)
			label.global_position = foot_transform.origin + Vector3(0.14, 0.14, 0.0)


static func _dim_mesh(node: Node) -> void:
	if node is MeshInstance3D:
		var mi := node as MeshInstance3D
		var mat := StandardMaterial3D.new()
		mat.albedo_color = Color(1.0, 1.0, 1.0, 0.25)
		mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
		mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
		mi.material_override = mat
	for child in node.get_children():
		_dim_mesh(child)


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
	camera.size = 4.0
	root.add_child(camera)
	camera.look_at_from_position(Vector3(0.0, 1.3, 6.0), Vector3(0.0, 1.3, 0.0), Vector3.UP)
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
