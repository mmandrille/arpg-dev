class_name LiveTargetingTrace
extends RefCounted

# Client-bot measurement only. Events deliberately exclude envelopes, account
# details, session IDs, entity IDs, and target coordinates.
const RESPONSE_DISTANCE := 0.015
const CORRECTION_DISTANCE := 0.01
const FRAME_STALL_MS := 33.3
const CombatFeelConfigScript := preload("res://scripts/combat_feel_config.gd")

var _origin_us: int = Time.get_ticks_usec()
var _actions: Array[Dictionary] = []
var _message_action: Dictionary = {}
var _active_index: int = -1
var _last_tick: int = -1
var _last_level: int = 0
var _last_player_id: String = ""
var _last_predicted_pos := Vector3.ZERO
var _last_anchor_pos := Vector3.ZERO
var _last_visual_pos := Vector3.ZERO
var _has_position_sample: bool = false
var _pending_local_upserts: int = 0
var _pending_reset_signal: String = ""
var _authoritative_updates: int = 0
var _stale_correction_ticks: int = 0
var _explicit_resets: int = 0
var _unattributed_position_changes: int = 0
var _frame_stalls: int = 0
var _max_frame_ms: float = 0.0
var _max_raw_authoritative_gap: float = 0.0
var _max_anchor_step: float = 0.0
var _max_visual_step: float = 0.0
var _max_visual_step_on_update: float = 0.0
var _max_visual_offset: float = 0.0
var _late_attack_dispatches: int = 0
var _late_local_swings: int = 0
var _last_clip: String = ""


func on_action(action_kind: String, main: Node) -> void:
	var visual: Node3D = main.get("character_visual") as Node3D
	var anim = main.get("player_anim")
	var clip := str(anim.current_clip()) if anim != null else ""
	_actions.append({
		"kind": action_kind,
		"input_us": Time.get_ticks_usec(),
		"visual_pos": visual.global_position if visual != null else Vector3.ZERO,
		"clip": clip,
		"visible_ms": -1.0,
		"dispatch_ms": -1.0,
		"ack_ms": -1.0,
	})
	_active_index = _actions.size() - 1
	_event("input", {"action": _active_index, "input_kind": action_kind})


func on_command_queued(msg_type: String, message_id: String) -> void:
	if not _is_targeting_command(msg_type):
		return
	if _active_index >= 0:
		_message_action[message_id] = _active_index
	_event("command_queued", {"command": msg_type, "action": _message_action.get(message_id, -1)})


func on_command_dispatch(msg_type: String, message_id: String) -> void:
	if not _is_targeting_command(msg_type):
		return
	var index := int(_message_action.get(message_id, -1))
	if index >= 0 and float(_actions[index]["dispatch_ms"]) < 0.0:
		_actions[index]["dispatch_ms"] = _elapsed_since(int(_actions[index]["input_us"]))
	if _active_index >= 0 and str(_actions[_active_index]["kind"]) == "floor" and msg_type == "action_intent":
		_late_attack_dispatches += 1
		_event("late_attack_dispatch", {"action": _active_index})
	_event("command_dispatch", {"command": msg_type, "action": index})


func on_envelope(env: Dictionary) -> void:
	var kind := str(env.get("type", ""))
	if kind == "session_snapshot":
		_pending_reset_signal = "snapshot"
		return
	if kind == "state_delta":
		_note_delta_signals(env.get("payload", {}))
		return
	if kind not in ["intent_accepted", "intent_rejected"]:
		return
	var payload = env.get("payload", {})
	if not payload is Dictionary:
		return
	var key := "accepted_message_id" if kind == "intent_accepted" else "rejected_message_id"
	var message_id := str(payload.get(key, ""))
	if not _message_action.has(message_id):
		return
	var index := int(_message_action[message_id])
	if float(_actions[index]["ack_ms"]) < 0.0:
		_actions[index]["ack_ms"] = _elapsed_since(int(_actions[index]["input_us"]))
	_event("authoritative_result", {"action": index, "result": "accepted" if kind == "intent_accepted" else "rejected", "server_tick": int(env.get("tick", 0))})
	_message_action.erase(message_id)


