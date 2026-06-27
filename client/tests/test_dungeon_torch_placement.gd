# Unit tests for dungeon torch placement.
# Run via: godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_dungeon_torch_placement.gd
extends SceneTree

const PlacementScript := preload("res://scripts/dungeon_torch_placement.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_perimeter_walls_produce_placements()
	_test_respects_max_count()
	_test_disabled_config_returns_empty()
	print("[gdtest] PASS: test_dungeon_torch_placement (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_perimeter_walls_produce_placements() -> void:
	var walls := [
		{"source": "perimeter", "kind": "wall", "position": {"x": 50, "y": -0.5}, "size": {"x": 100, "y": 1}},
		{"source": "perimeter", "kind": "wall", "position": {"x": 50, "y": 50.5}, "size": {"x": 100, "y": 1}},
		{"source": "perimeter", "kind": "wall", "position": {"x": -0.5, "y": 25}, "size": {"x": 1, "y": 50}},
		{"source": "perimeter", "kind": "wall", "position": {"x": 100.5, "y": 25}, "size": {"x": 1, "y": 50}},
	]
	var placements := PlacementScript.placements_from_walls(walls, {"enabled": true, "spacing": 10.0, "max_count": 8})
	_assert_true("perimeter walls produce torches", placements.size() >= 4)


func _test_respects_max_count() -> void:
	var walls := [
		{"source": "perimeter", "kind": "wall", "position": {"x": 50, "y": -0.5}, "size": {"x": 100, "y": 1}},
		{"source": "perimeter", "kind": "wall", "position": {"x": 50, "y": 50.5}, "size": {"x": 100, "y": 1}},
		{"source": "perimeter", "kind": "wall", "position": {"x": -0.5, "y": 25}, "size": {"x": 1, "y": 50}},
		{"source": "perimeter", "kind": "wall", "position": {"x": 100.5, "y": 25}, "size": {"x": 1, "y": 50}},
	]
	var placements := PlacementScript.placements_from_walls(walls, {"enabled": true, "spacing": 4.0, "max_count": 4})
	_assert_eq("max_count caps torch placements", placements.size(), 4)


func _test_disabled_config_returns_empty() -> void:
	var walls := [
		{"source": "perimeter", "kind": "wall", "position": {"x": 50, "y": -0.5}, "size": {"x": 100, "y": 1}},
	]
	var placements := PlacementScript.placements_from_walls(walls, {"enabled": false})
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
