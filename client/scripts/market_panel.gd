class_name MarketPanel
extends Control

signal market_action_requested(action: String, payload: Dictionary)

const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const PANEL_SIZE := Vector2(640, 520)
const BODY_FONT_SIZE := 19
const DETAIL_FONT_SIZE := 16
const DEFAULT_PUBLISH_PRICE_GOLD := 25

var market_entity_id: String = ""
var account_id: String = ""
var listings: Array = []
var stash_items: Array = []
var active_offers: Array = []
var selected_listing_id: String = ""
var _panel: DraggableWindow
var _status_label: Label
var _tabs: TabContainer
var _browse_rows: VBoxContainer
var _publish_rows: VBoxContainer
var _offer_rows: VBoxContainer
var _publish_price_spin: SpinBox


func _ready() -> void:
	_build()
	hide_display()


func show_market(entity_id: String, next_listings: Array, next_stash_items: Array = [], next_account_id: String = "", status: String = "") -> void:
	if _panel == null:
		_build()
	market_entity_id = entity_id
	listings = _dup_array(next_listings)
	stash_items = _dup_array(next_stash_items)
	active_offers = []
	account_id = next_account_id
	_status_label.text = status
	_rebuild_all()
	visible = true
	_panel.visible = true
	_panel.clamp_to_viewport()


func hide_display() -> void:
	visible = false
	if _panel != null:
		_panel.visible = false


func show_status(message: String, warning: bool = false) -> void:
	if _status_label == null:
		return
	_status_label.text = message
	_status_label.add_theme_color_override("font_color", Color("#ffcf5a") if warning else Color("#9fd7ff"))


func bot_select_tab(tab_name: String) -> void:
	if _tabs == null:
		return
	match tab_name:
		"publish":
			_tabs.current_tab = 1
		"offer":
			_tabs.current_tab = 2
		_:
			_tabs.current_tab = 0


func bot_set_publish_price(price_gold: int) -> void:
	if _publish_price_spin == null:
		return
	_publish_price_spin.value = max(1, price_gold)


func bot_click_publish_stash_item(stash_item_id: String = "", item_def_id: String = "", rolled: Variant = null, stash_index: int = 0) -> void:
	var item := _matching_stash_item(stash_item_id, item_def_id, rolled, stash_index)
	if item.is_empty():
		show_status("No matching stash item to publish", true)
		return
	_emit_market_action("publish", item)


func bot_click_purchase_listing(listing_id: String = "", item_def_id: String = "", price_gold: int = -1, listing_index: int = 0) -> void:
	var listing := _matching_listing(listing_id, item_def_id, price_gold, listing_index)
	if listing.is_empty():
		show_status("No matching listing to purchase", true)
		return
	_emit_purchase_action(listing)


func bot_click_view_offers(listing_id: String = "", item_def_id: String = "", price_gold: int = -1, listing_index: int = 0) -> void:
	var listing := _matching_listing(listing_id, item_def_id, price_gold, listing_index, true)
	if listing.is_empty():
		show_status("No matching seller listing", true)
		return
	selected_listing_id = str(listing.get("listing_id", ""))
	market_action_requested.emit("list_offers", {"listing_id": selected_listing_id})


func bot_click_accept_offer(offer_id: String = "", offer_index: int = 0) -> void:
	var offer := _matching_offer(offer_id, offer_index)
	if offer.is_empty():
		show_status("No matching offer", true)
		return
	market_action_requested.emit("accept_offer", {"listing_id": str(offer.get("listing_id", selected_listing_id)), "offer_id": str(offer.get("offer_id", ""))})


