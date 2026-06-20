class_name BotWaitHandlers
extends RefCounted

const BotQuestJournalAssertionsScript := preload("res://scripts/bot_quest_journal_assertions.gd")
const BotEliteObjectiveAssertionsScript := preload("res://scripts/bot_elite_objective_assertions.gd")
const BotEliteObjectiveMinimapAssertionsScript := preload("res://scripts/bot_elite_objective_minimap_assertions.gd")
const BotMercenaryPanelAssertionsScript := preload("res://scripts/bot_mercenary_panel_assertions.gd")
const BotMarketReceiptAssertionsScript := preload("res://scripts/bot_market_receipt_assertions.gd")
const BotMarketBadgeAssertionsScript := preload("res://scripts/bot_market_badge_assertions.gd")
const BotAssertionHandlersScript := preload("res://scripts/bot_assertion_handlers.gd")


static func evaluate(runner, step: Dictionary, stype: String, state: Dictionary) -> bool:
	match stype:
		"wait_ws_open":
			return bool(state.get("ws_open", false))
		"wait_main_menu":
			return bool(state.get("main_menu_visible", false))
		"wait_character_panel":
			return bool(state.get("character_panel_visible", false))
		"wait_multiplayer_panel":
			return bool(state.get("multiplayer_panel_visible", false))
		"wait_settings_panel":
			return bool(state.get("settings_panel_visible", false))
		"wait_pause_menu":
			return bool(state.get("pause_menu_visible", false))
		"wait_character_progression":
			return runner._progression_matches(step, state)
		"wait_skill_progression":
			return runner._skill_progression_matches(step, state)
		"wait_skill_bar":
			return runner._skill_bar_matches(step, state)
		"wait_damage_number":
			return runner._damage_number_matches(step, state)
		"wait_no_damage_number":
			return (state.get("damage_numbers", []) as Array).is_empty()
		"wait_entity_reaction":
			return runner._presentation_matches(step, state)
		"wait_movement_visual_smoothing":
			return BotAssertionHandlersScript.movement_visual_smoothing_matches(step, state)
		"wait_command_retarget_grace":
			return BotAssertionHandlersScript.command_retarget_grace_matches(step, state)
		"wait_wall_layout":
			return runner._wall_layout_matches(step, state)
		"wait_fog_of_war":
			return BotAssertionHandlersScript.fog_of_war_matches(step, state)
		"wait_shop_panel":
			if not bool(state.get("shop_panel_visible", false)):
				return false
			return runner._shop_offer_count_matches(step, state)
		"wait_stash_panel":
			if not bool(state.get("stash_panel_visible", false)):
				return false
			return runner._stash_item_count_matches(step, state)
		"wait_market_panel":
			if not bool(state.get("market_panel_visible", false)):
				return false
			return runner._market_listing_rows_match(step, state) and runner._market_offer_rows_match(step, state) and BotMarketReceiptAssertionsScript.matches(step, state)
		"wait_market_board_badges":
			return BotMarketBadgeAssertionsScript.matches(step, state)
		"wait_bishop_panel":
			if not bool(state.get("bishop_panel_visible", false)):
				return false
			return runner._bishop_panel_matches(step, state)
		"wait_mercenary_panel":
			if not bool(state.get("mercenary_panel_visible", false)):
				return false
			return BotMercenaryPanelAssertionsScript.matches(step, state)
		"wait_blacksmith_panel":
			if not bool(state.get("blacksmith_panel_visible", false)):
				return false
			return runner._blacksmith_panel_matches(step, state)
		"wait_boss_health_bar":
			return runner._boss_health_bar_matches(step, state)
		"wait_remote_player_count":
			return runner._remote_player_count_matches(step, state)
		"wait_quest_journal":
			return BotQuestJournalAssertionsScript.matches(step, state)
		"wait_elite_objective_tracker":
			return BotEliteObjectiveAssertionsScript.matches(step, state)
		"wait_elite_objective_minimap":
			return BotEliteObjectiveMinimapAssertionsScript.matches(step, state)
		"wait_ticks":
			return runner._wait_ticks(step, state)
		"wait_entity":
			var etype := str(step.get("entity_type", ""))
			var eids: Array = state.get("%s_ids" % etype, state.get("entities_by_type", {}).get(etype, []))
			return eids.size() > 0
		"wait_event":
			var evtypes := _event_types(step)
			var event_step := _event_match_step(step, state)
			var pending: Array = state.get("pending_events", [])
			for i in range(pending.size()):
				if evtypes.has(str(pending[i].get("event_type", ""))) and runner._event_matches(event_step, pending[i]):
					if runner._controller != null and runner._controller.has_method("consume_pending_event_at"):
						runner._controller.consume_pending_event_at(i)
					return true
			return false
		"wait_inventory_count":
			var def_id := str(step.get("item_def_id", ""))
			var want := int(step.get("equals", 0))
			return runner._inventory_count(state, def_id) == want
		"wait_loot_count":
			var min_count := int(step.get("min_count", 1))
			return (state.get("loot_ids", []) as Array).size() >= min_count
		"click_entity_until_event":
			var evtype := str(step.get("event_type", ""))
			var pending: Array = state.get("pending_events", [])
			var event_step := step.duplicate()
			for selector_key in ["entity_type", "entity_index", "monster_def_id", "interactable_def_id", "item_def_id", "rarity", "state", "is_boss"]:
				event_step.erase(selector_key)
			for i in range(pending.size()):
				var ev = pending[i]
				if str(ev.get("event_type", "")) == evtype and runner._event_matches(event_step, ev):
					if bool(step.get("consume_event", false)) and runner._controller != null and runner._controller.has_method("consume_pending_event_at"):
						runner._controller.consume_pending_event_at(i)
					return true
			var retry_s := float(step.get("retry_s", 0.25))
			if runner._step_elapsed - runner._last_retry_at >= retry_s:
				runner._last_retry_at = runner._step_elapsed
				runner.pending_action = {
					"type": "click_entity",
					"_type": "click_entity",
					"entity_type": str(step.get("entity_type", "")),
					"entity_index": int(step.get("entity_index", 0)),
				}
				for key in ["monster_def_id", "interactable_def_id", "item_def_id", "rarity", "state", "is_boss"]:
					if step.has(key):
						runner.pending_action[key] = step[key]
			return false
		"wait_inventory_item":
			var def_id := str(step.get("item_def_id", ""))
			var inv: Array = state.get("inventory", [])
			if def_id == "":
				return inv.size() > 0
			for item in inv:
				if str(item.get("item_def_id", "")) == def_id:
					return true
			return false
		"wait_loot_item":
			return (state.get("loot_ids", []) as Array).size() > 0
		"wait_hotbar_assigned":
			return runner._hotbar_slot_matches(step, state)
		"wait_hotbar_capacity":
			return runner._hotbar_capacity_matches(step, state)
		"wait_player_near":
			var tx := float(step.get("x", 0.0))
			var tz := float(step.get("z", 0.0))
			var max_dist := float(step.get("distance", 2.5))
			var pp: Dictionary = state.get("player_pos", {})
			var px := float(pp.get("x", 0.0))
			var pz := float(pp.get("z", 0.0))
			var dist := sqrt((px - tx) * (px - tx) + (pz - tz) * (pz - tz))
			return dist <= max_dist
		"wait_entity_near_player":
			return _entity_near_player(runner, step, state)
		"assert_entity_removed":
			# Treated as a wait step: server entity_remove may arrive in a
			# subsequent delta after the kill event. Times out via timeout_s.
			var etype := str(step.get("entity_type", ""))
			var eids: Array = state.get("%s_ids" % etype, [])
			return eids.is_empty()
	return false


