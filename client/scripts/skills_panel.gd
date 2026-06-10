class_name SkillsPanel
extends Control

signal allocate_skill_point_requested(skill_id: String)

const MAGIC_BOLT_ID := "magic_bolt"

var skill_progression: Dictionary = {}
var interactive: bool = true
var _hovered_skill_id: String = ""
var _hover_controls: Array[Control] = []
var _skill_function_keys: Array = []
var _right_click_skill_id: String = ""
var _panel: PanelContainer
var _points_label: Label
var _rank_label: Label
var _spend_button: Button
var _skill_block: Panel
var _skill_icon_label: Label
var _assigned_key_label: Label
var _tooltip: PanelContainer
var _tooltip_rank: Label
var _tooltip_body: Label


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


func set_skill_bindings(function_keys: Array, right_click_skill_id: String) -> void:
	_skill_function_keys = function_keys.duplicate(true)
	_right_click_skill_id = right_click_skill_id
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
		"assigned_key": _assigned_key_for_skill(MAGIC_BOLT_ID),
		"right_click_assigned": _right_click_skill_id == MAGIC_BOLT_ID,
		"tooltip_visible": _tooltip != null and _tooltip.visible,
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
	_show_tooltip(skill_id)


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
	root.add_theme_constant_override("separation", 10)
	root.custom_minimum_size = Vector2(304, 470)
	_panel.add_child(root)

	var title := _label("Skill Tree", 31, Color("#f0dfbb"))
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	root.add_child(title)

	var tree := Control.new()
	tree.custom_minimum_size = Vector2(304, 376)
	tree.mouse_filter = Control.MOUSE_FILTER_IGNORE
	root.add_child(tree)

	var backdrop := ColorRect.new()
	backdrop.color = Color("#151617")
	backdrop.custom_minimum_size = Vector2(304, 376)
	backdrop.mouse_filter = Control.MOUSE_FILTER_IGNORE
	tree.add_child(backdrop)

	var line := ColorRect.new()
	line.color = Color(0.05, 0.045, 0.04, 0.95)
	line.position = Vector2(151, 126)
	line.custom_minimum_size = Vector2(4, 126)
	line.mouse_filter = Control.MOUSE_FILTER_IGNORE
	tree.add_child(line)
	_add_disabled_slot(tree, Vector2(111, 252))
	_add_disabled_slot(tree, Vector2(176, 252))

	_skill_block = Panel.new()
	_skill_block.position = Vector2(112, 54)
	_skill_block.size = Vector2(80, 80)
	_skill_block.mouse_filter = Control.MOUSE_FILTER_STOP
	_skill_block.add_theme_stylebox_override("panel", _skill_block_style(false, false))
	tree.add_child(_skill_block)
	_bind_skill_hover(_skill_block, MAGIC_BOLT_ID)

	_skill_icon_label = _label("M", 42, Color("#d8d1c1"))
	_skill_icon_label.position = Vector2(8, 8)
	_skill_icon_label.size = Vector2(64, 64)
	_skill_icon_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_skill_icon_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	_skill_icon_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_skill_block.add_child(_skill_icon_label)

	_assigned_key_label = _badge_label("")
	_assigned_key_label.position = Vector2(50, 55)
	_assigned_key_label.custom_minimum_size = Vector2(30, 22)
	_skill_block.add_child(_assigned_key_label)

	_rank_label = _badge_label("")
	_rank_label.position = Vector2(196, 78)
	_rank_label.custom_minimum_size = Vector2(40, 28)
	tree.add_child(_rank_label)

	_tooltip = PanelContainer.new()
	_tooltip.visible = false
	_tooltip.position = Vector2(48, 150)
	_tooltip.custom_minimum_size = Vector2(208, 178)
	_tooltip.mouse_filter = Control.MOUSE_FILTER_STOP
	_tooltip.add_theme_stylebox_override("panel", _tooltip_style())
	tree.add_child(_tooltip)
	_bind_skill_hover(_tooltip, MAGIC_BOLT_ID)

	var tip_root := VBoxContainer.new()
	tip_root.add_theme_constant_override("separation", 6)
	tip_root.custom_minimum_size = Vector2(184, 154)
	_tooltip.add_child(tip_root)

	var tooltip_title := _label("Magic Bolt", 21, Color("#f0dfbb"))
	tip_root.add_child(tooltip_title)
	_tooltip_rank = _label("", 16, Color("#cfc3aa"))
	tip_root.add_child(_tooltip_rank)
	_tooltip_body = _label("", 15, Color("#b9ad97"))
	tip_root.add_child(_tooltip_body)
	_spend_button = Button.new()
	_spend_button.text = "+"
	_spend_button.tooltip_text = "Spend skill point"
	_spend_button.focus_mode = Control.FOCUS_NONE
	_spend_button.custom_minimum_size = Vector2(38, 30)
	_spend_button.pressed.connect(_on_spend_pressed)
	tip_root.add_child(_spend_button)
	_bind_skill_hover(_spend_button, MAGIC_BOLT_ID)

	_points_label = _label("", 18, Color("#bfc6c2"))
	_points_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	root.add_child(_points_label)

	_render()


