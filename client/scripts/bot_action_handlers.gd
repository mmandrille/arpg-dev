class_name BotActionHandlers
extends RefCounted


static func queue(runner, step: Dictionary, stype: String, state: Dictionary) -> void:
	if stype == "remember_session":
		runner._memory["session_id"] = str(state.get("current_session_id", ""))
		return
	if stype == "remember_player_position":
		runner._memory["player_pos"] = (state.get("player_pos", {}) as Dictionary).duplicate(true)
		return
	runner.pending_action = step.duplicate()
	runner.pending_action["_type"] = stype

