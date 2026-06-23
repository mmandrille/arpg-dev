class_name CameraImpactFeedback
extends RefCounted

const MAX_OFFSET := 0.22
const DECAY_SPEED := 14.0

static var _shake_strength: float = 0.0
static var _shake_offset := Vector3.ZERO


static func apply_from_damage(damage: int, max_hp: int) -> void:
	if damage <= 0:
		return
	var ratio := float(damage) / float(maxi(1, max_hp))
	_shake_strength = clampf(_shake_strength + ratio * 0.35, 0.0, 1.0)
	_shake_offset = Vector3(
		randf_range(-MAX_OFFSET, MAX_OFFSET) * _shake_strength,
		randf_range(-MAX_OFFSET * 0.35, MAX_OFFSET * 0.35) * _shake_strength,
		randf_range(-MAX_OFFSET, MAX_OFFSET) * _shake_strength,
	)


static func get_offset() -> Vector3:
	return _shake_offset


static func is_active() -> bool:
	return _shake_strength > 0.001


static func decay(delta: float) -> void:
	if _shake_strength <= 0.0:
		return
	var previous_strength := _shake_strength
	_shake_strength = maxf(0.0, _shake_strength - delta * DECAY_SPEED)
	if _shake_strength <= 0.001:
		_shake_strength = 0.0
		_shake_offset = Vector3.ZERO
		return
	if previous_strength > 0.0:
		_shake_offset *= _shake_strength / previous_strength
