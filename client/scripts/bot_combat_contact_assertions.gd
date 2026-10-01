class_name BotCombatContactAssertions
extends RefCounted


static func matches(stype: String, step: Dictionary, state: Dictionary) -> bool:
	var key := "last_monster_damage_feedback" if stype == "assert_combat_contact" else "attack_buffer"
	var got: Dictionary = state.get(key, {})
	if got.is_empty():
		return false
	for field in ["outcome", "blocked", "target_monster_def_id", "active", "target_id"]:
		if step.has(field) and got.get(field) != step.get(field):
			return false
	for count_field in ["impact_feedback_delta", "queued_count", "replaced_count", "expired_count", "cleared_count"]:
		if (step.has("%s_min" % count_field) or step.has("%s_max" % count_field)) and not got.has(count_field):
			return false
		if step.has("%s_min" % count_field) and int(got.get(count_field, -1)) < int(step.get("%s_min" % count_field)):
			return false
		if step.has("%s_max" % count_field) and int(got.get(count_field, -1)) > int(step.get("%s_max" % count_field)):
			return false
	return true
