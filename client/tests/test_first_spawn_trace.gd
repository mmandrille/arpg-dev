extends SceneTree

const SamplerScript := preload("res://scripts/perf_debug_sampler.gd")

var failures := 0


func _init() -> void:
	OS.set_environment("ARPG_FIRST_SPAWN_TRACE", "1")
	var sampler = SamplerScript.new()
	_check("trace enabled", sampler.trace_frames)
	sampler.begin_frame()
	sampler.sample(0.016, 1, 1, 0.0, {}, [], "balanced")
	_check("no spawn before monsters", not sampler._saw_monsters)
	sampler.begin_frame()
	sampler.sample(0.016, 1, 2, 0.0, {"m": {"type": "monster"}}, ["m"], "balanced")
	_check("first spawn detected", sampler._saw_monsters)
	_check("each frame recorded", sampler._frame_index == 2)
	OS.set_environment("ARPG_FIRST_SPAWN_TRACE", "0")
	if failures == 0:
		print("[gdtest] PASS: test_first_spawn_trace")
		quit(0)
	else:
		print("[gdtest] FAIL: test_first_spawn_trace (%d failures)" % failures)
		quit(1)


func _check(label: String, condition: bool) -> void:
	if not condition:
		failures += 1
		push_error("[gdtest] FAIL: " + label)
