## Bot wait/assert helpers for presentation debug slices.
class_name BotPresentationAssertions
extends RefCounted


static func mobility_skill_smoothing_matches(step: Dictionary, state: Dictionary) -> bool:
	return _mobility_smoothing_state_matches(step, state.get("mobility_skill_smoothing", {}))


static func dungeon_torch_lights_matches(step: Dictionary, state: Dictionary) -> bool:
	var torches: Dictionary = state.get("dungeon_torch_lights", {})
	if step.has("active") and bool(torches.get("active", false)) != bool(step.get("active", false)):
		return false
	if step.has("min_count") and int(torches.get("count", 0)) < int(step.get("min_count", 0)):
		return false
	if step.has("count") and int(torches.get("count", 0)) != int(step.get("count", 0)):
		return false

	return step.has("active") or step.has("min_count") or step.has("count")


static func _mobility_smoothing_state_matches(step: Dictionary, smoothing: Dictionary) -> bool:
	if step.has("active") and bool(smoothing.get("active", false)) != bool(step.get("active", false)):
		return false
	if step.has("skill_id") and str(smoothing.get("skill_id", "")) != str(step.get("skill_id", "")):
		return false

	return step.has("active") or step.has("skill_id")
