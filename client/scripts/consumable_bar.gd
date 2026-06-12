class_name ConsumableBar
extends Control

signal intent_requested(intent_type: String, payload: Dictionary)

const SLOT_COUNT := 10
const ItemIconDrawerScript := preload("res://scripts/item_icon_drawer.gd")
const HOTKEY_LABELS := ["1", "2", "3", "4", "5", "6", "7", "8", "9", "0"]

var inventory: Array = []
var item_rules: Dictionary:
	get: return ItemRulesLoader.item_rules
var item_presentations: Dictionary:
	get: return ItemRulesLoader.item_presentations
var progression: Dictionary = {}
var progression_rules: Dictionary = {}
var hotbar_capacity: int = 2
var hotbar: Array = []
var _slots: Array = []
var _slot_items: Array = []
var _drag_data: Dictionary = {}
var _interactive: bool = true
var _panel: PanelContainer
var _xp_bar: ProgressBar
var _xp_label: Label


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
		bar._drag_data = {"source": "consumable_bar", "slot_index": slot_index, "item": item}
		return bar._drag_data

	func _can_drop_data(_at_position: Vector2, data: Variant) -> bool:
		if not bar._interactive or typeof(data) != TYPE_DICTIONARY:
			return false
		var source := str(data.get("source", ""))
		var dragged: Dictionary = data.get("item", {})
		if dragged.is_empty():
			return false
		if source == "bag":
			return bar._is_slot_enabled(slot_index) and bar._is_consumable(dragged)
		if source == "consumable_bar":
			return bar._is_slot_enabled(slot_index)
		return false

	func _drop_data(_at_position: Vector2, data: Variant) -> void:
		bar._handle_drop_on_slot(slot_index, data)


func _ready() -> void:
	ItemRulesLoader.ensure_loaded()
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_load_progression_rules()
	_slot_items.resize(SLOT_COUNT)
	for i in SLOT_COUNT:
		_slot_items[i] = {}
	_build()


func _notification(what: int) -> void:
	if what == NOTIFICATION_DRAG_END and _interactive and not _drag_data.is_empty():
		if not get_viewport().gui_is_drag_successful() and str(_drag_data.get("source", "")) == "consumable_bar":
			var slot_index := int(_drag_data.get("slot_index", -1))
			if slot_index >= 0 and slot_index < SLOT_COUNT:
				intent_requested.emit("assign_hotbar_intent", {"slot_index": slot_index, "item_instance_id": null})
		_drag_data = {}


func set_interactive(enabled: bool) -> void:
	_interactive = enabled
	if _panel != null:
		_panel.mouse_filter = Control.MOUSE_FILTER_STOP if _interactive else Control.MOUSE_FILTER_IGNORE


func set_inventory_state(next_inventory: Array) -> void:
	inventory = []
	for item in next_inventory:
		inventory.append((item as Dictionary).duplicate(true))
	_rebuild_slot_items()
	_render()


func set_hotbar_state(next_capacity: int, next_hotbar: Array) -> void:
	hotbar_capacity = clamp(next_capacity, 2, SLOT_COUNT)
	hotbar = []
	for slot in next_hotbar:
		hotbar.append((slot as Dictionary).duplicate(true))
	_rebuild_slot_items()
	_render()


func apply_hotbar_update(slot_index: int, item_instance_id, item: Dictionary = {}) -> void:
	if slot_index < 0 or slot_index >= SLOT_COUNT:
		return
	while hotbar.size() < SLOT_COUNT:
		hotbar.append({"slot_index": hotbar.size(), "item_instance_id": null})
	var slot := {"slot_index": slot_index, "item_instance_id": item_instance_id}
	if not item.is_empty():
		slot["item"] = item.duplicate(true)
	hotbar[slot_index] = slot
	_rebuild_slot_items()
	_render()


func set_character_progression(next_progression: Dictionary) -> void:
	progression = next_progression.duplicate(true)
	_render_xp()


func use_slot(slot_index: int) -> void:
	if not _interactive or slot_index < 0 or slot_index >= SLOT_COUNT:
		return
	if not _is_slot_enabled(slot_index):
		return
	var item: Dictionary = _slot_items[slot_index]
	if item.is_empty():
		return
	intent_requested.emit("use_hotbar_intent", {"slot_index": slot_index})