func sample(main: Node, delta: float) -> void:
	var frame_ms := delta * 1000.0
	_max_frame_ms = maxf(_max_frame_ms, frame_ms)
	if frame_ms > FRAME_STALL_MS:
		_frame_stalls += 1
		_event("frame_stall", {"frame_ms": snappedf(frame_ms, 0.01)})
	_sample_position(main)
	if _active_index < 0:
		return
	var action: Dictionary = _actions[_active_index]
	var visual: Node3D = main.get("character_visual") as Node3D
	var anim = main.get("player_anim")
	var clip := str(anim.current_clip()) if anim != null else ""
	if str(action["kind"]) == "floor" and clip.begins_with("attack") and not _last_clip.begins_with("attack"):
		_late_local_swings += 1
		_event("late_local_swing", {"action": _active_index})
	_last_clip = clip
	if float(action["visible_ms"]) >= 0.0:
		return
	var moved := visual != null and visual.global_position.distance_to(action["visual_pos"]) > RESPONSE_DISTANCE
	var attacked := clip.begins_with("attack") and not str(action["clip"]).begins_with("attack")
	if moved or attacked:
		_actions[_active_index]["visible_ms"] = _elapsed_since(int(action["input_us"]))
		_event("first_visible_response", {"action": _active_index, "response": "attack" if attacked else "movement", "latency_ms": _actions[_active_index]["visible_ms"]})


func finish() -> Dictionary:
	var visible: Array[float] = []
	var dispatch: Array[float] = []
	var ack: Array[float] = []
	for action in _actions:
		if float(action["visible_ms"]) >= 0.0:
			visible.append(float(action["visible_ms"]))
		if float(action["dispatch_ms"]) >= 0.0:
			dispatch.append(float(action["dispatch_ms"]))
		if float(action["ack_ms"]) >= 0.0:
			ack.append(float(action["ack_ms"]))
	var summary := {
		"transport_profile": OS.get_environment("ARPG_BOT_TRANSPORT_PROFILE") if OS.get_environment("ARPG_BOT_TRANSPORT_PROFILE") != "" else "local",
		"actions": _actions.size(), "visible_samples": visible.size(), "dispatch_samples": dispatch.size(), "ack_samples": ack.size(),
		"input_visible_p50_ms": _percentile(visible, 0.50), "input_visible_p95_ms": _percentile(visible, 0.95),
		"input_dispatch_p95_ms": _percentile(dispatch, 0.95), "input_ack_p95_ms": _percentile(ack, 0.95),
		"frame_stalls_over_33ms": _frame_stalls, "max_frame_ms": snappedf(_max_frame_ms, 0.01),
		"authoritative_player_updates": _authoritative_updates, "stale_correction_ticks": _stale_correction_ticks,
		"explicit_resets": _explicit_resets, "unattributed_position_changes": _unattributed_position_changes,
		"max_raw_authoritative_gap": snappedf(_max_raw_authoritative_gap, 0.001),
		"max_anchor_step": snappedf(_max_anchor_step, 0.001),
		"max_visual_frame_step": snappedf(_max_visual_step, 0.001),
		"max_visual_step_on_update": snappedf(_max_visual_step_on_update, 0.001),
		"max_visual_offset": snappedf(_max_visual_offset, 0.001),
		"visual_offset_bound": CombatFeelConfigScript.movement_smoothing_max_offset(),
		"late_attack_dispatches": _late_attack_dispatches, "late_local_swings": _late_local_swings,
	}
	_event("summary", summary)
	return summary


func _note_delta_signals(raw_payload: Variant) -> void:
	if not raw_payload is Dictionary:
		return
	var payload := raw_payload as Dictionary
	var events = payload.get("events", [])
	if not events is Array:
		events = []
	for raw_event in events:
		if not raw_event is Dictionary:
			continue
		var event := raw_event as Dictionary
		if str(event.get("event_type", "")) == "level_changed":
			_pending_reset_signal = "level_change"
		elif str(event.get("event_type", "")) == "skill_cast" and str(event.get("entity_id", "")) == _last_player_id and str(event.get("skill_id", "")) in ["leap", "charge", "teleport"]:
			_pending_reset_signal = "local_mobility"
	var changes = payload.get("changes", [])
	if not changes is Array:
		changes = []
	for raw_change in changes:
		if not raw_change is Dictionary:
			continue
		var change := raw_change as Dictionary
		if str(change.get("op", "")) not in ["entity_spawn", "entity_update"]:
			continue
		var entity = change.get("entity", {})
		if entity is Dictionary and _last_player_id != "" and str(entity.get("id", "")) == _last_player_id and entity.has("position"):
			_pending_local_upserts += 1


