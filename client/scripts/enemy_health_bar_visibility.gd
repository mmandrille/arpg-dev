extends RefCounted
class_name EnemyHealthBarVisibility

const ClientSettingsScript := preload("res://scripts/client_settings.gd")


static func should_show(mode: String, entity_id: String, hovered_entity_id: String, pending_action_targets: Dictionary, pending_skill_casts: Dictionary) -> bool:
	if ClientSettingsScript.normalize_monster_health_bar_mode(mode) == ClientSettingsScript.MONSTER_HEALTH_BAR_ALWAYS:
		return true
	return entity_id == hovered_entity_id or _pending_targets_entity(entity_id, pending_action_targets, pending_skill_casts)


static func _pending_targets_entity(entity_id: String, pending_action_targets: Dictionary, pending_skill_casts: Dictionary) -> bool:
	for pending in pending_action_targets.values():
		if pending is Dictionary and str((pending as Dictionary).get("target_id", "")) == entity_id:
			return true
	for pending in pending_skill_casts.values():
		if pending is Dictionary and str((pending as Dictionary).get("target_id", "")) == entity_id:
			return true
	return false
