class_name SkillsPanel
extends Control

signal allocate_skill_point_requested(skill_id: String)

const SkillIconScript := preload("res://scripts/skill_icon.gd")
const DraggableWindowScript := preload("res://scripts/draggable_window.gd")

const SKILL_BLOCK_SIZE := Vector2(83, 83)
const SKILL_ICON_SIZE := Vector2(62, 62)
const SKILL_TREE_ORIGIN := Vector2(23, 70)
const SKILL_TREE_SPACING := Vector2(96, 127)
const SKILL_TREE_WIDTH := 395.0
const SKILL_TOOLTIP_SIZE := Vector2(218, 218)
const SKILL_TOOLTIP_GAP := 8.0

var skill_progression: Dictionary = {}
var character_progression: Dictionary = {}
var interactive: bool = true
var _hovered_skill_id: String = ""
var _hover_controls: Array[Control] = []
var _skill_function_keys: Array = []
var _right_click_skill_id: String = ""
var _selected_skill_id: String = ""
var _panel: DraggableWindow
var _points_label: Label
var _skill_blocks: Dictionary = {}
var _skill_icons: Dictionary = {}
var _skill_rank_labels: Dictionary = {}
var _assigned_key_labels: Dictionary = {}
var _tooltip: PanelContainer
var _tooltip_title: Label
var _tooltip_rank: Label
var _tooltip_body: RichTextLabel
var _tooltip_plain_text: String = ""


static func skill_tooltip_text(skill_id: String, rank: int, max_rank: int, skill_progression: Dictionary, character_progression: Dictionary) -> String:
	return "%s\nRank %d / %d\n%s" % [
		SkillRulesLoader.skill_display_name(skill_id),
		rank,
		max_rank,
		tooltip_plain_body(skill_id, rank, skill_progression, character_progression),
	]


static func tooltip_plain_body(skill_id: String, rank: int, skill_progression: Dictionary, character_progression: Dictionary) -> String:
	var def := SkillRulesLoader.skill_definition(skill_id)
	var summary := SkillRulesLoader.skill_summary(skill_id)
	if summary == "":
		summary = _static_kind_label(def)
	var text := summary
	text += "\nMana: %d" % _static_skill_mana_cost(def, rank)
	text += "\n%s" % _static_skill_cooldown_text(def)
	var next_lines := tooltip_next_rank_lines(skill_id, rank)
	for line in next_lines:
		text += "\n%s" % str(line)
	var requirement_lines := tooltip_requirement_lines(skill_id, skill_progression, character_progression)
	if not requirement_lines.is_empty():
		text += "\n\nRequires:"
		for line in requirement_lines:
			text += "\n%s" % str((line as Dictionary).get("text", ""))
	return text


static func tooltip_next_rank_lines(skill_id: String, rank: int) -> Array:
	var def := SkillRulesLoader.skill_definition(skill_id)
	var max_rank := int(def.get("max_rank", 1))
	var current_rank := maxi(rank, 1)
	if rank <= 0:
		return []
	if current_rank >= max_rank:
		return []
	var next_rank := current_rank + 1
	var lines: Array = []
	var mana_now := _static_skill_mana_cost(def, current_rank)
	var mana_next := _static_skill_mana_cost(def, next_rank)
	if mana_next != mana_now:
		lines.append("Mana: %d -> %d" % [mana_now, mana_next])
	var damage: Dictionary = def.get("damage", {})
	if str(damage.get("type", "")) == "rank_linear_range":
		var min_now := _static_ranked_stat_value(int(damage.get("min_base", 0)), int(damage.get("min_per_rank", 0)), current_rank)
		var max_now := _static_ranked_stat_value(int(damage.get("max_base", 0)), int(damage.get("max_per_rank", 0)), current_rank)
		var min_next := _static_ranked_stat_value(int(damage.get("min_base", 0)), int(damage.get("min_per_rank", 0)), next_rank)
		var max_next := _static_ranked_stat_value(int(damage.get("max_base", 0)), int(damage.get("max_per_rank", 0)), next_rank)
		if min_now != min_next or max_now != max_next:
			lines.append("Damage: %d-%d -> %d-%d" % [min_now, max_now, min_next, max_next])
	var effects: Array = def.get("effects", [])
	for effect in effects:
		if typeof(effect) != TYPE_DICTIONARY:
			continue
		var rec := effect as Dictionary
		if not rec.has("percent_base") and not rec.has("percent_per_rank"):
			continue
		var percent_now := _static_ranked_stat_value(int(rec.get("percent_base", 0)), int(rec.get("percent_per_rank", 0)), current_rank)
		var percent_next := _static_ranked_stat_value(int(rec.get("percent_base", 0)), int(rec.get("percent_per_rank", 0)), next_rank)
		if percent_now == percent_next:
			continue
		lines.append("%s: %d%% -> %d%%" % [_static_effect_label(rec), percent_now, percent_next])
	return lines


