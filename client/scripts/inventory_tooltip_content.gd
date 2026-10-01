class_name InventoryTooltipContent
extends RefCounted
## Pure item-tooltip content builder (v515). Extracted from InventoryPanel so it
## can be imported and tested without the panel or the scene tree. Display only:
## it reads item dictionaries and static rules and never mutates item state.

const StatLabels := preload("res://scripts/stat_labels.gd")
const ItemRequirementViews := preload("res://scripts/item_requirement_views.gd")
const ItemTooltipStatSectionsScript := preload("res://scripts/item_tooltip_stat_sections.gd")
const UniqueEffectTooltipScript := preload("res://scripts/unique_effect_tooltip.gd")
const WeaponRangeTooltipScript := preload("res://scripts/weapon_range_tooltip.gd")
const HOTKEY_LABELS := ["1", "2", "3", "4", "5", "6", "7", "8", "9", "0"]
const TOOLTIP_META_FONT_SIZE := 19


## Narrow typed input: the only panel state tooltip content depends on.
class Context:
	extends RefCounted

	var hotbar: Array = []

	func _init(next_hotbar: Array = []) -> void:
		hotbar = next_hotbar


static func tooltip_text(item: Dictionary, context: Context) -> String:
	return "\n".join(tooltip_text_lines(tooltip_lines(item, context)) + requirement_lines_as_summary(item) + comparison_text_lines(item) + UniqueEffectTooltipScript.text_lines_for_item(item))


static func tooltip_lines(item: Dictionary, context: Context) -> Array:
	var def: Dictionary = item_definition_for_item(item)
	var def_id := str(item.get("item_def_id", ""))
	var rarity := str(item.get("rarity", ""))
	var lines: Array = [item_name_tooltip_line(str(item.get("display_name", def.get("name", def_id))), rarity)]
	if rarity != "":
		lines.append(metadata_tooltip_line("Rarity: %s" % rarity.capitalize()))
	var summary_lines := detail_lines(item, false, false)
	if not summary_lines.is_empty():
		var metadata_lines := ItemTooltipStatSectionsScript.metadata_lines_from_summary(summary_lines)
		WeaponRangeTooltipScript.ensure_after_slot(metadata_lines, item)
		lines.append_array(compact_metadata_lines(metadata_lines))
		ItemTooltipStatSectionsScript.append_equipment_stat_sections(lines, item.get("rolled_stats", {}), def, ItemTooltipStatSectionsScript.TOOLTIP_STAT_SEPARATOR)
		append_hotbar_tooltip_line(lines, item, context.hotbar)
		return lines
	var slot := str(def.get("slot", ""))
	if slot != "":
		lines.append(metadata_tooltip_line("Slot: %s" % slot))
	else:
		var category := str(def.get("category", ""))
		if category != "":
			lines.append("Kind: %s" % category)
	var range_line := WeaponRangeTooltipScript.line_for_item(item)
	if range_line != "":
		lines.append(metadata_tooltip_line(range_line))
	var projectile_speed_line := WeaponRangeTooltipScript.projectile_speed_line_for_item(item)
	if projectile_speed_line != "":
		lines.append(metadata_tooltip_line(projectile_speed_line))
	if def.has("attack_mode"):
		lines.append("Mode: %s" % str(def["attack_mode"]))
	lines.append_array(consumable_effect_lines(def))
	var base_stat_lines := ItemTooltipStatSectionsScript.base_stat_lines_for(def)
	if not base_stat_lines.is_empty():
		lines.append(ItemTooltipStatSectionsScript.TOOLTIP_STAT_SEPARATOR)
		lines.append_array(base_stat_lines)
	var random_stat_lines := ItemTooltipStatSectionsScript.random_stat_lines_for(item.get("rolled_stats", {}), def)
	if not random_stat_lines.is_empty():
		lines.append(ItemTooltipStatSectionsScript.TOOLTIP_STAT_SEPARATOR)
		lines.append_array(random_stat_lines)
	elif def.has("damage"):
		var dmg: Dictionary = def["damage"]
		lines.append("Damage: %s-%s" % [str(dmg.get("min", "?")), str(dmg.get("max", "?"))])
	append_hotbar_tooltip_line(lines, item, context.hotbar)
	return lines


