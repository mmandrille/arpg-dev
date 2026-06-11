class_name InventoryPanel
extends Control

signal intent_requested(intent_type: String, payload: Dictionary)

const ItemTooltipPanelScript := preload("res://scripts/item_tooltip_panel.gd")
const StatLabels := preload("res://scripts/stat_labels.gd")
const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const SLOT_KIND_BAG := "bag"
const SLOT_KIND_EQUIP_PREFIX := "equip:"
const DRAG_SOURCE_SHOP_OFFER := "shop_offer"
const DRAG_SOURCE_STASH := "stash"
const BAG_COLUMNS := 5
const BASE_INVENTORY_ROWS := 3
const HOTKEY_LABELS := ["1", "2", "3", "4", "5", "6", "7", "8", "9", "0"]
const TITLE_FONT_SIZE := 33
const BODY_FONT_SIZE := 23
const SLOT_FONT_SIZE := 23
const ICON_FONT_SIZE := 22
const EQUIPMENT_SLOT_SIZE := Vector2(96, 58)
const EQUIPMENT_SLOTS := ["head", "amulet", "chest", "gloves", "belt", "boots", "ring_left", "ring_right", "main_hand", "off_hand"]
const EQUIPMENT_LABELS := {
	"head": "Head",
	"amulet": "Amulet",
	"chest": "Chest",
	"gloves": "Gloves",
	"belt": "Belt",
	"boots": "Boots",
	"ring_left": "Ring L",
	"ring_right": "Ring R",
	"main_hand": "Main",
	"off_hand": "Off"
}
const TOOLTIP_STAT_SEPARATOR := "----------------"
const ITEM_RARITY_BACKGROUNDS := {
	"common": Color("#343432"),
	"magic": Color("#1b3458"),
	"rare": Color("#5a4520"),
	"unique": Color("#5a2f17"),
}
const PAPER_DOLL_SLOT_POSITIONS := {
	"head": Vector2(122, 10),
	"amulet": Vector2(208, 40),
	"main_hand": Vector2(20, 116),
	"off_hand": Vector2(236, 116),
	"chest": Vector2(122, 112),
	"ring_left": Vector2(20, 198),
	"ring_right": Vector2(236, 198),
	"gloves": Vector2(36, 276),
	"belt": Vector2(122, 194),
	"boots": Vector2(122, 276),
}

var inventory: Array = []
var equipped: Dictionary = {}
var hotbar: Array = []
var inventory_rows: int = BASE_INVENTORY_ROWS
var inventory_capacity: int = BASE_INVENTORY_ROWS * BAG_COLUMNS
var gold: int = 0
var item_rules: Dictionary:
	get: return ItemRulesLoader.item_rules
var item_templates: Dictionary:
	get: return ItemRulesLoader.item_templates
var item_presentations: Dictionary:
	get: return ItemRulesLoader.item_presentations
var _panel: DraggableWindow
var _equipment_slots: Dictionary = {}
var _bag_grid: GridContainer
var _gold_label: Label
var _paper_doll_preview: Control
var _drag_data: Dictionary = {}
var _interactive: bool = true
var _gesture_hint: Label
var _gesture_tween: Tween
var _rendered_bag_slot_count: int = 0
var _shop_sell_entity_id: String = ""


class InventorySlotButton:
	extends Button

	var panel: InventoryPanel
	var slot_kind: String = ""
	var item: Dictionary = {}

	func _draw() -> void:
		if item.is_empty():
			return
		panel._draw_item_icon(self, item)

	func _gui_input(event: InputEvent) -> void:
		if not panel._interactive:
			return
		if event is InputEventMouseButton \
				and event.button_index == MOUSE_BUTTON_LEFT \
				and event.double_click \
				and slot_kind == SLOT_KIND_BAG \
				and not item.is_empty():
			panel._handle_double_click(item)

	func _get_drag_data(_at_position: Vector2) -> Variant:
		if not panel._interactive or item.is_empty():
			return null
		panel._drag_data = {"source": slot_kind, "item": item}
		var preview := Label.new()
		preview.text = str(item.get("item_def_id", "item"))
		preview.add_theme_color_override("font_color", Color("#e8dcc8"))
		set_drag_preview(preview)
		return panel._drag_data

	func _can_drop_data(_at_position: Vector2, data: Variant) -> bool:
		if typeof(data) != TYPE_DICTIONARY:
			return false
		var source := str(data.get("source", ""))
		var dragged: Dictionary = data.get("item", {})
		if dragged.is_empty():
			return false
		if panel._slot_kind_is_equipment(slot_kind):
			return source == SLOT_KIND_BAG and panel._item_can_equip_to(dragged, panel._slot_from_kind(slot_kind))
		if slot_kind == SLOT_KIND_BAG:
			if source == DRAG_SOURCE_SHOP_OFFER and str(data.get("offer_id", "")) != "" and str(data.get("shop_entity_id", "")) != "":
				return true
			if source == DRAG_SOURCE_STASH \
					and str(data.get("stash_entity_id", "")) != "" \
					and str(data.get("stash_item_id", "")) != "":
				return true
			return panel._slot_kind_is_equipment(source)
		return false

	func _drop_data(_at_position: Vector2, data: Variant) -> void:
		panel._handle_drop_on_slot(slot_kind, data)

	func _make_custom_tooltip(for_text: String) -> Object:
		if panel == null:
			return null
		if item.is_empty():
			return panel._make_text_tooltip(for_text)
		return panel._make_item_tooltip(item)


func _ready() -> void:
	ItemRulesLoader.ensure_loaded()
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_ensure_built()
	_render()
	visible = false


func toggle() -> void:
	if not _interactive:
		return
	visible = not visible
	_apply_interaction_filters()


func set_interactive(enabled: bool) -> void:
	_interactive = enabled
	_apply_interaction_filters()


func set_shop_sell_context(shop_entity_id: String) -> void:
	_shop_sell_entity_id = shop_entity_id


