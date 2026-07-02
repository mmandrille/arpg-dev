class_name CharacterStatsPanel
extends Control

signal allocate_stat_requested(stat: String)

const StatLabels := preload("res://scripts/stat_labels.gd")
const CharacterStatsBreakdown := preload("res://scripts/character_stats_breakdown.gd")
const StatTooltipLabelScript := preload("res://scripts/stat_tooltip_label.gd")
const CharacterPanelStyles := preload("res://scripts/character_panel_styles.gd")
const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const TextCatalogScript := preload("res://scripts/text_catalog.gd")
const BASE_STATS := StatLabels.BASE_STATS
const DERIVED_LABELS := {
	"damage_min": "Damage min",
	"damage_max": "Damage max",
	"armor": "Armor",
	"attack_speed": "Attack speed",
	"attack_interval_ticks": "Attack interval",
	"hit_chance": "Hit chance",
	"crit_chance": "Crit chance",
	"crit_damage": "Crit damage",
	"evade_chance": "Evade chance",
	"block_percent": "Block",
	"movement_speed": "Move speed",
	"max_hp": "HP",
	"max_mana": "Mana",
	"health_regen_per_second": "HP regen /s",
	"mana_regen_per_second": "Mana regen /s",
	"light_radius": "Light radius",
}
const DUAL_WIELD_DAMAGE_KEYS := ["damage_min", "damage_max"]

var progression: Dictionary = {}
var allocation_enabled: bool = false
var _panel: DraggableWindow
var _level_label: Label
var _xp_label: Label
var _points_label: Label
var _stat_value_labels: Dictionary = {}
var _stat_base_labels: Dictionary = {}
var _stat_effective_labels: Dictionary = {}
var _stat_buttons: Dictionary = {}
var _derived_name_labels: Dictionary = {}
var _derived_labels: Dictionary = {}
var _derived_off_labels: Dictionary = {}
var _derived_header_name: Label
var _derived_header_main: Label
var _derived_header_off: Label
var _derived_dual_wield_active: bool = false
var _derived_title: Label
var _derived_scroll: ScrollContainer
var _derived_container: VBoxContainer
var _hero_name: String = "Character"


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
	_sync_title()
	_render()


func set_hero_name(next_name: String) -> void:
	_hero_name = next_name.strip_edges()
	if _hero_name == "":
		_hero_name = "Character"
	_sync_title()


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
	var derived_labels := {}
	for key in _derived_labels.keys():
		var name_label: Label = _derived_name_labels.get(key, null)
		var label: Label = _derived_labels.get(key, null)
		var off_label: Label = _derived_off_labels.get(key, null)
		if label != null and name_label != null:
			if _derived_dual_wield_active and key in DUAL_WIELD_DAMAGE_KEYS and off_label != null:
				derived_labels[key] = "%s  %s  %s" % [name_label.text, label.text, off_label.text]
			else:
				derived_labels[key] = "%s  %s" % [name_label.text, label.text]
	var stat_labels := {}
	var stat_effective_styles := {}
	for stat in BASE_STATS:
		var label: Label = _stat_value_labels.get(stat, null)
		var base_label: Label = _stat_base_labels.get(stat, null)
		var effective_label: Label = _stat_effective_labels.get(stat, null)
		if label != null and base_label != null and effective_label != null:
			stat_labels[stat] = "%s  %s / %s" % [label.text, base_label.text, effective_label.text]
			stat_effective_styles[stat] = {
				"color": effective_label.get_theme_color("font_color").to_html(false),
				"outline_size": effective_label.get_theme_constant("outline_size"),
			}
	return {
		"visible": visible,
		"progression": progression.duplicate(true),
		"allocation_enabled": allocation_enabled,
		"stat_buttons": stat_buttons,
		"stat_labels": stat_labels,
		"stat_columns": ["NAME", "BASE", "EFFECTIVE"],
		"stat_effective_styles": stat_effective_styles,
		"stat_tooltips": _stat_tooltips_by_key(),
		"stat_mouse_filters": _stat_mouse_filters_by_key(),
		"derived_open": true,
		"derived_labels": derived_labels,
		"derived_columns": (["NAME", "MAIN", "OFF"] if _derived_dual_wield_active else ["NAME", "VALUE"]),
		"derived_title": _derived_title.text if _derived_title != null else "",
		"derived_title_is_button": false,
		"derived_tooltips": _derived_tooltips_by_key(),
		"derived_tooltips_main": _derived_slot_tooltips_by_key(_derived_labels),
		"derived_tooltips_off": _derived_slot_tooltips_by_key(_derived_off_labels),
		"derived_mouse_filters": _derived_mouse_filters_by_key(),
		"derived_scroll": _derived_scroll_debug_state(),
		"derived_tooltip_panel": _derived_tooltip_panel_debug_state(),
		"stat_breakdowns": CharacterStatsBreakdown.breakdowns_by_key(progression),
		"hero_name": _hero_name,
		"window": _panel.get_debug_state() if _panel != null else {},
	}


