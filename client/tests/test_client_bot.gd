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
	_test_missing_skill_fields_rejected()
	_test_menu_step_types_load()
	_test_multiplayer_menu_step_types_load()
	_test_multiplayer_menu_assertions()
	_test_v45_menu_step_types_load()
	_test_v45_menu_assertions()
	_test_character_stats_step_types_load()
	_test_skill_step_types_load()
	_test_character_info_step_types_load()
	_test_character_progression_assertions()
	_test_skill_assertions()
	_test_character_info_assertions()
	_test_timeout_failure_message_format()
	_test_pass_sentinel_format()
	_test_fail_sentinel_format()
	_test_client_settings_supported_size_labels()
	_test_client_settings_parse_size_label()
	_test_client_settings_size_from_data()
	_test_client_settings_floating_combat_text_from_data()
	_test_client_settings_top_right_status_text_from_data()
	_test_client_settings_create_game_session_type_from_data()
	_test_client_settings_create_game_session_type_save_shape()
	_test_combat_feedback_step_types_load()
	_test_combat_event_and_damage_number_assertions()
	_test_inventory_paper_doll_step_types_load()
	_test_inventory_paper_doll_assertions()
	_test_wall_layout_step_types_load()
	_test_wall_layout_assertions()
	_test_shop_step_types_load()
	_test_shop_assertions()

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
		{"type": "click_loot_item", "rolled": true},
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
	_assert_ne("click loot without selector rejected", BotScenarioRunnerScript.validate_step({"type": "click_loot_item"}, 0), "")


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


func _test_missing_skill_fields_rejected() -> void:
	var progression_err := BotScenarioRunnerScript.validate_step({"type": "wait_skill_progression", "timeout_s": 1.0}, 0)
	_assert_ne("wait_skill_progression without expectations rejected", progression_err, "")
	var button_err := BotScenarioRunnerScript.validate_step({"type": "assert_skill_button_enabled"}, 0)
	_assert_ne("assert_skill_button_enabled without enabled rejected", button_err, "")
	var bar_err := BotScenarioRunnerScript.validate_step({"type": "assert_skill_bar"}, 0)
	_assert_ne("assert_skill_bar without expectations rejected", bar_err, "")


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


func _test_multiplayer_menu_step_types_load() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = [
		{"type": "wait_multiplayer_panel", "timeout_s": 1.0},
		{"type": "assert_multiplayer_panel_visible", "visible": true},
		{"type": "assert_multiplayer_session_rows", "equals": 0},
		{"type": "assert_multiplayer_session_rows", "min_count": 1, "listed": true},
		{"type": "assert_multiplayer_session_rows", "session_id_env": "ARPG_EXPECTED_JOIN_SESSION_ID", "mode": "coop", "member_count_min": 1, "connected_count_min": 1},
		{"type": "click_menu_button", "button": "refresh_sessions"},
		{"type": "click_menu_button", "button": "host_listed_session"},
		{"type": "click_menu_button", "button": "join_first_listed_session"},
		{"type": "click_menu_button", "button": "select_expected_join_session"},
		{"type": "wait_remote_player_count", "at_least": 1, "timeout_s": 1.0},
		{"type": "assert_remote_player_count", "equals": 1},
	]
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_eq("multiplayer menu step scenario valid", err, "")
	_assert_ne("multiplayer row assertion without expectation rejected", BotScenarioRunnerScript.validate_step({"type": "assert_multiplayer_session_rows"}, 0), "")
	_assert_ne("remote player count requires expectation", BotScenarioRunnerScript.validate_step({"type": "assert_remote_player_count"}, 0), "")


