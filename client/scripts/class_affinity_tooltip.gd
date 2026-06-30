class_name ClassAffinityTooltip
extends RefCounted

const ACTIVE_COLOR := Color("#9ee6a8")
const INACTIVE_COLOR := Color("#ff6f6f")


static func lines_for_item(item: Dictionary, character_class: String = "") -> Array:
	var statuses = item.get("class_affinity_status", [])
	if typeof(statuses) == TYPE_ARRAY and not statuses.is_empty():
		return _lines_from_status(statuses)
	var affinities := _affinities_from_item(item)
	if affinities.is_empty():
		return []
	if character_class == "":
		return []
	return _lines_from_affinities(affinities, character_class)


static func _lines_from_status(statuses: Array) -> Array:
	var lines: Array = []
	for status in statuses:
		if typeof(status) != TYPE_DICTIONARY:
			continue
		var rec := status as Dictionary
		var display := str(rec.get("display", "")).strip_edges()
		if display == "":
			continue
		lines.append({
			"text": display,
			"color": ACTIVE_COLOR if bool(rec.get("active", false)) else INACTIVE_COLOR,
		})
	return lines


static func _lines_from_affinities(affinities: Array, character_class: String) -> Array:
	var lines: Array = []
	for affinity in affinities:
		if typeof(affinity) != TYPE_DICTIONARY:
			continue
		var rec := affinity as Dictionary
		var item_class := str(rec.get("class", ""))
		var mode := str(rec.get("mode", ""))
		var active := character_class == item_class
		if mode == "penalty_if_not_class":
			active = character_class != item_class
		var stat := str(rec.get("stat", ""))
		var value := int(rec.get("value", 0))
		var display := _display_line(stat, value, item_class, mode)
		if display == "":
			continue
		lines.append({
			"text": display,
			"color": ACTIVE_COLOR if active else INACTIVE_COLOR,
		})
	return lines


static func _affinities_from_item(item: Dictionary) -> Array:
	if item.has("class_affinities") and typeof(item.get("class_affinities")) == TYPE_ARRAY:
		return item.get("class_affinities", [])
	var rolled = item.get("rolled_stats", {})
	if typeof(rolled) != TYPE_DICTIONARY:
		return []
	var rolled_rec := rolled as Dictionary
	if typeof(rolled_rec.get("class_affinities", [])) == TYPE_ARRAY:
		return rolled_rec.get("class_affinities", [])
	return []


static func _display_line(stat: String, value: int, item_class: String, mode: String) -> String:
	var class_label := item_class.capitalize()
	if mode == "penalty_if_not_class":
		class_label = "non-%s" % class_label
	var sign := "+" if value >= 0 else ""
	var suffix := "%"
	match stat:
		"damage_percent", "attack_speed_percent", "reach_percent", "max_mana_percent":
			suffix = "%"
		_:
			suffix = ""
	var stat_label := stat.replace("_percent", "").replace("_", " ")
	return "%s%s%s %s (%s)" % [sign, value, suffix, stat_label, class_label]
