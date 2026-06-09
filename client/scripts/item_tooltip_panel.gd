class_name ItemTooltipPanel
extends PanelContainer

const BODY_FONT_SIZE := 23
const REQUIREMENT_FONT_SIZE := BODY_FONT_SIZE - 1
const ICON_FONT_SIZE := 32
const PREVIEW_SIZE := Vector2(96, 96)
const PREVIEW_GAP := 8
const CONTENT_WIDTH := 320.0
const MAIN_STAT_WIDTH := CONTENT_WIDTH - PREVIEW_SIZE.x - PREVIEW_GAP
const PRICE_WIDTH := 72.0


class ItemPreview:
	extends Control

	var item: Dictionary = {}
	var item_presentations: Dictionary = {}
	var fallback_label: String = ""
	var dimmed: bool = false

	func setup(next_item: Dictionary, next_presentations: Dictionary, next_fallback_label: String = "", next_dimmed: bool = false) -> void:
		item = next_item.duplicate(true)
		item_presentations = next_presentations.duplicate(true)
		fallback_label = next_fallback_label
		dimmed = next_dimmed
		custom_minimum_size = PREVIEW_SIZE
		size_flags_horizontal = Control.SIZE_SHRINK_BEGIN
		size_flags_vertical = Control.SIZE_SHRINK_BEGIN
		queue_redraw()

	func _draw() -> void:
		var rect := Rect2(Vector2.ZERO, size)
		draw_rect(rect, Color("#0a0908"), true)
		draw_rect(rect, Color("#5c4a1f"), false, 1.0)
		if item.is_empty():
			return

		var def_id := str(item.get("item_def_id", ""))
		var presentation: Dictionary = item_presentations.get(def_id, {})
		var icon: Dictionary = presentation.get("icon", {})
		var shape := str(icon.get("shape", "box"))
		var color := Color(str(icon.get("color", "#d8d0bd")))
		var accent := Color(str(icon.get("accent", "#6b5420")))
		if dimmed:
			color = color.darkened(0.35)
			accent = accent.darkened(0.35)
		var center := rect.get_center()
		var min_side = min(rect.size.x, rect.size.y)
		var label := str(icon.get("label", fallback_label if fallback_label != "" else _short_label(def_id)))

		match shape:
			"blade":
				var a := center + Vector2(-min_side * 0.26, min_side * 0.24)
				var b := center + Vector2(min_side * 0.26, -min_side * 0.24)
				draw_line(a, b, color, 10.0, true)
				draw_line(a + Vector2(-8, 8), a + Vector2(10, -10), accent, 7.0, true)
			"bow":
				draw_arc(center, min_side * 0.30, -1.28, 1.28, 26, color, 8.0, true)
				draw_line(center + Vector2(min_side * 0.18, -min_side * 0.30), center + Vector2(min_side * 0.18, min_side * 0.30), accent, 4.0, true)
			"badge", "coin":
				draw_circle(center, min_side * 0.27, color)
				draw_arc(center, min_side * 0.20, 0.0, TAU, 24, accent, 4.0, true)
			"leaf":
				var pts := PackedVector2Array([
					center + Vector2(0, -min_side * 0.34),
					center + Vector2(min_side * 0.28, -min_side * 0.02),
					center + Vector2(0, min_side * 0.31),
					center + Vector2(-min_side * 0.28, -min_side * 0.02),
				])
				draw_colored_polygon(pts, color)
				draw_line(center + Vector2(0, -min_side * 0.25), center + Vector2(0, min_side * 0.25), accent, 4.0, true)
			"potion":
				draw_rect(Rect2(center + Vector2(-min_side * 0.14, -min_side * 0.04), Vector2(min_side * 0.28, min_side * 0.32)), color, true)
				draw_rect(Rect2(center + Vector2(-min_side * 0.09, -min_side * 0.24), Vector2(min_side * 0.18, min_side * 0.18)), accent, true)
			_:
				draw_rect(Rect2(center - Vector2(min_side * 0.23, min_side * 0.23), Vector2(min_side * 0.46, min_side * 0.46)), color, true)

		var font := get_theme_default_font()
		var text_size := font.get_string_size(label, HORIZONTAL_ALIGNMENT_LEFT, -1, ICON_FONT_SIZE)
		draw_string(font, center + Vector2(-text_size.x * 0.5, min_side * 0.36), label, HORIZONTAL_ALIGNMENT_LEFT, -1, ICON_FONT_SIZE, Color("#f4ead8"))

	func _short_label(def_id: String) -> String:
		if def_id == "":
			return "?"
		var parts := def_id.split("_")
		var out := ""
		for part in parts:
			if part.length() > 0:
				out += part.substr(0, 1).to_upper()
		return out.substr(0, 3)


