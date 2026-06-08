class_name DirectionalAttackInput
extends RefCounted

const DIRECTION_EPSILON := 0.0001


static func direction_or_fallback(aim: Vector2, fallback: Vector2) -> Vector2:
	if aim.length_squared() >= DIRECTION_EPSILON:
		return aim.normalized()
	if fallback.length_squared() >= DIRECTION_EPSILON:
		return fallback.normalized()
	return Vector2(1.0, 0.0)


static func payload(direction: Vector2) -> Dictionary:
	var dir := direction_or_fallback(direction, Vector2(1.0, 0.0))
	return {"direction": {"x": dir.x, "y": dir.y}}


static func can_repeat(force_stand_held: bool, left_mouse_held: bool, gameplay_allowed: bool, player_hp: int) -> bool:
	return force_stand_held and left_mouse_held and gameplay_allowed and player_hp > 0
