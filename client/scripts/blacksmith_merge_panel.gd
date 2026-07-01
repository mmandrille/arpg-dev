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


func can_place_item_at(index: int, item: Dictionary) -> bool:
	return _can_place_item(index, item)


func place_item_at(index: int, item: Dictionary) -> bool:
	if not _can_place_item(index, item):
		return false
	_place_item(index, item)
	return true


func clear_merge_slot(index: int) -> void:
	_clear_slot(index)


func _build() -> void:
	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 8)
	add_child(root)

	_status_label = Label.new()
	_status_label.text = "Drag or click to place 3 same-level shards or stones"
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
	var button := MergeSlotButton.new()
	button.merge_panel = self
	button.slot_index = index
	button.custom_minimum_size = SLOT_SIZE
	button.focus_mode = Control.FOCUS_NONE
	button.mouse_default_cursor_shape = Control.CURSOR_POINTING_HAND
	button.disabled = false
	return button


func _try_fill_slot(index: int) -> void:
	var candidate := _next_available_consumable()
	if candidate.is_empty():
		show_status("No mergeable consumable in bag", true)
		return
	_place_item(index, candidate)


func _place_item(index: int, item: Dictionary) -> void:
	_slots[index] = item.duplicate(true)
	_refresh_slots()


func _clear_slot(index: int) -> void:
	_slots[index] = {}
	_refresh_slots()


func _next_available_consumable() -> Dictionary:
	var lock := _merge_lock_from_filled()
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
		if not _is_leveled_consumable(row):
			continue
		if lock.is_empty():
			return row
		if str(row.get("item_def_id", "")) != str(lock.get("def_id", "")):
			continue
		if _consumable_level(row) != int(lock.get("level", 0)):
			continue
		return row
	return {}


func _can_place_item(index: int, item: Dictionary) -> bool:
	if index < 0 or index >= _slots.size():
		return false
	if not _is_leveled_consumable(item):
		return false
	var item_instance_id := str(item.get("item_instance_id", ""))
	if item_instance_id == "" or _is_instance_used(item_instance_id, index):
		return false
	var lock := _merge_lock_from_filled()
	if lock.is_empty():
		return true
	if str(item.get("item_def_id", "")) != str(lock.get("def_id", "")):
		return false
	return _consumable_level(item) == int(lock.get("level", 0))


func _is_instance_used(item_instance_id: String, except_index: int) -> bool:
	for slot_index in range(_slots.size()):
		if slot_index == except_index:
			continue
		var slot: Dictionary = _slots[slot_index] if typeof(_slots[slot_index]) == TYPE_DICTIONARY else {}
		if str(slot.get("item_instance_id", "")) == item_instance_id:
			return true
	return false


func _merge_lock_from_filled() -> Dictionary:
	for slot in _slots:
		if typeof(slot) != TYPE_DICTIONARY:
			continue
		var row := slot as Dictionary
		if row.is_empty():
			continue
		return {
			"def_id": str(row.get("item_def_id", "")),
			"level": _consumable_level(row),
		}
	return {}


func _is_leveled_consumable(item: Dictionary) -> bool:
	var def_id := str(item.get("item_def_id", ""))
	return def_id == "upgrade_shard" or def_id == "renew_stone"


func _refresh_slots() -> void:
	var grid := get_child(0).get_child(1) as GridContainer
	for index in range(grid.get_child_count()):
		var button := grid.get_child(index) as MergeSlotButton
		var slot: Dictionary = _slots[index] if index < _slots.size() else {}
		if slot.is_empty():
			button.text = "+"
			button.tooltip_text = "Drag from bag or click to place consumable"
			button.add_theme_color_override("font_color", Color("#8aa4b8"))
			continue
		var level := _consumable_level(slot)
		var def_id := str(slot.get("item_def_id", "consumable"))
		button.text = "Lv%d" % level
		button.tooltip_text = "%s Lv%d" % [def_id.replace("_", " ").capitalize(), level]
		button.remove_theme_color_override("font_color")
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


class MergeSlotButton:
	extends Button

	var merge_panel: BlacksmithMergePanel
	var slot_index: int = 0

	func _can_drop_data(_at_position: Vector2, data: Variant) -> bool:
		if typeof(data) != TYPE_DICTIONARY:
			return false
		if str(data.get("source", "")) != "bag":
			return false
		var item: Dictionary = data.get("item", {})
		return merge_panel._can_place_item(slot_index, item)

	func _drop_data(_at_position: Vector2, data: Variant) -> void:
		var item: Dictionary = data.get("item", {})
		if item.is_empty():
			return
		merge_panel._place_item(slot_index, item)

	func _get_drag_data(_at_position: Vector2) -> Variant:
		if merge_panel == null or slot_index >= merge_panel._slots.size():
			return null
		var slot: Dictionary = merge_panel._slots[slot_index]
		if slot.is_empty():
			return null
		var data := {
			"source": "blacksmith_merge",
			"item": slot.duplicate(true),
			"merge_panel": merge_panel,
			"slot_index": slot_index,
		}
		var preview := Label.new()
		preview.text = str(slot.get("item_def_id", "consumable"))
		preview.add_theme_color_override("font_color", Color("#e8dcc8"))
		set_drag_preview(preview)
		return data

	func _gui_input(event: InputEvent) -> void:
		if event is InputEventMouseButton \
				and event.button_index == MOUSE_BUTTON_LEFT \
				and event.pressed:
			if event.double_click:
				merge_panel._clear_slot(slot_index)
				accept_event()
				return
			if merge_panel._slots[slot_index].is_empty():
				merge_panel._try_fill_slot(slot_index)
				accept_event()
