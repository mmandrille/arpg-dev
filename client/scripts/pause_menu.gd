extends Control
class_name PauseMenu

const TextCatalogScript := preload("res://scripts/text_catalog.gd")

signal resume_pressed
signal settings_pressed
signal return_to_menu_pressed
signal exit_pressed

var _title: Label
var _buttons_by_key: Dictionary = {}

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

	_title = Label.new()
	_title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_title.add_theme_font_size_override("font_size", 39)
	box.add_child(_title)

	box.add_child(_button("resume", resume_pressed.emit))
	box.add_child(_button("settings", settings_pressed.emit))
	box.add_child(_button("return_to_main_menu", return_to_menu_pressed.emit))
	box.add_child(_button("exit", exit_pressed.emit))
	refresh_texts()


func _button(key: String, callback: Callable) -> Button:
	var button := Button.new()
	_buttons_by_key[key] = button
	button.custom_minimum_size = Vector2(260, 42)
	button.pressed.connect(callback)
	return button


func refresh_texts() -> void:
	if _title != null:
		_title.text = TextCatalogScript.get_text("pause.title", "Paused")
	_set_button_text("resume", TextCatalogScript.get_text("pause.resume", "Resume"))
	_set_button_text("settings", TextCatalogScript.get_text("menu.settings", "Settings"))
	_set_button_text("return_to_main_menu", TextCatalogScript.get_text("pause.return_to_main_menu", "Return to Main Menu"))
	_set_button_text("exit", TextCatalogScript.get_text("menu.exit", "Exit"))


func _set_button_text(key: String, text: String) -> void:
	var button: Button = _buttons_by_key.get(key, null)
	if button != null:
		button.text = text
