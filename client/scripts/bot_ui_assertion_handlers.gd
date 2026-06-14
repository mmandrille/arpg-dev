class_name BotUiAssertionHandlers
extends RefCounted


static func try_evaluate(runner, step: Dictionary, stype: String, state: Dictionary) -> Dictionary:
	match stype:
		"assert_panel_visible":
			var want := bool(step.get("visible", true))
			var got := bool(state.get("inventory_panel_visible", false))
			if want != got:
				runner._fail("assert_panel_visible failed: want=%s got=%s step=%d scenario=%s" % [
					want, got, runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return _handled(false)
			return _handled(true)
		"assert_main_menu_visible":
			return _handled(runner._assert_bool_state("assert_main_menu_visible", "main_menu_visible", step, state))
		"assert_main_menu_actions":
			return _handled(runner._assert_main_menu_actions(step, state))
		"assert_character_panel_visible":
			return _handled(runner._assert_bool_state("assert_character_panel_visible", "character_panel_visible", step, state))
		"assert_character_panel":
			return _handled(runner._assert_character_panel(step, state))
		"assert_create_game_type":
			return _handled(runner._assert_create_game_type(step, state))
		"assert_current_session":
			return _handled(runner._assert_current_session(step, state))
		"assert_multiplayer_panel_visible":
			return _handled(runner._assert_bool_state("assert_multiplayer_panel_visible", "multiplayer_panel_visible", step, state))
		"assert_multiplayer_session_rows":
			return _handled(runner._assert_multiplayer_session_rows(step, state))
		"assert_multiplayer_filter":
			return _handled(runner._assert_multiplayer_filter(step, state))
		"assert_settings_panel_visible":
			return _handled(runner._assert_bool_state("assert_settings_panel_visible", "settings_panel_visible", step, state))
		"assert_pause_menu_visible":
			return _handled(runner._assert_bool_state("assert_pause_menu_visible", "pause_menu_visible", step, state))
		"assert_character_stats_panel_visible":
			return _handled(runner._assert_bool_state("assert_character_stats_panel_visible", "character_stats_panel_visible", step, state))
		"assert_character_info_panel_visible":
			return _handled(runner._assert_bool_state("assert_character_info_panel_visible", "character_info_panel_visible", step, state))
		"assert_character_info":
			return _handled(runner._assert_character_info(step, state))
	return {"handled": false, "ok": true}


static func _handled(ok: bool) -> Dictionary:
	return {"handled": true, "ok": ok}
