class_name ProjectileFlightPresentation
extends RefCounted

const CombatReachScript := preload("res://scripts/combat_reach.gd")
const ProjectileVisualsScript := preload("res://scripts/projectile_visuals.gd")
const SkillRulesLoaderScript := preload("res://scripts/skill_rules_loader.gd")
const WeaponRangeTooltipScript := preload("res://scripts/weapon_range_tooltip.gd")

const DEFAULT_FLIGHT_DURATION_S := 0.42
const SPAWN_OFFSET := 0.45
const DEFAULT_MAX_VISUAL_DISTANCE := 8.5

static var _monster_defs: Dictionary = {}
static var _monsters_loaded := false


static func try_spawn_from_entity_spawn(
	parent: Node,
	entities_root: Node3D,
	entity: Dictionary,
	entities: Dictionary,
	inventory: Array,
	equipped: Dictionary,
	character_class: String,
	player_id: String,
	player_anchor: Node3D,
	replay_step_delay: float,
	visual_replay_enabled: bool,
) -> bool:
	var target_id := str(entity.get("target_id", ""))
	if target_id == "":
		return false
	var path := _flight_path_from_entity(
		entity,
		entities,
		inventory,
		equipped,
		character_class,
		player_id,
		player_anchor,
	)
	if path.is_empty():
		return false
	_spawn_flight_visual(
		parent,
		entities_root,
		_visual_id_for_def(str(entity.get("projectile_def_id", ""))),
		path["start"],
		path["finish"],
		_flight_duration(replay_step_delay, visual_replay_enabled),
		0,
	)
	return true


static func spawn_from_motion_segment(
	parent: Node,
	entities_root: Node3D,
	entity: Dictionary,
	entities: Dictionary,
	from_pos: Vector3,
	to_pos: Vector3,
	segment_distance: float,
	inventory: Array,
	equipped: Dictionary,
	character_class: String,
	player_id: String,
	replay_step_delay: float,
	visual_replay_enabled: bool,
) -> void:
	if segment_distance <= 0.001:
		return
	var dir := Vector2(to_pos.x - from_pos.x, to_pos.z - from_pos.z)
	if dir.length_squared() <= 0.0001:
		return
	dir = dir.normalized()
	var max_range := _max_range_for_entity(entity, entities, inventory, equipped, character_class, player_id)
	var distance := minf(max_range, DEFAULT_MAX_VISUAL_DISTANCE)
	var start := Vector3(from_pos.x + dir.x * SPAWN_OFFSET, 0.0, from_pos.z + dir.y * SPAWN_OFFSET)
	var finish := Vector3(start.x + dir.x * distance, 0.0, start.z + dir.y * distance)
	_spawn_flight_visual(
		parent,
		entities_root,
		_visual_id_for_def(str(entity.get("projectile_def_id", ""))),
		start,
		finish,
		_flight_duration(replay_step_delay, visual_replay_enabled),
		0,
	)


static func spawn_from_skill_cast(
	parent: Node,
	entities_root: Node3D,
	event: Dictionary,
	entities: Dictionary,
	replay_step_delay: float,
	visual_replay_enabled: bool,
) -> void:
	var projectile_def_id := str(event.get("projectile_def_id", ""))
	if projectile_def_id == "":
		return
	var pos := _vec2_from_dict(event.get("position", {}))
	var dir := _vec2_from_dict(event.get("direction", {}))
	if dir.length_squared() <= 0.0001:
		return
	dir = dir.normalized()
	var max_range := maxf(float(event.get("range", 0.0)), 1.0)
	var start := Vector3(pos.x + dir.x * SPAWN_OFFSET, 0.0, pos.y + dir.y * SPAWN_OFFSET)
	var distance := minf(max_range, DEFAULT_MAX_VISUAL_DISTANCE)
	var target_id := str(event.get("target_entity_id", ""))
	if target_id != "" and entities.has(target_id):
		var target_node := (entities[target_id] as Dictionary).get("node", null) as Node3D
		if target_node != null:
			var target_pos := _node_position(target_node)
			var target_flat := Vector2(target_pos.x - start.x, target_pos.z - start.z)
			if target_flat.length_squared() > 0.0001:
				distance = clampf(target_flat.length(), 1.0, max_range)
	var arrow_count := 1
	var spread_degrees := 0.0
	if str(event.get("skill_id", "")) == "volley":
		arrow_count = 5
		spread_degrees = 32.0
	var duration := _flight_duration(replay_step_delay, visual_replay_enabled)
	for i in range(arrow_count):
		var shot_dir := dir
		if arrow_count > 1:
			var t := 0.0 if arrow_count == 1 else (float(i) / float(arrow_count - 1) - 0.5)
			shot_dir = dir.rotated(deg_to_rad(spread_degrees * t)).normalized()
		var finish := Vector3(start.x + shot_dir.x * distance, 0.0, start.z + shot_dir.y * distance)
		_spawn_flight_visual(parent, entities_root, projectile_def_id, start, finish, duration, i)