func clear_shop_sell_context() -> void:
	_shop_sell_entity_id = ""


func ensure_display_visible() -> void:
	visible = true
	_apply_interaction_filters()


func hide_display() -> void:
	visible = false


func bot_click_close() -> void:
	if _panel != null and _panel.close_button() != null:
		_panel.close_button().pressed.emit()


func bot_drag_window_by(delta: Vector2) -> void:
	if _panel != null:
		_panel.bot_drag_by(delta)


func show_gesture_hint(text: String) -> void:
	if _gesture_hint == null:
		return
	if _gesture_tween != null and is_instance_valid(_gesture_tween):
		_gesture_tween.kill()
	_gesture_hint.text = text
	_gesture_hint.modulate.a = 1.0
	_gesture_hint.visible = true
	_gesture_tween = create_tween()
	_gesture_tween.tween_interval(0.9)
	_gesture_tween.tween_property(_gesture_hint, "modulate:a", 0.0, 0.35)
	_gesture_tween.tween_callback(func() -> void:
		if _gesture_hint != null:
			_gesture_hint.visible = false)


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_reposition_panel()


func set_inventory_state(next_inventory: Array, next_equipped: Dictionary, next_inventory_rows: int = BASE_INVENTORY_ROWS, next_inventory_capacity: int = BASE_INVENTORY_ROWS * BAG_COLUMNS, next_gold: int = 0, next_hotbar: Array = []) -> void:
	inventory = []
	for item in next_inventory:
		inventory.append((item as Dictionary).duplicate(true))
	equipped = next_equipped.duplicate(true)
	hotbar = []
	for slot in next_hotbar:
		hotbar.append((slot as Dictionary).duplicate(true))
	inventory_rows = max(0, next_inventory_rows)
	inventory_capacity = max(0, next_inventory_capacity)
	gold = max(0, next_gold)
	_ensure_built()
	_rendered_bag_slot_count = _target_bag_slot_count()
	if _bag_grid != null:
		_render()


func get_debug_state() -> Dictionary:
	_ensure_built()
	return {
		"visible": visible,
		"bag_count": inventory.size(),
		"equipped": equipped.duplicate(true),
		"equipped_main_hand": equipped.get("main_hand", null),
		"main_hand_item": _equipped_item("main_hand"),
		"weapon_item": _equipped_item("main_hand"),
		"item_presentations": _debug_presentations(),
		"inventory_rows": inventory_rows,
		"inventory_capacity": inventory_capacity,
		"gold": gold,
		"hotbar_assigned_item_ids": _debug_hotbar_assigned_item_ids(),
		"hotbar_assigned_inventory_count": _hotbar_assigned_inventory_count(),
		"bag_columns": _bag_grid.columns if _bag_grid != null else BAG_COLUMNS,
		"available_slot_count": inventory_capacity,
		"rendered_slot_count": _rendered_bag_slot_count,
		"paper_doll_slot_ids": EQUIPMENT_SLOTS.duplicate(),
		"paper_doll_slots": _debug_paper_doll_slots(),
		"paper_doll_preview": {
			"exists": _paper_doll_preview != null,
			"name": _paper_doll_preview.name if _paper_doll_preview != null else "",
			"visible": _paper_doll_preview.visible if _paper_doll_preview != null else false,
		},
		"requirement_row_count": _requirement_row_count(),
		"equip_preview_row_count": _equip_preview_row_count(),
		"empty_slot_style": "gray_block",
		"window": _panel.get_debug_state() if _panel != null else {},
	}


func get_weapon_slot_screen_center() -> Vector2:
	return get_equipment_slot_screen_center("main_hand")


func get_equipment_slot_screen_center(slot: String) -> Vector2:
	return _slot_screen_center(_equipment_slots.get(slot, null))


func get_bag_area_screen_center() -> Vector2:
	if _bag_grid == null:
		return Vector2.ZERO
	for child in _bag_grid.get_children():
		if child is InventorySlotButton and child.slot_kind == SLOT_KIND_BAG and child.item.is_empty():
			return _slot_screen_center(child)
	for child in _bag_grid.get_children():
		if child is InventorySlotButton and child.slot_kind == SLOT_KIND_BAG:
			return _slot_screen_center(child)
	return Vector2.ZERO


func get_bag_item_screen_center(item_instance_id: String = "") -> Vector2:
	if _bag_grid == null:
		return Vector2.ZERO
	for child in _bag_grid.get_children():
		if child is InventorySlotButton and child.slot_kind == SLOT_KIND_BAG:
			if item_instance_id == "" or str(child.item.get("item_instance_id", "")) == item_instance_id:
				return _slot_screen_center(child)
	if item_instance_id != "":
		var bag_index := 0
		for item in inventory:
			if _is_equipped_instance(str(item.get("item_instance_id", ""))):
				if str(item.get("item_instance_id", "")) == item_instance_id:
					return _bag_cell_screen_center(bag_index)
				continue
			if str(item.get("item_instance_id", "")) == item_instance_id:
				return _bag_cell_screen_center(bag_index)
			bag_index += 1
	return Vector2.ZERO


func get_drop_outside_screen_point() -> Vector2:
	if _panel == null or not is_inside_tree():
		return Vector2(120, get_viewport_rect().size.y - 80)
	var panel_rect := _panel.get_global_rect()
	return Vector2(panel_rect.position.x - 48, panel_rect.position.y + panel_rect.size.y * 0.5)


func _slot_screen_center(slot: Control) -> Vector2:
	if slot == null or not is_inside_tree():
		return Vector2.ZERO
	return slot.get_global_rect().get_center()


