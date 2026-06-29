# Headless unit tests for PerfPhaseTimer (v370).
extends SceneTree

const PerfPhaseTimerScript := preload("res://scripts/perf_phase_timer.gd")

var _pass := 0
var _fail := 0


func _init() -> void:
	OS.set_environment("ARPG_PERF_DEBUG", "1")
	PerfPhaseTimerScript.ensure_enabled()
	_test_disabled_without_env()
	_test_accumulates_phases()
	_test_format_snapshot()
	_finish()


func _test_disabled_without_env() -> void:
	OS.set_environment("ARPG_PERF_DEBUG", "0")
	PerfPhaseTimerScript._enabled = false
	PerfPhaseTimerScript.ensure_enabled()
	PerfPhaseTimerScript.reset_frame()
	PerfPhaseTimerScript.add_ms("delta", 1000)
	_assert_eq("disabled timer ignores samples", PerfPhaseTimerScript.snapshot_ms().size(), 0)
	OS.set_environment("ARPG_PERF_DEBUG", "1")
	PerfPhaseTimerScript._enabled = false
	PerfPhaseTimerScript.ensure_enabled()


func _test_accumulates_phases() -> void:
	PerfPhaseTimerScript.reset_frame()
	PerfPhaseTimerScript.add_ms("delta", 1500)
	PerfPhaseTimerScript.add_ms("delta", 500)
	PerfPhaseTimerScript.add_ms("fog", 2500)
	var snap: Dictionary = PerfPhaseTimerScript.snapshot_ms()
	_assert_eq("delta accumulates", snap.get("delta", 0.0), 2.0)
	_assert_eq("fog recorded", snap.get("fog", 0.0), 2.5)


func _test_format_snapshot() -> void:
	PerfPhaseTimerScript.reset_frame()
	PerfPhaseTimerScript.add_ms("entities", 1200)
	PerfPhaseTimerScript.add_ms("delta", 800)
	var formatted := PerfPhaseTimerScript.format_snapshot()
	_assert_true("format includes delta", formatted.find("delta=") >= 0)
	_assert_true("format includes entities", formatted.find("entities=") >= 0)
	var ranked := PerfPhaseTimerScript.format_snapshot(true)
	_assert_true("ranked puts largest phase first", ranked.begins_with("entities="))


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass += 1
		return
	_fail += 1
	push_error("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_true(label: String, value: bool) -> void:
	_assert_eq(label, value, true)


func _finish() -> void:
	if _fail == 0:
		print("[gdtest] PASS: test_perf_phase_timer (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_perf_phase_timer (%d failures, %d assertions)" % [_fail, _pass])
		quit(1)
