class_name BotConnectionRecoveryAssertions
extends RefCounted


static func connection_recovery_matches(step: Dictionary, state: Dictionary) -> bool:
	var recovery: Dictionary = state.get("connection_recovery", {})
	var overlay: Dictionary = state.get("connection_overlay", {})
	if step.has("active") and bool(recovery.get("active", false)) != bool(step.get("active", false)):
		return false
	if step.has("blocks_input") and bool(recovery.get("blocks_input", false)) != bool(step.get("blocks_input", false)):
		return false
	if step.has("recovery_count") and int(recovery.get("recovery_count", 0)) != int(step.get("recovery_count", 0)):
		return false
	if step.has("overlay_visible") and bool(overlay.get("visible", false)) != bool(step.get("overlay_visible", false)):
		return false
	if step.has("ws_open") and bool(state.get("ws_open", false)) != bool(step.get("ws_open", false)):
		return false
	if step.has("title_contains"):
		var title := str(overlay.get("title", ""))
		if not title.contains(str(step.get("title_contains", ""))):
			return false
	return true


static func assert_connection_recovery(runner, step: Dictionary, state: Dictionary) -> bool:
	if connection_recovery_matches(step, state):
		return true
	runner._fail("assert_connection_recovery failed: want=%s recovery=%s overlay=%s ws_open=%s step=%d scenario=%s" % [
		str(step),
		str(state.get("connection_recovery", {})),
		str(state.get("connection_overlay", {})),
		str(state.get("ws_open", false)),
		runner._step_index,
		str(runner.scenario.get("id", "?")),
	])
	return false


static func assert_session_unchanged(runner, state: Dictionary) -> bool:
	var remembered_session := str(runner._memory.get("session_id", ""))
	var current_session := str(state.get("current_session_id", ""))
	if current_session != "" and remembered_session != "" and current_session == remembered_session:
		return true
	runner._fail("assert_session_unchanged failed: remembered=%s current=%s step=%d scenario=%s" % [
		remembered_session, current_session, runner._step_index, str(runner.scenario.get("id", "?"))
	])
	return false