func _bag_cell_screen_center(cell_index: int) -> Vector2:
	if _bag_grid == null or not is_inside_tree():
		return Vector2.ZERO
	var cell_w := 48.0
	var cell_h := 48.0
	var sep := float(_bag_grid.get_theme_constant("h_separation"))
	if sep <= 0.0:
		sep = 6.0
	var grid_rect := _bag_grid.get_global_rect()
	var col := cell_index % _bag_grid.columns
	var row := int(cell_index / _bag_grid.columns)
	return grid_rect.position + Vector2(
		col * (cell_w + sep) + cell_w * 0.5,
		row * (cell_h + sep) + cell_h * 0.5,
	)


func _notification(what: int) -> void:
	if what == NOTIFICATION_DRAG_END and _interactive and not _drag_data.is_empty():
		if not get_viewport().gui_is_drag_successful():
			var item: Dictionary = _drag_data.get("item", {})
			if not item.is_empty():
				intent_requested.emit("drop_intent", {"item_instance_id": str(item.get("item_instance_id", ""))})
		_drag_data = {}


func _build() -> void:
	if _panel != null:
		return
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = DraggableWindowScript.new()
	_panel.custom_minimum_size = Vector2(750, 460)
	_panel.configure("Inventory", Vector2(720, 396))
	_reposition_panel()
	_panel.set_layout_key("inventory")
	_panel.add_theme_stylebox_override("panel", _panel_style())
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	_panel.close_requested.connect(hide_display)
	add_child(_panel)

	var body := VBoxContainer.new()
	body.add_theme_constant_override("separation", 2)
	body.custom_minimum_size = Vector2(720, 396)
	_panel.set_content(body)

	_gesture_hint = Label.new()
	_gesture_hint.visible = false
	_gesture_hint.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_gesture_hint.add_theme_color_override("font_color", Color("#c9a227"))
	_gesture_hint.add_theme_font_size_override("font_size", 23)
	body.add_child(_gesture_hint)

	var root := HBoxContainer.new()
	root.add_theme_constant_override("separation", 18)
	root.custom_minimum_size = Vector2(720, 370)
	body.add_child(root)

	var left := VBoxContainer.new()
	left.custom_minimum_size = Vector2(350, 0)
	root.add_child(left)
	left.add_child(_title("Equipment"))
	var paper := Control.new()
	paper.custom_minimum_size = Vector2(340, 360)
	left.add_child(paper)
	_paper_doll_preview = Panel.new()
	_paper_doll_preview.name = "character_paper_doll"
	_paper_doll_preview.position = Vector2(124, 78)
	_paper_doll_preview.custom_minimum_size = Vector2(92, 210)
	_paper_doll_preview.size = _paper_doll_preview.custom_minimum_size
	_paper_doll_preview.add_theme_stylebox_override("panel", _paper_doll_style())
	paper.add_child(_paper_doll_preview)
	for slot in EQUIPMENT_SLOTS:
		var btn := _slot_button(_slot_kind_for_equipment(str(slot)), EQUIPMENT_SLOT_SIZE)
		btn.position = PAPER_DOLL_SLOT_POSITIONS.get(str(slot), Vector2.ZERO)
		btn.size = btn.custom_minimum_size
		_equipment_slots[str(slot)] = btn
		paper.add_child(btn)

	var right := VBoxContainer.new()
	right.custom_minimum_size = Vector2(350, 0)
	root.add_child(right)
	right.add_child(_caption("Bag"))
	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(340, 382)
	right.add_child(scroll)
	_bag_grid = GridContainer.new()
	_bag_grid.columns = BAG_COLUMNS
	_bag_grid.add_theme_constant_override("h_separation", 6)
	_bag_grid.add_theme_constant_override("v_separation", 6)
	scroll.add_child(_bag_grid)
	_gold_label = Label.new()
	_gold_label.text = "Gold: 0"
	_gold_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	_gold_label.add_theme_color_override("font_color", Color("#f4c84f"))
	_gold_label.add_theme_font_size_override("font_size", 26)
	right.add_child(_gold_label)
	_render()


func _ensure_built() -> void:
	if _panel != null:
		return
	_build()


func _render() -> void:
	if _bag_grid == null:
		return
	for slot in EQUIPMENT_SLOTS:
		_fill_slot(_equipment_slots.get(slot, null), _equipped_item(str(slot)))
	for child in _bag_grid.get_children():
		child.queue_free()
	var bag_items := _bag_items()
	_rendered_bag_slot_count = _target_bag_slot_count()
	for i in range(_rendered_bag_slot_count):
		var slot := _slot_button(SLOT_KIND_BAG, Vector2(58, 58))
		var item: Dictionary = bag_items[i] if i < bag_items.size() else {}
		_fill_slot(slot, item)
		_bag_grid.add_child(slot)
	if _gold_label != null:
		_gold_label.text = "Gold: %d" % gold
	_position_gesture_hint()


func _bag_items() -> Array:
	var items: Array = []
	for item in inventory:
		if _is_equipped_instance(str(item.get("item_instance_id", ""))):
			continue
		items.append(item)
	return items


func _target_bag_slot_count() -> int:
	var bag_item_count := int(_bag_items().size())
	var required_slots: int = inventory_capacity if inventory_capacity > bag_item_count else bag_item_count
	if required_slots <= 0:
		return 0
	return int(ceil(float(required_slots) / float(BAG_COLUMNS))) * BAG_COLUMNS


func _apply_interaction_filters() -> void:
	if _panel == null:
		return
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP if _interactive else Control.MOUSE_FILTER_IGNORE


func _position_gesture_hint() -> void:
	if _gesture_hint == null or _panel == null:
		return
	_gesture_hint.set_anchors_preset(Control.PRESET_TOP_WIDE)
	_gesture_hint.offset_top = 4
	_gesture_hint.offset_bottom = 22


