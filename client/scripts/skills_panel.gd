class_name SkillsPanel
extends Control

signal allocate_skill_point_requested(skill_id: String)

var skill_progression: Dictionary = {}
var character_progression: Dictionary = {}
var interactive: bool = true
var _hovered_skill_id: String = ""
var _hover_controls: Array[Control] = []
var _skill_function_keys: Array = []
var _right_click_skill_id: String = ""
var _selected_skill_id: String = ""
var _panel: PanelContainer
var _points_label: Label
var _spend_button: Button
var _skill_blocks: Dictionary = {}
var _skill_icon_labels: Dictionary = {}
var _skill_rank_labels: Dictionary = {}
var _assigned_key_labels: Dictionary = {}
var _tooltip: PanelContainer
var _tooltip_title: Label
var _tooltip_rank: Label
var _tooltip_body: Label


func _ready() -> void:
	SkillRulesLoader.ensure_loaded()
	_selected_skill_id = _current_skill_id()
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


func set_character_progression(next_progression: Dictionary) -> void:
	character_progression = next_progression.duplicate(true)
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
	var skill_id := _current_skill_id()
	var skill := _skill_row(skill_id)
	var requirement_status := _requirement_status(skill_id)
	var skill_states := {}
	for id in _tree_skill_ids():
		var row_skill_id := str(id)
		var row := _skill_row(row_skill_id)
		var row_requirements := _requirement_status(row_skill_id)
		skill_states[row_skill_id] = {
			"skill_id": row_skill_id,
			"skill_name": _skill_name(row_skill_id),
			"icon_label": _skill_icon_label_text(row_skill_id),
			"rank": int(row.get("rank", 0)),
			"max_rank": int(row.get("max_rank", int(_skill_def(row_skill_id).get("max_rank", 0)))),
			"can_spend": bool(row.get("can_spend", false)),
			"spend_button_enabled": _skill_spend_enabled(row_skill_id),
			"assigned_key": _assigned_key_for_skill(row_skill_id),
			"right_click_assigned": _right_click_skill_id == row_skill_id,
			"requirements_met": _requirements_met(row_requirements),
			"requirement_status": row_requirements,
			"tooltip_body": _tooltip_text(row_skill_id, maxi(int(row.get("rank", 0)), 1)),
		}
	return {
		"visible": visible,
		"unspent_skill_points": int(skill_progression.get("unspent_skill_points", 0)),
		"skill_id": skill_id,
		"skill_ids": _tree_skill_ids(),
		"skill_name": _skill_name(skill_id),
		"icon_label": _skill_icon_label_text(skill_id),
		"rank": int(skill.get("rank", 0)),
		"max_rank": int(skill.get("max_rank", int(_skill_def(skill_id).get("max_rank", 0)))),
		"can_spend": bool(skill.get("can_spend", false)),
		"spend_button_enabled": _spend_button != null and not _spend_button.disabled,
		"hovered_skill_id": _hovered_skill_id,
		"selected_skill_id": skill_id,
		"assigned_key": _assigned_key_for_skill(skill_id),
		"right_click_assigned": _right_click_skill_id == skill_id,
		"tooltip_visible": _tooltip != null and _tooltip.visible,
		"tooltip_body": _tooltip_body.text if _tooltip_body != null else "",
		"requirements_met": _requirements_met(requirement_status),
		"requirement_status": requirement_status,
		"skills": skill_states,
	}


func bot_click_skill_button(skill_id: String = "") -> void:
	if skill_id == "":
		skill_id = _current_skill_id()
	if not _select_skill(skill_id):
		return
	if _spend_button == null or _spend_button.disabled:
		return
	_spend_button.pressed.emit()


