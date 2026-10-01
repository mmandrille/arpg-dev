class_name AttackContactTrace
extends RefCounted

# Opt-in bot diagnostic. Rows stay in memory until the bot exits so tracing does
# not print once per rendered frame. Record the real window with an OS recorder.
const MAX_ROWS := 1200
const MAX_FRAMES_PER_SWING := 240

static var _enabled := false
static var _origin_usec := 0
static var _rows: Array = []
static var _dropped := 0
static var _sampling := false
static var _attack_clip := ""
static var _saw_attack_clip := false
static var _frames_in_swing := 0
static var _last_sample_usec := 0
static var _last_stage := ""
static var _overlay_layer: CanvasLayer
static var _overlay_label: Label


static func begin(scenario_id: String, viewport: Viewport = null) -> void:
	_reset()
	_enabled = OS.get_environment("ARPG_ATTACK_TRACE") == "1"
	if not _enabled:
		return
	_origin_usec = Time.get_ticks_usec()
	record("trace_start", {"scenario": scenario_id, "viewport": str(viewport.size) if viewport != null else "", "godot": Engine.get_version_info().get("string", "")})
	if viewport != null:
		_overlay_layer = CanvasLayer.new()
		_overlay_layer.layer = 100
		var panel := PanelContainer.new()
		panel.position = Vector2(8, 8)
		panel.mouse_filter = Control.MOUSE_FILTER_IGNORE
		_overlay_label = Label.new()
		_overlay_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
		panel.add_child(_overlay_label)
		_overlay_layer.add_child(panel)
		viewport.add_child(_overlay_layer)


static func begin_test() -> void:
	_reset()
	_enabled = true
	_origin_usec = Time.get_ticks_usec()


static func enabled() -> bool:
	return _enabled


static func record(stage: String, details: Dictionary = {}) -> void:
	if not _enabled:
		return
	var row := details.duplicate(true)
	row["stage"] = stage
	row["t_ms"] = snappedf(float(Time.get_ticks_usec() - _origin_usec) / 1000.0, 0.001)
	row["frame"] = Engine.get_process_frames()
	_append(row)
	_last_stage = stage
	if stage == "swing_start":
		_sampling = true
		_attack_clip = str(details.get("clip", ""))
		_saw_attack_clip = false
		_frames_in_swing = 0
		_last_sample_usec = 0


static func record_result(ev: Dictionary, server_tick: int) -> void:
	if not _enabled:
		return
	var event_type := str(ev.get("event_type", ""))
	if event_type not in ["monster_damaged", "monster_killed", "attack_missed", "attack_blocked"]:
		return
	record("result_received", {
		"event_type": event_type,
		"outcome": str(ev.get("outcome", "")),
		"blocked": bool(ev.get("blocked", false)),
		"target_id": str(ev.get("target_entity_id", ev.get("entity_id", ""))),
		"server_tick": server_tick,
	})


static func record_text(ev: Dictionary, entity_id: String) -> void:
	if not _enabled:
		return
	if str(ev.get("event_type", "")) not in ["monster_damaged", "monster_killed", "attack_missed", "attack_blocked"]:
		return
	record("feedback_text_requested", {"event_type": str(ev.get("event_type", "")), "outcome": str(ev.get("outcome", "")), "target_id": entity_id})


static func sample(state: Dictionary) -> void:
	if not _enabled:
		return
	var local: Dictionary = state.get("local_player_presentation", {})
	var animation: Dictionary = local.get("animation", {})
	var clip := str(animation.get("current_clip", ""))
	var clip_pos_ms := snappedf(float(animation.get("clip_position_s", 0.0)) * 1000.0, 0.1)
	if _overlay_label != null:
		_overlay_label.text = "v496 %0.1f ms | frame %d | %s %0.1f ms | %s" % [
			float(Time.get_ticks_usec() - _origin_usec) / 1000.0,
			Engine.get_process_frames(), clip, clip_pos_ms, _last_stage,
		]
	if not _sampling or _frames_in_swing >= MAX_FRAMES_PER_SWING:
		return
	_frames_in_swing += 1
	var now_usec := Time.get_ticks_usec()
	var lunge: Dictionary = animation.get("melee_lunge", {})
	var visible_text: Array = []
	for entry in state.get("damage_numbers", []):
		if entry is Dictionary:
			visible_text.append({"variant": str(entry.get("variant", "")), "text": str(entry.get("text", ""))})
	record("frame_sample", {
		"clip": clip,
		"clip_position_ms": clip_pos_ms,
		"frame_delta_ms": snappedf(float(now_usec - _last_sample_usec) / 1000.0, 0.001) if _last_sample_usec > 0 else 0.0,
		"moving": bool(animation.get("is_moving", false)),
		"lunge_offset": float(lunge.get("offset_length", 0.0)),
		"visible_text": visible_text,
		"server_tick": int(state.get("last_tick", 0)),
	})
	_last_sample_usec = now_usec
	if clip == _attack_clip:
		_saw_attack_clip = true
	elif _saw_attack_clip and clip in ["idle", "walk", "run"]:
		record("locomotion_return", {"clip": clip, "moving": bool(animation.get("is_moving", false))})
		_sampling = false


static func finish() -> void:
	if not _enabled:
		return
	record("trace_end", {"rows_dropped": _dropped})
	for row in _rows:
		print("[attack-trace] %s" % JSON.stringify(row))
	_reset()


static func rows_for_test() -> Array:
	return _rows.duplicate(true)


static func _append(row: Dictionary) -> void:
	if _rows.size() >= MAX_ROWS:
		_dropped += 1
		return
	_rows.append(row)


static func _reset() -> void:
	if _overlay_layer != null and is_instance_valid(_overlay_layer):
		_overlay_layer.queue_free()
	_overlay_layer = null
	_overlay_label = null
	_enabled = false
	_origin_usec = 0
	_rows.clear()
	_dropped = 0
	_sampling = false
	_attack_clip = ""
	_saw_attack_clip = false
	_frames_in_swing = 0
	_last_sample_usec = 0
	_last_stage = ""
