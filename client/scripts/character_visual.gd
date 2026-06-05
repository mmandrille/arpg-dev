extends Node3D
# Attaches the weapon mount socket to the rig's hand bone (spec §5.5).
# Named "right_hand_socket" so EquipmentVisualResolver.find_child() resolves it
# unchanged. Created in code (not the .tscn) to stay robust against the exact
# imported skeleton node path.

const MOUNT_BONE := "hand_r"
const SOCKET_NAME := "right_hand_socket"


func _ready() -> void:
	var skel := find_child("Skeleton3D", true, false) as Skeleton3D
	if skel == null:
		push_warning("[character] no Skeleton3D; cannot attach %s" % SOCKET_NAME)
		return
	if skel.find_child(SOCKET_NAME, false, false) != null:
		return  # already attached (e.g. duplicated instance)
	var att := BoneAttachment3D.new()
	att.name = SOCKET_NAME
	skel.add_child(att)
	att.bone_name = MOUNT_BONE
	if att.bone_idx < 0:
		push_warning("[character] bone %s not found on skeleton" % MOUNT_BONE)
