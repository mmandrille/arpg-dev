class_name InventoryPanel
extends Control

signal intent_requested(intent_type: String, payload: Dictionary)

const SLOT_KIND_BAG := "bag"
const SLOT_KIND_EQUIP_PREFIX := "equip:"
const SLOT_KIND_BAG_AREA := "bag_area"
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

var inventory: Array = []
var equipped: Dictionary = {}
var item_rules: Dictionary = {}
var item_templates: Dictionary = {}
var item_presentations: Dictionary = {}
var _panel: PanelContainer
var _equipment_slots: Dictionary = {}
var _bag_grid: GridContainer
var _drag_data: Dictionary = {}
var _interactive: bool = true
var _gesture_hint: Label
var _gesture_tween: Tween


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
		if slot_kind == SLOT_KIND_BAG_AREA:
			return panel._slot_kind_is_equipment(source)
		return false

	func _drop_data(_at_position: Vector2, data: Variant) -> void:
		panel._handle_drop_on_slot(slot_kind, data)


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_load_item_rules()
	_load_item_templates()
	_load_item_presentations()
	_build()
	visible = false


func toggle() -> void:
	if not _interactive:
		return
	visible = not visible
	_apply_interaction_filters()


func set_interactive(enabled: bool) -> void:
	_interactive = enabled
	_apply_interaction_filters()


func ensure_display_visible() -> void:
	visible = true
	_apply_interaction_filters()


func hide_display() -> void:
	visible = false


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


func set_inventory_state(next_inventory: Array, next_equipped: Dictionary) -> void:
	inventory = []
	for item in next_inventory:
		inventory.append((item as Dictionary).duplicate(true))
	equipped = next_equipped.duplicate(true)
	if _bag_grid != null:
		_render()


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"bag_count": inventory.size(),
		"equipped": equipped.duplicate(true),
		"equipped_main_hand": equipped.get("main_hand", null),
		"main_hand_item": _equipped_item("main_hand"),
		"weapon_item": _equipped_item("main_hand"),
		"item_presentations": _debug_presentations(),
	}


func get_weapon_slot_screen_center() -> Vector2:
	return get_equipment_slot_screen_center("main_hand")


func get_equipment_slot_screen_center(slot: String) -> Vector2:
	return _slot_screen_center(_equipment_slots.get(slot, null))


func get_bag_area_screen_center() -> Vector2:
	if _bag_grid == null:
		return Vector2.ZERO
	for child in _bag_grid.get_children():
		if child is InventorySlotButton and child.slot_kind == SLOT_KIND_BAG_AREA:
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
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = PanelContainer.new()
	_panel.custom_minimum_size = Vector2(560, 410)
	_panel.set_anchors_preset(Control.PRESET_BOTTOM_RIGHT)
	_reposition_panel()
	_panel.add_theme_stylebox_override("panel", _panel_style())
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	add_child(_panel)

	_gesture_hint = Label.new()
	_gesture_hint.visible = false
	_gesture_hint.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_gesture_hint.add_theme_color_override("font_color", Color("#c9a227"))
	_gesture_hint.add_theme_font_size_override("font_size", 15)
	_panel.add_child(_gesture_hint)

	var root := HBoxContainer.new()
	root.add_theme_constant_override("separation", 18)
	root.custom_minimum_size = Vector2(530, 380)
	_panel.add_child(root)

	var left := VBoxContainer.new()
	left.custom_minimum_size = Vector2(220, 0)
	root.add_child(left)
	left.add_child(_title("Equipment"))
	var equip_grid := GridContainer.new()
	equip_grid.columns = 2
	equip_grid.add_theme_constant_override("h_separation", 6)
	equip_grid.add_theme_constant_override("v_separation", 6)
	left.add_child(equip_grid)
	for slot in EQUIPMENT_SLOTS:
		var box := VBoxContainer.new()
		box.add_child(_caption(str(EQUIPMENT_LABELS.get(slot, slot))))
		var btn := _slot_button(_slot_kind_for_equipment(str(slot)), Vector2(84, 62))
		_equipment_slots[str(slot)] = btn
		box.add_child(btn)
		equip_grid.add_child(box)

	var right := VBoxContainer.new()
	right.custom_minimum_size = Vector2(290, 0)
	root.add_child(right)
	right.add_child(_caption("Bag"))
	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(285, 288)
	right.add_child(scroll)
	_bag_grid = GridContainer.new()
	_bag_grid.columns = 4
	_bag_grid.add_theme_constant_override("h_separation", 6)
	_bag_grid.add_theme_constant_override("v_separation", 6)
	scroll.add_child(_bag_grid)
	_render()


