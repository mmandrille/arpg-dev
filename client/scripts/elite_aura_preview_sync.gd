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
		_cull_far_aura_markers(entities, hero_pos, light_radius)


static func _cull_far_aura_markers(entities: Dictionary, hero_pos: Vector3, light_radius: float) -> void:
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		var node := rec.get("node", null) as Node3D
		if node == null or hero_pos.distance_to(node.global_position) <= light_radius:
			continue
		for marker_name in [
			PlayerStatusEffectMarkers.ELITE_COMMAND_RADIUS_PREVIEW_NAME,
			PlayerStatusEffectMarkers.ELITE_COMMAND_MARKER_NAME,
		]:
			var marker := node.find_child(marker_name, false, false)
			if marker != null:
				marker.visible = false


static func _aura_radius(dungeon_generation: Dictionary) -> float:
	var placement: Dictionary = dungeon_generation.get("monster_placement", {})
	var aura: Dictionary = placement.get("elite_aura", {})
	return maxf(float(aura.get("radius", 0.0)), 0.5)