static func _event_types(step: Dictionary) -> Array:
	var out: Array = []
	if step.has("event_types") and typeof(step.get("event_types")) == TYPE_ARRAY:
		for event_type in step.get("event_types", []):
			out.append(str(event_type))
	else:
		out.append(str(step.get("event_type", "")))
	return out


static func _event_match_step(step: Dictionary, state: Dictionary) -> Dictionary:
	var out := step.duplicate()
	if bool(step.get("source_is_local_player", false)):
		var local_player_id := str(state.get("local_player_id", ""))
		if local_player_id != "":
			out["source_entity_id"] = local_player_id
	if bool(step.get("target_is_local_player", false)):
		var local_player_id := str(state.get("local_player_id", ""))
		if local_player_id != "":
			out["target_entity_id"] = local_player_id
	var target_entity_id := _selected_event_entity_id(step, state, "target")
	if target_entity_id != "":
		out["target_entity_id"] = target_entity_id
	return out


static func _selected_event_entity_id(step: Dictionary, state: Dictionary, prefix: String) -> String:
	var entity_type := str(step.get("%s_entity_type" % prefix, ""))
	if entity_type == "":
		return ""
	var monster_def_id := str(step.get("%s_monster_def_id" % prefix, ""))
	var item_def_id := str(step.get("%s_item_def_id" % prefix, ""))
	var interactable_def_id := str(step.get("%s_interactable_def_id" % prefix, ""))
	var entity_index := int(step.get("%s_entity_index" % prefix, 0))
	var matches: Array = []
	for row in state.get("entities_debug", []):
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if str(rec.get("type", "")) != entity_type:
			continue
		if monster_def_id != "" and str(rec.get("monster_def_id", "")) != monster_def_id:
			continue
		if item_def_id != "" and str(rec.get("item_def_id", "")) != item_def_id:
			continue
		if interactable_def_id != "" and str(rec.get("interactable_def_id", "")) != interactable_def_id:
			continue
		matches.append(str(rec.get("id", "")))
	if entity_index < 0 or entity_index >= matches.size():
		return ""
	return str(matches[entity_index])


static func _entity_near_player(runner, step: Dictionary, state: Dictionary) -> bool:
	var pp: Dictionary = state.get("player_pos", {})
	var px := float(pp.get("x", 0.0))
	var pz := float(pp.get("z", 0.0))
	var max_dist := float(step.get("distance", 2.5))
	for row in state.get("entities_presentation_debug", []):
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if not runner._presentation_row_matches(step, rec):
			continue
		var pos: Dictionary = rec.get("position", {})
		var dx := float(pos.get("x", 0.0)) - px
		var dz := float(pos.get("z", 0.0)) - pz
		if sqrt(dx * dx + dz * dz) <= max_dist:
			return true
	return false
