class_name ConsumableBar
extends Control

signal intent_requested(intent_type: String, payload: Dictionary)

const SLOT_COUNT := 10
const HOTKEY_LABELS := ["1", "2", "3", "4", "5", "6", "7", "8", "9", "0"]

var inventory: Array = []
var item_rules: Dictionary = {}
var item_presentations: Dictionary = {}
var _slots: Array = []
var _slot_items: Array = []
var _interactive: bool = true
var _panel: PanelContainer


class ConsumableSlotButton:
	extends Button

	var bar: ConsumableBar
	var slot_index: int = 0
	var item: Dictionary = {}

	func _draw() -> void:
		if item.is_empty():
			return
		bar._draw_item_icon(self, item)

	func _get_drag_data(_at_position: Vector2) -> Variant:
		if not bar._interactive or item.is_empty():
			return null
		var preview := Label.new()
		preview.text = str(item.get("item_def_id", "item"))
		preview.add_theme_color_override("font_color", Color("#e8dcc8"))
		set_drag_preview(preview)
		return {"source": "consumable_bar", "slot_index": slot_index, "item": item}

	func _can_drop_data(_at_position: Vector2, data: Variant) -> bool:
		if not bar._interactive or typeof(data) != TYPE_DICTIONARY:
			return false
		var source := str(data.get("source", ""))
		var dragged: Dictionary = data.get("item", {})
		if dragged.is_empty():
			return false
		if source == "bag":
			return bar._is_consumable(dragged)
		if source == "consumable_bar":
			return true
		return false

	func _drop_data(_at_position: Vector2, data: Variant) -> void:
		bar._handle_drop_on_slot(slot_index, data)


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_load_item_rules()
	_load_item_presentations()
	_slot_items.resize(SLOT_COUNT)
	for i in SLOT_COUNT:
		_slot_items[i] = {}
	_build()


func set_interactive(enabled: bool) -> void:
	_interactive = enabled
	if _panel != null:
		_panel.mouse_filter = Control.MOUSE_FILTER_STOP if _interactive else Control.MOUSE_FILTER_IGNORE


func set_inventory_state(next_inventory: Array) -> void:
	inventory = []
	for item in next_inventory:
		inventory.append((item as Dictionary).duplicate(true))
	_prune_slots()
	_render()


func use_slot(slot_index: int) -> void:
	if not _interactive or slot_index < 0 or slot_index >= SLOT_COUNT:
		return
	var item: Dictionary = _slot_items[slot_index]
	if item.is_empty():
		return
	if not _inventory_has_item(str(item.get("item_instance_id", ""))):
		_slot_items[slot_index] = {}
		_render()
		return
	intent_requested.emit("use_intent", {"item_instance_id": str(item.get("item_instance_id", ""))})


func assign_slot(slot_index: int, item_instance_id: String) -> void:
	if slot_index < 0 or slot_index >= SLOT_COUNT:
		return
	var item := _find_inventory_item(item_instance_id)
	if item.is_empty() or not _is_consumable(item):
		return
	_slot_items[slot_index] = item.duplicate(true)
	_render()


func get_debug_state() -> Dictionary:
	var assigned: Array = []
	for i in SLOT_COUNT:
		var item: Dictionary = _slot_items[i]
		if item.is_empty():
			assigned.append(null)
		else:
			assigned.append({
				"slot": i + 1,
				"item_def_id": item.get("item_def_id", ""),
				"item_instance_id": item.get("item_instance_id", ""),
			})
	return {"assigned_slots": assigned}


func get_slot_screen_center(slot_index: int) -> Vector2:
	if slot_index < 0 or slot_index >= _slots.size():
		return Vector2.ZERO
	var slot: Control = _slots[slot_index]
	if slot == null or not is_inside_tree():
		return Vector2.ZERO
	return slot.get_global_rect().get_center()


func _sync_viewport_size() -> void:
	position = Vector2.ZERO
	size = get_viewport_rect().size
	if _panel != null:
		_position_panel()


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = PanelContainer.new()
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	_panel.add_theme_stylebox_override("panel", _panel_style())
	add_child(_panel)

	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 6)
	_panel.add_child(row)

	for i in SLOT_COUNT:
		var slot := ConsumableSlotButton.new()
		slot.bar = self
		slot.slot_index = i
		slot.custom_minimum_size = Vector2(52, 52)
		slot.focus_mode = Control.FOCUS_NONE
		slot.clip_text = true
		slot.add_theme_stylebox_override("normal", _slot_style(false))
		slot.add_theme_stylebox_override("hover", _slot_style(true))
		slot.add_theme_stylebox_override("pressed", _slot_style(true))
		slot.add_theme_color_override("font_color", Color("#8b7a62"))
		slot.add_theme_font_size_override("font_size", 10)
		slot.text = HOTKEY_LABELS[i]
		row.add_child(slot)
		_slots.append(slot)

	_position_panel()
	_render()


