class_name QuestJournalPanel
extends Control

const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const PANEL_SIZE := Vector2(320, 170)

var _panel: DraggableWindow
var _summary_label: Label
var _rows: VBoxContainer
var _objectives: Array = []


func _ready() -> void:
	_build()


func set_objectives(objectives: Array) -> void:
	_objectives = objectives.duplicate(true)
	_refresh()


func toggle() -> void:
	if _panel == null:
		_build()
	visible = not visible


func hide_display() -> void:
	visible = false


func ensure_display_visible() -> void:
	if _panel == null:
		_build()
	visible = true
	_panel.clamp_to_viewport()


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"objectives": _objectives.duplicate(true),
		"count": _objectives.size(),
		"summary": _summary_label.text if _summary_label != null else "",
		"window": _panel.get_debug_state() if _panel != null else {},
	}


func _build() -> void:
	if _panel != null:
		return
	visible = false
	_panel = DraggableWindowScript.new()
	_panel.configure("Quest Journal", PANEL_SIZE)
	_panel.custom_minimum_size = Vector2(PANEL_SIZE.x, PANEL_SIZE.y + DraggableWindowScript.TITLEBAR_HEIGHT)
	_panel.size = _panel.custom_minimum_size
	_panel.set_layout_key("quest_journal_panel")
	_panel.position = Vector2(64, 112)
	_panel.close_requested.connect(hide_display)
	add_child(_panel)

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 10)
	_panel.set_content(root)
	_summary_label = Label.new()
	_summary_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_summary_label.add_theme_font_size_override("font_size", 15)
	_summary_label.add_theme_color_override("font_color", Color("#f0dfbb"))
	root.add_child(_summary_label)
	_rows = VBoxContainer.new()
	_rows.add_theme_constant_override("separation", 6)
	root.add_child(_rows)
	_refresh()


func _refresh() -> void:
	if _rows == null:
		return
	for child in _rows.get_children():
		_rows.remove_child(child)
		child.queue_free()
	if _objectives.is_empty():
		_summary_label.text = "No active quests on this floor."
		return
	_summary_label.text = "Current floor objectives"
	for objective in _objectives:
		_rows.add_child(_objective_label(objective as Dictionary))


func _objective_label(objective: Dictionary) -> Label:
	var label := Label.new()
	var complete := bool(objective.get("complete", false))
	label.text = "%s %s" % ["Done:" if complete else "Active:", str(objective.get("title", "Quest objective"))]
	label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	label.add_theme_font_size_override("font_size", 14)
	label.add_theme_color_override("font_color", Color("#9ddc8f") if complete else Color("#79b8ff"))
	return label
