# Unit tests for local/remote co-op snapshot handling (v33).
# Run via: godot --headless --path client --script res://tests/test_coop_client.gd
extends SceneTree

const MainScript := preload("res://scripts/main.gd")
const NetClientScript := preload("res://scripts/net_client.gd")
const CharacterSelectPanelScript := preload("res://scripts/character_select_panel.gd")
const SettingsPanelScript := preload("res://scripts/settings_panel.gd")

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
	_test_loss_popup_shows_for_dead_local_player()
	_test_dead_character_rows_are_disabled()
	_test_character_panel_modes_for_v45()
	_test_settings_panel_create_game_type_sync()

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
	_assert_true("forced create hides empty label", not bool(forced.get("empty_visible", true)))
	panel.show_choose_or_create([
		{"character_id": "char_live", "name": "Alive", "created_at": "", "dead": false},
	], "Choose Character")
	var choose := panel.get_debug_state()
	_assert_eq("choose mode", str(choose.get("mode", "")), "choose_or_create")
	_assert_eq("choose title", str(choose.get("title", "")), "Choose Character")
	_assert_true("choose keeps create affordance", bool(choose.get("create_button_visible", false)))
	_assert_eq("choose character count", (choose.get("characters", []) as Array).size(), 1)
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


func _vec2_as_vec3(v: Vector2) -> Vector3:
	return Vector3(v.x, 0.0, v.y)