func bot_click_stat_button(stat: String) -> void:
	var btn: Button = _stat_buttons.get(stat, null)
	if btn == null or btn.disabled:
		return
	btn.pressed.emit()


func bot_click_close() -> void:
	if _panel != null and _panel.close_button() != null:
		_panel.close_button().pressed.emit()


func bot_drag_window_by(delta: Vector2) -> void:
	if _panel != null:
		_panel.bot_drag_by(delta)


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = DraggableWindowScript.new()
	_panel.custom_minimum_size = Vector2(330, 585)
	_panel.position = Vector2(16, 118)
	_panel.configure("Character", Vector2(304, 520))
	_panel.set_layout_key("character_stats")
	_panel.add_theme_stylebox_override("panel", CharacterPanelStyles.panel_style())
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	_panel.close_requested.connect(hide_display)
	add_child(_panel)

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 8)
	root.custom_minimum_size = Vector2(304, 520)
	_panel.set_content(root)

	_level_label = _value_label()
	_xp_label = _value_label()
	_points_label = _value_label()
	root.add_child(_level_label)
	root.add_child(_xp_label)
	root.add_child(_points_label)

	root.add_child(_section_label("Stats"))
	var stat_header := HBoxContainer.new()
	stat_header.add_theme_constant_override("separation", 4)
	stat_header.add_child(_header_label("NAME", 90, HORIZONTAL_ALIGNMENT_LEFT))
	stat_header.add_child(_header_label("BASE", 54, HORIZONTAL_ALIGNMENT_RIGHT))
	stat_header.add_child(_header_label("EFFECTIVE", 86, HORIZONTAL_ALIGNMENT_RIGHT))
	stat_header.add_child(_header_label("", 36, HORIZONTAL_ALIGNMENT_CENTER))
	root.add_child(stat_header)
	for stat in BASE_STATS:
		var row := HBoxContainer.new()
		row.add_theme_constant_override("separation", 4)
		var label := _derived_value_label()
		label.custom_minimum_size = Vector2(90, 28)
		label.mouse_filter = Control.MOUSE_FILTER_STOP
		row.add_child(label)
		var base_label := _table_value_label(54)
		row.add_child(base_label)
		var effective_label := _derived_value_label()
		effective_label.custom_minimum_size = Vector2(86, 28)
		effective_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
		effective_label.mouse_filter = Control.MOUSE_FILTER_STOP
		row.add_child(effective_label)
		var btn := Button.new()
		btn.text = "+"
		btn.tooltip_text = "Spend point"
		btn.custom_minimum_size = Vector2(36, 28)
		btn.focus_mode = Control.FOCUS_NONE
		btn.pressed.connect(_on_stat_button_pressed.bind(stat))
		row.add_child(btn)
		_stat_value_labels[stat] = label
		_stat_base_labels[stat] = base_label
		_stat_effective_labels[stat] = effective_label
		_stat_buttons[stat] = btn
		root.add_child(row)

	_derived_title = _section_label("Derived")
	_derived_title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_derived_title.custom_minimum_size = Vector2(180, 28)
	root.add_child(_derived_title)
	var derived_header := HBoxContainer.new()
	derived_header.add_theme_constant_override("separation", 6)
	_derived_header_name = _header_label("NAME", 188, HORIZONTAL_ALIGNMENT_LEFT)
	_derived_header_main = _header_label("VALUE", 72, HORIZONTAL_ALIGNMENT_RIGHT)
	_derived_header_off = _header_label("OFF", 72, HORIZONTAL_ALIGNMENT_RIGHT)
	_derived_header_off.visible = false
	derived_header.add_child(_derived_header_name)
	derived_header.add_child(_derived_header_main)
	derived_header.add_child(_derived_header_off)
	root.add_child(derived_header)

	_derived_scroll = ScrollContainer.new()
	_derived_scroll.visible = true
	_derived_scroll.custom_minimum_size = Vector2(304, 172)
	_derived_scroll.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_derived_scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_derived_scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	_derived_scroll.vertical_scroll_mode = ScrollContainer.SCROLL_MODE_SHOW_ALWAYS
	_derived_scroll.clip_contents = true
	root.add_child(_derived_scroll)

	_derived_container = VBoxContainer.new()
	_derived_container.add_theme_constant_override("separation", 3)
	_derived_container.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_derived_scroll.add_child(_derived_container)
	for key in DERIVED_LABELS.keys():
		var row := HBoxContainer.new()
		row.add_theme_constant_override("separation", 6)
		_derived_container.add_child(row)
		var name_label := _derived_value_label()
		name_label.custom_minimum_size = Vector2(188, 28)
		name_label.mouse_filter = Control.MOUSE_FILTER_STOP
		row.add_child(name_label)
		var label := _derived_value_label()
		label.custom_minimum_size = Vector2(72, 28)
		label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
		label.mouse_filter = Control.MOUSE_FILTER_STOP
		_derived_name_labels[key] = name_label
		_derived_labels[key] = label
		row.add_child(label)
		if key in DUAL_WIELD_DAMAGE_KEYS:
			var off_label := _derived_value_label()
			off_label.custom_minimum_size = Vector2(72, 28)
			off_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
			off_label.mouse_filter = Control.MOUSE_FILTER_STOP
			off_label.visible = false
			row.add_child(off_label)
			_derived_off_labels[key] = off_label

	_render()


