class_name ShopPanel
extends Control

signal intent_requested(intent_type: String, payload: Dictionary)

const ItemTooltipPanelScript := preload("res://scripts/item_tooltip_panel.gd")
const ItemIconDrawerScript := preload("res://scripts/item_icon_drawer.gd")
const MysterySilhouetteDrawer := preload("res://scripts/mystery_silhouette_drawer.gd")
const StatLabels := preload("res://scripts/stat_labels.gd")
const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const InventoryRenderGuardScript := preload("res://scripts/inventory_render_guard.gd")
const PANEL_SIZE := Vector2(360, 680)
const VENDOR_COLUMNS := 5
const VENDOR_ROWS := 10
const VENDOR_SLOT_COUNT := VENDOR_COLUMNS * VENDOR_ROWS
const SLOT_SIZE := Vector2(52, 52)
const SLOT_GAP := 6
const TITLE_FONT_SIZE := 33
const BODY_FONT_SIZE := 23
const DETAIL_FONT_SIZE := 20
const TOOLTIP_META_FONT_SIZE := BODY_FONT_SIZE - 4
const ICON_FONT_SIZE := 20
const PRICE_FONT_SIZE := 13
const DRAG_SOURCE_SHOP_OFFER := "shop_offer"
const DRAG_SOURCE_INVENTORY_BAG := "bag"
const MYSTERY_SLOT_STYLE := "magic"
const ITEM_IDENTITY_FIELDS := [
	"item_def_id",
	"item_template_id",
	"display_name",
	"rarity",
	"rolled_stats",
	"requirements",
	"requirement_status",
	"comparison",
	"equip_preview",
	"summary_lines",
]
const ITEM_RARITY_BACKGROUNDS := {
	"common": Color("#343432"),
	"magic": Color("#1b3458"),
	"rare": Color("#5a4520"),
	"unique": Color("#5a2f17"),
}

var shop_id: String = ""
var shop_entity_id: String = ""
var shop_title: String = "Vendor"
var offers: Array = []
var inventory: Array = []
var equipped: Dictionary = {}
var sell_appraisals: Array = []
var gold: int = 0
var item_rules: Dictionary:
	get: return ItemRulesLoader.item_rules
var item_templates: Dictionary:
	get: return ItemRulesLoader.item_templates
var item_presentations: Dictionary:
	get: return ItemRulesLoader.item_presentations

var _panel: DraggableWindow
var _title_label: Label
var _reroll_button: Button
var _status_label: Label
var _vendor_grid: GridContainer
var _buy_buttons: Dictionary = {}
var _interactive: bool = true
var _suppress_render_guard: bool = false


class ShopSlotButton:
	extends Button

	var panel: ShopPanel
	var offer: Dictionary = {}

	func _draw() -> void:
		if offer.is_empty():
			return
		panel._draw_item_icon(self, offer)

	func _gui_input(event: InputEvent) -> void:
		if not panel._interactive:
			return
		if event is InputEventMouseButton \
				and event.button_index == MOUSE_BUTTON_LEFT \
				and event.double_click \
				and not offer.is_empty():
			panel._handle_offer_double_click(offer)

	func _get_drag_data(_at_position: Vector2) -> Variant:
		if not panel._interactive or offer.is_empty() or not panel._offer_affordable(offer):
			return null
		var data := {
			"source": DRAG_SOURCE_SHOP_OFFER,
			"shop_entity_id": panel.shop_entity_id,
			"offer_id": str(offer.get("offer_id", "")),
			"item": offer,
		}
		set_drag_preview(panel._drag_preview(offer))
		return data

	func _can_drop_data(_at_position: Vector2, data: Variant) -> bool:
		if not panel._interactive or typeof(data) != TYPE_DICTIONARY:
			return false
		var rec := data as Dictionary
		if str(rec.get("source", "")) != DRAG_SOURCE_INVENTORY_BAG:
			return false
		var item: Dictionary = rec.get("item", {})
		return panel._inventory_item_sellable(item)

	func _drop_data(_at_position: Vector2, data: Variant) -> void:
		panel._handle_inventory_drop(data)

	func _make_custom_tooltip(for_text: String) -> Object:
		if panel == null:
			return null
		if offer.is_empty():
			return panel._make_text_tooltip("Drop inventory items here to sell.")
		return panel._make_offer_tooltip(offer)


func _ready() -> void:
	ItemRulesLoader.ensure_loaded()
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_build()
	visible = false


func show_shop(next_shop_entity_id: String, next_shop_id: String, next_offers: Array, next_gold: int, next_inventory: Array, next_equipped: Dictionary, next_title: String = "Town Vendor", next_sell_appraisals: Array = []) -> void:
	_suppress_render_guard = true
	shop_entity_id = next_shop_entity_id
	shop_id = next_shop_id
	shop_title = next_title
	offers = _dup_array(next_offers)
	sell_appraisals = _dup_array(next_sell_appraisals)
	set_inventory_state(next_inventory, next_equipped, next_gold)
	_suppress_render_guard = false
	visible = true
	_apply_interaction_filters()
	_render_if_changed()


