class_name SustainedClickInput
extends RefCounted

const HOLD_MOVE_EPSILON := 0.25

var active: bool = false
var mode: String = ""
var target_id: String = ""
var last_ground: Vector2 = Vector2.ZERO


func clear() -> void:
	active = false
	mode = ""
	target_id = ""
	last_ground = Vector2.ZERO


func begin_from_pick(pick: Dictionary) -> bool:
	var kind := str(pick.get("kind", ""))
	match kind:
		"monster":
			active = true
			mode = "attack"
			target_id = str(pick.get("target_id", ""))
			return true
		"floor":
			active = true
			mode = "move"
			var ground: Vector3 = pick.get("ground", Vector3.ZERO)
			last_ground = Vector2(ground.x, ground.z)
			return true
		_:
			clear()
			return false


func should_stop(player_hp: int, entities: Dictionary) -> bool:
	if not active:
		return true

	if player_hp <= 0:
		return true

	if mode != "attack":
		return false

	if target_id == "" or not entities.has(target_id):
		return true

	var rec: Dictionary = entities[target_id]
	if str(rec.get("type", "")) != "monster":
		return true

	if int(rec.get("hp", 1)) <= 0:
		return true

	return false


func can_repeat_move(ground: Vector3) -> bool:
	if mode != "move":
		return false

	var flat := Vector2(ground.x, ground.z)

	return flat.distance_to(last_ground) >= HOLD_MOVE_EPSILON


func mark_move_sent(ground: Vector3) -> void:
	last_ground = Vector2(ground.x, ground.z)
