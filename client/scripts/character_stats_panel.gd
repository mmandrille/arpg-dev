class_name CharacterStatsPanel
extends Control

signal allocate_stat_requested(stat: String)

const BASE_STATS := ["str", "dex", "vit", "magic"]
const DERIVED_LABELS := {
	"damage_min": "Damage min",
	"damage_max": "Damage max",
	"armor": "Armor",
	"attack_speed": "Attack speed",
	"hit_chance": "Hit chance",
	"crit_chance": "Crit chance",
	"crit_damage": "Crit damage",
	"block_percent": "Block",
	"movement_speed": "Move speed",
	"max_hp": "HP",
	"max_mana": "Mana",
}

var progression: Dictionary = {}
var allocation_enabled: bool = false
var _panel: PanelContainer
var _level_label: Label
var _xp_label: Label
var _points_label: Label
var _stat_value_labels: Dictionary = {}
var _stat_buttons: Dictionary = {}
var _derived_labels: Dictionary = {}


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


func set_progression(next_progression: Dictionary) -> void:
	progression = next_progression.duplicate(true)
	_render()


func set_allocation_enabled(enabled: bool) -> void:
	allocation_enabled = enabled
	_render_buttons()


func get_debug_state() -> Dictionary:
	var stat_buttons := {}
	for stat in BASE_STATS:
		var btn: Button = _stat_buttons.get(stat, null)
		stat_buttons[stat] = {
			"enabled": btn != null and not btn.disabled,
			"disabled": btn == null or btn.disabled,
		}
	return {
		"visible": visible,
		"progression": progression.duplicate(true),
		"allocation_enabled": allocation_enabled,
		"stat_buttons": stat_buttons,
		"stat_breakdowns": _breakdowns_by_key(),
	}


func bot_click_stat_button(stat: String) -> void:
	var btn: Button = _stat_buttons.get(stat, null)
	if btn == null or btn.disabled:
		return
	btn.pressed.emit()


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = PanelContainer.new()
	_panel.custom_minimum_size = Vector2(286, 420)
	_panel.position = Vector2(16, 118)
	_panel.add_theme_stylebox_override("panel", _panel_style())
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	add_child(_panel)

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 8)
	root.custom_minimum_size = Vector2(260, 390)
	_panel.add_child(root)

	var title := Label.new()
	title.text = "Character"
	title.add_theme_font_size_override("font_size", 18)
	title.add_theme_color_override("font_color", Color("#f0dfbb"))
	root.add_child(title)

	_level_label = _value_label()
	_xp_label = _value_label()
	_points_label = _value_label()
	root.add_child(_level_label)
	root.add_child(_xp_label)
	root.add_child(_points_label)

	root.add_child(_section_label("Stats"))
	for stat in BASE_STATS:
		var row := HBoxContainer.new()
		row.add_theme_constant_override("separation", 8)
		var label := _value_label()
		label.custom_minimum_size = Vector2(126, 24)
		row.add_child(label)
		var btn := Button.new()
		btn.text = "+"
		btn.tooltip_text = "Spend point"
		btn.custom_minimum_size = Vector2(32, 24)
		btn.focus_mode = Control.FOCUS_NONE
		btn.pressed.connect(_on_stat_button_pressed.bind(stat))
		row.add_child(btn)
		_stat_value_labels[stat] = label
		_stat_buttons[stat] = btn
		root.add_child(row)

	root.add_child(_section_label("Derived"))
	for key in DERIVED_LABELS.keys():
		var label := _value_label()
		_derived_labels[key] = label
		root.add_child(label)

	_render()


func _render() -> void:
	var level := int(progression.get("level", 1))
	var xp := int(progression.get("experience", 0))
	var remaining = progression.get("experience_to_next_level", null)
	var points := int(progression.get("unspent_stat_points", 0))
	_level_label.text = "Level %d" % level
	_xp_label.text = "XP %d%s" % [xp, "" if remaining == null else " (+%d)" % int(remaining)]
	_points_label.text = "Points %d" % points
	var base: Dictionary = progression.get("base_stats", {})
	for stat in BASE_STATS:
		var label: Label = _stat_value_labels.get(stat, null)
		if label != null:
			label.text = "%s  %d" % [stat.to_upper(), int(base.get(stat, 0))]
	var derived: Dictionary = progression.get("derived_stats", {})
	for key in DERIVED_LABELS.keys():
		var label: Label = _derived_labels.get(key, null)
		if label != null:
			label.text = "%s  %s" % [DERIVED_LABELS[key], _format_number(float(derived.get(key, 0.0)))]
			label.tooltip_text = _breakdown_summary(key)
	_render_buttons()


func _on_stat_button_pressed(stat: String) -> void:
	allocate_stat_requested.emit(stat)


func _render_buttons() -> void:
	var points := int(progression.get("unspent_stat_points", 0))
	for stat in BASE_STATS:
		var btn: Button = _stat_buttons.get(stat, null)
		if btn != null:
			btn.disabled = not allocation_enabled or points <= 0


func _format_number(value: float) -> String:
	if absf(value - roundf(value)) < 0.0001:
		return str(int(roundf(value)))
	return "%.2f" % value


func _breakdowns_by_key() -> Dictionary:
	var out := {}
	var rows: Array = progression.get("stat_breakdowns", [])
	for row in rows:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		out[str(rec.get("key", ""))] = rec.duplicate(true)
	return out


func _breakdown_summary(key: String) -> String:
	var rec: Dictionary = _breakdowns_by_key().get(key, {})
	if rec.is_empty():
		return ""
	var parts: PackedStringArray = PackedStringArray()
	var value := _format_number(float(rec.get("value", 0.0)))
	if rec.get("cap", null) != null:
		parts.append("%s: %s (cap %s)" % [DERIVED_LABELS.get(key, key), value, _format_number(float(rec.get("cap", 0.0)))])
	else:
		parts.append("%s: %s" % [DERIVED_LABELS.get(key, key), value])
	var sources: Array = rec.get("sources", [])
	for source in sources:
		if typeof(source) != TYPE_DICTIONARY:
			continue
		var source_rec := source as Dictionary
		var label := str(source_rec.get("label", source_rec.get("kind", "")))
		parts.append("%s %s" % [label, _format_number(float(source_rec.get("value", 0.0)))])
	return "\n".join(parts)


func _value_label() -> Label:
	var label := Label.new()
	label.add_theme_color_override("font_color", Color("#d8c7a6"))
	label.add_theme_font_size_override("font_size", 12)
	return label


func _section_label(text: String) -> Label:
	var label := _value_label()
	label.text = text
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