func apply_shop_refresh(next_offers: Array, next_sell_appraisals: Array) -> void:
	offers = _dup_array(next_offers)
	sell_appraisals = _dup_array(next_sell_appraisals)
	_apply_interaction_filters()
	_render_if_changed()


func hide_display() -> void:
	visible = false


func set_interactive(enabled: bool) -> void:
	_interactive = enabled
	_apply_interaction_filters()


func set_inventory_state(next_inventory: Array, next_equipped: Dictionary, next_gold: int) -> void:
	inventory = _dup_array(next_inventory)
	equipped = next_equipped.duplicate(true)
	gold = max(0, next_gold)
	if _panel != null and not _suppress_render_guard:
		_render_if_changed()


func _render_if_changed() -> void:
	if InventoryRenderGuardScript.should_render(self):
		_render_and_mark()


func _render_and_mark() -> void:
	_render()
	InventoryRenderGuardScript.mark_rendered(self)


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
		"mystery_offer_count": _offers_by_kind("mystery").size(),
		"buyback_offer_count": _offers_by_kind("buyback").size(),
		"reroll_visible": _reroll_button != null and _reroll_button.visible,
		"reroll_enabled": _reroll_enabled(),
		"reroll_cost": _reroll_cost(),
		"buy_buttons": _debug_buy_buttons(),
		"offer_rows": _debug_offer_rows(),
		"sell_rows": _debug_sell_rows(),
		"sell_row_count": _sell_rows().size(),
		"comparison_row_count": _comparison_row_count(),
		"requirement_row_count": _requirement_row_count(),
		"equip_preview_row_count": _equip_preview_row_count(),
		"vendor_columns": VENDOR_COLUMNS,
		"vendor_rows": VENDOR_ROWS,
		"vendor_slot_count": VENDOR_SLOT_COUNT,
		"occupied_vendor_slot_count": _vendor_items().size(),
		"header_gold_visible": false,
		"status": _status_label.text if _status_label != null else "",
		"window": _panel.get_debug_state() if _panel != null else {},
	}


func bot_click_buy_offer(offer_id: String = "", offer_kind: String = "", offer_index: int = 0) -> void:
	var matches := _matching_offers(offer_id, offer_kind)
	if offer_index < 0 or offer_index >= matches.size():
		return
	var selected: Dictionary = matches[offer_index]
	if not _offer_affordable(selected):
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


func bot_click_reroll() -> void:
	if _reroll_enabled():
		_emit_reroll()


func bot_click_close() -> void:
	if _panel != null and _panel.close_button() != null:
		_panel.close_button().pressed.emit()


func bot_drag_window_by(delta: Vector2) -> void:
	if _panel != null:
		_panel.bot_drag_by(delta)


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_reposition_panel()


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = DraggableWindowScript.new()
	_panel.custom_minimum_size = PANEL_SIZE
	_panel.configure(shop_title, Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 58))
	_panel.add_theme_stylebox_override("panel", _panel_style())
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	_panel.close_requested.connect(hide_display)
	add_child(_panel)
	_reposition_panel()
	_panel.set_layout_key("shop")

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 8)
	root.custom_minimum_size = Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 58)
	_panel.set_content(root)

	var header := HBoxContainer.new()
	header.add_theme_constant_override("separation", 10)
	root.add_child(header)
	_title_label = Label.new()
	_title_label.text = shop_title
	_title_label.visible = false
	_title_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_title_label.add_theme_color_override("font_color", Color("#f4d481"))
	_title_label.add_theme_font_size_override("font_size", TITLE_FONT_SIZE)

	_reroll_button = Button.new()
	_reroll_button.text = "Reroll"
	_reroll_button.focus_mode = Control.FOCUS_NONE
	_reroll_button.custom_minimum_size = Vector2(88, 36)
	_reroll_button.pressed.connect(_emit_reroll)
	header.add_child(_reroll_button)

	_status_label = Label.new()
	_status_label.text = ""
	_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_status_label.add_theme_color_override("font_color", Color("#b8aa91"))
	_status_label.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	root.add_child(_status_label)

	_vendor_grid = GridContainer.new()
	_vendor_grid.columns = VENDOR_COLUMNS
	_vendor_grid.add_theme_constant_override("h_separation", SLOT_GAP)
	_vendor_grid.add_theme_constant_override("v_separation", SLOT_GAP)
	_vendor_grid.custom_minimum_size = Vector2(
		VENDOR_COLUMNS * SLOT_SIZE.x + (VENDOR_COLUMNS - 1) * SLOT_GAP,
		VENDOR_ROWS * SLOT_SIZE.y + (VENDOR_ROWS - 1) * SLOT_GAP
	)
	root.add_child(_vendor_grid)
	_render()