func _sample_position(main: Node) -> void:
	var predicted: Vector3 = main.get("predicted_pos")
	var anchor: Node3D = main.get("player_anchor") as Node3D
	var visual: Node3D = main.get("character_visual") as Node3D
	var anchor_pos := anchor.global_position if anchor != null else predicted
	var visual_pos := visual.global_position if visual != null else anchor_pos
	var offset := Vector2(visual.position.x, visual.position.z).length() if visual != null else 0.0
	var tick := int(main.get("last_server_tick"))
	var level := int(main.get("current_level"))
	_last_player_id = str(main.get("player_id"))
	if not _has_position_sample:
		_has_position_sample = true
		_last_predicted_pos = predicted
		_last_anchor_pos = anchor_pos
		_last_visual_pos = visual_pos
		_last_tick = tick
		_last_level = level
		_pending_local_upserts = 0
		_pending_reset_signal = ""
		return
	var server_step := predicted.distance_to(_last_predicted_pos)
	var anchor_step := anchor_pos.distance_to(_last_anchor_pos)
	var visual_step := visual_pos.distance_to(_last_visual_pos)
	_max_visual_step = maxf(_max_visual_step, visual_step)
	_max_visual_offset = maxf(_max_visual_offset, offset)
	if tick != _last_tick and _pending_local_upserts == 0 and _pending_reset_signal == "" and level == _last_level and float(main.get("reconciliation_delta")) > CORRECTION_DISTANCE:
		_stale_correction_ticks += 1
	if level != _last_level:
		_pending_reset_signal = "level_change"
	var reset_signal := _pending_reset_signal
	if reset_signal in ["local_mobility", "level_change"] and server_step <= CORRECTION_DISTANCE and anchor_step <= CORRECTION_DISTANCE:
		reset_signal = ""
	if _pending_local_upserts > 0 or server_step > CORRECTION_DISTANCE or reset_signal != "":
		var source := "authoritative_update" if _pending_local_upserts > 0 else "local_position_change"
		if reset_signal != "":
			source = reset_signal
			_explicit_resets += 1
		elif _pending_local_upserts > 0:
			_authoritative_updates += 1
			_max_raw_authoritative_gap = maxf(_max_raw_authoritative_gap, float(main.get("reconciliation_delta")))
			_max_visual_step_on_update = maxf(_max_visual_step_on_update, visual_step)
		else:
			_unattributed_position_changes += 1
		_max_anchor_step = maxf(_max_anchor_step, anchor_step)
		_event("player_position_update", {
			"source": source, "server_tick": tick, "tick_gap": tick - _last_tick,
			"local_upserts": _pending_local_upserts,
			"raw_prediction_server_gap": snappedf(float(main.get("reconciliation_delta")), 0.001) if _pending_local_upserts > 0 else -1.0,
			"server_step": snappedf(server_step, 0.001), "anchor_step": snappedf(anchor_step, 0.001),
			"visual_step": snappedf(visual_step, 0.001), "visual_offset": snappedf(offset, 0.001),
			"offset_bound": CombatFeelConfigScript.movement_smoothing_max_offset(),
		})
	_last_predicted_pos = predicted
	_last_anchor_pos = anchor_pos
	_last_visual_pos = visual_pos
	_last_tick = tick
	_last_level = level
	_pending_local_upserts = 0
	if reset_signal != "" or _pending_reset_signal not in ["local_mobility", "level_change"]:
		_pending_reset_signal = ""


func _percentile(values: Array[float], fraction: float) -> float:
	if values.is_empty():
		return -1.0
	values.sort()
	return snappedf(values[mini(values.size() - 1, int(ceilf(values.size() * fraction)) - 1)], 0.01)


func _elapsed_since(start_us: int) -> float:
	return snappedf(float(Time.get_ticks_usec() - start_us) / 1000.0, 0.01)


func _is_targeting_command(msg_type: String) -> bool:
	return msg_type in ["move_to_intent", "action_intent", "directional_attack_intent"]


func _event(kind: String, fields: Dictionary) -> void:
	var record := {"event": kind, "t_ms": snappedf(float(Time.get_ticks_usec() - _origin_us) / 1000.0, 0.01)}
	for key in fields:
		record[key] = fields[key]
	print("[targeting-trace] %s" % JSON.stringify(record))