func bot_hover_skill(skill_id: String = "") -> void:
	if skill_id == "":
		skill_id = _current_skill_id()
	if not _select_skill(skill_id):
		return
	_hovered_skill_id = skill_id
	_show_tooltip(skill_id)


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = PanelContainer.new()
	_panel.custom_minimum_size = Vector2(330, 500)
	_panel.position = Vector2(362, 118)
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

	for raw_skill_id in _tree_skill_ids():
		var skill_id := str(raw_skill_id)
		var skill_block := Panel.new()
		skill_block.position = _skill_block_position(skill_id)
		skill_block.size = Vector2(80, 80)
		skill_block.mouse_filter = Control.MOUSE_FILTER_STOP
		skill_block.add_theme_stylebox_override("panel", _skill_block_style(false, false))
		skill_block.gui_input.connect(func(event: InputEvent) -> void:
			if event is InputEventMouseButton and event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
				_select_skill(skill_id)
				_show_tooltip(skill_id)
		)
		tree.add_child(skill_block)
		_bind_skill_hover(skill_block, skill_id)
		_skill_blocks[skill_id] = skill_block

		var icon_label := _label(_skill_icon_label_text(skill_id), 42, _skill_icon_color(skill_id))
		icon_label.position = Vector2(8, 8)
		icon_label.size = Vector2(64, 64)
		icon_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		icon_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
		icon_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
		skill_block.add_child(icon_label)
		_skill_icon_labels[skill_id] = icon_label

		var assigned_key_label := _badge_label("")
		assigned_key_label.position = Vector2(50, 55)
		assigned_key_label.custom_minimum_size = Vector2(30, 22)
		skill_block.add_child(assigned_key_label)
		_assigned_key_labels[skill_id] = assigned_key_label

		var rank_label := _badge_label("")
		rank_label.position = Vector2(0, 55)
		rank_label.custom_minimum_size = Vector2(40, 22)
		skill_block.add_child(rank_label)
		_skill_rank_labels[skill_id] = rank_label

	_tooltip = PanelContainer.new()
	_tooltip.visible = false
	_tooltip.position = Vector2(48, 150)
	_tooltip.custom_minimum_size = Vector2(208, 178)
	_tooltip.mouse_filter = Control.MOUSE_FILTER_STOP
	_tooltip.add_theme_stylebox_override("panel", _tooltip_style())
	tree.add_child(_tooltip)
	_hover_controls.append(_tooltip)

	var tip_root := VBoxContainer.new()
	tip_root.add_theme_constant_override("separation", 6)
	tip_root.custom_minimum_size = Vector2(184, 154)
	_tooltip.add_child(tip_root)

	_tooltip_title = _label(_skill_name(_current_skill_id()), 21, Color("#f0dfbb"))
	tip_root.add_child(_tooltip_title)
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
	_hover_controls.append(_spend_button)

	_points_label = _label("", 18, Color("#bfc6c2"))
	_points_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	root.add_child(_points_label)

	_render()


func _render() -> void:
	if _points_label == null or _spend_button == null:
		return
	if _selected_skill_id == "" or SkillRulesLoader.skill_definition(_selected_skill_id).is_empty():
		_selected_skill_id = SkillRulesLoader.first_skill_id()
	var unspent := int(skill_progression.get("unspent_skill_points", 0))
	var skill_id := _current_skill_id()
	var skill := _skill_row(skill_id)
	var rank := int(skill.get("rank", 0))
	var max_rank := int(skill.get("max_rank", int(_skill_def(skill_id).get("max_rank", 0))))
	_points_label.text = "Skill choices remaining  %d" % unspent
	_spend_button.disabled = not _skill_spend_enabled(skill_id)
	_spend_button.tooltip_text = "Spend point in %s" % _skill_name(skill_id)
	if _tooltip_title != null:
		_tooltip_title.text = _skill_name(skill_id)
	_tooltip_rank.text = "Rank %d / %d" % [rank, max_rank]
	_tooltip_body.text = _tooltip_text(skill_id, maxi(rank, 1))
	for raw_skill_id in _tree_skill_ids():
		var row_skill_id := str(raw_skill_id)
		var row := _skill_row(row_skill_id)
		var row_rank := int(row.get("rank", 0))
		var row_max_rank := int(row.get("max_rank", int(_skill_def(row_skill_id).get("max_rank", 0))))
		var unlocked := row_rank > 0
		var selected := row_skill_id == skill_id
		var block := _skill_blocks.get(row_skill_id, null) as Panel
		if block != null:
			block.add_theme_stylebox_override("panel", _skill_block_style(unlocked, selected or _right_click_skill_id == row_skill_id))
		var icon_label := _skill_icon_labels.get(row_skill_id, null) as Label
		if icon_label != null:
			icon_label.text = _skill_icon_label_text(row_skill_id)
			icon_label.add_theme_color_override("font_color", _skill_icon_color(row_skill_id))
			icon_label.modulate = Color(1, 1, 1, 1) if unlocked or selected else Color(0.5, 0.5, 0.5, 1)
		var assigned_key_label := _assigned_key_labels.get(row_skill_id, null) as Label
		if assigned_key_label != null:
			var assigned_key := _assigned_key_for_skill(row_skill_id)
			assigned_key_label.text = assigned_key
			assigned_key_label.visible = assigned_key != ""
		var rank_label := _skill_rank_labels.get(row_skill_id, null) as Label
		if rank_label != null:
			rank_label.text = "%d/%d" % [row_rank, row_max_rank]
			rank_label.modulate = Color(1, 1, 1, 1) if unlocked or selected else Color(0.45, 0.45, 0.45, 1)


