class_name PerfPhaseTimer
extends RefCounted

static var _enabled: bool = false
static var _phases_ms: Dictionary = {}
static var _trace_frames: bool = false
static var _frame_phases_ms: Dictionary = {}


static func ensure_enabled() -> void:
	_trace_frames = OS.get_environment("ARPG_FIRST_SPAWN_TRACE").to_lower() in ["1", "true", "yes", "on"]
	_enabled = _enabled or _trace_frames or OS.get_environment("ARPG_PERF_DEBUG").to_lower() in ["1", "true", "yes", "on"]


static func begin_frame() -> void:
	if _trace_frames:
		_frame_phases_ms.clear()


static func take_frame_snapshot() -> Dictionary:
	var result := _frame_phases_ms.duplicate()
	_frame_phases_ms.clear()
	return result


static func reset_frame() -> void:
	if not _enabled:
		return
	_phases_ms = {}


static func add_ms(phase: String, elapsed_usec: int) -> void:
	if not _enabled or phase == "":
		return
	var elapsed_ms := float(elapsed_usec) / 1000.0
	_phases_ms[phase] = float(_phases_ms.get(phase, 0.0)) + elapsed_ms
	if _trace_frames:
		_frame_phases_ms[phase] = float(_frame_phases_ms.get(phase, 0.0)) + elapsed_ms


static func measure_usec(phase: String, start_usec: int) -> void:
	add_ms(phase, Time.get_ticks_usec() - start_usec)


static func snapshot_ms() -> Dictionary:
	return _phases_ms.duplicate()


static func format_snapshot(rank_by_value: bool = false) -> String:
	if _phases_ms.is_empty():
		return ""
	var keys: Array = _phases_ms.keys()
	if rank_by_value:
		keys.sort_custom(func(a: String, b: String) -> bool:
			return float(_phases_ms[a]) > float(_phases_ms[b])
		)
	else:
		keys.sort()
	var parts: PackedStringArray = []
	for key in keys:
		parts.append("%s=%.2f" % [str(key), float(_phases_ms[key])])
	return " ".join(parts)
