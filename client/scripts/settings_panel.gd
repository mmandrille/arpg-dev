extends Control
class_name SettingsPanel

signal back_requested
signal size_selected(label: String)
signal floating_combat_text_toggled(enabled: bool)

var _buttons: Dictionary = {}
var _selected_label: String = ""
var _floating_toggle: CheckButton


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_STOP
	_build()
	visible = false


func show_settings(selected_label: String, floating_combat_text_enabled: bool = true) -> void:
	_sync_viewport_size()
	visible = true
	set_selected_size_label(selected_label)
	set_floating_combat_text_enabled(floating_combat_text_enabled)


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


func _build() -> void:
	var bg := ColorRect.new()
	bg.color = Color(0.03, 0.035, 0.04, 0.88)
	bg.set_anchors_preset(Control.PRESET_FULL_RECT)
	add_child(bg)

	var panel := PanelContainer.new()
	panel.custom_minimum_size = Vector2(420, 340)
	panel.set_anchors_preset(Control.PRESET_CENTER)
	panel.offset_left = -210
	panel.offset_right = 210
	panel.offset_top = -170
	panel.offset_bottom = 170
	add_child(panel)

	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 8)
	panel.add_child(box)

	var title := Label.new()
	title.text = "Settings"
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.add_theme_font_size_override("font_size", 30)
	box.add_child(title)

	for label in ["1280x720", "1600x900", "1920x1080"]:
		var button := Button.new()
		button.text = label
		button.custom_minimum_size = Vector2(320, 46)
		button.add_theme_font_size_override("font_size", 17)
		button.pressed.connect(func() -> void:
			size_selected.emit(label)
		)
		_buttons[label] = button
		box.add_child(button)

	_floating_toggle = CheckButton.new()
	_floating_toggle.text = "Floating combat text"
	_floating_toggle.button_pressed = true
	_floating_toggle.custom_minimum_size = Vector2(320, 42)
	_floating_toggle.add_theme_font_size_override("font_size", 17)
	_floating_toggle.toggled.connect(func(enabled: bool) -> void:
		floating_combat_text_toggled.emit(enabled)
	)
	box.add_child(_floating_toggle)

	var back := Button.new()
	back.text = "Back"
	back.custom_minimum_size = Vector2(320, 44)
	back.add_theme_font_size_override("font_size", 17)
	back.pressed.connect(back_requested.emit)
	box.add_child(back)
