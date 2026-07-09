class_name ItemTooltipStatSections
extends RefCounted

const StatLabels := preload("res://scripts/stat_labels.gd")
const WeaponRangeTooltipScript := preload("res://scripts/weapon_range_tooltip.gd")

const TOOLTIP_STAT_SEPARATOR := "----------------"

const ROLL_STAT_KEYS := [
	"str", "dex", "vit", "magic", "all_skills", "armor", "block_percent", "attack_speed_percent",
	"hit_chance", "crit_chance", "evade_chance", "max_hp", "max_mana", "health_regen_per_10_seconds",
	"mana_regen_per_10_seconds", "skill_damage_percent", "skill_cooldown_reduction_percent",
	"skill_mana_cost_reduction", "magic_find_percent", "hotbar_slots", "inventory_rows",
	"bonus_cold_damage", "bonus_fire_damage", "bonus_lightning_damage", "bonus_poison_damage",
]


static func is_summary_metadata_line(text: String) -> bool:
	var stripped := text.strip_edges()
	if stripped == "":
		return false
	if stripped.begins_with("Slot:") or stripped.begins_with("Kind:") or stripped.begins_with("Mode:"):
		return true
	if WeaponRangeTooltipScript.is_weapon_metadata_line(stripped):
		return true
	if stripped.begins_with("Set:") or stripped.find("set bonus:") >= 0:
		return true
	if stripped.begins_with("Restores "):
		return true

	return false


static func metadata_lines_from_summary(summary_lines: Array) -> Array:
	var out: Array = []
	for line in summary_lines:
		if typeof(line) == TYPE_DICTIONARY:
			out.append(line)
			continue
		var text := line_text(line)
		if is_summary_metadata_line(text):
			out.append(line)

	return out


static func append_equipment_stat_sections(lines: Array, rolled_stats: Variant, def: Dictionary, separator: String) -> void:
	var base_stat_lines := base_stat_lines_for(def)
	var random_stat_lines := random_stat_lines_for(rolled_stats, def)
	if not base_stat_lines.is_empty():
		lines.append(separator)
		lines.append_array(base_stat_lines)
	if not random_stat_lines.is_empty():
		lines.append(separator)
		lines.append_array(random_stat_lines)


static func base_stats_from_def(def: Dictionary) -> Dictionary:
	var stats: Dictionary = {}
	var stats_value = def.get("base_stats", {})
	if typeof(stats_value) == TYPE_DICTIONARY:
		for key in (stats_value as Dictionary).keys():
			var parsed = numeric_stat_or_null((stats_value as Dictionary).get(key, null))
			if parsed != null:
				stats[str(key)] = int(parsed)
	var damage = def.get("damage", {})
	if typeof(damage) == TYPE_DICTIONARY and not (damage as Dictionary).is_empty():
		stats["damage_min"] = int((damage as Dictionary).get("min", 0))
		stats["damage_max"] = int((damage as Dictionary).get("max", stats.get("damage_min", 0)))

	return stats


static func base_stat_lines_for(def: Dictionary) -> Array:
	var stats := base_stats_from_def(def)
	if stats.is_empty():
		return []

	return stat_lines_for_tooltip(stats, false)


static func random_stat_lines_for(stats_value: Variant, def: Dictionary) -> Array:
	if typeof(stats_value) != TYPE_DICTIONARY:
		return []
	var base_stats: Dictionary = base_stats_from_def(def)
	var deltas: Dictionary = {}
	for key in (stats_value as Dictionary).keys():
		var total_value = numeric_stat_or_null((stats_value as Dictionary).get(key, null))
		if total_value == null:
			continue
		var base_value = numeric_stat_or_null(base_stats.get(key, 0))
		var total := int(total_value)
		var base := int(base_value if base_value != null else 0)
		var delta := total - base
		if delta != 0:
			deltas[key] = delta

	return stat_lines_for_tooltip(deltas, true)


static func stat_lines_for_tooltip(stats: Dictionary, signed: bool) -> Array:
	var lines: Array = []
	if int(stats.get("damage_min", 0)) != 0 or int(stats.get("damage_max", 0)) != 0:
		if signed:
			if int(stats.get("damage_min", 0)) != 0:
				lines.append("%s: %s" % [StatLabels.display_name("damage_min"), format_stat_value(stats.get("damage_min", 0), false)])
			if int(stats.get("damage_max", 0)) != 0:
				lines.append("%s: %s" % [StatLabels.display_name("damage_max"), format_stat_value(stats.get("damage_max", 0), false)])
		else:
			lines.append("Damage: %s-%s" % [str(stats.get("damage_min", "?")), str(stats.get("damage_max", "?"))])
	for key in ROLL_STAT_KEYS:
		if not stats.has(key):
			continue
		var value := int(stats.get(key, 0))
		if value == 0:
			continue
		lines.append("%s: %s" % [
			StatLabels.display_name(key),
			format_stat_value(value, key in ["block_percent", "attack_speed_percent", "hit_chance", "crit_chance", "evade_chance", "skill_damage_percent", "skill_cooldown_reduction_percent", "magic_find_percent"]),
		])

	return lines


static func line_text(line: Variant) -> String:
	if typeof(line) == TYPE_DICTIONARY:
		return str((line as Dictionary).get("text", ""))
	return str(line)


static func numeric_stat_or_null(value: Variant):
	match typeof(value):
		TYPE_INT:
			return int(value)
		TYPE_FLOAT:
			return int(value)
		TYPE_STRING:
			if str(value).is_valid_int():
				return int(value)
	return null


static func format_stat_value(value: int, percent: bool) -> String:
	var sign := "+" if value > 0 else ""
	var suffix := "%" if percent else ""
	return "%s%d%s" % [sign, value, suffix]
