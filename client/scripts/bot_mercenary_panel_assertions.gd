class_name BotMercenaryPanelAssertions
extends RefCounted


static func matches(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("mercenary_panel", {})
	if step.has("visible") and bool(panel.get("visible", false)) != bool(step.get("visible", false)):
		return false
	for key in ["price", "gold", "hired_count"]:
		if step.has(key) and int(panel.get(key, -1)) != int(step.get(key, 0)):
			return false
	if step.has("affordable") and bool(panel.get("affordable", false)) != bool(step.get("affordable", false)):
		return false
	for key in ["service_id", "offer_id", "monster_def_id", "hired_entity_id"]:
		if step.has(key) and str(panel.get(key, "")) != str(step.get(key, "")):
			return false
	if step.has("status_contains") and not str(panel.get("status", "")).contains(str(step.get("status_contains", ""))):
		return false
	var companion_bar: Dictionary = state.get("companion_bar", {})
	if step.has("companion_bar_count") and int(companion_bar.get("count", -1)) != int(step.get("companion_bar_count", 0)):
		return false
	if step.has("companion_icon_kind") and not _companion_bar_has_icon_kind(companion_bar, str(step.get("companion_icon_kind", ""))):
		return false
	return true


static func _companion_bar_has_icon_kind(companion_bar: Dictionary, icon_kind: String) -> bool:
	for companion in companion_bar.get("companions", []):
		if typeof(companion) == TYPE_DICTIONARY and str((companion as Dictionary).get("icon_kind", "")) == icon_kind:
			return true
	return false
