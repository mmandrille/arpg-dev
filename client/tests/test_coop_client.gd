# Unit tests for local/remote co-op snapshot handling (v33).
# Run via: godot --headless --path client --script res://tests/test_coop_client.gd
extends SceneTree

const MainScript := preload("res://scripts/main.gd")
const NetClientScript := preload("res://scripts/net_client.gd")
const CharacterSelectPanelScript := preload("res://scripts/character_select_panel.gd")
const MultiplayerSessionsPanelScript := preload("res://scripts/multiplayer_sessions_panel.gd")
const PlayerHealthBarScript := preload("res://scripts/player_health_bar.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")
const SettingsPanelScript := preload("res://scripts/settings_panel.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")
const ShopPanelScript := preload("res://scripts/shop_panel.gd")
const StashPanelScript := preload("res://scripts/stash_panel.gd")
const CharacterStatsPanelScript := preload("res://scripts/character_stats_panel.gd")
const SkillsPanelScript := preload("res://scripts/skills_panel.gd")
const CharacterBarScript := preload("res://scripts/character_bar.gd")
const SkillBarScript := preload("res://scripts/skill_bar.gd")
const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const HealRainEffectScript := preload("res://scripts/heal_rain_effect.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_net_client_base_url_and_ws_url()
	_test_local_and_remote_players_apply_from_snapshot()
	_test_snapshot_wall_layout_rendering()
	_test_delta_wall_layout_replacement()
	_test_teardown_clears_wall_layout()
	_test_preset_world_wall_fallback()
	_test_local_player_model_front_faces_direction()
	_test_remote_player_delta_and_remove()
	_test_multiple_remote_players_update_and_remove_independently()
	_test_path_reject_clears_held_click_state()
	_test_capacity_reject_shows_bag_full_unequip_message()
	_test_no_mana_reject_shows_floating_text()
	_test_skill_cooldown_reject_shows_floating_text()
	_test_player_healed_spawns_heal_rain()
	_test_local_attack_range_uses_equipped_reach()
	_test_basic_attack_cooldown_uses_derived_interval()
	_test_character_bar_opens_stats_panel()
	_test_skill_function_key_selects_right_click_skill()
	_test_learned_skill_auto_selects_right_click()
	_test_skill_cast_payload_uses_direction_without_nearest_fallback()
	_test_loss_popup_shows_for_dead_local_player()
	_test_dead_character_rows_are_disabled()
	_test_character_panel_modes_for_v45()
	_test_multiplayer_sessions_panel_row_join_affordances()
	_test_settings_panel_create_game_type_sync()
	_test_status_text_toggle_hides_left_debug_not_level_hud()
	_test_player_hud_identity_uses_character_name_and_level()
	_test_character_stats_probability_values_use_percentages()
	_test_character_stats_window_chrome()
	_test_draggable_window_persists_layout()
	_test_actionable_panels_autoclose_out_of_range()
	_test_movement_closes_gameplay_panels()

	print("[gdtest] PASS: test_coop_client (%d passed, %d failed)" % [_pass_count, _fail_count])
	if _fail_count > 0:
		quit(1)
	else:
		quit(0)


func _make_main():
	var main = MainScript.new()
	main.player_anchor = Node3D.new()
	main.entities_root = Node3D.new()
	main.walls_root = Node3D.new()
	get_root().add_child(main.player_anchor)
	get_root().add_child(main.entities_root)
	get_root().add_child(main.walls_root)
	return main


func _test_net_client_base_url_and_ws_url() -> void:
	var http_client = NetClientScript.new("http://localhost:18080")
	http_client.ws_url = "/v0/ws?session_id=sess_1"
	http_client.token = "tok"
	_assert_eq("http websocket URL", http_client.websocket_url(), "ws://localhost:18080/v0/ws?session_id=sess_1&access_token=tok")
	var https_client = NetClientScript.new("https://example.test/some/path")
	https_client.ws_url = "/v0/ws?session_id=sess_2"
	https_client.token = "tok2"
	_assert_eq("https websocket URL", https_client.websocket_url(), "wss://example.test:443/v0/ws?session_id=sess_2&access_token=tok2")


func _test_local_and_remote_players_apply_from_snapshot() -> void:
	var main = _make_main()
	main._apply_snapshot({
		"server_tick": 1,
		"current_level": 0,
		"local_player_id": "1001",
		"party": [
			{"player_id": "1001", "role": "host", "connected": true},
			{"player_id": "1002", "role": "guest", "connected": true},
			{"player_id": "1003", "role": "guest", "connected": true},
		],
		"entities": [
			{"id": "1001", "type": "player", "position": {"x": 2.0, "y": 3.0}, "hp": 9, "max_hp": 10, "character_id": "char_host"},
			{"id": "1002", "type": "player", "position": {"x": 4.0, "y": 5.0}, "hp": 8, "max_hp": 10, "character_id": "char_guest"},
			{"id": "1003", "type": "player", "position": {"x": 6.0, "y": 7.0}, "hp": 10, "max_hp": 10, "character_id": "char_guest_2"},
		],
		"inventory": [],
		"equipped": {},
		"hotbar": [],
		"character_progression": {},
	})
	_assert_eq("local player id from snapshot", main.player_id, "1001")
	_assert_vec3("local player anchor position", main.player_anchor.position, Vector3(2.0, 0.0, 3.0))
	_assert_true("remote player entity stored", main.entities.has("1002"))
	_assert_eq("remote entity type", str(main.entities["1002"].get("type", "")), "player")
	_assert_eq("remote character metadata", str(main.entities["1002"].get("character_id", "")), "char_guest")
	_assert_eq("remote visual tint", str(main.entities["1002"].get("base_tint", "")), MainScript.REMOTE_PLAYER_TINT.to_html(false))
	_assert_true("remote player has character model", (main.entities["1002"]["node"] as Node3D).find_child("ModelRoot", true, false) != null)
	_assert_true("remote player has animation controller", main.entities["1002"].get("controller", null) != null)
	_assert_true("remote player has reaction controller", main.entities["1002"].get("reaction", null) != null)
	_assert_true("second remote player entity stored", main.entities.has("1003"))
	_assert_eq("second remote character metadata", str(main.entities["1003"].get("character_id", "")), "char_guest_2")
	_assert_eq("party count", main.party.size(), 3)
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_snapshot_wall_layout_rendering() -> void:
	var main = _make_main()
	main.current_world_id = "dungeon_levels"
	main._apply_snapshot({
		"server_tick": 1,
		"current_level": -1,
		"local_player_id": "1001",
		"party": [],
		"walls": [
			{"id": "wall_-1_0000", "position": {"x": 5.0, "y": 0.0}, "size": {"x": 10.0, "y": 1.0}, "source": "perimeter"},
			{"id": "wall_-1_0001", "position": {"x": 12.0, "y": 8.0}, "size": {"x": 3.0, "y": 1.0}, "source": "generated"},
		],
		"entities": [
			{"id": "1001", "type": "player", "position": {"x": 2.0, "y": 3.0}, "hp": 10, "max_hp": 10},
		],
		"inventory": [],
		"equipped": {},
		"hotbar": [],
		"character_progression": {},
	})
	_assert_eq("snapshot wall nodes", main.walls_root.get_child_count(), 2)
	_assert_eq("snapshot wall layout count", main.current_wall_layout.size(), 2)
	_assert_eq("snapshot generated wall count", int(main.get_bot_state().get("generated_wall_count", 0)), 1)
	_assert_eq("snapshot wall metadata source", str(main.walls_root.get_child(1).get_meta("source", "")), "generated")
	_assert_vec3("snapshot generated wall position", (main.walls_root.get_child(1) as Node3D).position, Vector3(12.0, 0.5, 8.0))
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_delta_wall_layout_replacement() -> void:
	var main = _make_main()
	main.current_world_id = "dungeon_levels"
	main._apply_snapshot({
		"server_tick": 1,
		"current_level": -1,
		"local_player_id": "1001",
		"party": [],
		"walls": [
			{"id": "wall_-1_old", "position": {"x": 1.0, "y": 1.0}, "size": {"x": 1.0, "y": 1.0}, "source": "generated"},
		],
		"entities": [
			{"id": "1001", "type": "player", "position": {"x": 2.0, "y": 3.0}, "hp": 10, "max_hp": 10},
		],
		"inventory": [],
		"equipped": {},
		"hotbar": [],
		"character_progression": {},
	})
	main._apply_delta({
		"events": [
			{"event_type": "level_changed", "entity_id": "1001", "to_level": -2},
		],
		"changes": [
			{"op": "wall_layout_update", "walls": [
				{"id": "wall_-2_0000", "position": {"x": 4.0, "y": 0.0}, "size": {"x": 8.0, "y": 1.0}, "source": "perimeter"},
				{"id": "wall_-2_0001", "position": {"x": 9.0, "y": 6.0}, "size": {"x": 1.0, "y": 4.0}, "source": "generated"},
			]},
			{"op": "entity_spawn", "entity": {"id": "2001", "type": "loot", "item_def_id": "gold", "position": {"x": 6.0, "y": 6.0}}},
		],
	})
	_assert_eq("delta current level", main.current_level, -2)
	_assert_eq("delta wall nodes replaced", main.walls_root.get_child_count(), 2)
	_assert_eq("delta generated wall count", int(main.get_bot_state().get("generated_wall_count", 0)), 1)
	_assert_eq("delta removed old wall", str(main.walls_root.get_child(0).get_meta("wall_id", "")), "wall_-2_0000")
	_assert_true("delta entity spawned after wall update", main.entities.has("2001"))
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_teardown_clears_wall_layout() -> void:
	var main = _make_main()
	main.current_world_id = "dungeon_levels"
	main._render_wall_layout([
		{"id": "wall_test", "position": {"x": 1.0, "y": 2.0}, "size": {"x": 3.0, "y": 4.0}, "source": "generated"},
	])
	_assert_eq("teardown precondition wall nodes", main.walls_root.get_child_count(), 1)
	main._teardown_gameplay_state(false)
	_assert_eq("teardown clears wall nodes", main.walls_root.get_child_count(), 0)
	_assert_eq("teardown clears wall layout", main.current_wall_layout.size(), 0)
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_preset_world_wall_fallback() -> void:
	var main = _make_main()
	main.current_world_id = "vertical_slice"
	main._apply_snapshot({
		"server_tick": 1,
		"current_level": 0,
		"local_player_id": "1001",
		"party": [],
		"entities": [
			{"id": "1001", "type": "player", "position": {"x": 2.0, "y": 3.0}, "hp": 10, "max_hp": 10},
		],
		"inventory": [],
		"equipped": {},
		"hotbar": [],
		"character_progression": {},
	})
	_assert_eq("preset fallback wall nodes", main.walls_root.get_child_count(), 4)
	_assert_eq("preset fallback wall source", str(main.walls_root.get_child(0).get_meta("source", "")), "preset")
	_assert_eq("preset fallback debug wall count", int(main.get_bot_state().get("wall_count", 0)), 4)
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_loss_popup_shows_for_dead_local_player() -> void:
	var main = _make_main()
	main.loss_popup = main._build_loss_popup()
	get_root().add_child(main.loss_popup)
	main._apply_snapshot({
		"server_tick": 1,
		"current_level": 0,
		"local_player_id": "1001",
		"party": [],
		"entities": [
			{"id": "1001", "type": "player", "position": {"x": 2.0, "y": 3.0}, "hp": 0, "max_hp": 10},
		],
		"inventory": [],
		"equipped": {},
		"hotbar": [],
		"character_progression": {},
	})
	_assert_true("dead local player shows loss popup", bool(main.get_bot_state().get("loss_popup_visible", false)))
	main.loss_popup.queue_free()
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_dead_character_rows_are_disabled() -> void:
	var panel: CharacterSelectPanel = CharacterSelectPanelScript.new()
	get_root().add_child(panel)
	panel._build()
	panel.show_continue([
		{"character_id": "char_dead", "name": "Fallen", "created_at": "2026-06-09T00:00:00Z", "dead": true},
		{"character_id": "char_live", "name": "Alive", "created_at": "2026-06-09T00:00:00Z", "dead": false},
	])
	var dead_row := panel._rows.get_child(0) as HBoxContainer
	var dead_button := dead_row.get_child(0) as Button
	_assert_true("dead character select disabled", dead_button.disabled)
	_assert_true("dead character has skull marker", dead_button.text.begins_with("☠"))
	_assert_true("dead character row keeps summary", dead_button.text.find("Lv 1") >= 0 and dead_button.text.find("0g") >= 0 and dead_button.text.find("D0") >= 0)
	var debug := panel.get_debug_state()
	var rows: Array = debug.get("character_rows", [])
	_assert_eq("character summary row count", rows.size(), 2)
	_assert_eq("dead character summary status", str((rows[0] as Dictionary).get("status", "")), "Dead")
	var started := {"id": ""}
	panel.start_requested.connect(func(character_id: String) -> void:
		started["id"] = character_id
	)
	panel.start_character_at_index(0)
	_assert_eq("dead character did not start", str(started["id"]), "")
	panel.start_character_at_index(1)
	_assert_eq("live character starts", str(started["id"]), "char_live")
	panel.queue_free()


func _test_character_panel_modes_for_v45() -> void:
	var panel: CharacterSelectPanel = CharacterSelectPanelScript.new()
	get_root().add_child(panel)
	panel._build()
	panel.show_forced_create("Create Character")
	var forced := panel.get_debug_state()
	_assert_eq("forced create mode", str(forced.get("mode", "")), "forced_create")
	_assert_eq("forced create title", str(forced.get("title", "")), "Create Character")
	_assert_true("forced create shows name field", bool(forced.get("name_field_visible", false)))
	_assert_eq("forced create uses check action", str(forced.get("create_button_text", "")), "✓")
	_assert_true("forced create hides empty label", not bool(forced.get("empty_visible", true)))
	panel.show_choose_or_create([
		{"character_id": "char_live", "name": "Alive", "created_at": "", "dead": false, "level": 4, "gold": 12, "deepest_dungeon_depth": 2},
	], "Choose Character")
	var choose := panel.get_debug_state()
	_assert_eq("choose mode", str(choose.get("mode", "")), "choose_or_create")
	_assert_eq("choose title", str(choose.get("title", "")), "Choose Character")
	_assert_true("choose keeps create affordance", bool(choose.get("create_button_visible", false)))
	_assert_true("choose hides create name field until requested", not bool(choose.get("name_field_visible", true)))
	_assert_eq("choose create action is explicit", str(choose.get("create_button_text", "")), "Create Character")
	_assert_eq("choose character count", (choose.get("characters", []) as Array).size(), 1)
	var choose_rows: Array = choose.get("character_rows", [])
	_assert_eq("choose row level", int((choose_rows[0] as Dictionary).get("level", 0)), 4)
	_assert_eq("choose row gold", int((choose_rows[0] as Dictionary).get("gold", 0)), 12)
	_assert_eq("choose row depth", int((choose_rows[0] as Dictionary).get("deepest_dungeon_depth", 0)), 2)
	_assert_true("choose row label includes summary", str((choose_rows[0] as Dictionary).get("label", "")).find("Lv 4") >= 0)
	_assert_eq("choose row default class", str((choose_rows[0] as Dictionary).get("character_class", "")), "barbarian")
	panel.submit_name()
	var expanded := panel.get_debug_state()
	_assert_true("create press reveals name field", bool(expanded.get("name_field_visible", false)))
	_assert_eq("expanded create action uses check", str(expanded.get("create_button_text", "")), "✓")
	_assert_true("expanded class picker visible", bool(expanded.get("class_picker_visible", false)))
	_assert_eq("default selected class", str(expanded.get("selected_class", "")), "barbarian")
	var options: Array = expanded.get("class_options", [])
	_assert_eq("three class options", options.size(), 3)
	_assert_true("barbarian tooltip includes skill", str((options[0] as Dictionary).get("tooltip", "")).contains("Skill: Rage"))
	panel.select_class("sorcerer")
	var sorc_state := panel.get_debug_state()
	_assert_eq("selected sorcerer", str(sorc_state.get("selected_class", "")), "sorcerer")
	_assert_true("sorcerer tooltip includes magic bolt", str(((sorc_state.get("class_options", []) as Array)[1] as Dictionary).get("tooltip", "")).contains("Magic Bolt"))
	var created := {"name": "", "class": ""}
	panel.create_requested.connect(func(name: String, character_class: String) -> void:
		created["name"] = name
		created["class"] = character_class
	)
	panel.set_name_text("Fresh Hero")
	panel.submit_name()
	_assert_eq("check creates character", str(created["name"]), "Fresh Hero")
	_assert_eq("check creates selected class", str(created["class"]), "sorcerer")
	panel.queue_free()


func _test_multiplayer_sessions_panel_row_join_affordances() -> void:
	var panel: MultiplayerSessionsPanel = MultiplayerSessionsPanelScript.new()
	get_root().add_child(panel)
	panel._build()
	panel.show_panel()
	panel.set_sessions([
		{"session_id": "sess_1", "host_display_name": "Host", "connected_count": 1, "member_count": 4, "world_id": "dungeon_levels", "mode": "coop", "listed": true},
	])
	var debug := panel.get_debug_state()
	var actions: Array = debug.get("actions", [])
	_assert_true("join selected action removed", not actions.has("join_selected_session"))
	var row := panel._rows.get_child(0) as HBoxContainer
	_assert_eq("session row has label and check", row.get_child_count(), 2)
	var row_button := row.get_child(0) as Button
	var join_button := row.get_child(1) as Button
	_assert_eq("session row join uses check", join_button.text, "✓")
	var joined := {"id": ""}
	panel.join_requested.connect(func(session_id: String) -> void:
		joined["id"] = session_id
	)
	join_button.pressed.emit()
	_assert_eq("row check joins session", str(joined["id"]), "sess_1")
	joined["id"] = ""
	var event := InputEventMouseButton.new()
	event.button_index = MOUSE_BUTTON_LEFT
	event.pressed = true
	event.double_click = true
	row_button.emit_signal("gui_input", event)
	_assert_eq("row double click joins session", str(joined["id"]), "sess_1")
	panel.queue_free()


func _test_settings_panel_create_game_type_sync() -> void:
	var panel: SettingsPanel = SettingsPanelScript.new()
	get_root().add_child(panel)
	panel._build()
	panel.show_settings("1920x1080", true, true, "solo")
	_assert_true("solo create game type selected", (panel._session_type_buttons["solo"] as Button).disabled)
	_assert_true("coop create game type available", not (panel._session_type_buttons["coop"] as Button).disabled)
	panel.set_create_game_session_type("coop")
	_assert_true("coop create game type selected", (panel._session_type_buttons["coop"] as Button).disabled)
	panel.queue_free()


func _test_status_text_toggle_hides_left_debug_not_level_hud() -> void:
	var main = _make_main()
	main._debug_label = Label.new()
	main._level_label = Label.new()
	main.client_settings = ClientSettingsScript.new()
	main.client_settings.status_text = false
	main.current_level = -3
	main._update_level_hud()
	main._update_debug()
	_assert_true("status text off hides left debug label", not main._debug_label.visible)
	_assert_true("status text off keeps right level visible", main._level_label.visible)
	_assert_true("right level still shows dungeon depth", main._level_label.text.begins_with("Level 3"))
	main._debug_label.free()
	main._level_label.free()
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_player_hud_identity_uses_character_name_and_level() -> void:
	var main = _make_main()
	main._health_bar = PlayerHealthBarScript.new()
	main._health_bar._build()
	main.player_id = "1001"
	main.party = [{"player_id": "1001", "display_name": "Astra"}]
	main.character_progression = {"level": 4}
	main._refresh_player_hud_identity()
	var state: Dictionary = main._health_bar.get_debug_state()
	_assert_eq("player hud identity name", str(state.get("character_name", "")), "Astra")
	_assert_eq("player hud identity level", int(state.get("level", 0)), 4)
	_assert_eq("player hud identity text", str(state.get("identity_text", "")), "Astra  Lv 4")
	main._health_bar.free()
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_character_stats_probability_values_use_percentages() -> void:
	var panel = CharacterStatsPanelScript.new()
	panel._build()
	panel.set_progression({
		"derived_stats": {
			"hit_chance": 0.5,
			"crit_chance": 0.06,
			"block_percent": 75,
		},
	})
	var state: Dictionary = panel.get_debug_state()
	var labels: Dictionary = state.get("derived_labels", {})
	_assert_eq("hit chance displays percent", str(labels.get("hit_chance", "")), "Hit chance  50%")
	_assert_eq("crit chance displays percent", str(labels.get("crit_chance", "")), "Crit chance  6%")
	_assert_eq("block chance displays percent", str(labels.get("block_percent", "")), "Block  75%")
	panel.free()


func _test_character_stats_window_chrome() -> void:
	var panel = CharacterStatsPanelScript.new()
	root.add_child(panel)
	panel._build()
	panel.ensure_display_visible()
	var state: Dictionary = panel.get_debug_state()
	var window: Dictionary = state.get("window", {})
	_assert_eq("stats window title", str(window.get("title", "")), "Character")
	_assert_true("stats window has close button", bool(window.get("close_visible", false)))
	_assert_true("stats window is draggable", bool(window.get("draggable", false)))
	panel.bot_drag_window_by(Vector2(32, 18))
	state = panel.get_debug_state()
	var moved_position: Dictionary = (state.get("window", {}) as Dictionary).get("position", {})
	_assert_eq("stats drag moved x", int(moved_position.get("x", 0)), 48)
	_assert_eq("stats drag moved y", int(moved_position.get("y", 0)), 136)
	panel.bot_drag_window_by(Vector2(-10000, -10000))
	state = panel.get_debug_state()
	var clamped_position: Dictionary = (state.get("window", {}) as Dictionary).get("position", {})
	_assert_eq("stats drag clamps x", int(clamped_position.get("x", -1)), 0)
	_assert_eq("stats drag clamps y", int(clamped_position.get("y", -1)), 0)
	panel.bot_click_close()
	_assert_true("stats close button hides panel", not panel.visible)
	panel.queue_free()


func _test_draggable_window_persists_layout() -> void:
	var old_path: String = DraggableWindowScript.layout_storage_path
	var old_force: bool = DraggableWindowScript.force_enable_persistence_for_tests
	DraggableWindowScript.layout_storage_path = "user://test_window_layout.cfg"
	DraggableWindowScript.force_enable_persistence_for_tests = true
	var first = DraggableWindowScript.new()
	root.add_child(first)
	first.custom_minimum_size = Vector2(200, 120)
	first.position = Vector2(24, 36)
	first.configure("Test", Vector2(180, 80))
	first.set_layout_key("test_panel")
	first.bot_drag_by(Vector2(40, 30))
	var saved_state: Dictionary = first.get_debug_state()
	_assert_true("window persistence enabled in forced test", bool(saved_state.get("persistence_enabled", false)))
	first.queue_free()

	var second = DraggableWindowScript.new()
	root.add_child(second)
	second.custom_minimum_size = Vector2(200, 120)
	second.position = Vector2(5, 5)
	second.configure("Test", Vector2(180, 80))
	second.set_layout_key("test_panel")
	var loaded_position: Dictionary = (second.get_debug_state().get("position", {}) as Dictionary)
	_assert_eq("window persisted x", int(loaded_position.get("x", 0)), 64)
	_assert_eq("window persisted y", int(loaded_position.get("y", 0)), 66)
	second.queue_free()
	DraggableWindowScript.layout_storage_path = old_path
	DraggableWindowScript.force_enable_persistence_for_tests = old_force


func _test_actionable_panels_autoclose_out_of_range() -> void:
	var main = _make_main()
	main.inventory_panel = InventoryPanelScript.new()
	main.shop_panel = ShopPanelScript.new()
	main.stash_panel = StashPanelScript.new()
	var vendor_node := Node3D.new()
	var stash_node := Node3D.new()
	main.entities_root.add_child(vendor_node)
	main.entities_root.add_child(stash_node)
	vendor_node.position = Vector3(1.0, 0.0, 0.0)
	stash_node.position = Vector3(1.0, 0.0, 0.0)
	main.entities["2001"] = {"node": vendor_node, "type": "interactable", "interactable_def_id": "town_vendor"}
	main.entities["2002"] = {"node": stash_node, "type": "interactable", "interactable_def_id": "town_stash"}

	main.inventory_panel.visible = true
	main.shop_panel.visible = true
	main.shop_panel.shop_entity_id = "2001"
	main._sync_actionable_panel_reach()
	_assert_true("shop remains visible while in range", main.shop_panel.visible)
	_assert_true("inventory remains visible while shop in range", main.inventory_panel.visible)
	vendor_node.position = Vector3(5.0, 0.0, 0.0)
	main._sync_actionable_panel_reach()
	_assert_true("shop closes out of range", not main.shop_panel.visible)
	_assert_true("inventory closes with out-of-range shop", not main.inventory_panel.visible)

	main.inventory_panel.visible = true
	main.stash_panel.visible = true
	main.stash_panel.stash_entity_id = "2002"
	main._sync_actionable_panel_reach()
	_assert_true("stash remains visible while in range", main.stash_panel.visible)
	stash_node.position = Vector3(5.0, 0.0, 0.0)
	main._sync_actionable_panel_reach()
	_assert_true("stash closes out of range", not main.stash_panel.visible)
	_assert_true("inventory closes with out-of-range stash", not main.inventory_panel.visible)

	main.inventory_panel.free()
	main.shop_panel.free()
	main.stash_panel.free()
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_movement_closes_gameplay_panels() -> void:
	var main = _make_main()
	main.inventory_panel = InventoryPanelScript.new()
	main.shop_panel = ShopPanelScript.new()
	main.stash_panel = StashPanelScript.new()
	main.character_stats_panel = CharacterStatsPanelScript.new()
	main.skills_panel = SkillsPanelScript.new()
	main.character_info_panel = PanelContainer.new()
	main.waypoint_panel = PanelContainer.new()

	main.inventory_panel.visible = true
	main.shop_panel.visible = true
	main.stash_panel.visible = true
	main.character_stats_panel.visible = true
	main.skills_panel.visible = true
	main.character_info_panel.visible = true
	main.waypoint_panel.visible = true
	main._close_gameplay_panels_for_movement()

	_assert_true("movement closes inventory panel", not main.inventory_panel.visible)
	_assert_true("movement closes shop panel", not main.shop_panel.visible)
	_assert_true("movement closes stash panel", not main.stash_panel.visible)
	_assert_true("movement closes character stats panel", not main.character_stats_panel.visible)
	_assert_true("movement closes skills panel", not main.skills_panel.visible)
	_assert_true("movement closes character info panel", not main.character_info_panel.visible)
	_assert_true("movement closes waypoint panel", not main.waypoint_panel.visible)

	main.inventory_panel.free()
	main.shop_panel.free()
	main.stash_panel.free()
	main.character_stats_panel.free()
	main.skills_panel.free()
	main.character_info_panel.free()
	main.waypoint_panel.free()
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_local_player_model_front_faces_direction() -> void:
	var main = _make_main()
	main.character_visual = Node3D.new()
	main.player_anchor.add_child(main.character_visual)
	main._face_direction(Vector2(1.0, 0.0))
	var front: Vector3 = main.character_visual.transform.basis.z
	front.y = 0.0
	_assert_vec3("local player visual front faces east", front.normalized(), Vector3(1.0, 0.0, 0.0))
	main._face_direction(Vector2(0.0, 1.0))
	front = main.character_visual.transform.basis.z
	front.y = 0.0
	_assert_vec3("local player visual front faces south", front.normalized(), Vector3(0.0, 0.0, 1.0))
	main._face_direction(Vector2.ZERO)
	_assert_vec3("local player zero facing ignored", _vec2_as_vec3(main._last_facing_direction), Vector3(0.0, 0.0, 1.0))
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_multiple_remote_players_update_and_remove_independently() -> void:
	var main = _make_main()
	main._apply_snapshot({
		"server_tick": 1,
		"current_level": 0,
		"local_player_id": "1001",
		"party": [],
		"entities": [
			{"id": "1001", "type": "player", "position": {"x": 1.0, "y": 1.0}, "hp": 10, "max_hp": 10},
			{"id": "1002", "type": "player", "position": {"x": 2.0, "y": 2.0}, "hp": 10, "max_hp": 10},
			{"id": "1003", "type": "player", "position": {"x": 3.0, "y": 3.0}, "hp": 10, "max_hp": 10},
		],
		"inventory": [],
		"equipped": {},
		"hotbar": [],
		"character_progression": {},
	})
	main._apply_delta({
		"events": [],
		"changes": [
			{"op": "entity_update", "entity": {"id": "1003", "type": "player", "position": {"x": 8.0, "y": 9.0}, "hp": 7, "max_hp": 10}},
		],
	})
	_assert_vec3("second remote player authoritative position", (main.entities["1003"]["node"] as Node3D).position, Vector3(8.0, 0.0, 9.0))
	_assert_vec3("first remote player untouched", (main.entities["1002"]["node"] as Node3D).position, Vector3(2.0, 0.0, 2.0))
	main._apply_delta({"events": [], "changes": [{"op": "entity_remove", "entity_id": "1003"}]})
	_assert_true("second remote removed", not main.entities.has("1003"))
	_assert_true("first remote remains", main.entities.has("1002"))
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_path_reject_clears_held_click_state() -> void:
	var main = _make_main()
	main._sustained_click.begin_from_pick({"kind": "monster", "target_id": "2001"})
	main.pending_interactable_action = {"target_id": "3001", "interactable_def_id": "stairs_down"}
	main.pending_action_targets["msg-no-path"] = {"target_id": "2001"}
	main._handle_intent_rejected({"rejected_message_id": "msg-no-path", "reason": "no_path"})
	_assert_true("no_path clears sustained click", not main._sustained_click.active)
	_assert_true("no_path clears pending interactable action", main.pending_interactable_action.is_empty())
	_assert_true("no_path removes pending action target", not main.pending_action_targets.has("msg-no-path"))
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_capacity_reject_shows_bag_full_unequip_message() -> void:
	var main = _make_main()
	main._handle_intent_rejected({"rejected_message_id": "msg-capacity", "reason": "capacity_would_overflow"})
	_assert_eq("capacity overflow hint text", main._last_inventory_feedback_text, MainScript.BAG_FULL_CANT_UNEQUIP_TEXT)
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_no_mana_reject_shows_floating_text() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main.player_anchor.position = Vector3(2.0, 0.0, 3.0)
	main.client_settings = ClientSettingsScript.new()
	main.damage_numbers_layer = CanvasLayer.new()
	main._camera = Camera3D.new()
	root.add_child(main.damage_numbers_layer)
	root.add_child(main._camera)
	main._camera.look_at_from_position(Vector3(2.0, 12.0, 13.0), main.player_anchor.position, Vector3.UP)
	main._handle_intent_rejected({"rejected_message_id": "msg-no-mana", "reason": "not_enough_mana"})
	var numbers := main._bot_damage_numbers()
	_assert_eq("no mana floating text count", numbers.size(), 1)
	_assert_eq("no mana floating text", str((numbers[0] as Dictionary).get("text", "")), MainScript.NO_MANA_TEXT)
	_assert_eq("no mana floating text variant", str((numbers[0] as Dictionary).get("variant", "")), "mana")
	main.damage_numbers_layer.queue_free()
	main._camera.queue_free()
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_skill_cooldown_reject_shows_floating_text() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main.player_anchor.position = Vector3(2.0, 0.0, 3.0)
	main.client_settings = ClientSettingsScript.new()
	main.damage_numbers_layer = CanvasLayer.new()
	main._camera = Camera3D.new()
	root.add_child(main.damage_numbers_layer)
	root.add_child(main._camera)
	main._camera.look_at_from_position(Vector3(2.0, 12.0, 13.0), main.player_anchor.position, Vector3.UP)
	main._apply_delta({"events": [{"event_type": "skill_cooldown_rejected", "entity_id": "1001", "skill_id": "heal", "reason": "skill_on_cooldown"}], "changes": []})
	var numbers := main._bot_damage_numbers()
	_assert_eq("cooldown floating text count", numbers.size(), 1)
	_assert_eq("cooldown floating text", str((numbers[0] as Dictionary).get("text", "")), "ON COOLDOWN")
	_assert_eq("cooldown floating text variant", str((numbers[0] as Dictionary).get("variant", "")), "skill_reject")
	main.damage_numbers_layer.queue_free()
	main._camera.queue_free()
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_player_healed_spawns_heal_rain() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main.player_anchor.position = Vector3(2.0, 0.0, 3.0)
	main.client_settings = ClientSettingsScript.new()
	main.damage_numbers_layer = CanvasLayer.new()
	main._camera = Camera3D.new()
	root.add_child(main.damage_numbers_layer)
	root.add_child(main._camera)
	main._camera.look_at_from_position(Vector3(2.0, 12.0, 13.0), main.player_anchor.position, Vector3.UP)
	main._apply_delta({"events": [{"event_type": "player_healed", "entity_id": "1001", "heal": 4}], "changes": []})
	var rain_count := 0
	for child in main.get_children():
		if child.get_script() == HealRainEffectScript:
			rain_count += 1
			_assert_float("heal rain radius", float(child.radius), MainScript.HEAL_RAIN_RADIUS)
	_assert_eq("heal rain count", rain_count, 1)
	var numbers := main._bot_damage_numbers()
	_assert_eq("heal floating text count", numbers.size(), 1)
	_assert_eq("heal floating text", str((numbers[0] as Dictionary).get("text", "")), "+4")
	main.damage_numbers_layer.queue_free()
	main._camera.queue_free()
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.queue_free()


func _test_local_attack_range_uses_equipped_reach() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main.player_anchor.position = Vector3.ZERO
	var near := Node3D.new()
	near.position = Vector3(1.95, 0.0, 0.0)
	var far := Node3D.new()
	far.position = Vector3(2.10, 0.0, 0.0)
	main.entities_root.add_child(near)
	main.entities_root.add_child(far)
	main.inventory = [{"item_instance_id": "sword_1", "item_def_id": "rusty_sword"}]
	main.equipped = {"main_hand": "sword_1"}
	main.entities["near"] = {"node": near, "type": "monster", "hp": 3}
	main.entities["far"] = {"node": far, "type": "monster", "hp": 3}
	_assert_true("near monster is inside equipped sword reach", main._target_in_local_attack_range("near"))
	_assert_true("far monster is outside equipped sword reach", not main._target_in_local_attack_range("far"))
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_character_bar_opens_stats_panel() -> void:
	var main = MainScript.new()
	main.character_stats_panel = CharacterStatsPanelScript.new()
	main.skills_panel = SkillsPanelScript.new()
	main.inventory_panel = InventoryPanelScript.new()
	main.character_bar = CharacterBarScript.new()
	root.add_child(main.character_stats_panel)
	root.add_child(main.skills_panel)
	root.add_child(main.inventory_panel)
	root.add_child(main.character_bar)
	main.character_stats_panel._build()
	main.character_bar._build()
	main.character_bar.open_character_requested.connect(main._open_character_panel_from_bar)
	main.character_progression = {
		"level": 2,
		"experience": 30,
		"unspent_stat_points": 2,
		"base_stats": {"str": 5, "agi": 5, "vit": 5, "magic": 5},
		"derived_stats": {"max_hp": 10},
	}
	main._refresh_progression_ui()
	var character_bar_state: Dictionary = main.character_bar.get_debug_state()
	_assert_eq("character bar is the stats button", str(character_bar_state.get("tooltip_text", "")), "Character")
	_assert_true("character bar stat badge visible with points", bool(character_bar_state.get("upgrade_badge_visible", false)))
	_assert_eq("character bar stat badge text", str(character_bar_state.get("upgrade_badge_text", "")), "+")
	main.skills_panel.visible = true
	main.inventory_panel.visible = true
	main.character_bar.open_slot()
	_assert_true("character bar opens stats panel", main.character_stats_panel.visible)
	main.character_bar.queue_free()
	main.character_stats_panel.free()
	main.skills_panel.free()
	main.inventory_panel.free()
	main.free()


func _test_basic_attack_cooldown_uses_derived_interval() -> void:
	var main = MainScript.new()
	main.character_progression = {"derived_stats": {"attack_interval_ticks": 7}}
	_assert_float("basic attack cooldown uses derived attack interval", main._basic_attack_cooldown_seconds(), 0.7)
	main.character_progression = {"derived_stats": {"attack_interval_ticks": 0}}
	_assert_float("basic attack cooldown falls back to default interval", main._basic_attack_cooldown_seconds(), 2.0)
	main.free()


func _test_skill_function_key_selects_right_click_skill() -> void:
	var main = MainScript.new()
	main.skill_bar = SkillBarScript.new()
	root.add_child(main.skill_bar)
	main.skill_progression = {
		"unspent_skill_points": 0,
		"skills": [
			{"skill_id": "magic_bolt", "rank": 1, "max_rank": 5, "can_spend": false},
			{"skill_id": "heal", "rank": 1, "max_rank": 5, "can_spend": false},
		],
	}
	var event := InputEventKey.new()
	event.keycode = KEY_F1
	_assert_eq("F1 maps to skill slot 0", main._skill_function_key_slot(event), 0)
	_assert_true("assign F1 to magic bolt", main._assign_skill_function_key(0, "magic_bolt"))
	_assert_eq("F1 binding stored", str(main.skill_function_keys[0]), "magic_bolt")
	_assert_eq("ranked binding selects immediately", main.right_click_skill_id, "magic_bolt")
	_assert_eq("ranked binding updates skill bar", str(main.skill_bar.get_debug_state().get("skill_id", "")), "magic_bolt")
	_assert_true("assign F2 to heal", main._assign_skill_function_key(1, "heal"))
	_assert_eq("F2 binding selects heal", main.right_click_skill_id, "heal")
	_assert_eq("F2 binding updates skill bar", str(main.skill_bar.get_debug_state().get("skill_id", "")), "heal")
	main.right_click_skill_id = ""
	_assert_true("pressing F1 selects right click skill", main._select_right_click_skill_from_function_key(0))
	_assert_eq("right click skill selected", main.right_click_skill_id, "magic_bolt")
	_assert_eq("pressing F1 updates skill bar", str(main.skill_bar.get_debug_state().get("skill_id", "")), "magic_bolt")

	main.right_click_skill_id = ""
	main.skill_progression = {
		"unspent_skill_points": 1,
		"skills": [{"skill_id": "magic_bolt", "rank": 0, "max_rank": 5, "can_spend": true}],
	}
	_assert_true("unranked skill can still be bound", main._assign_skill_function_key(1, "magic_bolt"))
	_assert_true("unranked binding cannot select right click", not main._select_right_click_skill_from_function_key(1))
	_assert_eq("right click stays empty for unranked skill", main.right_click_skill_id, "")
	main._apply_skill_bindings({
		"function_keys": ["heal", "magic_bolt", "", "", "", "", "", ""],
		"right_click_skill_id": "heal",
	})
	_assert_eq("snapshot restores F1 skill binding", str(main.skill_function_keys[0]), "heal")
	_assert_eq("snapshot restores right click skill", main.right_click_skill_id, "heal")
	main.skill_bar.queue_free()
	main.free()


func _test_learned_skill_auto_selects_right_click() -> void:
	var main = MainScript.new()
	main.skill_progression = {
		"unspent_skill_points": 0,
		"skills": [{"skill_id": "magic_bolt", "rank": 1, "max_rank": 5, "can_spend": false}],
	}
	main._refresh_skill_ui()
	_assert_eq("learned only active skill auto-selects right click", main.right_click_skill_id, "magic_bolt")
	main.skill_progression = {
		"unspent_skill_points": 1,
		"skills": [{"skill_id": "magic_bolt", "rank": 0, "max_rank": 5, "can_spend": true}],
	}
	main._refresh_skill_ui()
	_assert_eq("unlearned skill clears right click", main.right_click_skill_id, "")
	main.free()


func _test_skill_cast_payload_uses_direction_without_nearest_fallback() -> void:
	var main = _make_main()
	var monster := Node3D.new()
	monster.position = Vector3(2.0, 0.0, 0.0)
	main.entities_root.add_child(monster)
	main.entities["2001"] = {"node": monster, "type": "monster", "hp": 10}
	main.monster_ids = ["2001"]
	main.predicted_pos = Vector3.ZERO
	var payload := main._skill_cast_payload("magic_bolt", "", Vector2(0.0, -3.0), false)
	_assert_eq("right click direction payload has no target", payload.has("target_id"), false)
	var direction: Dictionary = payload.get("direction", {})
	_assert_float("right click direction x", float(direction.get("x", 99.0)), 0.0)
	_assert_float("right click direction y", float(direction.get("y", 99.0)), -1.0)
	var targeted := main._skill_cast_payload("magic_bolt", "", Vector2.ZERO, true)
	_assert_eq("skill slot fallback can still target nearest", str(targeted.get("target_id", "")), "2001")
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_remote_player_delta_and_remove() -> void:
	var main = _make_main()
	main._apply_snapshot({
		"server_tick": 1,
		"current_level": 0,
		"local_player_id": "1001",
		"party": [],
		"entities": [
			{"id": "1001", "type": "player", "position": {"x": 2.0, "y": 3.0}, "hp": 10, "max_hp": 10},
			{"id": "1002", "type": "player", "position": {"x": 4.0, "y": 5.0}, "hp": 10, "max_hp": 10},
		],
		"inventory": [],
		"equipped": {},
		"hotbar": [],
		"character_progression": {},
	})
	main._apply_delta({
		"events": [
			{"event_type": "player_damaged", "entity_id": "1002", "target_entity_id": "1002", "source_entity_id": "1003", "damage": 2},
		],
		"changes": [
			{"op": "entity_update", "entity": {"id": "1002", "type": "player", "position": {"x": 6.0, "y": 7.0}, "hp": 8, "max_hp": 10}},
		],
	})
	_assert_vec3("remote player authoritative position", (main.entities["1002"]["node"] as Node3D).position, Vector3(6.0, 0.0, 7.0))
	_assert_eq("remote hp updated", int(main.entities["1002"].get("hp", 0)), 8)
	_assert_eq("remote hit reaction", str(main.entities["1002"]["reaction"].get_debug_state().get("last_reaction", "")), "hit")
	_assert_vec3("local prediction untouched by remote delta", main.predicted_pos, Vector3(2.0, 0.0, 3.0))
	main._apply_delta({
		"events": [
			{"event_type": "player_killed", "entity_id": "1002", "target_entity_id": "1002", "source_entity_id": "1003", "damage": 8},
		],
		"changes": [],
	})
	_assert_true("remote death terminal reaction", bool(main.entities["1002"]["reaction"].get_debug_state().get("terminal", false)))
	main._apply_delta({"events": [], "changes": [{"op": "entity_remove", "entity_id": "1002"}]})
	_assert_true("remote player removed", not main.entities.has("1002"))
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _assert_true(name: String, cond: bool) -> void:
	if cond:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL: %s" % name)


func _assert_eq(name: String, got, want) -> void:
	if got == want:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL: %s got=%s want=%s" % [name, str(got), str(want)])


func _assert_vec3(name: String, got: Vector3, want: Vector3) -> void:
	if got.distance_to(want) <= 0.0001:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL: %s got=%s want=%s" % [name, str(got), str(want)])


func _assert_float(name: String, got: float, want: float) -> void:
	if absf(got - want) <= 0.0001:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL: %s got=%s want=%s" % [name, str(got), str(want)])


func _vec2_as_vec3(v: Vector2) -> Vector3:
	return Vector3(v.x, 0.0, v.y)
