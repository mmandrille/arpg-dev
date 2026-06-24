class_name PlayerMovementFeel
extends RefCounted

const MainConfigLoaderScript := preload("res://scripts/main_config_loader.gd")

var _hold_seconds: float = 0.0
var _last_dir := Vector2.ZERO
var _server_speed: float = 0.0  # tiles/tick from server derived_stats; 0 = use config fallback


func reset() -> void:
	_hold_seconds = 0.0
	_last_dir = Vector2.ZERO


func on_stop() -> void:
	_hold_seconds = 0.0
	_last_dir = Vector2.ZERO


func speed_multiplier(direction: Vector2, delta: float) -> float:
	if direction == Vector2.ZERO:
		on_stop()
		return 0.0
	var normalized := direction.normalized()
	if _last_dir != Vector2.ZERO:
		var dot := normalized.dot(_last_dir)
		var grace := MainConfigLoaderScript.movement_direction_grace_seconds()
		# Small correction (dot >= 0.5) after holding past grace window → no reset.
		# Sharp turn (dot < 0.5) or not yet past grace → reset.
		var is_small_correction := dot >= 0.5 and _hold_seconds >= grace
		if not is_small_correction and dot < 0.7:
			_hold_seconds = 0.0
	_last_dir = normalized
	_hold_seconds += maxf(delta, 0.0)
	var accel_seconds := MainConfigLoaderScript.movement_acceleration_seconds()
	if accel_seconds <= 0.0:
		return 1.0
	var min_factor := MainConfigLoaderScript.movement_min_speed_factor()
	var ramp := clampf(_hold_seconds / accel_seconds, min_factor, 1.0)
	return ramp


func set_server_speed(tiles_per_tick: float) -> void:
	_server_speed = tiles_per_tick


func effective_speed(direction: Vector2, delta: float) -> float:
	var base_speed := _server_speed if _server_speed > 0.0 else MainConfigLoaderScript.base_movement_speed()
	return base_speed * speed_multiplier(direction, delta)


func get_debug_state() -> Dictionary:
	return {
		"hold_seconds": _hold_seconds,
		"last_dir": {"x": _last_dir.x, "y": _last_dir.y},
	}
