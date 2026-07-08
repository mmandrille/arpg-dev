class_name MaterialWalletPanel
extends Control

const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const ItemIconDrawerScript := preload("res://scripts/item_icon_drawer.gd")
const ItemTooltipPanelScript := preload("res://scripts/item_tooltip_panel.gd")
const InventoryTransferRouterScript := preload("res://scripts/inventory_transfer_router.gd")

const PANEL_SIZE := Vector2(360, 280)
const BODY_FONT_SIZE := 14
const DETAIL_FONT_SIZE := 12
const WALLET_COLUMNS := 5
const BASE_WALLET_ROWS := 3
const SLOT_SIZE := Vector2(56, 56)
const SLOT_GAP := 6
const DRAG_SOURCE_RESOURCE_BAG := "resource_bag"

signal intent_requested(intent_type: String, payload: Dictionary)

var resource_wallet: Dictionary = {}
var resource_bag_items: Array = []
var wallet_rows: int = BASE_WALLET_ROWS
var _panel: DraggableWindow
var _scroll: ScrollContainer
var _bag_grid: GridContainer
var _badge_rows: VBoxContainer
var _empty_label: Label
var _row_debug: Array = []
var _interactive: bool = true


class WalletSlotButton:
	extends Button

	var panel: MaterialWalletPanel
	var slot_kind: String = ""
	var item: Dictionary = {}

	func _draw() -> void:
		if item.is_empty():
			return
		panel._draw_item_icon(self, item)

	func _get_drag_data(_at_position: Vector2) -> Variant:
		if not panel._interactive or item.is_empty() or slot_kind != "bag":
			return null
		var data := {
			"source": DRAG_SOURCE_RESOURCE_BAG,
			"bag_item_id": InventoryTransferRouterScript.resource_bag_item_id_from_item(item),
			"item": item.duplicate(true),
		}
		panel._set_drag_preview(self, data)
		return data

	func _gui_input(event: InputEvent) -> void:
		if not panel._interactive:
			return
		if event is InputEventMouseButton \
				and event.button_index == MOUSE_BUTTON_LEFT \
				and event.pressed \
				and event.double_click \
				and slot_kind == "bag" \
				and not item.is_empty():
			panel._emit_withdraw(item)
			accept_event()

	func _can_drop_data(_at_position: Vector2, data: Variant) -> bool:
		return panel._interactive and slot_kind == "bag" and panel._can_accept_drop(data)

	func _drop_data(_at_position: Vector2, data: Variant) -> void:
		panel._handle_drop_on_bag(data)

	func _make_custom_tooltip(for_text: String) -> Object:
		if panel == null:
			return null
		if item.is_empty():
			return panel._make_text_tooltip(for_text)
		return panel._make_item_tooltip(item)


func _ready() -> void:
	ItemRulesLoader.ensure_loaded()
	set_anchors_preset(Control.PRESET_FULL_RECT)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	visible = false


func show_wallet(next_wallet: Dictionary, next_bag_items: Array = []) -> void:
	resource_wallet = next_wallet.duplicate(true)
	resource_bag_items = _dup_array(next_bag_items)
	_ensure_wallet_rows()
	if _is_empty_display():
		hide_display()
		return
	_build()
	visible = true
	_render()


func set_wallet(next_wallet: Dictionary, next_bag_items: Array = []) -> void:
	resource_wallet = next_wallet.duplicate(true)
	resource_bag_items = _dup_array(next_bag_items)
	_ensure_wallet_rows()
	if not is_open():
		return
	if _is_empty_display():
		hide_display()
		return
	_render()


func hide_display() -> void:
	visible = false


func is_open() -> bool:
	return visible and _panel != null


func get_debug_state() -> Dictionary:
	return {
		"visible": is_open(),
		"wallet_rows": wallet_rows,
		"wallet_capacity": wallet_capacity(),
		"bag_item_count": resource_bag_items.size(),
		"row_count": _row_debug.size(),
		"rows": _row_debug.duplicate(true),
		"text": _debug_text(),
		"window": _panel.get_debug_state() if _panel != null else {},
	}


