class_name MaterialWalletPanel
extends Control

const DraggableWindowScript := preload("res://scripts/draggable_window.gd")

const PANEL_SIZE := Vector2(300, 216)
const BODY_FONT_SIZE := 14
const DETAIL_FONT_SIZE := 12

var resource_wallet: Dictionary = {}
var _panel: DraggableWindow
var _rows: VBoxContainer
var _empty_label: Label
var _row_debug: Array = []


func _ready() -> void:
	ItemRulesLoader.ensure_loaded()
	set_anchors_preset(Control.PRESET_FULL_RECT)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	visible = false


func show_wallet(next_wallet: Dictionary) -> void:
	resource_wallet = next_wallet.duplicate(true)
	if _wallet_resource_keys().is_empty():
		hide_display()
		return
	_build()
	visible = true
	_render()


func set_wallet(next_wallet: Dictionary) -> void:
	resource_wallet = next_wallet.duplicate(true)
	if not is_open():
		return
	if _wallet_resource_keys().is_empty():
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
		"row_count": _row_debug.size(),
		"rows": _row_debug.duplicate(true),
		"text": _debug_text(),
		"window": _panel.get_debug_state() if _panel != null else {},
	}


func _build() -> void:
	if _panel != null:
		return
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = DraggableWindowScript.new()
	_panel.configure("Material Wallet", PANEL_SIZE)
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
	title.text = "Account materials"
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.add_theme_font_size_override("font_size", 16)
	title.add_theme_color_override("font_color", Color("#f0dfbb"))
	root.add_child(title)

	_rows = VBoxContainer.new()
	_rows.add_theme_constant_override("separation", 6)
	root.add_child(_rows)

	_empty_label = Label.new()
	_empty_label.text = "No materials stored"
	_empty_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_empty_label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	_empty_label.add_theme_color_override("font_color", Color("#8f826b"))
	root.add_child(_empty_label)


func _render() -> void:
	if _rows == null:
		return
	for child in _rows.get_children():
		_rows.remove_child(child)
		child.queue_free()
	_row_debug.clear()
	for resource_id in _wallet_resource_keys():
		var amount := int(resource_wallet.get(resource_id, 0))
		_rows.add_child(_resource_row(str(resource_id), amount))
	if _empty_label != null:
		_empty_label.visible = _row_debug.is_empty()


func _resource_row(resource_id: String, amount: int) -> Control:
	var row := PanelContainer.new()
	row.add_theme_stylebox_override("panel", _row_style())

	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 2)
	row.add_child(box)

	var name := _resource_name(resource_id)
	var category := _resource_category(resource_id)
	var header := Label.new()
	header.text = "%s x%d" % [name, amount]
	header.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	header.add_theme_color_override("font_color", Color("#b8e6ff"))
	box.add_child(header)

	if category != "":
		box.add_child(_detail_label("Category: %s" % category))
	box.add_child(_detail_label("Stored account-wide"))

	_row_debug.append({
		"resource_id": resource_id,
		"name": name,
		"amount": amount,
		"category": category,
		"storage": "Stored account-wide",
		"text": _row_text(name, amount, category),
	})
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
		var amount := int(resource_wallet.get(key, 0))
		if amount > 0:
			out.append(key)
	return out


func _resource_name(resource_id: String) -> String:
	var def := ItemRulesLoader.item_definition(resource_id)
	if def.has("name"):
		return str(def.get("name", ""))
	return resource_id.replace("_", " ").capitalize()


func _resource_category(resource_id: String) -> String:
	var def := ItemRulesLoader.item_definition(resource_id)
	return str(def.get("category", "")).replace("_", " ").capitalize()


func _row_text(name: String, amount: int, category: String) -> String:
	var parts := ["%s x%d" % [name, amount]]
	if category != "":
		parts.append("Category: %s" % category)
	parts.append("Stored account-wide")
	return " | ".join(parts)


func _debug_text() -> String:
	var lines: Array = []
	for row in _row_debug:
		lines.append(str((row as Dictionary).get("text", "")))
	return "\n".join(lines)


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
