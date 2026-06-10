extends Control
class_name CharacterSelectPanel

signal back_requested
signal start_requested(character_id: String)
signal create_requested(name: String)
signal delete_requested(character_id: String)
signal rename_requested(character_id: String, name: String)

const MODE_CHOOSE_OR_CREATE := "choose_or_create"
const MODE_FORCED_CREATE := "forced_create"

var _title: Label
var _rows: VBoxContainer
var _empty_label: Label
var _create_row: HBoxContainer
var _name_edit: LineEdit
var _create_button: Button
var _error_label: Label
var _confirm_dialog: ConfirmationDialog
var _rename_dialog: ConfirmationDialog
var _rename_edit: LineEdit
var _characters: Array = []
var _delete_mode: bool = false
var _pending_delete_id: String = ""
var _pending_rename_id: String = ""
var _mode: String = MODE_FORCED_CREATE


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_STOP
	_build()
	visible = false


func show_continue(characters: Array) -> void:
	show_choose_or_create(characters)


func show_new_game() -> void:
	show_forced_create()


func show_choose_or_create(characters: Array, title: String = "Choose Character") -> void:
	_sync_viewport_size()
	visible = true
	_characters = characters.duplicate(true)
	_delete_mode = true
	_mode = MODE_CHOOSE_OR_CREATE
	_title.text = title
	_name_edit.text = ""
	_set_create_entry_expanded(false, "New character name")
	_error_label.text = ""
	_render_characters(characters)


func show_forced_create(title: String = "Create Character") -> void:
	_sync_viewport_size()
	visible = true
	_characters = []
	_delete_mode = false
	_mode = MODE_FORCED_CREATE
	_title.text = title
	_name_edit.text = ""
	_set_create_entry_expanded(true, "Character name")
	_error_label.text = ""
	_render_characters([])
	_focus_name_edit()


func hide_panel() -> void:
	visible = false


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)


func set_error(message: String) -> void:
	_error_label.text = message


func known_characters() -> Array:
	return _characters.duplicate(true)


func mode() -> String:
	return _mode


func title_text() -> String:
	return _title.text if _title != null else ""


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"mode": _mode,
		"title": title_text(),
		"characters": known_characters(),
		"character_rows": _debug_character_rows(),
		"name_field_visible": _name_edit != null and _name_edit.visible,
		"create_button_visible": _create_button != null and _create_button.visible,
		"create_button_text": _create_button.text if _create_button != null else "",
		"empty_visible": _empty_label != null and _empty_label.visible,
		"error": _error_label.text if _error_label != null else "",
	}


func set_name_text(name: String) -> void:
	if _name_edit != null and not _name_edit.visible:
		_expand_create_entry()
	_name_edit.text = name


func submit_name() -> void:
	_on_create_button_pressed()


func start_character_at_index(index: int) -> void:
	if index < 0 or index >= _characters.size():
		return
	var character: Dictionary = _characters[index]
	if bool(character.get("dead", false)):
		return
	start_requested.emit(str(character.get("character_id", "")))