func _fill_slot(slot: InventorySlotButton, item: Dictionary) -> void:
	if slot == null:
		return
	slot.item = item.duplicate(true)
	if item.is_empty():
		if _slot_kind_is_equipment(slot.slot_kind):
			var slot_name := _slot_from_kind(slot.slot_kind)
			slot.text = str(EQUIPMENT_LABELS.get(slot_name, slot_name))
			slot.tooltip_text = "Empty %s" % str(EQUIPMENT_LABELS.get(slot_name, slot_name))
			slot.add_theme_stylebox_override("normal", _empty_slot_style(false))
			slot.add_theme_stylebox_override("hover", _empty_slot_style(true))
			slot.add_theme_stylebox_override("pressed", _empty_slot_style(true))
		else:
			slot.text = ""
			slot.tooltip_text = "Empty"
			slot.add_theme_stylebox_override("normal", _slot_style(false))
			slot.add_theme_stylebox_override("hover", _slot_style(true))
			slot.add_theme_stylebox_override("pressed", _slot_style(true))
		slot.queue_redraw()
		return
	slot.text = ""
	slot.tooltip_text = _tooltip(item)
	var rarity := str(item.get("rarity", "common"))
	slot.add_theme_stylebox_override("normal", _item_slot_style(rarity, false))
	slot.add_theme_stylebox_override("hover", _item_slot_style(rarity, true))
	slot.add_theme_stylebox_override("pressed", _item_slot_style(rarity, true))
	slot.queue_redraw()


func _slot_button(kind: String, size: Vector2) -> InventorySlotButton:
	var btn := InventorySlotButton.new()
	btn.panel = self
	btn.slot_kind = kind
	btn.custom_minimum_size = size
	btn.focus_mode = Control.FOCUS_NONE
	btn.clip_text = true
	btn.add_theme_stylebox_override("normal", _slot_style(false))
	btn.add_theme_stylebox_override("hover", _slot_style(true))
	btn.add_theme_stylebox_override("pressed", _slot_style(true))
	btn.add_theme_color_override("font_color", Color("#e8dcc8"))
	btn.add_theme_font_size_override("font_size", SLOT_FONT_SIZE)
	return btn


func _draw_item_icon(slot: Control, item: Dictionary) -> void:
	var def_id := str(item.get("item_def_id", ""))
	var icon: Dictionary = item_presentations.get(def_id, {}).get("icon", {})
	var shape := str(icon.get("shape", "box"))
	var color := Color(str(icon.get("color", "#d8d0bd")))
	var accent := Color(str(icon.get("accent", "#6b5420")))
	var rect := Rect2(Vector2.ZERO, slot.size)
	var center := rect.get_center()
	var min_side = min(rect.size.x, rect.size.y)
	var label := str(icon.get("label", _short_label(def_id)))

	match shape:
		"blade":
			var a := center + Vector2(-min_side * 0.22, min_side * 0.22)
			var b := center + Vector2(min_side * 0.22, -min_side * 0.22)
			slot.draw_line(a, b, color, 5.0, true)
			slot.draw_line(a + Vector2(-4, 4), a + Vector2(5, -5), accent, 4.0, true)
		"bow":
			slot.draw_arc(center, min_side * 0.28, -1.25, 1.25, 18, color, 4.0, true)
			slot.draw_line(center + Vector2(min_side * 0.18, -min_side * 0.26), center + Vector2(min_side * 0.18, min_side * 0.26), accent, 2.0, true)
		"badge", "coin":
			slot.draw_circle(center, min_side * 0.24, color)
			slot.draw_arc(center, min_side * 0.17, 0.0, TAU, 18, accent, 2.0, true)
		"leaf":
			var pts := PackedVector2Array([
				center + Vector2(0, -min_side * 0.30),
				center + Vector2(min_side * 0.24, -min_side * 0.02),
				center + Vector2(0, min_side * 0.28),
				center + Vector2(-min_side * 0.24, -min_side * 0.02),
			])
			slot.draw_colored_polygon(pts, color)
			slot.draw_line(center + Vector2(0, -min_side * 0.22), center + Vector2(0, min_side * 0.22), accent, 2.0, true)
		"potion":
			slot.draw_rect(Rect2(center + Vector2(-min_side * 0.13, -min_side * 0.05), Vector2(min_side * 0.26, min_side * 0.28)), color, true)
			slot.draw_rect(Rect2(center + Vector2(-min_side * 0.08, -min_side * 0.22), Vector2(min_side * 0.16, min_side * 0.16)), accent, true)
		_:
			slot.draw_rect(Rect2(center - Vector2(min_side * 0.20, min_side * 0.20), Vector2(min_side * 0.40, min_side * 0.40)), color, true)

	var font := slot.get_theme_default_font()
	var font_size := ICON_FONT_SIZE
	var text_size := font.get_string_size(label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size)
	slot.draw_string(font, center + Vector2(-text_size.x * 0.5, min_side * 0.38), label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, Color("#f4ead8"))
	_draw_hotbar_badge(slot, item)


func _draw_hotbar_badge(slot: Control, item: Dictionary) -> void:
	var assigned_slots := _hotbar_slots_for_item(str(item.get("item_instance_id", "")))
	if assigned_slots.is_empty():
		return
	var label := "H%s" % _hotbar_label_for_slot(int(assigned_slots[0]))
	if assigned_slots.size() > 1:
		label = "H+"
	var badge_rect := Rect2(Vector2(slot.size.x - 28.0, 3.0), Vector2(24.0, 16.0))
	slot.draw_rect(badge_rect, Color("#15110a"), true)
	slot.draw_rect(badge_rect, Color("#d6a94d"), false, 1.0)
	var font := slot.get_theme_default_font()
	var font_size := 10
	var text_size := font.get_string_size(label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size)
	slot.draw_string(font, badge_rect.position + Vector2((badge_rect.size.x - text_size.x) * 0.5, 12.0), label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, Color("#f4ead8"))


func _title(text: String) -> Label:
	var label := _caption(text)
	label.add_theme_font_size_override("font_size", TITLE_FONT_SIZE)
	return label


func _caption(text: String) -> Label:
	var label := Label.new()
	label.text = text
	label.add_theme_color_override("font_color", Color("#c9a227"))
	label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	return label


