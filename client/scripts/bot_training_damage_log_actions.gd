class_name BotTrainingDamageLogActions
extends RefCounted


static func inject_event(main, action: Dictionary) -> void:
	if main == null or not main.has_method("bot_inject_training_damage_log_event"):
		return
	var event := {
		"event_type": "monster_damaged",
		"monster_def_id": str(action.get("monster_def_id", "town_training_doll")),
		"damage": int(action.get("damage", 1)),
		"outcome": str(action.get("outcome", "hit")),
		"damage_breakdown": action.get("damage_breakdown", []),
	}
	main.bot_inject_training_damage_log_event(event)


static func click_close(main) -> void:
	if main != null and main.has_method("bot_click_training_damage_log_close"):
		main.bot_click_training_damage_log_close()
