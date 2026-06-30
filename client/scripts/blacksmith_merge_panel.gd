class_name BlacksmithMergePanel
extends Control

signal merge_requested(item_instance_ids: Array)

const SLOT_SIZE := Vector2(44, 44)
const GRID_SIZE := 5

var bag_items: Array = []
var _slots: Array = []
var _status_label: Label
var _merge_button: Button


func _ready() -> void:
	ItemRulesLoader.ensure_loaded()
	_build()


func set_bag_items(next_items: Array) -> void:
	bag_items = next_items.duplicate(true)
	_refresh_slots()


func set_stash_items(next_items: Array) -> void:
	set_bag_items(next_items)


func get_debug_state() -> Dictionary:
	var filled: Array = []
	for slot in _slots:
		if slot is Dictionary and not str(slot.get("item_instance_id", "")).is_empty():
			filled.append(slot)
	return {
		"filled_count": filled.size(),
		"merge_enabled": _merge_button.disabled == false if _merge_button != null else false,
		"status": _status_label.text if _status_label != null else "",
	}


func show_status(message: String, warning: bool = false) -> void:
	if _status_label == null:
		return
	_status_label.text = message
	_status_label.add_theme_color_override("font_color", Color("#ffcf5a") if warning else Color("#9fd7ff"))


func bot_click_merge() -> void:
	_emit_merge()


func bot_fill_merge_slots(count: int = 3) -> void:
	for index in range(mini(count, GRID_SIZE * GRID_SIZE)):
		_try_fill_slot(index)
	_refresh_slots()


func _build() -> void:
	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 8)
	add_child(root)

	_status_label = Label.new()
	_status_label.text = "Place 3 same-level shards or stones from your bag"
	_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_status_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	root.add_child(_status_label)

	var grid := GridContainer.new()
	grid.columns = GRID_SIZE
	grid.add_theme_constant_override("h_separation", 4)
	grid.add_theme_constant_override("v_separation", 4)
	root.add_child(grid)

	_slots.clear()
	for index in range(GRID_SIZE * GRID_SIZE):
		var slot := _make_slot(index)
		grid.add_child(slot)
		_slots.append({})

	var button_row := CenterContainer.new()
	_merge_button = Button.new()
	_merge_button.text = "Merge"
	_merge_button.custom_minimum_size = Vector2(140, 36)
	_merge_button.disabled = true
	_merge_button.pressed.connect(_emit_merge)
	button_row.add_child(_merge_button)
	root.add_child(button_row)


func _make_slot(index: int) -> Button:
	var button := Button.new()
	button.custom_minimum_size = SLOT_SIZE
	button.text = ""
	button.focus_mode = Control.FOCUS_NONE
	button.set_meta("slot_index", index)
	button.gui_input.connect(func(event: InputEvent) -> void: _on_slot_gui_input(index, event))
	return button


func _on_slot_gui_input(index: int, event: InputEvent) -> void:
	if event is InputEventMouseButton and event.button_index == MOUSE_BUTTON_LEFT and event.pressed:
		if event.double_click:
			_clear_slot(index)
		else:
			_try_fill_slot(index)


func _try_fill_slot(index: int) -> void:
	var candidate := _next_available_consumable()
	if candidate.is_empty():
		show_status("No mergeable consumable in bag", true)
		return
	_slots[index] = candidate.duplicate(true)
	_refresh_slots()


func _clear_slot(index: int) -> void:
	_slots[index] = {}
	_refresh_slots()


func _next_available_consumable() -> Dictionary:
	var used: Dictionary = {}
	for slot in _slots:
		if typeof(slot) != TYPE_DICTIONARY:
			continue
		var item_instance_id := str((slot as Dictionary).get("item_instance_id", ""))
		if item_instance_id != "":
			used[item_instance_id] = true
	for item in bag_items:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var row := item as Dictionary
		var item_instance_id := str(row.get("item_instance_id", ""))
		if item_instance_id == "" or used.has(item_instance_id):
			continue
		return row
	return {}


func _refresh_slots() -> void:
	var grid := get_child(0).get_child(1) as GridContainer
	for index in range(grid.get_child_count()):
		var button := grid.get_child(index) as Button
		var slot: Dictionary = _slots[index] if index < _slots.size() else {}
		if slot.is_empty():
			button.text = ""
			button.tooltip_text = "Click to place consumable"
			continue
		var level := _consumable_level(slot)
		var def_id := str(slot.get("item_def_id", "consumable"))
		button.text = "Lv%d" % level
		button.tooltip_text = "%s Lv%d" % [def_id.replace("_", " ").capitalize(), level]
	_merge_button.disabled = not _can_merge()


func _can_merge() -> bool:
	var filled: Array = []
	for slot in _slots:
		if typeof(slot) == TYPE_DICTIONARY and not (slot as Dictionary).is_empty():
			filled.append(slot)
	if filled.size() != 3:
		return false
	var level := _consumable_level(filled[0])
	var def_id := str((filled[0] as Dictionary).get("item_def_id", ""))
	for slot in filled:
		if _consumable_level(slot) != level:
			return false
		if str((slot as Dictionary).get("item_def_id", "")) != def_id:
			return false
	return true


func _emit_merge() -> void:
	if not _can_merge():
		show_status("Need 3 same consumable type and level", true)
		return
	var ids: Array = []
	for slot in _slots:
		if typeof(slot) != TYPE_DICTIONARY:
			continue
		var item_instance_id := str((slot as Dictionary).get("item_instance_id", ""))
		if item_instance_id != "":
			ids.append(item_instance_id)
	merge_requested.emit(ids)
	for index in range(_slots.size()):
		_slots[index] = {}
	_refresh_slots()


func _consumable_level(item: Dictionary) -> int:
	var rolled = item.get("rolled_stats", {})
	if typeof(rolled) != TYPE_DICTIONARY:
		return 1
	var payload := rolled as Dictionary
	if typeof(payload.get("stats", {})) == TYPE_DICTIONARY:
		return maxi(1, int((payload.get("stats", {}) as Dictionary).get("item_level", 1)))
	return maxi(1, int(payload.get("item_level", 1)))