func _on_spend_pressed() -> void:
	if _spend_button == null or _spend_button.disabled:
		return
	allocate_skill_point_requested.emit(_current_skill_id())


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
	if not _select_skill(skill_id) or _tooltip == null:
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


func _current_skill_id() -> String:
	if _selected_skill_id != "" and not SkillRulesLoader.skill_definition(_selected_skill_id).is_empty():
		return _selected_skill_id
	return SkillRulesLoader.first_skill_id()


func _select_skill(skill_id: String) -> bool:
	if skill_id == "" or SkillRulesLoader.skill_definition(skill_id).is_empty():
		return false
	_selected_skill_id = skill_id
	_render()
	return true


func _tree_skill_ids() -> Array:
	return SkillRulesLoader.skill_ids_by_tree()


func _skill_block_position(skill_id: String) -> Vector2:
	var tree: Dictionary = _skill_def(skill_id).get("tree", {})
	var tier := maxi(1, int(tree.get("tier", 1)))
	var column := maxi(1, int(tree.get("column", 1)))
	return Vector2(18 + (column - 1) * 94, 54 + (tier - 1) * 98)


func _skill_spend_enabled(skill_id: String) -> bool:
	var skill := _skill_row(skill_id)
	var rank := int(skill.get("rank", 0))
	var max_rank := int(skill.get("max_rank", int(_skill_def(skill_id).get("max_rank", 0))))
	return interactive \
		and int(skill_progression.get("unspent_skill_points", 0)) > 0 \
		and rank < max_rank \
		and bool(skill.get("can_spend", false))


func _skill_def(skill_id: String) -> Dictionary:
	return SkillRulesLoader.skill_definition(skill_id)


func _skill_presentation(skill_id: String) -> Dictionary:
	return SkillRulesLoader.skill_presentation(skill_id)


func _skill_name(skill_id: String) -> String:
	var def := _skill_def(skill_id)
	return str(def.get("name", skill_id))


func _skill_icon_label_text(skill_id: String) -> String:
	var presentation := _skill_presentation(skill_id)
	var icon: Dictionary = presentation.get("icon", {})
	return str(icon.get("label", skill_id.substr(0, 1).to_upper()))


func _skill_icon_color(skill_id: String) -> Color:
	var presentation := _skill_presentation(skill_id)
	var icon: Dictionary = presentation.get("icon", {})
	return Color(str(icon.get("accent", "#d8d1c1")))


func _tooltip_text(skill_id: String, rank: int) -> String:
	var def := _skill_def(skill_id)
	var presentation := _skill_presentation(skill_id)
	var summary := str(presentation.get("summary", _kind_label(def)))
	var text := summary
	text += "\nMana: %d" % _skill_mana_cost(def, rank)
	text += "\n%s" % _skill_cooldown_text(def)
	var requirement_line := _requirement_line(skill_id)
	if requirement_line != "":
		text += "\n%s" % requirement_line
	return text


func _kind_label(def: Dictionary) -> String:
	var kind := str(def.get("kind", "skill")).replace("_", " ")
	return kind.capitalize()


func _skill_mana_cost(def: Dictionary, rank: int) -> int:
	var cost: Dictionary = def.get("cost", {})
	var mana: Dictionary = cost.get("mana", {})
	return maxi(0, int(mana.get("base", 0)) + int(mana.get("per_rank", 0)) * maxi(0, rank - 1))


