class_name EntityTickSmoothingRuntime
extends RefCounted

const EntityTickSmoothingScript := preload("res://scripts/entity_tick_smoothing.gd")
const MovementPresentationLoaderScript := preload("res://scripts/movement_presentation_loader.gd")

var _player_smoothing: EntityTickSmoothing
var _enabled := true
var _duration := 0.1
var _snap_distance := 2.0
var _config_loaded := false


func ensure_config() -> void:
	if _config_loaded:
		return
	_config_loaded = true
	MovementPresentationLoaderScript.ensure_loaded()
	var cfg := MovementPresentationLoaderScript.tick_smoothing()
	_enabled = bool(cfg.get("enabled", true))
	_duration = float(cfg.get("snapshot_interval_seconds", 0.1))
	_snap_distance = float(cfg.get("snap_distance", 2.0))
	if _player_smoothing != null:
		_player_smoothing.configure(_duration, _snap_distance)


func player_smoothing() -> EntityTickSmoothing:
	ensure_config()
	if _player_smoothing == null:
		_player_smoothing = EntityTickSmoothingScript.new()
		_player_smoothing.configure(_duration, _snap_distance)
	return _player_smoothing


func smoothing_for_rec(rec: Dictionary) -> EntityTickSmoothing:
	ensure_config()
	if not rec.has("tick_smoothing"):
		var smoothing := EntityTickSmoothingScript.new()
		smoothing.configure(_duration, _snap_distance)
		rec["tick_smoothing"] = smoothing
	return rec["tick_smoothing"] as EntityTickSmoothing


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
		node.position = smoothing.advance(delta)


func get_player_debug_state() -> Dictionary:
	return player_smoothing().get_debug_state()
