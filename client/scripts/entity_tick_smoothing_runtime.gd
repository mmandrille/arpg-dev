class_name EntityTickSmoothingRuntime
extends RefCounted

const EntityTickSmoothingScript := preload("res://scripts/entity_tick_smoothing.gd")
const MovementPresentationLoaderScript := preload("res://scripts/movement_presentation_loader.gd")

var _player_smoothing: EntityTickSmoothing
var _enabled := true
var _projectiles_enabled := true
var _loot_enabled := true
var _duration := 0.1
var _snap_distance := 2.0
var _projectile_snap_distance := 8.0
var _loot_snap_distance := 2.0
var _config_loaded := false


func ensure_config() -> void:
	if _config_loaded:
		return
	_config_loaded = true
	MovementPresentationLoaderScript.ensure_loaded()
	var cfg := MovementPresentationLoaderScript.tick_smoothing()
	_enabled = bool(cfg.get("enabled", true))
	_projectiles_enabled = bool(cfg.get("projectiles_enabled", true))
	_loot_enabled = bool(cfg.get("loot_enabled", true))
	_duration = float(cfg.get("snapshot_interval_seconds", 0.1))
	_snap_distance = float(cfg.get("snap_distance", 2.0))
	_projectile_snap_distance = float(cfg.get("projectile_snap_distance", _snap_distance))
	_loot_snap_distance = float(cfg.get("loot_snap_distance", _snap_distance))
	if _player_smoothing != null:
		_player_smoothing.configure(_duration, _snap_distance)


func player_smoothing() -> EntityTickSmoothing:
	ensure_config()
	if _player_smoothing == null:
		_player_smoothing = EntityTickSmoothingScript.new()
		_player_smoothing.configure(_duration, _snap_distance)
	return _player_smoothing


func smoothing_for_rec(rec: Dictionary, snap_distance: float = -1.0) -> EntityTickSmoothing:
	ensure_config()
	var threshold := snap_distance if snap_distance >= 0.0 else _snap_distance
	if not rec.has("tick_smoothing"):
		var smoothing := EntityTickSmoothingScript.new()
		smoothing.configure(_duration, threshold)
		rec["tick_smoothing"] = smoothing
	elif snap_distance >= 0.0:
		(rec["tick_smoothing"] as EntityTickSmoothing).configure(_duration, threshold)
	return rec["tick_smoothing"] as EntityTickSmoothing


func apply_projectile_authoritative(rec: Dictionary, node: Node3D, target: Vector3, is_new: bool) -> float:
	if node == null:
		return 0.0
	ensure_config()
	var smoothing := smoothing_for_rec(rec, _projectile_snap_distance)
	if is_new or not _enabled or not _projectiles_enabled:
		smoothing.reset(target)
		node.position = target
		_face_projectile_toward(node, target)
		return 0.0
	var prev := node.position
	smoothing.begin_segment(target, prev)
	if not smoothing.is_active():
		node.position = target
	else:
		_face_projectile_toward(node, target)
	return smoothing.last_segment_distance()


func apply_player_authoritative(anchor: Node3D, target: Vector3, snap: bool = false) -> void:
	if anchor == null:
		return
	ensure_config()
	var smoothing := player_smoothing()
	if not _enabled or snap:
		smoothing.reset(target)
		anchor.position = target
		return
	smoothing.begin_segment(target, anchor.position)


func apply_loot_authoritative(rec: Dictionary, node: Node3D, target: Vector3, is_new: bool) -> float:
	if node == null:
		return 0.0
	ensure_config()
	var smoothing := smoothing_for_rec(rec, _loot_snap_distance)
	if is_new or not _enabled or not _loot_enabled:
		smoothing.reset(target)
		node.position = target
		return 0.0
	var prev := node.position
	smoothing.begin_segment(target, prev)
	if not smoothing.is_active():
		node.position = target
		return smoothing.last_segment_distance()
	return smoothing.last_segment_distance()


