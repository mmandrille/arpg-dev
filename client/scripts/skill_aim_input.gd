class_name SkillAimInput
extends RefCounted

const SkillRulesLoaderScript := preload("res://scripts/skill_rules_loader.gd")
const DirectionalAttackInputScript := preload("res://scripts/directional_attack_input.gd")


static func skill_targeting(skill_id: String) -> String:
	var def := SkillRulesLoaderScript.skill_definition(skill_id)
	return str(def.get("targeting", "direction"))


static func uses_entity_pick_targeting(skill_id: String) -> bool:
	return skill_targeting(skill_id) == "direction_or_target"


static func build_cast_payload(
	skill_id: String,
	target_id: String = "",
	direction: Vector2 = Vector2.ZERO,
	use_nearest_fallback: bool = false,
	player_id: String = "",
	last_facing: Vector2 = Vector2.RIGHT,
	hover_dead_monster_id: String = "",
) -> Dictionary:
	var payload := {"skill_id": skill_id}
	var targeting := skill_targeting(skill_id)
	if targeting == "self":
		if player_id != "":
			payload["target_id"] = player_id
		else:
			payload["direction"] = {"x": last_facing.x, "y": last_facing.y}
		return payload
	if targeting == "self_or_ally_area":
		if player_id != "":
			payload["target_id"] = player_id
		else:
			payload["direction"] = {"x": last_facing.x, "y": last_facing.y}
		return payload
	if targeting == "direction_or_target_area" and target_id == "" and direction.length_squared() <= 0.0001 and use_nearest_fallback:
		if player_id != "":
			payload["target_id"] = player_id
			return payload
	var chosen_target := target_id
	if skill_id == "revive" and chosen_target == "" and hover_dead_monster_id != "":
		chosen_target = hover_dead_monster_id
	if targeting == "direction_or_target" and chosen_target != "":
		payload["target_id"] = chosen_target
		return payload
	if targeting == "direction_or_target" and chosen_target == "" and use_nearest_fallback:
		return {}
	var dir := DirectionalAttackInputScript.direction_or_fallback(direction, last_facing)
	if dir.length_squared() <= 0.0001:
		return {}
	payload["direction"] = {"x": dir.x, "y": dir.y}
	return payload


static func pending_highlights_entity(pending_skill: Dictionary) -> bool:
	if pending_skill.is_empty():
		return false
	if not pending_skill.has("target_id"):
		return false
	var skill_id := str(pending_skill.get("skill_id", ""))
	return uses_entity_pick_targeting(skill_id)
