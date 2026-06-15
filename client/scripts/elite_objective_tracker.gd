class_name EliteObjectiveTracker
extends PanelContainer

var _title_label: Label
var _detail_label: Label
var _state: Dictionary = {"visible": false, "status": "hidden", "remaining_leaders": 0}


func _ready() -> void:
	_build()


func set_state(state: Dictionary) -> void:
	_state = state.duplicate(true)
	if _title_label == null:
		_build()
	visible = bool(_state.get("visible", false))
	var status := str(_state.get("status", "hidden"))
	var remaining := int(_state.get("remaining_leaders", 0))
	if status == "active":
		_title_label.text = "Elite Objective"
		_detail_label.text = "Defeat elite leaders: %d remaining" % remaining
	elif status == "claim":
		_title_label.text = "Elite Objective"
		_detail_label.text = "Claim the marked reward chest"
	elif status == "complete":
		_title_label.text = "Elite Objective Complete"
		_detail_label.text = "Reward chest opened"
	else:
		_title_label.text = ""
		_detail_label.text = ""


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"status": str(_state.get("status", "hidden")),
		"remaining_leaders": int(_state.get("remaining_leaders", 0)),
		"title": _title_label.text if _title_label != null else "",
		"detail": _detail_label.text if _detail_label != null else "",
	}


func _build() -> void:
	if _title_label != null:
		return
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	set_anchors_preset(Control.PRESET_TOP_LEFT)
	offset_left = 18
	offset_top = 132
	custom_minimum_size = Vector2(260, 60)
	add_theme_stylebox_override("panel", _panel_style())
	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 2)
	add_child(root)
	_title_label = Label.new()
	_title_label.add_theme_font_size_override("font_size", 16)
	_title_label.add_theme_color_override("font_color", Color("#f0dfbb"))
	root.add_child(_title_label)
	_detail_label = Label.new()
	_detail_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_detail_label.add_theme_font_size_override("font_size", 13)
	_detail_label.add_theme_color_override("font_color", Color("#79b8ff"))
	root.add_child(_detail_label)
	set_state(_state)


func _panel_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color(0.035, 0.032, 0.026, 0.86)
	style.border_color = Color("#3d3322")
	style.border_width_left = 1
	style.border_width_top = 1
	style.border_width_right = 1
	style.border_width_bottom = 1
	style.content_margin_left = 10
	style.content_margin_top = 7
	style.content_margin_right = 10
	style.content_margin_bottom = 7
	return style