func show_offers(listing_id: String, offers: Array, status: String = "") -> void:
	selected_listing_id = listing_id
	active_offers = _dup_array(offers)
	if status != "":
		_status_label.text = status
	_tabs.current_tab = 2
	_rebuild_offer_rows()


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"market_entity_id": market_entity_id,
		"account_id": account_id,
		"listing_count": listings.size(),
		"listing_rows": _debug_listing_rows(),
		"stash_item_count": stash_items.size(),
		"stash_rows": _debug_stash_rows(),
		"offer_count": active_offers.size(),
		"offer_rows": _debug_offer_rows(),
		"publish_price_gold": _publish_price(),
		"selected_listing_id": selected_listing_id,
		"status": _status_label.text if _status_label != null else "",
		"tab": _tabs.current_tab if _tabs != null else -1,
		"window": _panel.get_debug_state() if _panel != null else {},
	}


func _build() -> void:
	if _panel != null:
		return
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	set_anchors_preset(Control.PRESET_FULL_RECT)

	_panel = DraggableWindowScript.new()
	_panel.configure("Market Board", PANEL_SIZE)
	_panel.set_layout_key("market_panel")
	_panel.position = Vector2(340, 68)
	_panel.close_requested.connect(hide_display)
	add_child(_panel)

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 8)
	_panel.set_content(root)

	_status_label = Label.new()
	_status_label.text = ""
	_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_status_label.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	_status_label.add_theme_color_override("font_color", Color("#9fd7ff"))
	root.add_child(_status_label)

	_tabs = TabContainer.new()
	_tabs.size_flags_vertical = Control.SIZE_EXPAND_FILL
	root.add_child(_tabs)

	_browse_rows = _tab_rows("Browse")
	_publish_rows = _tab_rows("Publish")
	_offer_rows = _tab_rows("Offer")


func _tab_rows(title: String) -> VBoxContainer:
	var scroll := ScrollContainer.new()
	scroll.name = title
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_tabs.add_child(scroll)
	var rows := VBoxContainer.new()
	rows.add_theme_constant_override("separation", 6)
	scroll.add_child(rows)
	return rows


func _rebuild_all() -> void:
	_rebuild_browse_rows()
	_rebuild_publish_rows()
	_rebuild_offer_rows()


func _rebuild_browse_rows() -> void:
	_clear_rows(_browse_rows)
	if listings.is_empty():
		_browse_rows.add_child(_empty_label("No active listings"))
		return
	for listing in listings:
		if typeof(listing) == TYPE_DICTIONARY:
			_browse_rows.add_child(_listing_row(listing as Dictionary, true))


func _rebuild_publish_rows() -> void:
	_clear_rows(_publish_rows)
	_publish_rows.add_child(_publish_price_row())
	if stash_items.is_empty():
		_publish_rows.add_child(_empty_label("Your stash has no items to publish"))
		return
	for item in stash_items:
		if typeof(item) == TYPE_DICTIONARY:
			_publish_rows.add_child(_stash_item_row(item as Dictionary, "Publish", "publish"))


func _rebuild_offer_rows() -> void:
	_clear_rows(_offer_rows)
	var selected := _selected_listing()
	if selected.is_empty():
		_offer_rows.add_child(_empty_label("Select another player's listing in Browse"))
		return
	_offer_rows.add_child(_listing_row(selected, false))
	if str(selected.get("seller_account_id", "")) == account_id:
		if active_offers.is_empty():
			_offer_rows.add_child(_empty_label("No active offers"))
			return
		for offer in active_offers:
			if typeof(offer) == TYPE_DICTIONARY:
				_offer_rows.add_child(_offer_row(offer as Dictionary))
		return
	if stash_items.is_empty():
		_offer_rows.add_child(_empty_label("Your stash has no items to offer"))
		return
	for item in stash_items:
		if typeof(item) == TYPE_DICTIONARY:
			_offer_rows.add_child(_stash_item_row(item as Dictionary, "Offer", "offer"))


