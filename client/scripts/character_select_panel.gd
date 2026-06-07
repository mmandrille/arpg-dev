extends Control
class_name CharacterSelectPanel

signal back_requested
signal start_requested(character_id: String)
signal create_requested(name: String)
signal delete_requested(character_id: String)

var _title: Label
var _rows: VBoxContainer
var _empty_label: Label
var _name_edit: LineEdit
var _create_button: Button
var _error_label: Label
var _confirm_dialog: ConfirmationDialog
var _characters: Array = []
var _delete_mode: bool = false
var _pending_delete_id: String = ""


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_STOP
	_build()
	visible = false


func show_continue(characters: Array) -> void:
	_sync_viewport_size()
	visible = true
	_characters = characters.duplicate(true)
	_delete_mode = true
	_title.text = "Continue"
	_name_edit.visible = false
	_create_button.visible = false
	_error_label.text = ""
	_render_characters(characters)


func show_new_game() -> void:
	_sync_viewport_size()
	visible = true
	_characters = []
	_delete_mode = false
	_title.text = "New Game"
	_name_edit.visible = true
	_create_button.visible = true
	_name_edit.text = ""
	_error_label.text = ""
	_render_characters([])
	_name_edit.grab_focus()


func hide_panel() -> void:
	visible = false


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)


func set_error(message: String) -> void:
	_error_label.text = message


func known_characters() -> Array:
	return _characters.duplicate(true)


func set_name_text(name: String) -> void:
	_name_edit.text = name


func submit_name() -> void:
	_create_from_input()


func start_character_at_index(index: int) -> void:
	if index < 0 or index >= _characters.size():
		return
	var character: Dictionary = _characters[index]
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
	_title.add_theme_font_size_override("font_size", 24)
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

	_name_edit = LineEdit.new()
	_name_edit.placeholder_text = "Character name"
	_name_edit.max_length = 24
	_name_edit.text_submitted.connect(func(_text: String) -> void: _create_from_input())
	box.add_child(_name_edit)

	_create_button = Button.new()
	_create_button.text = "Start"
	_create_button.pressed.connect(_create_from_input)
	box.add_child(_create_button)

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


func _render_characters(characters: Array) -> void:
	for child in _rows.get_children():
		child.queue_free()
	_empty_label.visible = characters.is_empty() and not _name_edit.visible
	for character in characters:
		if typeof(character) != TYPE_DICTIONARY:
			continue
		var row := HBoxContainer.new()
		row.add_theme_constant_override("separation", 6)
		row.custom_minimum_size = Vector2(360, 38)
		var name := str(character.get("name", "Hero"))
		var created := str(character.get("created_at", ""))
		var select_btn := Button.new()
		select_btn.text = name if created == "" else "%s  %s" % [name, created.left(10)]
		select_btn.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		select_btn.pressed.connect(func() -> void:
			start_requested.emit(str(character.get("character_id", "")))
		)
		row.add_child(select_btn)
		if _delete_mode:
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


func _prompt_delete(character_id: String, character_name: String) -> void:
	if character_id == "":
		return
	_pending_delete_id = character_id
	_confirm_dialog.dialog_text = "Delete %s? This cannot be undone." % character_name
	_confirm_dialog.popup_centered()


func _on_delete_confirmed() -> void:
	if _pending_delete_id == "":
		return
	delete_requested.emit(_pending_delete_id)
	_pending_delete_id = ""


func _create_from_input() -> void:
	var name := _name_edit.text.strip_edges()
	if name == "":
		set_error("Name required")
		return
	if name.length() > 24:
		set_error("Name is too long")
		return
	create_requested.emit(name)
