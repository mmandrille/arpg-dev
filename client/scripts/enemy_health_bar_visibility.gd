extends RefCounted
class_name EnemyHealthBarVisibility

const ClientSettingsScript := preload("res://scripts/client_settings.gd")


static func should_show(mode: String, entity_id: String, hovered_entity_id: String, pending_action_targets: Dictionary, pending_skill_casts: Dictionary) -> bool:
	if ClientSettingsScript.normalize_monster_health_bar_mode(mode) == ClientSettingsScript.MONSTER_HEALTH_BAR_ALWAYS:
		return true
	return entity_id == hovered_entity_id or _pending_targets_entity(entity_id, pending_action_targets, pending_skill_casts)


static func perspective_visible(node: Node3D, hero_pos: Vector3, is_perspective: bool, light_radius: float) -> bool:
	if not is_perspective or light_radius <= 0.0 or node == null:
		return true
	if hero_pos.distance_to(node.global_position) > light_radius:
		return false
	var world := node.get_world_3d()
	if world == null:
		return true
	var space := world.direct_space_state
	if space == null:
		return true
	var eye := Vector3(0.0, 1.0, 0.0)
	var params := PhysicsRayQueryParameters3D.create(hero_pos + eye, node.global_position + eye)
	params.collision_mask = 1
	return space.intersect_ray(params).is_empty()


static func _pending_targets_entity(entity_id: String, pending_action_targets: Dictionary, pending_skill_casts: Dictionary) -> bool:
	for pending in pending_action_targets.values():
		if pending is Dictionary and str((pending as Dictionary).get("target_id", "")) == entity_id:
			return true
	for pending in pending_skill_casts.values():
		if pending is Dictionary and str((pending as Dictionary).get("target_id", "")) == entity_id:
			return true
	return false
