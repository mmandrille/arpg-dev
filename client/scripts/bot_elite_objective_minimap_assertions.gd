class_name BotEliteObjectiveMinimapAssertions
extends RefCounted


static func matches(step: Dictionary, state: Dictionary) -> bool:
	var minimap: Dictionary = state.get("elite_objective_minimap", {})
	if step.has("visible") and bool(minimap.get("visible", false)) != bool(step.get("visible", false)):
		return false
	if step.has("has_pin") and bool(minimap.get("has_pin", false)) != bool(step.get("has_pin", false)):
		return false
	if step.has("status") and str(minimap.get("status", "")) != str(step.get("status", "")):
		return false
	if step.has("pin_x_min") and float(minimap.get("pin_x", 0.0)) < float(step.get("pin_x_min", 0.0)):
		return false
	if step.has("pin_x_max") and float(minimap.get("pin_x", 0.0)) > float(step.get("pin_x_max", 1.0)):
		return false
	if step.has("pin_y_min") and float(minimap.get("pin_y", 0.0)) < float(step.get("pin_y_min", 0.0)):
		return false
	if step.has("pin_y_max") and float(minimap.get("pin_y", 0.0)) > float(step.get("pin_y_max", 1.0)):
		return false
	return true
