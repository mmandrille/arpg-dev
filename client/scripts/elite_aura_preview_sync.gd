extends RefCounted

const AuraSoftLightsScript := preload("res://scripts/aura_soft_lights.gd")
const PlayerStatusEffectMarkers := preload("res://scripts/player_status_effect_markers.gd")

const AURA_LIGHT_NAME := AuraSoftLightsScript.AURA_LIGHT_NAME


static func sync_session(
	entities: Dictionary,
	dungeon_generation: Dictionary,
	perspective: bool,
	player_anchor: Node3D,
	light_radius: float,
	sanctuary_radius: float,
	holy_shield_radius: float,
) -> void:
	var hero_pos := Vector3.ZERO
	if player_anchor != null:
		hero_pos = player_anchor.global_position if player_anchor.is_inside_tree() else player_anchor.position
	sync(
		entities,
		dungeon_generation,
		perspective,
		hero_pos,
		light_radius,
		sanctuary_radius,
		holy_shield_radius,
	)


static func sync(
	entities: Dictionary,
	dungeon_generation: Dictionary,
	perspective: bool = false,
	hero_pos: Vector3 = Vector3.ZERO,
	light_radius: float = 0.0,
	sanctuary_radius: float = 5.0,
	holy_shield_radius: float = 5.0,
) -> void:
	var elite_radius := _aura_radius(dungeon_generation)
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
		if not bool(rec.get("monster_pack_leader", false)):
			continue
		var node := rec.get("node", null) as Node3D
		if node == null:
			continue
		var pack_id := str(rec.get("monster_pack_id", ""))
		var alive := int(rec.get("hp", 1)) > 0
		var preview_active := alive and pack_id != "" and active_pack_ids.has(pack_id)
		var state := AuraSoftLightsScript.build_state(
			rec.get("effect_ids", []) if alive else [],
			"monster",
			{
				"monster_pack_leader": true,
				"elite_radius_preview_active": preview_active,
				"elite_aura_radius": elite_radius,
				"sanctuary_radius": sanctuary_radius,
				"holy_shield_radius": holy_shield_radius,
			},
		)
		AuraSoftLightsScript.sync_aura(node, state)
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
		AURA_LIGHT_NAME,
		PlayerStatusEffectMarkers.BURNING_MARKER_NAME,
		PlayerStatusEffectMarkers.PINNING_ROOT_MARKER_NAME,
		PlayerStatusEffectMarkers.STUN_MARKER_NAME,
		PlayerStatusEffectMarkers.ROGUE_MARK_MARKER_NAME,
	]


static func _aura_radius(dungeon_generation: Dictionary) -> float:
	var placement: Dictionary = dungeon_generation.get("monster_placement", {})
	var aura: Dictionary = placement.get("elite_aura", {})
	return maxf(float(aura.get("radius", 0.0)), 0.5)
