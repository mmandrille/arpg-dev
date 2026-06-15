class_name BotEliteObjectiveAssertions
extends RefCounted


static func matches(step: Dictionary, state: Dictionary) -> bool:
	var tracker: Dictionary = state.get("elite_objective_tracker", {})
	if step.has("visible") and bool(tracker.get("visible", false)) != bool(step.get("visible", false)):
		return false
	if step.has("status") and str(tracker.get("status", "")) != str(step.get("status", "")):
		return false
	if step.has("remaining_leaders") and int(tracker.get("remaining_leaders", -1)) != int(step.get("remaining_leaders", 0)):
		return false
	return true
