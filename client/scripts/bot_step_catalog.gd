class_name BotStepCatalog
extends RefCounted

const BotActionStepValidatorScript := preload("res://scripts/bot_action_step_validator.gd")


const STEP_TYPES_WAIT := [
	"wait_ws_open", "wait_entity", "wait_event", "wait_inventory_item",
	"wait_inventory_count", "wait_loot_item", "wait_loot_count", "wait_hotbar_assigned",
	"wait_hotbar_capacity",
	"wait_player_near", "assert_entity_removed",
	"click_entity_until_event", "wait_main_menu", "wait_character_panel",
	"wait_multiplayer_panel", "wait_settings_panel", "wait_pause_menu", "wait_character_progression",
	"wait_skill_progression", "wait_skill_bar",
	"wait_damage_number", "wait_no_damage_number", "wait_entity_reaction",
	"wait_wall_layout", "wait_shop_panel", "wait_stash_panel", "wait_market_panel", "wait_bishop_panel", "wait_mercenary_panel", "wait_blacksmith_panel",
	"wait_boss_health_bar", "wait_remote_player_count",
	"wait_ticks", "wait_quest_journal", "wait_elite_objective_tracker", "wait_elite_objective_minimap",
]
const STEP_TYPES_ASSERT := [
	"assert_panel_visible", "assert_waypoint_panel_visible", "assert_equipped",
	"assert_unequipped", "assert_inventory_missing", "assert_inventory_count",
	"assert_loot_presentation", "assert_inventory_presentation",
	"assert_hotbar_assigned", "assert_player_hp", "assert_main_menu_visible",
	"assert_main_menu_actions", "assert_character_panel", "assert_create_game_type",
	"assert_current_session",
	"assert_character_panel_visible", "assert_settings_panel_visible",
	"assert_pause_menu_visible", "assert_session_changed",
	"assert_multiplayer_panel_visible", "assert_multiplayer_session_rows", "assert_multiplayer_filter",
	"assert_player_position_unchanged", "assert_character_stats_panel_visible",
	"assert_character_info_panel_visible", "assert_character_info",
	"assert_character_progression", "assert_stat_button_enabled", "assert_xp_bar",
	"assert_skills_panel_visible", "assert_skill_progression",
	"assert_skill_button_enabled", "assert_skill_bar",
	"assert_hotbar_capacity", "assert_hotbar_slot_disabled",
	"assert_inventory_capacity", "assert_bag_grid", "assert_paper_doll_layout",
	"assert_inventory_panel_details",
	"assert_floating_combat_text_enabled", "assert_damage_number", "assert_no_damage_number",
	"assert_entity_reaction",
	"assert_wall_layout", "assert_shop_panel_visible", "assert_shop_offer_count",
	"assert_shop_buy_button", "assert_shop_reroll_button", "assert_shop_sell_rows", "assert_shop_offer_details",
	"assert_shop_sell_details", "assert_stash_panel_visible", "assert_stash_item_count",
	"assert_stash_gold", "assert_stash_filter", "assert_market_panel_visible", "assert_market_listing_rows", "assert_market_offer_rows", "assert_boss_health_bar", "assert_audio_state", "assert_resource_wallet_panel",
	"assert_bishop_panel_visible", "assert_bishop_panel", "assert_mercenary_panel_visible", "assert_mercenary_panel", "assert_blacksmith_panel_visible", "assert_blacksmith_panel", "assert_boss_reward_status", "assert_remote_player_count",
	"assert_quest_journal", "assert_elite_objective_tracker", "assert_elite_objective_minimap",
]
const STEP_TYPES_ACTION := [
	"press_key", "click_entity", "click_loot_item", "click_floor",
	"drag_bag_to_weapon_slot", "drag_weapon_to_bag", "drag_bag_to_equipment_slot",
	"drag_equipment_to_bag", "drag_bag_to_outside", "assign_hotbar_slot",
	"use_hotbar_slot", "double_click_bag_item", "click_menu_button",
	"enter_character_name", "select_character", "select_character_class", "select_window_size",
	"set_floating_combat_text", "select_create_game_type",
	"remember_session", "remember_player_position", "click_stat_button",
	"click_skill_button", "use_skill_slot", "click_shop_buy_offer", "click_shop_reroll", "click_shop_sell_item",
	"drag_bag_to_stash", "drag_stash_to_bag", "click_stash_deposit_gold",
	"click_stash_withdraw_gold", "click_bishop_respec", "set_stash_search", "select_stash_sort",
	"set_multiplayer_search", "select_multiplayer_sort",
	"click_blacksmith_upgrade", "click_blacksmith_stage_item", "click_mercenary_stance",
	"set_market_publish_price", "click_market_publish_item", "click_market_purchase_listing",
	"click_market_view_offers", "click_market_cancel_listing", "click_market_accept_offer", "click_waypoint_level",
]
const WAIT_LOG_INTERVAL_S := 2.0

