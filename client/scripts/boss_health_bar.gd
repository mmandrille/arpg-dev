extends Control
class_name BossHealthBar

const PANEL_WIDTH := 390.0
const PANEL_TOP := 58.0
const BAR_SIZE := Vector2(360.0, 14.0)
const PHASE_BAR_SIZE := Vector2(360.0, 6.0)

var _panel: PanelContainer
var _title_label: Label
var _hp_label: Label
var _phase_label: Label
var _phase_fill: ColorRect
var _fill: ColorRect
var _boss_id: String = ""
var _boss_template_id: String = ""
var _title: String = ""
var _hp: int = 0
var _max_hp: int = 1
var _ratio: float = 0.0
var _phase_kind: String = ""
var _pattern_id: String = ""
var _phase_index: int = -1
var _duration_ticks: int = 0
var _remaining_ticks: int = 0
var _phase_ratio: float = 0.0


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_build()
	_sync_position()
	get_viewport().size_changed.connect(_sync_position)


func show_boss(entity_id: String, boss_template_id: String, title: String, hp: int, max_hp: int) -> void:
	_build()
	_boss_id = entity_id
	_boss_template_id = boss_template_id
	_title = title if title != "" else "Boss"
	_max_hp = maxi(1, max_hp)
	_hp = clampi(hp, 0, _max_hp)
	_ratio = clampf(float(_hp) / float(_max_hp), 0.0, 1.0)
	visible = _hp > 0
	_update_display()
	_sync_position()


func hide_boss() -> void:
	_boss_id = ""
	_boss_template_id = ""
	_title = ""
	_hp = 0
	_max_hp = 1
	_ratio = 0.0
	_clear_phase_fields()
	visible = false
	_update_display()


func set_phase_state(phase: Dictionary) -> void:
	_build()
	_phase_kind = str(phase.get("phase_kind", ""))
	_pattern_id = str(phase.get("pattern_id", ""))
	_phase_index = int(phase.get("phase_index", -1))
	_duration_ticks = maxi(0, int(phase.get("duration_ticks", 0)))
	_remaining_ticks = clampi(int(phase.get("remaining_ticks", _duration_ticks)), 0, _duration_ticks)
	_phase_ratio = _phase_ratio_from_ticks(_remaining_ticks, _duration_ticks)
	_update_display()


func clear_phase_state() -> void:
	_clear_phase_fields()
	_update_display()


func get_debug_state() -> Dictionary:
	return {
		"visible": visible and _boss_id != "",
		"boss_id": _boss_id,
		"boss_template_id": _boss_template_id,
		"title": _title,
		"hp": _hp,
		"max_hp": _max_hp,
		"ratio": _ratio,
		"phase_kind": _phase_kind,
		"pattern_id": _pattern_id,
		"phase_index": _phase_index,
		"duration_ticks": _duration_ticks,
		"remaining_ticks": _remaining_ticks,
		"phase_ratio": _phase_ratio,
	}


