extends Control
class_name MultiplayerSessionsPanel

signal host_requested
signal refresh_requested
signal join_requested(session_id: String)
signal back_requested

var _rows: VBoxContainer
var _empty_label: Label
var _error_label: Label
var _selected_session_id: String = ""
var _sessions: Array = []


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_STOP
	_build()
	visible = false


func show_panel() -> void:
	_sync_viewport_size()
	visible = true
	_error_label.text = ""


func hide_panel() -> void:
	visible = false


func set_error(message: String) -> void:
	_error_label.text = message


func set_sessions(sessions: Array) -> void:
	_sessions = sessions.duplicate(true)
	_selected_session_id = ""
	_render_sessions()


func select_first_session() -> void:
	if _sessions.is_empty():
		return
	var row: Dictionary = _sessions[0]
	_selected_session_id = str(row.get("session_id", ""))
	_render_sessions()


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"sessions": _sessions.duplicate(true),
		"selected_session_id": _selected_session_id,
		"error": _error_label.text if _error_label != null else "",
	}


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)


func _build() -> void:
	var bg := ColorRect.new()
	bg.color = Color(0.035, 0.04, 0.045, 0.93)
	bg.set_anchors_preset(Control.PRESET_FULL_RECT)
	add_child(bg)

	var panel := PanelContainer.new()
	panel.custom_minimum_size = Vector2(520, 430)
	panel.set_anchors_preset(Control.PRESET_CENTER)
	panel.offset_left = -260
	panel.offset_right = 260
	panel.offset_top = -215
	panel.offset_bottom = 215
	add_child(panel)

	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 8)
	panel.add_child(box)

	var title := Label.new()
	title.text = "Multiplayer"
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.add_theme_font_size_override("font_size", 36)
	box.add_child(title)

	var actions := HBoxContainer.new()
	actions.add_theme_constant_override("separation", 8)
	box.add_child(actions)
	actions.add_child(_button("Host Listed Session", host_requested.emit))
	actions.add_child(_button("Refresh Sessions", refresh_requested.emit))

	_empty_label = Label.new()
	_empty_label.text = "No listed sessions"
	_empty_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	box.add_child(_empty_label)

	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(480, 220)
	box.add_child(scroll)
	_rows = VBoxContainer.new()
	_rows.add_theme_constant_override("separation", 6)
	scroll.add_child(_rows)

	var bottom := HBoxContainer.new()
	bottom.add_theme_constant_override("separation", 8)
	box.add_child(bottom)
	bottom.add_child(_button("Join Selected", func() -> void:
		if _selected_session_id != "":
			join_requested.emit(_selected_session_id)
	))
	bottom.add_child(_button("Back", back_requested.emit))

	_error_label = Label.new()
	_error_label.add_theme_color_override("font_color", Color("#ff9b7a"))
	_error_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	box.add_child(_error_label)


func _button(text: String, callback: Callable) -> Button:
	var button := Button.new()
	button.text = text
	button.custom_minimum_size = Vector2(150, 40)
	button.pressed.connect(callback)
	return button


func _render_sessions() -> void:
	for child in _rows.get_children():
		child.queue_free()
	_empty_label.visible = _sessions.is_empty()
	for session in _sessions:
		if typeof(session) != TYPE_DICTIONARY:
			continue
		var row: Dictionary = session
		var session_id := str(row.get("session_id", ""))
		var button := Button.new()
		button.toggle_mode = true
		button.button_pressed = session_id == _selected_session_id
		button.text = "%s  %d/%d  %s" % [
			str(row.get("host_display_name", "Host")),
			int(row.get("connected_count", 0)),
			int(row.get("member_count", 0)),
			str(row.get("world_id", "")),
		]
		button.custom_minimum_size = Vector2(460, 38)
		button.pressed.connect(func() -> void:
			_selected_session_id = session_id
			_render_sessions()
		)
		_rows.add_child(button)