func _make_item_tooltip(item: Dictionary) -> Control:
	var tooltip := ItemTooltipPanelScript.new()
	tooltip.setup(
		item,
		item_presentations,
		_tooltip_lines(item),
		_requirement_lines(item),
		_comparison_entries(item),
		_item_gold_value(item),
		true,
		_short_label(str(item.get("item_def_id", "")))
	)
	return tooltip


func _make_text_tooltip(text: String) -> Control:
	var tooltip := ItemTooltipPanelScript.new()
	tooltip.setup({}, item_presentations, [text], [], [], -1, true, "")
	return tooltip


func _item_gold_value(item: Dictionary) -> int:
	for key in ["sell_price", "buy_price", "gold_value", "value"]:
		if item.has(key):
			return max(0, int(item.get(key, 0)))
	var buy_price := _item_buy_price(item)
	if buy_price > 0:
		return max(1, int(floor(float(buy_price) * _town_vendor_sell_multiplier())))
	return -1


func _item_buy_price(item: Dictionary) -> int:
	var def_id := str(item.get("item_def_id", ""))
	var shop := _town_vendor_rules()
	if def_id == "" or shop.is_empty():
		return 0
	if _is_generated_item(item):
		return _generated_buy_price(item, shop)
	var fixed_offers = shop.get("fixed_offers", [])
	if typeof(fixed_offers) == TYPE_ARRAY:
		for offer in fixed_offers:
			if typeof(offer) == TYPE_DICTIONARY and str((offer as Dictionary).get("item_def_id", "")) == def_id:
				return int((offer as Dictionary).get("buy_price", 0))
	return 0


func _generated_buy_price(item: Dictionary, shop: Dictionary) -> int:
	var template_id := str(item.get("item_template_id", item.get("item_def_id", "")))
	var template: Dictionary = item_templates.get(template_id, {})
	if template.is_empty():
		return 0
	var pricing: Dictionary = shop.get("pricing", {})
	var round_to := int(pricing.get("round_buy_to", 0))
	if round_to <= 0:
		return 0
	var rarity := str(item.get("rarity", "common"))
	var rarity_multipliers: Dictionary = pricing.get("rarity_multipliers", {})
	var multiplier := float(rarity_multipliers.get(rarity, 0.0))
	if multiplier <= 0.0:
		return 0
	var base_stats: Dictionary = template.get("base_stats", {})
	var final_stats: Dictionary = item.get("rolled_stats", {})
	var stat_weights: Dictionary = pricing.get("stat_weights", {})
	var slot_base: Dictionary = pricing.get("slot_base", {})
	var score := int(slot_base.get(str(template.get("slot", "")), 0))
	for stat in stat_weights.keys():
		var key := str(stat)
		var weight := int(stat_weights.get(key, 0))
		score += int(base_stats.get(key, 0)) * weight
	for stat in stat_weights.keys():
		var key := str(stat)
		var weight := int(stat_weights.get(key, 0))
		var delta := int(final_stats.get(key, 0)) - int(base_stats.get(key, 0))
		if delta > 0:
			score += delta * weight
	var raw = max(1.0, float(score) * multiplier)
	return int(ceil(raw / float(round_to))) * round_to


func _is_generated_item(item: Dictionary) -> bool:
	if str(item.get("item_template_id", "")) != "":
		return true
	var def_id := str(item.get("item_def_id", ""))
	return item_templates.has(def_id) and str(item.get("rarity", "")) != ""


func _town_vendor_rules() -> Dictionary:
	var shops: Dictionary = ItemRulesLoader.shop_rules.get("shops", {})
	return shops.get("town_vendor", {})


func _town_vendor_sell_multiplier() -> float:
	var shop := _town_vendor_rules()
	var pricing: Dictionary = shop.get("pricing", {})
	return float(pricing.get("sell_multiplier", 0.25))


func _reposition_panel() -> void:
	if _panel == null:
		return
	var margin := 20.0
	var panel_size := _panel.custom_minimum_size
	var viewport_size := get_viewport_rect().size if is_inside_tree() else Vector2(1280, 720)
	var bottom_margin := maxf(margin, minf(140.0, viewport_size.y * 0.16))
	_panel.position = Vector2(
		maxf(margin, viewport_size.x - panel_size.x - margin),
		maxf(margin, viewport_size.y - panel_size.y - bottom_margin)
	)


func _panel_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.07, 0.06, 0.05, 0.92)
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


func _item_slot_style(rarity: String, hover: bool) -> StyleBoxFlat:
	var s := _slot_style(hover)
	var base: Color = ITEM_RARITY_BACKGROUNDS.get(rarity.to_lower(), ITEM_RARITY_BACKGROUNDS["common"])
	s.bg_color = base.lightened(0.12) if hover else base
	s.border_color = base.lightened(0.46) if hover else base.lightened(0.28)
	return s


func _empty_slot_style(hover: bool) -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color("#3a3a37") if hover else Color("#242422")
	s.border_color = Color("#8a877d") if hover else Color("#5f5b52")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.content_margin_left = 4
	s.content_margin_top = 4
	s.content_margin_right = 4
	s.content_margin_bottom = 4
	return s


func _paper_doll_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color("#171715")
	s.border_color = Color("#5f5b52")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.corner_radius_top_left = 8
	s.corner_radius_top_right = 8
	s.corner_radius_bottom_right = 8
	s.corner_radius_bottom_left = 8
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


func _handle_double_click(item: Dictionary) -> void:
	if _shop_sell_entity_id != "" and not _is_equipped_instance(str(item.get("item_instance_id", ""))):
		intent_requested.emit("shop_sell_intent", {
			"shop_entity_id": _shop_sell_entity_id,
			"item_instance_id": str(item.get("item_instance_id", "")),
		})
		return
	var slot := _preferred_equip_slot(item)
	if slot != "":
		intent_requested.emit("equip_intent", {"item_instance_id": str(item.get("item_instance_id", "")), "slot": slot})
	elif _is_consumable(item):
		intent_requested.emit("use_intent", {"item_instance_id": str(item.get("item_instance_id", ""))})


