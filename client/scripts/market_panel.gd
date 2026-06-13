class_name MarketPanel
extends Control

signal market_action_requested(action: String, payload: Dictionary)
signal inventory_context_requested(context: String)
signal staged_offer_items_changed(item_instance_ids: Array)

const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const ItemIconDrawerScript := preload("res://scripts/item_icon_drawer.gd")
const ItemTooltipPanelScript := preload("res://scripts/item_tooltip_panel.gd")
const StatLabels := preload("res://scripts/stat_labels.gd")
const PANEL_SIZE := Vector2(640, 520)
const BODY_FONT_SIZE := 19
const DETAIL_FONT_SIZE := 16
const LISTING_ICON_SIZE := Vector2(52, 52)
const STAGE_SLOT_SIZE := Vector2(68, 54)
const DEFAULT_PUBLISH_PRICE_GOLD := 25

var market_entity_id: String = ""
var account_id: String = ""
var listings: Array = []
var inventory_items: Array = []
var active_offers: Array = []
var selected_listing_id: String = ""
var staged_publish_item: Dictionary = {}
var staged_offer_items: Array = []
var _offer_tab_visible: bool = false
var _panel: DraggableWindow
var _status_label: Label
var _tabs: TabContainer
var _browse_rows: VBoxContainer
var _publish_rows: VBoxContainer
var _offer_rows: VBoxContainer
var _publish_price_spin: SpinBox
var _publish_button: Button

class MarketItemIcon:
	extends Control

	var item: Dictionary = {}
	var presentations: Dictionary = {}

	func setup(next_item: Dictionary, next_presentations: Dictionary) -> void:
		item = next_item.duplicate(true)
		presentations = next_presentations.duplicate(true)
		custom_minimum_size = LISTING_ICON_SIZE
		queue_redraw()

	func _draw() -> void:
		draw_rect(Rect2(Vector2.ZERO, size), Color("#0a0908"), true)
		draw_rect(Rect2(Vector2.ZERO, size), Color("#5c4a1f"), false, 1.0)
		var def_id := str(item.get("item_def_id", ""))
		var icon: Dictionary = presentations.get(def_id, {}).get("icon", {})
		ItemIconDrawerScript.draw(self, Rect2(Vector2.ZERO, size), icon, str(icon.get("label", _short_label(def_id))), false, 0.38, 20)

	func _short_label(def_id: String) -> String:
		if def_id == "":
			return "?"
		var out := ""
		for part in def_id.split("_"):
			if str(part).length() > 0:
				out += str(part).substr(0, 1).to_upper()
		return out.substr(0, 3)

class MarketStageSlot:
	extends Button

	var panel: MarketPanel
	var slot_context: String = "publish"
	var slot_index: int = 0
	var item: Dictionary = {}

	func _draw() -> void:
		if item.is_empty():
			return
		panel._draw_item_icon(self, item)

	func _make_custom_tooltip(for_text: String) -> Object:
		if item.is_empty():
			return panel._make_text_tooltip(for_text)
		return panel._make_item_tooltip(item)

	func _get_drag_data(_at_position: Vector2) -> Variant:
		if item.is_empty() or slot_context == "view_offer":
			return null
		return {"source": "market_stage", "context": slot_context, "slot_index": slot_index, "item": item}

	func _can_drop_data(_at_position: Vector2, data: Variant) -> bool:
		if slot_context == "view_offer":
			return false
		return typeof(data) == TYPE_DICTIONARY and str(data.get("source", "")) == "bag" and typeof(data.get("item", {})) == TYPE_DICTIONARY

	func _drop_data(_at_position: Vector2, data: Variant) -> void:
		if slot_context == "view_offer":
			return
		panel.stage_inventory_item(str(slot_context), data.get("item", {}), slot_index)

func _ready() -> void:
	ItemRulesLoader.ensure_loaded()
	_build()
	hide_display()


func show_market(entity_id: String, next_listings: Array, next_stash_items: Array = [], next_account_id: String = "", status: String = "") -> void:
	if _panel == null:
		_build()
	market_entity_id = entity_id
	listings = _dup_array(next_listings)
	inventory_items = _dup_array(next_stash_items)
	active_offers = []
	account_id = next_account_id
	staged_publish_item = {}
	staged_offer_items = []
	active_offers = []
	selected_listing_id = ""
	_offer_tab_visible = false
	staged_offer_items_changed.emit([])
	_status_label.text = status
	_rebuild_all()
	visible = true
	_panel.visible = true
	_panel.clamp_to_viewport()


