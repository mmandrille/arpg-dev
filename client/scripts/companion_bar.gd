class_name CompanionBar
extends Control

const SLOT_SIZE := Vector2(44.0, 62.0)
const ICON_SIZE := Vector2(36.0, 36.0)
const BAR_HEIGHT := 6.0
const MAX_VISIBLE := 8
const TICK_RATE := 10.0

var _row: HBoxContainer
var _companions: Array = []


func _ready() -> void:
	_build()
	set_process(true)


func _process(delta: float) -> void:
	var changed := false
	for companion in _companions:
		if typeof(companion) != TYPE_DICTIONARY:
			continue
		var rec := companion as Dictionary
		var remaining := int(rec.get("remaining_ticks", 0))
		if remaining <= 0:
			continue
		rec["remaining_ticks"] = maxi(0, remaining - int(ceil(delta * TICK_RATE)))
		changed = true
	if changed:
		_render()


func set_companions(next_companions: Array) -> void:
	_companions = next_companions.duplicate(true)
	_companions.sort_custom(func(a: Dictionary, b: Dictionary) -> bool:
		return str(a.get("id", "")) < str(b.get("id", ""))
	)
	_render()


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"count": _companions.size(),
		"companions": _companions.duplicate(true),
		"slot_count": _row.get_child_count() if _row != null else 0,
	}


func _build() -> void:
	if _row != null:
		return
	set_anchors_preset(Control.PRESET_TOP_LEFT)
	position = Vector2(12.0, 40.0)
	mouse_filter = Control.MOUSE_FILTER_IGNORE

	_row = HBoxContainer.new()
	_row.add_theme_constant_override("separation", 6)
	_row.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(_row)
	_render()


func _render() -> void:
	if _row == null:
		return
	for child in _row.get_children():
		child.queue_free()
	visible = not _companions.is_empty()
	if not visible:
		return
	var count := mini(_companions.size(), MAX_VISIBLE)
	for i in range(count):
		_row.add_child(_make_slot(_companions[i] as Dictionary))


func _make_slot(companion: Dictionary) -> Control:
	var slot := PanelContainer.new()
	slot.custom_minimum_size = SLOT_SIZE
	slot.mouse_filter = Control.MOUSE_FILTER_IGNORE
	slot.add_theme_stylebox_override("panel", _slot_style())

	var stack := VBoxContainer.new()
	stack.mouse_filter = Control.MOUSE_FILTER_IGNORE
	stack.add_theme_constant_override("separation", 2)
	slot.add_child(stack)

	var icon := Control.new()
	icon.custom_minimum_size = ICON_SIZE
	icon.mouse_filter = Control.MOUSE_FILTER_IGNORE
	stack.add_child(icon)

	var portrait := ColorRect.new()
	portrait.position = Vector2(4.0, 4.0)
	portrait.size = Vector2(28.0, 28.0)
	portrait.color = _portrait_color(companion)
	portrait.mouse_filter = Control.MOUSE_FILTER_IGNORE
	icon.add_child(portrait)

	var glyph := Label.new()
	glyph.size = ICON_SIZE
	glyph.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	glyph.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	glyph.mouse_filter = Control.MOUSE_FILTER_IGNORE
	glyph.text = _glyph(companion)
	glyph.label_settings = _glyph_settings()
	icon.add_child(glyph)

	stack.add_child(_make_meter(Color(0.10, 0.70, 0.18, 0.94), _hp_ratio(companion)))
	if _has_duration(companion):
		stack.add_child(_make_meter(Color(0.55, 0.76, 1.0, 0.95), _duration_ratio(companion)))
	return slot


func _make_meter(fill_color: Color, ratio: float) -> ColorRect:
	var back := ColorRect.new()
	back.custom_minimum_size = Vector2(ICON_SIZE.x, BAR_HEIGHT)
	back.color = Color(0.08, 0.02, 0.02, 0.82)
	back.mouse_filter = Control.MOUSE_FILTER_IGNORE
	var fill := ColorRect.new()
	fill.size = Vector2(ICON_SIZE.x * clampf(ratio, 0.0, 1.0), BAR_HEIGHT)
	fill.color = fill_color
	fill.mouse_filter = Control.MOUSE_FILTER_IGNORE
	back.add_child(fill)
	return back


func _slot_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color(0.035, 0.030, 0.026, 0.84)
	style.border_color = Color(0.55, 0.46, 0.32, 0.92)
	style.set_border_width_all(1)
	style.corner_radius_top_left = 3
	style.corner_radius_top_right = 3
	style.corner_radius_bottom_left = 3
	style.corner_radius_bottom_right = 3
	style.content_margin_left = 4
	style.content_margin_right = 4
	style.content_margin_top = 4
	style.content_margin_bottom = 4
	return style


func _glyph_settings() -> LabelSettings:
	var settings := LabelSettings.new()
	settings.font_size = 15
	settings.font_color = Color("#f4e7c6")
	settings.outline_size = 3
	settings.outline_color = Color(0.02, 0.01, 0.0, 0.85)
	return settings


func _portrait_color(companion: Dictionary) -> Color:
	var tint := str(companion.get("visual_tint", ""))
	if tint.begins_with("#") and tint.length() == 7:
		return Color(tint)
	var monster_def_id := str(companion.get("monster_def_id", ""))
	if monster_def_id.find("wolf") >= 0:
		return Color("#141414")
	if monster_def_id.find("skeleton") >= 0:
		return Color("#c7b88d")
	if monster_def_id.find("mercenary") >= 0:
		return Color("#53738d")
	return Color("#5c4a36")


func _glyph(companion: Dictionary) -> String:
	var monster_def_id := str(companion.get("monster_def_id", ""))
	if monster_def_id.find("wolf") >= 0:
		return "W"
	if monster_def_id.find("skeleton") >= 0:
		return "S"
	if monster_def_id.find("mercenary") >= 0:
		return "M"
	if monster_def_id == "":
		return "C"
	return monster_def_id.substr(0, 1).to_upper()


func _hp_ratio(companion: Dictionary) -> float:
	var max_hp := maxi(1, int(companion.get("max_hp", 1)))
	var hp := clampi(int(companion.get("hp", max_hp)), 0, max_hp)
	return float(hp) / float(max_hp)


func _has_duration(companion: Dictionary) -> bool:
	return int(companion.get("total_ticks", 0)) > 0


func _duration_ratio(companion: Dictionary) -> float:
	var total := maxi(1, int(companion.get("total_ticks", 1)))
	var remaining := clampi(int(companion.get("remaining_ticks", total)), 0, total)
	return float(remaining) / float(total)
