class_name SkillBonusTooltip
extends RefCounted

const ACTIVE_COLOR := Color("#9ee6a8")
const INACTIVE_WRONG_CLASS_COLOR := Color("#ff6f6f")
const INACTIVE_UNLEARNED_COLOR := Color("#9a9a9a")


static func lines_for_item(item: Dictionary, character_class: String = "") -> Array:
	var statuses = item.get("skill_bonus_status", [])
	if typeof(statuses) != TYPE_ARRAY or statuses.is_empty():
		return []
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
			"color": _color_for_status(rec, character_class),
		})
	return lines


static func _color_for_status(rec: Dictionary, character_class: String) -> Color:
	if bool(rec.get("active", false)):
		return ACTIVE_COLOR
	var skill_class := str(rec.get("skill_class", ""))
	if skill_class != "" and character_class != "" and skill_class != character_class:
		return INACTIVE_WRONG_CLASS_COLOR

	return INACTIVE_UNLEARNED_COLOR