func _sync_title() -> void:
	if _panel != null:
		_panel.configure(_title_text(), Vector2(304, 520))


func _title_text() -> String:
	var class_id := str(progression.get("character_class", "")).strip_edges()
	if class_id == "":
		return _hero_name
	return "%s (%s)" % [_hero_name, TextCatalogScript.get_text("character.class.%s" % class_id, class_id.capitalize())]


func _render() -> void:
	var level := int(progression.get("level", 1))
	var xp := int(progression.get("experience", 0))
	var remaining = progression.get("experience_to_next_level", null)
	var points := int(progression.get("unspent_stat_points", 0))
	_level_label.text = "Level %d" % level
	_xp_label.text = "XP %d%s" % [xp, "" if remaining == null else " (+%d)" % int(remaining)]
	_points_label.text = "Points %d" % points
	var base: Dictionary = progression.get("base_stats", {})
	var effective := CharacterStatsBreakdown.effective_base_stats(progression)
	for stat in BASE_STATS:
		var label: Label = _stat_value_labels.get(stat, null)
		var base_label: Label = _stat_base_labels.get(stat, null)
		var effective_label: Label = _stat_effective_labels.get(stat, null)
		var base_value := int(base.get(stat, 0))
		var effective_value := int(effective.get(stat, base_value))
		var tooltip := CharacterStatsBreakdown.breakdown_summary(progression, stat, DERIVED_LABELS)
		if label != null:
			label.text = StatLabels.display_name(stat)
			label.tooltip_text = ""
		if base_label != null:
			base_label.text = str(base_value)
		if effective_label != null:
			effective_label.text = str(effective_value)
			effective_label.tooltip_text = tooltip
			effective_label.call("apply_effective_stat_style", base_value, effective_value)
	var derived: Dictionary = progression.get("derived_stats", {})
	var dual_wield := _dual_wield_damage_active(derived)
	if dual_wield != _derived_dual_wield_active:
		_derived_dual_wield_active = dual_wield
		_apply_derived_header_layout(dual_wield)
	var weapon_damage: Dictionary = derived.get("weapon_damage_by_slot", {})
	var main_damage: Dictionary = weapon_damage.get("main_hand", {})
	var off_damage: Dictionary = weapon_damage.get("off_hand", {})
	for key in DERIVED_LABELS.keys():
		var name_label: Label = _derived_name_labels.get(key, null)
		var label: Label = _derived_labels.get(key, null)
		var off_label: Label = _derived_off_labels.get(key, null)
		var tooltip := CharacterStatsBreakdown.breakdown_summary(progression, key, DERIVED_LABELS)
		if name_label != null:
			name_label.text = DERIVED_LABELS[key]
			if dual_wield and key in DUAL_WIELD_DAMAGE_KEYS:
				name_label.tooltip_text = ""
			else:
				name_label.tooltip_text = tooltip
		if label != null:
			if dual_wield and key == "damage_min":
				label.text = CharacterStatsBreakdown.format_stat_value(key, float(main_damage.get("min", derived.get("damage_min", 0.0))))
			elif dual_wield and key == "damage_max":
				label.text = CharacterStatsBreakdown.format_stat_value(key, float(main_damage.get("max", derived.get("damage_max", 0.0))))
			else:
				label.text = CharacterStatsBreakdown.format_stat_value(key, float(derived.get(key, 0.0)))
			if dual_wield and key in DUAL_WIELD_DAMAGE_KEYS:
				var damage_key := "min" if key == "damage_min" else "max"
				label.tooltip_text = CharacterStatsBreakdown.weapon_slot_tooltip(progression, key, main_damage, damage_key, DERIVED_LABELS)
			else:
				label.tooltip_text = tooltip
		if off_label != null:
			if dual_wield and key == "damage_min":
				off_label.visible = true
				off_label.text = CharacterStatsBreakdown.format_stat_value(key, float(off_damage.get("min", 0.0)))
				off_label.tooltip_text = CharacterStatsBreakdown.weapon_slot_tooltip(progression, "damage_min", off_damage, "min", DERIVED_LABELS)
			elif dual_wield and key == "damage_max":
				off_label.visible = true
				off_label.text = CharacterStatsBreakdown.format_stat_value(key, float(off_damage.get("max", 0.0)))
				off_label.tooltip_text = CharacterStatsBreakdown.weapon_slot_tooltip(progression, "damage_max", off_damage, "max", DERIVED_LABELS)
			else:
				off_label.visible = false
				off_label.text = ""
				off_label.tooltip_text = ""
	_render_buttons()


