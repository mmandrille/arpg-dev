class_name CharacterBar
extends Control

const MaterialWalletPanelScript := preload("res://scripts/material_wallet_panel.gd")

signal open_character_requested

var _interactive: bool = true
var _progression: Dictionary = {}
var _panel: PanelContainer
var _slot: Button
var _badge: Label
var _wallet_label: Label
var _cooldown_overlay: ColorRect
var _attack_recovery_remaining: float = 0.0
var _attack_recovery_total: float = 0.0
var _resource_wallet: Dictionary = {}
var _wallet_window: Control


func _ready() -> void:
	ItemRulesLoader.ensure_loaded()
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	set_process(true)
	_build()


func _process(delta: float) -> void:
	if _attack_recovery_remaining <= 0.0:
		return
	_attack_recovery_remaining = maxf(0.0, _attack_recovery_remaining - maxf(0.0, delta))
	_render()


func set_interactive(enabled: bool) -> void:
	_interactive = enabled
	_render()


func set_progression(next_progression: Dictionary) -> void:
	_progression = next_progression.duplicate(true)
	_render()


func set_resource_wallet(next_wallet: Dictionary) -> void:
	_resource_wallet = next_wallet.duplicate(true)
	_render()
	_sync_wallet_window()


func open_slot() -> void:
	if not _interactive:
		return
	open_character_requested.emit()


func open_wallet_window() -> void:
	if _wallet_rows().is_empty():
		return
	if _wallet_window == null:
		_wallet_window = MaterialWalletPanelScript.new()
		add_child(_wallet_window)
	_wallet_window.show_wallet(_resource_wallet)


func start_attack_recovery(duration_seconds: float) -> void:
	_attack_recovery_total = maxf(0.0, duration_seconds)
	_attack_recovery_remaining = _attack_recovery_total
	_render()