static func _flight_path_from_entity(
	entity: Dictionary,
	entities: Dictionary,
	inventory: Array,
	equipped: Dictionary,
	character_class: String,
	player_id: String,
	player_anchor: Node3D,
) -> Dictionary:
	var owner_id := str(entity.get("owner_id", ""))
	var target_id := str(entity.get("target_id", ""))
	var origin := _owner_origin(owner_id, entity, entities, player_id, player_anchor)
	var dir := Vector2.RIGHT
	if target_id != "" and entities.has(target_id):
		var target_pos := _entity_record_position(entities[target_id])
		var flat := Vector2(target_pos.x - origin.x, target_pos.z - origin.z)
		if flat.length_squared() > 0.0001:
			dir = flat.normalized()
		else:
			return {}
	else:
		return {}
	var max_range := _max_range_for_entity(entity, entities, inventory, equipped, character_class, player_id)
	var distance := minf(max_range, DEFAULT_MAX_VISUAL_DISTANCE)
	var target_flat := Vector2(
		_entity_record_position(entities[target_id]).x - origin.x,
		_entity_record_position(entities[target_id]).z - origin.z,
	)
	if target_flat.length_squared() > 0.0001:
		distance = clampf(target_flat.length(), 1.0, max_range)
	var start := Vector3(origin.x + dir.x * SPAWN_OFFSET, 0.0, origin.z + dir.y * SPAWN_OFFSET)
	var finish := Vector3(start.x + dir.x * distance, 0.0, start.z + dir.y * distance)
	return {"start": start, "finish": finish}


static func _max_range_for_entity(
	entity: Dictionary,
	entities: Dictionary,
	inventory: Array,
	equipped: Dictionary,
	character_class: String,
	player_id: String,
) -> float:
	var owner_id := str(entity.get("owner_id", ""))
	if owner_id != "" and owner_id == player_id:
		return CombatReachScript.local_player_attack_reach(inventory, equipped, character_class)
	if owner_id != "" and entities.has(owner_id):
		var owner: Dictionary = entities[owner_id]
		match str(owner.get("type", "")):
			"player":
				return CombatReachScript.local_player_attack_reach(inventory, equipped, character_class)
			"monster", "companion":
				return _monster_attack_range(str(owner.get("monster_def_id", "")))
	return maxf(DEFAULT_MAX_VISUAL_DISTANCE, 1.0)


static func _owner_origin(
	owner_id: String,
	entity: Dictionary,
	entities: Dictionary,
	player_id: String,
	player_anchor: Node3D,
) -> Vector3:
	if owner_id != "" and owner_id == player_id and player_anchor != null:
		return _node_position(player_anchor)
	if owner_id != "" and entities.has(owner_id):
		return _entity_record_position(entities[owner_id])

	return _entity_position(entity)


static func _monster_attack_range(monster_def_id: String) -> float:
	_ensure_monsters_loaded()
	if monster_def_id == "":
		return DEFAULT_MAX_VISUAL_DISTANCE
	var def: Dictionary = _monster_defs.get(monster_def_id, {})
	var attack_range := float(def.get("attack_range", 0.0))
	if attack_range > 0.0:
		return attack_range
	return DEFAULT_MAX_VISUAL_DISTANCE


static func _visual_id_for_def(projectile_def_id: String) -> String:
	if projectile_def_id == "":
		return ""
	var presentation := SkillRulesLoaderScript.skill_presentation(projectile_def_id)
	var visual := str(presentation.get("projectile_visual", ""))
	if visual != "":
		return visual
	var skill := SkillRulesLoaderScript.skill_definition(projectile_def_id)
	var projectile: Dictionary = skill.get("projectile", {})
	return str(projectile.get("visual", projectile_def_id))


static func _flight_duration(replay_step_delay: float, visual_replay_enabled: bool) -> float:
	if visual_replay_enabled:
		return maxf(replay_step_delay * 0.85, 0.32)
	return DEFAULT_FLIGHT_DURATION_S


static func _spawn_flight_visual(
	parent: Node,
	entities_root: Node3D,
	projectile_def_id: String,
	start: Vector3,
	finish: Vector3,
	duration: float,
	index: int,
) -> void:
	if projectile_def_id == "":
		return
	var node := ProjectileVisualsScript.make_node(projectile_def_id)
	node.name = "ProjectileFlight_%s" % projectile_def_id
	node.position = start
	entities_root.add_child(node)
	var flat := Vector2(finish.x - start.x, finish.z - start.z)
	if flat.length_squared() > 0.0001:
		var look_target := Vector3(finish.x, start.y, finish.z)
		if node.is_inside_tree():
			node.look_at(look_target, Vector3.UP)
		else:
			node.look_at_from_position(start, look_target, Vector3.UP)
	var tween := parent.create_tween()
	if index > 0:
		tween.tween_interval(0.025 * float(index))
	tween.tween_property(node, "position", finish, duration).set_trans(Tween.TRANS_LINEAR)
	tween.tween_callback(node.queue_free)


static func _ensure_monsters_loaded() -> void:
	if _monsters_loaded:
		return
	_monsters_loaded = true
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/monsters.v0.json")
	if not FileAccess.file_exists(path):
		return
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		return
	_monster_defs = (parsed as Dictionary).get("monsters", {})


static func _entity_position(entity: Dictionary) -> Vector3:
	var pos: Dictionary = entity.get("position", {})
	return Vector3(float(pos.get("x", 0.0)), 0.0, float(pos.get("y", 0.0)))


static func _entity_record_position(rec: Dictionary) -> Vector3:
	var node := rec.get("node", null) as Node3D
	if node != null:
		return _node_position(node)
	return Vector3.ZERO


static func _node_position(node: Node3D) -> Vector3:
	if node == null:
		return Vector3.ZERO
	return node.global_position if node.is_inside_tree() else node.position


static func _vec2_from_dict(value) -> Vector2:
	if value is Dictionary:
		return Vector2(float(value.get("x", 0.0)), float(value.get("y", 0.0)))
	return Vector2.ZERO
