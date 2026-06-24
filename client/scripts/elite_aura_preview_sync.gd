extends RefCounted

const PlayerStatusEffectMarkers := preload("res://scripts/player_status_effect_markers.gd")


static func sync(
	entities: Dictionary,
	dungeon_generation: Dictionary,
	perspective: bool = false,
	hero_pos: Vector3 = Vector3.ZERO,
	light_radius: float = 0.0,
) -> void:
	var radius := _aura_radius(dungeon_generation)
	var active_pack_ids: Dictionary = {}
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		if str(rec.get("type", "")) == "monster" \
				and int(rec.get("hp", 1)) > 0 \
				and PlayerStatusEffectMarkers.has_elite_command_effect_id(rec.get("effect_ids", [])):
			var pack_id := str(rec.get("monster_pack_id", ""))
			if pack_id != "":
				active_pack_ids[pack_id] = true
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		if str(rec.get("type", "")) != "monster":
			continue
		var pack_id := str(rec.get("monster_pack_id", ""))
		var active := int(rec.get("hp", 1)) > 0 and bool(rec.get("monster_pack_leader", false)) and pack_id != "" and active_pack_ids.has(pack_id)
		PlayerStatusEffectMarkers.sync_elite_command_radius_preview(rec.get("node", null) as Node3D, active, radius)
	if perspective and light_radius > 0.0:
		_cull_occluded_markers(entities, hero_pos, light_radius)


static func _cull_occluded_markers(entities: Dictionary, hero_pos: Vector3, light_radius: float) -> void:
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		var node := rec.get("node", null) as Node3D
		if node == null:
			continue
		var entity_pos := node.global_position
		var in_range := hero_pos.distance_to(entity_pos) <= light_radius
		var in_los := in_range and not _wall_blocks_los(node, hero_pos, entity_pos)
		if in_los:
			continue
		for marker_name in _all_marker_names():
			var marker := node.find_child(marker_name, false, false)
			if marker != null:
				marker.visible = false


static func _wall_blocks_los(ref_node: Node3D, from_pos: Vector3, to_pos: Vector3) -> bool:
	var world := ref_node.get_world_3d()
	if world == null:
		return false
	var space := world.direct_space_state
	if space == null:
		return false
	var eye := Vector3(0.0, 1.0, 0.0)
	var params := PhysicsRayQueryParameters3D.create(from_pos + eye, to_pos + eye)
	params.collision_mask = 1
	return not space.intersect_ray(params).is_empty()


static func _all_marker_names() -> Array:
	return [
		PlayerStatusEffectMarkers.HOLY_SHIELD_MARKER_NAME,
		PlayerStatusEffectMarkers.SANCTUARY_MARKER_NAME,
		PlayerStatusEffectMarkers.RAGE_MARKER_NAME,
		PlayerStatusEffectMarkers.BURNING_MARKER_NAME,
		PlayerStatusEffectMarkers.ELITE_COMMAND_MARKER_NAME,
		PlayerStatusEffectMarkers.ELITE_COMMAND_RADIUS_PREVIEW_NAME,
		PlayerStatusEffectMarkers.PINNING_ROOT_MARKER_NAME,
		PlayerStatusEffectMarkers.STUN_MARKER_NAME,
		PlayerStatusEffectMarkers.ROGUE_MARK_MARKER_NAME,
	]


static func _aura_radius(dungeon_generation: Dictionary) -> float:
	var placement: Dictionary = dungeon_generation.get("monster_placement", {})
	var aura: Dictionary = placement.get("elite_aura", {})
	return maxf(float(aura.get("radius", 0.0)), 0.5)