func _render() -> void:
	if _panel == null or _vendor_grid == null:
		return
	_buy_buttons = {}
	_title_label.text = shop_title
	_panel.configure(shop_title, Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 58))
	_sync_reroll_button()
	_clear_children(_vendor_grid)
	var rows := _vendor_items()
	for i in range(VENDOR_SLOT_COUNT):
		var slot := _slot_button()
		var offer: Dictionary = rows[i] if i < rows.size() else {}
		_fill_slot(slot, offer)
		_vendor_grid.add_child(slot)


func _vendor_items() -> Array:
	var rows: Array = []
	for offer in offers:
		if typeof(offer) == TYPE_DICTIONARY:
			rows.append((offer as Dictionary).duplicate(true))
	return rows


func _slot_button() -> ShopSlotButton:
	var btn := ShopSlotButton.new()
	btn.panel = self
	btn.custom_minimum_size = SLOT_SIZE
	btn.focus_mode = Control.FOCUS_NONE
	btn.clip_text = true
	btn.add_theme_stylebox_override("normal", _slot_style(false))
	btn.add_theme_stylebox_override("hover", _slot_style(true))
	btn.add_theme_stylebox_override("pressed", _slot_style(true))
	btn.add_theme_color_override("font_color", Color("#e8dcc8"))
	btn.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	return btn


func _fill_slot(slot: ShopSlotButton, offer: Dictionary) -> void:
	slot.offer = offer.duplicate(true)
	if offer.is_empty():
		slot.text = ""
		slot.tooltip_text = "Drop item to sell"
		slot.add_theme_stylebox_override("normal", _slot_style(false))
		slot.add_theme_stylebox_override("hover", _slot_style(true))
		slot.add_theme_stylebox_override("pressed", _slot_style(true))
		slot.queue_redraw()
		return
	slot.text = ""
	slot.tooltip_text = _tooltip(offer)
	var rarity := MYSTERY_SLOT_STYLE if _is_mystery_offer(offer) else str(offer.get("rarity", "common"))
	var affordable := _offer_affordable(offer)
	slot.add_theme_stylebox_override("normal", _item_slot_style(rarity, false, affordable))
	slot.add_theme_stylebox_override("hover", _item_slot_style(rarity, true, affordable))
	slot.add_theme_stylebox_override("pressed", _item_slot_style(rarity, true, affordable))
	_buy_buttons[str(offer.get("offer_id", ""))] = slot
	slot.queue_redraw()


func _draw_item_icon(slot: Control, item: Dictionary) -> void:
	if _is_mystery_offer(item):
		_draw_mystery_icon(slot, item)
		return
	var def_id := str(item.get("item_def_id", ""))
	var icon: Dictionary = item_presentations.get(def_id, {}).get("icon", {})
	var rect := Rect2(Vector2.ZERO, slot.size)
	var label := str(icon.get("label", _short_label(def_id)))
	ItemIconDrawerScript.draw(slot, rect, icon, label, not _offer_affordable(item), 0.10, ICON_FONT_SIZE)

	var font := slot.get_theme_default_font()
	var price := str(int(item.get("buy_price", 0)))
	var price_color := Color("#f4c84f") if _offer_affordable(item) else Color("#ff6f6f")
	var price_size := font.get_string_size(price, HORIZONTAL_ALIGNMENT_LEFT, -1, PRICE_FONT_SIZE)
	slot.draw_string(
		font,
		Vector2(slot.size.x - price_size.x - 3.0, slot.size.y - 4.0),
		price,
		HORIZONTAL_ALIGNMENT_LEFT,
		-1,
		PRICE_FONT_SIZE,
		price_color
	)


func _draw_mystery_icon(slot: Control, item: Dictionary) -> void:
	var rect := Rect2(Vector2.ZERO, slot.size)
	var center := rect.get_center()
	var min_side = min(rect.size.x, rect.size.y)
	var color := Color("#b6a6ff")
	var accent := Color("#f4d481")
	if not _offer_affordable(item):
		color = color.darkened(0.35)
		accent = accent.darkened(0.35)
	MysterySilhouetteDrawer.draw(slot, center, min_side, item, color, accent)
	var font := slot.get_theme_default_font()
	var label := "?"
	var text_size := font.get_string_size(label, HORIZONTAL_ALIGNMENT_LEFT, -1, ICON_FONT_SIZE)
	slot.draw_string(font, center + Vector2(-text_size.x * 0.5, min_side * 0.11), label, HORIZONTAL_ALIGNMENT_LEFT, -1, ICON_FONT_SIZE, Color("#f4ead8"))
	_draw_offer_price(slot, item)


