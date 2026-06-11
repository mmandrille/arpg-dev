class_name SkillBar
extends Control

signal cast_skill_requested(skill_id: String)
signal open_skills_requested

const TICK_DURATION_S := 1.0 / 10.0

var skill_progression: Dictionary = {}
var _interactive: bool = true
var _skill_id: String = ""
var _rank: int = 0
var _max_rank: int = 0
var _remaining_ticks: float = 0.0
var _total_ticks: int = 0
var _panel: PanelContainer
var _slot: Button
var _badge: Label
var _cooldown: ProgressBar
var _flash_timer: float = 0.0
var _flash_color := Color("#f0dfbb")


func _ready() -> void:
	SkillRulesLoader.ensure_loaded()
	_skill_id = _current_skill_id()
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	set_process(true)
	_build()


func _process(delta: float) -> void:
	if _remaining_ticks > 0.0:
		_remaining_ticks = maxf(0.0, _remaining_ticks - (delta / TICK_DURATION_S))
		_render()
	if _flash_timer > 0.0:
		_flash_timer = maxf(0.0, _flash_timer - delta)
		if _slot != null and _flash_timer <= 0.0:
			_slot.modulate = Color.WHITE


func set_skill_progression(next_progression: Dictionary) -> void:
	skill_progression = next_progression.duplicate(true)
	_sync_rank_from_progression()
	_render()


func set_skill_id(skill_id: String) -> void:
	if skill_id == "":
		skill_id = SkillRulesLoader.first_skill_id()
	if skill_id == _current_skill_id():
		return
	_skill_id = skill_id
	_remaining_ticks = 0.0
	_total_ticks = 0
	_sync_rank_from_progression()
	_render()


func set_skill_cooldowns(cooldowns: Array) -> void:
	var found := false
	for row in cooldowns:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if str(rec.get("skill_id", "")) != _current_skill_id():
			continue
		_remaining_ticks = float(rec.get("remaining_ticks", 0))
		_total_ticks = int(rec.get("total_ticks", 0))
		found = true
		break
	if not found:
		_remaining_ticks = 0.0
		_total_ticks = 0
	_render()


func set_interactive(enabled: bool) -> void:
	_interactive = enabled
	_render()


func flash_cast() -> void:
	_flash(Color("#f0dfbb"))


func flash_rejected() -> void:
	_flash(Color("#dd5b48"))


func use_slot() -> void:
	if not _slot_enabled():
		return
	cast_skill_requested.emit(_current_skill_id())


func get_debug_state() -> Dictionary:
	return {
		"skill_id": _current_skill_id(),
		"skill_name": _skill_name(_current_skill_id()),
		"rank": _rank,
		"max_rank": _max_rank,
		"unspent_skill_points": _unspent_skill_points(),
		"upgrade_badge_visible": _badge.visible if _badge != null else false,
		"upgrade_badge_text": _badge.text if _badge != null else "",
		"enabled": _slot_enabled(),
		"disabled": not _slot_enabled(),
		"remaining_ticks": int(ceil(_remaining_ticks)),
		"total_ticks": _total_ticks,
		"cooldown_fraction": _cooldown_fraction(),
		"slot_text": _slot.text if _slot != null else "",
		"tooltip_text": _slot.tooltip_text if _slot != null else "",
	}


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	if _panel != null:
		_position_panel()


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = PanelContainer.new()
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	_panel.add_theme_stylebox_override("panel", _panel_style())
	add_child(_panel)

	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 3)
	_panel.add_child(box)

	_slot = Button.new()
	_slot.text = _skill_icon_label_text(_current_skill_id())
	_slot.tooltip_text = _tooltip_text(_current_skill_id())
	_slot.focus_mode = Control.FOCUS_NONE
	_slot.custom_minimum_size = Vector2(52, 52)
	_slot.pressed.connect(func() -> void: open_skills_requested.emit())
	_slot.add_theme_font_size_override("font_size", 22)
	box.add_child(_slot)

	_badge = _make_upgrade_badge()
	_slot.add_child(_badge)

	_cooldown = ProgressBar.new()
	_cooldown.min_value = 0.0
	_cooldown.max_value = 1.0
	_cooldown.step = 0.001
	_cooldown.show_percentage = false
	_cooldown.custom_minimum_size = Vector2(52, 6)
	box.add_child(_cooldown)

	_position_panel()
	_render()


