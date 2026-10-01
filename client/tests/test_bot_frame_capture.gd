extends SceneTree

const CaptureScript := preload("res://scripts/bot_frame_capture.gd")
const StepCatalogScript := preload("res://scripts/bot_step_catalog.gd")


func _initialize() -> void:
	for name in ["v498_before_depth1", "v498-after-depth8", "A0"]:
		if not CaptureScript.valid_name(name):
			_fail("valid capture name rejected: %s" % name)
			return
	for name in ["", "../escape", "nested/file", "nested\\file", "name.png", "a".repeat(65)]:
		if CaptureScript.valid_name(name):
			_fail("unsafe capture name accepted: %s" % name)
			return
	if not CaptureScript.output_path("v498_depth1").ends_with("/.artifacts/bot-captures/v498_depth1.png"):
		_fail("capture path did not remain under checkout artifacts")
		return
	var scenario := {
		"id": "capture_test", "runner": "godot_client", "world_id": "dungeon_depth_one_lab",
		"client_steps": [{"type": "capture_frame", "name": "v498_depth1", "timeout_s": 5.0}],
	}
	if StepCatalogScript.validate_scenario(scenario) != "":
		_fail("safe capture scenario failed validation")
		return
	(scenario["client_steps"] as Array)[0]["name"] = "../escape"
	if StepCatalogScript.validate_scenario(scenario) == "":
		_fail("unsafe capture scenario passed validation")
		return
	(scenario["client_steps"] as Array)[0] = {
		"type": "capture_frame", "name": "safe", "timeout_s": 5.0,
		"skip_if_headless": "yes",
	}
	if StepCatalogScript.validate_scenario(scenario) == "":
		_fail("non-boolean headless skip passed validation")
		return
	(scenario["client_steps"] as Array)[0] = {"type": "set_graphics_quality", "quality": "ultra"}
	if StepCatalogScript.validate_scenario(scenario) == "":
		_fail("unsupported graphics tier passed validation")
		return
	print("[gdtest] PASS: test_bot_frame_capture")
	quit(0)


func _fail(message: String) -> void:
	push_error(message)
	print("[gdtest] FAIL: test_bot_frame_capture — %s" % message)
	quit(1)
