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


static func wall_occlusion_matches(step: Dictionary, state: Dictionary) -> bool:
	return wall_occlusion_mismatch(step, state) == ""


static func wall_occlusion_mismatch(step: Dictionary, state: Dictionary) -> String:
	var occlusion: Dictionary = state.get("wall_occlusion", {})
	for key in ["active"]:
		if step.has(key) and bool(occlusion.get(key, false)) != bool(step.get(key, true)):
			return "%s want=%s got=%s occlusion=%s" % [key, str(step.get(key, true)), str(occlusion.get(key, null)), str(occlusion)]
	for key in ["faded_wall_count", "target_count"]:
		var min_key := "%s_min" % key
		if step.has(min_key) and int(occlusion.get(key, 0)) < int(step.get(min_key, 0)):
			return "%s want_min=%s got=%s occlusion=%s" % [key, str(step.get(min_key, 0)), str(occlusion.get(key, null)), str(occlusion)]
		if step.has(key) and int(occlusion.get(key, -999999)) != int(step.get(key, 0)):
			return "%s want=%s got=%s occlusion=%s" % [key, str(step.get(key, 0)), str(occlusion.get(key, null)), str(occlusion)]
	for key in ["min_faded_alpha", "faded_alpha", "opaque_alpha"]:
		var max_key := "%s_max" % key
		var min_key := "%s_min" % key
		if step.has(max_key) and float(occlusion.get(key, 1.0)) > float(step.get(max_key, 1.0)):
			return "%s want_max=%s got=%s occlusion=%s" % [key, str(step.get(max_key, 1.0)), str(occlusion.get(key, null)), str(occlusion)]
		if step.has(min_key) and float(occlusion.get(key, 0.0)) < float(step.get(min_key, 0.0)):
			return "%s want_min=%s got=%s occlusion=%s" % [key, str(step.get(min_key, 0.0)), str(occlusion.get(key, null)), str(occlusion)]
		if step.has(key) and abs(float(occlusion.get(key, -999999.0)) - float(step.get(key, 0.0))) > float(step.get("tolerance", 0.001)):
			return "%s want=%s got=%s occlusion=%s" % [key, str(step.get(key, 0.0)), str(occlusion.get(key, null)), str(occlusion)]

	return ""