func _on_stat_button_pressed(stat: String) -> void:
	allocate_stat_requested.emit(stat)


func _render_buttons() -> void:
	var points := int(progression.get("unspent_stat_points", 0))
	for stat in BASE_STATS:
		var btn: Button = _stat_buttons.get(stat, null)
		if btn != null:
			btn.disabled = not allocation_enabled or points <= 0


func _dual_wield_damage_active(derived: Dictionary) -> bool:
	var weapon_damage = derived.get("weapon_damage_by_slot", null)
	if typeof(weapon_damage) != TYPE_DICTIONARY:
		return false
	return (weapon_damage as Dictionary).get("off_hand", null) != null


func _apply_derived_header_layout(dual_wield: bool) -> void:
	if _derived_header_name == null or _derived_header_main == null or _derived_header_off == null:
		return
	if dual_wield:
		_derived_header_name.custom_minimum_size = Vector2(120, 22)
		_derived_header_main.text = "MAIN"
		_derived_header_main.custom_minimum_size = Vector2(58, 22)
		_derived_header_off.visible = true
		_derived_header_off.custom_minimum_size = Vector2(58, 22)
		for key in DUAL_WIELD_DAMAGE_KEYS:
			var name_label: Label = _derived_name_labels.get(key, null)
			var label: Label = _derived_labels.get(key, null)
			var off_label: Label = _derived_off_labels.get(key, null)
			if name_label != null:
				name_label.custom_minimum_size = Vector2(120, 28)
			if label != null:
				label.custom_minimum_size = Vector2(58, 28)
			if off_label != null:
				off_label.custom_minimum_size = Vector2(58, 28)
		return
	_derived_header_name.custom_minimum_size = Vector2(188, 22)
	_derived_header_main.text = "VALUE"
	_derived_header_main.custom_minimum_size = Vector2(72, 22)
	_derived_header_off.visible = false
	for key in DUAL_WIELD_DAMAGE_KEYS:
		var name_label: Label = _derived_name_labels.get(key, null)
		var label: Label = _derived_labels.get(key, null)
		var off_label: Label = _derived_off_labels.get(key, null)
		if name_label != null:
			name_label.custom_minimum_size = Vector2(188, 28)
		if label != null:
			label.custom_minimum_size = Vector2(72, 28)
		if off_label != null:
			off_label.custom_minimum_size = Vector2(72, 28)


