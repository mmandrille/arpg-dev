extends Control
class_name PauseMenu

signal resume_pressed
signal settings_pressed
signal return_to_menu_pressed
signal exit_pressed


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_STOP
	_build()
	visible = false


func show_pause() -> void:
	_sync_viewport_size()
	visible = true


func hide_pause() -> void:
	visible = false


func _sync_viewport_size() -> void:
	set_anchors_preset(Control.PRESET_TOP_LEFT)
	position = Vector2.ZERO
	size = get_viewport_rect().size


func _build() -> void:
	var bg := ColorRect.new()
	bg.color = Color(0.02, 0.025, 0.03, 0.70)
	bg.set_anchors_preset(Control.PRESET_FULL_RECT)
	add_child(bg)

	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 8)
	box.custom_minimum_size = Vector2(300, 0)
	box.set_anchors_preset(Control.PRESET_CENTER)
	box.offset_left = -150
	box.offset_right = 150
	box.offset_top = -140
	box.offset_bottom = 140
	add_child(box)

	var title := Label.new()
	title.text = "Paused"
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.add_theme_font_size_override("font_size", 26)
	box.add_child(title)

	box.add_child(_button("Resume", resume_pressed.emit))
	box.add_child(_button("Settings", settings_pressed.emit))
	box.add_child(_button("Return to Main Menu", return_to_menu_pressed.emit))
	box.add_child(_button("Exit", exit_pressed.emit))


func _button(text: String, callback: Callable) -> Button:
	var button := Button.new()
	button.text = text
	button.custom_minimum_size = Vector2(260, 42)
	button.pressed.connect(callback)
	return button
