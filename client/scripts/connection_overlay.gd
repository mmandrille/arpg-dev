class_name ConnectionOverlay
extends Control

signal cancel_pressed
signal return_to_menu_pressed

var _title: Label
var _subtitle: Label
var _detail: Label
var _cancel_button: Button
var _menu_button: Button
var _button_row: HBoxContainer


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_STOP
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	visible = false
	_build()


func show_reconnecting(attempt: int, max_attempts: int) -> void:
	_title.text = "Reconnecting..."
	_subtitle.text = "Restoring your session"
	_detail.text = "Attempt %d of %d" % [attempt, max_attempts]
	_cancel_button.visible = attempt >= 1
	_menu_button.visible = false
	visible = true
	move_to_front()


func show_failed() -> void:
	_title.text = "Connection lost"
	_subtitle.text = "Could not reach the game server"
	_detail.text = "Your session may still be running on the server"
	_cancel_button.visible = false
	_menu_button.visible = true
	visible = true
	move_to_front()


func hide_overlay() -> void:
	visible = false


func _build() -> void:
	var bg := ColorRect.new()
	bg.color = Color(0.03, 0.035, 0.05, 0.94)
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(bg)

	var center := VBoxContainer.new()
	center.add_theme_constant_override("separation", 14)
	center.set_anchors_preset(Control.PRESET_CENTER)
	center.offset_left = -280
	center.offset_right = 280
	center.offset_top = -110
	center.offset_bottom = 110
	add_child(center)

	_title = Label.new()
	_title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_title.add_theme_font_size_override("font_size", 34)
	_title.add_theme_color_override("font_color", Color("#f1efe4"))
	center.add_child(_title)

	_subtitle = Label.new()
	_subtitle.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_subtitle.add_theme_font_size_override("font_size", 18)
	_subtitle.add_theme_color_override("font_color", Color("#c8c4b6"))
	center.add_child(_subtitle)

	_detail = Label.new()
	_detail.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_detail.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_detail.add_theme_font_size_override("font_size", 14)
	_detail.add_theme_color_override("font_color", Color("#9a9688"))
	center.add_child(_detail)

	_button_row = HBoxContainer.new()
	_button_row.alignment = BoxContainer.ALIGNMENT_CENTER
	_button_row.add_theme_constant_override("separation", 12)
	center.add_child(_button_row)

	_cancel_button = Button.new()
	_cancel_button.text = "Cancel"
	_cancel_button.visible = false
	_cancel_button.pressed.connect(func() -> void: cancel_pressed.emit())
	_button_row.add_child(_cancel_button)

	_menu_button = Button.new()
	_menu_button.text = "Return to menu"
	_menu_button.visible = false
	_menu_button.pressed.connect(func() -> void: return_to_menu_pressed.emit())
	_button_row.add_child(_menu_button)
