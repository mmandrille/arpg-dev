extends Control
class_name MultiplayerSessionsPanel

signal refresh_requested
signal join_requested(session_id: String)
signal back_requested

var _title: Label
var _rows: VBoxContainer
var _empty_label: Label
var _error_label: Label
var _search_input: LineEdit
var _sort_option: OptionButton
var _selected_session_id: String = ""
var _sessions: Array = []
var _visible_sessions: Array = []
var _search_text: String = ""
var _sort_mode: String = "recent"


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
	if _visible_sessions.is_empty():
		return
	var row: Dictionary = _visible_sessions[0]
	_selected_session_id = str(row.get("session_id", ""))
	_render_sessions()


func select_session(session_id: String) -> void:
	for session in _sessions:
		if typeof(session) == TYPE_DICTIONARY and str((session as Dictionary).get("session_id", "")) == session_id:
			_selected_session_id = session_id
			_render_sessions()
			return


func join_first_session() -> void:
	if _visible_sessions.is_empty():
		return
	var row: Dictionary = _visible_sessions[0]
	_join_session(str(row.get("session_id", "")))


func join_session(session_id: String) -> void:
	for session in _sessions:
		if typeof(session) == TYPE_DICTIONARY and str((session as Dictionary).get("session_id", "")) == session_id:
			_join_session(session_id)
			return


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"title": _title.text if _title != null else "",
		"sessions": _visible_sessions.duplicate(true),
		"total_session_count": _sessions.size(),
		"filtered_session_count": _visible_sessions.size(),
		"search_text": _search_text,
		"sort_mode": _sort_mode,
		"selected_session_id": _selected_session_id,
		"error": _error_label.text if _error_label != null else "",
		"actions": ["refresh_sessions", "join_first_listed_session", "join_expected_session", "back"],
	}


func bot_set_search(text: String) -> void:
	_search_text = text.strip_edges()
	if _search_input != null:
		_search_input.text = _search_text
	_render_sessions()


func bot_select_sort(mode: String) -> void:
	_set_sort_mode(mode)
	_render_sessions()


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

	_title = Label.new()
	_title.text = "Join Game"
	_title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_title.add_theme_font_size_override("font_size", 36)
	box.add_child(_title)

	var actions := HBoxContainer.new()
	actions.add_theme_constant_override("separation", 8)
	box.add_child(actions)
	actions.add_child(_button("Refresh", refresh_requested.emit))

	var filter_row := HBoxContainer.new()
	filter_row.add_theme_constant_override("separation", 8)
	box.add_child(filter_row)

	_search_input = LineEdit.new()
	_search_input.placeholder_text = "Search host, world, session"
	_search_input.custom_minimum_size = Vector2(300, 34)
	_search_input.text_changed.connect(func(text: String) -> void:
		_search_text = text.strip_edges()
		_render_sessions()
	)
	filter_row.add_child(_search_input)

	_sort_option = OptionButton.new()
	_sort_option.custom_minimum_size = Vector2(170, 34)
	_sort_option.add_item("Recent", 0)
	_sort_option.add_item("Host", 1)
	_sort_option.add_item("Players", 2)
	_sort_option.set_item_metadata(0, "recent")
	_sort_option.set_item_metadata(1, "host")
	_sort_option.set_item_metadata(2, "players")
	_sort_option.item_selected.connect(func(index: int) -> void:
		_set_sort_mode(str(_sort_option.get_item_metadata(index)))
		_render_sessions()
	)
	filter_row.add_child(_sort_option)

	_empty_label = Label.new()
	_empty_label.text = "No active games"
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
		_rows.remove_child(child)
		child.free()
	_visible_sessions = _filtered_sorted_sessions()
	if _selected_session_id != "" and not _has_visible_session(_selected_session_id):
		_selected_session_id = ""
	_empty_label.visible = _visible_sessions.is_empty()
	_empty_label.text = "No matching games" if _search_text != "" and not _sessions.is_empty() else "No active games"
	for session in _visible_sessions:
		if typeof(session) != TYPE_DICTIONARY:
			continue
		var row: Dictionary = session
		var session_id := str(row.get("session_id", ""))
		var row_box := HBoxContainer.new()
		row_box.add_theme_constant_override("separation", 6)
		row_box.custom_minimum_size = Vector2(460, 38)
		var button := Button.new()
		button.toggle_mode = true
		button.button_pressed = session_id == _selected_session_id
		button.text = "%s  %d/%d  %s" % [
			str(row.get("host_display_name", "Host")),
			int(row.get("connected_count", 0)),
			int(row.get("member_count", 0)),
			str(row.get("world_id", "")),
		]
		button.custom_minimum_size = Vector2(410, 38)
		button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		button.pressed.connect(func() -> void:
			_selected_session_id = session_id
			_render_sessions()
		)
		button.gui_input.connect(func(event: InputEvent) -> void:
			if event is InputEventMouseButton:
				var mouse_event := event as InputEventMouseButton
				if mouse_event.button_index == MOUSE_BUTTON_LEFT and mouse_event.pressed and mouse_event.double_click:
					_join_session(session_id)
		)
		row_box.add_child(button)

		var join_btn := Button.new()
		join_btn.text = "✓"
		join_btn.tooltip_text = "Join game"
		join_btn.custom_minimum_size = Vector2(38, 38)
		join_btn.focus_mode = Control.FOCUS_NONE
		join_btn.pressed.connect(func() -> void:
			_join_session(session_id)
		)
		row_box.add_child(join_btn)
		_rows.add_child(row_box)


