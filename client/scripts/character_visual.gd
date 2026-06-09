extends Node3D
# Attaches equipment sockets in code (not the .tscn) to stay robust against the
# exact imported skeleton node path. The weapon socket rides the hand bone; the
# rest are lightweight root-relative placeholders for fallback gear visuals.

const MOUNT_BONE := "hand_r"
const SOCKET_NAME := "right_hand_socket"
const FALLBACK_SOCKETS := {
	"off_hand_socket": Vector3(-0.38, 0.92, 0.02),
	"head_socket": Vector3(0.0, 1.55, 0.0),
	"chest_socket": Vector3(0.0, 1.08, 0.0),
	"gloves_socket": Vector3(0.0, 0.82, 0.0),
	"belt_socket": Vector3(0.0, 0.78, 0.0),
	"boots_socket": Vector3(0.0, 0.22, 0.0),
	"ring_left_socket": Vector3(-0.42, 0.82, 0.02),
	"ring_right_socket": Vector3(0.42, 0.82, 0.02),
	"amulet_socket": Vector3(0.0, 1.32, -0.06),
}


func _ready() -> void:
	_ensure_weapon_socket()
	_ensure_fallback_sockets()


func _ensure_weapon_socket() -> void:
	var skel := find_child("Skeleton3D", true, false) as Skeleton3D
	if skel == null:
		push_warning("[character] no Skeleton3D; cannot attach %s" % SOCKET_NAME)
		return
	if skel.find_child(SOCKET_NAME, false, false) != null:
		return
	var att := BoneAttachment3D.new()
	att.name = SOCKET_NAME
	skel.add_child(att)
	att.bone_name = MOUNT_BONE
	if att.bone_idx < 0:
		push_warning("[character] bone %s not found on skeleton" % MOUNT_BONE)


func _ensure_fallback_sockets() -> void:
	for socket_name in FALLBACK_SOCKETS.keys():
		if find_child(str(socket_name), true, false) != null:
			continue
		var socket := Node3D.new()
		socket.name = str(socket_name)
		socket.position = FALLBACK_SOCKETS[socket_name]
		add_child(socket)
