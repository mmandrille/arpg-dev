class_name BotMarketActions
extends RefCounted

const MarketFilterControlsScript := preload("res://scripts/market_filter_controls.gd")


static func dispatch(main, action: Dictionary) -> bool:
	var panel = _panel(main)
	if panel == null:
		return false
	match str(action.get("_type", action.get("type", ""))):
		"set_market_publish_price":
			if panel.has_method("bot_set_publish_price"):
				panel.bot_set_publish_price(int(action.get("price_gold", 1)))
		"click_market_publish_item":
			if panel.has_method("bot_click_publish_stash_item"):
				panel.bot_click_publish_stash_item(str(action.get("stash_item_id", "")), str(action.get("item_def_id", "")), action.get("rolled", null), int(action.get("stash_index", 0)))
		"click_market_purchase_listing":
			if panel.has_method("bot_click_purchase_listing"):
				panel.bot_click_purchase_listing(str(action.get("listing_id", "")), str(action.get("item_def_id", "")), int(action.get("price_gold", -1)), int(action.get("listing_index", 0)))
		"click_market_view_offers":
			if panel.has_method("bot_click_view_offers"):
				panel.bot_click_view_offers(str(action.get("listing_id", "")), str(action.get("item_def_id", "")), int(action.get("price_gold", -1)), int(action.get("listing_index", 0)))
		"click_market_cancel_listing":
			if panel.has_method("bot_click_cancel_listing"):
				panel.bot_click_cancel_listing(str(action.get("listing_id", "")), str(action.get("item_def_id", "")), int(action.get("price_gold", -1)), int(action.get("listing_index", 0)))
		"click_market_accept_offer", "click_market_cancel_offer":
			if panel.has_method("bot_click_offer_action"):
				panel.bot_click_offer_action("cancel_offer" if str(action.get("type", "")) == "click_market_cancel_offer" else "accept_offer", str(action.get("offer_id", "")), int(action.get("offer_index", 0)))
		"set_market_search":
			if panel.has_method("bot_set_market_search"):
				panel.bot_set_market_search(str(action.get("text", "")))
		"select_market_sort":
			if panel.has_method("bot_select_market_sort"):
				panel.bot_select_market_sort(str(action.get("mode", MarketFilterControlsScript.SORT_DEFAULT)))
		_:
			return false
	return true


static func summary(action: Dictionary) -> String:
	match str(action.get("_type", action.get("type", ""))):
		"set_market_publish_price":
			return "set_market_publish_price price=%s" % str(action.get("price_gold", 1))
		"click_market_publish_item":
			return "click_market_publish_item stash_item=%s item=%s rolled=%s stash_index=%s" % [str(action.get("stash_item_id", "")), str(action.get("item_def_id", "")), str(action.get("rolled", "")), str(action.get("stash_index", 0))]
		"click_market_purchase_listing":
			return "click_market_purchase_listing listing=%s item=%s price=%s index=%s" % [str(action.get("listing_id", "")), str(action.get("item_def_id", "")), str(action.get("price_gold", "")), str(action.get("listing_index", 0))]
		"click_market_view_offers":
			return "click_market_view_offers listing=%s item=%s price=%s index=%s" % [str(action.get("listing_id", "")), str(action.get("item_def_id", "")), str(action.get("price_gold", "")), str(action.get("listing_index", 0))]
		"click_market_cancel_listing":
			return "click_market_cancel_listing listing=%s item=%s price=%s index=%s" % [str(action.get("listing_id", "")), str(action.get("item_def_id", "")), str(action.get("price_gold", "")), str(action.get("listing_index", 0))]
		"click_market_accept_offer", "click_market_cancel_offer":
			return "%s offer=%s index=%s" % [str(action.get("type", "")), str(action.get("offer_id", "")), str(action.get("offer_index", 0))]
		"set_market_search":
			return "set_market_search text=%s" % str(action.get("text", ""))
		"select_market_sort":
			return "select_market_sort mode=%s" % str(action.get("mode", MarketFilterControlsScript.SORT_DEFAULT))
	return str(action.get("_type", action.get("type", "")))


static func _panel(main):
	return main.get("market_panel") if main != null else null
