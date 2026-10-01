extends SceneTree

const TraceScript := preload("res://scripts/live_targeting_trace.gd")
const CombatFeelConfigScript := preload("res://scripts/combat_feel_config.gd")
const MovementVisualSmoothingScript := preload("res://scripts/movement_visual_smoothing.gd")

var _failures: int = 0


class ProbeMain extends Node3D:
	var predicted_pos := Vector3.ZERO
	var reconciliation_delta: float = 0.0
	var last_server_tick: int = 0
	var current_level: int = 0
	var player_id: String = "local-player"
	var player_anchor: Node3D
	var character_visual: Node3D

	func _init() -> void:
		player_anchor = Node3D.new()
		add_child(player_anchor)
		character_visual = Node3D.new()
		player_anchor.add_child(character_visual)


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var main := ProbeMain.new()
	root.add_child(main)
	var trace = TraceScript.new()
	var smoothing = MovementVisualSmoothingScript.new()
	smoothing.reset(main.player_anchor, main.character_visual)
	trace.sample(main, 0.016)
	trace.on_envelope({"type": "state_delta", "payload": {"changes": "invalid", "events": false}})

	trace.on_envelope({"type": "state_delta", "payload": {
		"changes": [{"op": "entity_update", "entity": {"id": "local-player", "type": "player", "position": {"x": 1, "y": 0}}}],
		"events": [],
	}})
	main.predicted_pos = Vector3(1, 0, 0)
	main.player_anchor.position = main.predicted_pos
	smoothing.preserve_after_anchor_move(main.player_anchor, main.character_visual)
	main.reconciliation_delta = 1.0
	main.last_server_tick = 1
	trace.sample(main, 0.016)

	# A later tick without a local-player change must not turn the retained
	# reconciliation_delta into another measured correction.
	main.last_server_tick = 2
	trace.sample(main, 0.016)

	trace.on_envelope({"type": "state_delta", "payload": {"changes": [], "events": [{"event_type": "level_changed"}]}})
	main.current_level = 1
	main.last_server_tick = 3
	trace.sample(main, 0.016)
	# Arrival can be delivered in a later envelope/frame than level_changed.
	trace.on_envelope({"type": "state_delta", "payload": {
		"changes": [{"op": "entity_spawn", "entity": {"id": "local-player", "type": "player", "position": {"x": 10, "y": 0}}}],
		"events": [],
	}})
	main.predicted_pos = Vector3(10, 0, 0)
	main.player_anchor.position = main.predicted_pos
	smoothing.reset(main.player_anchor, main.character_visual)
	main.last_server_tick = 4
	trace.sample(main, 0.016)
	var summary: Dictionary = trace.finish()
	_check("one fresh authoritative update", int(summary.get("authoritative_player_updates", -1)) == 1)
	_check("retained correction counted as stale", int(summary.get("stale_correction_ticks", -1)) == 1)
	_check("level change separated", int(summary.get("explicit_resets", -1)) == 1)
	var raw_gap := float(summary.get("max_raw_authoritative_gap", -1))
	var visual_step := float(summary.get("max_visual_step_on_update", -1))
	var visual_offset := float(summary.get("max_visual_offset", -1))
	_check("fresh raw gap measured", is_equal_approx(raw_gap, 1.0))
	_check("visual step smaller than raw anchor gap", visual_step >= 0.0 and visual_step < raw_gap)
	_check("offset kept within configured bound", visual_offset >= 0.0 and visual_offset <= CombatFeelConfigScript.movement_smoothing_max_offset())
	var mobility_trace = TraceScript.new()
	mobility_trace.sample(main, 0.016)
	mobility_trace.on_envelope({"type": "state_delta", "payload": {"changes": [], "events": [{"event_type": "skill_cast", "entity_id": "local-player", "skill_id": "teleport"}]}})
	main.last_server_tick = 5
	mobility_trace.sample(main, 0.016)
	main.predicted_pos = Vector3(12, 0, 0)
	main.player_anchor.position = main.predicted_pos
	main.last_server_tick = 6
	mobility_trace.sample(main, 0.016)
	var mobility_summary: Dictionary = mobility_trace.finish()
	_check("mobility signal waits for position change", int(mobility_summary.get("explicit_resets", -1)) == 1)
	main.queue_free()
	if _failures == 0:
		print("[gdtest] PASS: test_live_targeting_trace")
	quit(1 if _failures > 0 else 0)


func _check(label: String, passed: bool) -> void:
	if not passed:
		_failures += 1
		push_error("[gdtest] FAIL: %s" % label)
