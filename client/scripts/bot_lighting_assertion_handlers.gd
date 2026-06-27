## Bot wait/assert helpers for dungeon torch presentation debug.
class_name BotLightingAssertionHandlers
extends RefCounted


static func dungeon_torch_lights_matches(step: Dictionary, state: Dictionary) -> bool:
	var torches: Dictionary = state.get("dungeon_torch_lights", {})
	if step.has("active") and bool(torches.get("active", false)) != bool(step.get("active", false)):
		return false
	if step.has("min_count") and int(torches.get("count", 0)) < int(step.get("min_count", 0)):
		return false
	if step.has("count") and int(torches.get("count", 0)) != int(step.get("count", 0)):
		return false

	return step.has("active") or step.has("min_count") or step.has("count")
