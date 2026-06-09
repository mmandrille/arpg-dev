class_name ShopPanel
extends Control

signal intent_requested(intent_type: String, payload: Dictionary)

const PANEL_SIZE := Vector2(560, 520)
const SECTION_GAP := 8

var shop_id: String = ""
var shop_entity_id: String = ""
var shop_title: String = "Vendor"
var offers: Array = []
var inventory: Array = []
var equipped: Dictionary = {}
var gold: int = 0

var _panel: PanelContainer
var _title_label: Label
var _gold_label: Label
var _status_label: Label
var _fixed_list: VBoxContainer
var _generated_list: VBoxContainer
var _sell_list: VBoxContainer
var _buy_buttons: Dictionary = {}
var _sell_buttons: Dictionary = {}
var _interactive: bool = true


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_build()
	visible = false


func show_shop(next_shop_entity_id: String, next_shop_id: String, next_offers: Array, next_gold: int, next_inventory: Array, next_equipped: Dictionary, next_title: String = "Town Vendor") -> void:
	shop_entity_id = next_shop_entity_id
	shop_id = next_shop_id
	shop_title = next_title
	offers = _dup_array(next_offers)
	set_inventory_state(next_inventory, next_equipped, next_gold)
	visible = true
	_apply_interaction_filters()
	_render()


func hide_display() -> void:
	visible = false


func set_interactive(enabled: bool) -> void:
	_interactive = enabled
	_apply_interaction_filters()


func set_inventory_state(next_inventory: Array, next_equipped: Dictionary, next_gold: int) -> void:
	inventory = _dup_array(next_inventory)
	equipped = next_equipped.duplicate(true)
	gold = max(0, next_gold)
	if _panel != null:
		_render()


func show_status(text: String, error: bool = false) -> void:
	if _status_label == null:
		return
	_status_label.text = text
	_status_label.add_theme_color_override("font_color", Color("#ff9f7a") if error else Color("#9ee6a8"))


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"shop_id": shop_id,
		"shop_entity_id": shop_entity_id,
		"gold": gold,
		"offer_count": offers.size(),
		"fixed_offer_count": _offers_by_kind("fixed").size(),
		"generated_offer_count": _offers_by_kind("generated").size(),
		"buy_buttons": _debug_buy_buttons(),
		"sell_rows": _debug_sell_rows(),
		"sell_row_count": _sellable_items().size(),
		"status": _status_label.text if _status_label != null else "",
	}


func bot_click_buy_offer(offer_id: String = "", offer_kind: String = "", offer_index: int = 0) -> void:
	var matches := _matching_offers(offer_id, offer_kind)
	if offer_index < 0 or offer_index >= matches.size():
		return
	var selected: Dictionary = matches[offer_index]
	var btn := _buy_buttons.get(str(selected.get("offer_id", "")), null) as Button
	if btn != null and btn.disabled:
		return
	_emit_buy(selected)


func bot_click_sell_item(item_def_id: String = "", rolled: Variant = null, bag_index: int = 0) -> void:
	var matches: Array = []
	for item in _sellable_items():
		if item_def_id != "" and str(item.get("item_def_id", "")) != item_def_id:
			continue
		if rolled != null and (str(item.get("item_template_id", "")) != "") != bool(rolled):
			continue
		matches.append(item)
	if bag_index < 0 or bag_index >= matches.size():
		return
	_emit_sell(matches[bag_index])


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_reposition_panel()


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = PanelContainer.new()
	_panel.custom_minimum_size = PANEL_SIZE
	_panel.set_anchors_preset(Control.PRESET_TOP_RIGHT)
	_panel.add_theme_stylebox_override("panel", _panel_style())
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	add_child(_panel)
	_reposition_panel()

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", SECTION_GAP)
	root.custom_minimum_size = Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 24)
	_panel.add_child(root)

	var header := HBoxContainer.new()
	header.add_theme_constant_override("separation", 10)
	root.add_child(header)
	_title_label = Label.new()
	_title_label.text = shop_title
	_title_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_title_label.add_theme_color_override("font_color", Color("#f4d481"))
	_title_label.add_theme_font_size_override("font_size", 28)
	header.add_child(_title_label)
	_gold_label = Label.new()
	_gold_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	_gold_label.add_theme_color_override("font_color", Color("#f4c84f"))
	_gold_label.add_theme_font_size_override("font_size", 24)
	header.add_child(_gold_label)

	_status_label = Label.new()
	_status_label.text = ""
	_status_label.clip_text = true
	_status_label.add_theme_color_override("font_color", Color("#9ee6a8"))
	root.add_child(_status_label)

	var columns := HBoxContainer.new()
	columns.add_theme_constant_override("separation", 12)
	columns.size_flags_vertical = Control.SIZE_EXPAND_FILL
	root.add_child(columns)

	var buy_col := VBoxContainer.new()
	buy_col.custom_minimum_size = Vector2(300, 0)
	buy_col.size_flags_vertical = Control.SIZE_EXPAND_FILL
	columns.add_child(buy_col)
	buy_col.add_child(_caption("Buy"))
	var buy_scroll := ScrollContainer.new()
	buy_scroll.custom_minimum_size = Vector2(300, 390)
	buy_scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	buy_col.add_child(buy_scroll)
	var buy_lists := VBoxContainer.new()
	buy_lists.add_theme_constant_override("separation", 10)
	buy_scroll.add_child(buy_lists)
	buy_lists.add_child(_subcaption("Potions"))
	_fixed_list = VBoxContainer.new()
	_fixed_list.add_theme_constant_override("separation", 5)
	buy_lists.add_child(_fixed_list)
	buy_lists.add_child(_subcaption("Generated"))
	_generated_list = VBoxContainer.new()
	_generated_list.add_theme_constant_override("separation", 5)
	buy_lists.add_child(_generated_list)

	var sell_col := VBoxContainer.new()
	sell_col.custom_minimum_size = Vector2(210, 0)
	sell_col.size_flags_vertical = Control.SIZE_EXPAND_FILL
	columns.add_child(sell_col)
	sell_col.add_child(_caption("Sell"))
	var sell_scroll := ScrollContainer.new()
	sell_scroll.custom_minimum_size = Vector2(210, 390)
	sell_scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	sell_col.add_child(sell_scroll)
	_sell_list = VBoxContainer.new()
	_sell_list.add_theme_constant_override("separation", 5)
	sell_scroll.add_child(_sell_list)
	_render()