func _test_multiplayer_menu_assertions() -> void:
	var runner := BotScenarioRunnerScript.new()
	var data := {
		"id": "multiplayer_assert_test",
		"runner": "godot_client",
		"world_id": "dungeon_levels",
		"client_steps": [
			{"type": "wait_multiplayer_panel", "timeout_s": 1.0},
			{"type": "assert_multiplayer_panel_visible", "visible": true},
			{"type": "assert_multiplayer_session_rows", "session_id": "sess_1", "min_count": 1, "selected": true, "listed": true, "mode": "coop", "member_count_min": 3, "connected_count_min": 1},
			{"type": "assert_current_session", "session_id": "sess_1", "mode": "coop", "listed": true},
			{"type": "assert_remote_player_count", "equals": 1},
		],
	}
	runner.load_scenario(data)
	var state := {
		"multiplayer_panel_visible": true,
		"multiplayer_panel": {
			"selected_session_id": "sess_1",
			"sessions": [
				{"session_id": "sess_1", "mode": "coop", "listed": true, "member_count": 3, "connected_count": 1},
			],
		},
		"current_session_id": "sess_1",
		"current_session_mode": "coop",
		"current_session_listed": true,
		"remote_player_ids": ["1002"],
	}
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	_assert_true("multiplayer menu assertions pass", runner.is_done() and runner.passed())


func _test_v45_menu_step_types_load() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = [
		{"type": "wait_main_menu", "timeout_s": 1.0},
		{"type": "assert_main_menu_actions", "labels": ["Create Game", "Join Game", "Settings", "Exit"], "actions": ["create_game", "join_game", "settings", "exit"]},
		{"type": "click_menu_button", "button": "create_game"},
		{"type": "assert_character_panel", "mode": "forced_create", "title": "Create Character", "name_field_visible": true, "create_button_visible": true},
		{"type": "select_create_game_type", "session_type": "solo"},
		{"type": "assert_create_game_type", "session_type": "solo"},
		{"type": "click_menu_button", "button": "join_game"},
		{"type": "click_menu_button", "button": "join_selected_session"},
		{"type": "assert_current_session", "exists": true, "mode": "solo", "listed": false, "session_id": "sess_1"},
	]
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_eq("v45 menu step scenario valid", err, "")
	_assert_ne("create game type action validates session_type", BotScenarioRunnerScript.validate_step({"type": "select_create_game_type"}, 0), "")
	_assert_ne("main menu actions requires expectation", BotScenarioRunnerScript.validate_step({"type": "assert_main_menu_actions"}, 0), "")
	_assert_ne("character panel assertion requires expectation", BotScenarioRunnerScript.validate_step({"type": "assert_character_panel"}, 0), "")
	_assert_ne("current session assertion requires expectation", BotScenarioRunnerScript.validate_step({"type": "assert_current_session"}, 0), "")


func _test_v45_menu_assertions() -> void:
	var runner := BotScenarioRunnerScript.new()
	var data := {
		"id": "v45_menu_assert_test",
		"runner": "godot_client",
		"world_id": "dungeon_levels",
		"client_steps": [
			{"type": "assert_main_menu_actions", "labels": ["Create Game", "Join Game", "Settings", "Exit"], "actions": ["create_game", "join_game"]},
			{"type": "assert_character_panel", "mode": "choose_or_create", "title": "Choose Character", "min_character_count": 1, "name_field_visible": true, "create_button_visible": true},
			{"type": "assert_create_game_type", "session_type": "coop"},
			{"type": "assert_current_session", "exists": true, "mode": "coop", "listed": true},
		],
	}
	runner.load_scenario(data)
	var state := {
		"main_menu_button_labels": ["Create Game", "Join Game", "Settings", "Exit"],
		"main_menu_actions": ["create_game", "join_game", "settings", "exit"],
		"character_panel": {
			"mode": "choose_or_create",
			"title": "Choose Character",
			"characters": [{"character_id": "char_1", "dead": false}],
			"name_field_visible": true,
			"create_button_visible": true,
		},
		"create_game_session_type": "coop",
		"current_session_id": "sess_1",
		"current_session_mode": "coop",
		"current_session_listed": true,
	}
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	_assert_true("v45 menu assertions pass", runner.is_done() and runner.passed())


