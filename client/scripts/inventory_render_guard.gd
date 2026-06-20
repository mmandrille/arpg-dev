extends RefCounted

const META_KEY := "inventory_render_state_key"


static func should_render(panel: Object) -> bool:
	var key := _state_key(panel)
	var old := str(panel.get_meta(META_KEY, ""))
	if key == old:
		return false
	panel.set_meta(META_KEY, key)
	return true


static func mark_rendered(panel: Object) -> void:
	panel.set_meta(META_KEY, _state_key(panel))


static func _state_key(panel: Object) -> String:
	return JSON.stringify(_stable({
		"inventory": panel.get("inventory"),
		"equipped": panel.get("equipped"),
		"active_weapon_set": panel.get("active_weapon_set"),
		"viewed_weapon_set": panel.get("viewed_weapon_set"),
		"weapon_sets": panel.get("weapon_sets"),
		"hotbar": panel.get("hotbar"),
		"hotbar_capacity": panel.get("hotbar_capacity"),
		"inventory_rows": panel.get("inventory_rows"),
		"inventory_capacity": panel.get("inventory_capacity"),
		"gold": panel.get("gold"),
		"stash_items": panel.get("stash_items"),
		"stash_gold": panel.get("stash_gold"),
		"stash_capacity": panel.get("stash_capacity"),
		"stash_entity_id": panel.get("stash_entity_id"),
		"stash_id": panel.get("stash_id"),
		"stash_title": panel.get("stash_title"),
		"container_mode": panel.get("container_mode"),
		"container_label": panel.get("container_label"),
		"interactive": panel.get("_interactive"),
		"search_text": panel.get("_search_text"),
		"sort_mode": panel.get("_sort_mode"),
		"unique_chest_tab": panel.get("_unique_chest_tab"),
		"market_context": panel.get("_market_context"),
		"market_hidden_item_ids": panel.get("_market_hidden_item_ids"),
	}))


static func _stable(value: Variant) -> Variant:
	match typeof(value):
		TYPE_DICTIONARY:
			var source := value as Dictionary
			var keys := source.keys()
			keys.sort()
			var stable := {}
			for key in keys:
				stable[str(key)] = _stable(source.get(key))
			return stable
		TYPE_ARRAY:
			var stable := []
			for entry in value as Array:
				stable.append(_stable(entry))
			return stable
	return value
