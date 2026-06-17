extends Control
class_name BossHealthBar

const PANEL_WIDTH := 450.0
const PANEL_TOP := 58.0
const BAR_SIZE := Vector2(350.0, 14.0)
const PHASE_BAR_SIZE := Vector2(350.0, 6.0)
const PORTRAIT_SIZE := Vector2(58.0, 58.0)
const REWARD_PANEL_WIDTH := 390.0

var _panel: PanelContainer
var _reward_panel: PanelContainer
var _portrait: Control
var _title_label: Label
var _hp_label: Label
var _phase_label: Label
var _reward_title_label: Label
var _reward_status_label: Label
var _reward_hint_label: Label
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
var _reward_boss_template_id: String = ""
var _reward_title: String = ""
var _reward_status: String = ""
var _reward_hint: String = ""


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_build()
	_sync_position()
	get_viewport().size_changed.connect(_sync_position)


func show_boss(entity_id: String, boss_template_id: String, title: String, hp: int, max_hp: int) -> void:
	_build()
	_clear_reward_fields()
	_boss_id = entity_id
	_boss_template_id = boss_template_id
	_title = title if title != "" else "Boss"
	_max_hp = maxi(1, max_hp)
	_hp = clampi(hp, 0, _max_hp)
	_ratio = clampf(float(_hp) / float(_max_hp), 0.0, 1.0)
	_panel.visible = _hp > 0
	_update_display()
	_update_reward_display()
	_sync_root_visibility()
	_sync_position()


func hide_boss() -> void:
	hide_live_boss()
	clear_reward_status()


func hide_live_boss() -> void:
	_boss_id = ""
	_boss_template_id = ""
	_title = ""
	_hp = 0
	_max_hp = 1
	_ratio = 0.0
	_clear_phase_fields()
	if _panel != null:
		_panel.visible = false
	_update_display()
	_sync_root_visibility()


func show_reward_status(boss_template_id: String, title: String, status: String, hint: String) -> void:
	_build()
	hide_live_boss()
	_reward_boss_template_id = boss_template_id
	_reward_title = title if title != "" else "Boss"
	_reward_status = status
	_reward_hint = hint
	if _reward_panel != null:
		_reward_panel.visible = _reward_status != ""
	_update_reward_display()
	_sync_root_visibility()
	_sync_position()


func clear_reward_status() -> void:
	_clear_reward_fields()
	_update_reward_display()
	_sync_root_visibility()


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
		"portrait_visible": visible and _boss_id != "",
		"portrait_kind": _portrait_kind(),
		"portrait_label": _portrait_label(),
		"reward_panel_visible": _reward_visible(),
		"reward_boss_template_id": _reward_boss_template_id,
		"reward_title": _reward_title,
		"reward_status": _reward_status,
		"reward_hint": _reward_hint,
	}


func _build() -> void:
	if _panel != null:
		return
	set_anchors_preset(Control.PRESET_FULL_RECT)
	z_index = 70
	visible = false

	_panel = PanelContainer.new()
	_panel.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_panel.visible = false
	_panel.custom_minimum_size = Vector2(PANEL_WIDTH, 88.0)
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

	var root := HBoxContainer.new()
	root.mouse_filter = Control.MOUSE_FILTER_IGNORE
	root.add_theme_constant_override("separation", 10)
	_panel.add_child(root)

	_portrait = Control.new()
	_portrait.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_portrait.custom_minimum_size = PORTRAIT_SIZE
	_portrait.draw.connect(func() -> void: _draw_portrait())
	root.add_child(_portrait)

	var details := VBoxContainer.new()
	details.mouse_filter = Control.MOUSE_FILTER_IGNORE
	details.add_theme_constant_override("separation", 5)
	root.add_child(details)

	var top_row := HBoxContainer.new()
	top_row.mouse_filter = Control.MOUSE_FILTER_IGNORE
	top_row.add_theme_constant_override("separation", 10)
	details.add_child(top_row)

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
	details.add_child(bar_bg)

	_fill = ColorRect.new()
	_fill.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_fill.color = Color("#b72323")
	_fill.size = BAR_SIZE
	bar_bg.add_child(_fill)

	_phase_label = Label.new()
	_phase_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_phase_label.add_theme_color_override("font_color", Color("#d7c6a0"))
	_phase_label.add_theme_font_size_override("font_size", 12)
	details.add_child(_phase_label)

	var phase_bg := ColorRect.new()
	phase_bg.mouse_filter = Control.MOUSE_FILTER_IGNORE
	phase_bg.custom_minimum_size = PHASE_BAR_SIZE
	phase_bg.color = Color(0.10, 0.085, 0.065, 0.88)
	details.add_child(phase_bg)

	_phase_fill = ColorRect.new()
	_phase_fill.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_phase_fill.color = Color("#f05a42")
	_phase_fill.size = PHASE_BAR_SIZE
	phase_bg.add_child(_phase_fill)

	_reward_panel = PanelContainer.new()
	_reward_panel.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_reward_panel.visible = false
	_reward_panel.custom_minimum_size = Vector2(REWARD_PANEL_WIDTH, 88.0)
	var reward_style := StyleBoxFlat.new()
	reward_style.bg_color = Color(0.055, 0.075, 0.055, 0.94)
	reward_style.border_color = Color("#d4b36a")
	reward_style.border_width_left = 1
	reward_style.border_width_top = 1
	reward_style.border_width_right = 1
	reward_style.border_width_bottom = 1
	reward_style.content_margin_left = 14
	reward_style.content_margin_top = 9
	reward_style.content_margin_right = 14
	reward_style.content_margin_bottom = 10
	_reward_panel.add_theme_stylebox_override("panel", reward_style)
	add_child(_reward_panel)

	var reward_root := VBoxContainer.new()
	reward_root.mouse_filter = Control.MOUSE_FILTER_IGNORE
	reward_root.add_theme_constant_override("separation", 4)
	_reward_panel.add_child(reward_root)

	_reward_title_label = Label.new()
	_reward_title_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_reward_title_label.add_theme_color_override("font_color", Color("#f3c251"))
	_reward_title_label.add_theme_font_size_override("font_size", 15)
	reward_root.add_child(_reward_title_label)

	_reward_status_label = Label.new()
	_reward_status_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_reward_status_label.add_theme_color_override("font_color", Color("#e8dcc8"))
	_reward_status_label.add_theme_font_size_override("font_size", 13)
	reward_root.add_child(_reward_status_label)

	_reward_hint_label = Label.new()
	_reward_hint_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_reward_hint_label.add_theme_color_override("font_color", Color("#9fd69f"))
	_reward_hint_label.add_theme_font_size_override("font_size", 12)
	reward_root.add_child(_reward_hint_label)

	_update_display()
	_update_reward_display()