func _test_character_stats_step_types_load() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = [
		{"type": "press_key", "keycode": "KEY_C"},
		{"type": "assert_character_stats_panel_visible", "visible": true},
		{"type": "wait_character_progression", "level": 2, "experience": 20, "timeout_s": 1.0},
		{"type": "assert_character_progression", "level": 2, "unspent_stat_points": 3, "vit": 5},
		{"type": "assert_stat_button_enabled", "stat": "vit", "enabled": true},
		{"type": "click_stat_button", "stat": "vit"},
		{"type": "assert_xp_bar", "level": 2, "experience": 20, "progress_min": 0.0, "progress_max": 0.01},
	]
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_eq("character stats step scenario valid", err, "")


func _test_skill_step_types_load() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = [
		{"type": "press_key", "keycode": "KEY_K"},
		{"type": "assert_skills_panel_visible", "visible": true},
		{"type": "wait_skill_progression", "unspent_skill_points": 1, "rank": 0, "max_rank": 5, "can_spend": true, "timeout_s": 1.0},
		{"type": "assert_skill_progression", "unspent_skill_points": 1, "skill_id": "magic_bolt", "rank": 0, "max_rank": 5},
		{"type": "assert_skill_button_enabled", "skill_id": "magic_bolt", "enabled": true},
		{"type": "click_skill_button", "skill_id": "magic_bolt"},
		{"type": "wait_skill_bar", "skill_id": "magic_bolt", "rank": 1, "enabled": true, "remaining_ticks": 0, "timeout_s": 1.0},
		{"type": "use_skill_slot", "skill_id": "magic_bolt", "monster_def_id": "dungeon_mob"},
		{"type": "assert_skill_bar", "skill_id": "magic_bolt", "rank": 1, "disabled": true, "remaining_ticks_min": 1, "total_ticks": 40},
	]
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_eq("skill client step scenario valid", err, "")


func _test_character_info_step_types_load() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = [
		{"type": "press_key", "keycode": "KEY_P"},
		{"type": "assert_character_info_panel_visible", "visible": true},
		{"type": "assert_character_info", "name": "Hero", "level": 2, "area": "Town"},
	]
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_eq("character info step scenario valid", err, "")
	_assert_ne("character info assertion without expectation rejected", BotScenarioRunnerScript.validate_step({"type": "assert_character_info"}, 0), "")