func hide_display() -> void:
	if not staged_offer_items.is_empty():
		staged_offer_items = []
		staged_offer_items_changed.emit([])
	visible = false
	if _panel != null:
		_panel.visible = false


func show_status(message: String, warning: bool = false) -> void:
	if _status_label == null:
		return
	_status_label.text = message
	_status_label.add_theme_color_override("font_color", Color("#ffcf5a") if warning else Color("#9fd7ff"))


func return_to_browse_after_offer() -> void:
	staged_offer_items = []
	staged_offer_items_changed.emit([])
	if _tabs != null:
		_tabs.current_tab = 0
	_offer_tab_visible = false
	_apply_offer_tab_visibility()
	inventory_context_requested.emit("")
	_rebuild_offer_rows()


func bot_select_tab(tab_name: String) -> void:
	if _tabs == null:
		return
	match tab_name:
		"publish":
			_tabs.current_tab = 1
			inventory_context_requested.emit("publish")
		"offer":
			_offer_tab_visible = true
			_apply_offer_tab_visibility()
			_tabs.current_tab = 2
			inventory_context_requested.emit("offer")
		_:
			_tabs.current_tab = 0
			inventory_context_requested.emit("")


func bot_set_publish_price(price_gold: int) -> void:
	if _publish_price_spin == null:
		return
	_publish_price_spin.value = max(1, price_gold)


func bot_click_publish_stash_item(stash_item_id: String = "", item_def_id: String = "", rolled: Variant = null, stash_index: int = 0) -> void:
	var item := _matching_inventory_item(stash_item_id, item_def_id, rolled, stash_index)
	if item.is_empty():
		show_status("No matching inventory item to publish", true)
		return
	stage_inventory_item("publish", item)
	_emit_publish_action()


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
	_offer_tab_visible = true
	_apply_offer_tab_visibility()
	_tabs.current_tab = 2
	_rebuild_offer_rows()


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"market_entity_id": market_entity_id,
		"account_id": account_id,
		"listing_count": listings.size(),
		"listing_rows": _debug_listing_rows(_foreign_listings()),
		"owned_listing_rows": _debug_listing_rows(_owned_listings()),
		"stash_item_count": inventory_items.size(),
		"stash_rows": _debug_inventory_rows(),
		"inventory_item_count": inventory_items.size(),
		"inventory_rows": _debug_inventory_rows(),
		"staged_publish_item": staged_publish_item.duplicate(true),
		"staged_offer_count": staged_offer_items.size(),
		"staged_offer_item_ids": _staged_offer_item_ids(),
		"staged_offer_slots": _debug_staged_offer_slots(),
		"offer_count": active_offers.size(),
		"offer_rows": _debug_offer_rows(),
		"publish_price_gold": _publish_price(),
		"publish_price_width": int(_publish_price_spin.custom_minimum_size.x) if _publish_price_spin != null else 0,
		"publish_button_width": int(_publish_button.custom_minimum_size.x) if _publish_button != null else 0,
		"publish_button_same_row": _publish_button != null and _publish_price_spin != null and _publish_button.get_parent() == _publish_price_spin.get_parent(),
		"publish_rows_centered": _rows_centered(_publish_rows),
		"offer_rows_top_aligned": _rows_top_aligned(_offer_rows),
		"selected_listing_id": selected_listing_id,
		"offer_tab_visible": _offer_tab_visible,
		"visible_tab_titles": _visible_tab_titles(),
		"status": _status_label.text if _status_label != null else "",
		"tab": _tabs.current_tab if _tabs != null else -1,
		"window": _panel.get_debug_state() if _panel != null else {},
	}


