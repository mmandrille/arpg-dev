extends Control
class_name SettingsPanel

signal back_requested
signal size_selected(label: String)
signal floating_combat_text_toggled(enabled: bool)
signal status_text_toggled(enabled: bool)
signal create_game_session_type_selected(session_type: String)

var _buttons: Dictionary = {}
var _session_type_buttons: Dictionary = {}
var _selected_label: String = ""
var _selected_session_type: String = "coop"
var _floating_toggle: CheckButton
var _status_text_toggle: CheckButton


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_STOP
	_build()
	visible = false


func show_settings(selected_label: String, floating_combat_text_enabled: bool = true, status_text_enabled: bool = true, create_game_session_type: String = "coop") -> void:
	_sync_viewport_size()
	visible = true
	set_selected_size_label(selected_label)
	set_floating_combat_text_enabled(floating_combat_text_enabled)
	set_status_text_enabled(status_text_enabled)
	set_create_game_session_type(create_game_session_type)


func hide_panel() -> void:
	visible = false


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)


func set_selected_size_label(label: String) -> void:
	_selected_label = label
	for key in _buttons.keys():
		var button: Button = _buttons[key]
		button.disabled = str(key) == _selected_label


func set_floating_combat_text_enabled(enabled: bool) -> void:
	if _floating_toggle == null:
		return
	_floating_toggle.button_pressed = enabled


func set_status_text_enabled(enabled: bool) -> void:
	if _status_text_toggle == null:
		return
	_status_text_toggle.button_pressed = enabled


func set_create_game_session_type(session_type: String) -> void:
	_selected_session_type = _normalize_session_type(session_type)
	for key in _session_type_buttons.keys():
		var button: Button = _session_type_buttons[key]
		button.disabled = str(key) == _selected_session_type


func _build() -> void:
	var bg := ColorRect.new()
	bg.color = Color(0.03, 0.035, 0.04, 0.88)
	bg.set_anchors_preset(Control.PRESET_FULL_RECT)
	add_child(bg)

	var panel := PanelContainer.new()
	panel.custom_minimum_size = Vector2(460, 520)
	panel.set_anchors_preset(Control.PRESET_CENTER)
	panel.offset_left = -230
	panel.offset_right = 230
	panel.offset_top = -260
	panel.offset_bottom = 260
	add_child(panel)

	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 8)
	panel.add_child(box)

	var title := Label.new()
	title.text = "Settings"
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.add_theme_font_size_override("font_size", 45)
	box.add_child(title)

	for label in ["1280x720", "1600x900", "1920x1080"]:
		var button := Button.new()
		button.text = label
		button.custom_minimum_size = Vector2(320, 46)
		button.add_theme_font_size_override("font_size", 39)
		button.pressed.connect(func() -> void:
			size_selected.emit(label)
		)
		_buttons[label] = button
		box.add_child(button)

	var session_type_label := Label.new()
	session_type_label.text = "Create Game Type"
	session_type_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	session_type_label.add_theme_font_size_override("font_size", 30)
	box.add_child(session_type_label)

	var session_type_row := HBoxContainer.new()
	session_type_row.add_theme_constant_override("separation", 8)
	box.add_child(session_type_row)
	_add_session_type_button(session_type_row, "Co-op", "coop")
	_add_session_type_button(session_type_row, "Solo", "solo")

	_floating_toggle = CheckButton.new()
	_floating_toggle.text = "Floating combat text"
	_floating_toggle.button_pressed = true
	_floating_toggle.custom_minimum_size = Vector2(320, 42)
	_floating_toggle.add_theme_font_size_override("font_size", 39)
	_floating_toggle.toggled.connect(func(enabled: bool) -> void:
		floating_combat_text_toggled.emit(enabled)
	)
	box.add_child(_floating_toggle)

	_status_text_toggle = CheckButton.new()
	_status_text_toggle.text = "status-text"
	_status_text_toggle.button_pressed = true
	_status_text_toggle.custom_minimum_size = Vector2(320, 42)
	_status_text_toggle.add_theme_font_size_override("font_size", 39)
	_status_text_toggle.toggled.connect(func(enabled: bool) -> void:
		status_text_toggled.emit(enabled)
	)
	box.add_child(_status_text_toggle)

	var back := Button.new()
	back.text = "Back"
	back.custom_minimum_size = Vector2(320, 44)
	back.add_theme_font_size_override("font_size", 39)
	back.pressed.connect(back_requested.emit)
	box.add_child(back)


func _add_session_type_button(parent: Control, label: String, session_type: String) -> void:
	var button := Button.new()
	button.text = label
	button.custom_minimum_size = Vector2(154, 42)
	button.add_theme_font_size_override("font_size", 34)
	button.pressed.connect(func() -> void:
		create_game_session_type_selected.emit(session_type)
	)
	_session_type_buttons[session_type] = button
	parent.add_child(button)


func _normalize_session_type(session_type: String) -> String:
	var normalized := session_type.strip_edges().to_lower()
	if normalized == "solo":
		return "solo"
	return "coop"
