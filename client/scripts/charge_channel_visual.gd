extends Node3D

const SKILL_ID := "charge"

var _active := false
var _character_visual: Node3D
var _off_hand_socket: Node3D
var _original_socket_transform := Transform3D.IDENTITY
var _has_original_socket_transform := false
var _guard: Node3D
var _trail_root: Node3D


func start(anchor: Node3D, character_visual: Node3D, direction: Vector2) -> void:
	if anchor == null or character_visual == null:
		return
	if get_parent() != anchor:
		if get_parent() != null:
			get_parent().remove_child(self)
		anchor.add_child(self)
	_character_visual = character_visual
	_active = true
	visible = true
	position = Vector3.ZERO
	_update_direction(direction)
	_pose_shield_socket()
	_ensure_guard()
	_ensure_trails()


func update_direction(direction: Vector2) -> void:
	if not _active:
		return
	_update_direction(direction)


func stop() -> void:
	if not _active:
		return
	_active = false
	visible = false
	_restore_shield_socket()


func is_active() -> bool:
	return _active


func get_debug_state() -> Dictionary:
	return {
		"active": _active,
		"has_guard": _guard != null and is_instance_valid(_guard),
		"trail_count": _trail_root.get_child_count() if _trail_root != null else 0,
		"has_shield_pose": _has_original_socket_transform,
	}


func _update_direction(direction: Vector2) -> void:
	if direction.length_squared() <= 0.0001:
		return
	var facing := direction.normalized()
	rotation.y = atan2(facing.x, facing.y)


func _pose_shield_socket() -> void:
	_off_hand_socket = _character_visual.find_child("off_hand_socket", true, false) as Node3D
	if _off_hand_socket == null:
		return
	if not _has_original_socket_transform:
		_original_socket_transform = _off_hand_socket.transform
		_has_original_socket_transform = true
	_off_hand_socket.transform = _original_socket_transform
	_off_hand_socket.position += Vector3(0.18, 0.24, 0.42)
	_off_hand_socket.rotation_degrees += Vector3(-38.0, 10.0, -28.0)


func _restore_shield_socket() -> void:
	if _off_hand_socket != null and is_instance_valid(_off_hand_socket) and _has_original_socket_transform:
		_off_hand_socket.transform = _original_socket_transform
	_has_original_socket_transform = false
	_off_hand_socket = null


func _ensure_guard() -> void:
	if _guard != null and is_instance_valid(_guard):
		return
	_guard = Node3D.new()
	_guard.name = "ChargeShieldGuard"
	add_child(_guard)

	var face := MeshInstance3D.new()
	face.name = "ShieldForwardFace"
	var mesh := CylinderMesh.new()
	mesh.top_radius = 0.48
	mesh.bottom_radius = 0.48
	mesh.height = 0.05
	mesh.radial_segments = 32
	face.mesh = mesh
	face.position = Vector3(-0.32, 1.08, 0.78)
	face.rotation_degrees = Vector3(90.0, 0.0, 0.0)
	face.material_override = _material(Color(0.52, 0.72, 1.0, 0.36))
	_guard.add_child(face)

	var rim := MeshInstance3D.new()
	rim.name = "ShieldForwardRim"
	var rim_mesh := TorusMesh.new()
	rim_mesh.inner_radius = 0.45
	rim_mesh.outer_radius = 0.51
	rim_mesh.rings = 8
	rim_mesh.ring_segments = 32
	rim.mesh = rim_mesh
	rim.position = face.position + Vector3(0.0, 0.0, 0.015)
	rim.rotation_degrees = Vector3(90.0, 0.0, 0.0)
	rim.material_override = _material(Color(0.95, 0.82, 0.34, 0.7))
	_guard.add_child(rim)


func _ensure_trails() -> void:
	if _trail_root != null and is_instance_valid(_trail_root):
		return
	_trail_root = Node3D.new()
	_trail_root.name = "ChargeSpeedTrail"
	add_child(_trail_root)
	for i in range(5):
		var streak := MeshInstance3D.new()
		streak.name = "ChargeSpeedStreak%d" % i
		var mesh := BoxMesh.new()
		mesh.size = Vector3(0.035 + float(i % 2) * 0.02, 0.035, 0.9 + float(i) * 0.13)
		streak.mesh = mesh
		streak.position = Vector3(-0.42 + float(i) * 0.21, 0.36 + float(i % 3) * 0.18, -0.74 - float(i) * 0.18)
		streak.material_override = _material(Color(0.42, 0.78, 1.0, 0.22 - float(i) * 0.018))
		_trail_root.add_child(streak)


func _material(color: Color) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	mat.no_depth_test = true
	return mat