func _build() -> void:
	var bg := ColorRect.new()
	bg.color = Color(0.03, 0.035, 0.04, 0.92)
	bg.set_anchors_preset(Control.PRESET_FULL_RECT)
	add_child(bg)

	var panel := PanelContainer.new()
	panel.custom_minimum_size = Vector2(430, 420)
	panel.set_anchors_preset(Control.PRESET_CENTER)
	panel.offset_left = -215
	panel.offset_right = 215
	panel.offset_top = -210
	panel.offset_bottom = 210
	add_child(panel)

	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 8)
	panel.add_child(box)

	_title = Label.new()
	_title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_title.add_theme_font_size_override("font_size", 36)
	box.add_child(_title)

	_empty_label = Label.new()
	_empty_label.text = "No characters"
	_empty_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	box.add_child(_empty_label)

	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(390, 210)
	box.add_child(scroll)
	_rows = VBoxContainer.new()
	_rows.add_theme_constant_override("separation", 6)
	scroll.add_child(_rows)

	_create_row = HBoxContainer.new()
	_create_row.add_theme_constant_override("separation", 8)
	box.add_child(_create_row)

	_name_edit = LineEdit.new()
	_name_edit.placeholder_text = "Character name"
	_name_edit.max_length = 24
	_name_edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_name_edit.text_submitted.connect(func(_text: String) -> void: _create_from_input())
	_create_row.add_child(_name_edit)

	_create_button = Button.new()
	_create_button.text = "Start"
	_create_button.custom_minimum_size = Vector2(150, 40)
	_create_button.pressed.connect(_on_create_button_pressed)
	_create_row.add_child(_create_button)

	_error_label = Label.new()
	_error_label.add_theme_color_override("font_color", Color("#ff9b7a"))
	_error_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	box.add_child(_error_label)

	var back := Button.new()
	back.text = "Back"
	back.pressed.connect(back_requested.emit)
	box.add_child(back)

	_confirm_dialog = ConfirmationDialog.new()
	_confirm_dialog.title = "Delete character?"
	_confirm_dialog.ok_button_text = "Yes"
	_confirm_dialog.cancel_button_text = "No"
	_confirm_dialog.confirmed.connect(_on_delete_confirmed)
	add_child(_confirm_dialog)

	_rename_dialog = ConfirmationDialog.new()
	_rename_dialog.title = "Rename character"
	_rename_dialog.ok_button_text = "Save"
	_rename_dialog.cancel_button_text = "Cancel"
	_rename_dialog.confirmed.connect(_on_rename_confirmed)
	var rename_box := VBoxContainer.new()
	rename_box.add_theme_constant_override("separation", 8)
	_rename_edit = LineEdit.new()
	_rename_edit.placeholder_text = "Character name"
	_rename_edit.max_length = 24
	_rename_edit.text_submitted.connect(func(_text: String) -> void:
		_on_rename_confirmed()
		_rename_dialog.hide()
	)
	rename_box.add_child(_rename_edit)
	_rename_dialog.add_child(rename_box)
	add_child(_rename_dialog)


func _render_characters(characters: Array) -> void:
	for child in _rows.get_children():
		child.queue_free()
	_empty_label.visible = characters.is_empty() and _mode == MODE_CHOOSE_OR_CREATE and not _name_edit.visible
	for character in characters:
		if typeof(character) != TYPE_DICTIONARY:
			continue
		var row := HBoxContainer.new()
		row.add_theme_constant_override("separation", 6)
		row.custom_minimum_size = Vector2(360, 38)
		var name := str(character.get("name", "Hero"))
		var dead := bool(character.get("dead", false))
		var select_btn := Button.new()
		var label := _character_row_label(character)
		if dead:
			label = "☠  " + label
		select_btn.text = label
		select_btn.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		select_btn.disabled = dead
		select_btn.tooltip_text = _character_row_tooltip(character)
		if dead:
			select_btn.tooltip_text = "Dead\n" + select_btn.tooltip_text
		select_btn.pressed.connect(func() -> void:
			if not dead:
				start_requested.emit(str(character.get("character_id", "")))
		)
		row.add_child(select_btn)
		if _delete_mode:
			var rename_btn := Button.new()
			rename_btn.text = "✎"
			rename_btn.tooltip_text = "Rename character"
			rename_btn.custom_minimum_size = Vector2(38, 38)
			rename_btn.focus_mode = Control.FOCUS_NONE
			var rename_character_id := str(character.get("character_id", ""))
			rename_btn.pressed.connect(func() -> void:
				_prompt_rename(rename_character_id, name)
			)
			row.add_child(rename_btn)

			var delete_btn := Button.new()
			delete_btn.text = "🗑"
			delete_btn.tooltip_text = "Delete character"
			delete_btn.custom_minimum_size = Vector2(38, 38)
			delete_btn.focus_mode = Control.FOCUS_NONE
			var character_id := str(character.get("character_id", ""))
			delete_btn.pressed.connect(func() -> void:
				_prompt_delete(character_id, name)
			)
			row.add_child(delete_btn)
		_rows.add_child(row)


