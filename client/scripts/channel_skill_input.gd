extends RefCounted

const DirectionalAttackInputScript := preload("res://scripts/directional_attack_input.gd")

var active_skill_id: String = ""
var update_cooldown: float = 0.0


static func is_channel_skill(skill_id: String) -> bool:
	return skill_id == "charge"


static func payload(skill_id: String, phase: String, direction: Vector2, fallback: Vector2) -> Dictionary:
	if skill_id == "" or not (phase in ["start", "update", "stop"]):
		return {}
	var out := {"skill_id": skill_id, "phase": phase}
	if phase != "stop":
		var dir := DirectionalAttackInputScript.direction_or_fallback(direction, fallback)
		if dir.length_squared() <= 0.0001:
			return {}
		out["direction"] = {"x": dir.x, "y": dir.y}
	return out


func start(skill_id: String, sent: bool, send_interval: float) -> bool:
	if sent:
		active_skill_id = skill_id
		update_cooldown = send_interval
	return sent


func tick(delta: float) -> bool:
	update_cooldown -= delta
	return active_skill_id != "" and update_cooldown <= 0.0


func mark_updated(send_interval: float) -> void:
	update_cooldown = send_interval


func stop() -> String:
	var skill_id := active_skill_id
	active_skill_id = ""
	return skill_id


func try_start(skill_id: String, direction: Vector2, fallback: Vector2, blocked_reason: String, send_payload: Callable, face_direction: Callable, show_reject: Callable, send_interval: float) -> bool:
	if blocked_reason != "":
		show_reject.call(blocked_reason)
		return false
	var out := payload(skill_id, "start", direction, fallback)
	if out.is_empty():
		show_reject.call("invalid_target")
		return false
	_face_payload(out, face_direction)
	return start(skill_id, bool(send_payload.call(out)), send_interval)


func tick_and_send(delta: float, pressed: bool, direction: Vector2, fallback: Vector2, send_payload: Callable, face_direction: Callable, send_interval: float) -> void:
	if active_skill_id == "":
		return
	if not pressed:
		stop_and_send(send_payload)
		return
	if not tick(delta):
		return
	var out := payload(active_skill_id, "update", direction, fallback)
	if out.is_empty():
		return
	_face_payload(out, face_direction)
	send_payload.call(out)
	mark_updated(send_interval)


func stop_and_send(send_payload: Callable) -> void:
	if active_skill_id == "":
		return
	var skill_id := stop()
	send_payload.call(payload(skill_id, "stop", Vector2.ZERO, Vector2.RIGHT))


func _face_payload(out: Dictionary, face_direction: Callable) -> void:
	if not out.has("direction"):
		return
	var dir: Dictionary = out["direction"]
	face_direction.call(Vector2(float(dir.get("x", 0.0)), float(dir.get("y", 0.0))))