func _listing_row(listing: Dictionary, selectable: bool) -> Control:
	var row := PanelContainer.new()
	row.add_theme_stylebox_override("panel", _row_style())
	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 4)
	row.add_child(box)

	var title := Label.new()
	title.text = _listing_title(listing)
	title.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	title.add_theme_color_override("font_color", _rarity_color(str(listing.get("rarity", "common"))))
	box.add_child(title)

	var detail := Label.new()
	detail.text = "%d gold - Listing %s - seller %s" % [int(listing.get("price_gold", 0)), str(listing.get("listing_id", "")), str(listing.get("seller_account_id", "")).substr(0, 10)]
	detail.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	detail.add_theme_color_override("font_color", Color("#b9aa8a"))
	box.add_child(detail)

	if selectable:
		var actions := HBoxContainer.new()
		actions.add_theme_constant_override("separation", 8)
		if str(listing.get("seller_account_id", "")) == account_id:
			var offers_btn := Button.new()
			offers_btn.text = "View Offers"
			offers_btn.custom_minimum_size = Vector2(136, 34)
			offers_btn.pressed.connect(func() -> void:
				selected_listing_id = str(listing.get("listing_id", ""))
				market_action_requested.emit("list_offers", {"listing_id": selected_listing_id})
			)
			actions.add_child(offers_btn)
		elif int(listing.get("price_gold", 0)) > 0:
			var buy_btn := Button.new()
			buy_btn.text = "Buy"
			buy_btn.custom_minimum_size = Vector2(86, 34)
			buy_btn.pressed.connect(func() -> void:
				_emit_purchase_action(listing)
			)
			actions.add_child(buy_btn)
		if str(listing.get("seller_account_id", "")) != account_id:
			var btn := Button.new()
			btn.text = "Make Offer"
			btn.custom_minimum_size = Vector2(136, 34)
			btn.pressed.connect(func() -> void:
				selected_listing_id = str(listing.get("listing_id", ""))
				_tabs.current_tab = 2
				_rebuild_offer_rows()
			)
			actions.add_child(btn)
		box.add_child(actions)
	return row


func _offer_row(offer: Dictionary) -> Control:
	var row := PanelContainer.new()
	row.add_theme_stylebox_override("panel", _row_style())
	var box := HBoxContainer.new()
	box.add_theme_constant_override("separation", 8)
	row.add_child(box)
	var info := VBoxContainer.new()
	info.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	box.add_child(info)
	var title := Label.new()
	title.text = "%d item offer" % _offer_items(offer).size()
	title.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	title.add_theme_color_override("font_color", Color("#e8dcc8"))
	info.add_child(title)
	var detail := Label.new()
	detail.text = "%s - bidder %s - %s" % [str(offer.get("offer_id", "")), str(offer.get("bidder_account_id", "")).substr(0, 10), _offer_item_names(offer)]
	detail.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	detail.add_theme_color_override("font_color", Color("#b9aa8a"))
	info.add_child(detail)
	var btn := Button.new()
	btn.text = "Accept"
	btn.custom_minimum_size = Vector2(110, 38)
	btn.disabled = str(offer.get("status", "active")) != "active"
	btn.pressed.connect(func() -> void:
		market_action_requested.emit("accept_offer", {"listing_id": str(offer.get("listing_id", selected_listing_id)), "offer_id": str(offer.get("offer_id", ""))})
	)
	box.add_child(btn)
	return row


func _stash_item_row(item: Dictionary, action_label: String, action: String) -> Control:
	var row := PanelContainer.new()
	row.add_theme_stylebox_override("panel", _row_style())
	var box := HBoxContainer.new()
	box.add_theme_constant_override("separation", 8)
	row.add_child(box)

	var info := VBoxContainer.new()
	info.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	box.add_child(info)
	var title := Label.new()
	title.text = _item_title(item)
	title.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	title.add_theme_color_override("font_color", _rarity_color(str(item.get("rarity", "common"))))
	info.add_child(title)
	var detail := Label.new()
	detail.text = _item_detail(item)
	detail.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	detail.add_theme_color_override("font_color", Color("#b9aa8a"))
	info.add_child(detail)

	var btn := Button.new()
	btn.text = action_label
	btn.custom_minimum_size = Vector2(110, 38)
	btn.pressed.connect(func() -> void:
		_emit_market_action(action, item)
	)
	box.add_child(btn)
	return row


