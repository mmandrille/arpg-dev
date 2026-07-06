extends RefCounted

# Class × item equipped-fit matrix for test_item_visuals.gd.

const CharacterScene := preload("res://scenes/character.tscn")
const ResolverScript := preload("res://scripts/equipment_visuals.gd")
const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")
const ClassIdleStanceScript := preload("res://scripts/class_idle_stance.gd")
const ClassBodyTintScript := preload("res://scripts/class_body_tint.gd")
const GearSocketsLoaderScript := preload("res://scripts/gear_sockets_loader.gd")

const CLASS_IDS := ["barbarian", "paladin", "rogue", "ranger", "sorcerer"]

const CLASS_MAIN_HAND := {
	"barbarian": "barbarian_axe",
	"paladin": "long_sword",
	"rogue": "rusty_sword",
	"ranger": "ranger_shortbow",
	"sorcerer": "sorcerer_staff",
}

const SHARED_SLOTS := {
	"head": "helm",
	"chest": "mail",
	"boots": "boots",
}

const SLOT_SOCKET := {
	"head": "head_socket",
	"chest": "chest_socket",
	"boots": "boots_socket",
	"main_hand": "right_hand_socket",
	"off_hand": "off_hand_socket",
}

const SOCKET_BONE := {
	"head_socket": "head",
	"chest_socket": "chest",
	"boots_socket": "foot_l",
	"right_hand_socket": "hand_r",
	"off_hand_socket": "hand_l",
}

const MIN_BONE_REST_Y := {
	"head": 1.2,
	"chest": 0.9,
	"main_hand": 0.6,
	"off_hand": 0.6,
	"boots": -0.05,
}

const MAX_LOCAL_SCALE := {
	"head": 5,
	"chest": 5,
	"main_hand": 5,
	"off_hand": 5,
	"boots": 5,
}

const MIN_GLOBAL_SCALE := {
	"head": 0.12,
	"chest": 0.14,
	"boots": 0.06,
}


func verify_all_classes(tree: SceneTree, fail: Callable) -> bool:
	for class_id in CLASS_IDS:
		if not await _verify_class(tree, class_id, fail):
			return false
	return true