func apply_entity_authoritative(rec: Dictionary, node: Node3D, target: Vector3, is_new: bool) -> float:
	if node == null:
		return 0.0
	ensure_config()
	var smoothing := smoothing_for_rec(rec)
	if is_new or not _enabled:
		smoothing.reset(target)
		node.position = target
		return 0.0
	var prev := node.position
	smoothing.begin_segment(target, prev)
	if not smoothing.is_active():
		node.position = target
		return smoothing.last_segment_distance()
	return smoothing.last_segment_distance()


func tick_player(anchor: Node3D, delta: float) -> void:
	if anchor == null or not _enabled:
		return
	ensure_config()
	anchor.position = player_smoothing().advance(delta)


func tick_entities(entities: Dictionary, delta: float) -> void:
	if not _enabled:
		return
	ensure_config()
	for rec in entities.values():
		if typeof(rec) != TYPE_DICTIONARY:
			continue
		var node := rec.get("node", null) as Node3D
		var smoothing := rec.get("tick_smoothing", null) as EntityTickSmoothing
		if node == null or smoothing == null:
			continue
		var prev := node.position
		node.position = smoothing.advance(delta)
		if str(rec.get("type", "")) == "projectile":
			_face_projectile_motion(node, prev, node.position)


func get_active_projectile_debug_state(entities: Dictionary) -> Dictionary:
	ensure_config()
	for rec in entities.values():
		if typeof(rec) != TYPE_DICTIONARY or str(rec.get("type", "")) != "projectile":
			continue
		var smoothing := rec.get("tick_smoothing", null) as EntityTickSmoothing
		if smoothing == null:
			continue
		var debug := smoothing.get_debug_state()
		if bool(debug.get("active", false)):
			return debug
	for rec in entities.values():
		if typeof(rec) != TYPE_DICTIONARY or str(rec.get("type", "")) != "projectile":
			continue
		var smoothing := rec.get("tick_smoothing", null) as EntityTickSmoothing
		if smoothing != null:
			return smoothing.get_debug_state()
	return {}


func get_active_loot_debug_state(entities: Dictionary) -> Dictionary:
	return _get_active_type_debug_state(entities, "loot")


func get_player_debug_state() -> Dictionary:
	return player_smoothing().get_debug_state()


func _get_active_type_debug_state(entities: Dictionary, entity_type: String) -> Dictionary:
	ensure_config()
	for rec in entities.values():
		if typeof(rec) != TYPE_DICTIONARY or str(rec.get("type", "")) != entity_type:
			continue
		var smoothing := rec.get("tick_smoothing", null) as EntityTickSmoothing
		if smoothing == null:
			continue
		var debug := smoothing.get_debug_state()
		if bool(debug.get("active", false)):
			return debug
	for rec in entities.values():
		if typeof(rec) != TYPE_DICTIONARY or str(rec.get("type", "")) != entity_type:
			continue
		var smoothing := rec.get("tick_smoothing", null) as EntityTickSmoothing
		if smoothing != null:
			return smoothing.get_debug_state()
	return {}


static func _face_projectile_toward(node: Node3D, target: Vector3) -> void:
	if node == null:
		return
	var from := node.position
	var flat := Vector2(target.x - from.x, target.z - from.z)
	if flat.length_squared() <= 0.0001:
		return
	var look_target := Vector3(target.x, from.y, target.z)
	if node.is_inside_tree():
		node.look_at(look_target, Vector3.UP)
	else:
		node.look_at_from_position(from, look_target, Vector3.UP)


static func _face_projectile_motion(node: Node3D, from: Vector3, to: Vector3) -> void:
	if node == null:
		return
	var flat := Vector2(to.x - from.x, to.z - from.z)
	if flat.length_squared() <= 0.0001:
		return
	var look_target := Vector3(to.x, from.y, to.z)
	node.look_at_from_position(from, look_target, Vector3.UP)
