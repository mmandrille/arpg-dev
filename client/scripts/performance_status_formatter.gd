extends RefCounted
class_name PerformanceStatusFormatter


static func format_status(fps: int, ping_ms: int, ws_state: String, server_tick: int, current_level: int, perf: Dictionary) -> String:
	var ping_text := "--" if ping_ms < 0 else "%dms" % ping_ms
	var lines := [
		"Performance Status",
		"FPS %d  Ping %s  WS %s" % [fps, ping_text, ws_state],
		"Tick %d  Level %s" % [server_tick, _level_text(current_level)],
	]
	if perf.is_empty():
		lines.append("Backend waiting for server performance sample")
		return "\n".join(lines)
	lines.append("Backend total %.1fms  sim %.1fms" % [_num(perf, "total_ms"), _num(perf, "sim_ms")])
	lines.append("AI %.1f  Path %.1f  Combat %.1f" % [_num(perf, "ai_ms"), _num(perf, "pathfind_ms"), _num(perf, "combat_ms")])
	lines.append("Persist %.1f  Broadcast %.1f  Budget %.1f %s" % [_num(perf, "persist_ms"), _num(perf, "broadcast_ms"), _num(perf, "tick_budget_ms"), _budget_text(perf)])
	lines.append("Path req %d  cache %d  nodes %d  moved %d" % [_int_value(perf, "path_requests"), _int_value(perf, "path_cache_hits"), _int_value(perf, "path_nodes_visited"), _int_value(perf, "monsters_moved")])
	lines.append("Room monsters %d/%d  walls %d  entities %d" % [_int_value(perf, "live_monsters"), _int_value(perf, "monsters"), _int_value(perf, "walls"), _int_value(perf, "entities")])
	lines.append("Loop inputs %d  results %d  changes %d  events %d" % [_int_value(perf, "inputs"), _int_value(perf, "results"), _int_value(perf, "changes"), _int_value(perf, "events")])
	if bool(perf.get("degradation_applied", false)):
		lines.append("Overload degradation applied")
	return "\n".join(lines)


static func _level_text(level: int) -> String:
	if level >= 0:
		return "Town"
	return "D%d" % abs(level)


static func _budget_text(perf: Dictionary) -> String:
	if bool(perf.get("tick_over_budget", false)):
		return "over +%.1fms" % _num(perf, "tick_overrun_ms")
	return "ok"


static func _num(perf: Dictionary, key: String) -> float:
	var value = perf.get(key, 0.0)
	if value is int or value is float:
		return float(value)
	return 0.0


static func _int_value(perf: Dictionary, key: String) -> int:
	var value = perf.get(key, 0)
	if value is int or value is float:
		return int(value)
	return 0
