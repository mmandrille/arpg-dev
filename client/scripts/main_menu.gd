extends Control
class_name MainMenu

signal continue_pressed
signal new_game_pressed
signal multiplayer_pressed
signal settings_pressed
signal exit_pressed


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_STOP
	_build()
	visible = false


func show_menu() -> void:
	_sync_viewport_size()
	visible = true


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)


func _build() -> void:
	var bg := ColorRect.new()
	bg.color = Color(0.05, 0.055, 0.06, 0.94)
	bg.set_anchors_preset(Control.PRESET_FULL_RECT)
	add_child(bg)

	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 10)
	box.custom_minimum_size = Vector2(340, 0)
	box.set_anchors_preset(Control.PRESET_CENTER)
	box.offset_left = -170
	box.offset_right = 170
	box.offset_top = -160
	box.offset_bottom = 160
	add_child(box)

	var title := Label.new()
	title.text = "ARPG Dev"
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.add_theme_font_size_override("font_size", 34)
	title.add_theme_color_override("font_color", Color("#f1efe4"))
	box.add_child(title)

	box.add_child(_button("Continue", continue_pressed.emit))
	box.add_child(_button("New Game", new_game_pressed.emit))
	box.add_child(_button("Multiplayer", multiplayer_pressed.emit))
	box.add_child(_button("Settings", settings_pressed.emit))
	box.add_child(_button("Exit", exit_pressed.emit))


func _button(text: String, callback: Callable) -> Button:
	var button := Button.new()
	button.text = text
	button.custom_minimum_size = Vector2(260, 44)
	button.pressed.connect(callback)
	return button