func assign_slot(slot_index: int, item_instance_id: String) -> void:
	if slot_index < 0 or slot_index >= SLOT_COUNT:
		return
	var item := _find_inventory_item(item_instance_id)
	if item.is_empty() or not _is_consumable(item):
		return
	intent_requested.emit("assign_hotbar_intent", {"slot_index": slot_index, "item_instance_id": item_instance_id})


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
	var level := int(progression.get("level", 1))
	var xp := int(progression.get("experience", 0))
	var remaining = progression.get("experience_to_next_level", null)
	return {
		"assigned_slots": assigned,
		"hotbar_capacity": hotbar_capacity,
		"xp_bar": {
			"level": level,
			"experience": xp,
			"experience_to_next_level": remaining,
			"progress": _xp_progress(level, xp, remaining),
			"label": _xp_label.text if _xp_label != null else "",
		},
	}


func get_slot_screen_center(slot_index: int) -> Vector2:
	if slot_index < 0 or slot_index >= _slots.size():
		return Vector2.ZERO
	var slot: Control = _slots[slot_index]
	if slot == null or not is_inside_tree():
		return Vector2.ZERO
	return slot.get_global_rect().get_center()


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	if _panel != null:
		_position_panel()
	if _xp_bar != null:
		_position_xp_bar()


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
		slot.add_theme_font_size_override("font_size", 15)
		slot.text = HOTKEY_LABELS[i]
		row.add_child(slot)
		_slots.append(slot)

	_position_panel()
	_build_xp_bar()
	_render()


func _position_panel() -> void:
	if _panel == null:
		return
	var vp := get_viewport_rect().size
	var panel_w := float(SLOT_COUNT * 58)
	_panel.position = Vector2((vp.x - panel_w) * 0.5, vp.y - 78)
	_panel.size = Vector2(panel_w, 64)


func _build_xp_bar() -> void:
	_xp_bar = ProgressBar.new()
	_xp_bar.min_value = 0.0
	_xp_bar.max_value = 1.0
	_xp_bar.step = 0.001
	_xp_bar.show_percentage = false
	_xp_bar.add_theme_stylebox_override("background", _xp_bar_bg_style())
	_xp_bar.add_theme_stylebox_override("fill", _xp_bar_fill_style())
	add_child(_xp_bar)
	_xp_label = Label.new()
	_xp_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_xp_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	_xp_label.add_theme_color_override("font_color", Color("#f0dfbb"))
	_xp_label.add_theme_font_size_override("font_size", 9)
	_xp_bar.add_child(_xp_label)
	_position_xp_bar()
	_render_xp()


func _position_xp_bar() -> void:
	if _xp_bar == null:
		return
	var vp := get_viewport_rect().size
	var panel_w := float(SLOT_COUNT * 58)
	_xp_bar.position = Vector2((vp.x - panel_w) * 0.5, vp.y - 12)
	_xp_bar.size = Vector2(panel_w, 8)
	if _xp_label != null:
		_xp_label.position = Vector2.ZERO
		_xp_label.size = _xp_bar.size


func _render() -> void:
	for i in SLOT_COUNT:
		var slot: ConsumableSlotButton = _slots[i]
		if slot == null:
			continue
		var item: Dictionary = _slot_items[i]
		slot.item = item.duplicate(true) if not item.is_empty() else {}
		slot.text = HOTKEY_LABELS[i]
		slot.tooltip_text = _tooltip(item)
		slot.disabled = not _is_slot_enabled(i)
		slot.modulate.a = 1.0 if _is_slot_enabled(i) else 0.42
		slot.queue_redraw()
	_render_xp()


func _render_xp() -> void:
	if _xp_bar == null:
		return
	var level := int(progression.get("level", 1))
	var xp := int(progression.get("experience", 0))
	var remaining = progression.get("experience_to_next_level", null)
	_xp_bar.value = _xp_progress(level, xp, remaining)
	if _xp_label != null:
		_xp_label.text = "Level %d" % level if remaining == null else "%d XP" % xp


