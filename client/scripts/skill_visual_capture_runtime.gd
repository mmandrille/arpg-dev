## Runs the opt-in visual replay capture directive through the existing viewport frame writer.
class_name SkillVisualCaptureRuntime
extends RefCounted

const BotFrameCaptureScript := preload("res://scripts/bot_frame_capture.gd")
const SkillVisualCaptureScript := preload("res://scripts/skill_visual_capture.gd")
const PlayerCameraControllerScript := preload("res://scripts/player_camera_controller.gd")
const CombatVfxScript := preload("res://scripts/combat_vfx.gd")
const CombatFeelConfigScript := preload("res://scripts/combat_feel_config.gd")

var _config: Dictionary = {}
var _camera_config: Dictionary = {}
var _camera_focus := Vector3.ZERO
var _camera_focus_active := false
var _captured_skill_ids: Dictionary = {}
var _jobs: Array = []
var _error := ""


func configure(visual_config: Dictionary) -> void:
	var capture_config = visual_config.get("capture_frame", {})
	_config = capture_config.duplicate(true) if typeof(capture_config) == TYPE_DICTIONARY else {}
	_camera_config = PlayerCameraControllerScript.visual_replay_camera_config(visual_config)
	_camera_focus_active = false


func apply_pending_camera_view(camera_controller: Variant) -> void:
	if not has_pending() or not _camera_focus_active or camera_controller == null:
		return
	camera_controller.apply_visual_replay_view(_camera_focus, float(_camera_config.get("zoom", 1.0)))


func capture_target_id(env: Dictionary) -> String:
	if _config.is_empty() or str(env.get("type", "")) != "state_delta":
		return ""
	var payload = env.get("payload", {})
	if typeof(payload) != TYPE_DICTIONARY:
		return ""
	var events = (payload as Dictionary).get("events", [])
	if typeof(events) != TYPE_ARRAY:
		return ""
	for raw_event in events:
		if typeof(raw_event) != TYPE_DICTIONARY:
			continue
		var selection := SkillVisualCaptureScript.resolve(_config, raw_event as Dictionary, {})
		if not selection.is_empty():
			return str((raw_event as Dictionary).get("target_entity_id", ""))
	return ""


func capture_replay_envelope(
	env: Dictionary,
	viewport: Viewport,
	scenario_title: String,
	tick: int,
	quality: String,
	flush_pending_deltas: Callable,
	camera_controller: Variant,
	player_anchor: Node3D,
	entities: Dictionary,
	camera: Camera3D,
	last_damage_feedback: Dictionary,
) -> void:
	var target_id := capture_target_id(env)
	if target_id == "":
		capture(env, viewport, scenario_title, tick, quality)
		return
	if str(env.get("type", "")) == "state_delta":
		# Replay deltas bypass the live poll loop; apply the hit before diagnostics and capture.
		flush_pending_deltas.call()
	if not _camera_config.is_empty() and camera_controller != null and player_anchor != null:
		_camera_focus = player_anchor.global_position
		if entities.has(target_id):
			var target_node := (entities[target_id] as Dictionary).get("node", null) as Node3D
			if target_node != null and is_instance_valid(target_node):
				_camera_focus = target_node.global_position
		_camera_focus_active = true
		camera_controller.apply_visual_replay_view(_camera_focus, float(_camera_config.get("zoom", 1.0)))
	var diagnostics := _presentation_diagnostics(target_id, entities, camera, viewport, last_damage_feedback)
	capture(env, viewport, scenario_title, tick, quality, diagnostics)


