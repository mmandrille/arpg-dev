class_name TrainingDamageLogPanel
extends Control

const CombatBreakdownFormat := preload("res://scripts/combat_breakdown_format.gd")

const PANEL_SIZE := Vector2(360, 420)
const TOWN_TRAINING_DOLL_DEF_ID := "town_training_doll"

var _panel: PanelContainer
var _close_button: Button
var _scroll: ScrollContainer
var _entries: VBoxContainer
var _entries_data: Array = []
var _user_closed: bool = false


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_build()
	hide_display()


func hide_display() -> void:
	visible = false


func bot_click_close() -> void:
	_on_close_pressed()


func is_open() -> bool:
	return visible


func get_entry_count() -> int:
	return _entries_data.size()


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"entry_count": _entries_data.size(),
		"user_closed": _user_closed,
	}


func on_training_doll_combat_event(event: Dictionary, monster_def_id: String) -> void:
	if monster_def_id != TOWN_TRAINING_DOLL_DEF_ID:
		return
	if not _event_has_breakdown(event):
		return
	_entries_data.append(event.duplicate(true))
	if not visible:
		if _user_closed:
			_user_closed = false
		_show_panel()
	_render_entries()


func _event_has_breakdown(event: Dictionary) -> bool:
	var breakdown = event.get("damage_breakdown", [])
	return typeof(breakdown) == TYPE_ARRAY and (breakdown as Array).size() > 0


func _show_panel() -> void:
	visible = true


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = PanelContainer.new()
	_panel.custom_minimum_size = PANEL_SIZE
	_panel.set_anchors_preset(Control.PRESET_TOP_RIGHT)
	_panel.offset_left = -PANEL_SIZE.x - 16.0
	_panel.offset_top = 72.0
	_panel.offset_right = -16.0
	_panel.offset_bottom = 72.0 + PANEL_SIZE.y
	add_child(_panel)
	var margin := MarginContainer.new()
	margin.add_theme_constant_override("margin_left", 10)
	margin.add_theme_constant_override("margin_right", 10)
	margin.add_theme_constant_override("margin_top", 10)
	margin.add_theme_constant_override("margin_bottom", 10)
	_panel.add_child(margin)
	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 8)
	margin.add_child(root)
	var header := HBoxContainer.new()
	root.add_child(header)
	var title := Label.new()
	title.text = "Combat Damage Log"
	title.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	header.add_child(title)
	_close_button = Button.new()
	_close_button.text = "X"
	_close_button.focus_mode = Control.FOCUS_NONE
	_close_button.pressed.connect(_on_close_pressed)
	header.add_child(_close_button)
	_scroll = ScrollContainer.new()
	_scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_scroll.custom_minimum_size = Vector2(PANEL_SIZE.x - 32.0, PANEL_SIZE.y - 72.0)
	root.add_child(_scroll)
	_entries = VBoxContainer.new()
	_entries.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_scroll.add_child(_entries)


func _on_close_pressed() -> void:
	_user_closed = true
	hide_display()


func _render_entries() -> void:
	for child in _entries.get_children():
		child.queue_free()
	for event in _entries_data:
		_entries.add_child(_make_entry(event))


func _make_entry(event: Dictionary) -> PanelContainer:
	var panel := PanelContainer.new()
	var margin := MarginContainer.new()
	margin.add_theme_constant_override("margin_left", 8)
	margin.add_theme_constant_override("margin_right", 8)
	margin.add_theme_constant_override("margin_top", 6)
	margin.add_theme_constant_override("margin_bottom", 6)
	panel.add_child(margin)
	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 4)
	margin.add_child(box)
	var title := Label.new()
	title.text = CombatBreakdownFormat.attack_title(event)
	box.add_child(title)
	var outcome := Label.new()
	outcome.text = CombatBreakdownFormat.outcome_label(event)
	box.add_child(outcome)
	var body := Label.new()
	body.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	body.text = CombatBreakdownFormat.lines_text(event)
	box.add_child(body)
	return panel
