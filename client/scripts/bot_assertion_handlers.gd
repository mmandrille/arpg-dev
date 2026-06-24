class_name BotAssertionHandlers
extends RefCounted

const BotUiAssertionHandlersScript := preload("res://scripts/bot_ui_assertion_handlers.gd")
const BotQuestJournalAssertionsScript := preload("res://scripts/bot_quest_journal_assertions.gd")
const BotEliteObjectiveAssertionsScript := preload("res://scripts/bot_elite_objective_assertions.gd")
const BotEliteObjectiveMinimapAssertionsScript := preload("res://scripts/bot_elite_objective_minimap_assertions.gd")
const BotMercenaryPanelAssertionsScript := preload("res://scripts/bot_mercenary_panel_assertions.gd")
const BotMarketBadgeAssertionsScript := preload("res://scripts/bot_market_badge_assertions.gd")
const FLOAT_BOUND_EPSILON := 0.00001


static func evaluate(runner, step: Dictionary, stype: String, state: Dictionary) -> bool:
	var ui_result := BotUiAssertionHandlersScript.try_evaluate(runner, step, stype, state)
	if bool(ui_result.get("handled", false)):
		return bool(ui_result.get("ok", false))
	match stype:
		"assert_character_progression":
			return runner._assert_character_progression(step, state)
		"assert_stat_button_enabled":
			return runner._assert_stat_button_enabled(step, state)
		"assert_xp_bar":
			return runner._assert_xp_bar(step, state)
		"assert_skills_panel_visible":
			return runner._assert_bool_state("assert_skills_panel_visible", "skills_panel_visible", step, state)
		"assert_skill_progression":
			return runner._assert_skill_progression(step, state)
		"assert_skill_button_enabled":
			return runner._assert_skill_button_enabled(step, state)
		"assert_skill_bar":
			if not runner._skill_bar_matches(step, state):
				runner._fail("assert_skill_bar failed: want=%s got=%s step=%d scenario=%s" % [
					str(runner._skill_bar_expectation(step)), str(state.get("skill_bar", {})), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_fog_of_war":
			return _assert_fog_of_war(runner, step, state)
		"assert_discovery_minimap":
			return _assert_discovery_minimap(runner, step, state)
		"assert_floating_combat_text_enabled":
			return runner._assert_bool_value("assert_floating_combat_text_enabled", step, bool(state.get("floating_combat_text_enabled", false)), bool(step.get("enabled", true)))
		"assert_damage_number":
			if not runner._damage_number_matches(step, state):
				runner._fail("assert_damage_number failed: want=%s damage_numbers=%s step=%d scenario=%s" % [
					str(step), str(state.get("damage_numbers", [])), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_no_damage_number":
			var numbers: Array = state.get("damage_numbers", [])
			if not numbers.is_empty():
				runner._fail("assert_no_damage_number failed: damage_numbers=%s step=%d scenario=%s" % [
					str(numbers), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_entity_reaction":
			if not runner._presentation_matches(step, state):
				runner._fail("assert_entity_reaction failed: want=%s local=%s entities=%s step=%d scenario=%s" % [
					str(step), str(state.get("local_player_presentation", {})),
					str(state.get("entities_presentation_debug", [])), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_movement_visual_smoothing":
			if not movement_visual_smoothing_matches(step, state):
				runner._fail("assert_movement_visual_smoothing failed: want=%s got=%s step=%d scenario=%s" % [
					str(step), str(state.get("movement_visual_smoothing", {})), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_command_retarget_grace":
			if not command_retarget_grace_matches(step, state):
				runner._fail("assert_command_retarget_grace failed: want=%s got=%s step=%d scenario=%s" % [
					str(step), str(state.get("command_retarget_grace", {})), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_melee_lunge":
			if not melee_lunge_matches(step, state):
				runner._fail("assert_melee_lunge failed: want=%s got=%s step=%d scenario=%s" % [
					str(step), str(_melee_lunge_state(state)), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_wall_layout":
			if not runner._wall_layout_matches(step, state):
				runner._fail("assert_wall_layout failed: want=%s wall_count=%d generated=%d non_perimeter=%d level=%d walls=%s step=%d scenario=%s" % [
					str(step),
					int(state.get("wall_count", 0)),
					int(state.get("generated_wall_count", 0)),
					int(state.get("non_perimeter_wall_count", 0)),
					int(state.get("current_level", 0)),
					str(state.get("walls", [])),
					runner._step_index,
					str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_shop_panel_visible":
			return runner._assert_bool_state("assert_shop_panel_visible", "shop_panel_visible", step, state)
		"assert_shop_offer_count":
			return runner._assert_shop_offer_count(step, state)
		"assert_shop_buy_button":
			return runner._assert_shop_buy_button(step, state)
		"assert_shop_reroll_button":
			return runner._assert_shop_reroll_button(step, state)
		"assert_shop_sell_rows":
			return runner._assert_shop_sell_rows(step, state)
		"assert_shop_offer_details":
			return runner._assert_shop_offer_details(step, state)
		"assert_shop_sell_details":
			return runner._assert_shop_sell_details(step, state)
		"assert_stash_panel_visible":
			return runner._assert_bool_state("assert_stash_panel_visible", "stash_panel_visible", step, state)
		"assert_stash_item_count":
			return runner._assert_stash_item_count(step, state)
		"assert_stash_gold":
			return runner._assert_stash_gold(step, state)
		"assert_stash_filter":
			return runner._assert_stash_filter(step, state)
		"assert_market_panel_visible":
			return runner._assert_bool_state("assert_market_panel_visible", "market_panel_visible", step, state)
		"assert_market_board_badges":
			if not BotMarketBadgeAssertionsScript.matches(step, state):
				runner._fail("assert_market_board_badges failed: want=%s badges=%s step=%d scenario=%s" % [
					str(step), str(state.get("market_board_badges", {})), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_market_listing_rows":
			return runner._assert_market_listing_rows(step, state)
		"assert_market_offer_rows":
			return runner._assert_market_offer_rows(step, state)
		"assert_bishop_panel_visible":
			return runner._assert_bool_state("assert_bishop_panel_visible", "bishop_panel_visible", step, state)
		"assert_bishop_panel":
			if not runner._bishop_panel_matches(step, state):
				runner._fail("assert_bishop_panel failed: want=%s panel=%s step=%d scenario=%s" % [
					str(step), str(state.get("bishop_panel", {})), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_mercenary_panel_visible":
			return runner._assert_bool_state("assert_mercenary_panel_visible", "mercenary_panel_visible", step, state)
		"assert_mercenary_panel":
			if not BotMercenaryPanelAssertionsScript.matches(step, state):
				runner._fail("assert_mercenary_panel failed: want=%s panel=%s companion_bar=%s step=%d scenario=%s" % [
					str(step), str(state.get("mercenary_panel", {})), str(state.get("companion_bar", {})), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_blacksmith_panel_visible":
			return runner._assert_bool_state("assert_blacksmith_panel_visible", "blacksmith_panel_visible", step, state)
		"assert_blacksmith_panel":
			if not runner._blacksmith_panel_matches(step, state):
				runner._fail("assert_blacksmith_panel failed: want=%s panel=%s step=%d scenario=%s" % [
					str(step), str(state.get("blacksmith_panel", {})), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_boss_health_bar":
			return runner._assert_boss_health_bar(step, state)
		"assert_boss_reward_status":
			var want_status := str(step.get("text", ""))
			var got_status := str(state.get("boss_reward_status", ""))
			if want_status != got_status:
				runner._fail("assert_boss_reward_status failed: want=%s got=%s step=%d scenario=%s" % [
					want_status, got_status, runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_resource_wallet_panel":
			return _assert_resource_wallet_panel(runner, step, state)
		"assert_audio_state":
			return _assert_audio_state(runner, step, state)
		"assert_remote_player_count":
			if not runner._remote_player_count_matches(step, state):
				runner._fail("assert_remote_player_count failed: want=%s remote_player_ids=%s step=%d scenario=%s" % [
					str(step), str(state.get("remote_player_ids", [])), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_quest_journal":
			if not BotQuestJournalAssertionsScript.matches(step, state):
				runner._fail("assert_quest_journal failed: want=%s panel=%s step=%d scenario=%s" % [
					str(step), str(state.get("quest_journal_panel", {})), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_elite_objective_tracker":
			if not BotEliteObjectiveAssertionsScript.matches(step, state):
				runner._fail("assert_elite_objective_tracker failed: want=%s tracker=%s step=%d scenario=%s" % [
					str(step), str(state.get("elite_objective_tracker", {})), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_elite_objective_minimap":
			if not BotEliteObjectiveMinimapAssertionsScript.matches(step, state):
				runner._fail("assert_elite_objective_minimap failed: want=%s minimap=%s step=%d scenario=%s" % [
					str(step), str(state.get("elite_objective_minimap", {})), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_session_changed":
			var remembered_session := str(runner._memory.get("session_id", ""))
			var current_session := str(state.get("current_session_id", ""))
			if current_session == "" or remembered_session == "" or current_session == remembered_session:
				runner._fail("assert_session_changed failed: remembered=%s current=%s step=%d scenario=%s" % [
					remembered_session, current_session, runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_player_position_unchanged":
			var remembered_pos: Dictionary = runner._memory.get("player_pos", {})
			var current_pos: Dictionary = state.get("player_pos", {})
			var tolerance := float(step.get("tolerance", 0.01))
			var dx := float(current_pos.get("x", 0.0)) - float(remembered_pos.get("x", 0.0))
			var dz := float(current_pos.get("z", 0.0)) - float(remembered_pos.get("z", 0.0))
			if sqrt(dx * dx + dz * dz) > tolerance:
				runner._fail("assert_player_position_unchanged failed: remembered=%s current=%s step=%d scenario=%s" % [
					str(remembered_pos), str(current_pos), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_waypoint_panel_visible":
			var want := bool(step.get("visible", true))
			var got := bool(state.get("waypoint_panel_visible", false))
			if want != got:
				runner._fail("assert_waypoint_panel_visible failed: want=%s got=%s step=%d scenario=%s" % [
					want, got, runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_equipped":
			var slot := str(step.get("slot", "main_hand"))
			var eq: Dictionary = state.get("equipped", {})
			var val = eq.get(slot, null)
			if val == null or str(val) == "":
				runner._fail("assert_equipped failed: slot=%s equipped=%s step=%d scenario=%s" % [
					slot, str(eq), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_unequipped":
			var slot := str(step.get("slot", "main_hand"))
			var eq: Dictionary = state.get("equipped", {})
			var val = eq.get(slot, null)
			if val != null and str(val) != "" and str(val) != "null":
				runner._fail("assert_unequipped failed: slot=%s still has %s step=%d scenario=%s" % [
					slot, str(val), runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_inventory_missing":
			var def_id := str(step.get("item_def_id", ""))
			var inv: Array = state.get("inventory", [])
			for item in inv:
				if str(item.get("item_def_id", "")) == def_id:
					runner._fail("assert_inventory_missing failed: %s still in inventory step=%d scenario=%s" % [
						def_id, runner._step_index, str(runner.scenario.get("id", "?"))
					])
					return false
			return true
		"assert_loot_presentation":
			var def_id := str(step.get("item_def_id", ""))
			var presentations: Dictionary = state.get("loot_presentations", {})
			if not bool(presentations.get(def_id, false)):
				runner._fail("assert_loot_presentation failed: %s missing from loot presentation state step=%d scenario=%s" % [
					def_id, runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_inventory_presentation":
			var def_id := str(step.get("item_def_id", ""))
			var panel: Dictionary = state.get("inventory_panel", {})
			var presentations: Dictionary = panel.get("item_presentations", {})
			if not bool(presentations.get(def_id, false)):
				runner._fail("assert_inventory_presentation failed: %s missing from inventory presentation state step=%d scenario=%s" % [
					def_id, runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_hotbar_assigned":
			if not runner._hotbar_slot_matches(step, state):
				runner._fail("assert_hotbar_assigned failed: slot=%d item_def_id=%s step=%d scenario=%s" % [
					int(step.get("slot_index", -1)), str(step.get("item_def_id", "")),
					runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_hotbar_capacity":
			if not runner._hotbar_capacity_matches(step, state):
				var bar: Dictionary = state.get("consumable_bar", {})
				runner._fail("assert_hotbar_capacity failed: got=%d want=%d step=%d scenario=%s" % [
					int(bar.get("hotbar_capacity", -1)), int(step.get("equals", -1)),
					runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_hotbar_slot_disabled":
			var disabled_slot_index := int(step.get("slot_index", -1))
			var bar: Dictionary = state.get("consumable_bar", {})
			var cap := int(bar.get("hotbar_capacity", 2))
			if disabled_slot_index < cap:
				runner._fail("assert_hotbar_slot_disabled failed: slot=%d capacity=%d step=%d scenario=%s" % [
					disabled_slot_index, cap, runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_inventory_capacity":
			return runner._assert_inventory_capacity(step, state)
		"assert_bag_grid":
			return runner._assert_bag_grid(step, state)
		"assert_paper_doll_layout":
			return runner._assert_paper_doll_layout(step, state)
		"assert_inventory_panel_details":
			return runner._assert_inventory_panel_details(step, state)
		"assert_player_hp":
			var want_hp := int(step.get("equals", -1))
			var got_hp := int(state.get("player_hp", -1))
			if got_hp != want_hp:
				runner._fail("assert_player_hp failed: want=%d got=%d step=%d scenario=%s" % [
					want_hp, got_hp, runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_inventory_count":
			var def_id := str(step.get("item_def_id", ""))
			var want := int(step.get("equals", 0))
			var got := int(runner._inventory_count(state, def_id))
			if got != want:
				runner._fail("assert_inventory_count failed: %s want=%d got=%d step=%d scenario=%s" % [
					def_id, want, got, runner._step_index, str(runner.scenario.get("id", "?"))
				])
				return false
			return true
		"assert_camera_mode":
			return _assert_camera_mode(runner, step, state)
	return true


static func fog_of_war_matches(step: Dictionary, state: Dictionary) -> bool:
	return _fog_of_war_mismatch(step, state) == ""


static func _assert_fog_of_war(runner, step: Dictionary, state: Dictionary) -> bool:
	var mismatch := _fog_of_war_mismatch(step, state)
	if mismatch != "":
		runner._fail("assert_fog_of_war failed: %s step=%d scenario=%s" % [
			mismatch, runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	return true


static func _fog_of_war_mismatch(step: Dictionary, state: Dictionary) -> String:
	var fog: Dictionary = state.get("fog_of_war", {})
	for key in ["enabled", "active", "organic_edge_enabled", "hero_centered_falloff", "world_space_visibility", "perspective_camera"]:
		if step.has(key) and bool(fog.get(key, false)) != bool(step.get(key, true)):
			return "%s want=%s got=%s fog=%s" % [key, str(step.get(key, true)), str(fog.get(key, null)), str(fog)]
	for key in ["light_radius", "gloom_radius", "organic_edge_px", "darkness_feather_px", "shadow_core_alpha", "shadow_gloom_alpha"]:
		var min_key := "%s_min" % key
		var max_key := "%s_max" % key
		if step.has(min_key) and float(fog.get(key, 0.0)) < float(step.get(min_key, 0.0)):
			return "%s want_min=%s got=%s fog=%s" % [key, str(step.get(min_key, 0.0)), str(fog.get(key, null)), str(fog)]
		if step.has(max_key) and float(fog.get(key, 0.0)) > float(step.get(max_key, 0.0)):
			return "%s want_max=%s got=%s fog=%s" % [key, str(step.get(max_key, 0.0)), str(fog.get(key, null)), str(fog)]
		if step.has(key) and abs(float(fog.get(key, -999999.0)) - float(step.get(key, 0.0))) > float(step.get("tolerance", 0.001)):
			return "%s want=%s got=%s fog=%s" % [key, str(step.get(key, 0.0)), str(fog.get(key, null)), str(fog)]
	for key in ["wall_count", "extra_occluder_count", "occluder_count", "shadow_count", "organic_edge_segments"]:
		var min_key := "%s_min" % key
		if step.has(min_key) and int(fog.get(key, 0)) < int(step.get(min_key, 0)):
			return "%s want_min=%s got=%s fog=%s" % [key, str(step.get(min_key, 0)), str(fog.get(key, null)), str(fog)]
		if step.has(key) and int(fog.get(key, -999999)) != int(step.get(key, 0)):
			return "%s want=%s got=%s fog=%s" % [key, str(step.get(key, 0)), str(fog.get(key, null)), str(fog)]
	return ""


static func _assert_discovery_minimap(runner, step: Dictionary, state: Dictionary) -> bool:
	var minimap: Dictionary = state.get("discovery_minimap", {})
	for key in ["visible", "toggle_visible", "has_pin", "has_quest_path"]:
		if step.has(key) and bool(minimap.get(key, false)) != bool(step.get(key, true)):
			runner._fail("assert_discovery_minimap failed: %s want=%s got=%s minimap=%s step=%d scenario=%s" % [
				key, str(step.get(key, true)), str(minimap.get(key, null)), str(minimap), runner._step_index, str(runner.scenario.get("id", "?"))
			])
			return false
	for key in ["display_mode"]:
		if step.has(key) and str(minimap.get(key, "")) != str(step.get(key, "")):
			runner._fail("assert_discovery_minimap failed: %s want=%s got=%s minimap=%s step=%d scenario=%s" % [
				key, str(step.get(key, "")), str(minimap.get(key, null)), str(minimap), runner._step_index, str(runner.scenario.get("id", "?"))
			])
			return false
	if step.has("session_key_present"):
		var has_session := str(minimap.get("session_key", "")) != ""
		if has_session != bool(step.get("session_key_present", true)):
			runner._fail("assert_discovery_minimap failed: session_key_present want=%s got=%s minimap=%s step=%d scenario=%s" % [
				str(step.get("session_key_present", true)), str(has_session), str(minimap), runner._step_index, str(runner.scenario.get("id", "?"))
			])
			return false
	for key in ["map_size_x", "map_size_y", "explored_count", "wall_count", "marker_count", "service_marker_count", "stairs_marker_count", "waypoint_marker_count", "objective_marker_count"]:
		var min_key := "%s_min" % key
		if step.has(min_key) and int(minimap.get(key, 0)) < int(step.get(min_key, 0)):
			runner._fail("assert_discovery_minimap failed: %s want_min=%s got=%s minimap=%s step=%d scenario=%s" % [
				key, str(step.get(min_key, 0)), str(minimap.get(key, null)), str(minimap), runner._step_index, str(runner.scenario.get("id", "?"))
			])
			return false
		if step.has(key) and int(minimap.get(key, -999999)) != int(step.get(key, 0)):
			runner._fail("assert_discovery_minimap failed: %s want=%s got=%s minimap=%s step=%d scenario=%s" % [
				key, str(step.get(key, 0)), str(minimap.get(key, null)), str(minimap), runner._step_index, str(runner.scenario.get("id", "?"))
			])
			return false
	for key in ["pin_x", "pin_y", "quest_path_start_x", "quest_path_start_y", "quest_path_end_x", "quest_path_end_y", "quest_path_angle_radians"]:
		var min_key := "%s_min" % key
		var max_key := "%s_max" % key
		var got := float(minimap.get(key, 0.0))
		if step.has(min_key) and got < float(step.get(min_key, 0.0)) - FLOAT_BOUND_EPSILON:
			runner._fail("assert_discovery_minimap failed: %s want_min=%s got=%s minimap=%s step=%d scenario=%s" % [
				key, str(step.get(min_key, 0.0)), str(minimap.get(key, null)), str(minimap), runner._step_index, str(runner.scenario.get("id", "?"))
			])
			return false
		if step.has(max_key) and got > float(step.get(max_key, 0.0)) + FLOAT_BOUND_EPSILON:
			runner._fail("assert_discovery_minimap failed: %s want_max=%s got=%s minimap=%s step=%d scenario=%s" % [
				key, str(step.get(max_key, 0.0)), str(minimap.get(key, null)), str(minimap), runner._step_index, str(runner.scenario.get("id", "?"))
			])
			return false
		if step.has(key) and not is_equal_approx(float(minimap.get(key, 0.0)), float(step.get(key, 0.0))):
			runner._fail("assert_discovery_minimap failed: %s want=%s got=%s minimap=%s step=%d scenario=%s" % [
				key, str(step.get(key, 0.0)), str(minimap.get(key, null)), str(minimap), runner._step_index, str(runner.scenario.get("id", "?"))
			])
			return false
	if step.has("panel_opacity_max") and float(minimap.get("panel_opacity", 1.0)) > float(step.get("panel_opacity_max", 1.0)):
		runner._fail("assert_discovery_minimap failed: panel_opacity max want=%s got=%s minimap=%s step=%d scenario=%s" % [
			str(step.get("panel_opacity_max", 1.0)), str(minimap.get("panel_opacity", null)), str(minimap), runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	if step.has("panel_opacity_min") and float(minimap.get("panel_opacity", 0.0)) < float(step.get("panel_opacity_min", 0.0)):
		runner._fail("assert_discovery_minimap failed: panel_opacity min want=%s got=%s minimap=%s step=%d scenario=%s" % [
			str(step.get("panel_opacity_min", 0.0)), str(minimap.get("panel_opacity", null)), str(minimap), runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	return true


static func movement_visual_smoothing_matches(step: Dictionary, state: Dictionary) -> bool:
	var smoothing: Dictionary = state.get("movement_visual_smoothing", {})
	if step.has("active") and bool(smoothing.get("active", false)) != bool(step.get("active", false)):
		return false
	if step.has("offset_min") and float(smoothing.get("offset_length", 0.0)) < float(step.get("offset_min", 0.0)):
		return false
	if step.has("offset_max") and float(smoothing.get("offset_length", 0.0)) > float(step.get("offset_max", 0.0)):
		return false
	return step.has("active") or step.has("offset_min") or step.has("offset_max")


static func command_retarget_grace_matches(step: Dictionary, state: Dictionary) -> bool:
	var grace: Dictionary = state.get("command_retarget_grace", {})
	var use_last := bool(step.get("last_dispatched", false))
	var expected_kind := str(step.get("kind", ""))
	if step.has("active") and bool(grace.get("active", false)) != bool(step.get("active", false)):
		return false
	if expected_kind != "":
		var got_kind := str(grace.get("last_dispatched_kind" if use_last else "kind", ""))
		if got_kind != expected_kind:
			return false
	for key in ["queued", "replaced", "dispatched", "expired"]:
		var want_key := "%s_min" % key
		if step.has(want_key) and int(grace.get("%s_count" % key, 0)) < int(step.get(want_key, 0)):
			return false
	if step.has("ground_x") or step.has("ground_z"):
		var gx := float(grace.get("last_dispatched_ground_x" if use_last else "ground_x", 0.0))
		var gz := float(grace.get("last_dispatched_ground_z" if use_last else "ground_z", 0.0))
		var want := Vector2(float(step.get("ground_x", gx)), float(step.get("ground_z", gz)))
		if Vector2(gx, gz).distance_to(want) > float(step.get("distance_max", 0.05)):
			return false
	return step.has("active") or expected_kind != "" or step.has("queued_min") or step.has("replaced_min") or step.has("dispatched_min") or step.has("expired_min") or step.has("ground_x") or step.has("ground_z")


static func melee_lunge_matches(step: Dictionary, state: Dictionary) -> bool:
	var lunge := _melee_lunge_state(state)
	if step.has("active") and bool(lunge.get("active", false)) != bool(step.get("active", false)):
		return false
	if step.has("count_min") and int(lunge.get("count", 0)) < int(step.get("count_min", 0)):
		return false
	if step.has("offset_min") and float(lunge.get("offset_length", 0.0)) < float(step.get("offset_min", 0.0)):
		return false
	if step.has("offset_max") and float(lunge.get("offset_length", 0.0)) > float(step.get("offset_max", 0.0)):
		return false
	return step.has("active") or step.has("count_min") or step.has("offset_min") or step.has("offset_max")


static func _melee_lunge_state(state: Dictionary) -> Dictionary:
	var local: Dictionary = state.get("local_player_presentation", {})
	var animation: Dictionary = local.get("animation", {})
	return animation.get("melee_lunge", {})


static func _assert_audio_state(runner, step: Dictionary, state: Dictionary) -> bool:
	var audio: Dictionary = state.get("audio", {})
	for key in ["ambient_zone", "ambient_active", "boss_music_active", "boss_music_layer", "last_cue", "last_skill_id"]:
		if step.has(key) and audio.get(key, null) != step[key]:
			runner._fail("assert_audio_state failed: key=%s want=%s got=%s audio=%s step=%d scenario=%s" % [
				key, str(step[key]), str(audio.get(key, null)), str(audio), runner._step_index, str(runner.scenario.get("id", "?"))
			])
			return false
	if step.has("boss_music_intensity_min") and float(audio.get("boss_music_intensity", 0.0)) < float(step.get("boss_music_intensity_min", 0.0)):
		runner._fail("assert_audio_state failed: boss_music_intensity_min want=%s got=%s audio=%s step=%d scenario=%s" % [
			str(step.get("boss_music_intensity_min", 0.0)), str(audio.get("boss_music_intensity", 0.0)), str(audio), runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	return true


static func _assert_resource_wallet_panel(runner, step: Dictionary, state: Dictionary) -> bool:
	var panel: Dictionary = state.get("character_bar", {})
	if step.has("visible") and bool(panel.get("wallet_visible", false)) != bool(step.get("visible", true)):
		runner._fail("assert_resource_wallet_panel visible failed: want=%s panel=%s step=%d scenario=%s" % [
			str(step.get("visible", true)), str(panel), runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	if step.has("text_contains") and not str(panel.get("wallet_text", "")).contains(str(step.get("text_contains", ""))):
		runner._fail("assert_resource_wallet_panel text failed: want contains=%s panel=%s step=%d scenario=%s" % [
			str(step.get("text_contains", "")), str(panel), runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	if step.has("tooltip_contains") and not str(panel.get("wallet_tooltip", "")).contains(str(step.get("tooltip_contains", ""))):
		runner._fail("assert_resource_wallet_panel tooltip failed: want contains=%s panel=%s step=%d scenario=%s" % [
			str(step.get("tooltip_contains", "")), str(panel), runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	var window: Dictionary = panel.get("wallet_window", {})
	if step.has("window_visible") and bool(window.get("visible", false)) != bool(step.get("window_visible", true)):
		runner._fail("assert_resource_wallet_panel window visible failed: want=%s panel=%s step=%d scenario=%s" % [
			str(step.get("window_visible", true)), str(panel), runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	if step.has("window_contains") and not str(window.get("text", "")).contains(str(step.get("window_contains", ""))):
		runner._fail("assert_resource_wallet_panel window text failed: want contains=%s panel=%s step=%d scenario=%s" % [
			str(step.get("window_contains", "")), str(panel), runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	if step.has("window_row_count_at_least") and int(window.get("row_count", 0)) < int(step.get("window_row_count_at_least", 0)):
		runner._fail("assert_resource_wallet_panel row count failed: want at least=%s panel=%s step=%d scenario=%s" % [
			str(step.get("window_row_count_at_least", 0)), str(panel), runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	return true


static func _assert_camera_mode(runner, step: Dictionary, state: Dictionary) -> bool:
	var want_mode := str(step.get("mode", ""))
	var got_mode := str(state.get("camera_mode", ""))
	if want_mode != "" and got_mode != want_mode:
		runner._fail("assert_camera_mode failed: mode want=%s got=%s step=%d scenario=%s" % [
			want_mode, got_mode, runner._step_index, str(runner.scenario.get("id", "?"))
		])
		return false
	if step.has("projection"):
		var want_proj := str(step.get("projection", ""))
		var got_proj := str(state.get("camera_projection", ""))
		if got_proj != want_proj:
			runner._fail("assert_camera_mode failed: projection want=%s got=%s step=%d scenario=%s" % [
				want_proj, got_proj, runner._step_index, str(runner.scenario.get("id", "?"))
			])
			return false
	if step.has("mouse_captured"):
		var want_cap := bool(step.get("mouse_captured", false))
		var got_cap := bool(state.get("mouse_captured", false))
		if got_cap != want_cap:
			runner._fail("assert_camera_mode failed: mouse_captured want=%s got=%s step=%d scenario=%s" % [
				want_cap, got_cap, runner._step_index, str(runner.scenario.get("id", "?"))
			])
			return false
	return true