func stage_inventory_item(context: String, item: Dictionary, slot_index: int = -1) -> void:
	if item.is_empty():
		return
	if context == "publish":
		staged_publish_item = item.duplicate(true)
		show_status("Ready to publish %s" % _item_title(staged_publish_item))
		_rebuild_publish_rows()
		return
	if context == "offer":
		var item_id := str(item.get("item_instance_id", ""))
		if item_id == "":
			show_status("Missing item id", true)
			return
		for staged in staged_offer_items:
			if typeof(staged) == TYPE_DICTIONARY and str((staged as Dictionary).get("item_instance_id", "")) == item_id:
				show_status("Item already in offer", true)
				return
		if staged_offer_items.size() >= 10:
			show_status("Offer is full", true)
			return
		if slot_index >= 0 and slot_index < staged_offer_items.size():
			staged_offer_items[slot_index] = item.duplicate(true)
		else:
			staged_offer_items.append(item.duplicate(true))
		show_status("Offer staged: %d/10" % staged_offer_items.size())
		staged_offer_items_changed.emit(_staged_offer_item_ids())
		_rebuild_offer_rows()


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
	_tabs.tab_changed.connect(func(tab: int) -> void:
		if tab == 1:
			inventory_context_requested.emit("publish")
		elif tab == 2:
			inventory_context_requested.emit("offer")
		else:
			inventory_context_requested.emit("")
	)
	root.add_child(_tabs)

	_browse_rows = _tab_rows("Browse")
	_publish_rows = _tab_rows("Publish")
	_offer_rows = _tab_rows("Offer")
	_apply_offer_tab_visibility()


func _tab_rows(title: String) -> VBoxContainer:
	var scroll := ScrollContainer.new()
	scroll.name = title
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_tabs.add_child(scroll)
	var rows := VBoxContainer.new()
	rows.add_theme_constant_override("separation", 6)
	if title == "Browse":
		scroll.add_child(rows)
	elif title == "Publish":
		var center := CenterContainer.new()
		center.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		center.size_flags_vertical = Control.SIZE_EXPAND_FILL
		scroll.add_child(center)
		rows.alignment = BoxContainer.ALIGNMENT_CENTER
		rows.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
		rows.size_flags_vertical = Control.SIZE_SHRINK_CENTER
		center.add_child(rows)
	else:
		var margin := MarginContainer.new()
		margin.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		margin.size_flags_vertical = Control.SIZE_EXPAND_FILL
		margin.add_theme_constant_override("margin_top", 16)
		scroll.add_child(margin)
		rows.alignment = BoxContainer.ALIGNMENT_BEGIN
		rows.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
		rows.size_flags_vertical = Control.SIZE_SHRINK_BEGIN
		margin.add_child(rows)
	return rows


func _rebuild_all() -> void:
	_rebuild_browse_rows()
	_rebuild_publish_rows()
	_rebuild_offer_rows()


func _rebuild_browse_rows() -> void:
	_clear_rows(_browse_rows)
	var browse_listings := _foreign_listings()
	if browse_listings.is_empty():
		_browse_rows.add_child(_empty_label("No active listings"))
		return
	for listing in browse_listings:
		_browse_rows.add_child(_listing_row(listing as Dictionary, "browse"))


func _rebuild_publish_rows() -> void:
	_clear_rows(_publish_rows)
	var owned := _owned_listings()
	for listing in owned:
		_publish_rows.add_child(_listing_row(listing as Dictionary, "publish"))
	_publish_rows.add_child(_stage_slot("publish", staged_publish_item, 0))
	_publish_rows.add_child(_publish_controls_row())
	_publish_rows.add_child(_empty_label("Double-click or drag an inventory item here"))


func _rebuild_offer_rows() -> void:
	_clear_rows(_offer_rows)
	var selected := _selected_listing()
	if selected.is_empty():
		_offer_rows.add_child(_empty_label("Select another player's listing in Browse"))
		return
	_offer_rows.add_child(_listing_row(selected, "readonly"))
	if str(selected.get("seller_account_id", "")) == account_id:
		if active_offers.is_empty():
			_offer_rows.add_child(_empty_label("No active offers"))
			return
		for offer in active_offers:
			if typeof(offer) == TYPE_DICTIONARY:
				_offer_rows.add_child(_offer_row(offer as Dictionary))
		return
	_offer_rows.add_child(_offer_grid())
	var offer_btn := Button.new()
	offer_btn.text = "Offer"
	offer_btn.custom_minimum_size = Vector2(140, 38)
	offer_btn.disabled = staged_offer_items.is_empty()
	offer_btn.pressed.connect(_emit_offer_action)
	_offer_rows.add_child(offer_btn)
	_offer_rows.add_child(_empty_label("Double-click or drag up to 10 inventory items"))


