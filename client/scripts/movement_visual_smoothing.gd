class_name MovementVisualSmoothing
extends RefCounted

const CATCH_UP_SPEED := 18.0
const MAX_OFFSET := 0.70
const RESET_DISTANCE := 1.50
const SETTLE_EPSILON := 0.01

var _has_anchor: bool = false
var _last_anchor_position := Vector3.ZERO


func reset(anchor: Node3D, visual: Node3D) -> void:
	_has_anchor = anchor != null
	_last_anchor_position = _anchor_position(anchor)
	if visual != null:
		visual.position.x = 0.0
		visual.position.z = 0.0


func preserve_after_anchor_move(anchor: Node3D, visual: Node3D) -> void:
	if anchor == null or visual == null:
		return
	var current := _anchor_position(anchor)
	if not _has_anchor:
		reset(anchor, visual)
		return
	var delta := _last_anchor_position - current
	_last_anchor_position = current
	if Vector2(delta.x, delta.z).length() > RESET_DISTANCE:
		reset(anchor, visual)
		return
	var offset := visual.position + Vector3(delta.x, 0.0, delta.z)
	offset.y = visual.position.y
	visual.position = _clamp_offset(offset)


func tick(delta: float, visual: Node3D) -> void:
	if visual == null:
		return
	var factor := clampf(1.0 - exp(-CATCH_UP_SPEED * maxf(delta, 0.0)), 0.0, 1.0)
	visual.position.x = lerpf(visual.position.x, 0.0, factor)
	visual.position.z = lerpf(visual.position.z, 0.0, factor)
	if Vector2(visual.position.x, visual.position.z).length() <= SETTLE_EPSILON:
		visual.position.x = 0.0
		visual.position.z = 0.0


func get_debug_state(visual: Node3D) -> Dictionary:
	var offset := Vector2.ZERO
	if visual != null:
		offset = Vector2(visual.position.x, visual.position.z)
	return {
		"active": offset.length() > SETTLE_EPSILON,
		"offset_length": offset.length(),
		"offset_x": offset.x,
		"offset_z": offset.y,
	}


func _clamp_offset(offset: Vector3) -> Vector3:
	var flat := Vector2(offset.x, offset.z)
	if flat.length() > MAX_OFFSET:
		flat = flat.normalized() * MAX_OFFSET
		offset.x = flat.x
		offset.z = flat.y
	return offset


func _anchor_position(anchor: Node3D) -> Vector3:
	return anchor.position if anchor != null else Vector3.ZERO
