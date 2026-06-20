class_name CommandRetargetGrace
extends RefCounted

const DEFAULT_GRACE_SECONDS := 0.22

var active: bool = false
var kind: String = ""
var ground := Vector2.ZERO
var remaining_seconds: float = 0.0
var queued_count: int = 0
var replaced_count: int = 0
var dispatched_count: int = 0
var expired_count: int = 0
var last_dispatched_kind: String = ""
var last_dispatched_ground := Vector2.ZERO


func clear() -> void:
	active = false
	kind = ""
	ground = Vector2.ZERO
	remaining_seconds = 0.0


func queue_floor(next_ground: Vector3, duration_seconds: float = DEFAULT_GRACE_SECONDS) -> void:
	if duration_seconds <= 0.0:
		clear()
		return
	if active:
		replaced_count += 1
	active = true
	kind = "floor"
	ground = Vector2(next_ground.x, next_ground.z)
	remaining_seconds = duration_seconds
	queued_count += 1


func tick(delta: float) -> void:
	if not active:
		return
	remaining_seconds = maxf(0.0, remaining_seconds - maxf(delta, 0.0))
	if remaining_seconds <= 0.0:
		expired_count += 1
		clear()


func pop_ready(local_cooldown: float) -> Dictionary:
	if not active or remaining_seconds <= 0.0 or local_cooldown > 0.0:
		return {}
	var out := {"kind": kind, "ground": Vector3(ground.x, 0.0, ground.y)}
	last_dispatched_kind = kind
	last_dispatched_ground = ground
	dispatched_count += 1
	clear()
	return out


func tick_and_dispatch(delta: float, local_cooldown: float, client, last_server_tick: int, before_dispatch: Callable, mark_walking: Callable) -> bool:
	tick(delta)
	if client == null:
		return false
	var command := pop_ready(local_cooldown)
	if command.is_empty():
		return false
	if before_dispatch.is_valid():
		before_dispatch.call()
	if mark_walking.is_valid():
		mark_walking.call()
	var command_ground: Vector3 = command.get("ground", Vector3.ZERO)
	client.send("move_to_intent", last_server_tick, {"position": {"x": command_ground.x, "y": command_ground.z}})
	return true


func get_debug_state() -> Dictionary:
	return {
		"active": active,
		"kind": kind,
		"ground_x": ground.x,
		"ground_z": ground.y,
		"remaining_seconds": remaining_seconds,
		"queued_count": queued_count,
		"replaced_count": replaced_count,
		"dispatched_count": dispatched_count,
		"expired_count": expired_count,
		"last_dispatched_kind": last_dispatched_kind,
		"last_dispatched_ground_x": last_dispatched_ground.x,
		"last_dispatched_ground_z": last_dispatched_ground.y,
	}
