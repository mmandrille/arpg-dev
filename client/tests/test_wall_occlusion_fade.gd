extends SceneTree

const WallOcclusionFadeScript := preload("res://scripts/wall_occlusion_fade.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_test_segment_intersects_wall_between_camera_and_hero()
	_test_non_box_wall_kind_skipped()
	_test_wall_qualifies_for_box_and_wood()
	_test_resolve_faded_walls_for_lab_layout()
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
