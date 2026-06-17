class_name BotActionStepValidator
extends RefCounted

const UNHANDLED := "__unhandled__"
const STEP_TYPES_ACTION := [
	"press_key", "click_entity", "click_loot_item", "click_floor",
	"drag_bag_to_weapon_slot", "drag_weapon_to_bag", "drag_bag_to_equipment_slot",
	"drag_equipment_to_bag", "drag_bag_to_outside", "assign_hotbar_slot",
	"use_hotbar_slot", "double_click_bag_item", "click_menu_button",
	"enter_character_name", "select_character", "select_character_class", "select_window_size",
	"set_floating_combat_text", "select_create_game_type",
	"remember_session", "remember_player_position", "click_stat_button",
	"click_skill_button", "use_skill_slot", "click_shop_buy_offer", "click_shop_reroll", "click_shop_sell_item",
	"open_resource_wallet_window",
	"drag_bag_to_stash", "drag_stash_to_bag", "click_stash_deposit_gold",
	"click_stash_withdraw_gold", "click_bishop_respec", "set_stash_search", "select_stash_sort",
	"set_multiplayer_search", "select_multiplayer_sort",
	"click_blacksmith_upgrade", "click_blacksmith_stage_item",
	"set_market_publish_price", "click_market_publish_item", "click_market_purchase_listing",
	"click_market_view_offers", "click_market_cancel_listing", "click_market_accept_offer", "click_market_cancel_offer",
	"set_market_search", "select_market_sort", "click_waypoint_level",
]


static func validate(step: Dictionary, stype: String, index: int) -> String:
	if stype not in STEP_TYPES_ACTION:
		return UNHANDLED
	if stype == "press_key":
		return _require_string(step, index, stype, "keycode")
	if stype == "click_entity":
		return _require_string(step, index, stype, "entity_type")
	if stype == "click_loot_item":
		if str(step.get("item_def_id", "")) == "" and not step.has("rolled"):
			return "client_steps[%d] (%s) requires item_def_id or rolled" % [index, stype]
	if stype == "click_floor":
		if not step.has("x") or not step.has("z"):
			return "client_steps[%d] (%s) requires x and z" % [index, stype]
	if stype == "click_waypoint_level":
		if not step.has("target_level"):
			return "client_steps[%d] (%s) requires target_level" % [index, stype]
	if stype in ["drag_bag_to_weapon_slot", "drag_bag_to_equipment_slot", "drag_bag_to_outside", "assign_hotbar_slot", "double_click_bag_item"]:
		if str(step.get("item_def_id", "")) == "":
			return "client_steps[%d] (%s) requires item_def_id" % [index, stype]
	if stype in ["drag_bag_to_equipment_slot", "drag_equipment_to_bag"]:
		if str(step.get("slot", "")) == "":
			return "client_steps[%d] (%s) requires slot" % [index, stype]
	if stype in ["assign_hotbar_slot", "use_hotbar_slot"]:
		if not step.has("slot_index"):
			return "client_steps[%d] (%s) requires slot_index" % [index, stype]
	if stype == "click_menu_button":
		return _require_string(step, index, stype, "button")
	if stype == "enter_character_name":
		return _require_string(step, index, stype, "name")
	if stype == "select_character_class":
		return _require_string(step, index, stype, "class_id")
	if stype == "select_window_size":
		return _require_string(step, index, stype, "size")
	if stype == "click_stat_button":
		return _require_string(step, index, stype, "stat")
	if stype == "select_create_game_type":
		var session_type := str(step.get("session_type", ""))
		if session_type != "coop" and session_type != "solo":
			return "client_steps[%d] (%s) requires session_type coop or solo" % [index, stype]
	if stype == "click_shop_buy_offer":
		if str(step.get("offer_id", "")) == "" and str(step.get("offer_kind", "")) == "":
			return "client_steps[%d] (%s) requires offer_id or offer_kind" % [index, stype]
	if stype == "drag_bag_to_stash":
		if str(step.get("item_def_id", "")) == "" and not step.has("rolled"):
			return "client_steps[%d] (%s) requires item_def_id or rolled" % [index, stype]
	if stype == "drag_stash_to_bag":
		if str(step.get("stash_item_id", "")) == "" and str(step.get("item_def_id", "")) == "" and not step.has("rolled"):
			return "client_steps[%d] (%s) requires stash_item_id, item_def_id, or rolled" % [index, stype]
	if stype in ["click_stash_deposit_gold", "click_stash_withdraw_gold"]:
		if int(step.get("amount", 0)) <= 0:
			return "client_steps[%d] (%s) requires positive amount" % [index, stype]
	if stype == "set_stash_search":
		if not step.has("text"):
			return "client_steps[%d] (%s) requires text" % [index, stype]
	if stype == "select_stash_sort":
		var mode := str(step.get("mode", ""))
		if not ["acquired", "name", "rarity", "slot", "unique_chest_sets"].has(mode):
			return "client_steps[%d] (%s) requires mode acquired, name, rarity, slot, or unique_chest_sets" % [index, stype]
	if stype == "set_multiplayer_search":
		if not step.has("text") and not step.has("text_env"):
			return "client_steps[%d] (%s) requires text or text_env" % [index, stype]
	if stype == "select_multiplayer_sort":
		var mode := str(step.get("mode", ""))
		if not ["recent", "host", "players"].has(mode):
			return "client_steps[%d] (%s) requires mode recent, host, or players" % [index, stype]
	if stype == "set_market_publish_price":
		if int(step.get("price_gold", 0)) <= 0:
			return "client_steps[%d] (%s) requires positive price_gold" % [index, stype]
	if stype == "set_market_search":
		if not step.has("text"):
			return "client_steps[%d] (%s) requires text" % [index, stype]
	if stype == "select_market_sort":
		var mode := str(step.get("mode", ""))
		if not ["default", "name", "price_low", "price_high", "status"].has(mode):
			return "client_steps[%d] (%s) requires mode default, name, price_low, price_high, or status" % [index, stype]
	if stype == "click_market_publish_item":
		if str(step.get("stash_item_id", "")) == "" and str(step.get("item_def_id", "")) == "" and not step.has("rolled"):
			return "client_steps[%d] (%s) requires stash_item_id, item_def_id, or rolled" % [index, stype]
	if stype in ["click_blacksmith_upgrade", "click_blacksmith_stage_item"]:
		if str(step.get("stash_item_id", "")) == "" and str(step.get("item_def_id", "")) == "":
			return "client_steps[%d] (%s) requires stash_item_id or item_def_id" % [index, stype]
	if stype in ["click_market_purchase_listing", "click_market_cancel_listing"]:
		if str(step.get("listing_id", "")) == "" and str(step.get("item_def_id", "")) == "" and not step.has("price_gold"):
			return "client_steps[%d] (%s) requires listing_id, item_def_id, or price_gold" % [index, stype]
	return ""


static func _require_string(step: Dictionary, index: int, stype: String, field: String) -> String:
	if str(step.get(field, "")) == "":
		return "client_steps[%d] (%s) requires %s" % [index, stype, field]
	return ""
