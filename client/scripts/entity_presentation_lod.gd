class_name EntityPresentationLod
extends RefCounted

const MainConfigLoaderScript := preload("res://scripts/main_config_loader.gd")

static var _lod_state: Dictionary = {}


static func should_apply(live_monster_count: int) -> bool:
	var cfg := MainConfigLoaderScript.presentation_lod()
	return live_monster_count >= int(cfg.get("min_live_monsters", 24))


static func apply_monster(node: Node3D, distance_sq: float, live_monster_count: int) -> void:
	if node == null or not should_apply(live_monster_count):
		_restore_monster(node)
		return
	var threshold := MainConfigLoaderScript.presentation_lod_distance_threshold()
	var cheap := distance_sq > threshold * threshold
	for child in node.find_children("*", "GeometryInstance3D", true, false):
		var geom := child as GeometryInstance3D
		if geom == null:
			continue
		var key := str(geom.get_instance_id())
		if not _lod_state.has(key):
			_lod_state[key] = {
				"cast_shadow": geom.cast_shadow,
				"gi_mode": geom.gi_mode,
			}
		if cheap:
			geom.cast_shadow = GeometryInstance3D.SHADOW_CASTING_SETTING_OFF
			geom.gi_mode = GeometryInstance3D.GI_MODE_DISABLED
		else:
			var saved: Dictionary = _lod_state[key]
			geom.cast_shadow = int(saved.get("cast_shadow", GeometryInstance3D.SHADOW_CASTING_SETTING_ON))
			geom.gi_mode = int(saved.get("gi_mode", GeometryInstance3D.GI_MODE_STATIC))


static func refresh_monsters(entities: Dictionary, monster_ids: Array, hero_pos: Vector3) -> void:
	var live_count := 0
	for mid in monster_ids:
		var rec: Dictionary = entities.get(str(mid), {})
		if str(rec.get("type", "")) != "monster":
			continue
		var node := rec.get("node", null) as Node3D
		if node == null:
			continue
		live_count += 1
	if not should_apply(live_count):
		for mid in monster_ids:
			var rec: Dictionary = entities.get(str(mid), {})
			_restore_monster(rec.get("node", null) as Node3D)
		return
	for mid in monster_ids:
		var rec: Dictionary = entities.get(str(mid), {})
		var node := rec.get("node", null) as Node3D
		if node == null:
			continue
		var dist_sq := Vector2(node.global_position.x - hero_pos.x, node.global_position.z - hero_pos.z).length_squared()
		apply_monster(node, dist_sq, live_count)


static func _restore_monster(node: Node3D) -> void:
	if node == null:
		return
	for child in node.find_children("*", "GeometryInstance3D", true, false):
		var geom := child as GeometryInstance3D
		if geom == null:
			continue
		var key := str(geom.get_instance_id())
		if not _lod_state.has(key):
			continue
		var saved: Dictionary = _lod_state[key]
		geom.cast_shadow = int(saved.get("cast_shadow", GeometryInstance3D.SHADOW_CASTING_SETTING_ON))
		geom.gi_mode = int(saved.get("gi_mode", GeometryInstance3D.GI_MODE_STATIC))


static func reset_for_tests() -> void:
	_lod_state.clear()
