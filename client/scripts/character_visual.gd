extends Node3D
# Attaches equipment sockets in code (not the .tscn) to stay robust against the
# exact imported skeleton node path. Sockets ride skeleton bones via BoneAttachment3D
# with data-driven offsets from shared/assets/gear_sockets.v0.json.

const GearSocketsLoaderScript := preload("res://scripts/gear_sockets_loader.gd")

const STATIC_SOCKET_POSITIONS := {
	"right_hand_socket": Vector3(0.36, 0.92, 0.0),
	"off_hand_socket": Vector3(-0.36, 0.92, 0.0),
	"head_socket": Vector3(0.0, 1.55, 0.0),
	"chest_socket": Vector3(0.0, 1.08, 0.0),
	"gloves_socket": Vector3(0.0, 0.82, 0.0),
	"belt_socket": Vector3(0.0, 0.78, 0.0),
	"boots_socket": Vector3(0.0, 0.22, 0.0),
	"ring_left_socket": Vector3(-0.42, 0.82, 0.02),
	"ring_right_socket": Vector3(0.42, 0.82, 0.02),
	"amulet_socket": Vector3(0.0, 1.32, -0.06),
}

var class_id: String = ""


func _ready() -> void:
	_ensure_gear_sockets()


func refresh_gear_sockets() -> void:
	_remove_gear_socket_nodes()
	_ensure_gear_sockets()


func _remove_gear_socket_nodes() -> void:
	var skel := find_child("Skeleton3D", true, false) as Skeleton3D
	var socket_names: Array = GearSocketsLoaderScript.sockets_for_class(class_id).keys()
	if skel != null:
		for child in skel.get_children():
			if socket_names.has(child.name):
				child.free()
		return
	for socket_name in socket_names:
		var existing := find_child(str(socket_name), true, false)
		if existing != null:
			existing.free()


func _ensure_gear_sockets() -> void:
	var skel := find_child("Skeleton3D", true, false) as Skeleton3D
	if skel == null:
		_ensure_static_sockets()
		return
	var sockets := GearSocketsLoaderScript.sockets_for_class(class_id)
	for socket_name in sockets.keys():
		if skel.find_child(str(socket_name), false, false) != null:
			continue
		var entry: Dictionary = sockets[socket_name]
		if typeof(entry) != TYPE_DICTIONARY:
			continue
		var bone_name := str(entry.get("bone", ""))
		if bone_name == "":
			continue
		var bone_idx := skel.find_bone(bone_name)
		if bone_idx < 0:
			push_warning("[character] bone %s not found for socket %s" % [bone_name, socket_name])
			continue
		var att := BoneAttachment3D.new()
		att.name = str(socket_name)
		skel.add_child(att)
		att.bone_idx = bone_idx
		if att.bone_idx < 0:
			att.queue_free()
			continue
		_apply_socket_transform(att, entry)


func _ensure_static_sockets() -> void:
	for socket_name in STATIC_SOCKET_POSITIONS.keys():
		if find_child(str(socket_name), true, false) != null:
			continue
		var socket := Node3D.new()
		socket.name = str(socket_name)
		socket.position = STATIC_SOCKET_POSITIONS[socket_name]
		add_child(socket)


func _apply_socket_transform(att: BoneAttachment3D, entry: Dictionary) -> void:
	var offset: Dictionary = entry.get("offset", {}) if typeof(entry.get("offset", {})) == TYPE_DICTIONARY else {}
	att.position = Vector3(
		float(offset.get("x", 0.0)),
		float(offset.get("y", 0.0)),
		float(offset.get("z", 0.0)),
	)
	var rotation: Dictionary = entry.get("rotation_degrees", {}) if typeof(entry.get("rotation_degrees", {})) == TYPE_DICTIONARY else {}
	att.rotation_degrees = Vector3(
		float(rotation.get("x", 0.0)),
		float(rotation.get("y", 0.0)),
		float(rotation.get("z", 0.0)),
	)


func _ensure_weapon_socket() -> void:
	refresh_gear_sockets()