const ALL_STEP_TYPES: Array = STEP_TYPES_WAIT + STEP_TYPES_ASSERT + STEP_TYPES_ACTION


static func validate_scenario(data: Dictionary) -> String:
	if str(data.get("runner", "")) != "godot_client":
		return "runner must be 'godot_client', got '%s'" % str(data.get("runner", ""))
	if str(data.get("id", "")) == "":
		return "id must be non-empty"
	if str(data.get("world_id", "")) == "":
		return "world_id must be non-empty"
	var steps = data.get("client_steps", null)
	if steps == null or typeof(steps) != TYPE_ARRAY or (steps as Array).size() == 0:
		return "client_steps must be a non-empty array"
	for i in range((steps as Array).size()):
		var step = (steps as Array)[i]
		if typeof(step) != TYPE_DICTIONARY:
			return "client_steps[%d] must be an object" % i
		var err := validate_step(step as Dictionary, i)
		if err != "":
			return err
	return ""



static func validate_step(step: Dictionary, index: int) -> String:
	var stype := str(step.get("type", ""))
	if stype == "":
		return "client_steps[%d].type is missing" % index
	if stype not in ALL_STEP_TYPES:
		return "client_steps[%d].type '%s' is unknown" % [index, stype]
	var action_err := BotActionStepValidatorScript.validate(step, stype, index)
	if action_err != BotActionStepValidatorScript.UNHANDLED:
		return action_err
	if stype in STEP_TYPES_WAIT and stype != "wait_loot_item":
		var timeout = step.get("timeout_s", null)
		if timeout == null or float(timeout) <= 0.0:
			return "client_steps[%d] (%s) requires a positive timeout_s" % [index, stype]
	if stype == "wait_entity" or stype == "assert_entity_removed":
		if str(step.get("entity_type", "")) == "":
			return "client_steps[%d] (%s) requires entity_type" % [index, stype]
	if stype == "wait_event":
		if str(step.get("event_type", "")) == "":
			return "client_steps[%d] (%s) requires event_type" % [index, stype]
	if stype == "click_entity_until_event":
		if str(step.get("entity_type", "")) == "" or str(step.get("event_type", "")) == "":
			return "client_steps[%d] (%s) requires entity_type and event_type" % [index, stype]
	if stype == "wait_player_near":
		if not step.has("x") or not step.has("z"):
			return "client_steps[%d] (%s) requires x and z" % [index, stype]
	if stype in ["wait_inventory_count", "assert_inventory_count"]:
		if not step.has("equals"):
			return "client_steps[%d] (%s) requires equals" % [index, stype]
	if stype == "wait_loot_count":
		if not step.has("min_count"):
			return "client_steps[%d] (%s) requires min_count" % [index, stype]
	if stype in ["assert_hotbar_assigned", "wait_hotbar_assigned"]:
		if not step.has("slot_index") or str(step.get("item_def_id", "")) == "":
			return "client_steps[%d] (%s) requires slot_index and item_def_id" % [index, stype]
	if stype in ["assert_hotbar_capacity", "wait_hotbar_capacity", "assert_hotbar_slot_disabled"]:
		if not step.has("equals") and not step.has("at_least") and stype in ["assert_hotbar_capacity", "wait_hotbar_capacity"]:
			return "client_steps[%d] (%s) requires equals or at_least" % [index, stype]
		if not step.has("slot_index") and stype == "assert_hotbar_slot_disabled":
			return "client_steps[%d] (%s) requires slot_index" % [index, stype]
	if stype == "assert_inventory_capacity":
		if not step.has("rows") and not step.has("capacity"):
			return "client_steps[%d] (%s) requires rows or capacity" % [index, stype]
	if stype == "assert_bag_grid":
		if not step.has("columns") and not step.has("available_slot_count"):
			return "client_steps[%d] (%s) requires columns or available_slot_count" % [index, stype]
	if stype == "assert_paper_doll_layout":
		if not step.has("slots"):
			return "client_steps[%d] (%s) requires slots" % [index, stype]
	if stype == "assert_player_hp":
		if not step.has("equals"):
			return "client_steps[%d] (%s) requires equals" % [index, stype]
	if stype == "assert_stat_button_enabled":
		if str(step.get("stat", "")) == "":
			return "client_steps[%d] (%s) requires stat" % [index, stype]
	if stype == "assert_stat_button_enabled":
		if not step.has("enabled"):
			return "client_steps[%d] (%s) requires enabled" % [index, stype]
	if stype == "assert_character_info":
		if not step.has("name") and not step.has("level") and not step.has("area"):
			return "client_steps[%d] (%s) requires name, level, or area" % [index, stype]
	if stype in ["wait_character_progression", "assert_character_progression"]:
		var has_any := false
		for key in ["level", "experience", "unspent_stat_points", "gold", "deepest_dungeon_depth", "str", "dex", "vit", "magic", "derived_stats", "player_max_hp", "stat_breakdowns"]:
			if step.has(key):
				has_any = true
		if not has_any:
			return "client_steps[%d] (%s) requires at least one progression expectation" % [index, stype]
	if stype in ["wait_skill_progression", "assert_skill_progression"]:
		var has_skill_expectation := false
		for key in ["unspent_skill_points", "rank", "max_rank", "can_spend"]:
			if step.has(key):
				has_skill_expectation = true
		if not has_skill_expectation:
			return "client_steps[%d] (%s) requires at least one skill progression expectation" % [index, stype]
	if stype == "assert_skill_button_enabled":
		if not step.has("enabled"):
			return "client_steps[%d] (%s) requires enabled" % [index, stype]
	if stype in ["wait_skill_bar", "assert_skill_bar"]:
		var has_bar_expectation := false
		for key in ["rank", "max_rank", "enabled", "disabled", "remaining_ticks", "remaining_ticks_min", "remaining_ticks_max", "total_ticks", "cooldown_fraction_min", "cooldown_fraction_max", "slot_text", "tooltip_contains"]:
			if step.has(key):
				has_bar_expectation = true
		if not has_bar_expectation:
			return "client_steps[%d] (%s) requires at least one skill bar expectation" % [index, stype]
	if stype == "wait_ticks":
		if int(step.get("ticks", 0)) <= 0:
			return "client_steps[%d] (%s) requires positive ticks" % [index, stype]
	if stype in ["wait_boss_health_bar", "assert_boss_health_bar"]:
		var has_boss_bar_expectation := false
		for key in ["visible", "boss_id", "boss_template_id", "title", "hp", "max_hp", "hp_min", "hp_max", "ratio_min", "ratio_max", "phase_kind", "pattern_id", "phase_index", "duration_ticks", "remaining_ticks_min", "remaining_ticks_max", "phase_ratio_min", "phase_ratio_max"]:
			if step.has(key):
				has_boss_bar_expectation = true
		if not has_boss_bar_expectation:
			return "client_steps[%d] (%s) requires at least one boss health bar expectation" % [index, stype]
	if stype == "assert_boss_reward_status":
		if str(step.get("text", "")) == "":
			return "client_steps[%d] (%s) requires text" % [index, stype]
	if stype in ["set_floating_combat_text", "assert_floating_combat_text_enabled"]:
		if not step.has("enabled"):
			return "client_steps[%d] (%s) requires enabled" % [index, stype]
	if stype == "assert_create_game_type":
		var session_type := str(step.get("session_type", ""))
		if session_type != "coop" and session_type != "solo":
			return "client_steps[%d] (%s) requires session_type coop or solo" % [index, stype]
	if stype == "assert_main_menu_actions":
		if not step.has("labels") and not step.has("actions"):
			return "client_steps[%d] (%s) requires labels or actions" % [index, stype]
	if stype == "assert_character_panel":
		var has_panel_expectation := false
		for key in ["visible", "mode", "title", "character_count", "min_character_count", "name_field_visible", "create_button_visible", "class_picker_visible", "selected_class", "row_character_class", "empty_visible", "character_id", "status", "level", "min_level", "gold", "min_gold", "deepest_dungeon_depth", "min_deepest_dungeon_depth", "label_contains"]:
			if step.has(key):
				has_panel_expectation = true
		if not has_panel_expectation:
			return "client_steps[%d] (%s) requires a panel expectation" % [index, stype]
	if stype == "assert_current_session":
		if not step.has("exists") and not step.has("mode") and not step.has("listed") and not step.has("session_id") and not step.has("session_id_env"):
			return "client_steps[%d] (%s) requires exists, mode, listed, session_id, or session_id_env" % [index, stype]
	if stype in ["wait_damage_number", "assert_damage_number"]:
		if not step.has("text") and not step.has("variant"):
			return "client_steps[%d] (%s) requires text or variant" % [index, stype]
	if stype in ["wait_wall_layout", "assert_wall_layout"]:
		if not step.has("equals") and not step.has("at_least") and not step.has("generated_at_least") and not step.has("non_perimeter_at_least") and not step.has("current_level"):
			return "client_steps[%d] (%s) requires a wall count or current_level expectation" % [index, stype]
	if stype in ["wait_shop_panel", "assert_shop_offer_count"]:
		if not step.has("equals") and not step.has("at_least"):
			return "client_steps[%d] (%s) requires equals or at_least" % [index, stype]
	if stype in ["wait_stash_panel", "assert_stash_item_count", "assert_stash_gold"]:
		if not step.has("equals") and not step.has("at_least"):
			return "client_steps[%d] (%s) requires equals or at_least" % [index, stype]
	if stype in ["wait_market_panel", "assert_market_listing_rows"]:
		if not step.has("equals") and not step.has("at_least") and not step.has("price_gold") and not step.has("item_def_id") and not step.has("rolled") and not step.has("offer_equals") and not step.has("offer_at_least") and not step.has("offer_item_def_id") and not step.has("offer_status"):
			return "client_steps[%d] (%s) requires a market listing expectation" % [index, stype]
	if stype == "assert_market_offer_rows":
		if not step.has("offer_equals") and not step.has("offer_at_least"):
			return "client_steps[%d] (%s) requires offer_equals or offer_at_least" % [index, stype]
	if stype in ["wait_bishop_panel", "assert_bishop_panel"]:
		if not step.has("price") and not step.has("gold") and not step.has("affordable") and not step.has("respec_enabled") and not step.has("service_id") and not step.has("visible") and not step.has("status_contains"):
			return "client_steps[%d] (%s) requires a bishop panel expectation" % [index, stype]
	if stype in ["wait_mercenary_panel", "assert_mercenary_panel"]:
		if not step.has("visible") and not step.has("price") and not step.has("gold") and not step.has("affordable") and not step.has("service_id") and not step.has("offer_id") and not step.has("monster_def_id") and not step.has("hired_entity_id") and not step.has("hired_count") and not step.has("selected_stance") and not step.has("status_contains") and not step.has("companion_bar_count") and not step.has("companion_icon_kind"):
			return "client_steps[%d] (%s) requires a mercenary panel expectation" % [index, stype]
	if stype in ["wait_blacksmith_panel", "assert_blacksmith_panel"]:
			if not step.has("visible") and not step.has("stash_gold_equals") and not step.has("stash_gold_at_least") and not step.has("item_count") and not step.has("status_contains") and not step.has("preview_contains") and not step.has("success_chance_percent") and not step.has("resource_item_def_id") and not step.has("resource_required_count") and not step.has("resource_inventory_count") and not step.has("resource_wallet_count") and not step.has("pity_failure_count") and not step.has("pity_threshold") and not step.has("pity_guaranteed") and not step.has("item_def_id") and not step.has("stash_item_id") and not step.has("item_level") and not step.has("upgrade_enabled"):
				return "client_steps[%d] (%s) requires a blacksmith panel expectation" % [index, stype]
	if stype == "assert_shop_buy_button":
		if str(step.get("offer_id", "")) == "":
			return "client_steps[%d] (%s) requires offer_id" % [index, stype]
	if stype == "assert_shop_reroll_button":
		if not step.has("visible") and not step.has("enabled") and not step.has("cost"):
			return "client_steps[%d] (%s) requires visible, enabled, or cost" % [index, stype]
	if stype in ["assert_shop_sell_rows", "assert_shop_offer_details", "assert_shop_sell_details"]:
		if not step.has("equals") and not step.has("at_least"):
			return "client_steps[%d] (%s) requires equals or at_least" % [index, stype]
	if stype == "assert_stash_filter":
		if not step.has("search_text") and not step.has("sort_mode") and not step.has("filtered_equals") and not step.has("first_item_def_id"):
			return "client_steps[%d] (%s) requires search_text, sort_mode, filtered_equals, or first_item_def_id" % [index, stype]
	if stype == "assert_multiplayer_session_rows":
		if not step.has("equals") and not step.has("min_count") and not step.has("selected") and not step.has("listed") and not step.has("session_id") and not step.has("session_id_env") and not step.has("mode") and not step.has("member_count_min") and not step.has("connected_count_min"):
			return "client_steps[%d] (%s) requires a row expectation" % [index, stype]
	if stype == "assert_multiplayer_filter":
		if not step.has("search_text") and not step.has("search_text_env") and not step.has("sort_mode") and not step.has("filtered_equals") and not step.has("total_at_least"):
			return "client_steps[%d] (%s) requires search_text, search_text_env, sort_mode, filtered_equals, or total_at_least" % [index, stype]
	if stype in ["wait_remote_player_count", "assert_remote_player_count"]:
		if not step.has("equals") and not step.has("at_least"):
			return "client_steps[%d] (%s) requires equals or at_least" % [index, stype]
	return ""