static func item_name_tooltip_line(text: String, rarity: String) -> Dictionary:
	return {"text": text, "color": rarity_color(rarity)}


static func metadata_tooltip_line(text: String) -> Dictionary:
	return {"text": text, "color": Color("#cdbd9f"), "font_size": TOOLTIP_META_FONT_SIZE}


static func compact_metadata_lines(lines: Array) -> Array:
	var out: Array = []
	for line in lines:
		var text := str(line)
		if WeaponRangeTooltipScript.is_weapon_metadata_line(text):
			out.append(metadata_tooltip_line(text))
		elif text.begins_with("Set:"):
			out.append_array(set_membership_tooltip_lines(text))
		elif text.find("set bonus:") >= 0:
			out.append_array(set_bonus_tooltip_lines(text))
		else:
			out.append(line)
	return out


static func set_membership_tooltip_lines(text: String) -> Array:
	var split_at := text.rfind(" (")
	if split_at < 0:
		return [{"text": text, "color": Color("#55e66f"), "font_size": TOOLTIP_META_FONT_SIZE}]
	return [
		{"text": text.substr(0, split_at), "color": Color("#55e66f"), "font_size": TOOLTIP_META_FONT_SIZE},
		{"text": text.substr(split_at + 1), "color": Color("#55e66f"), "font_size": TOOLTIP_META_FONT_SIZE},
	]


static func set_bonus_tooltip_lines(text: String) -> Array:
	var color := Color("#7df095") if text.find("(active)") >= 0 else Color("#7d8d7f")
	var split_at := text.find(": ")
	if split_at < 0:
		return [{"text": text, "color": color, "font_size": TOOLTIP_META_FONT_SIZE}]
	return [
		{"text": text.substr(0, split_at + 1), "color": color, "font_size": TOOLTIP_META_FONT_SIZE},
		{"text": text.substr(split_at + 2), "color": color, "font_size": TOOLTIP_META_FONT_SIZE},
	]


static func tooltip_text_lines(lines: Array) -> Array:
	var out: Array = []
	for line in lines:
		if typeof(line) == TYPE_DICTIONARY:
			out.append(str((line as Dictionary).get("text", "")))
		else:
			out.append(str(line))
	return out


static func rarity_color(rarity: String) -> Color:
	return InventoryPanelStyles.rarity_color(rarity)


static func append_hotbar_tooltip_line(lines: Array, item: Dictionary, hotbar: Array) -> void:
	var hotbar_labels := hotbar_labels_for_item(hotbar, str(item.get("item_instance_id", "")))
	if not hotbar_labels.is_empty():
		lines.append("Assigned to hotbar: %s" % ", ".join(hotbar_labels))


static func hotbar_slots_for_item(hotbar: Array, item_instance_id: String) -> Array:
	var slots: Array = []
	if item_instance_id == "":
		return slots
	for slot in hotbar:
		if typeof(slot) != TYPE_DICTIONARY:
			continue
		var rec := slot as Dictionary
		var assigned_id = rec.get("item_instance_id", null)
		if assigned_id != null and str(assigned_id) == item_instance_id:
			var slot_index := int(rec.get("slot_index", -1))
			if slot_index >= 0 and slot_index < HOTKEY_LABELS.size():
				slots.append(slot_index)
	return slots


static func hotbar_labels_for_item(hotbar: Array, item_instance_id: String) -> Array:
	var labels: Array = []
	for slot_index in hotbar_slots_for_item(hotbar, item_instance_id):
		var index := int(slot_index)
		if index >= 0 and index < HOTKEY_LABELS.size():
			labels.append(hotbar_label_for_slot(index))
	return labels