func _render() -> void:
	if _bag_grid == null:
		return
	for slot in EQUIPMENT_SLOTS:
		_fill_slot(_equipment_slots.get(slot, null), _equipped_item(str(slot)))
	for child in _bag_grid.get_children():
		child.queue_free()
	for item in inventory:
		if _is_equipped_instance(str(item.get("item_instance_id", ""))):
			continue
		var slot := _slot_button(SLOT_KIND_BAG, Vector2(58, 58))
		_fill_slot(slot, item)
		_bag_grid.add_child(slot)
	var bag_drop := _slot_button(SLOT_KIND_BAG_AREA, Vector2(58, 58))
	bag_drop.text = "+"
	bag_drop.tooltip_text = "Bag"
	_bag_grid.add_child(bag_drop)
	_position_gesture_hint()


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
		slot.text = ""
		slot.tooltip_text = "Empty"
		slot.queue_redraw()
		return
	var def_id := str(item.get("item_def_id", ""))
	slot.text = ""
	slot.tooltip_text = _tooltip(item)
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
	btn.add_theme_font_size_override("font_size", 14)
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
		"badge":
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
	var font_size := 12
	var text_size := font.get_string_size(label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size)
	slot.draw_string(font, center + Vector2(-text_size.x * 0.5, min_side * 0.38), label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, Color("#f4ead8"))


func _title(text: String) -> Label:
	var label := _caption(text)
	label.add_theme_font_size_override("font_size", 22)
	return label


func _caption(text: String) -> Label:
	var label := Label.new()
	label.text = text
	label.add_theme_color_override("font_color", Color("#c9a227"))
	label.add_theme_font_size_override("font_size", 15)
	return label


func _reposition_panel() -> void:
	if _panel == null:
		return
	var margin := 20.0
	var panel_size := _panel.custom_minimum_size
	var viewport_size := get_viewport_rect().size
	_panel.offset_right = -margin
	_panel.offset_bottom = -maxf(margin, minf(90.0, viewport_size.y * 0.10))
	_panel.offset_left = _panel.offset_right - panel_size.x
	_panel.offset_top = _panel.offset_bottom - panel_size.y
	if viewport_size.y > 0.0 and viewport_size.y + _panel.offset_top < margin:
		_panel.offset_top = -viewport_size.y + margin
	if viewport_size.x > 0.0 and viewport_size.x + _panel.offset_left < margin:
		_panel.offset_left = -viewport_size.x + margin


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


func _handle_double_click(item: Dictionary) -> void:
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
	elif slot_kind == SLOT_KIND_BAG_AREA:
		var source := str(data.get("source", ""))
		if _slot_kind_is_equipment(source):
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
	var def_id := str(item.get("item_def_id", ""))
	var def: Dictionary = _item_definition(def_id)
	var lines: Array[String] = [str(item.get("display_name", def.get("name", def_id)))]
	var rarity := str(item.get("rarity", ""))
	if rarity != "":
		lines.append("Rarity: %s" % rarity.capitalize())
	var slot := str(def.get("slot", ""))
	if slot != "":
		lines.append("Slot: %s" % slot)
	var rolled_stats: Dictionary = item.get("rolled_stats", {})
	if rolled_stats.has("damage_min") and rolled_stats.has("damage_max"):
		lines.append("Damage: %s-%s" % [str(rolled_stats.get("damage_min", "?")), str(rolled_stats.get("damage_max", "?"))])
	elif def.has("damage"):
		var dmg: Dictionary = def["damage"]
		lines.append("Damage: %s-%s" % [str(dmg.get("min", "?")), str(dmg.get("max", "?"))])
	if rolled_stats.has("max_hp"):
		lines.append("Max HP: +%s" % str(rolled_stats.get("max_hp", "?")))
	if rolled_stats.has("armor"):
		lines.append("Armor: +%s" % str(rolled_stats.get("armor", "?")))
	if rolled_stats.has("block_percent"):
		lines.append("Block: %s%%" % str(rolled_stats.get("block_percent", "?")))
	if rolled_stats.has("hotbar_slots"):
		lines.append("Hotbar slots: %s" % str(rolled_stats.get("hotbar_slots", "?")))
	if def.has("reach"):
		lines.append("Reach: %s" % str(def["reach"]))
	if def.has("attack_mode"):
		lines.append("Mode: %s" % str(def["attack_mode"]))
	var requirements: Dictionary = item.get("requirements", {})
	if requirements.has("level"):
		lines.append("Requires level %s" % str(requirements["level"]))
	return "\n".join(lines)


func _short_label(def_id: String) -> String:
	var def: Dictionary = item_rules.get(def_id, {})
	var name := str(def.get("name", def_id))
	var parts := name.split(" ")
	var out := ""
	for part in parts:
		if part.length() > 0:
			out += part.substr(0, 1).to_upper()
	return out.substr(0, 3)


func _load_item_rules() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/items.v0.json")
	if not FileAccess.file_exists(path):
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		item_rules = parsed.get("items", {})


func _load_item_templates() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/item_templates.v0.json")
	if not FileAccess.file_exists(path):
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		item_templates = parsed.get("templates", {})


func _item_definition(def_id: String) -> Dictionary:
	if item_rules.has(def_id):
		return item_rules.get(def_id, {})
	return item_templates.get(def_id, {})


func _load_item_presentations() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/item_presentations.v0.json")
	if not FileAccess.file_exists(path):
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		item_presentations = parsed.get("items", {})


func _debug_presentations() -> Dictionary:
	var out := {}
	for item in inventory:
		var def_id := str(item.get("item_def_id", ""))
		if def_id != "":
			out[def_id] = item_presentations.has(def_id)
	return out
