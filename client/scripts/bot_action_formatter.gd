class_name BotActionFormatter
extends RefCounted

const BotMarketActionsScript := preload("res://scripts/bot_market_actions.gd")

static func format_action(action: Dictionary) -> String:
	var stype := str(action.get("_type", action.get("type", "")))
	match stype:
		"click_entity":
			return "click_entity type=%s index=%s" % [
				str(action.get("entity_type", "")), str(action.get("entity_index", 0))
			]
		"click_loot_item":
			return "click_loot_item item=%s rolled=%s occurrence=%s" % [
				str(action.get("item_def_id", "")),
				str(action.get("rolled", "")),
				str(action.get("occurrence", 0))
			]
		"click_floor":
			return "click_floor x=%s z=%s" % [str(action.get("x", "")), str(action.get("z", ""))]
		"press_key":
			return "press_key %s" % str(action.get("keycode", ""))
		"click_menu_button":
			return "click_menu_button %s" % str(action.get("button", ""))
		"enter_character_name":
			return "enter_character_name %s" % str(action.get("name", ""))
		"select_character":
			return "select_character index=%s" % str(action.get("index", 0))
		"select_window_size":
			return "select_window_size %s" % str(action.get("size", ""))
		"select_create_game_type":
			return "select_create_game_type %s" % str(action.get("session_type", ""))
		"click_stat_button":
			return "click_stat_button %s" % str(action.get("stat", ""))
		"click_skill_button":
			return "click_skill_button %s" % str(action.get("skill_id", "magic_bolt"))
		"use_skill_slot":
			return "use_skill_slot skill=%s target=%s monster=%s force=%s direction=%s" % [
				str(action.get("skill_id", "magic_bolt")),
				str(action.get("target_id", "")),
				str(action.get("monster_def_id", "")),
				str(action.get("force_direct", false)),
				str(action.get("direction", {})),
			]
		"click_shop_buy_offer":
			return "click_shop_buy offer_id=%s kind=%s index=%s" % [
				str(action.get("offer_id", "")),
				str(action.get("offer_kind", "")),
				str(action.get("offer_index", 0)),
			]
		"click_shop_reroll":
			return "click_shop_reroll"
		"click_shop_sell_item":
			return "click_shop_sell item=%s rolled=%s bag_index=%s" % [
				str(action.get("item_def_id", "")),
				str(action.get("rolled", "")),
				str(action.get("bag_index", 0)),
			]
		"click_waypoint_level":
			return "click_waypoint_level target=%s" % str(action.get("target_level", ""))
		"drag_bag_to_stash":
			return "drag_bag_to_stash item=%s rolled=%s bag_index=%s" % [
				str(action.get("item_def_id", "")),
				str(action.get("rolled", "")),
				str(action.get("bag_index", 0)),
			]
		"drag_stash_to_bag":
			return "drag_stash_to_bag stash_item=%s item=%s rolled=%s stash_index=%s" % [
				str(action.get("stash_item_id", "")),
				str(action.get("item_def_id", "")),
				str(action.get("rolled", "")),
				str(action.get("stash_index", 0)),
			]
		"click_stash_deposit_gold":
			return "click_stash_deposit_gold amount=%s" % str(action.get("amount", 1))
		"click_stash_withdraw_gold":
			return "click_stash_withdraw_gold amount=%s" % str(action.get("amount", 1))
		"click_bishop_respec":
			return "click_bishop_respec"
		"click_bishop_debug":
			return "click_bishop_debug action=%s" % str(action.get("action", ""))
		"click_blacksmith_upgrade":
			return "click_blacksmith_upgrade stash_item=%s item=%s stash_index=%s" % [
				str(action.get("stash_item_id", "")),
				str(action.get("item_def_id", "")),
				str(action.get("stash_index", 0)),
			]
		"click_blacksmith_stage_item":
			return "click_blacksmith_stage_item stash_item=%s item=%s stash_index=%s" % [
				str(action.get("stash_item_id", "")),
				str(action.get("item_def_id", "")),
				str(action.get("stash_index", 0)),
			]
		"set_stash_search":
			return "set_stash_search text=%s" % str(action.get("text", ""))
		"select_stash_sort":
			return "select_stash_sort mode=%s" % str(action.get("mode", "acquired"))
		"set_market_publish_price", "click_market_publish_item", "click_market_purchase_listing", \
		"click_market_view_offers", "click_market_cancel_listing", "click_market_accept_offer", \
		"click_market_cancel_offer", "set_market_search", "select_market_sort":
			return BotMarketActionsScript.summary(action)
		"assign_hotbar_slot":
			return "assign_hotbar slot=%s item=%s bag_index=%s" % [
				str(action.get("slot_index", "")),
				str(action.get("item_def_id", "")),
				str(action.get("bag_index", "")),
			]
		"double_click_bag_item":
			return "double_click_bag item=%s bag_index=%s" % [
				str(action.get("item_def_id", "")), str(action.get("bag_index", ""))
			]
		_:
			return stype
