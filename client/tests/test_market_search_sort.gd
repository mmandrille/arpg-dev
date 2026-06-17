# Headless tests for market search and sort controls.
# Run via: godot --headless --path client --script res://tests/test_market_search_sort.gd
extends SceneTree

const MarketPanelScript := preload("res://scripts/market_panel.gd")
const MarketRowFiltersScript := preload("res://scripts/market_row_filters.gd")

var _pass: int = 0
var _fail: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_test_listing_filter_sort()
	_test_offer_and_receipt_filter_sort()
	await _test_panel_controls_filter_debug_rows()
	if _fail == 0:
		print("[gdtest] PASS: test_market_search_sort (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_market_search_sort (%d failures, %d assertions)" % [_fail, _pass])
		quit(1)


func _test_listing_filter_sort() -> void:
	var listings := [
		_listing("listing-mail-low", "cave_mail", "Iron Mail", "seller-a", 42),
		_listing("listing-blade", "cave_blade", "Cave Blade", "seller-b", 9),
		_listing("listing-mail-high", "cave_mail", "Aegis Mail", "seller-c", 80),
	]
	var rows := MarketRowFiltersScript.filter_sort_listings(listings, "mail", "price_high")
	_assert_eq("mail filter count", rows.size(), 2)
	_assert_eq("price high first", str((rows[0] as Dictionary).get("listing_id", "")), "listing-mail-high")
	_assert_false("source listings not tagged", (listings[0] as Dictionary).has("_market_source_index"))
	rows = MarketRowFiltersScript.filter_sort_listings(listings, "seller-b", "default")
	_assert_eq("seller search count", rows.size(), 1)
	_assert_eq("seller search row", str((rows[0] as Dictionary).get("item_def_id", "")), "cave_blade")
	rows = MarketRowFiltersScript.filter_sort_listings(listings, "", "name")
	_assert_eq("name sort first", str((rows[0] as Dictionary).get("display_name", "")), "Aegis Mail")


func _test_offer_and_receipt_filter_sort() -> void:
	var offers := [{
		"offer_id": "offer-active",
		"listing_id": "listing-mail",
		"status": "active",
		"bidder_account_id": "bidder-a",
		"listing": _listing("listing-mail", "cave_mail", "Iron Mail", "seller-a", 42),
		"items": [{"item_def_id": "cave_blade", "display_name": "Cave Blade"}],
	}, {
		"offer_id": "offer-canceled",
		"listing_id": "listing-ring",
		"status": "canceled",
		"bidder_account_id": "bidder-b",
		"listing": _listing("listing-ring", "cave_ring", "Cave Ring", "seller-a", 70),
		"items": [{"item_def_id": "red_potion", "display_name": "Red Potion"}],
	}]
	var offer_rows := MarketRowFiltersScript.filter_sort_offers(offers, "blade", "status")
	_assert_eq("offer item filter count", offer_rows.size(), 1)
	_assert_eq("offer item filter id", str((offer_rows[0] as Dictionary).get("offer_id", "")), "offer-active")
	offer_rows = MarketRowFiltersScript.filter_sort_offers(offers, "", "price_high")
	_assert_eq("offer price high", str((offer_rows[0] as Dictionary).get("offer_id", "")), "offer-canceled")

	var receipts := [
		{"action": "offer_canceled", "item_def_id": "cave_mail", "offer_id": "offer-canceled", "created_at": "2026-06-17T12:00:00Z"},
		{"action": "listing_purchased", "item_def_id": "cave_ring", "offer_id": "", "created_at": "2026-06-17T11:00:00Z"},
	]
	var receipt_rows := MarketRowFiltersScript.filter_sort_receipts(receipts, "canceled", "status")
	_assert_eq("receipt action filter count", receipt_rows.size(), 1)
	_assert_eq("receipt action row", str((receipt_rows[0] as Dictionary).get("action", "")), "offer_canceled")


func _test_panel_controls_filter_debug_rows() -> void:
	var panel := MarketPanelScript.new()
	root.add_child(panel)
	await process_frame
	panel.show_market("market-1", [
		_listing("listing-mail", "cave_mail", "Iron Mail", "other-a", 42),
		_listing("listing-blade", "cave_blade", "Cave Blade", "other-b", 9),
		_listing("listing-owned", "cave_bow", "Cave Bow", "me", 22),
	], [], "me", "Active listings")
	var state := panel.get_debug_state()
	_assert_true("market filter visible", bool(state.get("market_filter_visible", false)))
	_assert_eq("initial filter listing count", int(state.get("filtered_listing_count", -1)), 2)
	_assert_eq("initial owned listing count", int(state.get("filtered_owned_listing_count", -1)), 1)

	panel.bot_set_market_search("mail")
	state = panel.get_debug_state()
	_assert_eq("panel search text", str(state.get("market_search_text", "")), "mail")
	_assert_eq("panel search filters browse", int(state.get("filtered_listing_count", -1)), 1)
	_assert_eq("panel search row", str(((state.get("listing_rows", []) as Array)[0] as Dictionary).get("item_def_id", "")), "cave_mail")

	panel.bot_set_market_search("")
	panel.bot_select_market_sort("price_high")
	state = panel.get_debug_state()
	_assert_eq("panel sort mode", str(state.get("market_sort_mode", "")), "price_high")
	_assert_eq("panel high sort first", str(((state.get("listing_rows", []) as Array)[0] as Dictionary).get("item_def_id", "")), "cave_mail")
	panel.bot_select_market_sort("price_low")
	state = panel.get_debug_state()
	_assert_eq("panel low sort first", str(((state.get("listing_rows", []) as Array)[0] as Dictionary).get("item_def_id", "")), "cave_blade")
	panel.queue_free()


func _listing(listing_id: String, item_def_id: String, display_name: String, seller: String, price: int) -> Dictionary:
	return {
		"listing_id": listing_id,
		"item_def_id": item_def_id,
		"item_template_id": item_def_id,
		"display_name": display_name,
		"seller_account_id": seller,
		"price_gold": price,
		"status": "active",
		"expires_at": "2026-06-20T12:30:00Z",
	}


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass += 1
	else:
		_fail += 1
		push_error("assertion failed: " + label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _assert_eq(label: String, got, want) -> void:
	_assert_true("%s got=%s want=%s" % [label, str(got), str(want)], got == want)