func _handle_drop_on_slot(slot_kind: String, data: Variant) -> void:
	if typeof(data) != TYPE_DICTIONARY:
		return
	var item: Dictionary = data.get("item", {})
	if item.is_empty():
		return
	if _slot_kind_is_equipment(slot_kind):
		var slot := _slot_from_kind(slot_kind)
		if _item_can_equip_to(item, slot):
			intent_requested.emit("equip_intent", {"item_instance_id": str(item.get("item_instance_id", "")), "slot": slot})
	elif slot_kind == SLOT_KIND_BAG:
		var source := str(data.get("source", ""))
		if source == DRAG_SOURCE_SHOP_OFFER:
			intent_requested.emit("shop_buy_intent", {
				"shop_entity_id": str(data.get("shop_entity_id", "")),
				"offer_id": str(data.get("offer_id", "")),
			})
		elif source == DRAG_SOURCE_STASH:
			intent_requested.emit("stash_withdraw_item_intent", {
				"stash_entity_id": str(data.get("stash_entity_id", "")),
				"stash_item_id": str(data.get("stash_item_id", "")),
			})
		elif _slot_kind_is_equipment(source):
			intent_requested.emit("unequip_intent", {"slot": _slot_from_kind(source)})


func _item_can_equip_to(item: Dictionary, slot: String) -> bool:
	var def_id := str(item.get("item_def_id", ""))
	var def: Dictionary = _item_definition(def_id)
	if not bool(def.get("equippable", false)):
		return false
	var item_slot := str(def.get("slot", ""))
	if item_slot == "ring":
		return slot == "ring_left" or slot == "ring_right"
	return item_slot == slot


func _preferred_equip_slot(item: Dictionary) -> String:
	var def_id := str(item.get("item_def_id", ""))
	var def: Dictionary = _item_definition(def_id)
	if not bool(def.get("equippable", false)):
		return ""
	var item_slot := str(def.get("slot", ""))
	if item_slot == "ring":
		if equipped.get("ring_left", null) == null:
			return "ring_left"
		return "ring_right"
	return item_slot


func _is_consumable(item: Dictionary) -> bool:
	var def_id := str(item.get("item_def_id", ""))
	var def: Dictionary = _item_definition(def_id)
	return str(def.get("category", "")) == "consumable"


func _equipped_item(slot: String) -> Dictionary:
	var item_id = equipped.get(slot, null)
	if item_id == null:
		return {}
	for item in inventory:
		if str(item.get("item_instance_id", "")) == str(item_id):
			return item
	return {}


func _is_equipped_instance(item_instance_id: String) -> bool:
	if item_instance_id == "":
		return false
	for slot in EQUIPMENT_SLOTS:
		var equipped_id = equipped.get(str(slot), null)
		if equipped_id != null and str(equipped_id) == item_instance_id:
			return true
	return false


func _slot_kind_for_equipment(slot: String) -> String:
	return SLOT_KIND_EQUIP_PREFIX + slot


func _slot_kind_is_equipment(kind: String) -> bool:
	return kind.begins_with(SLOT_KIND_EQUIP_PREFIX)


func _slot_from_kind(kind: String) -> String:
	if not _slot_kind_is_equipment(kind):
		return ""
	return kind.substr(SLOT_KIND_EQUIP_PREFIX.length())


func _tooltip(item: Dictionary) -> String:
	return "\n".join(_tooltip_lines(item) + _requirement_lines_as_summary(item) + _comparison_text_lines(item))


func _tooltip_lines(item: Dictionary) -> Array:
	var def_id := str(item.get("item_def_id", ""))
	var def: Dictionary = _item_definition(def_id)
	var lines: Array = [str(item.get("display_name", def.get("name", def_id)))]
	var rarity := str(item.get("rarity", ""))
	if rarity != "":
		lines.append("Rarity: %s" % rarity.capitalize())
	var summary_lines := _detail_lines(item, false, false)
	if not summary_lines.is_empty():
		lines.append_array(summary_lines)
		_append_hotbar_tooltip_line(lines, item)
		return lines
	var slot := str(def.get("slot", ""))
	if slot != "":
		lines.append("Slot: %s" % slot)
	else:
		var category := str(def.get("category", ""))
		if category != "":
			lines.append("Kind: %s" % category)
	if def.has("reach"):
		lines.append("Reach: %s" % str(def["reach"]))
	if def.has("attack_mode"):
		lines.append("Mode: %s" % str(def["attack_mode"]))
	lines.append_array(_consumable_effect_lines(def))
	var base_stat_lines := _base_stat_lines(def)
	if not base_stat_lines.is_empty():
		lines.append(TOOLTIP_STAT_SEPARATOR)
		lines.append_array(base_stat_lines)
	var random_stat_lines := _random_stat_lines(item.get("rolled_stats", {}), def)
	if not random_stat_lines.is_empty():
		lines.append(TOOLTIP_STAT_SEPARATOR)
		lines.append_array(random_stat_lines)
	elif def.has("damage"):
		var dmg: Dictionary = def["damage"]
		lines.append("Damage: %s-%s" % [str(dmg.get("min", "?")), str(dmg.get("max", "?"))])
	_append_hotbar_tooltip_line(lines, item)
	return lines


func _base_stat_lines(def: Dictionary) -> Array:
	var stats_value = def.get("base_stats", {})
	if typeof(stats_value) != TYPE_DICTIONARY:
		return []
	return _stat_lines_for_tooltip(stats_value as Dictionary, false)


func _random_stat_lines(stats_value: Variant, def: Dictionary) -> Array:
	if typeof(stats_value) != TYPE_DICTIONARY:
		return []
	var base_stats: Dictionary = def.get("base_stats", {})
	var deltas: Dictionary = {}
	for key in (stats_value as Dictionary).keys():
		var total := int((stats_value as Dictionary).get(key, 0))
		var base := int(base_stats.get(key, 0))
		var delta := total - base
		if delta != 0:
			deltas[key] = delta
	return _stat_lines_for_tooltip(deltas, true)


