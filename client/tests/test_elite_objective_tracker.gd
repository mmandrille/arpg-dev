# Unit test for the elite objective HUD tracker.
extends SceneTree

const EliteObjectiveTrackerScript := preload("res://scripts/elite_objective_tracker.gd")

var _failures: int = 0


func _initialize() -> void:
	var tracker: EliteObjectiveTracker = EliteObjectiveTrackerScript.new()
	get_root().add_child(tracker)
	tracker._build()
	_assert_true("initial hidden", not bool(tracker.get_debug_state().get("visible", true)))
	tracker.set_state({"visible": true, "status": "active", "remaining_leaders": 2})
	var active := tracker.get_debug_state()
	_assert_eq("active status", str(active.get("status", "")), "active")
	_assert_eq("active remaining", int(active.get("remaining_leaders", -1)), 2)
	_assert_true("active detail", "2 remaining" in str(active.get("detail", "")))
	tracker.set_state({"visible": true, "status": "claim", "remaining_leaders": 0})
	_assert_true("claim text", "Claim" in str(tracker.get_debug_state().get("detail", "")))
	tracker.set_state({"visible": true, "status": "complete", "remaining_leaders": 0})
	_assert_eq("complete status", str(tracker.get_debug_state().get("status", "")), "complete")
	tracker.set_state({"visible": false, "status": "hidden", "remaining_leaders": 0})
	_assert_true("hidden state", not bool(tracker.get_debug_state().get("visible", true)))
	tracker.free()
	if _failures > 0:
		quit(1)
	print("[gdtest] PASS: test_elite_objective_tracker")
	quit(0)


func _assert_true(label: String, value: bool) -> void:
	if not value:
		_failures += 1
		printerr("[gdtest] FAIL %s" % label)


func _assert_eq(label: String, got, want) -> void:
	if got != want:
		_failures += 1
		printerr("[gdtest] FAIL %s got=%s want=%s" % [label, str(got), str(want)])
