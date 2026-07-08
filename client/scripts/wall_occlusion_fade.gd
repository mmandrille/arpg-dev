## Fades box wall meshes that block camera→entity sight lines on the XZ plane.
class_name WallOcclusionFade
extends RefCounted

const WallOcclusionPresentationLoaderScript := preload("res://scripts/wall_occlusion_presentation_loader.gd")

var _wall_renderer  # WallRenderer
var _debug: Dictionary = {}
var _last_camera_xz := Vector2(-99999.0, -99999.0)
var _last_target_signature := ""
var _last_wall_signature := ""
var _frames_since_rebuild := 0


func _init(wall_renderer) -> void:
	_wall_renderer = wall_renderer
	_reset_debug()


func sync(camera: Camera3D, wall_layout: Array, targets: Array, active: bool) -> void:
	if _wall_renderer == null:
		return
	if not active or camera == null or wall_layout.is_empty() or targets.is_empty():
		_wall_renderer.apply_occlusion_fades({})
		_set_debug(0, 1.0, 0, false)
		return
	var camera_xz := Vector2(camera.global_position.x, camera.global_position.z)
	var target_signature := _target_signature(targets)
	var wall_signature := _wall_signature(wall_layout)
	_frames_since_rebuild += 1
	var epsilon := WallOcclusionPresentationLoaderScript.move_epsilon()
	var min_frames := WallOcclusionPresentationLoaderScript.min_rebuild_interval_frames()
	if (
		_frames_since_rebuild < min_frames
		and camera_xz.distance_to(_last_camera_xz) <= epsilon
		and target_signature == _last_target_signature
		and wall_signature == _last_wall_signature
	):
		return
	_last_camera_xz = camera_xz
	_last_target_signature = target_signature
	_last_wall_signature = wall_signature
	_frames_since_rebuild = 0
	var faded := resolve_faded_walls(camera_xz, targets, wall_layout)
	_wall_renderer.apply_occlusion_fades(faded)
	var min_alpha := 1.0
	for wall_id in faded.keys():
		min_alpha = minf(min_alpha, float(faded[wall_id]))
	_set_debug(faded.size(), min_alpha, targets.size(), true)


func get_debug_state() -> Dictionary:
	return _debug.duplicate(true)


static func resolve_faded_walls(camera_xz: Vector2, targets: Array, wall_layout: Array) -> Dictionary:
	var faded: Dictionary = {}
	var inflate := WallOcclusionPresentationLoaderScript.segment_inflate()
	var faded_alpha := WallOcclusionPresentationLoaderScript.faded_alpha()
	for target in targets:
		if typeof(target) != TYPE_VECTOR3:
			continue
		var target_xz := Vector2(float(target.x), float(target.z))
		for wall in wall_layout:
			if typeof(wall) != TYPE_DICTIONARY:
				continue
			if not wall_qualifies_for_occlusion_fade(wall as Dictionary):
				continue
			var wall_id := str((wall as Dictionary).get("id", ""))
			if wall_id == "" or faded.has(wall_id):
				continue
			var pos: Dictionary = (wall as Dictionary).get("position", {})
			var size: Dictionary = (wall as Dictionary).get("size", {})
			var center := Vector2(float(pos.get("x", 0.0)), float(pos.get("y", 0.0)))
			var extents := Vector2(float(size.get("x", 0.0)), float(size.get("y", 0.0)))
			if segment_intersects_inflated_aabb(camera_xz, target_xz, center, extents, inflate):
				faded[wall_id] = faded_alpha

	return faded


static func wall_qualifies_for_occlusion_fade(wall: Dictionary) -> bool:
	if str(wall.get("source", "")) == "perimeter":
		return false
	var kind := str(wall.get("kind", "wall"))

	return kind == "" or kind == "wall" or kind == "wood"


static func segment_intersects_inflated_aabb(
	start: Vector2,
	end: Vector2,
	center: Vector2,
	size: Vector2,
	inflate: float,
) -> bool:
	if size.x <= 0.0 or size.y <= 0.0:
		return false
	var half := size * 0.5 + Vector2(inflate, inflate)
	var min_v := center - half
	var max_v := center + half
	var dx := end.x - start.x
	var dy := end.y - start.y
	var clip_x := _clip_segment_axis(start.x, dx, min_v.x, max_v.x, 0.0, 1.0)
	if clip_x.is_empty():
		return false
	var tmin := float(clip_x.get("tmin", 0.0))
	var tmax := float(clip_x.get("tmax", 1.0))
	var clip_y := _clip_segment_axis(start.y, dy, min_v.y, max_v.y, tmin, tmax)
	if clip_y.is_empty():
		return false
	tmin = float(clip_y.get("tmin", 0.0))
	tmax = float(clip_y.get("tmax", 1.0))

	return tmin <= tmax and tmax >= 0.0 and tmin <= 1.0


static func _clip_segment_axis(start: float, delta: float, min_v: float, max_v: float, tmin: float, tmax: float) -> Dictionary:
	if abs(delta) < 1e-12:
		if start < min_v or start > max_v:
			return {}
		return {"tmin": tmin, "tmax": tmax}
	var inv := 1.0 / delta
	var t1 := (min_v - start) * inv
	var t2 := (max_v - start) * inv
	if t1 > t2:
		var swap := t1
		t1 = t2
		t2 = swap
	tmin = maxf(tmin, t1)
	tmax = minf(tmax, t2)
	if tmin > tmax or tmax < 0.0 or tmin > 1.0:
		return {}

	return {"tmin": tmin, "tmax": tmax}


func reset() -> void:
	_last_camera_xz = Vector2(-99999.0, -99999.0)
	_last_target_signature = ""
	_last_wall_signature = ""
	_frames_since_rebuild = 0
	if _wall_renderer != null:
		_wall_renderer.apply_occlusion_fades({})
	_reset_debug()


func _target_signature(targets: Array) -> String:
	var parts: PackedStringArray = []
	for target in targets:
		if typeof(target) != TYPE_VECTOR3:
			continue
		var xz := Vector2(float(target.x), float(target.z))
		parts.append("%.2f,%.2f" % [xz.x, xz.y])

	return "|".join(parts)


func _wall_signature(wall_layout: Array) -> String:
	var parts: PackedStringArray = []
	for wall in wall_layout:
		if typeof(wall) != TYPE_DICTIONARY:
			continue
		parts.append(str((wall as Dictionary).get("id", "")))

	return "|".join(parts)


func _set_debug(faded_count: int, min_faded_alpha: float, target_count: int, active: bool) -> void:
	_debug = {
		"active": active,
		"faded_wall_count": faded_count,
		"min_faded_alpha": min_faded_alpha,
		"target_count": target_count,
		"faded_alpha": WallOcclusionPresentationLoaderScript.faded_alpha(),
		"opaque_alpha": WallOcclusionPresentationLoaderScript.opaque_alpha(),
	}


func _reset_debug() -> void:
	_set_debug(0, 1.0, 0, false)
