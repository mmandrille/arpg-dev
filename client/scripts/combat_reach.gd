class_name CombatReach
extends RefCounted

const APPROACH_STOP_MARGIN := 0.10


static func target_in_local_attack_range(player_anchor: Node3D, entities: Dictionary, inventory: Array, equipped: Dictionary, target_id: String, character_class: String = "") -> bool:
	if player_anchor == null or target_id == "" or not entities.has(target_id):
		return false
	var rec: Dictionary = entities[target_id]
	var target_node := rec.get("node", null) as Node3D
	if target_node == null:
		return false
	var target_position := _node_world_or_local_position(target_node)
	var player_position := _node_world_or_local_position(player_anchor)
	var flat := Vector2(target_position.x - player_position.x, target_position.z - player_position.z)
	var reach := _local_player_attack_reach(inventory, equipped, character_class)
	return flat.length() <= reach + _local_target_interaction_radius(rec) + ClientConstants.LOCAL_REACH_EPSILON


static func attack_approach_point(player_anchor: Node3D, entities: Dictionary, inventory: Array, equipped: Dictionary, target_id: String, fallback_direction: Vector2 = Vector2.RIGHT, character_class: String = "") -> Vector3:
	if player_anchor == null or target_id == "" or not entities.has(target_id):
		return Vector3.ZERO
	var rec: Dictionary = entities[target_id]
	var target_node := rec.get("node", null) as Node3D
	if target_node == null:
		return Vector3.ZERO
	var target_position := _node_world_or_local_position(target_node)
	var player_position := _node_world_or_local_position(player_anchor)
	var away_from_target := Vector2(player_position.x - target_position.x, player_position.z - target_position.z)
	if away_from_target.length_squared() <= 0.0001:
		away_from_target = fallback_direction.normalized()
	else:
		away_from_target = away_from_target.normalized()
	var stop_distance := maxf(0.0, _local_player_attack_reach(inventory, equipped, character_class) + _local_target_interaction_radius(rec) - APPROACH_STOP_MARGIN)
	return Vector3(target_position.x + away_from_target.x * stop_distance, 0.0, target_position.z + away_from_target.y * stop_distance)


static func local_player_attack_mode(inventory: Array, equipped: Dictionary) -> String:
	var item := _local_equipped_weapon_item(inventory, equipped)
	if item.is_empty():
		return "melee"
	var def := _local_equipped_weapon_definition(item)
	var mode := str(def.get("attack_mode", "melee"))
	return mode if mode != "" else "melee"


static func local_player_attack_reach(inventory: Array, equipped: Dictionary, character_class: String = "") -> float:
	return _local_player_attack_reach(inventory, equipped, character_class)


static func _local_player_attack_reach(inventory: Array, equipped: Dictionary, character_class: String = "") -> float:
	var main_reach := _slot_attack_reach(inventory, equipped, "main_hand")
	if not _has_rogue_offhand_melee_weapon(inventory, equipped, character_class):
		return main_reach
	var off_reach := _slot_attack_reach(inventory, equipped, "off_hand")
	return minf(main_reach, off_reach)


static func _slot_attack_reach(inventory: Array, equipped: Dictionary, slot: String) -> float:
	var item := _local_equipped_item_for_slot(inventory, equipped, slot)
	if item.is_empty():
		return ClientConstants.LOCAL_UNARMED_REACH if slot == "main_hand" else 0.0
	var def := _local_equipped_weapon_definition(item)
	var reach := float(def.get("reach", ClientConstants.LOCAL_UNARMED_REACH))
	return reach if reach > 0.0 else ClientConstants.LOCAL_UNARMED_REACH


static func is_rogue_dual_wield(inventory: Array, equipped: Dictionary, character_class: String) -> bool:
	return _has_rogue_offhand_melee_weapon(inventory, equipped, character_class)


static func _has_rogue_offhand_melee_weapon(inventory: Array, equipped: Dictionary, character_class: String) -> bool:
	if character_class != "rogue":
		return false
	var item := _local_equipped_item_for_slot(inventory, equipped, "off_hand")
	if item.is_empty():
		return false
	var def := _local_equipped_weapon_definition(item)
	return str(def.get("handedness", "")) == "one_handed" and str(def.get("attack_mode", "melee")) == "melee"


static func _local_equipped_weapon_item(inventory: Array, equipped: Dictionary) -> Dictionary:
	return _local_equipped_item_for_slot(inventory, equipped, "main_hand")


static func _local_equipped_item_for_slot(inventory: Array, equipped: Dictionary, slot: String) -> Dictionary:
	var raw_weapon_id = equipped.get(slot, null)
	if raw_weapon_id == null:
		return {}
	var weapon_id := str(raw_weapon_id)
	if weapon_id == "":
		return {}
	for item in inventory:
		var row: Dictionary = item
		if str(row.get("item_instance_id", "")) == weapon_id:
			return row
	return {}


static func _local_equipped_weapon_definition(item: Dictionary) -> Dictionary:
	ItemRulesLoader.ensure_loaded()
	var template_id := str(item.get("item_template_id", ""))
	if template_id != "":
		var template: Variant = ItemRulesLoader.item_templates.get(template_id, {})
		if typeof(template) == TYPE_DICTIONARY:
			return template
	var item_def_id := str(item.get("item_def_id", ""))
	if item_def_id != "":
		return ItemRulesLoader.item_definition(item_def_id)
	return {}


static func _local_target_interaction_radius(rec: Dictionary) -> float:
	match str(rec.get("type", "")):
		"monster":
			return ClientConstants.LOCAL_MONSTER_RADIUS
		"loot":
			return ClientConstants.LOCAL_LOOT_RADIUS
		"interactable":
			return ClientConstants.LOCAL_INTERACTABLE_RADIUS
		_:
			return 0.0


static func _node_world_or_local_position(node: Node3D) -> Vector3:
	if node == null:
		return Vector3.ZERO
	return node.global_position if node.is_inside_tree() else node.position
