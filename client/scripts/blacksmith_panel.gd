class_name BlacksmithPanel
extends Control

signal upgrade_requested(stash_item_id: String)

const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const PANEL_SIZE := Vector2(520, 440)
const BODY_FONT_SIZE := 18
const DETAIL_FONT_SIZE := 15

var blacksmith_entity_id: String = ""
var stash_items: Array = []
var stash_gold: int = 0
var base_cost: int = 100
var growth_cost: int = 50
var max_level: int = 3
var _panel: DraggableWindow
var _status_label: Label
var _gold_label: Label
var _rows: VBoxContainer


func _ready() -> void:
	_build()
	hide_display()


func show_blacksmith(entity_id: String, next_stash_items: Array, next_stash_gold: int, config: Dictionary, status: String = "") -> void:
	if _panel == null:
		_build()
	blacksmith_entity_id = entity_id
	stash_items = _dup_array(next_stash_items)
	stash_gold = next_stash_gold
	base_cost = int(config.get("item_upgrade_cost_gold", base_cost))
	growth_cost = int(config.get("item_upgrade_cost_growth_per_level", growth_cost))
	max_level = int(config.get("item_upgrade_max_level", max_level))
	_status_label.text = status
	_rebuild()
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


func update_after_upgrade(item: Dictionary, next_stash_gold: int, charged_cost: int) -> void:
	var id := str(item.get("stash_item_id", ""))
	for i in range(stash_items.size()):
		if typeof(stash_items[i]) == TYPE_DICTIONARY and str((stash_items[i] as Dictionary).get("stash_item_id", "")) == id:
			stash_items[i] = item.duplicate(true)
			break
	stash_gold = next_stash_gold
	show_status("Upgraded for %d gold" % charged_cost)
	_rebuild()


func bot_click_upgrade(stash_item_id: String = "", item_def_id: String = "", stash_index: int = 0) -> void:
	var item := _matching_item(stash_item_id, item_def_id, stash_index)
	if item.is_empty():
		show_status("No matching stash item", true)
		return
	_emit_upgrade(item)


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"blacksmith_entity_id": blacksmith_entity_id,
		"stash_gold": stash_gold,
		"item_count": stash_items.size(),
		"rows": _debug_rows(),
		"status": _status_label.text if _status_label != null else "",
		"window": _panel.get_debug_state() if _panel != null else {},
	}


func _build() -> void:
	if _panel != null:
		return
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	set_anchors_preset(Control.PRESET_FULL_RECT)

	_panel = DraggableWindowScript.new()
	_panel.configure("Blacksmith", PANEL_SIZE)
	_panel.set_layout_key("blacksmith_panel")
	_panel.position = Vector2(300, 88)
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

	_gold_label = Label.new()
	_gold_label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	_gold_label.add_theme_color_override("font_color", Color("#f4d481"))
	root.add_child(_gold_label)

	var scroll := ScrollContainer.new()
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	root.add_child(scroll)
	_rows = VBoxContainer.new()
	_rows.add_theme_constant_override("separation", 6)
	scroll.add_child(_rows)


func _rebuild() -> void:
	_gold_label.text = "Stash gold: %d" % stash_gold
	_clear_rows()
	if stash_items.is_empty():
		_rows.add_child(_empty_label("Your account stash has no upgradeable items"))
		return
	for item in stash_items:
		if typeof(item) == TYPE_DICTIONARY:
			_rows.add_child(_item_row(item as Dictionary))


func _item_row(item: Dictionary) -> Control:
	var row := PanelContainer.new()
	row.add_theme_stylebox_override("panel", _row_style())
	var box := HBoxContainer.new()
	box.add_theme_constant_override("separation", 10)
	row.add_child(box)

	var text := VBoxContainer.new()
	text.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	box.add_child(text)

	var title := Label.new()
	title.text = _item_title(item)
	title.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	title.add_theme_color_override("font_color", _rarity_color(str(item.get("rarity", ""))))
	text.add_child(title)

	var level := _item_level(item)
	var cost := _next_cost(level)
	var detail := Label.new()
	detail.text = "Level %d/%d  Next: %d gold" % [level, max_level, cost]
	detail.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	detail.add_theme_color_override("font_color", Color("#d8c8a8"))
	text.add_child(detail)

	var button := Button.new()
	button.text = "Upgrade"
	button.disabled = level >= max_level or stash_gold < cost
	button.tooltip_text = "Max level" if level >= max_level else ("Need %d gold" % cost if stash_gold < cost else "Upgrade item")
	button.pressed.connect(func() -> void: _emit_upgrade(item))
	box.add_child(button)
	return row


func _emit_upgrade(item: Dictionary) -> void:
	var level := _item_level(item)
	var cost := _next_cost(level)
	if level >= max_level:
		show_status("Item is already at max level", true)
		return
	if stash_gold < cost:
		show_status("Need %d gold" % cost, true)
		return
	var stash_item_id := str(item.get("stash_item_id", ""))
	if stash_item_id == "":
		show_status("Missing stash item id", true)
		return
	upgrade_requested.emit(stash_item_id)


func _matching_item(stash_item_id: String, item_def_id: String, stash_index: int) -> Dictionary:
	var matches: Array = []
	for value in stash_items:
		if typeof(value) != TYPE_DICTIONARY:
			continue
		var item := value as Dictionary
		if stash_item_id != "" and str(item.get("stash_item_id", "")) != stash_item_id:
			continue
		if item_def_id != "" and str(item.get("item_def_id", "")) != item_def_id:
			continue
		matches.append(item)
	if matches.is_empty() or stash_index < 0 or stash_index >= matches.size():
		return {}
	return (matches[stash_index] as Dictionary).duplicate(true)


func _debug_rows() -> Array:
	var rows: Array = []
	for value in stash_items:
		if typeof(value) != TYPE_DICTIONARY:
			continue
		var item := value as Dictionary
		var level := _item_level(item)
		rows.append({
			"stash_item_id": str(item.get("stash_item_id", "")),
			"item_def_id": str(item.get("item_def_id", "")),
			"display_name": _item_title(item),
			"rarity": str(item.get("rarity", "")),
			"item_level": level,
			"next_cost_gold": _next_cost(level),
			"upgrade_enabled": level < max_level and stash_gold >= _next_cost(level),
		})
	return rows


func _item_level(item: Dictionary) -> int:
	var rolled = item.get("rolled_stats", {})
	if typeof(rolled) == TYPE_DICTIONARY:
		var payload := rolled as Dictionary
		if typeof(payload.get("stats", {})) == TYPE_DICTIONARY:
			return int((payload.get("stats", {}) as Dictionary).get("item_level", 0))
		return int(payload.get("item_level", 0))
	return 0


func _next_cost(level: int) -> int:
	return base_cost + level * growth_cost


func _item_title(item: Dictionary) -> String:
	var display := str(item.get("display_name", ""))
	if display != "":
		return display
	return str(item.get("item_def_id", "Unknown item")).replace("_", " ").capitalize()


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


func _clear_rows() -> void:
	for child in _rows.get_children():
		child.queue_free()


func _empty_label(text: String) -> Label:
	var empty := Label.new()
	empty.text = text
	empty.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	empty.add_theme_color_override("font_color", Color("#e8dcc8"))
	return empty


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