func _skill_cooldown_text(def: Dictionary) -> String:
	var cooldown: Dictionary = def.get("cooldown", {})
	if str(cooldown.get("type", "")) == "attack_interval_multiplier":
		var multiplier := float(cooldown.get("multiplier", 1.0))
		if is_equal_approx(multiplier, roundf(multiplier)):
			return "Cooldown: attack x%d" % int(roundf(multiplier))
		return "Cooldown: attack x%.1f" % multiplier
	return "Cooldown: %s" % str(cooldown.get("type", "none"))


func _requirement_line(skill_id: String) -> String:
	var rows := _requirement_status(skill_id)
	if rows.is_empty():
		return ""
	var labels: Array[String] = []
	for row in rows:
		var rec := row as Dictionary
		var label := str(rec.get("label", rec.get("stat", "")))
		var required := int(rec.get("required", 0))
		var current := int(rec.get("current", 0))
		if bool(rec.get("met", false)):
			labels.append("%s %d" % [label, required])
		else:
			labels.append("%s %d (%d)" % [label, required, current])
	return "Requires %s" % ", ".join(labels)


func _requirement_status(skill_id: String) -> Array:
	var def := _skill_def(skill_id)
	var requirements: Dictionary = def.get("requirements", {})
	var out: Array = []
	var target_rank := _requirement_target_rank(skill_id)
	var level_required := _ranked_requirement_value(int(requirements.get("level", 0)), int(requirements.get("level_per_rank", 0)), target_rank)
	if level_required > 0:
		var current_level := int(character_progression.get("level", 1))
		out.append({
			"stat": "level",
			"label": "Level",
			"required": level_required,
			"current": current_level,
			"met": current_level >= level_required,
		})
	var stats: Dictionary = requirements.get("stats", {})
	var stats_per_rank: Dictionary = requirements.get("stats_per_rank", {})
	for stat in ["str", "dex", "vit", "magic"]:
		if not stats.has(stat) and not stats_per_rank.has(stat):
			continue
		var required := _ranked_requirement_value(int(stats.get(stat, 0)), int(stats_per_rank.get(stat, 0)), target_rank)
		if required <= 0:
			continue
		var current := _current_stat_value(stat)
		out.append({
			"stat": stat,
			"label": _stat_label(stat),
			"required": required,
			"current": current,
			"met": current >= required,
		})
	var skills: Array = requirements.get("skills", [])
	for prereq in skills:
		if typeof(prereq) != TYPE_DICTIONARY:
			continue
		var rec := prereq as Dictionary
		var prereq_id := str(rec.get("skill_id", ""))
		var required_rank := int(rec.get("rank", 0))
		if prereq_id == "" or required_rank <= 0:
			continue
		var current_rank := int(_skill_row(prereq_id).get("rank", 0))
		out.append({
			"stat": prereq_id,
			"label": _skill_name(prereq_id),
			"required": required_rank,
			"current": current_rank,
			"met": current_rank >= required_rank,
		})
	return out


func _requirement_target_rank(skill_id: String) -> int:
	var skill := _skill_row(skill_id)
	var rank := int(skill.get("rank", 0))
	var max_rank := int(skill.get("max_rank", int(_skill_def(skill_id).get("max_rank", 1))))
	if max_rank <= 0:
		max_rank = 1
	if rank >= max_rank:
		return max_rank
	return rank + 1


func _ranked_requirement_value(base: int, per_rank: int, rank: int) -> int:
	return maxi(0, base + per_rank * maxi(0, rank - 1))


func _requirements_met(rows: Array) -> bool:
	for row in rows:
		if typeof(row) == TYPE_DICTIONARY and not bool((row as Dictionary).get("met", false)):
			return false
	return true


func _current_stat_value(stat: String) -> int:
	var stats: Dictionary = character_progression.get("base_stats", {})
	return int(stats.get(stat, 0))


func _stat_label(stat: String) -> String:
	match stat:
		"str":
			return "Strength"
		"dex":
			return "Dexterity"
		"vit":
			return "Vitality"
		"magic":
			return "Magic"
		_:
			return stat.capitalize()


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