func _emit_market_action(action: String, item: Dictionary) -> void:
	var stash_item_id := str(item.get("stash_item_id", ""))
	if stash_item_id == "":
		show_status("Missing stash item id", true)
		return
	if action == "publish":
		market_action_requested.emit("publish", {
			"stash_item_id": stash_item_id,
			"price_gold": _publish_price(),
		})
		return
	if action == "offer":
		if selected_listing_id == "":
			show_status("Select a listing first", true)
			return
		market_action_requested.emit("offer", {"listing_id": selected_listing_id, "stash_item_ids": [stash_item_id]})


func _emit_purchase_action(listing: Dictionary) -> void:
	var listing_id := str(listing.get("listing_id", ""))
	if listing_id == "":
		show_status("Missing listing id", true)
		return
	market_action_requested.emit("purchase", {"listing_id": listing_id, "price_gold": int(listing.get("price_gold", 0))})


func _selected_listing() -> Dictionary:
	for listing in listings:
		if typeof(listing) == TYPE_DICTIONARY and str((listing as Dictionary).get("listing_id", "")) == selected_listing_id:
			return listing as Dictionary
	for listing in listings:
		if typeof(listing) == TYPE_DICTIONARY and str((listing as Dictionary).get("seller_account_id", "")) != account_id:
			selected_listing_id = str((listing as Dictionary).get("listing_id", ""))
			return listing as Dictionary
	return {}


func _matching_listing(listing_id: String = "", item_def_id: String = "", price_gold: int = -1, listing_index: int = 0, seller_owned: bool = false) -> Dictionary:
	var matches: Array = []
	for listing in listings:
		if typeof(listing) != TYPE_DICTIONARY:
			continue
		var rec := listing as Dictionary
		if seller_owned != (str(rec.get("seller_account_id", "")) == account_id):
			continue
		if listing_id != "" and str(rec.get("listing_id", "")) != listing_id:
			continue
		if item_def_id != "" and str(rec.get("item_def_id", "")) != item_def_id:
			continue
		if price_gold >= 0 and int(rec.get("price_gold", 0)) != price_gold:
			continue
		matches.append(rec)
	if matches.is_empty():
		return {}
	var index = clampi(listing_index, 0, matches.size() - 1)
	return (matches[index] as Dictionary).duplicate(true)


func _matching_offer(offer_id: String = "", offer_index: int = 0) -> Dictionary:
	var matches: Array = []
	for offer in active_offers:
		if typeof(offer) != TYPE_DICTIONARY:
			continue
		var rec := offer as Dictionary
		if offer_id != "" and str(rec.get("offer_id", "")) != offer_id:
			continue
		matches.append(rec)
	if matches.is_empty():
		return {}
	var index = clampi(offer_index, 0, matches.size() - 1)
	return (matches[index] as Dictionary).duplicate(true)


func _publish_price_row() -> Control:
	var current_price := _publish_price()
	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 8)
	var label := Label.new()
	label.text = "Price"
	label.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	label.add_theme_color_override("font_color", Color("#e8dcc8"))
	row.add_child(label)
	_publish_price_spin = SpinBox.new()
	_publish_price_spin.min_value = 1
	_publish_price_spin.max_value = 999999
	_publish_price_spin.step = 1
	_publish_price_spin.value = current_price
	_publish_price_spin.custom_minimum_size = Vector2(140, 34)
	row.add_child(_publish_price_spin)
	return row


func _publish_price() -> int:
	if _publish_price_spin == null:
		return DEFAULT_PUBLISH_PRICE_GOLD
	return max(1, int(_publish_price_spin.value))


func _matching_stash_item(stash_item_id: String = "", item_def_id: String = "", rolled: Variant = null, stash_index: int = 0) -> Dictionary:
	var matches: Array = []
	for item in stash_items:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := item as Dictionary
		if stash_item_id != "" and str(rec.get("stash_item_id", "")) != stash_item_id:
			continue
		if item_def_id != "" and str(rec.get("item_def_id", "")) != item_def_id:
			continue
		if rolled != null and (str(rec.get("item_template_id", "")) != "") != bool(rolled):
			continue
		matches.append(rec)
	if matches.is_empty():
		return {}
	var index = clampi(stash_index, 0, matches.size() - 1)
	return (matches[index] as Dictionary).duplicate(true)


