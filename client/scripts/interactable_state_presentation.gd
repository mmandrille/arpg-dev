## Interactable open-state tints and door/chest motion.
class_name InteractableStatePresentation
extends RefCounted

const ChestPresentationScript := preload("res://scripts/chest_presentation.gd")
const DoorPresentationScript := preload("res://scripts/door_presentation.gd")
const TownNodeFactoryScript := preload("res://scripts/town_node_factory.gd")


static func apply_state_tint(rec: Dictionary, state: String) -> void:
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	var def_id := str(rec.get("interactable_def_id", ""))
	if def_id == "treasure_chest" or def_id == "town_stash" or def_id == "town_unique_chest":
		ChestPresentationScript.sync_objective_marker(node, bool(rec.get("elite_objective", false)), state == "open")
		ChestPresentationScript.sync_quest_marker(node, bool(rec.get("quest_reward", false)), state == "open")
		if state == "open":
			ChestPresentationScript.sync_open_burst(node, true)
		var glow := node.find_child("ChestInnerGlow", true, false) as MeshInstance3D
		if glow != null:
			glow.visible = state == "open"
		var lock := node.find_child("ChestLockPlate", true, false) as MeshInstance3D
		if lock != null:
			var lock_mat := StandardMaterial3D.new()
			if state == "locked" or state == "disabled":
				lock_mat.albedo_color = Color("#7a2f2d")
			elif state == "open":
				lock_mat.albedo_color = Color("#f0cf72")
				lock_mat.emission_enabled = true
				lock_mat.emission = Color("#8a6122")
			else:
				lock_mat.albedo_color = Color("#c77dff") if def_id == "town_unique_chest" else (Color("#d1b15d") if def_id == "town_stash" else Color("#8d8f8f"))
			lock.material_override = lock_mat
		return
	if def_id == "teleporter":
		var core := node.get_child(1) as MeshInstance3D if node.get_child_count() > 1 else null
		if core == null:
			return
		var mat := StandardMaterial3D.new()
		if state == "disabled" or state == "locked":
			mat.albedo_color = Color(0.30, 0.16, 0.18)
			mat.emission_enabled = false
		else:
			mat.albedo_color = Color(0.15, 0.62, 0.70)
			mat.emission_enabled = true
			mat.emission = Color(0.05, 0.55, 0.68)
		core.material_override = mat
		return
	if def_id == "stairs_down" or def_id == "stairs_up":
		var base := node.get_child(0) as MeshInstance3D if node.get_child_count() > 0 else null
		if base == null:
			return
		var mat := StandardMaterial3D.new()
		mat.albedo_color = TownNodeFactoryScript.stair_base_color(def_id, state)
		base.material_override = mat


static func animate_open_motion(node: Node3D, state: String, host: Node) -> bool:
	var chest_pivot := node.find_child("ChestLidPivot", true, false) as Node3D
	if chest_pivot != null:
		var chest_rot := deg_to_rad(-68.0) if state == "open" else 0.0
		var chest_tween := host.create_tween()
		chest_tween.tween_property(chest_pivot, "rotation:x", chest_rot, 0.22)
		return true
	var pivot := node.find_child("DoorPivot", true, false) as Node3D
	if pivot == null:
		return false
	if state == "open":
		DoorPresentationScript.sync_open_burst(node, true)
	var target_rot := deg_to_rad(90.0) if state == "open" else 0.0
	var tween := host.create_tween()
	tween.tween_property(pivot, "rotation:y", target_rot, 0.25)
	return true
