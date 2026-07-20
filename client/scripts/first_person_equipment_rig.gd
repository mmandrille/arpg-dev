class_name FirstPersonEquipmentRig
extends RefCounted

const AnimationControllerScript := preload("res://scripts/animation_controller.gd")
const CharacterScene := preload("res://scenes/character.tscn")

var _camera: Camera3D
var _enabled := false
var _cfg: Dictionary = {}
var _rig: Node3D
var _animation: AnimationController
var _mounted: Dictionary = {}
var _state: Dictionary = {}
var _attack_count := 0


func set_eye_view(camera: Camera3D, enabled: bool, cfg: Dictionary) -> void:
	_camera = camera
	_enabled = enabled and camera != null and is_instance_valid(camera)
	_cfg = cfg.duplicate(true)
	if not _enabled:
		clear()
		return
	_ensure_rig()
	_apply_rig_transform()
	_update_root_state()


func clear() -> void:
	for slot in _mounted.keys():
		clear_slot(str(slot))
	if _rig != null and is_instance_valid(_rig):
		_rig.queue_free()
	_rig = null
	_animation = null
	_state = {}


func clear_slot(slot: String) -> void:
	var mounted = _mounted.get(slot, null)
	if mounted != null and is_instance_valid(mounted):
		(mounted as Node3D).queue_free()
	_mounted.erase(slot)
	_state[slot] = {"active": false, "visible": false, "slot": slot}


func mount_slot(slot: String, node: Node3D, metadata: Dictionary) -> bool:
	if not _enabled or node == null:
		clear_slot(slot)
		return false
	_ensure_rig()
	var socket_name := str(metadata.get("mount_socket", "right_hand_socket"))
	var socket := _rig.find_child(socket_name, true, false) as Node3D
	if socket == null:
		clear_slot(slot)
		_state[slot]["warning"] = "missing_socket"
		_state[slot]["mount_socket"] = socket_name
		return false
	clear_slot(slot)
	socket.add_child(node)
	_mounted[slot] = node
	_state[slot] = metadata.duplicate(true)
	_state[slot]["active"] = true
	_state[slot]["visible"] = node.visible
	_state[slot]["attack_count"] = _attack_count
	_state[slot]["socket_parent"] = socket.name
	_state[slot]["node_parent"] = node.get_parent().name if node.get_parent() != null else ""
	_state[slot]["node_path"] = str(node.get_path()) if node.is_inside_tree() else ""
	return true


func animation_controller():
	if not _enabled:
		return null
	_ensure_rig()
	return _animation


func record_attack(slot: String, clip: String) -> void:
	if not _enabled:
		return
	_attack_count += 1
	_state["last_attack_clip"] = clip
	for key in _mounted.keys():
		var slot_state: Dictionary = _state.get(str(key), {}).duplicate(true)
		slot_state["attack_count"] = _attack_count
		if str(key) == slot:
			slot_state["last_attack_clip"] = clip
		_state[str(key)] = slot_state


func get_debug_state() -> Dictionary:
	var out := _state.duplicate(true)
	out["rig_active"] = _rig != null and is_instance_valid(_rig)
	out["enabled"] = _enabled
	out["attack_count"] = _attack_count
	out["animation_clip"] = _animation.current_clip() if _animation != null else ""
	out["rig_path"] = str(_rig.get_path()) if _rig != null and _rig.is_inside_tree() else ""
	out["parent"] = _rig.get_parent().name if _rig != null and _rig.get_parent() != null else ""
	return out


func _ensure_rig() -> void:
	if _rig != null and is_instance_valid(_rig):
		return
	if _camera == null or not is_instance_valid(_camera):
		return
	_rig = CharacterScene.instantiate() as Node3D
	_rig.name = "FirstPersonEquipmentRig"
	_camera.add_child(_rig)
	if _rig.has_method("refresh_gear_sockets"):
		_rig.call("refresh_gear_sockets")
	var ap := _rig.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		_animation = AnimationControllerScript.new(ap)
	_configure_body_geometry()


func _apply_rig_transform() -> void:
	if _rig == null:
		return
	_rig.position = _vec3(_cfg.get("first_person_rig_position", [0.0, -1.2, -0.42]))
	_rig.rotation_degrees = _vec3(_cfg.get("first_person_rig_rotation_degrees", [0.0, 0.0, 0.0]))
	var scale := float(_cfg.get("first_person_rig_scale", 0.78))
	_rig.scale = Vector3.ONE * scale


func _configure_body_geometry() -> void:
	if _rig == null:
		return
	var body_visible := bool(_cfg.get("first_person_body_visible", true))
	for node in _rig.find_children("*", "GeometryInstance3D", true, false):
		var geom := node as GeometryInstance3D
		geom.cast_shadow = GeometryInstance3D.SHADOW_CASTING_SETTING_OFF
		geom.visible = body_visible


func _update_root_state() -> void:
	_state["rig"] = {
		"active": _rig != null and is_instance_valid(_rig),
		"parent": _rig.get_parent().name if _rig != null and _rig.get_parent() != null else "",
		"right_hand_socket": _rig.find_child("right_hand_socket", true, false) != null if _rig != null else false,
		"off_hand_socket": _rig.find_child("off_hand_socket", true, false) != null if _rig != null else false,
	}


func _vec3(value) -> Vector3:
	if value is Array and (value as Array).size() >= 3:
		return Vector3(float(value[0]), float(value[1]), float(value[2]))
	return Vector3.ZERO
