class_name CombatReach
extends RefCounted


static func target_in_local_attack_range(player_anchor: Node3D, entities: Dictionary, inventory: Array, equipped: Dictionary, target_id: String) -> bool:
	if player_anchor == null or target_id == "" or not entities.has(target_id):
		return false
	var rec: Dictionary = entities[target_id]
	var target_node := rec.get("node", null) as Node3D
	if target_node == null:
		return false
	var target_position := _node_world_or_local_position(target_node)
	var player_position := _node_world_or_local_position(player_anchor)
	var flat := Vector2(target_position.x - player_position.x, target_position.z - player_position.z)
	var reach := _local_player_attack_reach(inventory, equipped)
	return flat.length() <= reach + _local_target_interaction_radius(rec) + ClientConstants.LOCAL_REACH_EPSILON


static func _local_player_attack_reach(inventory: Array, equipped: Dictionary) -> float:
	var item := _local_equipped_weapon_item(inventory, equipped)
	if item.is_empty():
		return ClientConstants.LOCAL_UNARMED_REACH
	var def := _local_equipped_weapon_definition(item)
	var reach := float(def.get("reach", ClientConstants.LOCAL_UNARMED_REACH))
	return reach if reach > 0.0 else ClientConstants.LOCAL_UNARMED_REACH


static func _local_equipped_weapon_item(inventory: Array, equipped: Dictionary) -> Dictionary:
	var raw_weapon_id = equipped.get("main_hand", null)
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