func wallet_capacity() -> int:
	return wallet_rows * WALLET_COLUMNS


func _build() -> void:
	if _panel != null:
		return
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = DraggableWindowScript.new()
	_panel.configure("Resources", PANEL_SIZE)
	_panel.custom_minimum_size = Vector2(PANEL_SIZE.x, PANEL_SIZE.y + DraggableWindowScript.TITLEBAR_HEIGHT)
	_panel.size = _panel.custom_minimum_size
	_panel.position = Vector2(414, 348)
	_panel.set_layout_key("material_wallet")
	_panel.close_requested.connect(hide_display)
	add_child(_panel)

	var root := VBoxContainer.new()
	root.custom_minimum_size = PANEL_SIZE
	root.add_theme_constant_override("separation", 8)
	_panel.set_content(root)

	var title := Label.new()
	title.text = "Account resources"
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.add_theme_font_size_override("font_size", 16)
	title.add_theme_color_override("font_color", Color("#f0dfbb"))
	root.add_child(title)

	_scroll = ScrollContainer.new()
	_scroll.custom_minimum_size = Vector2(PANEL_SIZE.x - 16, 168)
	_scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	root.add_child(_scroll)

	_bag_grid = GridContainer.new()
	_bag_grid.columns = WALLET_COLUMNS
	_bag_grid.add_theme_constant_override("h_separation", SLOT_GAP)
	_bag_grid.add_theme_constant_override("v_separation", SLOT_GAP)
	_scroll.add_child(_bag_grid)

	_badge_rows = VBoxContainer.new()
	_badge_rows.add_theme_constant_override("separation", 6)
	root.add_child(_badge_rows)

	_empty_label = Label.new()
	_empty_label.text = "No materials stored"
	_empty_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_empty_label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	_empty_label.add_theme_color_override("font_color", Color("#8f826b"))
	root.add_child(_empty_label)


func _render() -> void:
	if _bag_grid == null:
		return
	_ensure_wallet_rows()
	_clear_children(_bag_grid)
	_row_debug.clear()
	var slots := wallet_capacity()
	for i in range(slots):
		var item: Dictionary = {}
		if i < resource_bag_items.size():
			item = resource_bag_items[i] as Dictionary
		var slot := _slot_button("bag")
		_fill_slot(slot, item)
		_bag_grid.add_child(slot)
		if not item.is_empty():
			_row_debug.append(_debug_item_row(item))

	_clear_children(_badge_rows)
	for resource_id in _wallet_resource_keys():
		var amount := int(resource_wallet.get(resource_id, 0))
		_badge_rows.add_child(_badge_row(str(resource_id), amount))

	if _empty_label != null:
		_empty_label.visible = resource_bag_items.is_empty() and _wallet_resource_keys().is_empty()


func _ensure_wallet_rows() -> void:
	wallet_rows = max(BASE_WALLET_ROWS, wallet_rows)
	while resource_bag_items.size() > wallet_rows * WALLET_COLUMNS:
		wallet_rows += 1


func _is_empty_display() -> bool:
	return resource_bag_items.is_empty() and _wallet_resource_keys().is_empty()


func _slot_button(kind: String) -> WalletSlotButton:
	var slot := WalletSlotButton.new()
	slot.panel = self
	slot.slot_kind = kind
	slot.custom_minimum_size = SLOT_SIZE
	slot.focus_mode = Control.FOCUS_NONE
	slot.mouse_default_cursor_shape = Control.CURSOR_POINTING_HAND
	slot.add_theme_stylebox_override("normal", _slot_style(false))
	slot.add_theme_stylebox_override("hover", _slot_style(true))
	slot.add_theme_stylebox_override("pressed", _slot_style(true))
	return slot


