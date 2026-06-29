extends RefCounted
class_name PerfDebugSampler

const PerfPhaseTimerScript := preload("res://scripts/perf_phase_timer.gd")

const SAMPLE_INTERVAL_SECONDS := 1.0

var enabled: bool = false
var _elapsed: float = 0.0
var _frames: int = 0


func _init() -> void:
	enabled = _truthy(OS.get_environment("ARPG_PERF_DEBUG"))
	if enabled:
		PerfPhaseTimerScript.ensure_enabled()


func sample(delta: float, ready_state: int, tick: int, reconciliation_delta: float, entities: Dictionary, monster_ids: Array) -> void:
	if not enabled:
		return
	_elapsed += delta
	_frames += 1
	if _elapsed < SAMPLE_INTERVAL_SECONDS:
		return
	var counts := _entity_counts(entities)
	var avg_frame_ms: float = (_elapsed / float(max(1, _frames))) * 1000.0
	var phase_suffix := ""
	if enabled:
		phase_suffix = " " + PerfPhaseTimerScript.format_snapshot()
		PerfPhaseTimerScript.reset_frame()
	print("[client-perf] fps=%d avg_frame_ms=%.2f process_ms=%.2f physics_ms=%.2f tick=%d ws=%d recon_delta=%.3f entities=%d monsters=%d live_monsters=%d projectiles=%d loot=%d interactables=%d nodes=%d objects=%d draw_calls=%d primitives=%d%s" % [
		int(Engine.get_frames_per_second()),
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
	_elapsed = 0.0
	_frames = 0


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
