extends RefCounted
class_name PerfDebugSampler

const PerfPhaseTimerScript := preload("res://scripts/perf_phase_timer.gd")

const SAMPLE_INTERVAL_SECONDS := 1.0
const MAX_FRAME_INTERVALS := 2048

var enabled: bool = false
var trace_frames: bool = false
var capture_frame_batches: bool = false
var _elapsed: float = 0.0
var _frames: int = 0
var _last_frame_us: int = 0
var _frame_intervals_us: PackedInt32Array = PackedInt32Array()
var _dropped_frame_intervals: int = 0
var _frame_index: int = 0
var _frame_start_usec: int = 0
var _previous_frame_start_usec: int = 0
var _saw_monsters: bool = false


func _init() -> void:
	trace_frames = _truthy(OS.get_environment("ARPG_FIRST_SPAWN_TRACE"))
	capture_frame_batches = _truthy(OS.get_environment("ARPG_PERF_DEBUG"))
	enabled = trace_frames or capture_frame_batches
	if enabled:
		PerfPhaseTimerScript.ensure_enabled()


func begin_frame() -> void:
	if not trace_frames:
		return
	_frame_start_usec = Time.get_ticks_usec()
	PerfPhaseTimerScript.begin_frame()


func sample(delta: float, ready_state: int, tick: int, reconciliation_delta: float, entities: Dictionary, monster_ids: Array, graphics_quality: String = "unknown", settings: ClientSettings = null) -> void:
	if not enabled:
		return
	if trace_frames:
		_trace_frame(delta, tick, entities, monster_ids, graphics_quality)
	if capture_frame_batches:
		var now_us := Time.get_ticks_usec()
		if _last_frame_us > 0:
			var interval_us := now_us - _last_frame_us
			if interval_us > 0 and interval_us <= 2147483647 and _frame_intervals_us.size() < MAX_FRAME_INTERVALS:
				_frame_intervals_us.append(interval_us)
			else:
				_dropped_frame_intervals += 1
		_last_frame_us = now_us
	_elapsed += delta
	_frames += 1
	if _elapsed < SAMPLE_INTERVAL_SECONDS:
		return
	var counts := _entity_counts(entities)
	var avg_frame_ms: float = (_elapsed / float(max(1, _frames))) * 1000.0
	var phase_suffix := ""
	if enabled:
		phase_suffix = " " + PerfPhaseTimerScript.format_snapshot(true)
		PerfPhaseTimerScript.reset_frame()
	print("[client-perf] fps=%d vsync=%d avg_frame_ms=%.2f process_ms=%.2f physics_ms=%.2f tick=%d ws=%d recon_delta=%.3f entities=%d monsters=%d live_monsters=%d projectiles=%d loot=%d interactables=%d nodes=%d objects=%d draw_calls=%d primitives=%d%s" % [
		int(Engine.get_frames_per_second()),
		int(DisplayServer.window_get_vsync_mode()),
		avg_frame_ms,
		float(Performance.get_monitor(Performance.TIME_PROCESS)) * 1000.0,
		float(Performance.get_monitor(Performance.TIME_PHYSICS_PROCESS)) * 1000.0,
		tick,
		ready_state,
		reconciliation_delta,
		entities.size(),
		counts.get("monsters", 0),
		monster_ids.size(),
		counts.get("projectiles", 0),
		counts.get("loot", 0),
		counts.get("interactables", 0),
		int(Performance.get_monitor(Performance.OBJECT_NODE_COUNT)),
		int(Performance.get_monitor(Performance.RENDER_TOTAL_OBJECTS_IN_FRAME)),
		int(Performance.get_monitor(Performance.RENDER_TOTAL_DRAW_CALLS_IN_FRAME)),
		int(Performance.get_monitor(Performance.RENDER_TOTAL_PRIMITIVES_IN_FRAME)),
		phase_suffix,
	])
	if capture_frame_batches:
		_print_frame_batch(tick, settings, graphics_quality)
	_elapsed = 0.0
	_frames = 0
	_frame_intervals_us.clear()
	_dropped_frame_intervals = 0


func _print_frame_batch(tick: int, settings: ClientSettings, graphics_quality: String) -> void:
	var samples := PackedStringArray()
	for interval_us in _frame_intervals_us:
		samples.append(str(interval_us))
	var size := DisplayServer.window_get_size()
	var camera := settings.camera_mode if settings != null else "unknown"
	print("[client-frame-batch] tick=%d n=%d dropped=%d renderer=%s quality=%s camera=%s width=%d height=%d world=%s seed=%s us=%s" % [
		tick,
		_frame_intervals_us.size(),
		_dropped_frame_intervals,
		RenderingServer.get_current_rendering_method(),
		graphics_quality,
		camera,
		size.x,
		size.y,
		OS.get_environment("ARPG_WORLD_ID"),
		OS.get_environment("ARPG_SEED"),
		",".join(samples),
	])


func _trace_frame(delta: float, tick: int, entities: Dictionary, monster_ids: Array, graphics_quality: String) -> void:
	var now_usec := Time.get_ticks_usec()
	var count := monster_ids.size()
	var first_spawn := count > 0 and not _saw_monsters
	_saw_monsters = _saw_monsters or count > 0
	var row := {
		"frame": _frame_index,
		"tick": tick,
		"first_spawn": first_spawn,
		"monsters": count,
		"entities": entities.size(),
		"frame_delta_ms": delta * 1000.0,
		"frame_interval_ms": float(_frame_start_usec - _previous_frame_start_usec) / 1000.0 if _previous_frame_start_usec > 0 else 0.0,
		"process_wall_ms": float(now_usec - _frame_start_usec) / 1000.0 if _frame_start_usec > 0 else 0.0,
		"engine_process_ms": float(Performance.get_monitor(Performance.TIME_PROCESS)) * 1000.0,
		"phases_ms": PerfPhaseTimerScript.take_frame_snapshot(),
		"draw_calls": int(Performance.get_monitor(Performance.RENDER_TOTAL_DRAW_CALLS_IN_FRAME)),
		"resources": int(Performance.get_monitor(Performance.OBJECT_RESOURCE_COUNT)),
		"nodes": int(Performance.get_monitor(Performance.OBJECT_NODE_COUNT)),
		"static_memory_mib": float(Performance.get_monitor(Performance.MEMORY_STATIC)) / 1048576.0,
	}
	if _frame_index == 0:
		row["renderer"] = RenderingServer.get_current_rendering_method()
		row["graphics_quality"] = graphics_quality
	print("[client-spawn-frame] " + JSON.stringify(row))
	_previous_frame_start_usec = _frame_start_usec
	_frame_index += 1


func _entity_counts(entities: Dictionary) -> Dictionary:
	var counts := {"monsters": 0, "projectiles": 0, "loot": 0, "interactables": 0}
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		match str(rec.get("type", "")):
			"monster":
				counts["monsters"] += 1
			"projectile":
				counts["projectiles"] += 1
			"loot":
				counts["loot"] += 1
			"interactable":
				counts["interactables"] += 1
	return counts


func _truthy(value: String) -> bool:
	return value.to_lower() in ["1", "true", "yes", "on"]
