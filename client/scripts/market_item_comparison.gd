class_name MarketItemComparison
extends RefCounted

const StatLabels := preload("res://scripts/stat_labels.gd")
const STAT_ORDER := ["damage_min", "damage_max", "str", "dex", "vit", "magic", "all_skills", "armor", "block_percent", "attack_speed_percent", "hit_chance", "crit_chance", "evade_chance", "max_hp", "max_mana", "health_regen_per_10_seconds", "mana_regen_per_10_seconds", "skill_damage_percent", "skill_cooldown_reduction_percent", "skill_mana_cost_reduction", "magic_find_percent", "hotbar_slots", "inventory_rows"]


static func comparison_for_item(item: Dictionary, inventory: Array, equipped: Dictionary) -> Dictionary:
	ItemRulesLoader.ensure_loaded()
	var stats := stats_for_item(item)
	var slot := _comparison_slot(_item_slot(item), equipped)
	if slot == "" or stats.is_empty():
		return {}
	var equipped_item := _equipped_item(slot, inventory, equipped)
	var equipped_stats := stats_for_item(equipped_item)
	var deltas := _comparison_deltas(stats, equipped_stats)
	if deltas.is_empty():
		return {}
	return {
		"slot": slot,
		"equipped_item_instance_id": str(equipped_item.get("item_instance_id", "")),
		"deltas": deltas,
	}


static func entries_for_item(item: Dictionary, inventory: Array, equipped: Dictionary) -> Array:
	return entries_for_comparison(comparison_for_item(item, inventory, equipped))


static func entries_for_comparison(comparison: Dictionary) -> Array:
	var entries: Array = []
	for line in text_lines_for_comparison(comparison):
		var delta := _line_delta(line)
		entries.append({"text": line, "color": _comparison_color(delta)})
	return entries


static func text_lines_for_comparison(comparison: Dictionary) -> Array:
	var lines: Array = []
	var deltas: Array = comparison.get("deltas", [])
	for delta in deltas:
		if typeof(delta) != TYPE_DICTIONARY:
			continue
		var rec := delta as Dictionary
		var diff := int(rec.get("delta", 0))
		var sign := "+" if diff >= 0 else ""
		lines.append("%s%d %s vs equipped" % [sign, diff, StatLabels.display_name(str(rec.get("stat", "")))])
	return lines


static func add_row_labels(parent: VBoxContainer, item: Dictionary, inventory: Array, equipped: Dictionary, font_size: int) -> void:
	for entry in entries_for_item(item, inventory, equipped):
		var label := Label.new()
		label.text = str((entry as Dictionary).get("text", ""))
		label.add_theme_font_size_override("font_size", font_size)
		label.add_theme_color_override("font_color", (entry as Dictionary).get("color", Color("#d8c7a6")))
		parent.add_child(label)


static func enrich_debug_listing_rows(rows: Array, source: Array, inventory: Array, equipped: Dictionary) -> Array:
	var listings: Array = []
	for value in source:
		if typeof(value) == TYPE_DICTIONARY:
			listings.append(value as Dictionary)
	for i in range(rows.size()):
		if typeof(rows[i]) != TYPE_DICTIONARY:
			continue
		var listing: Dictionary = listings[i] if i < listings.size() else {}
		var comparison := comparison_for_item(listing, inventory, equipped)
		var lines := text_lines_for_comparison(comparison)
		var row := rows[i] as Dictionary
		row["comparison_count"] = lines.size()
		row["comparison_lines"] = lines
		row["comparison_visible"] = not lines.is_empty()
	return rows


static func stats_for_item(item: Dictionary) -> Dictionary:
	if item.is_empty():
		return {}
	var stats: Dictionary = _base_stats_for_item(item)
	var rolled = item.get("rolled_stats", {})
	if typeof(rolled) == TYPE_DICTIONARY:
		for key in (rolled as Dictionary).keys():
			var parsed = _numeric_stat_or_null((rolled as Dictionary).get(key, null))
			if parsed != null:
				stats[str(key)] = int(parsed)
	return stats


