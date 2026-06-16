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
		"assert_inventory_panel_details":
			if not runner._assert_inventory_panel_details(step, state):
				return _handled(false)
			if _has_set_collection_expectation(step):
				return _handled(_set_collection_matches(runner, step, state))
			return _handled(true)
	return {"handled": false, "ok": true}


static func _handled(ok: bool) -> Dictionary:
	return {"handled": true, "ok": ok}


static func _has_set_collection_expectation(step: Dictionary) -> bool:
	for key in ["set_collection_set", "set_owned_count", "set_equipped_count", "set_piece_name", "set_piece_state", "set_bonus_required", "set_bonus_active"]:
		if step.has(key):
			return true
	return false


static func _set_collection_matches(runner, step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("inventory_panel", {})
	var collection: Dictionary = panel.get("set_collection", {})
	for set_state in collection.get("sets", []):
		var rec := set_state as Dictionary
		if step.has("set_collection_set") and str(rec.get("set_name", "")) != str(step.get("set_collection_set", "")):
			continue
		if step.has("set_owned_count") and int(rec.get("owned_count", -1)) != int(step.get("set_owned_count", 0)):
			continue
		if step.has("set_equipped_count") and int(rec.get("equipped_count", -1)) != int(step.get("set_equipped_count", 0)):
			continue
		if step.has("set_piece_name") and not _piece_matches(rec, step):
			continue
		if step.has("set_bonus_required") and not _bonus_matches(rec, step):
			continue
		return true
	runner._fail("assert_inventory_panel_details set collection failed: want=%s collection=%s step=%d scenario=%s" % [
		str(step), str(collection), runner._step_index, str(runner.scenario.get("id", "?"))
	])
	return false


static func _piece_matches(set_state: Dictionary, step: Dictionary) -> bool:
	for piece in set_state.get("pieces", []):
		var rec := piece as Dictionary
		if str(rec.get("name", "")) != str(step.get("set_piece_name", "")):
			continue
		return not step.has("set_piece_state") or str(rec.get("state", "")) == str(step.get("set_piece_state", ""))
	return false


static func _bonus_matches(set_state: Dictionary, step: Dictionary) -> bool:
	for bonus in set_state.get("bonuses", []):
		var rec := bonus as Dictionary
		if int(rec.get("required_pieces", 0)) != int(step.get("set_bonus_required", 0)):
			continue
		return not step.has("set_bonus_active") or bool(rec.get("active", false)) == bool(step.get("set_bonus_active", false))
	return false
