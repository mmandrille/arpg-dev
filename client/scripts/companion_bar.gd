class_name CompanionBar
extends Control

signal companion_selected(companion: Dictionary)

class CompanionIcon extends Control:
	var companion: Dictionary = {}

	func _init(next_companion: Dictionary = {}) -> void:
		companion = next_companion.duplicate(true)
		custom_minimum_size = ICON_SIZE
		mouse_filter = Control.MOUSE_FILTER_IGNORE

	func _draw() -> void:
		var kind := CompanionBar._icon_kind(companion)
		match kind:
			"revived":
				_draw_skull_icon()
			"wolf":
				_draw_wolf_icon()
			"mercenary":
				_draw_mercenary_icon()
			_:
				_draw_generic_icon()

	func _draw_skull_icon() -> void:
		draw_rect(Rect2(Vector2(4, 4), Vector2(28, 28)), Color("#191512"), true)
		draw_circle(Vector2(18, 15), 10.0, Color("#d2c9aa"))
		draw_rect(Rect2(Vector2(11, 22), Vector2(14, 7)), Color("#d2c9aa"), true)
		draw_circle(Vector2(14, 15), 2.6, Color("#17110e"))
		draw_circle(Vector2(22, 15), 2.6, Color("#17110e"))
		draw_polygon(PackedVector2Array([Vector2(18, 17), Vector2(15.8, 21), Vector2(20.2, 21)]), PackedColorArray([Color("#17110e")]))
		for x in [13.0, 17.0, 21.0]:
			draw_line(Vector2(x, 23), Vector2(x, 28), Color("#6d604e"), 1.0)

	func _draw_wolf_icon() -> void:
		draw_rect(Rect2(Vector2(4, 4), Vector2(28, 28)), Color("#131614"), true)
		var fur := Color("#d7d0bf")
		var shadow := Color("#151412")
		draw_polygon(PackedVector2Array([
			Vector2(9, 14), Vector2(16, 7), Vector2(28, 11), Vector2(31, 17),
			Vector2(23, 22), Vector2(15, 25), Vector2(8, 20)
		]), PackedColorArray([fur]))
		draw_polygon(PackedVector2Array([Vector2(14, 8), Vector2(18, 3), Vector2(21, 11)]), PackedColorArray([fur]))
		draw_polygon(PackedVector2Array([Vector2(21, 11), Vector2(31, 13), Vector2(28, 17)]), PackedColorArray([Color("#f0ead8")]))
		draw_circle(Vector2(23.5, 13.8), 1.4, shadow)
		draw_line(Vector2(27, 18), Vector2(31, 18), shadow, 1.5)
		draw_line(Vector2(12, 20), Vector2(8, 26), fur, 2.0)

	func _draw_mercenary_icon() -> void:
		draw_rect(Rect2(Vector2(4, 4), Vector2(28, 28)), Color("#17202a"), true)
		var steel := Color("#d7dde5")
		var hilt := Color("#9a6a2f")
		draw_line(Vector2(12, 27), Vector2(26, 9), steel, 3.0)
		draw_line(Vector2(24, 8), Vector2(28, 6), steel, 2.0)
		draw_line(Vector2(14, 23), Vector2(20, 28), hilt, 3.0)
		draw_line(Vector2(11, 23), Vector2(17, 29), Color("#e0b45b"), 1.5)
		draw_circle(Vector2(12, 12), 5.2, Color("#d7a72f"))
		draw_arc(Vector2(12, 12), 3.2, 0.0, TAU, 18, Color("#f5dc75"), 1.2)

	func _draw_generic_icon() -> void:
		draw_rect(Rect2(Vector2(4, 4), Vector2(28, 28)), CompanionBar._portrait_color(companion), true)
		draw_circle(Vector2(18, 18), 8.0, Color("#f4e7c6"))

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
	var debug_companions: Array = []
	for companion in _companions:
		if typeof(companion) != TYPE_DICTIONARY:
			continue
		var rec := (companion as Dictionary).duplicate(true)
		rec["icon_kind"] = _icon_kind(rec)
		debug_companions.append(rec)
	return {
		"visible": visible,
		"count": _companions.size(),
		"companions": debug_companions,
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
	slot.mouse_filter = Control.MOUSE_FILTER_STOP
	slot.add_theme_stylebox_override("panel", _slot_style())
	slot.gui_input.connect(func(event: InputEvent) -> void:
		if event is InputEventMouseButton and event.button_index == MOUSE_BUTTON_LEFT and event.pressed:
			companion_selected.emit(companion.duplicate(true))
	)

	var stack := VBoxContainer.new()
	stack.mouse_filter = Control.MOUSE_FILTER_IGNORE
	stack.add_theme_constant_override("separation", 2)
	slot.add_child(stack)

	var icon := Control.new()
	icon.custom_minimum_size = ICON_SIZE
	icon.mouse_filter = Control.MOUSE_FILTER_IGNORE
	stack.add_child(icon)

	icon.add_child(CompanionIcon.new(companion))

	stack.add_child(_make_meter(Color(0.10, 0.70, 0.18, 0.94), _hp_ratio(companion)))
	if _has_duration(companion):
		stack.add_child(_make_meter(Color(0.55, 0.76, 1.0, 0.95), _duration_ratio(companion)))
	return slot


func bot_click_slot(index: int = 0) -> void:
	if index < 0 or index >= _companions.size():
		return
	var companion: Dictionary = _companions[index]
	companion_selected.emit(companion.duplicate(true))


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


static func _portrait_color(companion: Dictionary) -> Color:
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


static func _icon_kind(companion: Dictionary) -> String:
	if int(companion.get("total_ticks", 0)) > 0:
		return "revived"
	var monster_def_id := str(companion.get("monster_def_id", ""))
	if monster_def_id.find("wolf") >= 0:
		return "wolf"
	if monster_def_id.find("mercenary") >= 0:
		return "mercenary"
	return "generic"


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
