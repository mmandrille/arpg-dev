class_name BotAssertionHandlers
extends RefCounted

const BotUiAssertionHandlersScript := preload("res://scripts/bot_ui_assertion_handlers.gd")
const BotQuestJournalAssertionsScript := preload("res://scripts/bot_quest_journal_assertions.gd")
const BotEliteObjectiveAssertionsScript := preload("res://scripts/bot_elite_objective_assertions.gd")
const BotEliteObjectiveMinimapAssertionsScript := preload("res://scripts/bot_elite_objective_minimap_assertions.gd")


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
	return true