func _position_panel() -> void:
	if _panel == null:
		return
	var vp := get_viewport_rect().size
	var panel_w := float(SLOT_COUNT * 58)
	_panel.position = Vector2((vp.x - panel_w) * 0.5, vp.y - 78)
	_panel.size = Vector2(panel_w, 64)


func _render() -> void:
	for i in SLOT_COUNT:
		var slot: ConsumableSlotButton = _slots[i]
		if slot == null:
			continue
		var item: Dictionary = _slot_items[i]
		slot.item = item.duplicate(true) if not item.is_empty() else {}
		slot.text = HOTKEY_LABELS[i]
		slot.tooltip_text = _tooltip(item)
		slot.queue_redraw()


func _handle_drop_on_slot(slot_index: int, data: Variant) -> void:
	if typeof(data) != TYPE_DICTIONARY:
		return
	var dragged: Dictionary = data.get("item", {})
	if dragged.is_empty() or not _is_consumable(dragged):
		return
	_slot_items[slot_index] = dragged.duplicate(true)


func _prune_slots() -> void:
	for i in SLOT_COUNT:
		var item: Dictionary = _slot_items[i]
		if item.is_empty():
			continue
		if not _inventory_has_item(str(item.get("item_instance_id", ""))):
			_slot_items[i] = {}


func _inventory_has_item(item_instance_id: String) -> bool:
	return not _find_inventory_item(item_instance_id).is_empty()


func _find_inventory_item(item_instance_id: String) -> Dictionary:
	for item in inventory:
		if str(item.get("item_instance_id", "")) == item_instance_id:
			return item
	return {}


func _is_consumable(item: Dictionary) -> bool:
	var def_id := str(item.get("item_def_id", ""))
	var def: Dictionary = item_rules.get(def_id, {})
	return str(def.get("category", "")) == "consumable"


func _tooltip(item: Dictionary) -> String:
	if item.is_empty():
		return "Empty consumable slot"
	var def_id := str(item.get("item_def_id", ""))
	var def: Dictionary = item_rules.get(def_id, {})
	var name := str(def.get("name", def_id))
	var heal: Dictionary = def.get("heal", {})
	if heal.is_empty():
		return name
	return "%s (heal %s-%s)" % [name, str(heal.get("min", "?")), str(heal.get("max", "?"))]


func _draw_item_icon(slot: Control, item: Dictionary) -> void:
	var def_id := str(item.get("item_def_id", ""))
	var icon: Dictionary = item_presentations.get(def_id, {}).get("icon", {})
	var shape := str(icon.get("shape", "box"))
	var color := Color(str(icon.get("color", "#d8d0bd")))
	var accent := Color(str(icon.get("accent", "#6b5420")))
	var rect := Rect2(Vector2.ZERO, slot.size)
	var center := rect.get_center()
	var min_side = min(rect.size.x, rect.size.y)

	match shape:
		"potion":
			slot.draw_rect(Rect2(center + Vector2(-min_side * 0.13, -min_side * 0.05), Vector2(min_side * 0.26, min_side * 0.28)), color, true)
			slot.draw_rect(Rect2(center + Vector2(-min_side * 0.08, -min_side * 0.22), Vector2(min_side * 0.16, min_side * 0.16)), accent, true)
		_:
			slot.draw_rect(Rect2(center - Vector2(min_side * 0.20, min_side * 0.20), Vector2(min_side * 0.40, min_side * 0.40)), color, true)


func _panel_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.07, 0.06, 0.05, 0.88)
	s.border_color = Color("#6b5420")
	s.border_width_left = 2
	s.border_width_top = 2
	s.border_width_right = 2
	s.border_width_bottom = 2
	s.content_margin_left = 8
	s.content_margin_top = 6
	s.content_margin_right = 8
	s.content_margin_bottom = 6
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


func _load_item_rules() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/items.v0.json")
	var parsed = _read_json(path)
	if typeof(parsed) == TYPE_DICTIONARY:
		item_rules = parsed.get("items", {})


func _load_item_presentations() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/item_presentations.v0.json")
	var parsed = _read_json(path)
	if typeof(parsed) == TYPE_DICTIONARY:
		item_presentations = parsed.get("items", {})


func _read_json(path: String) -> Variant:
	if not FileAccess.file_exists(path):
		return null
	var text := FileAccess.get_file_as_string(path)
	return JSON.parse_string(text)
