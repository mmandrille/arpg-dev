class_name ItemRequirementViews
extends RefCounted


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
