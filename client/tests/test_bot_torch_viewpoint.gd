extends SceneTree

const ViewpointScript := preload("res://scripts/bot_torch_viewpoint.gd")
const StepCatalogScript := preload("res://scripts/bot_step_catalog.gd")


func _initialize() -> void:
	var target := ViewpointScript.nearest_approach(
		[{"x": 2.0, "z": 0.0}, {"x": 20.0, "z": 0.0}],
		{"x": 0.0, "z": 0.0}, 2.0)
	if target.get("mount", {}) != {"x": 2.0, "z": 0.0}:
		_fail("nearest torch was not selected")
		return
	if target.get("approach", {}) != {"x": 0.0, "z": 0.0}:
		_fail("approach did not stop on the player side of the wall")
		return
	var pinned := ViewpointScript.nearest_approach(
		[{"x": 2.0, "z": 0.0}, {"x": 20.0, "z": 0.0}],
		{"x": 0.0, "z": 0.0}, 2.0, 1)
	if pinned.get("mount", {}) != {"x": 20.0, "z": 0.0}:
		_fail("pinned torch index was not selected")
		return
	if not ViewpointScript.nearest_approach([], {"x": 0.0, "z": 0.0}, 2.0).is_empty():
		_fail("empty torch layout selected a target")
		return
	var scenario := {
		"id": "torch_fixture_test", "runner": "godot_client", "world_id": "dungeon_levels",
		"client_steps": [
			{"type": "approach_nearest_torch", "stand_off": 2.5},
			{"type": "wait_selected_torch_in_view", "max_distance": 5.0, "timeout_s": 10.0},
		],
	}
	if StepCatalogScript.validate_scenario(scenario) != "":
		_fail("valid torch fixture was rejected")
		return
	(scenario["client_steps"] as Array)[0]["stand_off"] = -1.0
	if StepCatalogScript.validate_scenario(scenario) == "":
		_fail("invalid torch approach was accepted")
		return
	(scenario["client_steps"] as Array)[0]["stand_off"] = 2.5
	(scenario["client_steps"] as Array)[0]["torch_index"] = "invalid"
	if StepCatalogScript.validate_scenario(scenario) == "":
		_fail("noninteger torch index was accepted")
		return
	print("[gdtest] PASS: test_bot_torch_viewpoint")
	quit(0)


func _fail(message: String) -> void:
	push_error(message)
	print("[gdtest] FAIL: test_bot_torch_viewpoint — %s" % message)
	quit(1)