func _handle_drop_on_slot(slot_index: int, data: Variant) -> void:
	if typeof(data) != TYPE_DICTIONARY:
		return
	var dragged: Dictionary = data.get("item", {})
	if dragged.is_empty() or not _is_consumable(dragged):
		return
	if not _is_slot_enabled(slot_index):
		return
	intent_requested.emit("assign_hotbar_intent", {"slot_index": slot_index, "item_instance_id": str(dragged.get("item_instance_id", ""))})


func _rebuild_slot_items() -> void:
	_slot_items.resize(SLOT_COUNT)
	for i in SLOT_COUNT:
		_slot_items[i] = {}
	for slot in hotbar:
		var slot_index := int((slot as Dictionary).get("slot_index", -1))
		if slot_index < 0 or slot_index >= SLOT_COUNT:
			continue
		var item_id = (slot as Dictionary).get("item_instance_id", null)
		if item_id == null:
			continue
		var item := _find_inventory_item(str(item_id))
		if item.is_empty() and (slot as Dictionary).has("item"):
			item = ((slot as Dictionary).get("item", {}) as Dictionary).duplicate(true)
		if not item.is_empty() and _is_consumable(item):
			_slot_items[slot_index] = item.duplicate(true)


func _inventory_has_item(item_instance_id: String) -> bool:
	return not _find_inventory_item(item_instance_id).is_empty()


func _is_slot_enabled(slot_index: int) -> bool:
	return slot_index >= 0 and slot_index < hotbar_capacity


func _find_inventory_item(item_instance_id: String) -> Dictionary:
	for item in inventory:
		if str(item.get("item_instance_id", "")) == item_instance_id:
			return item
	return {}


func _is_consumable(item: Dictionary) -> bool:
	var def_id := str(item.get("item_def_id", ""))
	var def: Dictionary = item_rules.get(def_id, {})
	return str(def.get("category", "")) == "consumable"


func _xp_progress(level: int, xp: int, remaining) -> float:
	if remaining == null:
		return 1.0
	var prev_threshold := 0
	var next_threshold := xp + int(remaining)
	var levels: Array = progression_rules.get("experience_curve", {}).get("levels", [])
	for row in levels:
		var row_level := int(row.get("level", 0))
		var threshold := int(row.get("next_level_total_xp", 0))
		if row_level == level - 1:
			prev_threshold = threshold
		if row_level == level:
			next_threshold = threshold
	var needed = max(1, next_threshold - prev_threshold)
	var current = clamp(xp - prev_threshold, 0, needed)
	return float(current) / float(needed)


func _tooltip(item: Dictionary) -> String:
	if item.is_empty():
		return "Empty consumable slot"
	var def_id := str(item.get("item_def_id", ""))
	var def: Dictionary = item_rules.get(def_id, {})
	var name := str(def.get("name", def_id))
	var heal: Dictionary = def.get("heal", {})
	var mana_restore: Dictionary = def.get("mana_restore", {})
	if not mana_restore.is_empty():
		return "%s (mana %s-%s)" % [name, str(mana_restore.get("min", "?")), str(mana_restore.get("max", "?"))]
	if heal.is_empty():
		return name
	return "%s (heal %s-%s)" % [name, str(heal.get("min", "?")), str(heal.get("max", "?"))]


func _draw_item_icon(slot: Control, item: Dictionary) -> void:
	var def_id := str(item.get("item_def_id", ""))
	var icon: Dictionary = item_presentations.get(def_id, {}).get("icon", {})
	var rect := Rect2(Vector2.ZERO, slot.size)
	ItemIconDrawerScript.draw(slot, rect, icon, "", false, 0.38, 16)


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


func _xp_bar_bg_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.025, 0.022, 0.018, 0.90)
	s.border_color = Color("#3f3423")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	return s


func _xp_bar_fill_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color("#c9a227")
	return s




func _load_progression_rules() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/character_progression.v0.json")
	var parsed = _read_json(path)
	if typeof(parsed) == TYPE_DICTIONARY:
		progression_rules = parsed


func _read_json(path: String) -> Variant:
	if not FileAccess.file_exists(path):
		return null
	var text := FileAccess.get_file_as_string(path)
	return JSON.parse_string(text)
