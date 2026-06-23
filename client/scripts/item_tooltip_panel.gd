class_name ItemTooltipPanel
extends PanelContainer

const BODY_FONT_SIZE := 23
const ClientConstantsScript := preload("res://scripts/client_constants.gd")
const REQUIREMENT_FONT_SIZE := BODY_FONT_SIZE - 1
const ICON_FONT_SIZE := 32
const ItemIconDrawerScript := preload("res://scripts/item_icon_drawer.gd")
const TooltipMouseGuardScript := preload("res://scripts/tooltip_mouse_guard.gd")
const PREVIEW_SIZE := Vector2(96, 96)
const PREVIEW_GAP := 8
const CONTENT_WIDTH := 360.0
const MAIN_STAT_WIDTH := CONTENT_WIDTH - PREVIEW_SIZE.x - PREVIEW_GAP
const PRICE_WIDTH := 132.0
const LEVEL_WIDTH := 132.0
const FOOTER_TOP_GAP := 10


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
		mouse_filter = Control.MOUSE_FILTER_IGNORE
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
		var label := str(icon.get("label", fallback_label if fallback_label != "" else _short_label(def_id)))
		ItemIconDrawerScript.draw(self, rect, icon, label, dimmed, 0.36, ICON_FONT_SIZE)

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
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	var rarity := str(item.get("rarity", "common")).to_lower()
	add_theme_stylebox_override("panel", _tooltip_style(rarity))
	var has_item := not item.is_empty()

	var root := VBoxContainer.new()
	root.mouse_filter = Control.MOUSE_FILTER_IGNORE
	root.add_theme_constant_override("separation", 4)
	root.custom_minimum_size = Vector2(CONTENT_WIDTH, 0)
	add_child(root)

	var top_row := HBoxContainer.new()
	top_row.mouse_filter = Control.MOUSE_FILTER_IGNORE
	top_row.add_theme_constant_override("separation", PREVIEW_GAP)
	root.add_child(top_row)

	var main_stats := VBoxContainer.new()
	main_stats.mouse_filter = Control.MOUSE_FILTER_IGNORE
	main_stats.add_theme_constant_override("separation", 2)
	main_stats.custom_minimum_size = Vector2(MAIN_STAT_WIDTH if has_item else CONTENT_WIDTH, 0)
	main_stats.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	top_row.add_child(main_stats)

	for line in main_lines:
		main_stats.add_child(_tooltip_label(_entry_text(line), _entry_color(line, Color("#e8dcc8")), MAIN_STAT_WIDTH if has_item else CONTENT_WIDTH, _entry_font_size(line, BODY_FONT_SIZE)))

	if has_item:
		var preview := ItemPreview.new()
		preview.setup(item, item_presentations, fallback_label, price >= 0 and not affordable)
		top_row.add_child(preview)

	var level_text := _item_level_text(item)
	var visible_requirement_lines := _visible_requirement_lines(requirement_lines, not level_text.begins_with("Item level"))
	if not visible_requirement_lines.is_empty():
		root.add_child(_tooltip_spacer(8))
		root.add_child(_tooltip_label("Requirements", Color("#c9a227")))
		for line in visible_requirement_lines:
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

	if price >= 0 or level_text != "":
		root.add_child(_tooltip_spacer(FOOTER_TOP_GAP))
		var footer := HBoxContainer.new()
		footer.name = "GoldValueFooter"
		footer.mouse_filter = Control.MOUSE_FILTER_IGNORE
		footer.add_theme_constant_override("separation", 0)
		footer.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		var level_label := Label.new()
		level_label.name = "ItemLevelLabel"
		level_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
		level_label.text = level_text
		level_label.custom_minimum_size = Vector2(LEVEL_WIDTH, 0)
		level_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_LEFT
		level_label.add_theme_color_override("font_color", Color("#9a9489"))
		level_label.add_theme_font_size_override("font_size", REQUIREMENT_FONT_SIZE)
		footer.add_child(level_label)
		var spacer := Control.new()
		spacer.mouse_filter = Control.MOUSE_FILTER_IGNORE
		spacer.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		footer.add_child(spacer)
		if price >= 0:
			var price_label := Label.new()
			price_label.name = "GoldValueLabel"
			price_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
			price_label.text = "%d gold" % price
			price_label.custom_minimum_size = Vector2(PRICE_WIDTH, 0)
			price_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
			price_label.add_theme_color_override("font_color", Color("#f4c84f") if affordable else Color("#ff6f6f"))
			price_label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
			footer.add_child(price_label)
		root.add_child(footer)
	TooltipMouseGuardScript.ignore_mouse(self)


func debug_gold_value_text() -> String:
	var label := find_child("GoldValueLabel", true, false) as Label
	if label == null:
		return ""
	return label.text


func debug_item_level_text() -> String:
	var label := find_child("ItemLevelLabel", true, false) as Label
	if label == null:
		return ""
	return label.text


