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
	btn.text = "Accept"
	btn.custom_minimum_size = Vector2(110, 38)
	btn.disabled = outgoing or str(offer.get("status", "active")) != "active"
	btn.pressed.connect(func() -> void:
		panel.market_action_requested.emit("accept_offer", {"listing_id": str(offer.get("listing_id", panel.selected_listing_id)), "offer_id": str(offer.get("offer_id", ""))})
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