func _presentation_diagnostics(
	target_id: String,
	entities: Dictionary,
	camera: Camera3D,
	viewport: Viewport,
	last_damage_feedback: Dictionary,
) -> Dictionary:
	var diagnostics := {
		"reaction_controller_present": false,
		"enemy_impact_feedback_enabled": CombatFeelConfigScript.enemy_impact_feedback_enabled(),
		"impact_feedback_delta": 0,
		"impact_feedback_count": 0,
		"skill_vfx_burst_count": 0,
		"skill_vfx_bursts": [],
	}
	if entities.has(target_id):
		var record: Dictionary = entities[target_id]
		var reaction = record.get("reaction", null)
		diagnostics["reaction_controller_present"] = reaction != null
		if reaction != null and reaction.has_method("get_debug_state"):
			diagnostics["impact_feedback_count"] = int((reaction.call("get_debug_state") as Dictionary).get("impact_feedback_count", 0))
		var node := record.get("node", null) as Node3D
		if node != null and node.get_parent() != null:
			for child in node.get_parent().get_children():
				if child is GPUParticles3D and child.name.begins_with(CombatVfxScript.BURST_PREFIX):
					var burst := child as GPUParticles3D
					var world_position := burst.global_position
					var burst_state := {
						"name": burst.name,
						"world_position": {"x": world_position.x, "y": world_position.y, "z": world_position.z},
						"emitting": burst.emitting,
						"amount": burst.amount,
						"lifetime": burst.lifetime,
					}
					if camera != null and is_instance_valid(camera):
						var screen_position := camera.unproject_position(world_position)
						burst_state["camera_behind"] = camera.is_position_behind(world_position)
						burst_state["screen_position"] = {"x": screen_position.x, "y": screen_position.y}
						burst_state["on_screen"] = viewport.get_visible_rect().has_point(screen_position)
					(diagnostics["skill_vfx_bursts"] as Array).append(burst_state)
	if str(last_damage_feedback.get("target_entity_id", "")) == target_id:
		diagnostics["impact_feedback_delta"] = int(last_damage_feedback.get("impact_feedback_delta", 0))
	return diagnostics


func capture(
	env: Dictionary,
	viewport: Viewport,
	scenario_title: String,
	tick: int,
	quality: String,
	presentation_diagnostics: Dictionary = {},
) -> void:
	if _config.is_empty() or str(env.get("type", "")) != "state_delta":
		return
	var payload: Dictionary = env.get("payload", {})
	var events = payload.get("events", [])
	if typeof(events) != TYPE_ARRAY:
		return
	for raw_event in events:
		if typeof(raw_event) != TYPE_DICTIONARY:
			continue
		var event: Dictionary = raw_event
		var selection := SkillVisualCaptureScript.resolve(_config, event, _captured_skill_ids)
		if selection.is_empty():
			continue
		var skill_id := str(selection["skill_id"])
		var capture_name := str(selection["name"])
		if bool(_config.get("skip_if_headless", false)) and DisplayServer.get_name() == "headless":
			print("[bot-capture] SKIP windowed-only frame %s in headless visual replay" % capture_name)
			continue
		var fixture := {
			"visual_replay": true,
			"scenario": scenario_title,
			"tick": tick,
			"skill_id": skill_id,
			"effect_id": selection["effect_id"],
			"event_type": event.get("event_type", ""),
			"outcome": event.get("outcome", ""),
			"blocked": bool(event.get("blocked", false)),
			"monster_def_id": event.get("monster_def_id", ""),
			"target_entity_id": event.get("target_entity_id", ""),
			"presentation_diagnostics": presentation_diagnostics.duplicate(true),
		}
		var job = BotFrameCaptureScript.CaptureJob.new()
		_capture_after_frames(
			viewport,
			capture_name,
			quality,
			fixture,
			job,
			SkillVisualCaptureScript.capture_delay_frames(_config),
		)
		_jobs.append({"skill_id": skill_id, "name": capture_name, "job": job})


func _capture_after_frames(
	viewport: Viewport,
	name: String,
	quality: String,
	fixture: Dictionary,
	job: BotFrameCaptureScript.CaptureJob,
	delay_frames: int,
) -> void:
	for _frame in range(delay_frames):
		await RenderingServer.frame_post_draw
	var capture_job = BotFrameCaptureScript.start(viewport, name, quality, fixture)
	while not capture_job.done:
		await RenderingServer.frame_post_draw
	job.error = capture_job.error
	job.done = true


func tick() -> void:
	for index in range(_jobs.size() - 1, -1, -1):
		var entry: Dictionary = _jobs[index]
		var job: Variant = entry.get("job", null)
		if job == null or not job.done:
			continue
		if job.error != "":
			_error = str(job.error)
			push_error("skill visual capture failed for %s: %s" % [entry.get("skill_id", ""), job.error])
		else:
			print("[client] skill visual capture saved: %s" % str(entry.get("name", "")))
		_jobs.remove_at(index)


func has_pending() -> bool:
	return not _jobs.is_empty()


func error_message() -> String:
	return _error