func debug_requirement_texts() -> Array:
	var out: Array = []
	var in_requirements := false
	for child in find_children("*", "Label", true, false):
		var label := child as Label
		if label == null:
			continue
		if label.text == "Requirements":
			in_requirements = true
			continue
		if label.name == "GoldValueLabel" or label.name == "ItemLevelLabel":
			in_requirements = false
			continue
		if in_requirements and label.text != "":
			out.append(label.text)
	return out


static func border_width_for_rarity(rarity: String) -> int:
	var temp := ItemTooltipPanel.new()
	var style := temp._tooltip_style(rarity)

	return style.border_width_left


func debug_border_width() -> int:
	var style := get_theme_stylebox("panel") as StyleBoxFlat
	if style == null:
		return 0

	return style.border_width_left


func debug_first_main_line_color() -> String:
	var label := _first_main_line_label()
	if label == null:
		return ""
	return label.get_theme_color("font_color").to_html(false)


func debug_main_line_font_sizes() -> Array:
	var out: Array = []
	for child in find_children("*", "Label", true, false):
		var label := child as Label
		if label != null and label.name != "GoldValueLabel" and label.text != "":
			out.append({"text": label.text, "font_size": label.get_theme_font_size("font_size")})
	return out


func _first_main_line_label() -> Label:
	for child in find_children("*", "Label", true, false):
		var label := child as Label
		if label != null and label.name != "GoldValueLabel" and label.text != "":
			return label
	return null


func _tooltip_label(text: String, color: Color, width: float = CONTENT_WIDTH, font_size: int = BODY_FONT_SIZE) -> Label:
	var label := Label.new()
	label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	label.text = text
	label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	label.custom_minimum_size = Vector2(width, 0)
	label.add_theme_color_override("font_color", color)
	label.add_theme_font_size_override("font_size", font_size)
	return label


func _tooltip_spacer(height: int) -> Control:
	var spacer := Control.new()
	spacer.mouse_filter = Control.MOUSE_FILTER_IGNORE
	spacer.custom_minimum_size = Vector2(0, height)
	return spacer


func _tooltip_separator() -> ColorRect:
	var separator := ColorRect.new()
	separator.mouse_filter = Control.MOUSE_FILTER_IGNORE
	separator.color = Color("#6b5420")
	separator.custom_minimum_size = Vector2(0, 1)
	separator.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	return separator


func _tooltip_style(rarity: String = "common") -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.07, 0.06, 0.05, 0.97)
	var border: Color = ClientConstantsScript.LOOT_LABEL_RARITY_COLORS.get(rarity.to_lower(), Color("#8b6914"))
	s.border_color = border
	var border_width: int = 2 if rarity.to_lower() in ["magic", "rare", "unique"] else 1
	s.border_width_left = border_width
	s.border_width_top = border_width
	s.border_width_right = border_width
	s.border_width_bottom = border_width
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


func _entry_font_size(value, fallback: int) -> int:
	if typeof(value) != TYPE_DICTIONARY:
		return fallback
	var rec := value as Dictionary
	return int(rec.get("font_size", fallback))


func _visible_requirement_lines(requirement_lines: Array, hide_level_requirement: bool = true) -> Array:
	var out: Array = []
	for line in requirement_lines:
		if hide_level_requirement and _is_level_line(_entry_text(line)):
			continue
		out.append(line)
	return out


func _item_level_text(item: Dictionary) -> String:
	if item.is_empty():
		return ""
	var item_level = _explicit_item_level(item)
	if item_level != null:
		return "Item level %s" % _format_level_value(item_level)
	var requirements = item.get("requirements", {})
	if typeof(requirements) == TYPE_DICTIONARY:
		var req := requirements as Dictionary
		if req.has("level"):
			return "Level %s" % _format_level_value(req.get("level", 0))
	var statuses = item.get("requirement_status", [])
	if typeof(statuses) == TYPE_ARRAY:
		for status in statuses:
			if typeof(status) != TYPE_DICTIONARY:
				continue
			var rec := status as Dictionary
			if str(rec.get("stat", "")) == "level" and int(rec.get("required", 0)) > 0:
				return "Level %s" % _format_level_value(rec.get("required", 0))
	return ""


func _explicit_item_level(item: Dictionary):
	if item.has("item_level"):
		return item.get("item_level", 0)
	var stats = item.get("stats", {})
	if typeof(stats) == TYPE_DICTIONARY and (stats as Dictionary).has("item_level"):
		return (stats as Dictionary).get("item_level", 0)
	var rolled_stats = item.get("rolled_stats", {})
	if typeof(rolled_stats) == TYPE_DICTIONARY and (rolled_stats as Dictionary).has("item_level"):
		return (rolled_stats as Dictionary).get("item_level", 0)
	return null


func _format_level_value(value: Variant) -> String:
	if typeof(value) == TYPE_FLOAT:
		var number := float(value)
		if is_equal_approx(number, float(int(number))):
			return str(int(number))
	return str(value)


func _is_level_line(text: String) -> bool:
	return text.strip_edges().to_lower().begins_with("level ")