func get_debug_state() -> Dictionary:
	return {
		"enabled": _interactive,
		"disabled": not _interactive,
		"unspent_stat_points": _unspent_stat_points(),
		"upgrade_badge_visible": _badge.visible if _badge != null else false,
		"upgrade_badge_text": _badge.text if _badge != null else "",
		"upgrade_badge_position": _vec2_debug(_badge.position if _badge != null else Vector2.ZERO),
		"upgrade_badge_font_size": _badge.label_settings.font_size if _badge != null and _badge.label_settings != null else 0,
		"upgrade_badge_outline_size": _badge.label_settings.outline_size if _badge != null and _badge.label_settings != null else 0,
		"slot_text": _slot.text if _slot != null else "",
		"tooltip_text": _slot.tooltip_text if _slot != null else "",
		"wallet_visible": _wallet_label.visible if _wallet_label != null else false,
		"wallet_text": _wallet_label.text if _wallet_label != null else "",
		"wallet_tooltip": _wallet_label.tooltip_text if _wallet_label != null else "",
		"wallet_rows": _wallet_rows(),
		"wallet_details": _wallet_detail_lines(),
		"wallet_window": _wallet_window.get_debug_state() if _wallet_window != null else {"visible": false, "row_count": 0, "rows": [], "text": ""},
		"attack_recovery_remaining": _attack_recovery_remaining,
		"attack_recovery_total": _attack_recovery_total,
		"attack_recovery_fraction": _attack_recovery_fraction(),
		"cooldown_overlay_visible": _cooldown_overlay.visible if _cooldown_overlay != null else false,
		"cooldown_overlay_size": _vec2_debug(_cooldown_overlay.size if _cooldown_overlay != null else Vector2.ZERO),
		"cooldown_overlay_position": _vec2_debug(_cooldown_overlay.position if _cooldown_overlay != null else Vector2.ZERO),
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

	_cooldown_overlay = ColorRect.new()
	_cooldown_overlay.color = Color(0.18, 0.18, 0.18, 0.62)
	_cooldown_overlay.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_cooldown_overlay.visible = false
	_slot.add_child(_cooldown_overlay)

	_badge = _make_upgrade_badge()
	_slot.add_child(_badge)

	var spacer := Control.new()
	spacer.custom_minimum_size = Vector2(52, 6)
	box.add_child(spacer)

	_wallet_label = Label.new()
	_wallet_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_wallet_label.add_theme_font_size_override("font_size", 13)
	_wallet_label.add_theme_color_override("font_color", Color("#b8e6ff"))
	_wallet_label.clip_text = true
	_wallet_label.custom_minimum_size = Vector2(56, 18)
	_wallet_label.mouse_filter = Control.MOUSE_FILTER_STOP
	_wallet_label.gui_input.connect(_on_wallet_gui_input)
	box.add_child(_wallet_label)

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
	if _cooldown_overlay != null:
		var fraction := _attack_recovery_fraction()
		var slot_size := _slot.custom_minimum_size
		_cooldown_overlay.visible = fraction > 0.0
		_cooldown_overlay.position = Vector2(0.0, slot_size.y * (1.0 - fraction))
		_cooldown_overlay.size = Vector2(slot_size.x, slot_size.y * fraction)
	_render_wallet()


func _unspent_stat_points() -> int:
	return int(_progression.get("unspent_stat_points", 0))


func _attack_recovery_fraction() -> float:
	if _attack_recovery_total <= 0.0:
		return 0.0
	return clampf(_attack_recovery_remaining / _attack_recovery_total, 0.0, 1.0)


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


func _render_wallet() -> void:
	if _wallet_label == null:
		return
	var rows := _wallet_rows()
	_wallet_label.visible = not rows.is_empty()
	_wallet_label.text = "  ".join(rows)
	_wallet_label.tooltip_text = "\n".join(_wallet_detail_lines())


func _sync_wallet_window() -> void:
	if _wallet_window != null:
		_wallet_window.set_wallet(_resource_wallet)


func _on_wallet_gui_input(event: InputEvent) -> void:
	if event is InputEventMouseButton and event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
		open_wallet_window()
		accept_event()


func _wallet_rows() -> Array:
	var rows: Array = []
	for key in _wallet_resource_keys():
		var amount := int(_resource_wallet.get(key, 0))
		rows.append("%s %d" % [_resource_label(str(key)), amount])
	return rows


func _wallet_detail_lines() -> Array:
	var lines: Array = []
	for key in _wallet_resource_keys():
		var resource_id := str(key)
		var amount := int(_resource_wallet.get(resource_id, 0))
		lines.append("%s x%d" % [_resource_name(resource_id), amount])
		var category := _resource_category(resource_id)
		if category != "":
			lines.append("Category: %s" % category)
		lines.append("Stored account-wide")
	return lines


func _wallet_resource_keys() -> Array:
	var out: Array = []
	var keys: Array = _resource_wallet.keys()
	keys.sort()
	for key in keys:
		var amount := int(_resource_wallet.get(key, 0))
		if amount <= 0:
			continue
		out.append(key)
	return out


func _resource_label(resource_id: String) -> String:
	if resource_id == "upgrade_shard":
		return "Upgrade"
	if resource_id == "respec_badge":
		return "Respec"
	if resource_id == "stat_badge":
		return "Stat"
	if resource_id == "skill_badge":
		return "Skill"
	if resource_id == "resurrection_badge":
		return "Resurrect"
	return _resource_name(resource_id)


func _resource_name(resource_id: String) -> String:
	var def := ItemRulesLoader.item_definition(resource_id)
	if def.has("name"):
		return str(def.get("name", ""))
	return resource_id.replace("_", " ").capitalize()


func _resource_category(resource_id: String) -> String:
	var def := ItemRulesLoader.item_definition(resource_id)
	return str(def.get("category", "")).replace("_", " ").capitalize()


func _vec2_debug(v: Vector2) -> Dictionary:
	return {"x": v.x, "y": v.y}


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
