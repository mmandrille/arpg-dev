extends Control
class_name PauseMenu

const TextCatalogScript := preload("res://scripts/text_catalog.gd")

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
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)


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
	title.text = TextCatalogScript.get_text("pause.title", "Paused")
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.add_theme_font_size_override("font_size", 39)
	box.add_child(title)

	box.add_child(_button(TextCatalogScript.get_text("pause.resume", "Resume"), resume_pressed.emit))
	box.add_child(_button(TextCatalogScript.get_text("menu.settings", "Settings"), settings_pressed.emit))
	box.add_child(_button(TextCatalogScript.get_text("pause.return_to_main_menu", "Return to Main Menu"), return_to_menu_pressed.emit))
	box.add_child(_button(TextCatalogScript.get_text("menu.exit", "Exit"), exit_pressed.emit))


func _button(text: String, callback: Callable) -> Button:
	var button := Button.new()
	button.text = text
	button.custom_minimum_size = Vector2(260, 42)
	button.pressed.connect(callback)
	return button