static func hotbar_label_for_slot(slot_index: int) -> String:
	if slot_index >= 0 and slot_index < HOTKEY_LABELS.size():
		return HOTKEY_LABELS[slot_index]
	return str(slot_index + 1)


static func consumable_effect_lines(def: Dictionary) -> Array:
	var lines: Array = []
	var heal = def.get("heal", {})
	if typeof(heal) == TYPE_DICTIONARY and not (heal as Dictionary).is_empty():
		lines.append("Restores %s HP" % range_text(heal as Dictionary))
	var mana_restore = def.get("mana_restore", {})
	if typeof(mana_restore) == TYPE_DICTIONARY and not (mana_restore as Dictionary).is_empty():
		lines.append("Restores %s mana" % range_text(mana_restore as Dictionary))
	return lines


static func range_text(value: Dictionary) -> String:
	var min_value := int(value.get("min", 0))
	var max_value := int(value.get("max", min_value))
	if min_value == max_value:
		return str(min_value)
	return "%d-%d" % [min_value, max_value]


static func detail_lines(item: Dictionary, include_requirements: bool = true, include_comparison: bool = true) -> Array:
	var lines: Array = []
	var summary = item.get("summary_lines", [])
	if typeof(summary) != TYPE_ARRAY:
		return lines
	for line in summary:
		var text := str(line)
		if text == "":
			continue
		if not include_requirements and is_requirement_summary_line(text):
			continue
		if not include_comparison and is_comparison_summary_line(text):
			continue
		lines.append(text)
	return lines


static func requirement_lines(item: Dictionary) -> Array:
	var lines: Array = []
	var statuses = item.get("requirement_status", [])
	if typeof(statuses) == TYPE_ARRAY:
		for status in statuses:
			if typeof(status) != TYPE_DICTIONARY:
				continue
			var formatted := ItemRequirementViews.format_requirement_status(status as Dictionary, requirement_color(true), requirement_color(false))
			if not formatted.is_empty():
				lines.append(formatted)
	if not lines.is_empty():
		return lines
	var requirements: Dictionary = item.get("requirements", {})
	if requirements.has("level"):
		lines.append("Level %s" % str(requirements["level"]))
	for key in requirements.keys():
		var stat := str(key)
		if stat == "level":
			continue
		lines.append("%s %s" % [display_stat(stat), str(requirements.get(key, ""))])
	var summary = item.get("summary_lines", [])
	if typeof(summary) == TYPE_ARRAY:
		for line in summary:
			var parsed := requirement_from_summary_line(str(line))
			if parsed != "" and not lines.has(parsed):
				lines.append(parsed)
	return lines


static func requirement_lines_as_summary(item: Dictionary) -> Array:
	var lines: Array = []
	for line in requirement_lines(item):
		var text := entry_text(line)
		if text.to_lower().begins_with("level "):
			lines.append("Requires %s" % text.to_lower())
		else:
			lines.append("Requires %s" % text)
	return lines


static func is_requirement_summary_line(text: String) -> bool:
	return requirement_from_summary_line(text) != ""


static func is_comparison_summary_line(text: String) -> bool:
	return comparison_delta_from_line(text) != null


static func requirement_from_summary_line(text: String) -> String:
	var normalized := text.strip_edges()
	if not normalized.to_lower().begins_with("requires "):
		return ""
	var rest := normalized.substr("Requires ".length()).strip_edges()
	if rest.to_lower().begins_with("level "):
		return "Level %s" % rest.substr("level ".length()).strip_edges()
	return rest.capitalize()