func _render() -> void:
	if _points_label == null or _rank_label == null or _spend_button == null:
		return
	var unspent := int(skill_progression.get("unspent_skill_points", 0))
	var skill := _skill_row(MAGIC_BOLT_ID)
	var rank := int(skill.get("rank", 0))
	var max_rank := int(skill.get("max_rank", 0))
	var unlocked := rank > 0
	_points_label.text = "Skill choices remaining  %d" % unspent
	_rank_label.text = "%d / %d" % [rank, max_rank]
	_spend_button.disabled = not interactive or unspent <= 0 or rank >= max_rank or not bool(skill.get("can_spend", false))
	_tooltip_rank.text = "Rank %d / %d" % [rank, max_rank]
	_tooltip_body.text = "Projectile spell\nMana: %d\nCooldown: attack x2" % (2 + maxi(rank, 1))
	var assigned_key := _assigned_key_for_skill(MAGIC_BOLT_ID)
	_assigned_key_label.text = assigned_key
	_assigned_key_label.visible = assigned_key != ""
	_skill_block.add_theme_stylebox_override("panel", _skill_block_style(unlocked, _right_click_skill_id == MAGIC_BOLT_ID))
	_skill_icon_label.modulate = Color(1, 1, 1, 1) if unlocked else Color(0.42, 0.42, 0.42, 1)
	_rank_label.modulate = Color(1, 1, 1, 1) if unlocked else Color(0.45, 0.45, 0.45, 1)


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
		_show_tooltip(skill_id)
	)
	control.mouse_exited.connect(func() -> void:
		if not _mouse_over_skill_controls():
			_hovered_skill_id = ""
			_hide_tooltip()
	)


func _mouse_over_skill_controls() -> bool:
	var mouse_pos := get_viewport().get_mouse_position()
	for control in _hover_controls:
		if control != null and control.is_inside_tree() and control.get_global_rect().has_point(mouse_pos):
			return true
	return false


func _show_tooltip(skill_id: String) -> void:
	if skill_id != MAGIC_BOLT_ID or _tooltip == null:
		return
	_tooltip.visible = true


func _hide_tooltip() -> void:
	if _tooltip != null:
		_tooltip.visible = false


func _assigned_key_for_skill(skill_id: String) -> String:
	for i in range(_skill_function_keys.size()):
		if str(_skill_function_keys[i]) == skill_id:
			return "F%d" % (i + 1)
	return ""


func _add_disabled_slot(parent: Control, position: Vector2) -> void:
	var slot := PanelContainer.new()
	slot.position = position
	slot.custom_minimum_size = Vector2(40, 40)
	slot.mouse_filter = Control.MOUSE_FILTER_IGNORE
	slot.add_theme_stylebox_override("panel", _disabled_slot_style())
	parent.add_child(slot)


func _badge_label(text: String) -> Label:
	var label := _label(text, 16, Color("#d9d0bd"))
	label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	label.add_theme_stylebox_override("normal", _badge_style())
	return label


func _label(text: String, size: int, color: Color) -> Label:
	var label := Label.new()
	label.text = text
	label.add_theme_font_size_override("font_size", size)
	label.add_theme_color_override("font_color", color)
	return label


func _apply_mouse_filter() -> void:
	if _panel != null:
		_panel.mouse_filter = Control.MOUSE_FILTER_STOP if visible else Control.MOUSE_FILTER_IGNORE


func _panel_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.07, 0.072, 0.07, 0.96)
	s.border_color = Color("#7a6535")
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


func _skill_block_style(unlocked: bool, selected: bool) -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color("#5a4028") if unlocked else Color("#201a16")
	s.border_color = Color("#c9a76a") if selected else (Color("#8a7245") if unlocked else Color("#3b342c"))
	s.border_width_left = 2
	s.border_width_top = 2
	s.border_width_right = 2
	s.border_width_bottom = 2
	s.corner_radius_top_left = 2
	s.corner_radius_top_right = 2
	s.corner_radius_bottom_left = 2
	s.corner_radius_bottom_right = 2
	return s


func _tooltip_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.045, 0.042, 0.036, 0.96)
	s.border_color = Color("#7a6535")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.corner_radius_top_left = 4
	s.corner_radius_top_right = 4
	s.corner_radius_bottom_left = 4
	s.corner_radius_bottom_right = 4
	s.content_margin_left = 10
	s.content_margin_right = 10
	s.content_margin_top = 10
	s.content_margin_bottom = 10
	return s


func _badge_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.025, 0.025, 0.025, 0.92)
	s.border_color = Color("#50463a")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	return s


func _disabled_slot_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.05, 0.046, 0.042, 0.9)
	s.border_color = Color("#2f2c28")
	s.border_width_left = 2
	s.border_width_top = 2
	s.border_width_right = 2
	s.border_width_bottom = 2
	return s
