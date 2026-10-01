class_name BotActionHandlers
extends RefCounted

const BotTorchViewpointScript := preload("res://scripts/bot_torch_viewpoint.gd")

static func queue(runner, step: Dictionary, stype: String, state: Dictionary) -> void:
	if stype == "approach_nearest_torch":
		var torch_state: Dictionary = state.get("dungeon_torch_lights", {})
		var target := BotTorchViewpointScript.nearest_approach(
			torch_state.get("positions", []), state.get("player_pos", {}),
			float(step.get("stand_off", 2.5)), int(step.get("torch_index", -1)))
		if target.is_empty():
			runner._fail("approach_nearest_torch found no active torch")
			return
		runner._memory["selected_torch"] = target["mount"]
		var approach: Dictionary = target["approach"]
		print("[bot-client] selected torch mount=%s approach=%s" % [
			str(target["mount"]), str(approach)])
		runner.pending_action = {
			"type": "click_floor", "_type": "click_floor",
			"x": approach["x"], "z": approach["z"],
		}
		return
	if stype == "set_graphics_quality":
		var main = runner._controller._main
		main.client_settings.set_graphics_quality(str(step.get("quality", "balanced")), false, true)
		main._sync_fog_performance_throttle()
		main._sync_fog_and_dungeon_lighting()
		return
	if stype == "remember_session":
		runner._memory["session_id"] = str(state.get("current_session_id", ""))
		return
	if stype == "remember_player_position":
		runner._memory["player_pos"] = (state.get("player_pos", {}) as Dictionary).duplicate(true)
		return
	runner.pending_action = step.duplicate()
	runner.pending_action["_type"] = stype
