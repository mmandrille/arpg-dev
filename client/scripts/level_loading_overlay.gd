class_name LevelLoadingOverlay
extends Control

var _title: Label
var _subtitle: Label
var _detail: Label


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_STOP
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	visible = false
	_build()


func show_for_level(level: int, level_name: String, traveling: bool = false) -> void:
	_title.text = "Traveling..." if traveling else "Generating dungeon..."
	_subtitle.text = level_name if level_name != "" else "Level %d" % abs(level)
	_detail.text = "Preparing walls, monsters, and loot"
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
	center.offset_left = -260
	center.offset_right = 260
	center.offset_top = -90
	center.offset_bottom = 90
	add_child(center)

	_title = Label.new()
	_title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_title.add_theme_font_size_override("font_size", 34)
	_title.add_theme_color_override("font_color", Color("#f1efe4"))
	center.add_child(_title)

	_subtitle = Label.new()
	_subtitle.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_subtitle.add_theme_font_size_override("font_size", 22)
	_subtitle.add_theme_color_override("font_color", Color("#d0a84f"))
	center.add_child(_subtitle)

	_detail = Label.new()
	_detail.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_detail.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_detail.add_theme_font_size_override("font_size", 16)
	_detail.add_theme_color_override("font_color", Color("#9a9488"))
	center.add_child(_detail)
