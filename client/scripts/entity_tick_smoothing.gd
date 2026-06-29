class_name EntityTickSmoothing
extends RefCounted

var _from := Vector3.ZERO
var _to := Vector3.ZERO
var _display := Vector3.ZERO
var _elapsed := 0.0
var _duration := 0.1
var _base_duration := 0.1
var _snap_distance := 2.0
var _active := false
var _last_segment_distance := 0.0


func configure(duration: float, snap_distance: float) -> void:
	_base_duration = maxf(duration, 0.001)
	_duration = _base_duration
	_snap_distance = snap_distance


func reset(pos: Vector3) -> void:
	_from = pos
	_to = pos
	_display = pos
	_elapsed = 0.0
	_active = false
	_last_segment_distance = 0.0


func begin_segment(target: Vector3, current: Vector3, duration_override: float = -1.0) -> void:
	var flat_delta := Vector2(target.x - current.x, target.z - current.z).length()
	_last_segment_distance = flat_delta
	if flat_delta >= _snap_distance or flat_delta <= 0.001:
		reset(target)
		return
	_duration = _base_duration if duration_override < 0.0 else maxf(duration_override, 0.001)
	_from = current
	_to = target
	_display = current
	_elapsed = 0.0
	_active = true


func advance(delta: float) -> Vector3:
	if not _active:
		return _display
	_elapsed += maxf(delta, 0.0)
	var t := clampf(_elapsed / _duration, 0.0, 1.0)
	_display = _from.lerp(_to, t)
	if t >= 1.0:
		_active = false
		_display = _to
	return _display


func is_active() -> bool:
	return _active


func last_segment_distance() -> float:
	return _last_segment_distance


func get_debug_state() -> Dictionary:
	return {
		"active": _active,
		"elapsed": _elapsed,
		"duration": _duration,
		"segment_duration": _duration,
		"segment_distance": _last_segment_distance,
		"display_x": _display.x,
		"display_z": _display.z,
		"target_x": _to.x,
		"target_z": _to.z,
	}
