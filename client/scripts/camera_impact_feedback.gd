class_name CameraImpactFeedback
extends RefCounted

const MAX_OFFSET := 0.22
const DECAY_SPEED := 14.0

static var _shake_strength: float = 0.0


static func apply_from_damage(camera: Camera3D, damage: int, max_hp: int) -> void:
	if camera == null or damage <= 0:
		return
	var ratio := float(damage) / float(maxi(1, max_hp))
	_shake_strength = clampf(_shake_strength + ratio * 0.35, 0.0, 1.0)
	var offset := Vector3(
		randf_range(-MAX_OFFSET, MAX_OFFSET) * _shake_strength,
		randf_range(-MAX_OFFSET * 0.35, MAX_OFFSET * 0.35) * _shake_strength,
		randf_range(-MAX_OFFSET, MAX_OFFSET) * _shake_strength,
	)
	camera.position = ClientConstants.CAMERA_FOLLOW_OFFSET + offset


static func decay(camera: Camera3D, delta: float) -> void:
	if camera == null or _shake_strength <= 0.0:
		return
	_shake_strength = maxf(0.0, _shake_strength - delta * DECAY_SPEED)
	if _shake_strength <= 0.001:
		_shake_strength = 0.0
		camera.position = ClientConstants.CAMERA_FOLLOW_OFFSET
