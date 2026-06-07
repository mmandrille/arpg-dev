# Unit tests for BotScenarioRunner: scenario parsing, step validation,
# timeout failure format, and PASS/FAIL sentinel formatting.
# No live server required. Run via: godot --headless --path client --script res://tests/test_client_bot.gd
extends SceneTree

const BotScenarioRunnerScript := preload("res://scripts/bot_scenario_runner.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")

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
	_test_full_equipment_step_types_load()
	_test_missing_menu_button_rejected()
	_test_missing_character_name_rejected()
	_test_missing_window_size_rejected()
	_test_missing_stat_button_fields_rejected()
	_test_missing_progression_expectation_rejected()
	_test_menu_step_types_load()
	_test_character_stats_step_types_load()
	_test_character_progression_assertions()
	_test_timeout_failure_message_format()
	_test_pass_sentinel_format()
	_test_fail_sentinel_format()
	_test_client_settings_supported_size_labels()
	_test_client_settings_parse_size_label()
	_test_client_settings_size_from_data()

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


func _test_full_equipment_step_types_load() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = [
		{"type": "drag_bag_to_equipment_slot", "item_def_id": "cave_helm", "slot": "head"},
		{"type": "drag_equipment_to_bag", "slot": "head"},
		{"type": "click_loot_item", "item_def_id": "cave_belt"},
		{"type": "assign_hotbar_slot", "slot_index": 5, "item_def_id": "red_potion"},
		{"type": "wait_hotbar_assigned", "slot_index": 5, "item_def_id": "red_potion", "timeout_s": 1.0},
		{"type": "use_hotbar_slot", "slot_index": 5},
		{"type": "assert_hotbar_assigned", "slot_index": 5, "item_def_id": "red_potion"},
		{"type": "assert_hotbar_capacity", "equals": 10},
		{"type": "wait_hotbar_capacity", "at_least": 3, "timeout_s": 1.0},
		{"type": "assert_hotbar_slot_disabled", "slot_index": 5},
	]
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_eq("full equipment client step scenario valid", err, "")


func _test_missing_menu_button_rejected() -> void:
	var err := BotScenarioRunnerScript.validate_step({"type": "click_menu_button"}, 0)
	_assert_ne("click_menu_button without button rejected", err, "")


func _test_missing_character_name_rejected() -> void:
	var err := BotScenarioRunnerScript.validate_step({"type": "enter_character_name"}, 0)
	_assert_ne("enter_character_name without name rejected", err, "")


func _test_missing_window_size_rejected() -> void:
	var err := BotScenarioRunnerScript.validate_step({"type": "select_window_size"}, 0)
	_assert_ne("select_window_size without size rejected", err, "")


func _test_missing_stat_button_fields_rejected() -> void:
	var click_err := BotScenarioRunnerScript.validate_step({"type": "click_stat_button"}, 0)
	_assert_ne("click_stat_button without stat rejected", click_err, "")
	var assert_err := BotScenarioRunnerScript.validate_step({"type": "assert_stat_button_enabled", "stat": "vit"}, 0)
	_assert_ne("assert_stat_button_enabled without enabled rejected", assert_err, "")


func _test_missing_progression_expectation_rejected() -> void:
	var err := BotScenarioRunnerScript.validate_step({"type": "wait_character_progression", "timeout_s": 1.0}, 0)
	_assert_ne("wait_character_progression without expectations rejected", err, "")


func _test_menu_step_types_load() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = [
		{"type": "wait_main_menu", "timeout_s": 1.0},
		{"type": "click_menu_button", "button": "new_game"},
		{"type": "enter_character_name", "name": "Bot Hero"},
		{"type": "select_character", "index": 0},
		{"type": "select_window_size", "size": "1600x900"},
		{"type": "remember_session"},
		{"type": "assert_session_changed"},
		{"type": "remember_player_position"},
		{"type": "assert_player_position_unchanged"},
	]
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_eq("menu step scenario valid", err, "")


func _test_character_stats_step_types_load() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = [
		{"type": "press_key", "keycode": "KEY_C"},
		{"type": "assert_character_stats_panel_visible", "visible": true},
		{"type": "wait_character_progression", "level": 2, "experience": 20, "timeout_s": 1.0},
		{"type": "assert_character_progression", "level": 2, "unspent_stat_points": 5, "vit": 5},
		{"type": "assert_stat_button_enabled", "stat": "vit", "enabled": true},
		{"type": "click_stat_button", "stat": "vit"},
		{"type": "assert_xp_bar", "level": 2, "experience": 20, "progress_min": 0.0, "progress_max": 0.01},
	]
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_eq("character stats step scenario valid", err, "")


func _test_character_progression_assertions() -> void:
	var runner := BotScenarioRunnerScript.new()
	var data := {
		"id": "stats_assert_test",
		"runner": "godot_client",
		"world_id": "dungeon_levels",
		"client_steps": [
			{"type": "assert_character_progression", "level": 2, "experience": 20, "unspent_stat_points": 5, "vit": 5, "derived_stats": {"max_hp": 10}, "player_max_hp": 10},
			{"type": "assert_stat_button_enabled", "stat": "vit", "enabled": true},
			{"type": "assert_xp_bar", "level": 2, "experience": 20, "progress_min": 0.0, "progress_max": 0.01},
		],
	}
	runner.load_scenario(data)
	var state := {
		"character_progression": {
			"level": 2,
			"experience": 20,
			"unspent_stat_points": 5,
			"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": 5},
			"derived_stats": {"max_hp": 10},
		},
		"player_max_hp": 10,
		"character_stats_panel": {"stat_buttons": {"vit": {"enabled": true}}},
		"consumable_bar": {"xp_bar": {"level": 2, "experience": 20, "progress": 0.0}},
	}
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	_assert_true("character progression assertions pass", runner.is_done() and runner.passed())


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


func _test_client_settings_supported_size_labels() -> void:
	var labels := ClientSettingsScript.supported_size_labels()
	_assert_true("settings include 1280x720", labels.has("1280x720"))
	_assert_true("settings include 1600x900", labels.has("1600x900"))
	_assert_true("settings include 1920x1080", labels.has("1920x1080"))


func _test_client_settings_parse_size_label() -> void:
	_assert_eq("settings parse valid size", ClientSettingsScript.parse_size_label("1600x900"), Vector2i(1600, 900))
	_assert_eq("settings parse invalid size fallback", ClientSettingsScript.parse_size_label("1440x900"), Vector2i(1920, 1080))
	_assert_eq("settings parse malformed fallback", ClientSettingsScript.parse_size_label("bad"), Vector2i(1920, 1080))


func _test_client_settings_size_from_data() -> void:
	var valid := {"window_size": {"width": 1280, "height": 720}}
	_assert_eq("settings data valid", ClientSettingsScript.size_from_data(valid), Vector2i(1280, 720))
	var invalid := {"window_size": {"width": 777, "height": 444}}
	_assert_eq("settings data invalid fallback", ClientSettingsScript.size_from_data(invalid), Vector2i(1920, 1080))
	_assert_eq("settings data missing fallback", ClientSettingsScript.size_from_data({}), Vector2i(1920, 1080))


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
