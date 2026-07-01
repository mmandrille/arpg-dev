extends RefCounted

# Equipment resolver probes for test_item_visuals.gd.


const ResolverScript := preload("res://scripts/equipment_visuals.gd")


func verify_equipped_fallback_resolver(tree: SceneTree, fail: Callable) -> bool:
	var mount := _make_mount_root(tree)
	var resolver = ResolverScript.new(mount)
	var inventory := [
		{"item_instance_id": "2001", "item_def_id": "helm", "slot": "head", "equipped": true, "rarity": "rare"},
		{"item_instance_id": "2002", "item_def_id": "amulet", "slot": "amulet", "equipped": true, "rarity": "magic"},
		{"item_instance_id": "2003", "item_def_id": "mail", "slot": "chest", "equipped": true, "rarity": "common"},
		{"item_instance_id": "2004", "item_def_id": "gloves", "slot": "gloves", "equipped": true, "rarity": "magic"},
		{"item_instance_id": "2005", "item_def_id": "belt", "slot": "belt", "equipped": true, "rarity": "rare"},
		{"item_instance_id": "2006", "item_def_id": "boots", "slot": "boots", "equipped": true, "rarity": "common"},
		{"item_instance_id": "2007", "item_def_id": "ring", "slot": "ring_left", "equipped": true, "rarity": "magic"},
		{"item_instance_id": "2008", "item_def_id": "ring", "slot": "ring_right", "equipped": true, "rarity": "rare"},
		{"item_instance_id": "2009", "item_def_id": "bow", "slot": "main_hand", "equipped": true, "rarity": "rare"},
		{"item_instance_id": "2010", "item_def_id": "shield", "slot": "off_hand", "equipped": true, "rarity": "magic"},
	]
	resolver.apply_snapshot({
		"inventory": inventory,
		"equipped": {
			"head": "2001",
			"amulet": "2002",
			"chest": "2003",
			"gloves": "2004",
			"belt": "2005",
			"boots": "2006",
			"ring_left": "2007",
			"ring_right": "2008",
			"main_hand": "2009",
			"off_hand": "2010",
		},
	})
	var state: Dictionary = resolver.get_debug_state()
	if not state["warnings"].is_empty():
		fail.call("resolver emitted warnings for complete equipment fallback map: %s" % state["warnings"])
		return false
	var equipped_visuals: Dictionary = state["equipped_visuals"]
	for slot in ["head", "amulet", "chest", "gloves", "belt", "boots", "ring_left", "ring_right", "main_hand", "off_hand"]:
		if not equipped_visuals.has(slot):
			fail.call("resolver did not mount slot %s: %s" % [slot, equipped_visuals])
			return false
		var mounted: Dictionary = equipped_visuals[slot]
		if not bool(mounted.get("visible", false)):
			fail.call("resolver mounted invisible slot %s: %s" % [slot, mounted])
			return false
	for slot in ["head", "chest", "boots", "off_hand"]:
		if not bool((equipped_visuals[slot] as Dictionary).get("procedural_fallback", false)):
			fail.call("resolver did not use procedural fallback for slot %s: %s" % [slot, equipped_visuals[slot]])
			return false
	if str(equipped_visuals["ring_right"].get("mount_socket", "")) != "ring_right_socket":
		fail.call("ring_right mounted to wrong socket: %s" % equipped_visuals["ring_right"])
		return false
	if str(equipped_visuals["head"].get("tint", "")) != "ffd75e":
		fail.call("rare head tint mismatch: %s" % equipped_visuals["head"])
		return false
	var head_node := mount.find_child("fallback_equipment_head_v0", true, false) as Node3D
	if head_node == null or head_node.position.y < 0.12:
		fail.call("helmet fallback not raised above head socket: %s" % str(head_node.position if head_node != null else null))
		return false
	var chest_node := mount.find_child("fallback_equipment_chest_v0", true, false) as Node3D
	if chest_node == null or chest_node.position.z < 0.10:
		fail.call("chest fallback not pushed out from torso: %s" % str(chest_node.position if chest_node != null else null))
		return false
	var boots_node := mount.find_child("fallback_equipment_boots_v0", true, false) as Node3D
	if boots_node == null or absf(boots_node.rotation_degrees.z) > 0.001:
		fail.call("boots fallback rotated away from left/right foot layout: %s" % str(boots_node.rotation_degrees if boots_node != null else null))
		return false
	var left_boot := boots_node.find_child("left_boot", true, false) as Node3D
	var right_boot := boots_node.find_child("right_boot", true, false) as Node3D
	if left_boot == null or right_boot == null or left_boot.position.x > -0.5 or right_boot.position.x < 0.5:
		fail.call("boots fallback not split across feet: left=%s right=%s" % [
			str(left_boot.position if left_boot != null else null),
			str(right_boot.position if right_boot != null else null),
		])
		return false

	resolver.apply_snapshot({
		"inventory": [{"item_instance_id": "3001", "item_def_id": "future_helmet", "slot": "head", "equipped": true, "rarity": "magic"}],
		"equipped": {"head": "3001"},
	})
	state = resolver.get_debug_state()
	equipped_visuals = state["equipped_visuals"]
	if not equipped_visuals.has("head") or str(equipped_visuals["head"].get("asset_id", "")) != "fallback_equipment_head_v0":
		fail.call("unmapped future item did not use head fallback: %s" % equipped_visuals)
		return false
	if str(equipped_visuals["head"].get("tint", "")) != "5aa7ff":
		fail.call("unmapped future item magic tint mismatch: %s" % equipped_visuals["head"])
		return false
	mount.queue_free()

	return true