func _fill_slot(slot: WalletSlotButton, item: Dictionary) -> void:
	slot.item = item.duplicate(true) if not item.is_empty() else {}
	if item.is_empty():
		slot.tooltip_text = ""
	else:
		slot.tooltip_text = "\n".join(_tooltip_text_lines(_tooltip_lines(item)))
	slot.queue_redraw()


func _draw_item_icon(slot: Control, item: Dictionary) -> void:
	var def_id := str(item.get("item_def_id", ""))
	var icon: Dictionary = ItemRulesLoader.item_presentations.get(def_id, {}).get("icon", {})
	var rect := Rect2(Vector2.ZERO, slot.size)
	var label := str(icon.get("label", _short_label(def_id)))
	ItemIconDrawerScript.draw(slot, rect, icon, label, false, 0.18, DETAIL_FONT_SIZE)


func _short_label(def_id: String) -> String:
	var parts := def_id.replace("_", " ").split(" ")
	var out := ""
	for part in parts:
		if part.length() > 0:
			out += part.substr(0, 1).to_upper()
	return out.substr(0, 3)


func _set_drag_preview(slot: Control, data: Dictionary) -> void:
	var preview := PanelContainer.new()
	preview.custom_minimum_size = SLOT_SIZE
	preview.add_theme_stylebox_override("panel", _slot_style(true))
	var icon_host := Control.new()
	icon_host.custom_minimum_size = SLOT_SIZE
	preview.add_child(icon_host)
	var item: Dictionary = data.get("item", {})
	if not item.is_empty():
		ItemIconDrawerScript.draw(icon_host, Rect2(Vector2.ZERO, SLOT_SIZE), ItemRulesLoader.item_presentations.get(str(item.get("item_def_id", "")), {}).get("icon", {}), _short_label(str(item.get("item_def_id", ""))), false, 0.18, DETAIL_FONT_SIZE)
	set_drag_preview(preview)


func _can_accept_drop(data: Variant) -> bool:
	if typeof(data) != TYPE_DICTIONARY:
		return false
	var source := str((data as Dictionary).get("source", ""))
	return source == "bag" or source == InventoryTransferRouterScript.DRAG_SOURCE_STASH


func _handle_drop_on_bag(data: Variant) -> void:
	if typeof(data) != TYPE_DICTIONARY:
		return
	var payload: Dictionary = data as Dictionary
	var source := str(payload.get("source", ""))
	if source == "bag":
		var item: Dictionary = payload.get("item", {})
		var item_id := str(item.get("item_instance_id", ""))
		if item_id == "":
			return
		intent_requested.emit("resource_bag_deposit_item_intent", {"item_instance_id": item_id})
		return
	if source == InventoryTransferRouterScript.DRAG_SOURCE_STASH:
		var stash_entity_id := str(payload.get("stash_entity_id", ""))
		var stash_item_id := str(payload.get("stash_item_id", ""))
		if stash_entity_id == "" or stash_item_id == "":
			return
		intent_requested.emit("resource_bag_deposit_stash_item_intent", {
			"stash_entity_id": stash_entity_id,
			"stash_item_id": stash_item_id,
		})


func _emit_withdraw(item: Dictionary) -> void:
	var bag_item_id := InventoryTransferRouterScript.resource_bag_item_id_from_item(item)
	if bag_item_id == "":
		return
	intent_requested.emit("resource_bag_withdraw_item_intent", {"bag_item_id": bag_item_id})


func _badge_row(resource_id: String, amount: int) -> Control:
	var row := PanelContainer.new()
	row.add_theme_stylebox_override("panel", _row_style())
	var box := VBoxContainer.new()
	row.add_child(box)
	var name := _resource_name(resource_id)
	var header := Label.new()
	header.text = "%s x%d" % [name, amount]
	header.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	header.add_theme_color_override("font_color", Color("#b8e6ff"))
	box.add_child(header)
	box.add_child(_detail_label("Stored account-wide"))
	return row


func _detail_label(text: String) -> Label:
	var label := Label.new()
	label.text = text
	label.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	label.add_theme_color_override("font_color", Color("#b7ad98"))
	return label