func _test_character_progression_assertions() -> void:
	var runner := BotScenarioRunnerScript.new()
	var data := {
		"id": "stats_assert_test",
		"runner": "godot_client",
		"world_id": "dungeon_levels",
		"client_steps": [
			{"type": "assert_character_progression", "level": 2, "experience": 20, "unspent_stat_points": 3, "vit": 5, "derived_stats": {"max_hp": 10}, "player_max_hp": 10},
			{"type": "assert_stat_button_enabled", "stat": "vit", "enabled": true},
			{"type": "assert_xp_bar", "level": 2, "experience": 20, "progress_min": 0.0, "progress_max": 0.01},
		],
	}
	runner.load_scenario(data)
	var state := {
		"character_progression": {
			"level": 2,
			"experience": 20,
			"unspent_stat_points": 3,
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


func _test_skill_assertions() -> void:
	var runner := BotScenarioRunnerScript.new()
	var data := {
		"id": "skill_assert_test",
		"runner": "godot_client",
		"world_id": "dungeon_levels",
		"client_steps": [
			{"type": "assert_skill_progression", "unspent_skill_points": 1, "skill_id": "magic_bolt", "rank": 0, "max_rank": 5, "can_spend": true},
			{"type": "assert_skill_button_enabled", "skill_id": "magic_bolt", "enabled": true},
			{"type": "wait_skill_bar", "skill_id": "magic_bolt", "rank": 1, "disabled": true, "remaining_ticks_min": 1, "total_ticks": 40, "timeout_s": 1.0},
		],
	}
	runner.load_scenario(data)
	var state := {
		"skill_progression": {
			"unspent_skill_points": 1,
			"skills": [{"skill_id": "magic_bolt", "rank": 0, "max_rank": 5, "can_spend": true}],
		},
		"skills_panel": {"spend_button_enabled": true},
		"skill_bar": {"skill_id": "magic_bolt", "rank": 1, "max_rank": 5, "enabled": false, "disabled": true, "remaining_ticks": 38, "total_ticks": 40, "cooldown_fraction": 0.95},
	}
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	_assert_true("skill assertions pass", runner.is_done() and runner.passed())


func _test_character_info_assertions() -> void:
	var runner := BotScenarioRunnerScript.new()
	var data := {
		"id": "character_info_assert_test",
		"runner": "godot_client",
		"world_id": "dungeon_levels",
		"client_steps": [
			{"type": "assert_character_info_panel_visible", "visible": true},
			{"type": "assert_character_info", "name": "Hero", "level": 2, "area": "Dungeon lvl5 - Depth 5"},
		],
	}
	runner.load_scenario(data)
	var state := {
		"character_info_panel_visible": true,
		"character_info_panel": {
			"name": "Hero",
			"level": 2,
			"area": "Dungeon lvl5 - Depth 5",
		},
	}
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	_assert_true("character info assertions pass", runner.is_done() and runner.passed())


func _test_wall_layout_step_types_load() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = [
		{"type": "wait_wall_layout", "current_level": -1, "at_least": 5, "generated_at_least": 1, "non_perimeter_at_least": 1, "timeout_s": 1.0},
		{"type": "assert_wall_layout", "current_level": -1, "generated_at_least": 1},
	]
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_eq("wall layout step scenario valid", err, "")
	_assert_ne("wall layout assertion without expectation rejected", BotScenarioRunnerScript.validate_step({"type": "assert_wall_layout"}, 0), "")


func _test_wall_layout_assertions() -> void:
	var runner := BotScenarioRunnerScript.new()
	var data := {
		"id": "wall_layout_assert_test",
		"runner": "godot_client",
		"world_id": "dungeon_levels",
		"client_steps": [
			{"type": "wait_wall_layout", "current_level": -1, "at_least": 5, "generated_at_least": 1, "non_perimeter_at_least": 1, "timeout_s": 1.0},
			{"type": "assert_wall_layout", "current_level": -1, "generated_at_least": 1},
		],
	}
	runner.load_scenario(data)
	var state := {
		"current_level": -1,
		"wall_count": 8,
		"generated_wall_count": 3,
		"non_perimeter_wall_count": 3,
		"walls": [
			{"id": "wall_-1_0000", "source": "perimeter"},
			{"id": "wall_-1_0004", "source": "generated"},
		],
	}
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	_assert_true("wall layout assertions pass", runner.is_done() and runner.passed())


func _test_shop_step_types_load() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = [
		{"type": "wait_shop_panel", "equals": 7, "timeout_s": 1.0},
		{"type": "assert_shop_panel_visible", "visible": true},
		{"type": "assert_shop_offer_count", "offer_kind": "fixed", "equals": 2},
		{"type": "assert_shop_offer_count", "offer_kind": "generated", "equals": 5},
		{"type": "assert_shop_offer_count", "offer_kind": "buyback", "equals": 1},
		{"type": "assert_shop_buy_button", "offer_id": "fixed:red_potion", "enabled": true},
		{"type": "assert_shop_sell_rows", "rolled": true, "at_least": 1},
		{"type": "click_shop_buy_offer", "offer_id": "fixed:red_potion"},
		{"type": "click_shop_buy_offer", "offer_kind": "generated", "offer_index": 0},
		{"type": "click_shop_sell_item", "rolled": true, "bag_index": 0},
	]
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_eq("shop client step scenario valid", err, "")
	_assert_ne("shop panel wait without expectation rejected", BotScenarioRunnerScript.validate_step({"type": "wait_shop_panel", "timeout_s": 1.0}, 0), "")
	_assert_ne("shop buy button without offer rejected", BotScenarioRunnerScript.validate_step({"type": "assert_shop_buy_button"}, 0), "")
	_assert_ne("shop click buy without selector rejected", BotScenarioRunnerScript.validate_step({"type": "click_shop_buy_offer"}, 0), "")


func _test_shop_assertions() -> void:
	var runner := BotScenarioRunnerScript.new()
	var data := {
		"id": "shop_assert_test",
		"runner": "godot_client",
		"world_id": "dungeon_levels",
		"client_steps": [
			{"type": "wait_shop_panel", "equals": 7, "timeout_s": 1.0},
			{"type": "assert_shop_panel_visible", "visible": true},
			{"type": "assert_shop_offer_count", "offer_kind": "generated", "equals": 5},
			{"type": "assert_shop_buy_button", "offer_id": "fixed:red_potion", "enabled": true},
			{"type": "assert_shop_sell_rows", "rolled": true, "at_least": 1},
		],
	}
	runner.load_scenario(data)
	var state := {
		"shop_panel_visible": true,
		"shop_panel": {
			"offer_count": 7,
			"fixed_offer_count": 2,
			"generated_offer_count": 5,
			"buyback_offer_count": 0,
			"buy_buttons": {
				"fixed:red_potion": {"enabled": true},
			},
			"sell_rows": [
				{"item_instance_id": "1", "item_def_id": "cave_bow", "item_template_id": "cave_bow"},
			],
		},
	}
	for _i in range(5):
		runner.tick(0.016, state)
	_assert_true("shop assertions pass", runner.is_done() and runner.passed())


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


func _test_client_settings_floating_combat_text_from_data() -> void:
	_assert_eq("settings floating text defaults on", ClientSettingsScript.floating_combat_text_from_data({}), true)
	_assert_eq("settings floating text parses off", ClientSettingsScript.floating_combat_text_from_data({"floating_combat_text": false}), false)


func _test_client_settings_top_right_status_text_from_data() -> void:
	_assert_eq("settings top-right text defaults on", ClientSettingsScript.top_right_status_text_from_data({}), true)
	_assert_eq("settings top-right text parses off", ClientSettingsScript.top_right_status_text_from_data({"top_right_status_text": false}), false)


func _test_client_settings_create_game_session_type_from_data() -> void:
	_assert_eq("settings create game type defaults coop", ClientSettingsScript.create_game_session_type_from_data({}), "coop")
	_assert_eq("settings create game type parses solo", ClientSettingsScript.create_game_session_type_from_data({"create_game_session_type": "solo"}), "solo")
	_assert_eq("settings create game type normalizes invalid", ClientSettingsScript.create_game_session_type_from_data({"create_game_session_type": "lan"}), "coop")
	_assert_eq("settings create game label coop", ClientSettingsScript.create_game_session_type_label("coop"), "Co-op")
	_assert_eq("settings create game label solo", ClientSettingsScript.create_game_session_type_label("solo"), "Solo")


func _test_client_settings_create_game_session_type_save_shape() -> void:
	var path := "user://test_settings_v45.json"
	var absolute_path := ProjectSettings.globalize_path(path)
	if FileAccess.file_exists(path):
		DirAccess.remove_absolute(absolute_path)
	var settings = ClientSettingsScript.new(path)
	settings.set_create_game_session_type("solo", false)
	settings.save()
	var parsed = JSON.parse_string(FileAccess.get_file_as_string(path))
	_assert_eq("settings save includes create game session type", str((parsed as Dictionary).get("create_game_session_type", "")), "solo")
	var reloaded = ClientSettingsScript.new(path)
	reloaded.load()
	_assert_eq("settings reload restores create game session type", reloaded.create_game_session_type, "solo")
	DirAccess.remove_absolute(absolute_path)


func _test_combat_feedback_step_types_load() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = [
		{"type": "set_floating_combat_text", "enabled": false},
		{"type": "assert_floating_combat_text_enabled", "enabled": false},
		{"type": "set_floating_combat_text", "enabled": true},
		{"type": "wait_event", "event_type": "monster_damaged", "outcome": "block", "blocked": true, "timeout_s": 1.0},
		{"type": "wait_damage_number", "variant": "block", "text": "BLOCK", "timeout_s": 1.0},
		{"type": "wait_no_damage_number", "timeout_s": 1.0},
		{"type": "assert_damage_number", "variant": "block", "text": "BLOCK"},
		{"type": "assert_no_damage_number"},
		{"type": "assert_character_progression", "stat_breakdowns": [
			{"key": "block_percent", "min_value": 6, "cap": 75, "source_kinds": ["equipment_base", "equipment_roll"]}
		]},
	]
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_eq("combat feedback client step scenario valid", err, "")


func _test_combat_event_and_damage_number_assertions() -> void:
	var runner := BotScenarioRunnerScript.new()
	var data := {
		"id": "combat_feedback_assert_test",
		"runner": "godot_client",
		"world_id": "combat_stat_lab",
		"client_steps": [
			{"type": "wait_event", "event_type": "monster_damaged", "outcome": "block", "blocked": true, "timeout_s": 1.0},
			{"type": "assert_damage_number", "variant": "block", "text": "BLOCK"},
			{"type": "assert_floating_combat_text_enabled", "enabled": true},
			{"type": "assert_character_progression", "stat_breakdowns": [
				{"key": "block_percent", "min_value": 6, "cap": 75, "source_kinds": ["equipment_base", "equipment_roll"]}
			]},
		],
	}
	runner.load_scenario(data)
	var state := {
		"pending_events": [{"event_type": "monster_damaged", "outcome": "block", "blocked": true}],
		"damage_numbers": [{"variant": "block", "text": "BLOCK"}],
		"floating_combat_text_enabled": true,
		"character_progression": {
			"stat_breakdowns": [{
				"key": "block_percent",
				"value": 10,
				"uncapped_value": 10,
				"cap": 75,
				"sources": [
					{"kind": "equipment_base", "value": 5},
					{"kind": "equipment_roll", "value": 5}
				],
			}],
		},
	}
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	_assert_true("combat feedback assertions pass", runner.is_done() and runner.passed())


func _test_inventory_paper_doll_step_types_load() -> void:
	var data := _make_valid_scenario()
	data["client_steps"] = [
		{"type": "assert_inventory_capacity", "rows": 3, "capacity": 15},
		{"type": "assert_bag_grid", "columns": 5, "available_slot_count": 15},
		{"type": "assert_paper_doll_layout", "slots": ["head", "belt"], "preview": true},
	]
	var err := BotScenarioRunnerScript.validate_scenario(data)
	_assert_eq("inventory paper doll client step scenario valid", err, "")
	_assert_ne("inventory capacity without expectation rejected", BotScenarioRunnerScript.validate_step({"type": "assert_inventory_capacity"}, 0), "")
	_assert_ne("bag grid without expectation rejected", BotScenarioRunnerScript.validate_step({"type": "assert_bag_grid"}, 0), "")
	_assert_ne("paper doll layout without slots rejected", BotScenarioRunnerScript.validate_step({"type": "assert_paper_doll_layout"}, 0), "")


func _test_inventory_paper_doll_assertions() -> void:
	var runner := BotScenarioRunnerScript.new()
	var data := {
		"id": "inventory_paper_doll_assert_test",
		"runner": "godot_client",
		"world_id": "inventory_capacity_lab",
		"client_steps": [
			{"type": "assert_inventory_capacity", "rows": 4, "capacity": 20},
			{"type": "assert_bag_grid", "columns": 5, "available_slot_count": 20},
			{"type": "assert_paper_doll_layout", "slots": ["head", "belt", "main_hand", "off_hand"], "preview": true},
		],
	}
	runner.load_scenario(data)
	var state := {
		"inventory_rows": 4,
		"inventory_capacity": 20,
		"inventory_panel": {
			"inventory_rows": 4,
			"inventory_capacity": 20,
			"bag_columns": 5,
			"available_slot_count": 20,
			"paper_doll_preview": {"exists": true, "visible": true},
			"paper_doll_slots": {
				"head": {"exists": true},
				"belt": {"exists": true},
				"main_hand": {"exists": true},
				"off_hand": {"exists": true},
			},
		},
	}
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	runner.tick(0.016, state)
	_assert_true("inventory paper doll assertions pass", runner.is_done() and runner.passed())


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