func _update_display() -> void:
	if _title_label == null or _hp_label == null or _fill == null or _phase_label == null or _phase_fill == null:
		return
	if _portrait != null:
		_portrait.queue_redraw()
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
	if _panel == null or _reward_panel == null:
		return
	var viewport_size := get_viewport_rect().size
	var x := maxf(8.0, (viewport_size.x - PANEL_WIDTH) * 0.5)
	_panel.set_deferred("position", Vector2(x, PANEL_TOP))
	var reward_x := maxf(8.0, (viewport_size.x - REWARD_PANEL_WIDTH) * 0.5)
	_reward_panel.set_deferred("position", Vector2(reward_x, PANEL_TOP))


func _clear_phase_fields() -> void:
	_phase_kind = ""
	_pattern_id = ""
	_phase_index = -1
	_duration_ticks = 0
	_remaining_ticks = 0
	_phase_ratio = 0.0


func _clear_reward_fields() -> void:
	_reward_boss_template_id = ""
	_reward_title = ""
	_reward_status = ""
	_reward_hint = ""
	if _reward_panel != null:
		_reward_panel.visible = false


func _update_reward_display() -> void:
	if _reward_title_label == null or _reward_status_label == null or _reward_hint_label == null:
		return
	_reward_title_label.text = _reward_title
	_reward_status_label.text = _reward_status
	_reward_hint_label.text = _reward_hint


func _reward_visible() -> bool:
	return _reward_panel != null and _reward_panel.visible and _reward_status != ""


func _sync_root_visibility() -> void:
	visible = (_panel != null and _panel.visible and _boss_id != "") or _reward_visible()


func _phase_ratio_from_ticks(remaining_ticks: int, duration_ticks: int) -> float:
	if duration_ticks <= 0:
		return 0.0
	return clampf(float(remaining_ticks) / float(duration_ticks), 0.0, 1.0)


func _portrait_kind() -> String:
	if _boss_id == "":
		return ""
	if _boss_template_id == "cave_warden":
		return "cave_warden"
	return "boss"


func _portrait_label() -> String:
	match _portrait_kind():
		"cave_warden":
			return "CW"
		"boss":
			return "B"
		_:
			return ""


func _draw_portrait() -> void:
	if _portrait == null:
		return
	var rect := Rect2(Vector2.ZERO, PORTRAIT_SIZE)
	_portrait.draw_rect(rect, Color("#211812"), true)
	_portrait.draw_rect(rect.grow(-1.0), Color("#9a7425"), false, 1.5)
	var center := rect.get_center()
	match _portrait_kind():
		"cave_warden":
			_portrait.draw_circle(center + Vector2(0.0, 2.0), 19.0, Color("#5f4733"))
			_portrait.draw_arc(center + Vector2(-12.0, -10.0), 12.0, 2.9, 5.3, 18, Color("#d4b36a"), 3.0, true)
			_portrait.draw_arc(center + Vector2(12.0, -10.0), 12.0, 4.1, 6.5, 18, Color("#d4b36a"), 3.0, true)
			_portrait.draw_circle(center + Vector2(-7.0, -2.0), 2.4, Color("#f05a42"))
			_portrait.draw_circle(center + Vector2(7.0, -2.0), 2.4, Color("#f05a42"))
			_portrait.draw_line(center + Vector2(-9.0, 12.0), center + Vector2(9.0, 12.0), Color("#d4b36a"), 2.0)
		"boss":
			_portrait.draw_circle(center, 18.0, Color("#604138"))
			_portrait.draw_circle(center + Vector2(-6.0, -3.0), 2.0, Color("#f05a42"))
			_portrait.draw_circle(center + Vector2(6.0, -3.0), 2.0, Color("#f05a42"))
		_:
			pass
