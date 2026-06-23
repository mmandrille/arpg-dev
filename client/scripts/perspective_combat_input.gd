## PerspectiveCombatInput — computes aim direction for perspective camera modes.
##
## Unlike isometric (mouse cursor → ground pick), perspective modes aim from
## the viewport center so the player shoots where the camera is looking.
class_name PerspectiveCombatInput
extends RefCounted


## Returns a normalized flat (XZ) aim direction for perspective camera modes.
## Casts a ray from the viewport center and projects to the ground plane at the
## player's Y level. Falls back to flattening the camera's forward vector when
## the ray is nearly horizontal.
static func flat_aim_direction(camera: Camera3D, player_anchor: Node3D) -> Vector2:
	if camera == null or player_anchor == null:
		return Vector2(1.0, 0.0)
	var vp := camera.get_viewport()
	if vp == null:
		return Vector2(1.0, 0.0)
	var center := vp.get_visible_rect().get_center()
	var from := camera.project_ray_origin(center)
	var dir := camera.project_ray_normal(center)
	# Project to ground plane at player Y.
	if abs(dir.y) > 0.001:
		var t := (player_anchor.global_position.y - from.y) / dir.y
		if t > 0.0:
			var hit := from + dir * t
			var flat := hit - player_anchor.global_position
			if flat.length_squared() > 0.0001:
				return Vector2(flat.x, flat.z).normalized()
	# Fallback: flatten camera forward to XZ plane.
	var fwd := -camera.global_transform.basis.z
	var fwd2 := Vector2(fwd.x, fwd.z)
	if fwd2.length_squared() > 0.0001:
		return fwd2.normalized()
	return Vector2(1.0, 0.0)
