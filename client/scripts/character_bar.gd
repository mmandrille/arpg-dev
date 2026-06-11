class_name CharacterBar
extends Control

signal open_character_requested

var _interactive: bool = true
var _progression: Dictionary = {}
var _panel: PanelContainer
var _slot: Button
var _badge: Label


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_build()


func set_interactive(enabled: bool) -> void:
	_interactive = enabled
	_render()


func set_progression(next_progression: Dictionary) -> void:
	_progression = next_progression.duplicate(true)
	_render()


func open_slot() -> void:
	if not _interactive:
		return
	open_character_requested.emit()


func get_debug_state() -> Dictionary:
	return {
		"enabled": _interactive,
		"disabled": not _interactive,
		"unspent_stat_points": _unspent_stat_points(),
		"upgrade_badge_visible": _badge.visible if _badge != null else false,
		"upgrade_badge_text": _badge.text if _badge != null else "",
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
	_slot.text = "C"
	_slot.tooltip_text = "Character"
	_slot.focus_mode = Control.FOCUS_NONE
	_slot.custom_minimum_size = Vector2(52, 52)
	_slot.pressed.connect(open_slot)
	_slot.add_theme_font_size_override("font_size", 22)
	box.add_child(_slot)

	_badge = _make_upgrade_badge()
	_slot.add_child(_badge)

	var spacer := Control.new()
	spacer.custom_minimum_size = Vector2(52, 6)
	box.add_child(spacer)

	_position_panel()
	_render()


func _position_panel() -> void:
	var vp := get_viewport_rect().size
	_panel.position = Vector2((vp.x * 0.5) - 366.0, vp.y - 78.0)
	_panel.size = Vector2(64.0, 64.0)


func _render() -> void:
	if _slot == null:
		return
	_slot.disabled = not _interactive
	if _badge != null:
		_badge.visible = _unspent_stat_points() > 0


func _unspent_stat_points() -> int:
	return int(_progression.get("unspent_stat_points", 0))


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