func _listing_row(listing: Dictionary, mode: String) -> Control:
	var row := PanelContainer.new()
	row.add_theme_stylebox_override("panel", _row_style())
	var outer := HBoxContainer.new()
	outer.add_theme_constant_override("separation", 8)
	row.add_child(outer)

	var icon := MarketItemIcon.new()
	icon.setup(listing, ItemRulesLoader.item_presentations)
	outer.add_child(icon)

	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 4)
	box.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	outer.add_child(box)

	var title := Label.new()
	title.text = _listing_title(listing)
	title.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	title.add_theme_color_override("font_color", _rarity_color(str(listing.get("rarity", "common"))))
	box.add_child(title)

	var detail := Label.new()
	detail.text = "%d gold - seller %s" % [int(listing.get("price_gold", 0)), str(listing.get("seller_account_id", "")).substr(0, 10)]
	detail.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	detail.add_theme_color_override("font_color", Color("#b9aa8a"))
	box.add_child(detail)

	for stat_line in _listing_stat_lines(listing):
		var stat_label := Label.new()
		stat_label.text = stat_line
		stat_label.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
		stat_label.add_theme_color_override("font_color", Color("#cfc3aa"))
		box.add_child(stat_label)

	if mode != "readonly":
		var actions := HBoxContainer.new()
		actions.add_theme_constant_override("separation", 8)
		if mode == "publish":
			var offers_btn := Button.new()
			offers_btn.text = "View Offers"
			offers_btn.custom_minimum_size = Vector2(136, 34)
			offers_btn.pressed.connect(func() -> void:
				selected_listing_id = str(listing.get("listing_id", ""))
				market_action_requested.emit("list_offers", {"listing_id": selected_listing_id})
			)
			actions.add_child(offers_btn)
		elif mode == "browse" and int(listing.get("price_gold", 0)) > 0:
			var buy_btn := Button.new()
			buy_btn.text = "Buy"
			buy_btn.custom_minimum_size = Vector2(86, 34)
			buy_btn.pressed.connect(func() -> void:
				_emit_purchase_action(listing)
			)
			actions.add_child(buy_btn)
		if mode == "browse":
			var btn := Button.new()
			btn.text = "Make Offer"
			btn.custom_minimum_size = Vector2(136, 34)
			btn.pressed.connect(func() -> void:
				selected_listing_id = str(listing.get("listing_id", ""))
				_offer_tab_visible = true
				_apply_offer_tab_visibility()
				_tabs.current_tab = 2
				inventory_context_requested.emit("offer")
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
	detail.text = "Bidder %s" % str(offer.get("bidder_account_id", "")).substr(0, 10)
	detail.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	detail.add_theme_color_override("font_color", Color("#b9aa8a"))
	info.add_child(detail)
	info.add_child(_offer_items_grid(_offer_items(offer)))
	var btn := Button.new()
	btn.text = "Accept"
	btn.custom_minimum_size = Vector2(110, 38)
	btn.disabled = str(offer.get("status", "active")) != "active"
	btn.pressed.connect(func() -> void:
		market_action_requested.emit("accept_offer", {"listing_id": str(offer.get("listing_id", selected_listing_id)), "offer_id": str(offer.get("offer_id", ""))})
	)
	box.add_child(btn)
	return row


func _stage_slot(context: String, item: Dictionary, slot_index: int) -> Control:
	var btn := MarketStageSlot.new()
	btn.panel = self
	btn.slot_context = context
	btn.slot_index = slot_index
	btn.item = item.duplicate(true)
	btn.custom_minimum_size = STAGE_SLOT_SIZE
	btn.text = "" if not item.is_empty() else "Empty"
	btn.tooltip_text = _item_detail(item) if not item.is_empty() else "Drop inventory item"
	btn.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	btn.add_theme_color_override("font_color", _rarity_color(str(item.get("rarity", "common"))) if not item.is_empty() else Color("#8f826b"))
	btn.add_theme_stylebox_override("normal", _stage_slot_style(str(item.get("rarity", "common")), false))
	btn.add_theme_stylebox_override("hover", _stage_slot_style(str(item.get("rarity", "common")), true))
	btn.add_theme_stylebox_override("pressed", _stage_slot_style(str(item.get("rarity", "common")), true))
	return btn


func _offer_grid() -> Control:
	var grid := GridContainer.new()
	grid.columns = 5
	grid.add_theme_constant_override("h_separation", 6)
	grid.add_theme_constant_override("v_separation", 6)
	for i in range(10):
		var item: Dictionary = staged_offer_items[i] if i < staged_offer_items.size() else {}
		grid.add_child(_stage_slot("offer", item, i))
	return grid


func _offer_items_grid(items: Array) -> Control:
	var grid := GridContainer.new()
	grid.columns = 5
	grid.add_theme_constant_override("h_separation", 6)
	grid.add_theme_constant_override("v_separation", 6)
	var index := 0
	for value in items:
		if typeof(value) != TYPE_DICTIONARY:
			continue
		grid.add_child(_stage_slot("view_offer", value as Dictionary, index))
		index += 1
	return grid


func _emit_publish_action() -> void:
	if staged_publish_item.is_empty():
		show_status("Choose an inventory item first", true)
		return
	market_action_requested.emit("publish_inventory", {
		"item_instance_id": str(staged_publish_item.get("item_instance_id", "")),
		"price_gold": _publish_price(),
	})


func _emit_offer_action() -> void:
	if selected_listing_id == "":
		show_status("Select a listing first", true)
		return
	var ids: Array = []
	for item in staged_offer_items:
		if typeof(item) == TYPE_DICTIONARY:
			ids.append(str((item as Dictionary).get("item_instance_id", "")))
	if ids.is_empty():
		show_status("Choose inventory items first", true)
		return
	market_action_requested.emit("offer_inventory", {"listing_id": selected_listing_id, "item_instance_ids": ids})


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


func _owned_listings() -> Array:
	var out: Array = []
	for listing in listings:
		if typeof(listing) == TYPE_DICTIONARY and str((listing as Dictionary).get("seller_account_id", "")) == account_id:
			out.append(listing)
	return out


func _foreign_listings() -> Array:
	var out: Array = []
	for listing in listings:
		if typeof(listing) == TYPE_DICTIONARY and str((listing as Dictionary).get("seller_account_id", "")) != account_id:
			out.append(listing)
	return out


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


func _publish_controls_row() -> Control:
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
	_publish_price_spin.custom_minimum_size = Vector2(180, 38)
	_publish_price_spin.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(_publish_price_spin)
	_publish_button = Button.new()
	_publish_button.text = "Publish"
	_publish_button.custom_minimum_size = Vector2(180, 38)
	_publish_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_publish_button.disabled = staged_publish_item.is_empty()
	_publish_button.pressed.connect(_emit_publish_action)
	row.add_child(_publish_button)
	return row


func _publish_price() -> int:
	if _publish_price_spin == null:
		return DEFAULT_PUBLISH_PRICE_GOLD
	return max(1, int(_publish_price_spin.value))


func _matching_inventory_item(stash_item_id: String = "", item_def_id: String = "", rolled: Variant = null, stash_index: int = 0) -> Dictionary:
	var matches: Array = []
	for item in inventory_items:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := item as Dictionary
		if stash_item_id != "" and str(rec.get("item_instance_id", rec.get("stash_item_id", ""))) != stash_item_id:
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


func _debug_listing_rows(source: Array) -> Array:
	var rows: Array = []
	for listing in source:
		if typeof(listing) != TYPE_DICTIONARY:
			continue
		var rec := listing as Dictionary
		rows.append({
			"listing_id": str(rec.get("listing_id", "")),
			"item_def_id": str(rec.get("item_def_id", "")),
			"item_template_id": str(rec.get("item_template_id", "")),
			"seller_account_id": str(rec.get("seller_account_id", "")),
			"price_gold": int(rec.get("price_gold", 0)),
			"visible_detail": "%d gold - seller %s" % [int(rec.get("price_gold", 0)), str(rec.get("seller_account_id", "")).substr(0, 10)],
			"has_icon": str(rec.get("item_def_id", "")) != "",
			"stat_lines": _listing_stat_lines(rec),
		})
	return rows


func _apply_offer_tab_visibility() -> void:
	if _tabs == null:
		return
	if _tabs.get_tab_count() >= 3:
		_tabs.set_tab_hidden(2, not _offer_tab_visible)


func _visible_tab_titles() -> Array:
	var titles: Array = []
	if _tabs == null:
		return titles
	for i in range(_tabs.get_tab_count()):
		if not _tabs.is_tab_hidden(i):
			titles.append(_tabs.get_tab_title(i))
	return titles


func _debug_inventory_rows() -> Array:
	var rows: Array = []
	for item in inventory_items:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := item as Dictionary
		rows.append({
			"item_instance_id": str(rec.get("item_instance_id", "")),
			"stash_item_id": str(rec.get("stash_item_id", "")),
			"item_def_id": str(rec.get("item_def_id", "")),
			"item_template_id": str(rec.get("item_template_id", "")),
		})
	return rows


func _staged_offer_item_ids() -> Array:
	var ids: Array = []
	for item in staged_offer_items:
		if typeof(item) == TYPE_DICTIONARY:
			ids.append(str((item as Dictionary).get("item_instance_id", "")))
	return ids


func _debug_staged_offer_slots() -> Array:
	var slots: Array = []
	for i in range(10):
		var item: Dictionary = staged_offer_items[i] if i < staged_offer_items.size() and typeof(staged_offer_items[i]) == TYPE_DICTIONARY else {}
		slots.append({
			"index": i,
			"occupied": not item.is_empty(),
			"item_def_id": str(item.get("item_def_id", "")),
			"has_icon": not item.is_empty() and str(item.get("item_def_id", "")) != "",
			"slot_size": {"x": int(STAGE_SLOT_SIZE.x), "y": int(STAGE_SLOT_SIZE.y)},
			"uses_shared_tooltip": not item.is_empty(),
		})
	return slots


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
			"item_slots": _debug_offer_item_slots(rec),
		})
	return rows