func _draw_offer_price(slot: Control, item: Dictionary) -> void:
	var font := slot.get_theme_default_font()
	var price := str(int(item.get("buy_price", 0)))
	var price_color := Color("#f4c84f") if _offer_affordable(item) else Color("#ff6f6f")
	var price_size := font.get_string_size(price, HORIZONTAL_ALIGNMENT_LEFT, -1, PRICE_FONT_SIZE)
	slot.draw_string(
		font,
		Vector2(slot.size.x - price_size.x - 3.0, slot.size.y - 4.0),
		price,
		HORIZONTAL_ALIGNMENT_LEFT,
		-1,
		PRICE_FONT_SIZE,
		price_color
	)


func _drag_preview(offer: Dictionary) -> Control:
	var label := Label.new()
	label.text = _offer_name(offer)
	var rarity := MYSTERY_SLOT_STYLE if _is_mystery_offer(offer) else str(offer.get("rarity", "common"))
	label.add_theme_color_override("font_color", _rarity_color(rarity))
	label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	return label


func _handle_offer_double_click(offer: Dictionary) -> void:
	if not _offer_affordable(offer):
		show_status("insufficient gold", true)
		return
	_emit_buy(offer)


func _handle_inventory_drop(data: Variant) -> void:
	if typeof(data) != TYPE_DICTIONARY:
		return
	var item: Dictionary = (data as Dictionary).get("item", {})
	if not _inventory_item_sellable(item):
		return
	_emit_sell(item)


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


func _emit_reroll() -> void:
	if shop_entity_id == "" or not _reroll_enabled():
		return
	intent_requested.emit("shop_reroll_intent", {
		"shop_entity_id": shop_entity_id,
	})


func _sync_reroll_button() -> void:
	if _reroll_button == null:
		return
	var cost := _reroll_cost()
	_reroll_button.visible = _is_mystery_shop()
	_reroll_button.disabled = not _reroll_enabled()
	_reroll_button.text = "Reroll %d" % cost
	_reroll_button.tooltip_text = "Refresh mystery offers for %d gold" % cost


func _reroll_enabled() -> bool:
	return _interactive and _is_mystery_shop() and gold >= _reroll_cost()


func _reroll_cost() -> int:
	return int(ItemRulesLoader.shop_rules.get("shops", {}).get(shop_id, {}).get("mystery_offers", {}).get("reroll_cost", 0))


func _is_mystery_shop() -> bool:
	return shop_id == "town_mystery_seller" and _reroll_cost() > 0


func _offer_affordable(offer: Dictionary) -> bool:
	return gold >= int(offer.get("buy_price", 0))


func _inventory_item_sellable(item: Dictionary) -> bool:
	var item_id := str(item.get("item_instance_id", ""))
	if item_id == "" or _is_equipped_instance(item_id):
		return false
	if sell_appraisals.is_empty():
		return true
	for row in _sell_rows():
		if str((row as Dictionary).get("item_instance_id", "")) == item_id:
			return true
	return false


func _sellable_items() -> Array:
	var out: Array = []
	for item in inventory:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := item as Dictionary
		if _inventory_item_sellable(rec):
			out.append(rec)
	return out


func _sell_rows() -> Array:
	if sell_appraisals.is_empty():
		var items: Array = []
		for item in inventory:
			if typeof(item) == TYPE_DICTIONARY:
				var rec := item as Dictionary
				var item_id := str(rec.get("item_instance_id", ""))
				if item_id != "" and not _is_equipped_instance(item_id):
					items.append(rec)
		return items
	var live_ids := {}
	for item in inventory:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := item as Dictionary
		var item_id := str(rec.get("item_instance_id", ""))
		if item_id != "" and not _is_equipped_instance(item_id):
			live_ids[item_id] = true
	var out: Array = []
	for appraisal in sell_appraisals:
		if typeof(appraisal) != TYPE_DICTIONARY:
			continue
		var rec := appraisal as Dictionary
		if live_ids.get(str(rec.get("item_instance_id", "")), false):
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
	for offer in offers:
		if typeof(offer) != TYPE_DICTIONARY:
			continue
		var rec := offer as Dictionary
		var offer_id := str(rec.get("offer_id", ""))
		out[offer_id] = {
			"enabled": _offer_affordable(rec),
			"text": "Buy",
		}
	return out


func _debug_sell_rows() -> Array:
	var out: Array = []
	for item in _sell_rows():
		out.append({
			"item_instance_id": str(item.get("item_instance_id", "")),
			"item_def_id": str(item.get("item_def_id", "")),
			"item_template_id": str(item.get("item_template_id", "")),
			"display_name": _item_name(item),
			"rarity": str(item.get("rarity", "")),
			"slot": str(item.get("slot", "")),
			"category": str(item.get("category", "")),
			"sell_price": int(item.get("sell_price", 0)),
			"summary_lines": _detail_lines(item),
			"comparison_count": _comparison_count(item),
			"requirement_count": _requirement_lines(item).size(),
			"equip_preview_count": _equip_preview_count(item),
		})
	return out


