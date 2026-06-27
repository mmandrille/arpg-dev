class_name MobilitySkillPresentation
extends RefCounted

const MovementPresentationLoaderScript := preload("res://scripts/movement_presentation_loader.gd")
const SkillRulesLoaderScript := preload("res://scripts/skill_rules_loader.gd")
const MainConfigLoaderScript := preload("res://scripts/main_config_loader.gd")

var _owner: Node
var _active: Dictionary = {}


func bind_owner(owner: Node) -> void:
	_owner = owner


func is_active(entity_id: String) -> bool:
	return _active.has(entity_id)


func is_any_active() -> bool:
	return not _active.is_empty()


func clear_entity(entity_id: String) -> void:
	if not _active.has(entity_id):
		return
	var state: Dictionary = _active[entity_id]
	for key in ["tween_main", "tween_y"]:
		if not state.has(key):
			continue
		var tween = state[key]
		if is_instance_valid(tween):
			tween.kill()
	_active.erase(entity_id)


func play_from_skill_cast(
	entity_id: String,
	anchor: Node3D,
	ev: Dictionary,
	landing: Vector3,
	on_finish: Callable = Callable(),
	face_direction: Callable = Callable(),
) -> void:
	if not _mobility_enabled() or _owner == null or anchor == null:
		return
	match str(ev.get("skill_id", "")):
		"leap":
			_play_leap(entity_id, anchor, ev, landing, on_finish, face_direction)
		"charge":
			_play_charge(entity_id, anchor, ev, landing, on_finish, face_direction)
		"teleport":
			_play_teleport(entity_id, anchor, ev, landing, on_finish)
		_:
			pass


func play_travel_teleport(
	entity_id: String,
	anchor: Node3D,
	from_pos: Vector3,
	to_pos: Vector3,
	on_finish: Callable = Callable(),
) -> void:
	if not _mobility_enabled() or _owner == null or anchor == null:
		if on_finish.is_valid():
			on_finish.call(entity_id, to_pos)
		return
	_play_teleport_between(entity_id, anchor, from_pos, to_pos, _teleport_travel_duration(), "teleport_travel", on_finish)


func get_debug_state() -> Dictionary:
	if _active.is_empty():
		return {"active": false}
	var first_key: String = str(_active.keys()[0])
	var state: Dictionary = _active[first_key]

	return {
		"active": true,
		"entity_id": first_key,
		"skill_id": str(state.get("skill_id", "")),
	}


func _play_leap(
	entity_id: String,
	anchor: Node3D,
	ev: Dictionary,
	landing: Vector3,
	on_finish: Callable,
	face_direction: Callable,
) -> void:
	if not ev.has("position"):
		return
	clear_entity(entity_id)
	var start_2d := _vec2_from_dict(ev.get("position", {}))
	var start := Vector3(start_2d.x, landing.y, start_2d.y)
	var skill_id := "leap"
	var visual_duration := _skill_presentation_float(skill_id, "visual_duration")
	var apex := (start + landing) * 0.5 + Vector3(0.0, _skill_presentation_float(skill_id, "visual_height"), 0.0)
	anchor.position = start
	var tween_main := _owner.create_tween()
	tween_main.set_parallel(true)
	tween_main.tween_property(anchor, "position:x", landing.x, visual_duration).set_trans(Tween.TRANS_SINE).set_ease(Tween.EASE_IN_OUT)
	tween_main.tween_property(anchor, "position:z", landing.z, visual_duration).set_trans(Tween.TRANS_SINE).set_ease(Tween.EASE_IN_OUT)
	tween_main.chain().tween_callback(func() -> void:
		_finish(entity_id, anchor, landing, on_finish)
	)
	var tween_y := _owner.create_tween()
	tween_y.tween_property(anchor, "position:y", apex.y, visual_duration * 0.48).set_trans(Tween.TRANS_SINE).set_ease(Tween.EASE_OUT)
	tween_y.tween_property(anchor, "position:y", landing.y, visual_duration * 0.52).set_trans(Tween.TRANS_BOUNCE).set_ease(Tween.EASE_OUT)
	_active[entity_id] = {"skill_id": skill_id, "tween_main": tween_main, "tween_y": tween_y}