func _position_panel() -> void:
	var vp := get_viewport_rect().size
	_panel.position = Vector2((vp.x * 0.5) + 302.0, vp.y - 78.0)
	_panel.size = Vector2(64.0, 64.0)


func _render() -> void:
	if _slot == null or _cooldown == null:
		return
	_slot.disabled = not _interactive
	_slot.text = _skill_icon_label_text(_current_skill_id()) if _rank > 0 else "-"
	_slot.tooltip_text = _tooltip_text(_current_skill_id())
	if _badge != null:
		_badge.visible = _unspent_skill_points() > 0
	_cooldown.value = _cooldown_fraction()


func _slot_enabled() -> bool:
	return _interactive and _rank > 0 and _remaining_ticks <= 0.0


func _cooldown_fraction() -> float:
	if _total_ticks <= 0:
		return 0.0
	return clampf(_remaining_ticks / float(_total_ticks), 0.0, 1.0)


func _unspent_skill_points() -> int:
	return int(skill_progression.get("unspent_skill_points", 0))


func _skill_row(skill_id: String) -> Dictionary:
	var rows: Array = skill_progression.get("skills", [])
	for row in rows:
		if typeof(row) == TYPE_DICTIONARY and str((row as Dictionary).get("skill_id", "")) == skill_id:
			return (row as Dictionary)
	return {}


func _current_skill_id() -> String:
	if _skill_id != "" and not SkillRulesLoader.skill_definition(_skill_id).is_empty():
		return _skill_id
	return SkillRulesLoader.first_skill_id()


func _sync_rank_from_progression() -> void:
	var skill := _skill_row(_current_skill_id())
	_rank = int(skill.get("rank", 0))
	_max_rank = int(skill.get("max_rank", int(SkillRulesLoader.skill_definition(_current_skill_id()).get("max_rank", 0))))


func _skill_name(skill_id: String) -> String:
	var def := SkillRulesLoader.skill_definition(skill_id)
	return str(def.get("name", skill_id))


func _skill_icon_label_text(skill_id: String) -> String:
	var presentation := SkillRulesLoader.skill_presentation(skill_id)
	var icon: Dictionary = presentation.get("icon", {})
	return str(icon.get("label", skill_id.substr(0, 1).to_upper()))


func _tooltip_text(skill_id: String) -> String:
	var presentation := SkillRulesLoader.skill_presentation(skill_id)
	var summary := str(presentation.get("summary", "Skill"))
	return "%s\n%s" % [_skill_name(skill_id), summary]


func _flash(color: Color) -> void:
	if _slot == null:
		return
	_flash_color = color
	_flash_timer = 0.2
	_slot.modulate = _flash_color


func _make_upgrade_badge() -> Label:
	var badge := Label.new()
	badge.text = "+"
	badge.mouse_filter = Control.MOUSE_FILTER_IGNORE
	badge.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	badge.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	badge.position = Vector2(34.0, -2.0)
	badge.size = Vector2(20.0, 20.0)
	badge.add_theme_font_size_override("font_size", 18)
	badge.add_theme_color_override("font_color", Color("#5eff62"))
	badge.add_theme_color_override("font_shadow_color", Color(0.0, 0.0, 0.0, 0.85))
	badge.add_theme_constant_override("shadow_offset_x", 1)
	badge.add_theme_constant_override("shadow_offset_y", 1)
	return badge


func _panel_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.06, 0.055, 0.045, 0.88)
	s.border_color = Color("#5c4a1f")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.corner_radius_top_left = 6
	s.corner_radius_top_right = 6
	s.corner_radius_bottom_left = 6
	s.corner_radius_bottom_right = 6
	s.content_margin_left = 6
	s.content_margin_right = 6
	s.content_margin_top = 5
	s.content_margin_bottom = 5
	return s
