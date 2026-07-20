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
var _hand_proxies: Dictionary = {}
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
	_hand_proxies = {}
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
	out["hands_visible"] = _visible_hand_proxy_count() > 0
	out["hand_proxy_count"] = _visible_hand_proxy_count()
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
	_ensure_hand_proxies()


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


func _ensure_hand_proxies() -> void:
	if _rig == null:
		return
	if not bool(_cfg.get("first_person_hands_visible", true)):
		return
	_add_hand_proxy("right_hand_socket", "FirstPersonRightHand", 1.0)
	_add_hand_proxy("off_hand_socket", "FirstPersonLeftHand", -1.0)


func _add_hand_proxy(socket_name: String, proxy_name: String, side: float) -> void:
	var socket := _rig.find_child(socket_name, true, false) as Node3D
	if socket == null or _hand_proxies.has(socket_name):
		return
	var root := Node3D.new()
	root.name = proxy_name
	socket.add_child(root)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.86, 0.68, 0.46, 1.0)
	mat.roughness = 0.72
	var forearm := MeshInstance3D.new()
	forearm.name = "Forearm"
	var forearm_mesh := CapsuleMesh.new()
	forearm_mesh.radius = 0.055
	forearm_mesh.height = 0.48
	forearm.mesh = forearm_mesh
	forearm.material_override = mat
	forearm.cast_shadow = GeometryInstance3D.SHADOW_CASTING_SETTING_OFF
	forearm.position = Vector3(-0.05 * side, -0.24, 0.08)
	forearm.rotation_degrees = Vector3(10.0, 0.0, 12.0 * side)
	root.add_child(forearm)
	var hand := MeshInstance3D.new()
	hand.name = "Hand"
	var hand_mesh := SphereMesh.new()
	hand_mesh.radius = 0.075
	hand_mesh.height = 0.12
	hand.mesh = hand_mesh
	hand.material_override = mat
	hand.cast_shadow = GeometryInstance3D.SHADOW_CASTING_SETTING_OFF
	hand.position = Vector3(0.0, -0.03, 0.03)
	root.add_child(hand)
	_hand_proxies[socket_name] = root


func _visible_hand_proxy_count() -> int:
	var count := 0
	for proxy in _hand_proxies.values():
		if proxy != null and is_instance_valid(proxy) and bool((proxy as Node3D).visible):
			count += 1
	return count


func _update_root_state() -> void:
	_state["rig"] = {
		"active": _rig != null and is_instance_valid(_rig),
		"parent": _rig.get_parent().name if _rig != null and _rig.get_parent() != null else "",
		"right_hand_socket": _rig.find_child("right_hand_socket", true, false) != null if _rig != null else false,
		"off_hand_socket": _rig.find_child("off_hand_socket", true, false) != null if _rig != null else false,
		"hands_visible": _visible_hand_proxy_count() > 0,
		"hand_proxy_count": _visible_hand_proxy_count(),
	}


func _vec3(value) -> Vector3:
	if value is Array and (value as Array).size() >= 3:
		return Vector3(float(value[0]), float(value[1]), float(value[2]))
	return Vector3.ZERO
