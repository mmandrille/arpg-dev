## FogLosShadowCache caches LOS shadow polygon builds across frames.
class_name FogLosShadowCache
extends RefCounted

const HeroVisibilityFieldScript := preload("res://scripts/hero_visibility_field.gd")
const FogPresentationLoaderScript := preload("res://scripts/fog_presentation_loader.gd")

var rebuild_count: int = 0
var cache_hits: int = 0
var last_rebuild_reason: String = ""

var _built: Dictionary = {}
var _valid: bool = false
var _hard_invalidate_pending: bool = true
var _performance_throttle_frames: int = 0
var _frames_since_rebuild: int = 0
var _cached_hero_world := Vector2.ZERO
var _cached_center_px := Vector2.ZERO
var _cached_viewport_size := Vector2.ZERO


func hard_invalidate() -> void:
	_hard_invalidate_pending = true
	_valid = false


func set_performance_throttle_frames(frames: int) -> void:
	_performance_throttle_frames = maxi(0, frames)


func tick_frame() -> void:
	if _valid:
		_frames_since_rebuild += 1


func is_valid() -> bool:
	return _valid


func get_built() -> Dictionary:
	return _built


func resolve(
	camera: Camera3D,
	hero_world: Vector2,
	shadow_reach: float,
	viewport_size: Vector2,
	occluders: Array,
	shadow_cfg: Dictionary,
	center_px: Vector2,
) -> Dictionary:
	var cache_cfg: Dictionary = FogPresentationLoaderScript.shadow_cache()
	var move_epsilon := float(cache_cfg.get("move_epsilon", 0.006))
	var viewport_epsilon := float(cache_cfg.get("viewport_size_epsilon_px", 1.0))
	var reason := ""

	if _hard_invalidate_pending or not _valid:
		reason = "hard_invalidate" if _hard_invalidate_pending else "initial"
	elif hero_world.distance_to(_cached_hero_world) > move_epsilon:
		reason = "hero_move"
	elif center_px.distance_to(_cached_center_px) > move_epsilon:
		reason = "center_move"
	elif (
		absf(viewport_size.x - _cached_viewport_size.x) > viewport_epsilon
		or absf(viewport_size.y - _cached_viewport_size.y) > viewport_epsilon
	):
		reason = "viewport"
	else:
		cache_hits += 1
		return _built

	var soft_reason := reason in ["hero_move", "center_move", "viewport"]
	if (
		soft_reason
		and _performance_throttle_frames > 0
		and _valid
		and _frames_since_rebuild < _performance_throttle_frames
	):
		cache_hits += 1
		return _built

	_built = HeroVisibilityFieldScript.build_shadow_polygons(
		camera,
		hero_world,
		shadow_reach,
		viewport_size,
		occluders,
		shadow_cfg,
		center_px,
	)
	_valid = true
	_hard_invalidate_pending = false
	_cached_hero_world = hero_world
	_cached_center_px = center_px
	_cached_viewport_size = viewport_size
	_frames_since_rebuild = 0
	rebuild_count += 1
	last_rebuild_reason = reason if reason != "" else "rebuild"

	return _built