func _debug_offer_rows() -> Array:
	var out: Array = []
	for offer in offers:
		if typeof(offer) != TYPE_DICTIONARY:
			continue
		var rec := offer as Dictionary
		var mystery := _is_mystery_offer(rec)
		out.append({
			"offer_id": str(rec.get("offer_id", "")),
			"kind": str(rec.get("kind", "")),
			"item_def_id": "" if mystery else str(rec.get("item_def_id", "")),
			"item_template_id": "" if mystery else str(rec.get("item_template_id", "")),
			"display_name": _offer_name(rec),
			"rarity": "" if mystery else str(rec.get("rarity", "")),
			"slot": str(rec.get("slot", "")),
			"category": str(rec.get("category", "")),
			"buy_price": int(rec.get("buy_price", 0)),
			"source_depth": int(rec.get("source_depth", 0)),
			"source_depth_min": _source_depth_min(rec),
			"source_depth_max": _source_depth_max(rec),
			"concealed": mystery,
			"mystery_label": _mystery_label(rec) if mystery else "",
			"mystery_silhouette": MysterySilhouetteDrawer.key(rec) if mystery else "",
			"identity_field_count": _identity_field_count(rec) if mystery else 0,
			"summary_lines": _detail_lines(rec),
			"comparison_count": _comparison_count(rec),
			"requirement_count": _requirement_lines(rec).size(),
			"equip_preview_count": _equip_preview_count(rec),
		})
	return out


func _comparison_row_count() -> int:
	var total := 0
	for offer in offers:
		if typeof(offer) == TYPE_DICTIONARY:
			total += _comparison_count(offer as Dictionary)
	for item in _sell_rows():
		total += _comparison_count(item)
	return total


func _requirement_row_count() -> int:
	var total := 0
	for offer in offers:
		if typeof(offer) == TYPE_DICTIONARY:
			total += _requirement_lines(offer as Dictionary).size()
	for item in _sell_rows():
		total += _requirement_lines(item).size()
	return total


func _equip_preview_row_count() -> int:
	var total := 0
	for offer in offers:
		if typeof(offer) == TYPE_DICTIONARY:
			total += _equip_preview_count(offer as Dictionary)
	for item in _sell_rows():
		total += _equip_preview_count(item)
	return total


func _offer_name(offer: Dictionary) -> String:
	if _is_mystery_offer(offer):
		return _mystery_label(offer)
	var name := str(offer.get("display_name", ""))
	if name != "":
		return name
	var def_id := str(offer.get("item_def_id", ""))
	var def := _item_definition(def_id)
	return str(def.get("name", offer.get("item_template_id", def_id if def_id != "" else "item")))


func _item_name(item: Dictionary) -> String:
	var name := str(item.get("display_name", ""))
	if name != "":
		return name
	var def_id := str(item.get("item_def_id", ""))
	var def := _item_definition(def_id)
	return str(def.get("name", item.get("item_template_id", def_id if def_id != "" else "item")))


func _tooltip(row: Dictionary) -> String:
	if _is_mystery_offer(row):
		return "\n".join(_mystery_tooltip_lines(row))
	var lines: Array = [_offer_name(row)]
	var rarity := str(row.get("rarity", ""))
	if rarity != "":
		lines.append("Rarity: %s" % rarity.capitalize())
	lines.append_array(_detail_lines(row, true, true))
	return "\n".join(lines)


func _make_offer_tooltip(offer: Dictionary) -> Control:
	var tooltip := ItemTooltipPanelScript.new()
	if _is_mystery_offer(offer):
		tooltip.setup(
			{},
			item_presentations,
			_mystery_tooltip_lines(offer),
			[],
			[],
			int(offer.get("buy_price", 0)),
			_offer_affordable(offer),
			"?"
		)
		return tooltip
	tooltip.setup(
		offer,
		item_presentations,
		_tooltip_lines(offer),
		_requirement_lines(offer),
		_comparison_entries(offer),
		int(offer.get("buy_price", 0)),
		_offer_affordable(offer),
		_short_label(str(offer.get("item_def_id", "")))
	)
	return tooltip


func _make_text_tooltip(text: String) -> Control:
	var tooltip := ItemTooltipPanelScript.new()
	tooltip.setup({}, item_presentations, [text], [], [], -1, true, "")
	return tooltip


func _tooltip_lines(row: Dictionary) -> Array:
	if _is_mystery_offer(row):
		return _mystery_tooltip_lines(row)
	var rarity := str(row.get("rarity", ""))
	var lines: Array = [_item_name_tooltip_line(_offer_name(row), rarity)]
	if rarity != "":
		lines.append(_metadata_tooltip_line("Rarity: %s" % rarity.capitalize()))
	lines.append_array(_compact_metadata_lines(_detail_lines(row, false, false)))
	return lines


func _item_name_tooltip_line(text: String, rarity: String) -> Dictionary:
	return {"text": text, "color": _rarity_color(rarity)}