func setup(item: Dictionary, item_presentations: Dictionary, main_lines: Array, requirement_lines: Array, comparison_entries: Array, price: int = -1, affordable: bool = true, fallback_label: String = "") -> void:
	_clear_children(self)
	add_theme_stylebox_override("panel", _tooltip_style())
	var has_item := not item.is_empty()

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 4)
	root.custom_minimum_size = Vector2(CONTENT_WIDTH, 0)
	add_child(root)

	var top_row := HBoxContainer.new()
	top_row.add_theme_constant_override("separation", PREVIEW_GAP)
	root.add_child(top_row)

	var main_stats := VBoxContainer.new()
	main_stats.add_theme_constant_override("separation", 2)
	main_stats.custom_minimum_size = Vector2(MAIN_STAT_WIDTH if has_item else CONTENT_WIDTH, 0)
	main_stats.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	top_row.add_child(main_stats)

	for line in main_lines:
		main_stats.add_child(_tooltip_label(str(line), Color("#e8dcc8"), MAIN_STAT_WIDTH if has_item else CONTENT_WIDTH))

	if has_item:
		var preview := ItemPreview.new()
		preview.setup(item, item_presentations, fallback_label, price >= 0 and not affordable)
		top_row.add_child(preview)

	if not requirement_lines.is_empty():
		root.add_child(_tooltip_spacer(8))
		root.add_child(_tooltip_label("Requirements", Color("#c9a227")))
		for line in requirement_lines:
			root.add_child(_tooltip_label(_entry_text(line), _entry_color(line, Color("#d8c7a6")), CONTENT_WIDTH, REQUIREMENT_FONT_SIZE))
	if not comparison_entries.is_empty():
		root.add_child(_tooltip_spacer(6))
		root.add_child(_tooltip_separator())
		root.add_child(_tooltip_spacer(4))
		for entry in comparison_entries:
			if typeof(entry) != TYPE_DICTIONARY:
				continue
			var rec := entry as Dictionary
			var color: Color = rec.get("color", Color("#d8c7a6"))
			root.add_child(_tooltip_label(str(rec.get("text", "")), color))

	if price >= 0:
		var footer := HBoxContainer.new()
		footer.add_theme_constant_override("separation", 0)
		var spacer := Control.new()
		spacer.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		footer.add_child(spacer)
		var price_label := Label.new()
		price_label.text = str(price)
		price_label.custom_minimum_size = Vector2(PRICE_WIDTH, 0)
		price_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
		price_label.add_theme_color_override("font_color", Color("#f4c84f") if affordable else Color("#ff6f6f"))
		price_label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
		footer.add_child(price_label)
		root.add_child(footer)


func _tooltip_label(text: String, color: Color, width: float = CONTENT_WIDTH, font_size: int = BODY_FONT_SIZE) -> Label:
	var label := Label.new()
	label.text = text
	label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	label.custom_minimum_size = Vector2(width, 0)
	label.add_theme_color_override("font_color", color)
	label.add_theme_font_size_override("font_size", font_size)
	return label


func _tooltip_spacer(height: int) -> Control:
	var spacer := Control.new()
	spacer.custom_minimum_size = Vector2(0, height)
	return spacer


func _tooltip_separator() -> ColorRect:
	var separator := ColorRect.new()
	separator.color = Color("#6b5420")
	separator.custom_minimum_size = Vector2(0, 1)
	separator.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	return separator


func _tooltip_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.07, 0.06, 0.05, 0.97)
	s.border_color = Color("#8b6914")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.content_margin_left = 10
	s.content_margin_top = 8
	s.content_margin_right = 10
	s.content_margin_bottom = 8
	return s


func _clear_children(node: Node) -> void:
	for child in node.get_children():
		child.queue_free()


func _entry_text(value) -> String:
	if typeof(value) == TYPE_DICTIONARY:
		return str((value as Dictionary).get("text", ""))
	return str(value)


func _entry_color(value, fallback: Color) -> Color:
	if typeof(value) != TYPE_DICTIONARY:
		return fallback
	var rec := value as Dictionary
	var color = rec.get("color", fallback)
	if color is Color:
		return color
	return fallback