func _wallet_resource_keys() -> Array:
	var out: Array = []
	var keys: Array = resource_wallet.keys()
	keys.sort()
	for key in keys:
		if str(key) == "upgrade_shard":
			continue
		var amount := int(resource_wallet.get(key, 0))
		if amount > 0:
			out.append(key)
	return out


func _resource_name(resource_id: String) -> String:
	var def := ItemRulesLoader.item_definition(resource_id)
	if def.has("name"):
		return str(def.get("name", ""))
	return resource_id.replace("_", " ").capitalize()


func _make_item_tooltip(item: Dictionary) -> Control:
	var tooltip := ItemTooltipPanelScript.new()
	tooltip.setup(
		item,
		ItemRulesLoader.item_presentations,
		_tooltip_lines(item),
		[],
		[],
		-1,
		true,
		_short_label(str(item.get("item_def_id", ""))),
		[]
	)
	return tooltip


func _make_text_tooltip(text: String) -> Control:
	var tooltip := ItemTooltipPanelScript.new()
	tooltip.setup({}, ItemRulesLoader.item_presentations, [text], [], [], -1, true, "")
	return tooltip


func _tooltip_lines(item: Dictionary) -> Array:
	var def_id := str(item.get("item_def_id", ""))
	var def := ItemRulesLoader.item_definition(def_id)
	var name := str(item.get("display_name", def.get("name", _resource_name(def_id))))
	var lines: Array = [{"text": name, "color": Color("#f0dfbb")}]
	var category := str(def.get("category", ""))
	if category != "":
		lines.append({"text": "Kind: %s" % category.capitalize(), "color": Color("#cdbd9f"), "font_size": DETAIL_FONT_SIZE})
	if def.has("description"):
		lines.append(str(def.get("description", "")))
	return lines


func _tooltip_text_lines(lines: Array) -> Array:
	var out: Array = []
	for line in lines:
		if typeof(line) == TYPE_DICTIONARY:
			out.append(str((line as Dictionary).get("text", "")))
		else:
			out.append(str(line))
	return out


func _debug_item_row(item: Dictionary) -> Dictionary:
	var bag_id := str(item.get("stash_item_id", item.get("bag_item_id", "")))
	var def_id := str(item.get("item_def_id", ""))
	return {
		"bag_item_id": bag_id,
		"item_def_id": def_id,
		"text": "%s (%s)" % [_resource_name(def_id), bag_id],
	}


func _debug_text() -> String:
	var lines: Array = []
	for row in _row_debug:
		lines.append(str((row as Dictionary).get("text", "")))
	for key in _wallet_resource_keys():
		lines.append("%s x%d" % [_resource_name(str(key)), int(resource_wallet.get(key, 0))])
	return "\n".join(lines)


func _dup_array(values: Array) -> Array:
	var out: Array = []
	for value in values:
		if typeof(value) == TYPE_DICTIONARY:
			out.append((value as Dictionary).duplicate(true))
	return out


func _clear_children(node: Node) -> void:
	if node == null:
		return
	for child in node.get_children():
		node.remove_child(child)
		child.queue_free()


func _row_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.065, 0.06, 0.052, 0.94)
	s.border_color = Color("#3c3324")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.corner_radius_top_left = 5
	s.corner_radius_top_right = 5
	s.corner_radius_bottom_left = 5
	s.corner_radius_bottom_right = 5
	s.content_margin_left = 8
	s.content_margin_right = 8
	s.content_margin_top = 6
	s.content_margin_bottom = 6
	return s


func _slot_style(hover: bool) -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.08, 0.075, 0.065, 0.96 if hover else 0.9)
	s.border_color = Color("#6a5a3d" if hover else "#3c3324")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.corner_radius_top_left = 4
	s.corner_radius_top_right = 4
	s.corner_radius_bottom_left = 4
	s.corner_radius_bottom_right = 4
	return s
