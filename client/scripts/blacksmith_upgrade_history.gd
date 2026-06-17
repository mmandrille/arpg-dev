class_name BlacksmithUpgradeHistory
extends VBoxContainer

const MAX_ENTRIES := 4
const BODY_FONT_SIZE := 13
const DETAIL_FONT_SIZE := 12

var _entries: Array = []
var _title: Label
var _rows: VBoxContainer


func _ready() -> void:
	_build()
	_render()


func record_attempt(recipe_label: String, item_name: String, success: bool, cost_gold: int) -> void:
	if _rows == null:
		_build()
	_entries.push_front({
		"recipe_label": recipe_label,
		"item_name": item_name,
		"result": "Success" if success else "Failed",
		"cost_gold": max(0, cost_gold),
		"text": "%s: %s %s (%d gold)" % [recipe_label, "Success" if success else "Failed", item_name, max(0, cost_gold)],
	})
	while _entries.size() > MAX_ENTRIES:
		_entries.pop_back()
	_render()


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"row_count": _entries.size(),
		"rows": _entries.duplicate(true),
		"text": _debug_text(),
		"max_entries": MAX_ENTRIES,
	}


func _build() -> void:
	if _rows != null:
		return
	add_theme_constant_override("separation", 3)
	_title = Label.new()
	_title.text = "Recent upgrades"
	_title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_title.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	_title.add_theme_color_override("font_color", Color("#d8c8a8"))
	add_child(_title)

	_rows = VBoxContainer.new()
	_rows.add_theme_constant_override("separation", 2)
	add_child(_rows)


func _render() -> void:
	visible = not _entries.is_empty()
	if _rows == null:
		return
	for child in _rows.get_children():
		child.queue_free()
	for entry in _entries:
		var label := Label.new()
		label.text = str((entry as Dictionary).get("text", ""))
		label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
		label.add_theme_color_override("font_color", Color("#b8e6ff"))
		_rows.add_child(label)


func _debug_text() -> String:
	var lines: Array = []
	for entry in _entries:
		lines.append(str((entry as Dictionary).get("text", "")))
	return "\n".join(lines)