func _stat_tooltips_by_key() -> Dictionary:
	var out := {}
	for key in _stat_effective_labels.keys():
		var label: Label = _stat_effective_labels.get(key, null)
		if label != null and label.tooltip_text != "":
			out[key] = label.tooltip_text
	return out


func _stat_mouse_filters_by_key() -> Dictionary:
	var out := {}
	for key in _stat_effective_labels.keys():
		var label: Label = _stat_effective_labels.get(key, null)
		if label != null:
			out[key] = int(label.mouse_filter)
	return out


func _derived_tooltips_by_key() -> Dictionary:
	var out := {}
	for key in _derived_labels.keys():
		var label: Label = _derived_labels.get(key, null)
		if label != null and label.tooltip_text != "":
			out[key] = label.tooltip_text
	return out


func _derived_slot_tooltips_by_key(labels_by_key: Dictionary) -> Dictionary:
	var out := {}
	for key in labels_by_key.keys():
		var label: Label = labels_by_key.get(key, null)
		if label != null and label.visible and label.tooltip_text != "":
			out[key] = label.tooltip_text
	return out


func _derived_mouse_filters_by_key() -> Dictionary:
	var out := {}
	for key in _derived_labels.keys():
		var label: Label = _derived_labels.get(key, null)
		if label != null:
			out[key] = int(label.mouse_filter)
	return out


func _derived_scroll_debug_state() -> Dictionary:
	if _derived_scroll == null:
		return {}
	var bar := _derived_scroll.get_v_scroll_bar()
	return {
		"visible": _derived_scroll.visible,
		"vertical_scroll_mode": int(_derived_scroll.vertical_scroll_mode),
		"horizontal_scroll_mode": int(_derived_scroll.horizontal_scroll_mode),
		"scrollbar_on_right": true,
		"scrollbar_visible": bar != null and bar.visible,
		"row_count": _derived_container.get_child_count() if _derived_container != null else 0,
		"viewport_height": _derived_scroll.custom_minimum_size.y,
	}


func _derived_tooltip_panel_debug_state() -> Dictionary:
	for key in _derived_labels.keys():
		var label: Label = _derived_labels.get(key, null)
		if label == null or label.tooltip_text == "":
			continue
		var tooltip = label._make_custom_tooltip(label.tooltip_text)
		var panel := tooltip as PanelContainer
		if panel == null:
			return {"custom": false}
		var style := panel.get_theme_stylebox("panel") as StyleBoxFlat
		var alpha := style.bg_color.a if style != null else 0.0
		panel.queue_free()
		return {
			"custom": true,
			"background_alpha": alpha,
		}
	return {}


func _value_label() -> Label:
	var label := Label.new()
	label.add_theme_color_override("font_color", Color("#d8c7a6"))
	label.add_theme_font_size_override("font_size", 23)
	return label


func _table_value_label(width: float) -> Label:
	var label := _value_label()
	label.custom_minimum_size = Vector2(width, 28)
	label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	return label


func _derived_value_label() -> Label:
	var label := StatTooltipLabelScript.new()
	label.add_theme_color_override("font_color", Color("#d8c7a6"))
	label.add_theme_font_size_override("font_size", 23)
	return label


func _header_label(text: String, width: float, align: HorizontalAlignment) -> Label:
	var label := _section_label(text)
	label.custom_minimum_size = Vector2(width, 22)
	label.horizontal_alignment = align
	label.add_theme_font_size_override("font_size", 15)
	return label


func _section_label(text: String) -> Label:
	var label := _value_label()
	label.text = text
	label.add_theme_color_override("font_color", Color("#c9a227"))
	return label


func _apply_mouse_filter() -> void:
	if _panel != null:
		_panel.mouse_filter = Control.MOUSE_FILTER_STOP if visible else Control.MOUSE_FILTER_IGNORE