func _set_create_entry_expanded(expanded: bool, placeholder: String) -> void:
	_name_edit.visible = expanded
	_name_edit.placeholder_text = placeholder
	_create_button.visible = true
	if expanded:
		_create_button.text = "✓"
		_create_button.tooltip_text = "Create character"
		_create_button.custom_minimum_size = Vector2(44, 40)
		_create_button.size_flags_horizontal = Control.SIZE_SHRINK_END
	else:
		_create_button.text = "Create Character"
		_create_button.tooltip_text = "Create a new character"
		_create_button.custom_minimum_size = Vector2(150, 40)
		_create_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL


func _expand_create_entry() -> void:
	_set_create_entry_expanded(true, "New character name")
	_error_label.text = ""
	_render_characters(_characters)
	_focus_name_edit()


func _on_create_button_pressed() -> void:
	if _mode == MODE_CHOOSE_OR_CREATE and not _name_edit.visible:
		_expand_create_entry()
		return
	_create_from_input()


func _focus_name_edit() -> void:
	if _name_edit != null and _name_edit.is_inside_tree():
		_name_edit.grab_focus()


func _prompt_delete(character_id: String, character_name: String) -> void:
	if character_id == "":
		return
	_pending_delete_id = character_id
	_confirm_dialog.dialog_text = "Delete %s? This cannot be undone." % character_name
	_confirm_dialog.popup_centered()


func _prompt_rename(character_id: String, character_name: String) -> void:
	if character_id == "":
		return
	_pending_rename_id = character_id
	_rename_edit.text = character_name
	_rename_dialog.popup_centered(Vector2i(320, 120))
	_rename_edit.grab_focus()
	_rename_edit.select_all()


func _on_delete_confirmed() -> void:
	if _pending_delete_id == "":
		return
	delete_requested.emit(_pending_delete_id)
	_pending_delete_id = ""


func _on_rename_confirmed() -> void:
	if _pending_rename_id == "":
		return
	var name := _rename_edit.text.strip_edges()
	if name == "":
		set_error("Name required")
		return
	if name.length() > 24:
		set_error("Name is too long")
		return
	rename_requested.emit(_pending_rename_id, name)
	_pending_rename_id = ""


func _create_from_input() -> void:
	var name := _name_edit.text.strip_edges()
	if name == "":
		set_error("Name required")
		return
	if name.length() > 24:
		set_error("Name is too long")
		return
	create_requested.emit(name)


func _character_level(character: Dictionary) -> int:
	return max(1, int(character.get("level", 1)))


func _character_gold(character: Dictionary) -> int:
	return max(0, int(character.get("gold", 0)))


func _character_depth(character: Dictionary) -> int:
	return max(0, int(character.get("deepest_dungeon_depth", 0)))


func _character_status(character: Dictionary) -> String:
	return "Dead" if bool(character.get("dead", false)) else "Ready"


func _character_row_label(character: Dictionary) -> String:
	var name := str(character.get("name", "Hero"))
	return "%s  Lv %d | %dg | D%d | %s" % [
		name,
		_character_level(character),
		_character_gold(character),
		_character_depth(character),
		_character_status(character),
	]


func _character_row_tooltip(character: Dictionary) -> String:
	var created := str(character.get("created_at", ""))
	var lines := [
		"Level %d" % _character_level(character),
		"Gold %d" % _character_gold(character),
		"Deepest depth %d" % _character_depth(character),
		"Status %s" % _character_status(character),
	]
	if created != "":
		lines.append("Created %s" % created.left(10))
	return "\n".join(lines)


func _debug_character_rows() -> Array:
	var rows := []
	for character in _characters:
		if typeof(character) != TYPE_DICTIONARY:
			continue
		var rec: Dictionary = character
		rows.append({
			"character_id": str(rec.get("character_id", "")),
			"name": str(rec.get("name", "Hero")),
			"dead": bool(rec.get("dead", false)),
			"level": _character_level(rec),
			"gold": _character_gold(rec),
			"deepest_dungeon_depth": _character_depth(rec),
			"status": _character_status(rec),
			"label": _character_row_label(rec),
			"tooltip": _character_row_tooltip(rec),
		})
	return rows
