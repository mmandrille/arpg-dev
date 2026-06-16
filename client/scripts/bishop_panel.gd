class_name BishopPanel
extends Control

signal respec_requested(bishop_entity_id: String)
signal revive_all_requested(bishop_entity_id: String)
signal debug_requested(action: String, bishop_entity_id: String)

const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const PANEL_SIZE := Vector2(320, 360)
const TITLE_FONT_SIZE := 29
const BODY_FONT_SIZE := 20

var bishop_entity_id: String = ""
var service_id: String = "bishop"
var price: int = 0
var affordable: bool = false
var gold: int = 0
var debug_enabled: bool = false

var _panel: DraggableWindow
var _title_label: Label
var _body_label: Label
var _respec_button: Button
var _revive_all_button: Button
var _debug_level_button: Button
var _debug_skill_button: Button
var _debug_stat_button: Button
var _status_label: Label
var _interactive: bool = true


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_build()
	visible = false


func show_bishop(next_entity_id: String, next_service_id: String, next_price: int, next_affordable: bool, next_gold: int) -> void:
	bishop_entity_id = next_entity_id
	service_id = next_service_id
	price = max(0, next_price)
	affordable = next_affordable
	gold = max(0, next_gold)
	if _status_label != null:
		_status_label.text = ""
	visible = true
	_render()


func set_gold(next_gold: int) -> void:
	gold = max(0, next_gold)
	affordable = gold >= price
	_render()


func hide_display() -> void:
	visible = false


func set_interactive(enabled: bool) -> void:
	_interactive = enabled
	_render()


func set_debug_enabled(enabled: bool) -> void:
	debug_enabled = enabled
	_render()


func show_status(text: String, error: bool = false) -> void:
	if _status_label == null:
		return
	_status_label.text = text
	_status_label.add_theme_color_override("font_color", Color("#ff9f7a") if error else Color("#9ee6a8"))


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"bishop_entity_id": bishop_entity_id,
		"service_id": service_id,
		"price": price,
		"gold": gold,
		"affordable": affordable,
		"respec_enabled": _respec_enabled(),
		"revive_all_enabled": _revive_all_enabled(),
		"debug_enabled": debug_enabled,
		"status": _status_label.text if _status_label != null else "",
		"window": _panel.get_debug_state() if _panel != null else {},
	}


func bot_click_respec() -> void:
	if _respec_enabled():
		_emit_respec()


func bot_click_revive_all() -> void:
	if _revive_all_enabled():
		_emit_revive_all()


func bot_click_debug(action: String) -> void:
	if _debug_action_enabled():
		_emit_debug(action)