func _metadata_tooltip_line(text: String) -> Dictionary:
	return {"text": text, "color": Color("#cdbd9f"), "font_size": TOOLTIP_META_FONT_SIZE}


func _compact_metadata_lines(lines: Array) -> Array:
	var out: Array = []
	for line in lines:
		var text := str(line)
		if text.begins_with("Slot:"):
			out.append(_metadata_tooltip_line(text))
		else:
			out.append(line)
	return out


func _detail_lines(row: Dictionary, include_requirements: bool = true, include_comparison: bool = true) -> Array:
	if _is_mystery_offer(row):
		return _mystery_detail_lines(row)
	var lines: Array = []
	var summary = row.get("summary_lines", [])
	if typeof(summary) == TYPE_ARRAY:
		for line in summary:
			var text := str(line)
			if text != "":
				if not include_requirements and _is_requirement_summary_line(text):
					continue
				if not include_comparison and _is_comparison_summary_line(text):
					continue
				lines.append(text)
	if lines.is_empty():
		var slot := str(row.get("slot", ""))
		if slot != "":
			lines.append("Slot: %s" % slot)
		else:
			var category := str(row.get("category", ""))
			if category != "":
				lines.append("Kind: %s" % category)
		lines.append_array(_stat_lines(row.get("rolled_stats", {})))
		var req = row.get("requirements", {})
		if include_requirements and typeof(req) == TYPE_DICTIONARY and int((req as Dictionary).get("level", 0)) > 0:
			lines.append("Requires level %d" % int((req as Dictionary).get("level", 0)))
	if include_comparison:
		lines.append_array(_comparison_lines(row.get("comparison", {})))
	return lines


func _requirement_lines(row: Dictionary) -> Array:
	if _is_mystery_offer(row):
		return []
	var lines: Array = []
	var statuses = row.get("requirement_status", [])
	if typeof(statuses) == TYPE_ARRAY:
		for status in statuses:
			if typeof(status) != TYPE_DICTIONARY:
				continue
			var status_rec := status as Dictionary
			var stat := str(status_rec.get("stat", ""))
			var required := int(status_rec.get("required", 0))
			if stat == "" or required <= 0:
				continue
			var current := int(status_rec.get("current", 0))
			var met := bool(status_rec.get("met", current >= required))
			var suffix := "" if met else "(%d)" % (current - required)
			lines.append({
				"text": "%s %d%s" % [_display_stat(stat), required, suffix],
				"color": _requirement_color(met),
			})
	if not lines.is_empty():
		return lines
	var req = row.get("requirements", {})
	if typeof(req) == TYPE_DICTIONARY:
		var rec := req as Dictionary
		if int(rec.get("level", 0)) > 0:
			lines.append("Level %d" % int(rec.get("level", 0)))
		for key in rec.keys():
			var stat := str(key)
			if stat == "level":
				continue
			lines.append("%s %s" % [_display_stat(stat), str(rec.get(key, ""))])
	var summary = row.get("summary_lines", [])
	if typeof(summary) == TYPE_ARRAY:
		for line in summary:
			var parsed := _requirement_from_summary_line(str(line))
			if parsed != "" and not lines.has(parsed):
				lines.append(parsed)
	return lines


func _is_requirement_summary_line(text: String) -> bool:
	return _requirement_from_summary_line(text) != ""


func _is_comparison_summary_line(text: String) -> bool:
	return _comparison_delta_from_line(text) != null


func _requirement_from_summary_line(text: String) -> String:
	var normalized := text.strip_edges()
	if not normalized.to_lower().begins_with("requires "):
		return ""
	var rest := normalized.substr("Requires ".length()).strip_edges()
	if rest.to_lower().begins_with("level "):
		return "Level %s" % rest.substr("level ".length()).strip_edges()
	return rest.capitalize()


func _comparison_entries(row: Dictionary) -> Array:
	if _is_mystery_offer(row):
		return []
	var entries: Array = []
	_append_equip_preview_entries(entries, row)
	var comparison = row.get("comparison", {})
	if typeof(comparison) == TYPE_DICTIONARY:
		var deltas = (comparison as Dictionary).get("deltas", [])
		if typeof(deltas) == TYPE_ARRAY:
			for delta in deltas:
				if typeof(delta) != TYPE_DICTIONARY:
					continue
				var rec := delta as Dictionary
				var diff := int(rec.get("delta", 0))
				var sign := "+" if diff >= 0 else ""
				entries.append({
					"text": "%s%s %s vs equipped" % [sign, str(diff), _display_stat(str(rec.get("stat", "")))],
					"color": _comparison_color(diff),
				})
	var summary = row.get("summary_lines", [])
	if typeof(summary) == TYPE_ARRAY:
		for line in summary:
			var text := str(line)
			var diff_value = _comparison_delta_from_line(text)
			if diff_value == null:
				continue
			var duplicate := false
			for entry in entries:
				if typeof(entry) == TYPE_DICTIONARY and str((entry as Dictionary).get("text", "")) == text:
					duplicate = true
					break
			if duplicate:
				continue
			entries.append({
				"text": text,
				"color": _comparison_color(int(diff_value)),
			})
	return entries