func _stat_lines_for_tooltip(stats: Dictionary, signed: bool) -> Array:
	var lines: Array = []
	if int(stats.get("damage_min", 0)) > 0 or int(stats.get("damage_max", 0)) > 0:
		if signed:
			if int(stats.get("damage_min", 0)) != 0:
				lines.append("%s: %s" % [_display_stat("damage_min"), _format_stat_value(stats.get("damage_min", 0), false)])
			if int(stats.get("damage_max", 0)) != 0:
				lines.append("%s: %s" % [_display_stat("damage_max"), _format_stat_value(stats.get("damage_max", 0), false)])
		else:
			lines.append("Damage: %s-%s" % [str(stats.get("damage_min", "?")), str(stats.get("damage_max", "?"))])
	for key in ["armor", "block_percent", "attack_speed_percent", "max_hp", "max_mana", "health_regen_per_10_seconds", "mana_regen_per_10_seconds", "hotbar_slots", "inventory_rows"]:
		if not stats.has(key):
			continue
		var value := int(stats.get(key, 0))
		if value == 0:
			continue
		lines.append("%s: %s" % [_display_stat(key), _format_stat_value(value, key == "block_percent" or key == "attack_speed_percent")])
	return lines


func _format_stat_value(value: int, percent: bool) -> String:
	var sign := "+" if value > 0 else ""
	var suffix := "%" if percent else ""
	return "%s%d%s" % [sign, value, suffix]


func _append_hotbar_tooltip_line(lines: Array, item: Dictionary) -> void:
	var hotbar_labels := _hotbar_labels_for_item(str(item.get("item_instance_id", "")))
	if not hotbar_labels.is_empty():
		lines.append("Assigned to hotbar: %s" % ", ".join(hotbar_labels))


func _hotbar_slots_for_item(item_instance_id: String) -> Array:
	var slots: Array = []
	if item_instance_id == "":
		return slots
	for slot in hotbar:
		if typeof(slot) != TYPE_DICTIONARY:
			continue
		var rec := slot as Dictionary
		var assigned_id = rec.get("item_instance_id", null)
		if assigned_id != null and str(assigned_id) == item_instance_id:
			var slot_index := int(rec.get("slot_index", -1))
			if slot_index >= 0 and slot_index < HOTKEY_LABELS.size():
				slots.append(slot_index)
	return slots


func _hotbar_labels_for_item(item_instance_id: String) -> Array:
	var labels: Array = []
	for slot_index in _hotbar_slots_for_item(item_instance_id):
		var index := int(slot_index)
		if index >= 0 and index < HOTKEY_LABELS.size():
			labels.append(_hotbar_label_for_slot(index))
	return labels


func _hotbar_label_for_slot(slot_index: int) -> String:
	if slot_index >= 0 and slot_index < HOTKEY_LABELS.size():
		return HOTKEY_LABELS[slot_index]
	return str(slot_index + 1)


func _hotbar_assigned_inventory_count() -> int:
	var total := 0
	for item in _bag_items():
		if not _hotbar_slots_for_item(str((item as Dictionary).get("item_instance_id", ""))).is_empty():
			total += 1
	return total


func _consumable_effect_lines(def: Dictionary) -> Array:
	var lines: Array = []
	var heal = def.get("heal", {})
	if typeof(heal) == TYPE_DICTIONARY and not (heal as Dictionary).is_empty():
		lines.append("Restores %s HP" % _range_text(heal as Dictionary))
	var mana_restore = def.get("mana_restore", {})
	if typeof(mana_restore) == TYPE_DICTIONARY and not (mana_restore as Dictionary).is_empty():
		lines.append("Restores %s mana" % _range_text(mana_restore as Dictionary))
	return lines


func _range_text(value: Dictionary) -> String:
	var min_value := int(value.get("min", 0))
	var max_value := int(value.get("max", min_value))
	if min_value == max_value:
		return str(min_value)
	return "%d-%d" % [min_value, max_value]


func _detail_lines(item: Dictionary, include_requirements: bool = true, include_comparison: bool = true) -> Array:
	var lines: Array = []
	var summary = item.get("summary_lines", [])
	if typeof(summary) != TYPE_ARRAY:
		return lines
	for line in summary:
		var text := str(line)
		if text == "":
			continue
		if not include_requirements and _is_requirement_summary_line(text):
			continue
		if not include_comparison and _is_comparison_summary_line(text):
			continue
		lines.append(text)
	return lines


func _requirement_lines(item: Dictionary) -> Array:
	var lines: Array = []
	var statuses = item.get("requirement_status", [])
	if typeof(statuses) == TYPE_ARRAY:
		for status in statuses:
			if typeof(status) != TYPE_DICTIONARY:
				continue
			var rec := status as Dictionary
			var stat := str(rec.get("stat", ""))
			var required := int(rec.get("required", 0))
			if stat == "" or required <= 0:
				continue
			var current := int(rec.get("current", 0))
			var met := bool(rec.get("met", current >= required))
			var suffix := "" if met else "(%d)" % (current - required)
			lines.append({
				"text": "%s %d%s" % [_display_stat(stat), required, suffix],
				"color": _requirement_color(met),
			})
	if not lines.is_empty():
		return lines
	var requirements: Dictionary = item.get("requirements", {})
	if requirements.has("level"):
		lines.append("Level %s" % str(requirements["level"]))
	for key in requirements.keys():
		var stat := str(key)
		if stat == "level":
			continue
		lines.append("%s %s" % [_display_stat(stat), str(requirements.get(key, ""))])
	var summary = item.get("summary_lines", [])
	if typeof(summary) == TYPE_ARRAY:
		for line in summary:
			var parsed := _requirement_from_summary_line(str(line))
			if parsed != "" and not lines.has(parsed):
				lines.append(parsed)
	return lines