func _filtered_sorted_sessions() -> Array:
	var out: Array = []
	var needle := _search_text.strip_edges().to_lower()
	for session in _sessions:
		if typeof(session) != TYPE_DICTIONARY:
			continue
		var row: Dictionary = (session as Dictionary).duplicate(true)
		if needle != "" and not _session_matches(row, needle):
			continue
		out.append(row)
	out.sort_custom(func(a: Dictionary, b: Dictionary) -> bool:
		return _session_sort_less(a, b)
	)
	return out


func _session_matches(row: Dictionary, needle: String) -> bool:
	var haystack := " ".join(PackedStringArray([
		str(row.get("session_id", "")),
		str(row.get("host_display_name", "")),
		str(row.get("world_id", "")),
		str(row.get("mode", "")),
	]))
	return haystack.to_lower().find(needle) >= 0


func _session_sort_less(a: Dictionary, b: Dictionary) -> bool:
	match _sort_mode:
		"host":
			var host_a := str(a.get("host_display_name", "")).to_lower()
			var host_b := str(b.get("host_display_name", "")).to_lower()
			if host_a == host_b:
				return str(a.get("session_id", "")) < str(b.get("session_id", ""))
			return host_a < host_b
		"players":
			var connected_a := int(a.get("connected_count", 0))
			var connected_b := int(b.get("connected_count", 0))
			if connected_a == connected_b:
				return str(a.get("updated_at", "")) > str(b.get("updated_at", ""))
			return connected_a > connected_b
		_:
			var updated_a := str(a.get("updated_at", ""))
			var updated_b := str(b.get("updated_at", ""))
			if updated_a == updated_b:
				return str(a.get("session_id", "")) < str(b.get("session_id", ""))
			return updated_a > updated_b


func _has_visible_session(session_id: String) -> bool:
	for session in _visible_sessions:
		if typeof(session) == TYPE_DICTIONARY and str((session as Dictionary).get("session_id", "")) == session_id:
			return true
	return false


func _set_sort_mode(mode: String) -> void:
	_sort_mode = mode if mode in ["recent", "host", "players"] else "recent"
	if _sort_option == null:
		return
	for i in range(_sort_option.item_count):
		if str(_sort_option.get_item_metadata(i)) == _sort_mode:
			_sort_option.select(i)
			return


func _join_session(session_id: String) -> void:
	if session_id == "":
		return
	_selected_session_id = session_id
	join_requested.emit(session_id)
