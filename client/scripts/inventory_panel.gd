class_name InventoryPanel
extends Control

signal intent_requested(intent_type: String, payload: Dictionary)

const SLOT_KIND_BAG := "bag"
const SLOT_KIND_WEAPON := "weapon"
const SLOT_KIND_BAG_AREA := "bag_area"

var inventory: Array = []
var equipped: Dictionary = {"weapon": null}
var item_rules: Dictionary = {}
var _panel: PanelContainer
var _weapon_slot: InventorySlotButton
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
		if slot_kind == SLOT_KIND_WEAPON:
			return source == SLOT_KIND_BAG and panel._is_weapon(dragged)
		if slot_kind == SLOT_KIND_BAG_AREA:
			return source == SLOT_KIND_WEAPON
		return false

	func _drop_data(_at_position: Vector2, data: Variant) -> void:
		panel._handle_drop_on_slot(slot_kind, data)


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_load_item_rules()
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
	position = Vector2.ZERO
	size = get_viewport_rect().size


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
		"equipped_weapon": equipped.get("weapon", null),
		"weapon_item": _equipped_weapon_item(),
	}


func get_weapon_slot_screen_center() -> Vector2:
	return _slot_screen_center(_weapon_slot)


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
			if str(item.get("item_instance_id", "")) == str(equipped.get("weapon", "")):
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
	_panel.custom_minimum_size = Vector2(430, 300)
	_panel.set_anchors_preset(Control.PRESET_BOTTOM_RIGHT)
	_panel.offset_left = -450
	_panel.offset_top = -330
	_panel.offset_right = -20
	_panel.offset_bottom = -30
	_panel.add_theme_stylebox_override("panel", _panel_style())
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	add_child(_panel)

	_gesture_hint = Label.new()
	_gesture_hint.visible = false
	_gesture_hint.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_gesture_hint.add_theme_color_override("font_color", Color("#c9a227"))
	_gesture_hint.add_theme_font_size_override("font_size", 11)
	_panel.add_child(_gesture_hint)

	var root := HBoxContainer.new()
	root.add_theme_constant_override("separation", 18)
	root.custom_minimum_size = Vector2(400, 270)
	_panel.add_child(root)

	var left := VBoxContainer.new()
	left.custom_minimum_size = Vector2(120, 0)
	root.add_child(left)
	left.add_child(_title("Inventory"))
	left.add_child(_caption("Weapon"))
	_weapon_slot = _slot_button(SLOT_KIND_WEAPON, Vector2(76, 76))
	left.add_child(_weapon_slot)

	var right := VBoxContainer.new()
	right.custom_minimum_size = Vector2(250, 0)
	root.add_child(right)
	right.add_child(_caption("Bag"))
	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(245, 225)
	right.add_child(scroll)
	_bag_grid = GridContainer.new()
	_bag_grid.columns = 4
	_bag_grid.add_theme_constant_override("h_separation", 6)
	_bag_grid.add_theme_constant_override("v_separation", 6)
	scroll.add_child(_bag_grid)
	_render()


func _render() -> void:
	if _weapon_slot == null or _bag_grid == null:
		return
	_fill_slot(_weapon_slot, _equipped_weapon_item())
	for child in _bag_grid.get_children():
		child.queue_free()
	for item in inventory:
		if str(item.get("item_instance_id", "")) == str(equipped.get("weapon", "")):
			continue
		var slot := _slot_button(SLOT_KIND_BAG, Vector2(48, 48))
		_fill_slot(slot, item)
		_bag_grid.add_child(slot)
	var bag_drop := _slot_button(SLOT_KIND_BAG_AREA, Vector2(48, 48))
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
	slot.item = item.duplicate(true)
	if item.is_empty():
		slot.text = ""
		slot.tooltip_text = "Empty"
		return
	var def_id := str(item.get("item_def_id", ""))
	slot.text = _short_label(def_id)
	slot.tooltip_text = _tooltip(def_id)


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
	btn.add_theme_font_size_override("font_size", 10)
	return btn


func _title(text: String) -> Label:
	var label := _caption(text)
	label.add_theme_font_size_override("font_size", 18)
	return label


func _caption(text: String) -> Label:
	var label := Label.new()
	label.text = text
	label.add_theme_color_override("font_color", Color("#c9a227"))
	label.add_theme_font_size_override("font_size", 12)
	return label


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
	if _is_weapon(item):
		intent_requested.emit("equip_intent", {"item_instance_id": str(item.get("item_instance_id", "")), "slot": "weapon"})


func _handle_drop_on_slot(slot_kind: String, data: Variant) -> void:
	if typeof(data) != TYPE_DICTIONARY:
		return
	var item: Dictionary = data.get("item", {})
	if item.is_empty():
		return
	if slot_kind == SLOT_KIND_WEAPON and _is_weapon(item):
		intent_requested.emit("equip_intent", {"item_instance_id": str(item.get("item_instance_id", "")), "slot": "weapon"})
	elif slot_kind == SLOT_KIND_BAG_AREA:
		intent_requested.emit("unequip_intent", {"slot": "weapon"})


func _is_weapon(item: Dictionary) -> bool:
	var def_id := str(item.get("item_def_id", ""))
	var def: Dictionary = item_rules.get(def_id, {})
	return bool(def.get("equippable", false)) and str(def.get("slot", "")) == "weapon"


func _equipped_weapon_item() -> Dictionary:
	var weapon_id = equipped.get("weapon", null)
	if weapon_id == null:
		return {}
	for item in inventory:
		if str(item.get("item_instance_id", "")) == str(weapon_id):
			return item
	return {}


func _tooltip(def_id: String) -> String:
	var def: Dictionary = item_rules.get(def_id, {})
	var lines: Array[String] = [str(def.get("name", def_id))]
	var slot := str(def.get("slot", ""))
	if slot != "":
		lines.append("Slot: %s" % slot)
	if def.has("damage"):
		var dmg: Dictionary = def["damage"]
		lines.append("Damage: %s-%s" % [str(dmg.get("min", "?")), str(dmg.get("max", "?"))])
	if def.has("reach"):
		lines.append("Reach: %s" % str(def["reach"]))
	if def.has("attack_mode"):
		lines.append("Mode: %s" % str(def["attack_mode"]))
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
