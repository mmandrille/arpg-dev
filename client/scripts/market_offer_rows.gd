class_name MarketOfferRows
extends RefCounted
static func offer_row(panel: MarketPanel, offer: Dictionary, outgoing: bool = false) -> Control:
	var row := PanelContainer.new()
	row.add_theme_stylebox_override("panel", panel._row_style())
	var box := HBoxContainer.new()
	box.add_theme_constant_override("separation", 8)
	row.add_child(box)
	var info := VBoxContainer.new()
	info.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	box.add_child(info)
	_add_offer_text(panel, info, offer, outgoing)
	info.add_child(panel._offer_items_grid(panel._offer_items(offer)))
	var btn := Button.new()
	btn.text = "Cancel" if outgoing else "Accept"
	btn.custom_minimum_size = Vector2(110, 38)
	btn.disabled = str(offer.get("status", "active")) != "active"
	btn.pressed.connect(func() -> void:
		panel.market_action_requested.emit("cancel_offer" if outgoing else "accept_offer", {"listing_id": str(offer.get("listing_id", panel.selected_listing_id)), "offer_id": str(offer.get("offer_id", ""))})
	)
	box.add_child(btn)
	return row
static func my_offers_button(panel: MarketPanel) -> Control:
	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 8)
	var btn := Button.new()
	btn.text = "My Offers"
	btn.custom_minimum_size = Vector2(136, 34)
	btn.pressed.connect(func() -> void: panel.market_action_requested.emit("list_my_offers", {}))
	row.add_child(btn)
	return row
static func receipts_button(panel: MarketPanel) -> Control:
	var row := HBoxContainer.new()
	var btn := Button.new()
	btn.text = "Receipts"
	btn.custom_minimum_size = Vector2(136, 34)
	btn.pressed.connect(func() -> void: panel.market_action_requested.emit("list_market_receipts", {}))
	row.add_child(btn)
	return row
static func receipt_row(panel: MarketPanel, receipt: Dictionary) -> Control:
	var row := PanelContainer.new()
	row.add_theme_stylebox_override("panel", panel._row_style())
	var info := VBoxContainer.new()
	row.add_child(info)
	var title := Label.new()
	title.text = str(receipt.get("action", "")).replace("_", " ").capitalize()
	title.add_theme_font_size_override("font_size", MarketPanel.BODY_FONT_SIZE)
	title.add_theme_color_override("font_color", Color("#e8dcc8"))
	info.add_child(title)
	var detail := Label.new()
	detail.text = "%s %s" % [str(receipt.get("item_def_id", "")).replace("_", " ").capitalize(), str(receipt.get("created_at", "")).substr(0, 19)]
	detail.add_theme_font_size_override("font_size", MarketPanel.DETAIL_FONT_SIZE)
	detail.add_theme_color_override("font_color", Color("#b9aa8a"))
	info.add_child(detail)
	return row
static func show_receipts(panel: MarketPanel, receipts: Array, status: String = "") -> void:
	panel.selected_listing_id = ""
	panel.market_receipts = panel._dup_array(receipts)
	panel.offer_view_mode = "receipts"
	if status != "": panel._status_label.text = status
	panel._offer_tab_visible = true; panel._apply_offer_tab_visibility(); panel._tabs.current_tab = 2; panel._rebuild_offer_rows()
static func rebuild_receipts(panel: MarketPanel, rows: VBoxContainer, receipts: Array) -> void:
	if receipts.is_empty():
		rows.add_child(panel._empty_label("No receipts"))
		return
	for receipt in receipts:
		if typeof(receipt) == TYPE_DICTIONARY:
			rows.add_child(receipt_row(panel, receipt as Dictionary))
static func debug_receipt_rows(receipts: Array) -> Array:
	var out: Array = []
	for receipt in receipts:
		if typeof(receipt) == TYPE_DICTIONARY:
			var rec := receipt as Dictionary
			out.append({"action": str(rec.get("action", "")), "listing_id": str(rec.get("listing_id", "")), "offer_id": str(rec.get("offer_id", "")), "item_def_id": str(rec.get("item_def_id", "")), "stash_item_id": str(rec.get("stash_item_id", ""))})
	return out
static func debug_offer_rows(panel: MarketPanel, offers: Array) -> Array:
	var rows: Array = []
	for offer in offers:
		if typeof(offer) != TYPE_DICTIONARY:
			continue
		var rec := offer as Dictionary
		rows.append({
			"offer_id": str(rec.get("offer_id", "")),
			"listing_id": str(rec.get("listing_id", "")),
			"bidder_account_id": str(rec.get("bidder_account_id", "")),
			"status": str(rec.get("status", "")),
			"listing_item_def_id": str(panel._offer_listing(rec).get("item_def_id", "")),
			"listing_price_gold": int(panel._offer_listing(rec).get("price_gold", 0)),
			"item_count": panel._offer_items(rec).size(),
			"item_def_ids": panel._offer_item_def_ids(rec),
			"item_slots": panel._debug_offer_item_slots(rec),
		})
	return rows
static func _add_offer_text(panel: MarketPanel, info: VBoxContainer, offer: Dictionary, outgoing: bool) -> void:
	var title := Label.new()
	title.text = "%d item offer" % panel._offer_items(offer).size()
	title.add_theme_font_size_override("font_size", MarketPanel.BODY_FONT_SIZE)
	title.add_theme_color_override("font_color", Color("#e8dcc8"))
	info.add_child(title)
	var detail := Label.new()
	detail.text = panel._outgoing_offer_detail(offer) if outgoing else "Bidder %s" % str(offer.get("bidder_account_id", "")).substr(0, 10)
	detail.add_theme_font_size_override("font_size", MarketPanel.DETAIL_FONT_SIZE)
	detail.add_theme_color_override("font_color", Color("#b9aa8a"))
	info.add_child(detail)
	if outgoing:
		info.add_child(panel._empty_label("Listing: %s" % panel._listing_title(panel._offer_listing(offer))))
