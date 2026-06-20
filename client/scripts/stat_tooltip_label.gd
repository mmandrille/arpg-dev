class_name StatTooltipLabel
extends Label

const NORMAL_COLOR := Color("#d8c7a6")
const BOOST_COLOR := Color("#8ee68e")
const PENALTY_COLOR := Color("#ff8b7a")


func _make_custom_tooltip(for_text: String) -> Object:
	var panel := PanelContainer.new()
	panel.mouse_filter = Control.MOUSE_FILTER_IGNORE
	panel.add_theme_stylebox_override("panel", _tooltip_style())
	var label := Label.new()
	label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	label.text = for_text
	label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	label.custom_minimum_size = Vector2(460, 0)
	label.add_theme_color_override("font_color", Color("#f3ead8"))
	label.add_theme_font_size_override("font_size", 16)
	panel.add_child(label)
	return panel


func apply_effective_stat_style(base_value: int, effective_value: int) -> void:
	add_theme_color_override("font_color", _effective_color(base_value, effective_value))
	add_theme_font_size_override("font_size", 24)
	add_theme_constant_override("outline_size", 1)
	add_theme_color_override("font_outline_color", Color(0, 0, 0, 0.7))


func _effective_color(base_value: int, effective_value: int) -> Color:
	if effective_value > base_value:
		return BOOST_COLOR
	if effective_value < base_value:
		return PENALTY_COLOR
	return NORMAL_COLOR


func _tooltip_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.055, 0.048, 0.038, 1.0)
	s.border_color = Color("#b98a22")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.corner_radius_top_left = 4
	s.corner_radius_top_right = 4
	s.corner_radius_bottom_left = 4
	s.corner_radius_bottom_right = 4
	s.content_margin_left = 10
	s.content_margin_right = 10
	s.content_margin_top = 8
	s.content_margin_bottom = 8
	return s
