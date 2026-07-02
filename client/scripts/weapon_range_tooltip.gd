class_name WeaponRangeTooltip
extends RefCounted

const WHOLE_TILE_EPSILON := 0.05
const WEAPON_METADATA_PREFIXES := ["Slot:", "Range:", "Reach:", "Projectile speed:"]


static func line_for_item(item: Dictionary) -> String:
	var reach := reach_for_item(item)
	if reach <= 0.0:
		return ""

	return format_range_line(reach)


static func projectile_speed_line_for_item(item: Dictionary) -> String:
	var speed := projectile_speed_for_item(item)
	if speed <= 0.0:
		return ""

	return format_projectile_speed_line(speed)


static func reach_for_item(item: Dictionary) -> float:
	return _weapon_float_stat(item, "reach")


static func projectile_speed_for_item(item: Dictionary) -> float:
	return _weapon_float_stat(item, "projectile_speed")


static func format_range_line(reach: float) -> String:
	var tiles := reach
	if _is_whole_number(tiles):
		var count := int(round(tiles))
		var unit := "tile" if count == 1 else "tiles"
		return "Range: %d %s" % [count, unit]

	return "Range: %.1f tiles" % tiles


static func format_projectile_speed_line(speed: float) -> String:
	if _is_whole_number(speed):
		return "Projectile speed: %d tiles/s" % int(round(speed))

	return "Projectile speed: %.1f tiles/s" % speed


static func ensure_after_slot(lines: Array, item: Dictionary) -> void:
	var range_line := line_for_item(item)
	var speed_line := projectile_speed_line_for_item(item)
	if range_line == "" and speed_line == "":
		return

	var insert_at := _insert_index_after_slot(lines)
	if range_line != "" and not _has_prefix(lines, "Range:") and not _has_prefix(lines, "Reach:"):
		lines.insert(insert_at, range_line)
		insert_at += 1

	if speed_line != "" and not _has_prefix(lines, "Projectile speed:"):
		var speed_insert_at := _insert_index_after_range(lines)
		if speed_insert_at < 0:
			speed_insert_at = insert_at
		lines.insert(speed_insert_at, speed_line)


static func _weapon_float_stat(item: Dictionary, key: String) -> float:
	ItemRulesLoader.ensure_loaded()
	var template_id := str(item.get("item_template_id", ""))
	if template_id != "":
		var template: Variant = ItemRulesLoader.item_templates.get(template_id, {})
		if typeof(template) == TYPE_DICTIONARY:
			var value := float((template as Dictionary).get(key, 0.0))
			if value > 0.0:
				return value

	var item_def_id := str(item.get("item_def_id", ""))
	if item_def_id != "":
		var def := ItemRulesLoader.item_definition(item_def_id)
		if def.has(key):
			return float(def.get(key, 0.0))

	return 0.0


static func _insert_index_after_slot(lines: Array) -> int:
	var slot_index := _index_of_prefix(lines, "Slot:")
	if slot_index >= 0:
		return slot_index + 1

	return lines.size()


static func _insert_index_after_range(lines: Array) -> int:
	var range_index := _index_of_prefix(lines, "Range:")
	if range_index < 0:
		range_index = _index_of_prefix(lines, "Reach:")
	if range_index >= 0:
		return range_index + 1

	return -1


static func _has_prefix(lines: Array, prefix: String) -> bool:
	return _index_of_prefix(lines, prefix) >= 0


static func _index_of_prefix(lines: Array, prefix: String) -> int:
	for i in range(lines.size()):
		if _line_text(lines[i]).begins_with(prefix):
			return i

	return -1


static func _line_text(line: Variant) -> String:
	if typeof(line) == TYPE_DICTIONARY:
		return str((line as Dictionary).get("text", ""))

	return str(line)


static func is_weapon_metadata_line(text: String) -> bool:
	for prefix in WEAPON_METADATA_PREFIXES:
		if text.begins_with(prefix):
			return true

	return false


static func _is_whole_number(value: float) -> bool:
	return abs(value - round(value)) <= WHOLE_TILE_EPSILON
