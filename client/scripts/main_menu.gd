extends Control
class_name MainMenu

const TextCatalogScript := preload("res://scripts/text_catalog.gd")

signal create_game_pressed
signal join_game_pressed
signal settings_pressed
signal exit_pressed

var _button_labels: Array = []
var _buttons_by_action: Dictionary = {}
var _title: Label


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

	_title = Label.new()
	_title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_title.add_theme_font_size_override("font_size", 51)
	_title.add_theme_color_override("font_color", Color("#f1efe4"))
	box.add_child(_title)

	box.add_child(_button(TextCatalogScript.get_text("menu.create_game", "Create Game"), create_game_pressed.emit, "create_game"))
	box.add_child(_button(TextCatalogScript.get_text("menu.join_game", "Join Game"), join_game_pressed.emit, "join_game"))
	box.add_child(_button(TextCatalogScript.get_text("menu.settings", "Settings"), settings_pressed.emit))
	box.add_child(_button(TextCatalogScript.get_text("menu.exit", "Exit"), exit_pressed.emit))
	refresh_texts()


func _button(text: String, callback: Callable, action: String = "") -> Button:
	var button := Button.new()
	button.text = text
	if action != "":
		button.set_meta("action", action)
		_buttons_by_action[action] = button
	_button_labels.append(text)
	button.custom_minimum_size = Vector2(260, 44)
	button.pressed.connect(callback)
	return button


func button_labels() -> Array:
	return _button_labels.duplicate()


func available_actions() -> Array:
	return ["create_game", "join_game", "settings", "exit"]


func refresh_texts() -> void:
	if _title != null:
		_title.text = TextCatalogScript.get_text("app.title", "ARPG Dev")
	_set_button_text("create_game", TextCatalogScript.get_text("menu.create_game", "Create Game"))
	_set_button_text("join_game", TextCatalogScript.get_text("menu.join_game", "Join Game"))
	_set_button_text("settings", TextCatalogScript.get_text("menu.settings", "Settings"))
	_set_button_text("exit", TextCatalogScript.get_text("menu.exit", "Exit"))
	_button_labels = [
		TextCatalogScript.get_text("menu.create_game", "Create Game"),
		TextCatalogScript.get_text("menu.join_game", "Join Game"),
		TextCatalogScript.get_text("menu.settings", "Settings"),
		TextCatalogScript.get_text("menu.exit", "Exit"),
	]


func _set_button_text(action: String, text: String) -> void:
	var button: Button = _buttons_by_action.get(action, null)
	if button != null:
		button.text = text
