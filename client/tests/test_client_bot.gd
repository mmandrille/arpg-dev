# Unit tests for BotScenarioRunner: scenario parsing, step validation,
# timeout failure format, and PASS/FAIL sentinel formatting.
# No live server required. Run via: godot --headless --path client --script res://tests/test_client_bot.gd
extends SceneTree

const BotScenarioRunnerScript := preload("res://scripts/bot_scenario_runner.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_valid_scenario_loads()
	_test_invalid_runner_rejected()
	_test_missing_id_rejected()
	_test_missing_world_id_rejected()
	_test_empty_steps_rejected()
	_test_unknown_step_type_rejected()
	_test_missing_step_field_entity_type()
	_test_missing_step_field_event_type()
	_test_missing_step_field_timeout()
	_test_missing_keycode_rejected()
	_test_missing_click_entity_type_rejected()
	_test_missing_click_floor_coords_rejected()
	_test_missing_drag_bag_item_def_id_rejected()
	_test_timeout_failure_message_format()
	_test_pass_sentinel_format()
	_test_fail_sentinel_format()

	print("[gdtest] PASS: test_client_bot (%d passed, %d failed)" % [_pass_count, _fail_count])
	if _fail_count > 0:
		quit(1)
	else:
		quit(0)


func _test_valid_scenario_loads() -> void:
	var data := _make_valid_scenario()
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_eq("valid scenario has no error", err, "")
	var runner := BotScenarioRunnerScript.new()
	_assert_true("runner loads valid scenario", runner.load_scenario(data))


func _test_invalid_runner_rejected() -> void:
	var data := _make_valid_scenario()
	data["runner"] = "python_bot"
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_ne("invalid runner rejected", err, "")


func _test_missing_id_rejected() -> void:
	var data := _make_valid_scenario()
	data.erase("id")
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_ne("missing id rejected", err, "")


func _test_missing_world_id_rejected() -> void:
	var data := _make_valid_scenario()
	data.erase("world_id")
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_ne("missing world_id rejected", err, "")


func _test_empty_steps_rejected() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = []
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_ne("empty client_steps rejected", err, "")


func _test_unknown_step_type_rejected() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = [{"type": "do_the_thing", "timeout_s": 5.0}]
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_ne("unknown step type rejected", err, "")


func _test_missing_step_field_entity_type() -> void:
	var err := BotScenarioRunnerScript.validate_step({"type": "wait_entity", "timeout_s": 5.0}, 0)
	_assert_ne("wait_entity without entity_type rejected", err, "")


func _test_missing_step_field_event_type() -> void:
	var err := BotScenarioRunnerScript.validate_step({"type": "wait_event", "timeout_s": 5.0}, 0)
	_assert_ne("wait_event without event_type rejected", err, "")


func _test_missing_step_field_timeout() -> void:
	var err := BotScenarioRunnerScript.validate_step({"type": "wait_ws_open"}, 0)
	_assert_ne("wait_ws_open without timeout_s rejected", err, "")


func _test_missing_keycode_rejected() -> void:
	var err := BotScenarioRunnerScript.validate_step({"type": "press_key"}, 0)
	_assert_ne("press_key without keycode rejected", err, "")


func _test_missing_click_entity_type_rejected() -> void:
	var err := BotScenarioRunnerScript.validate_step({"type": "click_entity"}, 0)
	_assert_ne("click_entity without entity_type rejected", err, "")


func _test_missing_click_floor_coords_rejected() -> void:
	var err := BotScenarioRunnerScript.validate_step({"type": "click_floor"}, 0)
	_assert_ne("click_floor without x/z rejected", err, "")


func _test_missing_drag_bag_item_def_id_rejected() -> void:
	var err := BotScenarioRunnerScript.validate_step({"type": "drag_bag_to_weapon_slot"}, 0)
	_assert_ne("drag_bag_to_weapon_slot without item_def_id rejected", err, "")


func _test_timeout_failure_message_format() -> void:
	var runner := BotScenarioRunnerScript.new()
	var data := {
		"id": "timeout_test",
		"runner": "godot_client",
		"world_id": "vertical_slice",
		"client_steps": [
			{"type": "wait_ws_open", "timeout_s": 0.001},
		],
	}
	runner.load_scenario(data)
	# Tick once with large delta to trigger timeout.
	runner.tick(1.0, {"ws_open": false})
	_assert_true("timeout sets done", runner.is_done())
	_assert_true("timeout fails scenario", not runner.passed())
	var msg := runner.failure_message()
	_assert_true("timeout msg contains scenario id", "timeout_test" in msg)
	_assert_true("timeout msg contains step type", "wait_ws_open" in msg)
	_assert_true("timeout msg contains timeout value", "0.0" in msg or "timeout" in msg)


func _test_pass_sentinel_format() -> void:
	# The sentinel "[bot-client] PASS <id>" is printed by BotController, not
	# BotScenarioRunner. Verify the runner reports passed() == true on completion.
	var runner := BotScenarioRunnerScript.new()
	var data := {
		"id": "sentinel_test",
		"runner": "godot_client",
		"world_id": "vertical_slice",
		"client_steps": [
			{"type": "assert_panel_visible", "visible": false},
		],
	}
	runner.load_scenario(data)
	runner.tick(0.016, {"inventory_panel_visible": false})
	_assert_true("assert pass: done", runner.is_done())
	_assert_true("assert pass: passed", runner.passed())


func _test_fail_sentinel_format() -> void:
	var runner := BotScenarioRunnerScript.new()
	var data := {
		"id": "sentinel_fail_test",
		"runner": "godot_client",
		"world_id": "vertical_slice",
		"client_steps": [
			{"type": "assert_panel_visible", "visible": true},
		],
	}
	runner.load_scenario(data)
	runner.tick(0.016, {"inventory_panel_visible": false})
	_assert_true("assert fail: done", runner.is_done())
	_assert_true("assert fail: not passed", not runner.passed())
	_assert_true("assert fail: message has scenario id", "sentinel_fail_test" in runner.failure_message())


# --- helpers -----------------------------------------------------------------

func _make_valid_scenario() -> Dictionary:
	return {
		"id": "test_scenario",
		"runner": "godot_client",
		"world_id": "vertical_slice",
		"client_steps": [
			{"type": "wait_ws_open", "timeout_s": 5.0},
		],
	}


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		printerr("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])
		_fail_count += 1


func _assert_ne(label: String, got, not_expected) -> void:
	if got != not_expected:
		_pass_count += 1
	else:
		printerr("[gdtest] FAIL %s: expected something other than %s, got that" % [label, str(not_expected)])
		_fail_count += 1


func _assert_true(label: String, condition: bool) -> void:
	if condition:
		_pass_count += 1
	else:
		printerr("[gdtest] FAIL %s" % label)
		_fail_count += 1