func _play_charge(
	entity_id: String,
	anchor: Node3D,
	ev: Dictionary,
	landing: Vector3,
	on_finish: Callable,
	face_direction: Callable,
) -> void:
	if not ev.has("position"):
		return
	clear_entity(entity_id)
	var start_2d := _vec2_from_dict(ev.get("position", {}))
	var start := Vector3(start_2d.x, landing.y, start_2d.y)
	var distance_tiles := maxf(0.01, Vector2(start.x, start.z).distance_to(Vector2(landing.x, landing.z)))
	var speed_tiles_per_second := MainConfigLoaderScript.base_movement_speed() * _skill_mobility_float("charge", "speed_multiplier")
	var duration := maxf(0.08, distance_tiles / speed_tiles_per_second)
	anchor.position = start
	if face_direction.is_valid():
		face_direction.call(Vector2(landing.x - start.x, landing.z - start.z))
	var tween_main := _owner.create_tween()
	tween_main.set_parallel(true)
	tween_main.tween_property(anchor, "position:x", landing.x, duration).set_trans(Tween.TRANS_LINEAR).set_ease(Tween.EASE_IN_OUT)
	tween_main.tween_property(anchor, "position:z", landing.z, duration).set_trans(Tween.TRANS_LINEAR).set_ease(Tween.EASE_IN_OUT)
	tween_main.chain().tween_callback(func() -> void:
		_finish(entity_id, anchor, landing, on_finish)
	)
	_active[entity_id] = {"skill_id": "charge", "tween_main": tween_main}


func _play_teleport(
	entity_id: String,
	anchor: Node3D,
	ev: Dictionary,
	landing: Vector3,
	on_finish: Callable,
) -> void:
	var start_2d := _vec2_from_dict(ev.get("position", {}))
	var start := Vector3(start_2d.x, landing.y, start_2d.y)
	_play_teleport_between(entity_id, anchor, start, landing, _teleport_skill_duration(), "teleport", on_finish)


func _play_teleport_between(
	entity_id: String,
	anchor: Node3D,
	start: Vector3,
	landing: Vector3,
	duration: float,
	skill_id: String,
	on_finish: Callable,
) -> void:
	clear_entity(entity_id)
	anchor.position = start
	var tween_main := _owner.create_tween()
	tween_main.tween_property(anchor, "position", landing, maxf(0.04, duration)).set_trans(Tween.TRANS_SINE).set_ease(Tween.EASE_IN_OUT)
	tween_main.tween_callback(func() -> void:
		_finish(entity_id, anchor, landing, on_finish)
	)
	_active[entity_id] = {"skill_id": skill_id, "tween_main": tween_main}


func _finish(entity_id: String, anchor: Node3D, landing: Vector3, on_finish: Callable) -> void:
	clear_entity(entity_id)
	if anchor != null and is_instance_valid(anchor):
		anchor.position = landing
	if on_finish.is_valid():
		on_finish.call(entity_id, landing)


static func _vec2_from_dict(raw) -> Vector2:
	if typeof(raw) != TYPE_DICTIONARY:
		return Vector2.ZERO
	return Vector2(float(raw.get("x", 0.0)), float(raw.get("y", 0.0)))


static func _skill_presentation_float(skill_id: String, field: String) -> float:
	var presentation: Dictionary = SkillRulesLoaderScript.skill_presentation(skill_id)
	var value := float(presentation.get(field, 0.0))
	if value > 0.0:
		return value
	return 0.1


static func _skill_mobility_float(skill_id: String, field: String) -> float:
	var def: Dictionary = SkillRulesLoaderScript.skill_definition(skill_id)
	var mobility: Dictionary = def.get("mobility", {})
	var value := float(mobility.get(field, 0.0))
	if value > 0.0:
		return value
	return 0.1


static func _mobility_enabled() -> bool:
	MovementPresentationLoaderScript.ensure_loaded()
	return bool(MovementPresentationLoaderScript.mobility_smoothing().get("enabled", true))


static func _teleport_skill_duration() -> float:
	MovementPresentationLoaderScript.ensure_loaded()
	return float(MovementPresentationLoaderScript.mobility_smoothing().get("teleport_duration_seconds", 0.12))


static func _teleport_travel_duration() -> float:
	MovementPresentationLoaderScript.ensure_loaded()
	return float(MovementPresentationLoaderScript.mobility_smoothing().get("teleport_travel_duration_seconds", 0.18))
