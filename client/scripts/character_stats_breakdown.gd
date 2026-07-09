class_name CharacterStatsBreakdown
extends RefCounted

const StatLabels := preload("res://scripts/stat_labels.gd")

const BASE_STATS := StatLabels.BASE_STATS
const FRACTION_PERCENT_STATS := ["hit_chance", "crit_chance", "evade_chance"]
const WHOLE_PERCENT_STATS := ["block_percent"]
const TILES_PER_TICK_STATS := ["movement_speed"]


static func breakdowns_by_key(progression: Dictionary) -> Dictionary:
	var out := {}
	var rows: Array = progression.get("stat_breakdowns", [])
	for row in rows:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		out[str(rec.get("key", ""))] = rec.duplicate(true)
	return out


static func breakdown_summary(progression: Dictionary, key: String, derived_labels: Dictionary) -> String:
	var rec: Dictionary = breakdowns_by_key(progression).get(key, {})
	if rec.is_empty():
		return ""
	var parts: PackedStringArray = PackedStringArray()
	var value := format_stat_value(key, float(rec.get("value", 0.0)))
	parts.append("%s formula:" % breakdown_display_name(key, derived_labels))
	var sources: Array = rec.get("sources", [])
	var formula_terms := source_formula_terms(progression, key, sources, item_names_by_instance_id(sources))
	parts.append_array(formula_terms)
	if rec.get("cap", null) != null:
		parts.append("= %s (cap %s)" % [value, format_stat_value(key, float(rec.get("cap", 0.0)))])
	else:
		parts.append("= %s" % value)
	return "\n".join(parts)


static func weapon_slot_tooltip(
	progression: Dictionary,
	key: String,
	slot_damage: Dictionary,
	damage_key: String,
	derived_labels: Dictionary,
) -> String:
	var sources_key := "min_sources" if damage_key == "min" else "max_sources"
	var sources: Array = slot_damage.get(sources_key, [])
	if sources.is_empty():
		return breakdown_summary(progression, key, derived_labels)
	return breakdown_summary_from_sources(
		progression,
		key,
		float(slot_damage.get(damage_key, 0.0)),
		sources,
		derived_labels,
	)


static func breakdown_summary_from_sources(
	progression: Dictionary,
	key: String,
	value: float,
	sources: Array,
	derived_labels: Dictionary,
) -> String:
	if sources.is_empty():
		return ""
	var parts: PackedStringArray = PackedStringArray()
	parts.append("%s formula:" % breakdown_display_name(key, derived_labels))
	var formula_terms := source_formula_terms(progression, key, sources, item_names_by_instance_id(sources))
	parts.append_array(formula_terms)
	parts.append("= %s" % format_stat_value(key, value))
	return "\n".join(parts)


static func effective_base_stats(progression: Dictionary) -> Dictionary:
	var base: Dictionary = progression.get("base_stats", {})
	var effective = progression.get("effective_base_stats", null)
	if typeof(effective) == TYPE_DICTIONARY:
		return effective as Dictionary
	return base


static func format_number(value: float) -> String:
	if absf(value - roundf(value)) < 0.0001:
		return str(int(roundf(value)))
	var out := "%.2f" % value
	while out.ends_with("0"):
		out = out.left(out.length() - 1)
	if out.ends_with("."):
		out = out.left(out.length() - 1)
	return out


static func format_stat_value(key: String, value: float) -> String:
	if key in FRACTION_PERCENT_STATS:
		return "%s%%" % format_number(value * 100.0)
	if key in WHOLE_PERCENT_STATS:
		return "%s%%" % format_number(value)
	if key in TILES_PER_TICK_STATS:
		return "%.1f t/s" % (value * 10.0)
	return format_number(value)


static func breakdown_display_name(key: String, derived_labels: Dictionary) -> String:
	if key in BASE_STATS:
		return StatLabels.display_name(key)
	return str(derived_labels.get(key, key))


static func format_stat_delta(key: String, value: float) -> String:
	var formatted := format_stat_value(key, value)
	if value > 0.0:
		return "+%s" % formatted
	return formatted


static func source_formula_terms(
	progression: Dictionary,
	key: String,
	sources: Array,
	item_names_by_id: Dictionary,
) -> PackedStringArray:
	var terms := PackedStringArray()
	for source in sources:
		if typeof(source) != TYPE_DICTIONARY:
			continue
		var source_rec := source as Dictionary
		var source_text := source_formula_source(progression, key, source_rec, item_names_by_id)
		terms.append("%s (%s)" % [format_stat_delta(key, float(source_rec.get("value", 0.0))), source_text])
	return terms


static func item_names_by_instance_id(sources: Array) -> Dictionary:
	var out := {}
	for source in sources:
		if typeof(source) != TYPE_DICTIONARY:
			continue
		var source_rec := source as Dictionary
		var item_id := str(source_rec.get("item_instance_id", "")).strip_edges()
		var kind := str(source_rec.get("kind", "")).strip_edges()
		if item_id == "" or not (kind == "equipment_base" or kind == "equipment_roll"):
			continue
		var label := str(source_rec.get("label", "")).strip_edges()
		if label != "" and not out.has(item_id):
			out[item_id] = label
	return out


static func source_formula_source(
	progression: Dictionary,
	key: String,
	source_rec: Dictionary,
	item_names_by_id: Dictionary,
) -> String:
	var label := str(source_rec.get("label", source_rec.get("kind", ""))).strip_edges()
	if label == "":
		label = "Source"
	var kind := str(source_rec.get("kind", "")).strip_edges()
	var item_id := str(source_rec.get("item_instance_id", "")).strip_edges()
	if item_id != "" and (kind == "equipment_base" or kind == "equipment_roll"):
		var item_label := str(item_names_by_id.get(item_id, label))
		if key == "damage_min" or key == "damage_max":
			if kind == "equipment_base":
				return "%s base damage" % item_label
			return "%s rolled damage" % item_label
		return item_label
	var detail := label
	var kind_label := source_kind_label(kind)
	if kind == "character_formula":
		var stat_key := formula_source_stat_key(label)
		if stat_key != "":
			var base_stats: Dictionary = progression.get("base_stats", {})
			var effective_stats := effective_base_stats(progression)
			detail = "%d %s" % [int(effective_stats.get(stat_key, 0)), label]
	if kind_label != "" and kind_label != "Source":
		detail += ", %s" % kind_label
	return detail


static func source_kind_label(kind: String) -> String:
	match kind:
		"base_stat":
			return "Base stat"
		"character_formula":
			return "Character formula"
		"equipment_base":
			return "Item base"
		"equipment_roll":
			return "Item roll"
		"skill_effect":
			return "Skill effect"
		"passive_skill":
			return "Passive skill"
		"set_bonus":
			return "Set bonus"
		"buff":
			return "Buff"
		"debuff":
			return "Debuff"
		"cap":
			return "Cap"
		"clamp":
			return "Clamp"
		_:
			if kind == "":
				return "Source"
			return kind.replace("_", " ").capitalize()


static func formula_source_stat_key(label: String) -> String:
	match label:
		"Strength":
			return "str"
		"Dexterity":
			return "dex"
		"Vitality":
			return "vit"
		"Magic":
			return "magic"
	return ""