static func comparison_entries(item: Dictionary) -> Array:
	var entries: Array = []
	append_equip_preview_entries(entries, item)
	var comparison = item.get("comparison", {})
	if typeof(comparison) == TYPE_DICTIONARY:
		var deltas = (comparison as Dictionary).get("deltas", [])
		if typeof(deltas) == TYPE_ARRAY:
			for delta in deltas:
				if typeof(delta) != TYPE_DICTIONARY:
					continue
				var rec := delta as Dictionary
				var diff := float(rec.get("delta", 0.0))
				var sign := "+" if diff >= 0 else ""
				entries.append({
					"text": "%s%s %s vs equipped" % [sign, format_delta(diff), display_stat(str(rec.get("stat", "")))],
					"color": comparison_color(diff),
				})
	var summary = item.get("summary_lines", [])
	if typeof(summary) == TYPE_ARRAY:
		for line in summary:
			var text := str(line)
			var diff_value = comparison_delta_from_line(text)
			if diff_value == null:
				continue
			var duplicate := false
			for entry in entries:
				if typeof(entry) == TYPE_DICTIONARY and str((entry as Dictionary).get("text", "")) == text:
					duplicate = true
					break
			if duplicate:
				continue
			entries.append({
				"text": text,
				"color": comparison_color(float(diff_value)),
			})
	return entries


static func append_equip_preview_entries(entries: Array, item: Dictionary) -> void:
	var preview = item.get("equip_preview", {})
	if typeof(preview) != TYPE_DICTIONARY:
		return
	var deltas = (preview as Dictionary).get("deltas", [])
	if typeof(deltas) != TYPE_ARRAY:
		return
	for delta in deltas:
		if typeof(delta) != TYPE_DICTIONARY:
			continue
		var rec := delta as Dictionary
		var diff := float(rec.get("delta", 0.0))
		var sign := "+" if diff >= 0 else ""
		entries.append({
			"text": "%s%s %s preview" % [sign, format_delta(diff), display_stat(str(rec.get("stat", "")))],
			"color": comparison_color(diff),
		})


static func comparison_delta_from_line(text: String):
	var stripped := text.strip_edges()
	if not stripped.contains("vs equipped"):
		return null
	if stripped.length() == 0 or (not stripped.begins_with("+") and not stripped.begins_with("-")):
		return null
	var first_space := stripped.find(" ")
	if first_space <= 1:
		return null
	return float(stripped.substr(0, first_space))


static func comparison_text_lines(item: Dictionary) -> Array:
	var lines: Array = []
	for entry in comparison_entries(item):
		if typeof(entry) == TYPE_DICTIONARY:
			lines.append(str((entry as Dictionary).get("text", "")))
	return lines


static func entry_text(value) -> String:
	if typeof(value) == TYPE_DICTIONARY:
		return str((value as Dictionary).get("text", ""))
	return str(value)


static func requirement_color(met: bool) -> Color:
	return Color("#9ee6a8") if met else Color("#ff6f6f")


static func equip_preview_count(item: Dictionary) -> int:
	var preview = item.get("equip_preview", {})
	if typeof(preview) != TYPE_DICTIONARY:
		return 0
	var deltas = (preview as Dictionary).get("deltas", [])
	if typeof(deltas) != TYPE_ARRAY:
		return 0
	return (deltas as Array).size()


static func comparison_color(delta: float) -> Color:
	if delta > 0:
		return Color("#9ee6a8")
	if delta < 0:
		return Color("#ff9f7a")
	return Color("#d8c7a6")


static func display_stat(stat: String) -> String:
	return StatLabels.display_name(stat)


static func format_delta(delta: float) -> String:
	if absf(delta - roundf(delta)) < 0.0001:
		return str(int(roundf(delta)))
	return "%.2f" % delta


static func item_definition(def_id: String) -> Dictionary:
	return ItemRulesLoader.item_definition(def_id)


static func item_definition_for_item(item: Dictionary) -> Dictionary:
	var def_id := str(item.get("item_def_id", ""))
	var def := item_definition(def_id)
	if not def.is_empty():
		return def
	var template_id := str(item.get("item_template_id", ""))
	if template_id != "":
		def = item_definition(template_id)
		if not def.is_empty():
			return def
	var item_slot := str(item.get("slot", ""))
	if item_slot != "":
		return {
			"equippable": true,
			"slot": item_slot,
		}
	return {}
