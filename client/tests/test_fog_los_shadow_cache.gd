extends SceneTree

const FogLosShadowCacheScript := preload("res://scripts/fog_los_shadow_cache.gd")
const FogPresentationLoaderScript := preload("res://scripts/fog_presentation_loader.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	FogPresentationLoaderScript.reset_for_tests()
	FogPresentationLoaderScript.ensure_loaded()
	await _test_consecutive_resolve_hits_cache()
	await _test_layout_hard_invalidate_rebuilds()
	await _test_hero_move_within_epsilon_hits()
	await _test_hero_move_beyond_epsilon_rebuilds()
	await _test_performance_throttle_defers_soft_rebuild()
	print("[gdtest] PASS: test_fog_los_shadow_cache (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_consecutive_resolve_hits_cache() -> void:
	var cache := FogLosShadowCacheScript.new()
	var occluders := [{"position": {"x": 3.0, "y": 0.0}, "size": {"x": 1.0, "y": 3.0}}]
	var hero_world := Vector2(0.0, 0.0)
	var center_px := Vector2(640.0, 360.0)
	var viewport_size := Vector2(1280.0, 720.0)
	var shadow_cfg: Dictionary = FogPresentationLoaderScript.shadow()
	cache.resolve(null, hero_world, 12.0, viewport_size, occluders, shadow_cfg, center_px)
	cache.resolve(null, hero_world, 12.0, viewport_size, occluders, shadow_cfg, center_px)
	_assert_eq("static resolve rebuild count", cache.rebuild_count, 1)
	_assert_true("static resolve cache hit", cache.cache_hits >= 1)
	_assert_true("cache valid after resolve", cache.is_valid())


func _test_layout_hard_invalidate_rebuilds() -> void:
	var cache := FogLosShadowCacheScript.new()
	var occluders_a := [{"position": {"x": 3.0, "y": 0.0}, "size": {"x": 1.0, "y": 3.0}}]
	var occluders_b := [{"position": {"x": 4.0, "y": 0.0}, "size": {"x": 1.0, "y": 3.0}}]
	var hero_world := Vector2(0.0, 0.0)
	var center_px := Vector2(640.0, 360.0)
	var viewport_size := Vector2(1280.0, 720.0)
	var shadow_cfg: Dictionary = FogPresentationLoaderScript.shadow()
	cache.resolve(null, hero_world, 12.0, viewport_size, occluders_a, shadow_cfg, center_px)
	cache.hard_invalidate()
	cache.resolve(null, hero_world, 12.0, viewport_size, occluders_b, shadow_cfg, center_px)
	_assert_eq("layout invalidate rebuild count", cache.rebuild_count, 2)
	_assert_eq("layout invalidate reason", cache.last_rebuild_reason, "hard_invalidate")


func _test_hero_move_within_epsilon_hits() -> void:
	var cache := FogLosShadowCacheScript.new()
	var occluders := [{"position": {"x": 3.0, "y": 0.0}, "size": {"x": 1.0, "y": 3.0}}]
	var hero_world := Vector2(0.0, 0.0)
	var nudged := Vector2(0.003, 0.0)
	var center_px := Vector2(640.0, 360.0)
	var viewport_size := Vector2(1280.0, 720.0)
	var shadow_cfg: Dictionary = FogPresentationLoaderScript.shadow()
	cache.resolve(null, hero_world, 12.0, viewport_size, occluders, shadow_cfg, center_px)
	cache.resolve(null, nudged, 12.0, viewport_size, occluders, shadow_cfg, center_px)
	_assert_eq("epsilon hero move rebuild count", cache.rebuild_count, 1)
	_assert_true("epsilon hero move cache hit", cache.cache_hits >= 1)


func _test_hero_move_beyond_epsilon_rebuilds() -> void:
	var cache := FogLosShadowCacheScript.new()
	var occluders := [{"position": {"x": 3.0, "y": 0.0}, "size": {"x": 1.0, "y": 3.0}}]
	var hero_world := Vector2(0.0, 0.0)
	var moved := Vector2(0.05, 0.0)
	var center_px := Vector2(640.0, 360.0)
	var viewport_size := Vector2(1280.0, 720.0)
	var shadow_cfg: Dictionary = FogPresentationLoaderScript.shadow()
	cache.resolve(null, hero_world, 12.0, viewport_size, occluders, shadow_cfg, center_px)
	cache.resolve(null, moved, 12.0, viewport_size, occluders, shadow_cfg, center_px)
	_assert_eq("hero move rebuild count", cache.rebuild_count, 2)
	_assert_eq("hero move reason", cache.last_rebuild_reason, "hero_move")


func _test_performance_throttle_defers_soft_rebuild() -> void:
	var cache := FogLosShadowCacheScript.new()
	cache.set_performance_throttle_frames(3)
	var occluders := [{"position": {"x": 3.0, "y": 0.0}, "size": {"x": 1.0, "y": 3.0}}]
	var hero_world := Vector2(0.0, 0.0)
	var moved := Vector2(0.05, 0.0)
	var center_px := Vector2(640.0, 360.0)
	var viewport_size := Vector2(1280.0, 720.0)
	var shadow_cfg: Dictionary = FogPresentationLoaderScript.shadow()
	cache.resolve(null, hero_world, 12.0, viewport_size, occluders, shadow_cfg, center_px)
	cache.tick_frame()
	cache.resolve(null, moved, 12.0, viewport_size, occluders, shadow_cfg, center_px)
	_assert_eq("throttled soft move rebuild count", cache.rebuild_count, 1)
	_assert_true("throttled soft move cache hit", cache.cache_hits >= 1)
	cache.tick_frame()
	cache.tick_frame()
	cache.tick_frame()
	cache.resolve(null, moved + Vector2(0.05, 0.0), 12.0, viewport_size, occluders, shadow_cfg, center_px)
	_assert_eq("throttle expired rebuild count", cache.rebuild_count, 2)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
