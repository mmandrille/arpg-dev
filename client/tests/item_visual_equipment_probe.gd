extends RefCounted

# Equipment resolver probes for test_item_visuals.gd.


const ResolverScript := preload("res://scripts/equipment_visuals.gd")


func verify_full_loadout_resolver(tree: SceneTree, fail: Callable) -> bool:
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
		{"item_instance_id": "2009", "item_def_id": "long_sword", "slot": "main_hand", "equipped": true, "rarity": "rare"},
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
		fail.call("resolver emitted warnings for a complete loadout: %s" % state["warnings"])
		return false
	var equipped_visuals: Dictionary = state["equipped_visuals"]
	# ADR-0018 P3b: only weapons mount meshes; armor recolours the hero, jewelry has no world visual.
	var expected_kind := {
		"head": "headgear", "chest": "tint", "gloves": "tint", "belt": "tint", "boots": "tint",
		"amulet": "none", "ring_left": "none", "ring_right": "none", "main_hand": "mesh", "off_hand": "mesh",
	}
	for slot in expected_kind.keys():
		var mounted: Dictionary = equipped_visuals.get(slot, {})
		if str(mounted.get("kind", "")) != str(expected_kind[slot]):
			fail.call("slot %s kind %s, want %s: %s" % [slot, mounted.get("kind", ""), expected_kind[slot], mounted])
			return false
	for slot in ["main_hand", "off_hand"]:
		if not bool((equipped_visuals[slot] as Dictionary).get("visible", false)):
			fail.call("weapon slot %s mounted invisible: %s" % [slot, equipped_visuals[slot]])
			return false
	for socket_name in ["head_socket", "amulet_socket", "chest_socket", "gloves_socket", "belt_socket", "boots_socket", "ring_left_socket", "ring_right_socket"]:
		var socket := mount.find_child(socket_name, false, false)
		if socket != null and socket.get_child_count() > 0:
			fail.call("armor socket %s must stay empty, has %s" % [socket_name, socket.get_children()])
			return false

	resolver.apply_snapshot({
		"inventory": [{"item_instance_id": "3001", "item_def_id": "future_helmet", "slot": "head", "equipped": true, "rarity": "magic"}],
		"equipped": {"head": "3001"},
	})
	state = resolver.get_debug_state()
	if not state["warnings"].is_empty():
		fail.call("an uncoloured future head item must degrade quietly: %s" % state["warnings"])
		return false
	var future_head: Dictionary = (state["equipped_visuals"] as Dictionary).get("head", {})
	if str(future_head.get("kind", "")) != "headgear" or bool((state["armor_look"] as Dictionary).get("regions", {}).has("headgear")):
		fail.call("uncoloured head item should show headgear without a tint: %s / %s" % [future_head, state["armor_look"]])
		return false
	mount.queue_free()

	return true


func verify_off_hand_weapon_resolver(tree: SceneTree, fail: Callable) -> bool:
	var mount := _make_mount_root(tree)
	var resolver = ResolverScript.new(mount)
	resolver.apply_snapshot({
		"inventory": [
			{"item_instance_id": "4001", "item_def_id": "rusty_sword", "slot": "main_hand", "equipped": true, "rarity": "common"},
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
	var node := mount.find_child(resolver.asset_id_for("rusty_sword"), true, false) as Node3D
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