func verify_off_hand_weapon_resolver(tree: SceneTree, fail: Callable) -> bool:
	var mount := _make_mount_root(tree)
	var resolver = ResolverScript.new(mount)
	resolver.apply_snapshot({
		"inventory": [
			{"item_instance_id": "4001", "item_def_id": "starter_rogue_sword", "slot": "main_hand", "equipped": true, "rarity": "common"},
		],
		"equipped": {"off_hand": "4001"},
	})
	var state: Dictionary = resolver.get_debug_state()
	if not state["warnings"].is_empty():
		fail.call("resolver emitted warnings for rogue off-hand sword: %s" % state["warnings"])
		return false
	var equipped_visuals: Dictionary = state["equipped_visuals"]
	if not equipped_visuals.has("off_hand"):
		fail.call("rogue starter sword did not mount off hand: %s" % equipped_visuals)
		return false
	var off_hand: Dictionary = equipped_visuals["off_hand"]
	if str(off_hand.get("mount_socket", "")) != "off_hand_socket":
		fail.call("rogue starter sword off hand used wrong socket: %s" % off_hand)
		return false
	if bool(off_hand.get("procedural_fallback", false)):
		fail.call("rogue starter sword off hand used shield fallback: %s" % off_hand)
		return false
	var node := mount.find_child("weapon_rusty_sword_v0", true, false) as Node3D
	if node == null:
		fail.call("rogue starter sword off hand node missing")
		return false
	if absf(node.rotation_degrees.z - 180.0) > 0.01 or node.position.z < 0.07:
		fail.call("rogue starter sword off hand transform not mirrored: pos=%s rot=%s" % [str(node.position), str(node.rotation_degrees)])
		return false
	mount.queue_free()

	return true


func _make_mount_root(tree: SceneTree) -> Node3D:
	var mount := Node3D.new()
	mount.name = "CharacterVisual"
	for socket_name in [
		"right_hand_socket",
		"off_hand_socket",
		"head_socket",
		"amulet_socket",
		"chest_socket",
		"gloves_socket",
		"belt_socket",
		"boots_socket",
		"ring_left_socket",
		"ring_right_socket",
	]:
		var socket := Node3D.new()
		socket.name = str(socket_name)
		mount.add_child(socket)
	tree.get_root().add_child(mount)

	return mount
