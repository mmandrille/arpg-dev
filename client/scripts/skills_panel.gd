class_name SkillsPanel
extends Control

signal allocate_skill_point_requested(skill_id: String)

const MAGIC_BOLT_ID := "magic_bolt"

var skill_progression: Dictionary = {}
var interactive: bool = true
var _hovered_skill_id: String = ""
var _hover_controls: Array[Control] = []
var _panel: PanelContainer
var _points_label: Label
var _rank_label: Label
var _spend_button: Button


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_build()
	visible = false


func toggle() -> void:
	visible = not visible
	_apply_mouse_filter()


func ensure_display_visible() -> void:
	visible = true
	_apply_mouse_filter()


func hide_display() -> void:
	visible = false
	_apply_mouse_filter()


func set_skill_progression(next_progression: Dictionary) -> void:
	skill_progression = next_progression.duplicate(true)
	_render()


func set_interactive(enabled: bool) -> void:
	interactive = enabled
	_render()


func hovered_skill_id() -> String:
	return _hovered_skill_id


func get_debug_state() -> Dictionary:
	var skill := _skill_row(MAGIC_BOLT_ID)
	return {
		"visible": visible,
		"unspent_skill_points": int(skill_progression.get("unspent_skill_points", 0)),
		"skill_id": MAGIC_BOLT_ID,
		"rank": int(skill.get("rank", 0)),
		"max_rank": int(skill.get("max_rank", 0)),
		"can_spend": bool(skill.get("can_spend", false)),
		"spend_button_enabled": _spend_button != null and not _spend_button.disabled,
		"hovered_skill_id": _hovered_skill_id,
	}


func bot_click_skill_button(skill_id: String = MAGIC_BOLT_ID) -> void:
	if skill_id != MAGIC_BOLT_ID:
		return
	if _spend_button == null or _spend_button.disabled:
		return
	_spend_button.pressed.emit()


func bot_hover_skill(skill_id: String = MAGIC_BOLT_ID) -> void:
	if skill_id != MAGIC_BOLT_ID:
		return
	_hovered_skill_id = skill_id


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = PanelContainer.new()
	_panel.custom_minimum_size = Vector2(330, 500)
	_panel.position = Vector2(16, 118)
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	_panel.add_theme_stylebox_override("panel", _panel_style())
	add_child(_panel)

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 8)
	root.custom_minimum_size = Vector2(304, 470)
	_panel.add_child(root)

	var title := _label("Skills", 33, Color("#f0dfbb"))
	root.add_child(title)
	_points_label = _value_label()
	root.add_child(_points_label)

	root.add_child(_section_label("Active"))
	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 8)
	row.mouse_filter = Control.MOUSE_FILTER_PASS
	root.add_child(row)
	var name_label := _value_label("Magic Bolt")
	name_label.custom_minimum_size = Vector2(178, 30)
	row.add_child(name_label)
	_rank_label = _value_label()
	_rank_label.custom_minimum_size = Vector2(62, 30)
	row.add_child(_rank_label)
	_spend_button = Button.new()
	_spend_button.text = "+"
	_spend_button.tooltip_text = "Spend skill point"
	_spend_button.focus_mode = Control.FOCUS_NONE
	_spend_button.custom_minimum_size = Vector2(38, 30)
	_spend_button.pressed.connect(_on_spend_pressed)
	row.add_child(_spend_button)
	for control in [row, name_label, _rank_label, _spend_button]:
		_bind_skill_hover(control, MAGIC_BOLT_ID)

	_render()


func _render() -> void:
	if _points_label == null or _rank_label == null or _spend_button == null:
		return
	var unspent := int(skill_progression.get("unspent_skill_points", 0))
	var skill := _skill_row(MAGIC_BOLT_ID)
	var rank := int(skill.get("rank", 0))
	var max_rank := int(skill.get("max_rank", 0))
	_points_label.text = "Skill points %d" % unspent
	_rank_label.text = "%d / %d" % [rank, max_rank]
	_spend_button.disabled = not interactive or unspent <= 0 or rank >= max_rank or not bool(skill.get("can_spend", false))


func _on_spend_pressed() -> void:
	if _spend_button == null or _spend_button.disabled:
		return
	allocate_skill_point_requested.emit(MAGIC_BOLT_ID)


func _skill_row(skill_id: String) -> Dictionary:
	var rows: Array = skill_progression.get("skills", [])
	for row in rows:
		if typeof(row) == TYPE_DICTIONARY and str((row as Dictionary).get("skill_id", "")) == skill_id:
			return (row as Dictionary)
	return {}


func _bind_skill_hover(control: Control, skill_id: String) -> void:
	if control == null:
		return
	if control.mouse_filter == Control.MOUSE_FILTER_IGNORE:
		control.mouse_filter = Control.MOUSE_FILTER_PASS
	_hover_controls.append(control)
	control.mouse_entered.connect(func() -> void:
		_hovered_skill_id = skill_id
	)
	control.mouse_exited.connect(func() -> void:
		if not _mouse_over_skill_controls():
			_hovered_skill_id = ""
	)


func _mouse_over_skill_controls() -> bool:
	var mouse_pos := get_viewport().get_mouse_position()
	for control in _hover_controls:
		if control != null and control.is_inside_tree() and control.get_global_rect().has_point(mouse_pos):
			return true
	return false


func _label(text: String, size: int, color: Color) -> Label:
	var label := Label.new()
	label.text = text
	label.add_theme_font_size_override("font_size", size)
	label.add_theme_color_override("font_color", color)
	return label


func _value_label(text: String = "") -> Label:
	return _label(text, 23, Color("#d8c7a6"))


func _section_label(text: String) -> Label:
	var label := _value_label(text)
	label.add_theme_color_override("font_color", Color("#c9a227"))
	return label


func _apply_mouse_filter() -> void:
	if _panel != null:
		_panel.mouse_filter = Control.MOUSE_FILTER_STOP if visible else Control.MOUSE_FILTER_IGNORE


func _panel_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.06, 0.055, 0.045, 0.92)
	s.border_color = Color("#6b5420")
	s.border_width_left = 2
	s.border_width_top = 2
	s.border_width_right = 2
	s.border_width_bottom = 2
	s.corner_radius_top_left = 6
	s.corner_radius_top_right = 6
	s.corner_radius_bottom_left = 6
	s.corner_radius_bottom_right = 6
	s.content_margin_left = 12
	s.content_margin_right = 12
	s.content_margin_top = 12
	s.content_margin_bottom = 12
	return s