func _rows_centered(rows: VBoxContainer) -> bool:
	if rows == null:
		return false
	return rows.get_parent() is CenterContainer \
		and rows.alignment == BoxContainer.ALIGNMENT_CENTER \
		and rows.size_flags_horizontal == Control.SIZE_SHRINK_CENTER \
		and rows.size_flags_vertical == Control.SIZE_SHRINK_CENTER


func _rows_top_aligned(rows: VBoxContainer) -> bool:
	if rows == null:
		return false
	return rows.get_parent() is MarginContainer \
		and rows.alignment == BoxContainer.ALIGNMENT_BEGIN \
		and rows.size_flags_horizontal == Control.SIZE_SHRINK_CENTER \
		and rows.size_flags_vertical == Control.SIZE_SHRINK_BEGIN


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


func _debug_offer_item_slots(offer: Dictionary) -> Array:
	var slots: Array = []
	var index := 0
	for item in _offer_items(offer):
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := item as Dictionary
		slots.append({
			"index": index,
			"item_def_id": str(rec.get("item_def_id", "")),
			"has_icon": str(rec.get("item_def_id", "")) != "",
			"uses_shared_tooltip": true,
			"slot_size": {"x": int(STAGE_SLOT_SIZE.x), "y": int(STAGE_SLOT_SIZE.y)},
		})
		index += 1
	return slots


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