func _append_equip_preview_entries(entries: Array, row: Dictionary) -> void:
	var preview = row.get("equip_preview", {})
	if typeof(preview) != TYPE_DICTIONARY:
		return
	var deltas = (preview as Dictionary).get("deltas", [])
	if typeof(deltas) != TYPE_ARRAY:
		return
	for delta in deltas:
		if typeof(delta) != TYPE_DICTIONARY:
			continue
		var rec := delta as Dictionary
		var diff := int(rec.get("delta", 0))
		var sign := "+" if diff >= 0 else ""
		entries.append({
			"text": "%s%s %s preview" % [sign, str(diff), _display_stat(str(rec.get("stat", "")))],
			"color": _comparison_color(diff),
		})


func _comparison_delta_from_line(text: String):
	var stripped := text.strip_edges()
	if not stripped.contains("vs equipped"):
		return null
	if stripped.length() == 0 or (not stripped.begins_with("+") and not stripped.begins_with("-")):
		return null
	var first_space := stripped.find(" ")
	if first_space <= 1:
		return null
	return int(stripped.substr(0, first_space))


func _comparison_color(delta: int) -> Color:
	if delta > 0:
		return Color("#9ee6a8")
	if delta < 0:
		return Color("#ff9f7a")
	return Color("#d8c7a6")


func _stat_lines(stats_value: Variant) -> Array:
	if typeof(stats_value) != TYPE_DICTIONARY:
		return []
	var stats := stats_value as Dictionary
	var lines: Array = []
	if int(stats.get("damage_min", 0)) > 0 or int(stats.get("damage_max", 0)) > 0:
		lines.append("Damage %d-%d" % [int(stats.get("damage_min", 0)), int(stats.get("damage_max", 0))])
	for key in ["str", "dex", "vit", "magic", "all_skills", "armor", "block_percent", "attack_speed_percent", "hit_chance", "crit_chance", "evade_chance", "max_hp", "max_mana", "health_regen_per_10_seconds", "mana_regen_per_10_seconds", "skill_damage_percent", "skill_cooldown_reduction_percent", "skill_mana_cost_reduction", "hotbar_slots", "inventory_rows"]:
		var value := int(stats.get(key, 0))
		if value > 0:
			lines.append("%s %s" % [_display_stat(key), _format_stat_value(key, value)])
	return lines


func _format_stat_value(stat: String, value: int) -> String:
	if stat == "block_percent" or stat == "attack_speed_percent" or stat == "hit_chance" or stat == "crit_chance" or stat == "evade_chance" or stat == "skill_damage_percent" or stat == "skill_cooldown_reduction_percent":
		return "+%d%%" % value
	return "+%d" % value


func _comparison_lines(comparison_value: Variant) -> Array:
	if typeof(comparison_value) != TYPE_DICTIONARY:
		return []
	var comparison := comparison_value as Dictionary
	var deltas = comparison.get("deltas", [])
	if typeof(deltas) != TYPE_ARRAY:
		return []
	var lines: Array = []
	for delta in deltas:
		if typeof(delta) != TYPE_DICTIONARY:
			continue
		var rec := delta as Dictionary
		var diff := int(rec.get("delta", 0))
		var sign := "+" if diff >= 0 else ""
		lines.append("%s%s %s vs equipped" % [sign, str(diff), _display_stat(str(rec.get("stat", "")))])
	return lines


func _comparison_count(row: Dictionary) -> int:
	if _is_mystery_offer(row):
		return 0
	var comparison = row.get("comparison", {})
	if typeof(comparison) != TYPE_DICTIONARY:
		return 0
	var deltas = (comparison as Dictionary).get("deltas", [])
	if typeof(deltas) != TYPE_ARRAY:
		return 0
	return (deltas as Array).size()


func _requirement_color(met: bool) -> Color:
	return Color("#9ee6a8") if met else Color("#ff6f6f")


func _equip_preview_count(row: Dictionary) -> int:
	if _is_mystery_offer(row):
		return 0
	var preview = row.get("equip_preview", {})
	if typeof(preview) != TYPE_DICTIONARY:
		return 0
	var deltas = (preview as Dictionary).get("deltas", [])
	if typeof(deltas) != TYPE_ARRAY:
		return 0
	return (deltas as Array).size()


func _is_mystery_offer(row: Dictionary) -> bool:
	return str(row.get("kind", "")) == "mystery" \
		or bool(row.get("concealed", false)) \
		or str(row.get("offer_id", "")).begins_with("mystery:")