static func _base_stats_for_item(item: Dictionary) -> Dictionary:
	var def := _item_definition(item)
	var stats: Dictionary = {}
	var base = def.get("base_stats", {})
	if typeof(base) == TYPE_DICTIONARY:
		for key in (base as Dictionary).keys():
			var parsed = _numeric_stat_or_null((base as Dictionary).get(key, null))
			if parsed != null:
				stats[str(key)] = int(parsed)
	var damage = def.get("damage", {})
	if typeof(damage) == TYPE_DICTIONARY:
		stats["damage_min"] = int((damage as Dictionary).get("min", 0))
		stats["damage_max"] = int((damage as Dictionary).get("max", stats.get("damage_min", 0)))
	return stats


static func _comparison_deltas(offered: Dictionary, equipped_stats: Dictionary) -> Array:
	var deltas: Array = []
	var seen := {}
	for stat in STAT_ORDER:
		if offered.has(stat) or equipped_stats.has(stat):
			_append_delta(deltas, stat, offered, equipped_stats)
			seen[stat] = true
	var extras: Array = []
	for stat in offered.keys():
		if not seen.has(str(stat)):
			extras.append(str(stat))
	for stat in equipped_stats.keys():
		if not seen.has(str(stat)) and not extras.has(str(stat)):
			extras.append(str(stat))
	extras.sort()
	for stat in extras:
		_append_delta(deltas, stat, offered, equipped_stats)
	return deltas


static func _append_delta(deltas: Array, stat: String, offered: Dictionary, equipped_stats: Dictionary) -> void:
	var offered_value := int(offered.get(stat, 0))
	var equipped_value := int(equipped_stats.get(stat, 0))
	if offered_value == 0 and equipped_value == 0:
		return
	deltas.append({"stat": stat, "offered": offered_value, "equipped": equipped_value, "delta": offered_value - equipped_value})


static func _equipped_item(slot: String, inventory: Array, equipped: Dictionary) -> Dictionary:
	var item_id = equipped.get(slot, null)
	if item_id == null or str(item_id) == "":
		return {}
	for item in inventory:
		if typeof(item) == TYPE_DICTIONARY and str((item as Dictionary).get("item_instance_id", "")) == str(item_id):
			return item as Dictionary
	return {}


static func _comparison_slot(slot: String, equipped: Dictionary) -> String:
	if slot == "ring":
		if str(equipped.get("ring_left", "")) != "":
			return "ring_left"
		if str(equipped.get("ring_right", "")) != "":
			return "ring_right"
		return "ring_left"
	return slot


static func _item_slot(item: Dictionary) -> String:
	var slot := str(item.get("slot", ""))
	if slot != "":
		return slot
	return str(_item_definition(item).get("slot", ""))


static func _item_definition(item: Dictionary) -> Dictionary:
	var template_id := str(item.get("item_template_id", item.get("item_def_id", "")))
	if template_id != "" and ItemRulesLoader.item_templates.has(template_id):
		return ItemRulesLoader.item_templates.get(template_id, {})
	return ItemRulesLoader.item_definition(str(item.get("item_def_id", "")))


static func _numeric_stat_or_null(value: Variant):
	match typeof(value):
		TYPE_INT:
			return int(value)
		TYPE_FLOAT:
			return int(value)
		TYPE_STRING:
			if str(value).is_valid_int():
				return int(value)
			if str(value).is_valid_float():
				return int(float(value))
	return null


static func _line_delta(line: String) -> int:
	var first_space := line.find(" ")
	if first_space <= 1:
		return 0
	return int(line.substr(0, first_space))


static func _comparison_color(delta: int) -> Color:
	if delta > 0:
		return Color("#9ee6a8")
	if delta < 0:
		return Color("#ff9f7a")
	return Color("#d8c7a6")
