class_name BotStewardHuntAssertions
extends RefCounted


static func banner_matches(step: Dictionary, state: Dictionary) -> bool:
	var banner: Dictionary = state.get("steward_hunt_banner", {})
	if step.has("visible") and bool(banner.get("visible", false)) != bool(step.get("visible", false)):
		return false
	if step.has("contains"):
		var needle := str(step.get("contains", ""))
		var hay := "%s %s" % [str(banner.get("title", "")), str(banner.get("detail", ""))]
		if needle != "" and not hay.contains(needle):
			return false
	return true


static func panel_matches(step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("quest_steward_panel", {})
	if step.has("visible") and bool(panel.get("visible", false)) != bool(step.get("visible", false)):
		return false
	if step.has("offer_count") and int(panel.get("offer_count", 0)) != int(step.get("offer_count", 0)):
		return false
	return true