func _render() -> void:
	if _panel == null:
		return
	_buy_buttons = {}
	_sell_buttons = {}
	_title_label.text = shop_title
	_gold_label.text = "Gold: %d" % gold
	_render_offer_list(_fixed_list, _offers_by_kind("fixed"))
	_render_offer_list(_generated_list, _offers_by_kind("generated"))
	_render_sell_list()


func _render_offer_list(list: VBoxContainer, rows: Array) -> void:
	if list == null:
		return
	_clear_children(list)
	if rows.is_empty():
		list.add_child(_empty_row("No offers"))
		return
	for offer in rows:
		if typeof(offer) != TYPE_DICTIONARY:
			continue
		var rec := offer as Dictionary
		var row := HBoxContainer.new()
		row.add_theme_constant_override("separation", 6)
		row.custom_minimum_size = Vector2(0, 34)
		var name := Label.new()
		name.text = _offer_name(rec)
		name.clip_text = true
		name.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		name.add_theme_color_override("font_color", _rarity_color(str(rec.get("rarity", "common"))))
		row.add_child(name)
		var price := Label.new()
		price.text = str(int(rec.get("buy_price", 0)))
		price.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
		price.custom_minimum_size = Vector2(48, 0)
		price.add_theme_color_override("font_color", Color("#f4c84f"))
		row.add_child(price)
		var btn := Button.new()
		btn.text = "Buy"
		btn.custom_minimum_size = Vector2(58, 30)
		btn.disabled = gold < int(rec.get("buy_price", 0))
		btn.tooltip_text = "Buy %s" % _offer_name(rec)
		btn.pressed.connect(func() -> void:
			_emit_buy(rec)
		)
		row.add_child(btn)
		_buy_buttons[str(rec.get("offer_id", ""))] = btn
		list.add_child(row)


func _render_sell_list() -> void:
	if _sell_list == null:
		return
	_clear_children(_sell_list)
	var sellable := _sellable_items()
	if sellable.is_empty():
		_sell_list.add_child(_empty_row("Bag is empty"))
		return
	for item in sellable:
		var row := HBoxContainer.new()
		row.add_theme_constant_override("separation", 6)
		row.custom_minimum_size = Vector2(0, 34)
		var name := Label.new()
		name.text = _item_name(item)
		name.clip_text = true
		name.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		name.add_theme_color_override("font_color", _rarity_color(str(item.get("rarity", "common"))))
		row.add_child(name)
		var btn := Button.new()
		btn.text = "Sell"
		btn.custom_minimum_size = Vector2(58, 30)
		btn.tooltip_text = "Sell %s" % _item_name(item)
		btn.pressed.connect(func() -> void:
			_emit_sell(item)
		)
		row.add_child(btn)
		_sell_buttons[str(item.get("item_instance_id", ""))] = btn
		_sell_list.add_child(row)


func _emit_buy(offer: Dictionary) -> void:
	if shop_entity_id == "" or offer.is_empty():
		return
	intent_requested.emit("shop_buy_intent", {
		"shop_entity_id": shop_entity_id,
		"offer_id": str(offer.get("offer_id", "")),
	})


func _emit_sell(item: Dictionary) -> void:
	if shop_entity_id == "" or item.is_empty():
		return
	intent_requested.emit("shop_sell_intent", {
		"shop_entity_id": shop_entity_id,
		"item_instance_id": str(item.get("item_instance_id", "")),
	})


