extends SceneTree

const WallOcclusionFadeScript := preload("res://scripts/wall_occlusion_fade.gd")
const WallRendererScript := preload("res://scripts/wall_renderer.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")
const WallOcclusionPresentationLoaderScript := preload("res://scripts/wall_occlusion_presentation_loader.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_test_segment_intersects_wall_between_camera_and_hero()
	_test_non_box_wall_kind_skipped()
	_test_wall_qualifies_for_box_and_wood()
	_test_perimeter_wall_skipped_for_occlusion_fade()
	_test_resolve_faded_walls_for_lab_layout()
	_test_backdrop_walls_stay_opaque_in_lab_layout()
	_test_faded_wall_disables_pick_collision()
	_test_sync_throttles_rebuilds()
	print("[gdtest] PASS: test_wall_occlusion_fade (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_segment_intersects_wall_between_camera_and_hero() -> void:
	var camera_xz := Vector2(11.0, 20.0)
	var hero_xz := Vector2(2.0, 5.0)
	var center := Vector2(8.0, 9.5)
	var size := Vector2(16.0, 1.0)
	_assert_true(
		"north wall blocks camera to hero",
		WallOcclusionFadeScript.segment_intersects_inflated_aabb(camera_xz, hero_xz, center, size, 0.02),
	)


func _test_non_box_wall_kind_skipped() -> void:
	var faded := WallOcclusionFadeScript.resolve_faded_walls(
		Vector2(11.0, 20.0),
		[Vector3(2.0, 0.0, 5.0)],
		[
			{
				"id": "rock_1",
				"kind": "rock",
				"position": {"x": 8.0, "y": 9.5},
				"size": {"x": 16.0, "y": 1.0},
			},
		],
	)
	_assert_eq("rock wall skipped", faded.size(), 0)


func _test_wall_qualifies_for_box_and_wood() -> void:
	_assert_true("default wall qualifies", WallOcclusionFadeScript.wall_qualifies_for_occlusion_fade({}))
	_assert_true("wood wall qualifies", WallOcclusionFadeScript.wall_qualifies_for_occlusion_fade({"kind": "wood"}))
	_assert_false("water skipped", WallOcclusionFadeScript.wall_qualifies_for_occlusion_fade({"kind": "water"}))


func _test_perimeter_wall_skipped_for_occlusion_fade() -> void:
	_assert_false(
		"perimeter wall skipped",
		WallOcclusionFadeScript.wall_qualifies_for_occlusion_fade({"kind": "wall", "source": "perimeter"}),
	)
	var faded := WallOcclusionFadeScript.resolve_faded_walls(
		Vector2(11.0, 20.0),
		[Vector3(8.0, 0.0, 5.0)],
		[
			{
				"id": "south_perimeter",
				"kind": "wall",
				"source": "perimeter",
				"position": {"x": 8.0, "y": -0.5},
				"size": {"x": 16.0, "y": 1.0},
			},
		],
	)
	_assert_eq("perimeter wall not faded", faded.size(), 0)


func _test_resolve_faded_walls_for_lab_layout() -> void:
	var faded := WallOcclusionFadeScript.resolve_faded_walls(
		Vector2(11.0, 20.0),
		[Vector3(2.0, 0.0, 5.0)],
		[
			{
				"id": "north_wall",
				"kind": "wall",
				"position": {"x": 8.0, "y": 9.5},
				"size": {"x": 16.0, "y": 1.0},
			},
		],
	)
	_assert_eq("one faded wall", faded.size(), 1)
	_assert_true("north wall faded", faded.has("north_wall"))


func _test_backdrop_walls_stay_opaque_in_lab_layout() -> void:
	var faded := WallOcclusionFadeScript.resolve_faded_walls(
		Vector2(11.0, 20.0),
		[Vector3(8.0, 0.0, 5.0)],
		[
			{
				"id": "south_wall",
				"kind": "wall",
				"position": {"x": 8.0, "y": 0.5},
				"size": {"x": 16.0, "y": 1.0},
			},
			{
				"id": "north_wall",
				"kind": "wall",
				"position": {"x": 8.0, "y": 9.5},
				"size": {"x": 16.0, "y": 1.0},
			},
			{
				"id": "west_wall",
				"kind": "wall",
				"position": {"x": 0.5, "y": 5.0},
				"size": {"x": 1.0, "y": 10.0},
			},
			{
				"id": "east_wall",
				"kind": "wall",
				"position": {"x": 15.5, "y": 5.0},
				"size": {"x": 1.0, "y": 10.0},
			},
		],
	)
	_assert_true("north wall faded for centered hero", faded.has("north_wall"))
	_assert_false("south backdrop wall stays opaque", faded.has("south_wall"))
	_assert_false("west backdrop wall stays opaque", faded.has("west_wall"))


func _test_faded_wall_disables_pick_collision() -> void:
	var root := Node3D.new()
	get_root().add_child(root)
	var renderer = WallRendererScript.new(root, GroundWallFactoryScript.new())
	renderer.set_level(-4)
	renderer.render_wall_layout([{
		"id": "fade_wall",
		"position": {"x": 4.0, "y": 4.0},
		"size": {"x": 2.0, "y": 1.0},
		"source": "generated",
	}])
	var body := root.get_child(0) as StaticBody3D
	var shape := body.get_child(0) as CollisionShape3D
	_assert_false("wall collision starts enabled", shape.disabled)
	renderer.apply_occlusion_fades({"fade_wall": 0.34})
	_assert_true("faded wall disables collision", shape.disabled)
	_assert_false("faded wall is not ray pickable", body.input_ray_pickable)
	renderer.apply_occlusion_fades({})
	_assert_false("opaque wall restores collision", shape.disabled)
	_assert_true("opaque wall is ray pickable", body.input_ray_pickable)
	root.queue_free()


class FakeWallRenderer:
	extends RefCounted
	var apply_calls := 0

	func apply_occlusion_fades(_faded: Dictionary) -> void:
		apply_calls += 1


func _test_sync_throttles_rebuilds() -> void:
	var renderer := FakeWallRenderer.new()
	var fade = WallOcclusionFadeScript.new(renderer)
	var camera := Camera3D.new()
	get_root().add_child(camera)
	camera.global_position = Vector3(11.0, 10.0, 20.0)
	var walls := [{"id": "w1", "position": {"x": 8.0, "y": 9.5}, "size": {"x": 16.0, "y": 1.0}}]
	var min_frames := WallOcclusionPresentationLoaderScript.min_rebuild_interval_frames()
	fade.sync(camera, walls, [Vector3(2.0, 0.0, 5.0)], true)
	_assert_eq("first sync rebuilds immediately", renderer.apply_calls, 1)
	for i in range(min_frames * 4):
		fade.sync(camera, walls, [Vector3(2.0, 0.0, 5.0)], true)
	_assert_eq("unchanged state never rebuilds", renderer.apply_calls, 1)
	var frames := min_frames * 6
	for i in range(frames):
		fade.sync(camera, walls, [Vector3(2.0 + 0.1 * float(i + 1), 0.0, 5.0)], true)
	_assert_true("moving target rebuilds at most once per interval", renderer.apply_calls - 1 <= int(ceil(float(frames) / float(min_frames))))
	_assert_true("moving target still rebuilds", renderer.apply_calls > 1)
	var before_inactive: int = renderer.apply_calls
	for i in range(5):
		fade.sync(camera, walls, [], true)
	_assert_eq("inactive clears fades once, not every frame", renderer.apply_calls, before_inactive + 1)
	fade.sync(camera, walls, [Vector3(2.0, 0.0, 5.0)], true)
	_assert_eq("reactivation rebuilds immediately", renderer.apply_calls, before_inactive + 2)
	camera.free()


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("FAIL: %s" % label)


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("FAIL: %s got=%s want=%s" % [label, str(got), str(want)])


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)