func _mystery_label(row: Dictionary) -> String:
	var label := str(row.get("mystery_label", ""))
	if label != "":
		return label
	var slot := _display_token(str(row.get("slot", "")))
	if slot != "":
		return "Unidentified %s" % slot
	return "Unidentified item"


func _mystery_tooltip_lines(row: Dictionary) -> Array:
	var lines: Array = [_mystery_label(row)]
	lines.append_array(_mystery_detail_lines(row))
	return lines


func _mystery_detail_lines(row: Dictionary) -> Array:
	var lines: Array = []
	var slot := _display_token(str(row.get("slot", "")))
	if slot != "":
		lines.append("Slot: %s" % slot)
	var category := _display_token(str(row.get("category", "")))
	if category != "":
		lines.append("Kind: %s" % category)
	var silhouette := MysterySilhouetteDrawer.label(row)
	if silhouette != "":
		lines.append("Silhouette: %s" % silhouette)
	var source := _source_window_line(row)
	if source != "":
		lines.append(source)
	return lines


func _source_window_line(row: Dictionary) -> String:
	var min_depth := _source_depth_min(row)
	var max_depth := _source_depth_max(row)
	if min_depth <= 0 and max_depth <= 0:
		return ""
	if max_depth <= min_depth:
		return "Source depth: %d" % min_depth
	return "Source depths: %d-%d" % [min_depth, max_depth]


func _source_depth_min(row: Dictionary) -> int:
	if row.has("source_depth_min"):
		return int(row.get("source_depth_min", 0))
	return int(row.get("source_depth", 0))


func _source_depth_max(row: Dictionary) -> int:
	if row.has("source_depth_max"):
		return int(row.get("source_depth_max", 0))
	return int(row.get("source_depth", 0))


func _identity_field_count(row: Dictionary) -> int:
	var count := 0
	for key in ITEM_IDENTITY_FIELDS:
		if not row.has(key):
			continue
		var value = row.get(key)
		if key == "display_name" and str(value) == _mystery_label(row):
			continue
		if _has_identity_value(value):
			count += 1
	return count


func _has_identity_value(value) -> bool:
	match typeof(value):
		TYPE_NIL:
			return false
		TYPE_STRING:
			return str(value) != ""
		TYPE_ARRAY:
			return not (value as Array).is_empty()
		TYPE_DICTIONARY:
			return not (value as Dictionary).is_empty()
		_:
			return true


func _display_token(value: String) -> String:
	return value.replace("_", " ").capitalize()


func _display_stat(stat: String) -> String:
	return StatLabels.display_name(stat)


func _rarity_color(rarity: String) -> Color:
	match rarity.to_lower():
		"magic":
			return Color("#93c5fd")
		"rare":
			return Color("#f4d481")
		"unique":
			return Color("#ffb26b")
		"set":
			return Color("#55e66f")
		_:
			return Color("#e8dcc8")


func _short_label(def_id: String) -> String:
	var def: Dictionary = _item_definition(def_id)
	var name := str(def.get("name", def_id))
	var parts := name.split(" ")
	var out := ""
	for part in parts:
		if part.length() > 0:
			out += part.substr(0, 1).to_upper()
	return out.substr(0, 3)


func _item_definition(def_id: String) -> Dictionary:
	if item_rules.has(def_id):
		return item_rules.get(def_id, {})
	return item_templates.get(def_id, {})




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


func _slot_style(hover: bool) -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color("#3d2e10") if hover else Color("#0a0908")
	s.border_color = Color("#8b6914") if hover else Color("#5c4a1f")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.content_margin_left = 4
	s.content_margin_top = 4
	s.content_margin_right = 4
	s.content_margin_bottom = 4
	return s


func _item_slot_style(rarity: String, hover: bool, affordable: bool) -> StyleBoxFlat:
	var s := _slot_style(hover)
	var base: Color = ITEM_RARITY_BACKGROUNDS.get(rarity.to_lower(), ITEM_RARITY_BACKGROUNDS["common"])
	if not affordable:
		base = base.darkened(0.45)
	s.bg_color = base.lightened(0.12) if hover else base
	s.border_color = base.lightened(0.46) if hover else base.lightened(0.28)
	return s


func _tooltip_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.07, 0.06, 0.05, 0.97)
	s.border_color = Color("#8b6914")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.content_margin_left = 10
	s.content_margin_top = 8
	s.content_margin_right = 10
	s.content_margin_bottom = 8
	return s


func _reposition_panel() -> void:
	if _panel == null:
		return
	var margin := 20.0
	var panel_size := _panel.custom_minimum_size
	var viewport_size := get_viewport_rect().size if is_inside_tree() else Vector2(1280, 720)
	var desired := Vector2(margin, margin + 54.0)
	_panel.position = Vector2(
		clampf(desired.x, margin, maxf(margin, viewport_size.x - panel_size.x - margin)),
		clampf(desired.y, margin, maxf(margin, viewport_size.y - panel_size.y - margin))
	)


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