func _build() -> void:
	if _panel != null:
		return
	set_anchors_preset(Control.PRESET_FULL_RECT)
	z_index = 70
	visible = false

	_panel = PanelContainer.new()
	_panel.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_panel.custom_minimum_size = Vector2(PANEL_WIDTH, 72.0)
	var style := StyleBoxFlat.new()
	style.bg_color = Color(0.055, 0.045, 0.035, 0.92)
	style.border_color = Color("#9a7425")
	style.border_width_left = 1
	style.border_width_top = 1
	style.border_width_right = 1
	style.border_width_bottom = 1
	style.content_margin_left = 12
	style.content_margin_top = 7
	style.content_margin_right = 12
	style.content_margin_bottom = 8
	_panel.add_theme_stylebox_override("panel", style)
	add_child(_panel)

	var root := VBoxContainer.new()
	root.mouse_filter = Control.MOUSE_FILTER_IGNORE
	root.add_theme_constant_override("separation", 5)
	_panel.add_child(root)

	var top_row := HBoxContainer.new()
	top_row.mouse_filter = Control.MOUSE_FILTER_IGNORE
	top_row.add_theme_constant_override("separation", 10)
	root.add_child(top_row)

	_title_label = Label.new()
	_title_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_title_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_title_label.add_theme_color_override("font_color", Color("#f3c251"))
	_title_label.add_theme_font_size_override("font_size", 16)
	top_row.add_child(_title_label)

	_hp_label = Label.new()
	_hp_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_hp_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	_hp_label.add_theme_color_override("font_color", Color("#e8dcc8"))
	_hp_label.add_theme_font_size_override("font_size", 14)
	top_row.add_child(_hp_label)

	var bar_bg := ColorRect.new()
	bar_bg.mouse_filter = Control.MOUSE_FILTER_IGNORE
	bar_bg.custom_minimum_size = BAR_SIZE
	bar_bg.color = Color(0.12, 0.045, 0.035, 0.96)
	root.add_child(bar_bg)

	_fill = ColorRect.new()
	_fill.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_fill.color = Color("#b72323")
	_fill.size = BAR_SIZE
	bar_bg.add_child(_fill)

	_phase_label = Label.new()
	_phase_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_phase_label.add_theme_color_override("font_color", Color("#d7c6a0"))
	_phase_label.add_theme_font_size_override("font_size", 12)
	root.add_child(_phase_label)

	var phase_bg := ColorRect.new()
	phase_bg.mouse_filter = Control.MOUSE_FILTER_IGNORE
	phase_bg.custom_minimum_size = PHASE_BAR_SIZE
	phase_bg.color = Color(0.10, 0.085, 0.065, 0.88)
	root.add_child(phase_bg)

	_phase_fill = ColorRect.new()
	_phase_fill.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_phase_fill.color = Color("#f05a42")
	_phase_fill.size = PHASE_BAR_SIZE
	phase_bg.add_child(_phase_fill)

	_update_display()


func _update_display() -> void:
	if _title_label == null or _hp_label == null or _fill == null or _phase_label == null or _phase_fill == null:
		return
	_title_label.text = _title
	_hp_label.text = "%d / %d" % [_hp, _max_hp] if _boss_id != "" else ""
	_fill.size.x = BAR_SIZE.x * _ratio
	if _ratio > 0.6:
		_fill.color = Color("#b72323")
	elif _ratio > 0.3:
		_fill.color = Color("#c77622")
	else:
		_fill.color = Color("#d93a2f")
	if _phase_kind == "":
		_phase_label.text = ""
		_phase_fill.size.x = 0.0
		return
	_phase_label.text = "%s %d  %d / %d" % [
		_phase_kind.capitalize(),
		_phase_index,
		_remaining_ticks,
		_duration_ticks,
	]
	_phase_fill.size.x = PHASE_BAR_SIZE.x * _phase_ratio
	match _phase_kind:
		"telegraph":
			_phase_fill.color = Color("#f05a42")
		"active":
			_phase_fill.color = Color("#d32f2f")
		"recovery":
			_phase_fill.color = Color("#51b56d")
		_:
			_phase_fill.color = Color("#d7c6a0")


func _sync_position() -> void:
	if _panel == null:
		return
	var viewport_size := get_viewport_rect().size
	var x := maxf(8.0, (viewport_size.x - PANEL_WIDTH) * 0.5)
	_panel.set_deferred("position", Vector2(x, PANEL_TOP))


func _clear_phase_fields() -> void:
	_phase_kind = ""
	_pattern_id = ""
	_phase_index = -1
	_duration_ticks = 0
	_remaining_ticks = 0
	_phase_ratio = 0.0


func _phase_ratio_from_ticks(remaining_ticks: int, duration_ticks: int) -> float:
	if duration_ticks <= 0:
		return 0.0
	return clampf(float(remaining_ticks) / float(duration_ticks), 0.0, 1.0)
