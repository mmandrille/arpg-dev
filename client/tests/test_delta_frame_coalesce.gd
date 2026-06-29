# Headless unit tests for DeltaFrameCoalesce (v377).
extends SceneTree

const DeltaFrameCoalesceScript := preload("res://scripts/delta_frame_coalesce.gd")

var _pass := 0
var _fail := 0


func _init() -> void:
	_test_merges_events_and_changes()
	_test_last_performance_payload_wins()
	_finish()


func _test_merges_events_and_changes() -> void:
	var merged := DeltaFrameCoalesceScript.merge_pending([
		{"events": [{"event_type": "a"}], "changes": [{"op": "gold_update"}]},
		{"events": [{"event_type": "b"}], "changes": [{"op": "inventory_add"}]},
	])
	_assert_eq("events merged", (merged.get("events", []) as Array).size(), 2)
	_assert_eq("changes merged", (merged.get("changes", []) as Array).size(), 2)


func _test_last_performance_payload_wins() -> void:
	var merged := DeltaFrameCoalesceScript.merge_pending([
		{"performance": {"live_monsters": 1}},
		{"performance": {"live_monsters": 9}},
	])
	var perf: Dictionary = merged.get("performance", {})
	_assert_eq("performance last-wins", int(perf.get("live_monsters", 0)), 9)


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass += 1
	else:
		_fail += 1
		push_error("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])


func _finish() -> void:
	if _fail == 0:
		print("[gdtest] PASS: test_delta_frame_coalesce (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_delta_frame_coalesce (%d failures)" % _fail)
		quit(1)