func _make_item_tooltip(item: Dictionary) -> Control:
	var tooltip := ItemTooltipPanelScript.new()
	tooltip.setup(
		item,
		ItemRulesLoader.item_presentations,
		_tooltip_lines(item),
		_requirement_lines(item),
		[],
		-1,
		true,
		_short_label(str(item.get("item_def_id", "")))
	)
	return tooltip


func _make_text_tooltip(text: String) -> Control:
	var tooltip := ItemTooltipPanelScript.new()
	tooltip.setup({}, ItemRulesLoader.item_presentations, [text], [], [], -1, true, "")
	return tooltip


func _tooltip_lines(item: Dictionary) -> Array:
	var rarity := str(item.get("rarity", ""))
	var lines: Array = [{"text": _item_title(item), "color": _rarity_color(rarity)}]
	if rarity != "":
		lines.append({"text": "Rarity: %s" % rarity.capitalize(), "color": Color("#cdbd9f"), "font_size": 19})
	var slot := str(item.get("slot", ""))
	if slot != "":
		lines.append({"text": "Slot: %s" % slot, "color": Color("#cdbd9f"), "font_size": 19})
	for line in _listing_stat_lines(item):
		var text := str(line)
		if text.begins_with("Level "):
			continue
		lines.append(text.replace("Base ", "").replace("Rolled ", ""))
	return lines


func _requirement_lines(item: Dictionary) -> Array:
	var requirements := _item_requirements(item)
	var lines: Array = []
	if _stat_int(requirements.get("level", 0)) > 0:
		lines.append("Level %d" % _stat_int(requirements.get("level", 0)))
	for key in requirements.keys():
		var stat := str(key)
		if stat == "level":
			continue
		lines.append("%s %d" % [StatLabels.display_name(stat), _stat_int(requirements.get(key, 0))])
	return lines