static func tooltip_requirement_lines(skill_id: String, skill_progression: Dictionary, character_progression: Dictionary) -> Array:
	var rows := tooltip_requirement_status(skill_id, skill_progression, character_progression)
	if rows.is_empty():
		return []
	var lines: Array = []
	for row in rows:
		var rec := row as Dictionary
		var label := str(rec.get("label", rec.get("stat", "")))
		var required := int(rec.get("required", 0))
		var current := int(rec.get("current", 0))
		var met := bool(rec.get("met", current >= required))
		var suffix := "" if met else "(%d)" % (current - required)
		lines.append({
			"text": "%s %d%s" % [label, required, suffix],
			"color": _static_requirement_color(met),
		})
	return lines


static func tooltip_requirement_status(skill_id: String, skill_progression: Dictionary, character_progression: Dictionary) -> Array:
	var def := SkillRulesLoader.skill_definition(skill_id)
	var requirements: Dictionary = def.get("requirements", {})
	var out: Array = []
	var target_rank := _static_requirement_target_rank(skill_id, skill_progression)
	var level_required := _static_ranked_requirement_value(int(requirements.get("level", 0)), int(requirements.get("level_per_rank", 0)), target_rank)
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
		var required := _static_ranked_requirement_value(int(stats.get(stat, 0)), int(stats_per_rank.get(stat, 0)), target_rank)
		if required <= 0:
			continue
		var current := _static_current_stat_value(character_progression, stat)
		out.append({
			"stat": stat,
			"label": _static_stat_label(stat),
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
		var current_rank := int(_static_skill_row(skill_progression, prereq_id).get("rank", 0))
		out.append({
			"stat": prereq_id,
			"label": SkillRulesLoader.skill_display_name(prereq_id),
			"required": required_rank,
			"current": current_rank,
			"met": current_rank >= required_rank,
		})
	return out


static func _static_skill_row(skill_progression: Dictionary, skill_id: String) -> Dictionary:
	var rows: Array = skill_progression.get("skills", [])
	for row in rows:
		if typeof(row) == TYPE_DICTIONARY and str((row as Dictionary).get("skill_id", "")) == skill_id:
			return row as Dictionary
	return {}


static func _static_requirement_target_rank(skill_id: String, skill_progression: Dictionary) -> int:
	var skill := _static_skill_row(skill_progression, skill_id)
	var rank := int(skill.get("rank", 0))
	var max_rank := int(skill.get("max_rank", int(SkillRulesLoader.skill_definition(skill_id).get("max_rank", 1))))
	if max_rank <= 0:
		max_rank = 1
	if rank >= max_rank:
		return max_rank
	return rank + 1


static func _static_ranked_requirement_value(base: int, per_rank: int, rank: int) -> int:
	return maxi(0, base + per_rank * maxi(0, rank - 1))


static func _static_ranked_stat_value(base: int, per_rank: int, rank: int) -> int:
	return base + per_rank * maxi(0, rank - 1)


static func _static_effect_label(effect: Dictionary) -> String:
	match str(effect.get("type", "")):
		"area_percent_heal":
			return "Heal"
		"stat_percent_buff", "area_stat_percent_buff":
			return "Buff"
		_:
			return "Effect"


static func _static_current_stat_value(character_progression: Dictionary, stat: String) -> int:
	var stats: Dictionary = character_progression.get("base_stats", {})
	return int(stats.get(stat, 0))


static func _static_stat_label(stat: String) -> String:
	match stat:
		"str":
			return TextCatalog.get_text("stat.strength", "Strength")
		"dex":
			return TextCatalog.get_text("stat.dexterity", "Dexterity")
		"vit":
			return TextCatalog.get_text("stat.vitality", "Vitality")
		"magic":
			return "Magic"
		_:
			return stat.capitalize()


static func _static_skill_mana_cost(def: Dictionary, rank: int) -> int:
	var cost: Dictionary = def.get("cost", {})
	var mana: Dictionary = cost.get("mana", {})
	return maxi(0, int(mana.get("base", 0)) + int(mana.get("per_rank", 0)) * maxi(0, rank - 1))


static func _static_skill_cooldown_text(def: Dictionary) -> String:
	var cooldown: Dictionary = def.get("cooldown", {})
	if str(cooldown.get("type", "")) == "attack_interval_multiplier":
		var multiplier := float(cooldown.get("multiplier", 1.0))
		if is_equal_approx(multiplier, roundf(multiplier)):
			return "Cooldown: attack x%d" % int(roundf(multiplier))
		return "Cooldown: attack x%.1f" % multiplier
	return "Cooldown: %s" % str(cooldown.get("type", "none"))


static func _static_kind_label(def: Dictionary) -> String:
	var kind := str(def.get("kind", "skill")).replace("_", " ")
	return kind.capitalize()


static func _static_requirement_color(met: bool) -> String:
	return "#b9d6a3" if met else "#e07a67"


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
	for id in _visible_skill_ids():
		var row_skill_id := str(id)
		var row := _skill_row(row_skill_id)
		var row_requirements := _requirement_status(row_skill_id)
		skill_states[row_skill_id] = {
			"skill_id": row_skill_id,
			"skill_name": _skill_name(row_skill_id),
			"icon_label": _skill_icon_label_text(row_skill_id),
			"icon_shape": _skill_icon_shape(row_skill_id),
			"rank": int(row.get("rank", 0)),
			"max_rank": int(row.get("max_rank", int(_skill_def(row_skill_id).get("max_rank", 0)))),
			"can_spend": bool(row.get("can_spend", false)),
			"spend_button_enabled": _skill_spend_enabled(row_skill_id),
			"visual_state": _skill_visual_state(row_skill_id),
			"assigned_key": _assigned_key_for_skill(row_skill_id),
			"right_click_assigned": _right_click_skill_id == row_skill_id,
			"requirements_met": _requirements_met(row_requirements),
			"requirement_status": row_requirements,
			"tooltip_body": _tooltip_plain_text_for(row_skill_id, int(row.get("rank", 0))),
		}
	return {
		"visible": visible,
		"unspent_skill_points": int(skill_progression.get("unspent_skill_points", 0)),
		"skill_id": skill_id,
		"skill_ids": _visible_skill_ids(),
		"skill_name": _skill_name(skill_id),
		"icon_label": _skill_icon_label_text(skill_id),
		"icon_shape": _skill_icon_shape(skill_id),
		"rank": int(skill.get("rank", 0)),
		"max_rank": int(skill.get("max_rank", int(_skill_def(skill_id).get("max_rank", 0)))),
		"can_spend": bool(skill.get("can_spend", false)),
		"spend_button_enabled": _skill_spend_enabled(skill_id),
		"spend_button_visible": false,
		"visual_state": _skill_visual_state(skill_id),
		"hovered_skill_id": _hovered_skill_id,
		"selected_skill_id": skill_id,
		"assigned_key": _assigned_key_for_skill(skill_id),
		"right_click_assigned": _right_click_skill_id == skill_id,
		"tooltip_visible": _tooltip != null and _tooltip.visible,
		"tooltip_position": _vec2_debug(_tooltip.position if _tooltip != null else Vector2.ZERO),
		"tooltip_mouse_filter": _tooltip.mouse_filter if _tooltip != null else -1,
		"tooltip_body": _tooltip_plain_text,
		"tooltip_rich_text": _tooltip_body.text if _tooltip_body != null else "",
		"points_label_visible": _points_label != null and _points_label.visible,
		"requirements_met": _requirements_met(requirement_status),
		"requirement_status": requirement_status,
		"window": _panel.get_debug_state() if _panel != null else {},
		"skills": skill_states,
	}


func bot_click_skill_button(skill_id: String = "") -> void:
	if skill_id == "":
		skill_id = _current_skill_id()
	_click_skill(skill_id)


func bot_click_close() -> void:
	if _panel != null and _panel.close_button() != null:
		_panel.close_button().pressed.emit()


func bot_drag_window_by(delta: Vector2) -> void:
	if _panel != null:
		_panel.bot_drag_by(delta)


func bot_hover_skill(skill_id: String = "") -> void:
	if skill_id == "":
		skill_id = _current_skill_id()
	if not _select_skill(skill_id):
		return
	_hovered_skill_id = skill_id
	_show_tooltip(skill_id)


func bot_leave_skill_tooltip() -> void:
	_hovered_skill_id = ""
	_hide_tooltip()


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = DraggableWindowScript.new()
	_panel.custom_minimum_size = Vector2(429, 650)
	_panel.position = Vector2(362, 118)
	_panel.configure("Skills", Vector2(395, 567))
	_panel.set_layout_key("skills")
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	_panel.add_theme_stylebox_override("panel", _panel_style())
	_panel.close_requested.connect(hide_display)
	add_child(_panel)

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 10)
	root.custom_minimum_size = Vector2(395, 567)
	_panel.set_content(root)

	var tree := Control.new()
	tree.custom_minimum_size = Vector2(395, 463)
	tree.mouse_filter = Control.MOUSE_FILTER_IGNORE
	root.add_child(tree)

	var backdrop := ColorRect.new()
	backdrop.color = Color("#151617")
	backdrop.custom_minimum_size = Vector2(395, 463)
	backdrop.mouse_filter = Control.MOUSE_FILTER_IGNORE
	tree.add_child(backdrop)

	var line := ColorRect.new()
	line.color = Color(0.05, 0.045, 0.04, 0.95)
	line.position = Vector2(196, 164)
	line.custom_minimum_size = Vector2(5, 164)
	line.mouse_filter = Control.MOUSE_FILTER_IGNORE
	tree.add_child(line)
	_add_disabled_slot(tree, Vector2(144, 328))
	_add_disabled_slot(tree, Vector2(229, 328))

	for raw_skill_id in _tree_skill_ids():
		var skill_id := str(raw_skill_id)
		var skill_block := Panel.new()
		skill_block.position = _skill_block_position(skill_id)
		skill_block.size = SKILL_BLOCK_SIZE
		skill_block.custom_minimum_size = SKILL_BLOCK_SIZE
		skill_block.mouse_filter = Control.MOUSE_FILTER_STOP
		skill_block.add_theme_stylebox_override("panel", _skill_block_style("disabled", false))
		skill_block.gui_input.connect(func(event: InputEvent) -> void:
			if event is InputEventMouseButton and event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
				_click_skill(skill_id)
		)
		tree.add_child(skill_block)
		_bind_skill_hover(skill_block, skill_id)
		_skill_blocks[skill_id] = skill_block

		var icon = SkillIconScript.new()
		icon.position = Vector2(10, 10)
		icon.size = SKILL_ICON_SIZE
		icon.configure(skill_id, _skill_presentation(skill_id))
		skill_block.add_child(icon)
		_skill_icons[skill_id] = icon

		var assigned_key_label := _badge_label("")
		assigned_key_label.position = Vector2(53, 64)
		assigned_key_label.custom_minimum_size = Vector2(30, 19)
		skill_block.add_child(assigned_key_label)
		_assigned_key_labels[skill_id] = assigned_key_label

		var rank_label := _badge_label("")
		rank_label.position = Vector2(0, 64)
		rank_label.custom_minimum_size = Vector2(45, 19)
		skill_block.add_child(rank_label)
		_skill_rank_labels[skill_id] = rank_label

	_tooltip = PanelContainer.new()
	_tooltip.visible = false
	_tooltip.custom_minimum_size = SKILL_TOOLTIP_SIZE
	_tooltip.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_tooltip.add_theme_stylebox_override("panel", _tooltip_style())
	tree.add_child(_tooltip)

	var tip_root := VBoxContainer.new()
	tip_root.add_theme_constant_override("separation", 6)
	tip_root.custom_minimum_size = Vector2(184, 154)
	_tooltip.add_child(tip_root)

	_tooltip_title = _label(_skill_name(_current_skill_id()), 21, Color("#f0dfbb"))
	tip_root.add_child(_tooltip_title)
	_tooltip_rank = _label("", 16, Color("#cfc3aa"))
	tip_root.add_child(_tooltip_rank)
	_tooltip_body = RichTextLabel.new()
	_tooltip_body.bbcode_enabled = true
	_tooltip_body.fit_content = true
	_tooltip_body.scroll_active = false
	_tooltip_body.custom_minimum_size = Vector2(184, 96)
	_tooltip_body.add_theme_font_size_override("normal_font_size", 15)
	_tooltip_body.add_theme_color_override("default_color", Color("#b9ad97"))
	tip_root.add_child(_tooltip_body)
	_points_label = _label("", 18, Color("#bfc6c2"))
	_points_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	root.add_child(_points_label)

	_render()


func _render() -> void:
	if _points_label == null:
		return
	if _selected_skill_id == "" or not _skill_is_visible(_selected_skill_id):
		_selected_skill_id = str(_visible_skill_ids().front()) if not _visible_skill_ids().is_empty() else SkillRulesLoader.first_skill_id()
	var unspent := int(skill_progression.get("unspent_skill_points", 0))
	var skill_id := _current_skill_id()
	var skill := _skill_row(skill_id)
	var rank := int(skill.get("rank", 0))
	var max_rank := int(skill.get("max_rank", int(_skill_def(skill_id).get("max_rank", 0))))
	_points_label.text = "Skill choices remaining  %d" % unspent
	if _tooltip_title != null:
		_tooltip_title.text = _skill_name(skill_id)
	_tooltip_rank.text = "Rank %d / %d" % [rank, max_rank]
	_set_tooltip_body(skill_id, rank)
	for raw_skill_id in _tree_skill_ids():
		var row_skill_id := str(raw_skill_id)
		var row_visible := _skill_is_visible(row_skill_id)
		var row := _skill_row(row_skill_id)
		var row_rank := int(row.get("rank", 0))
		var row_max_rank := int(row.get("max_rank", int(_skill_def(row_skill_id).get("max_rank", 0))))
		var visual_state := _skill_visual_state(row_skill_id)
		var selected := row_skill_id == skill_id
		var block := _skill_blocks.get(row_skill_id, null) as Panel
		if block != null:
			block.visible = row_visible
			block.position = _skill_block_position(row_skill_id)
			block.add_theme_stylebox_override("panel", _skill_block_style(visual_state, selected or _right_click_skill_id == row_skill_id))
		var icon = _skill_icons.get(row_skill_id, null)
		if icon != null:
			icon.configure(row_skill_id, _skill_presentation(row_skill_id))
			icon.modulate = _skill_icon_modulate(visual_state, selected)
		var assigned_key_label := _assigned_key_labels.get(row_skill_id, null) as Label
		if assigned_key_label != null:
			var assigned_key := _assigned_key_for_skill(row_skill_id)
			assigned_key_label.text = assigned_key
			assigned_key_label.visible = assigned_key != ""
		var rank_label := _skill_rank_labels.get(row_skill_id, null) as Label
		if rank_label != null:
			rank_label.text = "%d/%d" % [row_rank, row_max_rank]
			rank_label.modulate = _skill_rank_modulate(visual_state, selected)


func _click_skill(skill_id: String) -> void:
	if not _select_skill(skill_id):
		return
	if _skill_spend_enabled(skill_id):
		allocate_skill_point_requested.emit(skill_id)
	else:
		_show_tooltip(skill_id)


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
	_position_tooltip_below_skill(skill_id)
	_tooltip.visible = true
	if _points_label != null:
		_points_label.visible = false


func _hide_tooltip() -> void:
	if _tooltip != null:
		_tooltip.visible = false
	if _points_label != null:
		_points_label.visible = true


func _position_tooltip_below_skill(skill_id: String) -> void:
	if _tooltip == null:
		return
	var block := _skill_blocks.get(skill_id, null) as Control
	if block == null:
		return
	var x := clampf(block.position.x, 0.0, maxf(0.0, SKILL_TREE_WIDTH - SKILL_TOOLTIP_SIZE.x))
	var y := block.position.y + SKILL_BLOCK_SIZE.y + SKILL_TOOLTIP_GAP
	_tooltip.position = Vector2(x, y)


func _assigned_key_for_skill(skill_id: String) -> String:
	for i in range(_skill_function_keys.size()):
		if str(_skill_function_keys[i]) == skill_id:
			return "F%d" % (i + 1)
	return ""


func _current_skill_id() -> String:
	if _selected_skill_id != "" and _skill_is_visible(_selected_skill_id):
		return _selected_skill_id
	var visible_skill_ids := _visible_skill_ids()
	if not visible_skill_ids.is_empty():
		return str(visible_skill_ids.front())
	return SkillRulesLoader.first_skill_id()


func _select_skill(skill_id: String) -> bool:
	if skill_id == "" or not _skill_is_visible(skill_id):
		return false
	_selected_skill_id = skill_id
	_render()
	return true


func _tree_skill_ids() -> Array:
	return SkillRulesLoader.skill_ids_by_tree()


func _visible_skill_ids() -> Array:
	var character_class := str(character_progression.get("character_class", ""))
	if character_class == "":
		return _tree_skill_ids()
	var out: Array = []
	for raw_skill_id in _tree_skill_ids():
		var skill_id := str(raw_skill_id)
		if str(_skill_def(skill_id).get("class", "")) == character_class:
			out.append(skill_id)
	return out


func _skill_is_visible(skill_id: String) -> bool:
	if skill_id == "" or SkillRulesLoader.skill_definition(skill_id).is_empty():
		return false
	var character_class := str(character_progression.get("character_class", ""))
	return character_class == "" or str(_skill_def(skill_id).get("class", "")) == character_class


func _skill_block_position(skill_id: String) -> Vector2:
	var tree: Dictionary = _skill_def(skill_id).get("tree", {})
	var tier := maxi(1, int(tree.get("tier", 1)))
	var visible_ids := _visible_skill_ids()
	var column := maxi(1, int(tree.get("column", 1)))
	if not visible_ids.is_empty():
		var visible_index := visible_ids.find(skill_id)
		if visible_index >= 0:
			column = visible_index + 1
	var row_ids: Array = []
	for raw_skill_id in visible_ids:
		var row_skill_id := str(raw_skill_id)
		var row_tree: Dictionary = _skill_def(row_skill_id).get("tree", {})
		if maxi(1, int(row_tree.get("tier", 1))) == tier:
			row_ids.append(row_skill_id)
	var row_count := row_ids.size()
	var centered_offset := 0.0
	if row_count > 0:
		var row_width := (float(row_count - 1) * SKILL_TREE_SPACING.x) + SKILL_BLOCK_SIZE.x
		centered_offset = maxf(0.0, (SKILL_TREE_WIDTH - row_width) * 0.5)
		var row_index := row_ids.find(skill_id)
		if row_index >= 0:
			column = row_index + 1
	return Vector2(SKILL_TREE_ORIGIN.x + centered_offset + (column - 1) * SKILL_TREE_SPACING.x, SKILL_TREE_ORIGIN.y + (tier - 1) * SKILL_TREE_SPACING.y)


func _skill_spend_enabled(skill_id: String) -> bool:
	var skill := _skill_row(skill_id)
	var rank := int(skill.get("rank", 0))
	var max_rank := int(skill.get("max_rank", int(_skill_def(skill_id).get("max_rank", 0))))
	return interactive \
		and int(skill_progression.get("unspent_skill_points", 0)) > 0 \
		and rank < max_rank \
		and bool(skill.get("can_spend", false))


func _skill_visual_state(skill_id: String) -> String:
	var rank := int(_skill_row(skill_id).get("rank", 0))
	if _skill_spend_enabled(skill_id):
		return "highlight"
	if rank <= 0:
		return "disabled"
	return "normal"


func _skill_def(skill_id: String) -> Dictionary:
	return SkillRulesLoader.skill_definition(skill_id)


func _skill_presentation(skill_id: String) -> Dictionary:
	return SkillRulesLoader.skill_presentation(skill_id)


func _skill_name(skill_id: String) -> String:
	return SkillRulesLoader.skill_display_name(skill_id)


func _skill_icon_label_text(skill_id: String) -> String:
	var presentation := _skill_presentation(skill_id)
	var icon: Dictionary = presentation.get("icon", {})
	return str(icon.get("label", skill_id.substr(0, 1).to_upper()))


func _skill_icon_color(skill_id: String) -> Color:
	var presentation := _skill_presentation(skill_id)
	var icon: Dictionary = presentation.get("icon", {})
	return Color(str(icon.get("accent", "#d8d1c1")))


func _skill_icon_shape(skill_id: String) -> String:
	var presentation := _skill_presentation(skill_id)
	var icon: Dictionary = presentation.get("icon", {})
	return str(icon.get("shape", "bolt"))


func _set_tooltip_body(skill_id: String, rank: int) -> void:
	_tooltip_plain_text = _tooltip_plain_text_for(skill_id, rank)
	if _tooltip_body != null:
		_tooltip_body.text = _tooltip_rich_text_for(skill_id, rank)


func _tooltip_plain_text_for(skill_id: String, rank: int) -> String:
	return tooltip_plain_body(skill_id, rank, skill_progression, character_progression)


func _tooltip_rich_text_for(skill_id: String, rank: int) -> String:
	var def := _skill_def(skill_id)
	var presentation := _skill_presentation(skill_id)
	var summary := SkillRulesLoader.skill_summary(skill_id)
	if summary == "":
		summary = _kind_label(def)
	var lines: Array[String] = [
		_escape_bbcode(summary),
		_escape_bbcode("Mana: %d" % _skill_mana_cost(def, rank)),
		_escape_bbcode(_skill_cooldown_text(def)),
	]
	var next_lines := tooltip_next_rank_lines(skill_id, rank)
	for line in next_lines:
		lines.append(_next_rank_rich_line(str(line)))
	var requirement_lines := _requirement_lines(skill_id)
	if not requirement_lines.is_empty():
		lines.append("")
		lines.append(_escape_bbcode("Requires:"))
		for line in requirement_lines:
			var rec := line as Dictionary
			var text := _escape_bbcode(str(rec.get("text", "")))
			var color := str(rec.get("color", "#b9ad97"))
			lines.append("[color=%s]%s[/color]" % [color, text])
	return "\n".join(lines)


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


func _requirement_lines(skill_id: String) -> Array:
	return tooltip_requirement_lines(skill_id, skill_progression, character_progression)


func _requirement_color(met: bool) -> String:
	return "#b9d6a3" if met else "#e07a67"


func _escape_bbcode(text: String) -> String:
	return text.replace("[", "\\[").replace("]", "\\]")


func _next_rank_rich_line(text: String) -> String:
	var marker := " -> "
	var marker_index := text.find(marker)
	if marker_index < 0:
		return _escape_bbcode(text)
	var left := text.substr(0, marker_index)
	var right := text.substr(marker_index)
	return "%s[color=#6fd66f]%s[/color]" % [_escape_bbcode(left), _escape_bbcode(right)]



func _requirement_status(skill_id: String) -> Array:
	return tooltip_requirement_status(skill_id, skill_progression, character_progression)


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
			return TextCatalog.get_text("stat.strength", "Strength")
		"dex":
			return TextCatalog.get_text("stat.dexterity", "Dexterity")
		"vit":
			return TextCatalog.get_text("stat.vitality", "Vitality")
		"magic":
			return "Magic"
		_:
			return stat.capitalize()


func _add_disabled_slot(parent: Control, position: Vector2) -> void:
	var slot := PanelContainer.new()
	slot.position = position
	slot.custom_minimum_size = Vector2(52, 52)
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


func _skill_block_style(visual_state: String, selected: bool) -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	match visual_state:
		"highlight":
			s.bg_color = Color("#8d681f")
			s.border_color = Color("#ffe08a")
		"normal":
			s.bg_color = Color("#5a4028")
			s.border_color = Color("#c9a76a") if selected else Color("#8a7245")
		_:
			s.bg_color = Color("#151515")
			s.border_color = Color("#4a4a4a") if selected else Color("#2f2f2f")
	if selected:
		match visual_state:
			"highlight":
				s.border_color = Color("#f0dfbb")
			"normal":
				s.border_color = Color("#c9a76a")
			_:
				s.border_color = Color("#5a5a5a")
	var border_width := 3 if visual_state == "highlight" else 2
	s.border_width_left = border_width
	s.border_width_top = border_width
	s.border_width_right = border_width
	s.border_width_bottom = border_width
	s.corner_radius_top_left = 2
	s.corner_radius_top_right = 2
	s.corner_radius_bottom_left = 2
	s.corner_radius_bottom_right = 2
	return s


func _skill_icon_modulate(visual_state: String, selected: bool) -> Color:
	match visual_state:
		"highlight":
			return Color(1.35, 1.24, 0.82, 1)
		"normal":
			return Color(1, 1, 1, 1)
		_:
			return Color(0.34, 0.34, 0.34, 1) if not selected else Color(0.55, 0.55, 0.55, 1)


func _skill_rank_modulate(visual_state: String, selected: bool) -> Color:
	match visual_state:
		"highlight":
			return Color(1, 0.93, 0.72, 1)
		"normal":
			return Color(1, 1, 1, 1)
		_:
			return Color(0.38, 0.38, 0.38, 1) if not selected else Color(0.55, 0.55, 0.55, 1)


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


func _vec2_debug(value: Vector2) -> Dictionary:
	return {"x": value.x, "y": value.y}


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