func _verify_class(tree: SceneTree, class_id: String, fail: Callable) -> bool:
	var character := CharacterScene.instantiate() as Node3D
	tree.get_root().add_child(character)
	_apply_class_model(character, class_id)
	await tree.process_frame
	await tree.process_frame
	var resolver = ResolverScript.new(character)
	resolver.set_character_class(class_id)
	var snapshot := _snapshot_for_class(class_id)
	resolver.apply_snapshot(snapshot)
	var ap := character.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		ap.play("idle")
	await tree.process_frame
	await tree.process_frame
	var state: Dictionary = resolver.get_debug_state()
	var equipped_visuals: Dictionary = state.get("equipped_visuals", {})
	var skel := character.find_child("Skeleton3D", true, false) as Skeleton3D
	if skel == null:
		fail.call("%s missing Skeleton3D after class model apply" % class_id)
		character.queue_free()
		return false
	for slot in snapshot["equipped"].keys():
		var slot_name := str(slot)
		if not equipped_visuals.has(slot_name):
			fail.call("%s missing equipped slot %s: %s" % [class_id, slot_name, equipped_visuals])
			character.queue_free()
			return false
		var mounted: Dictionary = equipped_visuals[slot_name]
		if not bool(mounted.get("visible", false)):
			fail.call("%s slot %s mounted invisible: %s" % [class_id, slot_name, mounted])
			character.queue_free()
			return false
		if bool(mounted.get("procedural_fallback", false)):
			fail.call("%s slot %s used procedural fallback unexpectedly: %s" % [class_id, slot_name, mounted])
			character.queue_free()
			return false
		var asset_id := str(mounted.get("asset_id", ""))
		var node := character.find_child(asset_id, true, false) as Node3D
		if node == null:
			fail.call("%s slot %s node %s missing" % [class_id, slot_name, asset_id])
			character.queue_free()
			return false
		var socket_name := str(SLOT_SOCKET.get(slot_name, mounted.get("mount_socket", "")))
		var socket := skel.find_child(socket_name, false, false)
		if socket == null:
			fail.call("%s slot %s socket %s missing on skeleton" % [class_id, slot_name, socket_name])
			character.queue_free()
			return false
		if node.get_parent() != socket:
			fail.call("%s slot %s parent %s != socket %s" % [
				class_id, slot_name, str(node.get_parent().name if node.get_parent() != null else ""), socket_name,
			])
			character.queue_free()
			return false
		if socket is BoneAttachment3D:
			var attachment := socket as BoneAttachment3D
			var expected_bone := str(SOCKET_BONE.get(socket_name, ""))
			if expected_bone != "" and skel.get_bone_name(attachment.bone_idx) != expected_bone:
				fail.call("%s socket %s bound to %s want %s" % [
					class_id, socket_name, skel.get_bone_name(attachment.bone_idx), expected_bone,
				])
				character.queue_free()
				return false
			var bone_y := skel.get_bone_global_rest(attachment.bone_idx).origin.y
			var min_y := float(MIN_BONE_REST_Y.get(slot_name, 0.0))
			if bone_y < min_y:
				fail.call("%s slot %s bone rest y %.3f below %.3f" % [class_id, slot_name, bone_y, min_y])
				character.queue_free()
				return false
		var local_scale := node.scale
		var max_scale := float(MAX_LOCAL_SCALE.get(slot_name, 1.5))
		if absf(local_scale.x) > max_scale or absf(local_scale.y) > max_scale or absf(local_scale.z) > max_scale:
			fail.call("%s slot %s local scale too large: %s" % [class_id, slot_name, str(local_scale)])
			character.queue_free()
			return false
		var min_global := float(MIN_GLOBAL_SCALE.get(slot_name, 0.0))
		if min_global > 0.0:
			var node_global_scale := node.global_transform.basis.get_scale()
			var global_max := maxf(
				absf(node_global_scale.x),
				maxf(absf(node_global_scale.y), absf(node_global_scale.z))
			)
			if global_max < min_global:
				fail.call("%s slot %s global scale %.3f below %.3f" % [
					class_id, slot_name, global_max, min_global,
				])
				character.queue_free()
				return false
		if slot_name == "boots":
			continue
	character.queue_free()
	return true


func _snapshot_for_class(class_id: String) -> Dictionary:
	var inventory: Array = []
	var equipped: Dictionary = {}
	var next_id := 6000
	for slot in SHARED_SLOTS.keys():
		var item_def_id := str(SHARED_SLOTS[slot])
		var iid := str(next_id)
		next_id += 1
		inventory.append({
			"item_instance_id": iid,
			"item_def_id": item_def_id,
			"slot": str(slot),
			"equipped": true,
			"rarity": "common",
		})
		equipped[str(slot)] = iid
	var main_hand := str(CLASS_MAIN_HAND.get(class_id, "rusty_sword"))
	var main_id := str(next_id)
	next_id += 1
	inventory.append({
		"item_instance_id": main_id,
		"item_def_id": main_hand,
		"slot": "main_hand",
		"equipped": true,
		"rarity": "common",
	})
	equipped["main_hand"] = main_id
	if class_id == "paladin":
		var shield_id := str(next_id)
		inventory.append({
			"item_instance_id": shield_id,
			"item_def_id": "shield",
			"slot": "off_hand",
			"equipped": true,
			"rarity": "common",
		})
		equipped["off_hand"] = shield_id
	return {"inventory": inventory, "equipped": equipped}


func _apply_class_model(character: Node3D, class_id: String) -> void:
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
	if character.has_method("refresh_gear_sockets"):
		character.call("refresh_gear_sockets")
	ClassBodyTintScript.apply_to_model(model, class_id)