func _draw_item_icon(slot: Control, item: Dictionary) -> void:
	var def_id := str(item.get("item_def_id", ""))
	var icon: Dictionary = ItemRulesLoader.item_presentations.get(def_id, {}).get("icon", {})
	var rect := Rect2(Vector2.ZERO, slot.size)
	ItemIconDrawerScript.draw(slot, rect, icon, str(icon.get("label", _short_label(def_id))), false, 0.24, 22)


func _short_label(def_id: String) -> String:
	if def_id == "":
		return "?"
	var out := ""
	for part in def_id.split("_"):
		if str(part).length() > 0:
			out += str(part).substr(0, 1).to_upper()
	return out.substr(0, 3)


func _listing_stat_lines(item: Dictionary) -> Array:
	var lines: Array = []
	var requirements: Dictionary = _item_requirements(item)
	if _stat_int(requirements.get("level", 0)) > 0:
		lines.append("Level %d" % _stat_int(requirements.get("level", 0)))
	var base_stats := _item_base_stats(item)
	for stat in _ordered_stat_keys(base_stats):
		lines.append("Base %s: %s" % [StatLabels.display_name(stat), _signed_stat_value(stat, _stat_int(base_stats.get(stat, 0)))])
	var rolled_stats: Dictionary = item.get("rolled_stats", {})
	for stat in _ordered_stat_keys(rolled_stats):
		var base := _stat_int(base_stats.get(stat, 0))
		var total := _stat_int(rolled_stats.get(stat, 0))
		var delta := total - base
		if delta != 0:
			lines.append("Rolled %s: %s" % [StatLabels.display_name(stat), _signed_stat_value(stat, delta)])
	return lines


func _stat_int(value) -> int:
	match typeof(value):
		TYPE_INT:
			return int(value)
		TYPE_FLOAT:
			return int(value)
		TYPE_STRING:
			var text := str(value)
			if text.is_valid_int():
				return int(text)
			if text.is_valid_float():
				return int(float(text))
	return 0


func _item_base_stats(item: Dictionary) -> Dictionary:
	var template_id := str(item.get("item_template_id", item.get("item_def_id", "")))
	var template: Dictionary = ItemRulesLoader.item_templates.get(template_id, {})
	return (template.get("base_stats", {}) as Dictionary).duplicate(true) if typeof(template.get("base_stats", {})) == TYPE_DICTIONARY else {}


func _item_requirements(item: Dictionary) -> Dictionary:
	var requirements = item.get("requirements", {})
	if typeof(requirements) == TYPE_DICTIONARY and not (requirements as Dictionary).is_empty():
		return (requirements as Dictionary).duplicate(true)
	var template_id := str(item.get("item_template_id", item.get("item_def_id", "")))
	var template: Dictionary = ItemRulesLoader.item_templates.get(template_id, {})
	return (template.get("requirements", {}) as Dictionary).duplicate(true) if typeof(template.get("requirements", {})) == TYPE_DICTIONARY else {}


func _ordered_stat_keys(stats: Dictionary) -> Array:
	var order := ["damage_min", "damage_max", "armor", "block_percent", "attack_speed_percent", "max_hp", "max_mana", "health_regen_per_10_seconds", "mana_regen_per_10_seconds", "skill_damage_percent", "hotbar_slots", "inventory_rows", "str", "dex", "vit", "magic", "all_skills"]
	var keys: Array = []
	for stat in order:
		if stats.has(stat):
			keys.append(stat)
	for stat in stats.keys():
		if not keys.has(str(stat)):
			keys.append(str(stat))
	return keys


func _signed_stat_value(stat: String, value: int) -> String:
	var suffix := "%" if stat in ["block_percent", "attack_speed_percent", "skill_damage_percent"] else ""
	return "%+d%s" % [value, suffix]


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


func _stage_slot_style(rarity: String, hover: bool) -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	var base := Color("#242422")
	match rarity:
		"magic":
			base = Color("#16304e")
		"rare":
			base = Color("#4b3a18")
		"unique":
			base = Color("#4b2815")
	s.bg_color = base.lightened(0.12) if hover else base
	s.border_color = base.lightened(0.46) if hover else base.lightened(0.28)
	s.set_border_width_all(1)
	s.set_content_margin_all(4)
	return s


func _dup_array(values: Array) -> Array:
	var out: Array = []
	for value in values:
		out.append((value as Dictionary).duplicate(true) if typeof(value) == TYPE_DICTIONARY else value)
	return out