func _debug_listing_rows() -> Array:
	var rows: Array = []
	for listing in listings:
		if typeof(listing) != TYPE_DICTIONARY:
			continue
		var rec := listing as Dictionary
		rows.append({
			"listing_id": str(rec.get("listing_id", "")),
			"item_def_id": str(rec.get("item_def_id", "")),
			"item_template_id": str(rec.get("item_template_id", "")),
			"seller_account_id": str(rec.get("seller_account_id", "")),
			"price_gold": int(rec.get("price_gold", 0)),
		})
	return rows


func _debug_stash_rows() -> Array:
	var rows: Array = []
	for item in stash_items:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := item as Dictionary
		rows.append({
			"stash_item_id": str(rec.get("stash_item_id", "")),
			"item_def_id": str(rec.get("item_def_id", "")),
			"item_template_id": str(rec.get("item_template_id", "")),
		})
	return rows


func _debug_offer_rows() -> Array:
	var rows: Array = []
	for offer in active_offers:
		if typeof(offer) != TYPE_DICTIONARY:
			continue
		var rec := offer as Dictionary
		rows.append({
			"offer_id": str(rec.get("offer_id", "")),
			"listing_id": str(rec.get("listing_id", "")),
			"bidder_account_id": str(rec.get("bidder_account_id", "")),
			"status": str(rec.get("status", "")),
			"item_count": _offer_items(rec).size(),
			"item_def_ids": _offer_item_def_ids(rec),
		})
	return rows


func _offer_items(offer: Dictionary) -> Array:
	var items: Array = offer.get("items", [])
	return items if items is Array else []


func _offer_item_names(offer: Dictionary) -> String:
	var names: Array = []
	for item in _offer_items(offer):
		if typeof(item) == TYPE_DICTIONARY:
			names.append(_item_title(item as Dictionary))
	return ", ".join(names)


func _offer_item_def_ids(offer: Dictionary) -> Array:
	var ids: Array = []
	for item in _offer_items(offer):
		if typeof(item) == TYPE_DICTIONARY:
			ids.append(str((item as Dictionary).get("item_def_id", "")))
	return ids


func _clear_rows(rows: VBoxContainer) -> void:
	for child in rows.get_children():
		child.queue_free()


func _empty_label(text: String) -> Label:
	var empty := Label.new()
	empty.text = text
	empty.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	empty.add_theme_color_override("font_color", Color("#e8dcc8"))
	return empty


func _listing_title(listing: Dictionary) -> String:
	var display := str(listing.get("display_name", ""))
	if display != "":
		return display
	return str(listing.get("item_def_id", "Unknown item")).replace("_", " ").capitalize()


func _item_title(item: Dictionary) -> String:
	var display := str(item.get("display_name", ""))
	if display != "":
		return display
	return str(item.get("item_def_id", "Unknown item")).replace("_", " ").capitalize()


func _item_detail(item: Dictionary) -> String:
	var lines: Array = item.get("summary_lines", [])
	if not lines.is_empty():
		return str(lines[0])
	var slot := str(item.get("slot", ""))
	if slot != "":
		return "Slot: %s" % slot.replace("_", " ")
	return str(item.get("stash_item_id", ""))


func _rarity_color(rarity: String) -> Color:
	match rarity:
		"magic":
			return Color("#93c5fd")
		"rare":
			return Color("#f4d481")
		"unique":
			return Color("#ffb26b")
		_:
			return Color("#e8dcc8")


func _row_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.07, 0.065, 0.052, 0.95)
	s.border_color = Color("#3b3020")
	s.set_border_width_all(1)
	s.set_content_margin_all(8)
	return s


func _dup_array(values: Array) -> Array:
	var out: Array = []
	for value in values:
		out.append((value as Dictionary).duplicate(true) if typeof(value) == TYPE_DICTIONARY else value)
	return out
