class_name BlacksmithUpgradePreview
extends RefCounted


static func item_level(item: Dictionary) -> int:
	var rolled = item.get("rolled_stats", {})
	if typeof(rolled) == TYPE_DICTIONARY:
		var payload := rolled as Dictionary
		if typeof(payload.get("stats", {})) == TYPE_DICTIONARY:
			return int((payload.get("stats", {}) as Dictionary).get("item_level", 0))
		return int(payload.get("item_level", 0))
	return 0


static func next_cost(level: int, base_cost: int, growth_cost: int) -> int:
	return base_cost + level * growth_cost


static func pity_failure_count(item: Dictionary) -> int:
	var rolled = item.get("rolled_stats", {})
	if typeof(rolled) != TYPE_DICTIONARY:
		return 0
	var pity = (rolled as Dictionary).get("upgrade_pity", {})
	return max(0, int((pity as Dictionary).get("failures", 0))) if typeof(pity) == TYPE_DICTIONARY else 0


static func pity_guaranteed(item: Dictionary, pity_failure_threshold: int) -> bool:
	return pity_failure_threshold > 0 and pity_failure_count(item) >= pity_failure_threshold


static func preview_lines(item: Dictionary, context: Dictionary) -> Array:
	var max_level := int(context.get("max_level", 0))
	var success_chance_percent := int(context.get("success_chance_percent", 100))
	var pity_failure_threshold := int(context.get("pity_failure_threshold", 0))
	var resource_count := int(context.get("resource_count", 0))
	var wallet_gold := int(context.get("wallet_gold", 0))
	var resource_wallet_count := int(context.get("resource_wallet_count", 0))
	var resource_name := str(context.get("resource_name", "resource"))
	var base_cost := int(context.get("base_cost", 0))
	var growth_cost := int(context.get("growth_cost", 0))
	var level := item_level(item)
	var cost := next_cost(level, base_cost, growth_cost)
	var lines: Array = []
	var stats := _summary_stat_map(item)
	if stats.is_empty():
		stats = _stats_map(item)
	if level >= max_level:
		return ["Max level reached"]
	var guaranteed := pity_guaranteed(item, pity_failure_threshold)
	lines.append("Success chance: %d%%" % success_chance_percent)
	if pity_failure_threshold > 0:
		lines.append("Next upgrade guaranteed" if guaranteed else "Pity: %d/%d failures" % [pity_failure_count(item), pity_failure_threshold])
	lines.append("On success: Level %d -> %d" % [level, min(max_level, level + 1)])
	for key in _ordered_upgrade_stat_keys(stats):
		var current := int(stats.get(key, 0))
		var next := current
		if str(key) == "item_level":
			next = min(max_level, level + 1)
		elif current > 0:
			next = current + 1
		if next != current:
			lines.append("%s: %d -> %d" % [_display_stat(str(key)), current, next])
	if _failure_possible(success_chance_percent, guaranteed):
		lines.append(_failure_line(item, pity_failure_threshold))
	lines.append(_spend_line(cost, resource_count, resource_name))
	if wallet_gold >= cost and (resource_count <= 0 or resource_wallet_count >= resource_count):
		lines.append(_after_attempt_line(wallet_gold - cost, resource_wallet_count - resource_count, resource_count, resource_name))
	return lines


static func _failure_possible(success_chance_percent: int, guaranteed: bool) -> bool:
	return not guaranteed and success_chance_percent < 100


static func _failure_line(item: Dictionary, pity_failure_threshold: int) -> String:
	if pity_failure_threshold <= 0:
		return "On failure: item unchanged"
	var current := pity_failure_count(item)
	var next: int = min(pity_failure_threshold, current + 1)
	return "On failure: item unchanged; pity %d -> %d failures" % [current, next]


static func _spend_line(cost: int, resource_count: int, resource_name: String) -> String:
	if resource_count > 0:
		return "Spend on attempt: %d gold, %d %s" % [cost, resource_count, resource_name]
	return "Spend on attempt: %d gold" % cost


