class_name MovementInputPresenter
extends RefCounted

const ClientConstantsScript := preload("res://scripts/client_constants.gd")

var requires_fresh_input: bool = false
var walk_linger: float = 0.0


static func intent_starts_motion(intent_type: String, payload: Dictionary) -> bool:
	if intent_type == "move_to_intent":
		return true
	if intent_type != "move_intent":
		return false
	var direction = payload.get("direction", {})
	if typeof(direction) != TYPE_DICTIONARY:
		return false
	return absf(float(direction.get("x", 0.0))) > 0.0001 or absf(float(direction.get("y", 0.0))) > 0.0001


static func is_force_stand_held() -> bool:
	return Input.is_key_pressed(KEY_SHIFT)


static func send_stop_intent(client, last_server_tick: int, player_hp: int) -> void:
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	client.send("move_intent", last_server_tick, {"direction": {"x": 0, "y": 0}, "duration_ticks": 1})


func mark_walking() -> void:
	walk_linger = ClientConstantsScript.WALK_ANIMATION_LINGER_SECONDS


func tick_walk_linger(delta: float, entities: Dictionary) -> void:
	walk_linger = maxf(0.0, walk_linger - delta)
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		if not rec.has("walk_linger"):
			continue
		rec["walk_linger"] = maxf(0.0, float(rec.get("walk_linger", 0.0)) - delta)
		var ctrl = rec.get("controller", null)
		if ctrl == null:
			continue
		var hp := int(rec.get("hp", 1))
		ctrl.set_locomotion(float(rec.get("walk_linger", 0.0)) > 0.0 and hp > 0)


func local_player_is_walking(
	client,
	player_hp: int,
	input_blocked: bool,
) -> bool:
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN:
		return false
	if player_hp <= 0 or input_blocked or is_force_stand_held() or requires_fresh_input:
		return false
	if walk_linger > 0.0:
		return true
	return Input.is_key_pressed(KEY_W) or Input.is_key_pressed(KEY_A) \
		or Input.is_key_pressed(KEY_S) or Input.is_key_pressed(KEY_D)


func apply_force_stand_hold() -> void:
	requires_fresh_input = true


func on_keyboard_released() -> void:
	requires_fresh_input = false


func begin_force_stand(
	hold_allowed: bool,
	client,
	player_hp: int,
	player_anchor: Node3D,
	on_stop_feel: Callable,
	on_clear_attacks: Callable,
	on_reconcile: Callable,
	last_server_tick: int,
) -> void:
	if not hold_allowed or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	if on_clear_attacks.is_valid():
		on_clear_attacks.call()
	requires_fresh_input = true
	if on_stop_feel.is_valid():
		on_stop_feel.call()
	if player_anchor != null and on_reconcile.is_valid():
		on_reconcile.call()
	send_stop_intent(client, last_server_tick, player_hp)


func try_send_keyboard_move(
	raw_input: Vector2,
	delta: float,
	send_cooldown: float,
	client,
	last_server_tick: int,
	player_hp: int,
	camera_controller,
	player_movement_feel,
	before_send: Callable,
	on_predict_move: Callable,
	on_mark_walking: Callable,
) -> float:
	if raw_input == Vector2.ZERO or requires_fresh_input or send_cooldown > 0.0:
		return send_cooldown
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return send_cooldown
	var dir := Vector2.ZERO
	if camera_controller != null:
		dir = camera_controller.camera_relative_flat_direction(raw_input)
	if before_send.is_valid():
		before_send.call()
	var move_speed: float = player_movement_feel.effective_speed(dir, delta)
	if on_predict_move.is_valid():
		on_predict_move.call(dir, move_speed)
	if on_mark_walking.is_valid():
		on_mark_walking.call()
	client.send("move_intent", last_server_tick, {"direction": {"x": dir.x, "y": dir.y}, "duration_ticks": 2})
	return ClientConstantsScript.SEND_INTERVAL