func bot_click_close() -> void:
	if _panel != null and _panel.close_button() != null:
		_panel.close_button().pressed.emit()


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_reposition_panel()


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = DraggableWindowScript.new()
	_panel.custom_minimum_size = PANEL_SIZE
	_panel.configure("Bishop", Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 52))
	_panel.add_theme_stylebox_override("panel", _panel_style())
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	_panel.close_requested.connect(hide_display)
	add_child(_panel)
	_reposition_panel()
	_panel.set_layout_key("bishop")

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 10)
	root.custom_minimum_size = Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 52)
	_panel.set_content(root)

	_title_label = Label.new()
	_title_label.text = "Bishop"
	_title_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_title_label.add_theme_font_size_override("font_size", TITLE_FONT_SIZE)
	_title_label.add_theme_color_override("font_color", Color("#f4e5d2"))
	root.add_child(_title_label)

	_body_label = Label.new()
	_body_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_body_label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	_body_label.add_theme_color_override("font_color", Color("#d9c8b5"))
	root.add_child(_body_label)

	_respec_button = Button.new()
	_respec_button.custom_minimum_size = Vector2(PANEL_SIZE.x - 60, 42)
	_respec_button.pressed.connect(_emit_respec)
	root.add_child(_respec_button)

	_revive_all_button = Button.new()
	_revive_all_button.custom_minimum_size = Vector2(PANEL_SIZE.x - 60, 42)
	_revive_all_button.text = "Revive all dead heroes"
	_revive_all_button.pressed.connect(_emit_revive_all)
	root.add_child(_revive_all_button)

	_debug_level_button = Button.new()
	_debug_level_button.custom_minimum_size = Vector2(PANEL_SIZE.x - 60, 36)
	_debug_level_button.text = "Debug: gain level"
	_debug_level_button.pressed.connect(func() -> void: _emit_debug("level"))
	root.add_child(_debug_level_button)

	_debug_skill_button = Button.new()
	_debug_skill_button.custom_minimum_size = Vector2(PANEL_SIZE.x - 60, 36)
	_debug_skill_button.text = "Debug: gain skill point"
	_debug_skill_button.pressed.connect(func() -> void: _emit_debug("skill_point"))
	root.add_child(_debug_skill_button)

	_debug_stat_button = Button.new()
	_debug_stat_button.custom_minimum_size = Vector2(PANEL_SIZE.x - 60, 36)
	_debug_stat_button.text = "Debug: gain stat point"
	_debug_stat_button.pressed.connect(func() -> void: _emit_debug("stat_point"))
	root.add_child(_debug_stat_button)

	_status_label = Label.new()
	_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_status_label.add_theme_font_size_override("font_size", BODY_FONT_SIZE - 2)
	_status_label.add_theme_color_override("font_color", Color("#9ee6a8"))
	root.add_child(_status_label)
	_render()


func _render() -> void:
	if _panel == null:
		return
	_panel.configure("Bishop", Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 52))
	if _body_label != null:
		_body_label.text = "Restored health and mana."
	if _respec_button != null:
		_respec_button.text = "Respec" if price == 0 else "Respec - %d gold" % price
		_respec_button.disabled = not _respec_enabled()
	if _revive_all_button != null:
		_revive_all_button.disabled = not _revive_all_enabled()
	for button in [_debug_level_button, _debug_skill_button, _debug_stat_button]:
		if button != null:
			button.visible = debug_enabled
			button.disabled = not _debug_action_enabled()
	if _status_label != null and _status_label.text == "":
		_status_label.text = "Gold: %d" % gold
		_status_label.add_theme_color_override("font_color", Color("#d9c8b5"))


func _emit_respec() -> void:
	if not _respec_enabled():
		show_status("Not enough gold", true)
		return
	respec_requested.emit(bishop_entity_id)


func _emit_revive_all() -> void:
	if not _revive_all_enabled():
		return
	revive_all_requested.emit(bishop_entity_id)


func _emit_debug(action: String) -> void:
	if not _debug_action_enabled():
		return
	debug_requested.emit(action, bishop_entity_id)


func _respec_enabled() -> bool:
	return _interactive and visible and bishop_entity_id != "" and affordable and gold >= price


func _revive_all_enabled() -> bool:
	return _interactive and visible and bishop_entity_id != ""


func _debug_action_enabled() -> bool:
	return _interactive and visible and bishop_entity_id != "" and debug_enabled


func _reposition_panel() -> void:
	if _panel == null:
		return
	var viewport_size := get_viewport_rect().size
	_panel.position = Vector2(maxf(16.0, viewport_size.x - PANEL_SIZE.x - 22.0), 106.0)


func _panel_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color("#211818")
	style.border_color = Color("#8d2424")
	style.border_width_left = 2
	style.border_width_top = 2
	style.border_width_right = 2
	style.border_width_bottom = 2
	style.corner_radius_top_left = 8
	style.corner_radius_top_right = 8
	style.corner_radius_bottom_left = 8
	style.corner_radius_bottom_right = 8
	style.content_margin_left = 12
	style.content_margin_top = 12
	style.content_margin_right = 12
	style.content_margin_bottom = 12
	return style