static func _after_attempt_line(next_gold: int, next_resource: int, resource_count: int, resource_name: String) -> String:
	if resource_count > 0:
		return "After attempt: %d gold, %d %s" % [max(0, next_gold), max(0, next_resource), resource_name]
	return "After attempt: %d gold" % max(0, next_gold)


static func _stats_map(item: Dictionary) -> Dictionary:
	var base_stats := _template_base_stats(item)
	var rolled: Variant = item.get("rolled_stats", {})
	if typeof(rolled) == TYPE_DICTIONARY:
		var payload := _dictionary_from_variant(rolled)
		if typeof(payload.get("stats", {})) == TYPE_DICTIONARY:
			var nested := _dictionary_from_variant(payload.get("stats", {}))
			_merge_missing_stats(nested, base_stats)
			return nested
		var out := payload
		var summary_stats := _summary_stat_map(item)
		for key in summary_stats.keys():
			if not out.has(key):
				out[key] = summary_stats.get(key)
		_merge_missing_stats(out, base_stats)
		return out
	var summary_only := _summary_stat_map(item)
	_merge_missing_stats(summary_only, base_stats)
	return summary_only


static func _template_base_stats(item: Dictionary) -> Dictionary:
	var def_id := str(item.get("item_def_id", ""))
	var template: Dictionary = ItemRulesLoader.item_definition(def_id)
	if typeof(template.get("base_stats", {})) != TYPE_DICTIONARY:
		return {}
	return _dictionary_from_variant(template.get("base_stats", {}))


static func _merge_missing_stats(target: Dictionary, fallback: Dictionary) -> void:
	for key in fallback.keys():
		if not target.has(key):
			target[key] = fallback.get(key)


static func _dictionary_from_variant(value: Variant) -> Dictionary:
	var parsed = JSON.parse_string(JSON.stringify(value))
	if typeof(parsed) != TYPE_DICTIONARY:
		return {}
	var out := {}
	for key in (parsed as Dictionary).keys():
		out[str(key)] = (parsed as Dictionary).get(key)
	return out


static func _summary_stat_map(item: Dictionary) -> Dictionary:
	var out := {}
	var summary: Variant = item.get("summary_lines", [])
	if typeof(summary) != TYPE_ARRAY:
		return out
	for line in summary as Array:
		var text := str(line)
		if text.begins_with("Armor"):
			out["armor"] = _last_int_in_text(text)
		elif text.begins_with("Block"):
			out["block_percent"] = _last_int_in_text(text)
		elif text.begins_with("Min damage"):
			out["damage_min"] = _last_int_in_text(text)
		elif text.begins_with("Max damage"):
			out["damage_max"] = _last_int_in_text(text)
	return out


static func _last_int_in_text(text: String) -> int:
	var regex := RegEx.new()
	regex.compile("-?\\d+")
	var matches := regex.search_all(text)
	if matches.is_empty():
		return 0
	return int(matches[matches.size() - 1].get_string())


static func _ordered_upgrade_stat_keys(stats: Dictionary) -> Array:
	var out: Array = []
	for key in ["armor", "block_percent", "damage_min", "damage_max", "str", "dex", "vit", "magic", "max_hp", "max_mana", "attack_speed_percent", "health_regen_per_10_seconds", "mana_regen_per_10_seconds", "skill_damage_percent", "hotbar_slots", "inventory_rows", "item_level"]:
		if stats.has(key):
			out.append(key)
	for key in stats.keys():
		if not out.has(str(key)):
			out.append(str(key))
	return out


static func _display_stat(stat: String) -> String:
	match stat:
		"block_percent":
			return "Block"
		"damage_min":
			return "Min damage"
		"damage_max":
			return "Max damage"
		"max_hp":
			return "Max HP"
		"max_mana":
			return "Max mana"
		"item_level":
			return "Item level"
		_:
			return stat.replace("_", " ").capitalize()