func _requirement_lines_as_summary(item: Dictionary) -> Array:
	var lines: Array = []
	for line in _requirement_lines(item):
		var text := _entry_text(line)
		if text.to_lower().begins_with("level "):
			lines.append("Requires %s" % text.to_lower())
		else:
			lines.append("Requires %s" % text)
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


func _comparison_entries(item: Dictionary) -> Array:
	var entries: Array = []
	_append_equip_preview_entries(entries, item)
	var comparison = item.get("comparison", {})
	if typeof(comparison) == TYPE_DICTIONARY:
		var deltas = (comparison as Dictionary).get("deltas", [])
		if typeof(deltas) == TYPE_ARRAY:
			for delta in deltas:
				if typeof(delta) != TYPE_DICTIONARY:
					continue
				var rec := delta as Dictionary
				var diff := float(rec.get("delta", 0.0))
				var sign := "+" if diff >= 0 else ""
				entries.append({
					"text": "%s%s %s vs equipped" % [sign, _format_delta(diff), _display_stat(str(rec.get("stat", "")))],
					"color": _comparison_color(diff),
				})
	var summary = item.get("summary_lines", [])
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
				"color": _comparison_color(float(diff_value)),
			})
	return entries


func _append_equip_preview_entries(entries: Array, item: Dictionary) -> void:
	var preview = item.get("equip_preview", {})
	if typeof(preview) != TYPE_DICTIONARY:
		return
	var deltas = (preview as Dictionary).get("deltas", [])
	if typeof(deltas) != TYPE_ARRAY:
		return
	for delta in deltas:
		if typeof(delta) != TYPE_DICTIONARY:
			continue
		var rec := delta as Dictionary
		var diff := float(rec.get("delta", 0.0))
		var sign := "+" if diff >= 0 else ""
		entries.append({
			"text": "%s%s %s preview" % [sign, _format_delta(diff), _display_stat(str(rec.get("stat", "")))],
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
	return float(stripped.substr(0, first_space))


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
		var diff := float(rec.get("delta", 0.0))
		var sign := "+" if diff >= 0 else ""
		lines.append("%s%s %s vs equipped" % [sign, _format_delta(diff), _display_stat(str(rec.get("stat", "")))])
	return lines


func _comparison_text_lines(item: Dictionary) -> Array:
	var lines: Array = []
	for entry in _comparison_entries(item):
		if typeof(entry) == TYPE_DICTIONARY:
			lines.append(str((entry as Dictionary).get("text", "")))
	return lines


func _entry_text(value) -> String:
	if typeof(value) == TYPE_DICTIONARY:
		return str((value as Dictionary).get("text", ""))
	return str(value)


func _requirement_color(met: bool) -> Color:
	return Color("#9ee6a8") if met else Color("#ff6f6f")


func _requirement_row_count() -> int:
	var total := 0
	for item in inventory:
		if typeof(item) == TYPE_DICTIONARY:
			total += _requirement_lines(item as Dictionary).size()
	for slot in EQUIPMENT_SLOTS:
		var item := _equipped_item(str(slot))
		if not item.is_empty():
			total += _requirement_lines(item).size()
	return total


func _equip_preview_row_count() -> int:
	var total := 0
	for item in inventory:
		if typeof(item) == TYPE_DICTIONARY:
			total += _equip_preview_count(item as Dictionary)
	for slot in EQUIPMENT_SLOTS:
		var item := _equipped_item(str(slot))
		if not item.is_empty():
			total += _equip_preview_count(item)
	return total


func _equip_preview_count(item: Dictionary) -> int:
	var preview = item.get("equip_preview", {})
	if typeof(preview) != TYPE_DICTIONARY:
		return 0
	var deltas = (preview as Dictionary).get("deltas", [])
	if typeof(deltas) != TYPE_ARRAY:
		return 0
	return (deltas as Array).size()


func _comparison_color(delta: float) -> Color:
	if delta > 0:
		return Color("#9ee6a8")
	if delta < 0:
		return Color("#ff9f7a")
	return Color("#d8c7a6")


func _display_stat(stat: String) -> String:
	return StatLabels.display_name(stat)


func _format_delta(delta: float) -> String:
	if absf(delta - roundf(delta)) < 0.0001:
		return str(int(roundf(delta)))
	return "%.2f" % delta


func _short_label(def_id: String) -> String:
	var def: Dictionary = item_rules.get(def_id, {})
	var name := str(def.get("name", def_id))
	var parts := name.split(" ")
	var out := ""
	for part in parts:
		if part.length() > 0:
			out += part.substr(0, 1).to_upper()
	return out.substr(0, 3)


func _item_definition(def_id: String) -> Dictionary:
	return ItemRulesLoader.item_definition(def_id)


func _debug_presentations() -> Dictionary:
	var out := {}
	for item in inventory:
		var def_id := str(item.get("item_def_id", ""))
		if def_id != "":
			out[def_id] = item_presentations.has(def_id)
	return out


func _debug_hotbar_assigned_item_ids() -> Array:
	var ids: Array = []
	for item in _bag_items():
		var item_id := str((item as Dictionary).get("item_instance_id", ""))
		if item_id != "" and not _hotbar_slots_for_item(item_id).is_empty():
			ids.append(item_id)
	return ids


func _debug_paper_doll_slots() -> Dictionary:
	var out := {}
	for slot in EQUIPMENT_SLOTS:
		var btn: InventorySlotButton = _equipment_slots.get(str(slot), null)
		out[str(slot)] = {
			"exists": btn != null,
			"position": {
				"x": btn.position.x if btn != null else 0.0,
				"y": btn.position.y if btn != null else 0.0,
			},
			"empty": _equipped_item(str(slot)).is_empty(),
			"label": str(EQUIPMENT_LABELS.get(str(slot), str(slot))),
		}
	return out
