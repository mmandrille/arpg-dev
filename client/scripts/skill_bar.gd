class_name SkillBar
extends Control

signal cast_skill_requested(skill_id: String)
signal open_skills_requested

const TICK_DURATION_S := 1.0 / 10.0
const SkillIconScript := preload("res://scripts/skill_icon.gd")
const SkillsPanelScript := preload("res://scripts/skills_panel.gd")

var skill_progression: Dictionary = {}
var character_progression: Dictionary = {}
var _interactive: bool = true
var _skill_id: String = ""
var _rank: int = 0
var _max_rank: int = 0
var _current_mana: int = 0
var _max_mana: int = 0
var _remaining_ticks: float = 0.0
var _total_ticks: int = 0
var _panel: PanelContainer
var _slot: Button
var _slot_icon
var _badge: Label
var _mana_cost_label: Label
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


func set_character_progression(next_progression: Dictionary) -> void:
	character_progression = next_progression.duplicate(true)
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


func set_player_mana(current_mana: int, max_mana: int) -> void:
	_current_mana = maxi(0, current_mana)
	_max_mana = maxi(0, max_mana)
	_render()


func set_interactive(enabled: bool) -> void:
	_interactive = enabled
	_render()


func flash_cast() -> void:
	_flash(Color("#f0dfbb"))


func flash_rejected() -> void:
	_flash(Color("#dd5b48"))


func use_slot() -> void:
	if not _cast_enabled():
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
		"upgrade_badge_position": _vec2_debug(_badge.position if _badge != null else Vector2.ZERO),
		"upgrade_badge_font_size": _badge.label_settings.font_size if _badge != null and _badge.label_settings != null else 0,
		"upgrade_badge_outline_size": _badge.label_settings.outline_size if _badge != null and _badge.label_settings != null else 0,
		"enabled": _cast_enabled(),
		"disabled": not _cast_enabled(),
		"greyed": _is_greyed(),
		"mana": _current_mana,
		"mana_cost": _mana_cost(),
		"mana_cost_visible": _mana_cost_label.visible if _mana_cost_label != null else false,
		"mana_cost_text": _mana_cost_label.text if _mana_cost_label != null else "",
		"not_enough_mana": _rank > 0 and _current_mana < _mana_cost(),
		"remaining_ticks": int(ceil(_remaining_ticks)),
		"total_ticks": _total_ticks,
		"cooldown_fraction": _cooldown_fraction(),
		"cooldown_visible": _cooldown.visible if _cooldown != null else false,
		"slot_text": _slot.text if _slot != null else "",
		"icon_shape": _slot_icon.shape if _slot_icon != null else "",
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
	_slot.text = ""
	_slot.tooltip_text = _tooltip_text(_current_skill_id())
	_slot.focus_mode = Control.FOCUS_NONE
	_slot.custom_minimum_size = Vector2(52, 52)
	_slot.pressed.connect(func() -> void: open_skills_requested.emit())
	_slot.add_theme_font_size_override("font_size", 22)
	box.add_child(_slot)

	_slot_icon = SkillIconScript.new()
	_slot_icon.position = Vector2(4, 4)
	_slot_icon.size = Vector2(44, 44)
	_slot.add_child(_slot_icon)

	_badge = _make_upgrade_badge()
	_slot.add_child(_badge)

	_mana_cost_label = _make_mana_cost_label()
	_slot.add_child(_mana_cost_label)

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
	_slot.disabled = false
	_slot.text = "-" if _rank <= 0 else ""
	_slot.tooltip_text = _tooltip_text(_current_skill_id())
	if _slot_icon != null:
		_slot_icon.visible = _rank > 0
		_slot_icon.configure(_current_skill_id(), SkillRulesLoader.skill_presentation(_current_skill_id()))
		_slot_icon.modulate = Color(0.38, 0.38, 0.38, 0.72) if _is_greyed() else Color.WHITE
	if _mana_cost_label != null:
		_mana_cost_label.visible = _rank > 0
		_mana_cost_label.text = str(_mana_cost())
		_mana_cost_label.modulate = Color("#8a8a8a") if _current_mana < _mana_cost() else Color("#9fe4ff")
	if _badge != null:
		_badge.visible = _unspent_skill_points() > 0
	var cooldown_fraction := _cooldown_fraction()
	_cooldown.visible = cooldown_fraction > 0.0
	_cooldown.value = cooldown_fraction


func _cast_enabled() -> bool:
	return _interactive and _rank > 0 and _remaining_ticks <= 0.0 and _current_mana >= _mana_cost()


func _is_greyed() -> bool:
	return _rank > 0 and (not _interactive or _remaining_ticks > 0.0 or _current_mana < _mana_cost())


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
	return SkillRulesLoader.skill_display_name(skill_id)


func _skill_icon_label_text(skill_id: String) -> String:
	var presentation := SkillRulesLoader.skill_presentation(skill_id)
	var icon: Dictionary = presentation.get("icon", {})
	return str(icon.get("label", skill_id.substr(0, 1).to_upper()))


func _tooltip_text(skill_id: String) -> String:
	return SkillsPanelScript.skill_tooltip_text(skill_id, _rank, _max_rank, skill_progression, character_progression)


func _mana_cost() -> int:
	var def := SkillRulesLoader.skill_definition(_current_skill_id())
	var cost: Dictionary = def.get("cost", {})
	var mana: Dictionary = cost.get("mana", {})
	return maxi(0, int(mana.get("base", 0)) + int(mana.get("per_rank", 0)) * maxi(0, _rank - 1))


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
	badge.position = Vector2(-3.0, -5.0)
	badge.size = Vector2(24.0, 24.0)
	var settings := LabelSettings.new()
	settings.font_size = 22
	settings.font_color = Color("#5eff62")
	settings.outline_size = 3
	settings.outline_color = Color(0.0, 0.0, 0.0, 0.95)
	badge.label_settings = settings
	return badge


func _make_mana_cost_label() -> Label:
	var label := Label.new()
	label.text = ""
	label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	label.vertical_alignment = VERTICAL_ALIGNMENT_BOTTOM
	label.position = Vector2(24.0, 29.0)
	label.size = Vector2(24.0, 18.0)
	label.add_theme_font_size_override("font_size", 14)
	label.add_theme_color_override("font_color", Color("#9fe4ff"))
	label.add_theme_color_override("font_shadow_color", Color(0.0, 0.0, 0.0, 0.9))
	label.add_theme_constant_override("shadow_offset_x", 1)
	label.add_theme_constant_override("shadow_offset_y", 1)
	return label


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


func _vec2_debug(v: Vector2) -> Dictionary:
	return {"x": v.x, "y": v.y}