func _sellable_items() -> Array:
	var out: Array = []
	for item in inventory:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := item as Dictionary
		var item_id := str(rec.get("item_instance_id", ""))
		if item_id == "" or _is_equipped_instance(item_id):
			continue
		out.append(rec)
	return out


func _is_equipped_instance(item_instance_id: String) -> bool:
	for slot in equipped.keys():
		var equipped_id = equipped.get(slot, null)
		if equipped_id != null and str(equipped_id) == item_instance_id:
			return true
	return false


func _offers_by_kind(kind: String) -> Array:
	var out: Array = []
	for offer in offers:
		if typeof(offer) == TYPE_DICTIONARY and str((offer as Dictionary).get("kind", "")) == kind:
			out.append(offer)
	return out


func _matching_offers(offer_id: String, offer_kind: String) -> Array:
	var out: Array = []
	for offer in offers:
		if typeof(offer) != TYPE_DICTIONARY:
			continue
		var rec := offer as Dictionary
		if offer_id != "" and str(rec.get("offer_id", "")) != offer_id:
			continue
		if offer_kind != "" and str(rec.get("kind", "")) != offer_kind:
			continue
		out.append(rec)
	out.sort_custom(func(a, b) -> bool:
		var ap := int((a as Dictionary).get("buy_price", 0))
		var bp := int((b as Dictionary).get("buy_price", 0))
		if ap == bp:
			return str((a as Dictionary).get("offer_id", "")) < str((b as Dictionary).get("offer_id", ""))
		return ap < bp
	)
	return out


func _debug_buy_buttons() -> Dictionary:
	var out := {}
	for offer_id in _buy_buttons.keys():
		var btn := _buy_buttons[offer_id] as Button
		out[str(offer_id)] = {
			"enabled": btn != null and not btn.disabled,
			"text": btn.text if btn != null else "",
		}
	return out


func _debug_sell_rows() -> Array:
	var out: Array = []
	for item in _sellable_items():
		out.append({
			"item_instance_id": str(item.get("item_instance_id", "")),
			"item_def_id": str(item.get("item_def_id", "")),
			"item_template_id": str(item.get("item_template_id", "")),
			"display_name": _item_name(item),
		})
	return out


func _offer_name(offer: Dictionary) -> String:
	var name := str(offer.get("display_name", ""))
	if name != "":
		return name
	return str(offer.get("item_template_id", offer.get("item_def_id", "item")))


func _item_name(item: Dictionary) -> String:
	var name := str(item.get("display_name", ""))
	if name != "":
		return name
	return str(item.get("item_template_id", item.get("item_def_id", "item")))


func _rarity_color(rarity: String) -> Color:
	match rarity.to_lower():
		"magic":
			return Color("#93c5fd")
		"rare":
			return Color("#f4d481")
		"unique":
			return Color("#ffb26b")
		_:
			return Color("#e8dcc8")


func _empty_row(text: String) -> Label:
	var label := Label.new()
	label.text = text
	label.add_theme_color_override("font_color", Color("#a99d8b"))
	return label


func _caption(text: String) -> Label:
	var label := Label.new()
	label.text = text
	label.add_theme_color_override("font_color", Color("#c9a227"))
	label.add_theme_font_size_override("font_size", 22)
	return label


func _subcaption(text: String) -> Label:
	var label := Label.new()
	label.text = text
	label.add_theme_color_override("font_color", Color("#9b8a6b"))
	label.add_theme_font_size_override("font_size", 16)
	return label


func _panel_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.07, 0.06, 0.05, 0.93)
	s.border_color = Color("#6b5420")
	s.border_width_left = 2
	s.border_width_top = 2
	s.border_width_right = 2
	s.border_width_bottom = 2
	s.content_margin_left = 14
	s.content_margin_top = 12
	s.content_margin_right = 14
	s.content_margin_bottom = 12
	return s


func _reposition_panel() -> void:
	if _panel == null:
		return
	var margin := 20.0
	var viewport_size := get_viewport_rect().size
	_panel.offset_right = -margin
	_panel.offset_left = _panel.offset_right - PANEL_SIZE.x
	_panel.offset_top = margin + 54.0
	_panel.offset_bottom = _panel.offset_top + PANEL_SIZE.y
	if viewport_size.x > 0.0 and viewport_size.x + _panel.offset_left < margin:
		_panel.offset_left = -viewport_size.x + margin
	if viewport_size.y > 0.0 and _panel.offset_bottom > viewport_size.y - margin:
		var overflow := _panel.offset_bottom - (viewport_size.y - margin)
		_panel.offset_top -= overflow
		_panel.offset_bottom -= overflow


func _apply_interaction_filters() -> void:
	if _panel == null:
		return
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP if _interactive and visible else Control.MOUSE_FILTER_IGNORE


func _clear_children(node: Node) -> void:
	for child in node.get_children():
		child.queue_free()


func _dup_array(rows: Array) -> Array:
	var out: Array = []
	for row in rows:
		if typeof(row) == TYPE_DICTIONARY:
			out.append((row as Dictionary).duplicate(true))
		else:
			out.append(row)
	return out
