class_name ItemRequirementViews
extends RefCounted

const TextCatalogScript := preload("res://scripts/text_catalog.gd")
const StatLabels := preload("res://scripts/stat_labels.gd")

static func requirements_met(item: Dictionary) -> bool:
	if item.is_empty():
		return true
	if item.has("requirements_met"):
		return bool(item.get("requirements_met", true))
	var statuses = item.get("requirement_status", [])
	if typeof(statuses) != TYPE_ARRAY:
		return true
	for status in statuses:
		if typeof(status) != TYPE_DICTIONARY:
			continue
		if not bool((status as Dictionary).get("met", true)):
			return false

	return true


static func has_requirement_data(item: Dictionary) -> bool:
	if item.has("requirements_met"):
		return true
	var statuses = item.get("requirement_status", [])
	if typeof(statuses) == TYPE_ARRAY and not statuses.is_empty():
		return true
	var req = item.get("requirements", {})
	return typeof(req) == TYPE_DICTIONARY and not (req as Dictionary).is_empty()


static func shows_invalid_requirement_warning(item: Dictionary, equippable: bool) -> bool:
	if item.is_empty() or not equippable:
		return false
	if bool(item.get("_blocked_by_two_handed", false)):
		return false
	if not has_requirement_data(item):
		return false

	return not requirements_met(item)


static func format_requirement_status(status: Dictionary, met_color: Color = Color("#d8c7a6"), unmet_color: Color = Color("#ff8f70")) -> Dictionary:
	if typeof(status) != TYPE_DICTIONARY:
		return {}
	var stat := str(status.get("stat", ""))
	var met := bool(status.get("met", true))
	if stat == "class":
		var class_id := str(status.get("class_id", ""))
		if class_id == "":
			return {}
		var label := TextCatalogScript.get_text("character.class.%s" % class_id, class_id.capitalize())
		return {
			"text": label,
			"color": met_color if met else unmet_color,
		}
	var required := int(status.get("required", 0))
	if stat == "" or required <= 0:
		return {}
	var current := int(status.get("current", 0))
	if not status.has("met"):
		met = current >= required
	var suffix := "" if met else "(%d)" % (current - required)
	return {
		"text": "%s %d%s" % [StatLabels.display_name(stat), required, suffix],
		"color": met_color if met else unmet_color,
	}
