# Unit tests for dungeon torch placement.
# Run via: godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_dungeon_torch_placement.gd
extends SceneTree

const PlacementScript := preload("res://scripts/dungeon_torch_placement.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_wall_segments_spawn_bounded_torches()
	_test_deterministic_per_level()
	_test_respects_shader_cap()
	_test_disabled_config_returns_empty()
	print("[gdtest] PASS: test_dungeon_torch_placement (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_wall_segments_spawn_bounded_torches() -> void:
	var walls := [
		{"id": "north", "source": "perimeter", "kind": "wall", "position": {"x": 50, "y": -0.5}, "size": {"x": 100, "y": 1}},
		{"id": "gen_a", "source": "generated", "kind": "wall", "position": {"x": 20, "y": 20}, "size": {"x": 12, "y": 1}},
	]
	var cfg := {
		"enabled": true,
		"wall_segment_tiles": 10.0,
		"torches_per_segment_min": 0,
		"torches_per_segment_max": 2,
		"max_shader_torches": 32,
	}
	var placements := PlacementScript.placements_from_walls(walls, cfg, -1)
	_assert_true("solid walls can spawn torches", placements.size() >= 1)
	_assert_true("torch count stays within segment budget", placements.size() <= 24)


func _test_deterministic_per_level() -> void:
	var walls := [
		{"id": "north", "source": "perimeter", "kind": "wall", "position": {"x": 50, "y": -0.5}, "size": {"x": 40, "y": 1}},
	]
	var cfg := {
		"enabled": true,
		"wall_segment_tiles": 10.0,
		"torches_per_segment_min": 0,
		"torches_per_segment_max": 2,
		"max_shader_torches": 32,
	}
	var first := PlacementScript.placements_from_walls(walls, cfg, -1)
	var second := PlacementScript.placements_from_walls(walls, cfg, -1)
	var other_level := PlacementScript.placements_from_walls(walls, cfg, -2)
	_assert_eq("same level yields stable torch layout", str(first), str(second))
	_assert_true("level changes can change torch layout", str(first) != str(other_level) or first.is_empty())


func _test_respects_shader_cap() -> void:
	var walls := [
		{"id": "north", "source": "perimeter", "kind": "wall", "position": {"x": 50, "y": -0.5}, "size": {"x": 200, "y": 1}},
	]
	var placements := PlacementScript.placements_from_walls(
		walls,
		{
			"enabled": true,
			"wall_segment_tiles": 5.0,
			"torches_per_segment_min": 2,
			"torches_per_segment_max": 2,
			"max_shader_torches": 6,
		},
		-1,
	)
	_assert_eq("shader cap limits torch placements", placements.size(), 6)


func _test_disabled_config_returns_empty() -> void:
	var walls := [
		{"id": "north", "source": "perimeter", "kind": "wall", "position": {"x": 50, "y": -0.5}, "size": {"x": 100, "y": 1}},
	]
	var placements := PlacementScript.placements_from_walls(walls, {"enabled": false}, -1)
	_assert_eq("disabled config yields no torches", placements.size(), 0)


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
