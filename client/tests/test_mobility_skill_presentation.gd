# Unit tests for mobility skill presentation (v353).
extends SceneTree

const MobilitySkillPresentationScript := preload("res://scripts/mobility_skill_presentation.gd")
const MovementPresentationLoaderScript := preload("res://scripts/movement_presentation_loader.gd")

var _pass_count: int = 0
var _fail_count: int = 0
var _finished_entity_id: String = ""


func _initialize() -> void:
	_test_leap_marks_entity_active()
	_test_clear_entity_clears_active()
	_test_teleport_starts_active()
	print("[gdtest] PASS: test_mobility_skill_presentation (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_leap_marks_entity_active() -> void:
	MovementPresentationLoaderScript.reset_for_tests()
	var owner := Node.new()
	get_root().add_child(owner)
	var presentation := MobilitySkillPresentationScript.new()
	presentation.bind_owner(owner)
	var anchor := Node3D.new()
	owner.add_child(anchor)
	presentation.play_from_skill_cast(
		"p1",
		anchor,
		{"skill_id": "leap", "position": {"x": 0.0, "y": 0.0}},
		Vector3(2.0, 0.0, 0.0),
	)
	_assert_true("leap active", presentation.is_active("p1"))
	var debug := presentation.get_debug_state()
	_assert_true("debug active", bool(debug.get("active", false)))
	_assert_eq("debug skill", str(debug.get("skill_id", "")), "leap")
	owner.queue_free()
	MovementPresentationLoaderScript.reset_for_tests()


func _test_clear_entity_clears_active() -> void:
	MovementPresentationLoaderScript.reset_for_tests()
	var owner := Node.new()
	get_root().add_child(owner)
	var presentation := MobilitySkillPresentationScript.new()
	presentation.bind_owner(owner)
	var anchor := Node3D.new()
	owner.add_child(anchor)
	presentation.play_from_skill_cast(
		"p1",
		anchor,
		{"skill_id": "charge", "position": {"x": 0.0, "y": 0.0}},
		Vector3(1.0, 0.0, 0.0),
	)
	presentation.clear_entity("p1")
	_assert_false("cleared inactive", presentation.is_active("p1"))
	owner.queue_free()
	MovementPresentationLoaderScript.reset_for_tests()


func _test_teleport_starts_active() -> void:
	MovementPresentationLoaderScript.reset_for_tests()
	var owner := Node.new()
	get_root().add_child(owner)
	var presentation := MobilitySkillPresentationScript.new()
	presentation.bind_owner(owner)
	var anchor := Node3D.new()
	owner.add_child(anchor)
	presentation.play_from_skill_cast(
		"p2",
		anchor,
		{"skill_id": "teleport", "position": {"x": 1.0, "y": 2.0}},
		Vector3(4.0, 0.0, 3.0),
	)
	_assert_true("teleport active", presentation.is_active("p2"))
	owner.queue_free()
	MovementPresentationLoaderScript.reset_for_tests()


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


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)
