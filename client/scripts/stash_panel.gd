class_name StashPanel
extends Control

signal intent_requested(intent_type: String, payload: Dictionary)

const StatLabels := preload("res://scripts/stat_labels.gd")
const ItemIconDrawerScript := preload("res://scripts/item_icon_drawer.gd")
const ItemTooltipPanelScript := preload("res://scripts/item_tooltip_panel.gd")
const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const PANEL_SIZE := Vector2(390, 520)
const COLUMNS := 5
const STASH_VISIBLE_ROWS := 6
const SLOT_SIZE := Vector2(50, 50)
const SLOT_GAP := 6
const TITLE_FONT_SIZE := 31
const BODY_FONT_SIZE := 21
const DETAIL_FONT_SIZE := 18
const ICON_FONT_SIZE := 18
const SORT_ACQUIRED := "acquired"
const SORT_NAME := "name"
const SORT_RARITY := "rarity"
const SORT_SLOT := "slot"
const SORT_MODES := [SORT_ACQUIRED, SORT_NAME, SORT_RARITY, SORT_SLOT]
const DRAG_SOURCE_INVENTORY_BAG := "bag"
const DRAG_SOURCE_STASH := "stash"
const DRAG_SOURCE_CORPSE := "corpse"
const DRAG_SOURCE_UNIQUE_CHEST := "unique_chest"
const ITEM_RARITY_BACKGROUNDS := {
	"common": Color("#343432"),
	"magic": Color("#1b3458"),
	"rare": Color("#5a4520"),
	"unique": Color("#5a2f17"),
}

var stash_entity_id: String = ""
var stash_id: String = "account_stash"
var stash_title: String = "Account Stash"
var stash_items: Array = []
var stash_gold: int = 0
var stash_capacity: int = 50
var container_mode: String = "stash"
var container_label: String = "Stash"
var inventory: Array = []
var equipped: Dictionary = {}
var hotbar: Array = []
var gold: int = 0
var item_rules: Dictionary:
	get: return ItemRulesLoader.item_rules
var item_templates: Dictionary:
	get: return ItemRulesLoader.item_templates
var item_presentations: Dictionary:
	get: return ItemRulesLoader.item_presentations

var _panel: DraggableWindow
var _title_label: Label
var _gold_label: Label
var _status_label: Label
var _search_field: LineEdit
var _sort_option: OptionButton
var _section_title_label: Label
var _stash_grid: GridContainer
var _withdraw_buttons: Dictionary = {}
var _deposit_gold_button: Button
var _withdraw_gold_button: Button
var _gold_amount_bar: HBoxContainer
var _gold_amount_label: Label
var _gold_amount_input: LineEdit
var _gold_amount_ok_button: Button
var _gold_amount_mode: String = ""
var _interactive: bool = true
var _search_text: String = ""
var _sort_mode: String = SORT_ACQUIRED


class StashSlotButton:
	extends Button

	var panel: StashPanel
	var item: Dictionary = {}
	var slot_kind: String = "stash"

	func _gui_input(event: InputEvent) -> void:
		if not panel._interactive:
			return
		if event is InputEventMouseButton \
				and event.button_index == MOUSE_BUTTON_LEFT \
				and event.pressed \
				and event.double_click \
				and slot_kind == "stash" \
				and not item.is_empty():
			panel._emit_withdraw(item)
			accept_event()

	func _draw() -> void:
		if item.is_empty():
			return
		panel._draw_item_icon(self, item)

	func _get_drag_data(_at_position: Vector2) -> Variant:
		if not panel._interactive or item.is_empty() or slot_kind != "stash":
			return null
		var data := {
			"source": panel._drag_source_for_container(),
			"stash_entity_id": panel.stash_entity_id,
			"stash_item_id": str(item.get("stash_item_id", "")),
			"corpse_entity_id": panel.stash_entity_id if panel.container_mode == "corpse" else "",
			"item_instance_id": str(item.get("item_instance_id", item.get("stash_item_id", ""))),
			"item": item,
		}
		var preview := Label.new()
		preview.text = str(item.get("display_name", item.get("item_def_id", "item")))
		preview.add_theme_color_override("font_color", Color("#e8dcc8"))
		set_drag_preview(preview)
		return data

	func _can_drop_data(_at_position: Vector2, data: Variant) -> bool:
		if not panel._interactive or typeof(data) != TYPE_DICTIONARY or slot_kind != "stash":
			return false
		var source := str((data as Dictionary).get("source", ""))
		var dragged: Dictionary = (data as Dictionary).get("item", {})
		return (source == DRAG_SOURCE_INVENTORY_BAG or source.begins_with("equip:")) \
			and panel.stash_entity_id != "" \
			and not dragged.is_empty() \
			and str(dragged.get("item_instance_id", "")) != ""

	func _drop_data(_at_position: Vector2, data: Variant) -> void:
		panel._handle_drop_on_stash(data)

	func _make_custom_tooltip(for_text: String) -> Object:
		if panel == null:
			return null
		if item.is_empty():
			return panel._make_text_tooltip(for_text)
		return panel._make_item_tooltip(item)


func _ready() -> void:
	ItemRulesLoader.ensure_loaded()
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_build()
	visible = false


func show_stash(next_entity_id: String, next_stash_id: String, next_items: Array, next_stash_gold: int, next_capacity: int, next_inventory: Array, next_equipped: Dictionary, next_gold: int, next_hotbar: Array = [], next_title: String = "Account Stash") -> void:
	stash_entity_id = next_entity_id
	stash_id = next_stash_id
	stash_title = next_title
	container_mode = "stash"
	container_label = "Stash"
	set_stash_state(next_items, next_stash_gold, next_capacity)
	set_inventory_state(next_inventory, next_equipped, next_gold, next_hotbar)
	visible = true
	_apply_interaction_filters()
	_render()


func show_corpse(next_entity_id: String, corpse_name: String, corpse_items: Array, next_inventory: Array, next_equipped: Dictionary, next_gold: int, next_hotbar: Array = []) -> void:
	stash_entity_id = next_entity_id
	stash_id = "hero_corpse"
	stash_title = "%s's Body" % corpse_name if corpse_name != "" else "Hero Body"
	container_mode = "corpse"
	container_label = "Corpse"
	var mapped_items := []
	for item in corpse_items:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := (item as Dictionary).duplicate(true)
		rec["stash_item_id"] = str(rec.get("item_instance_id", rec.get("stash_item_id", "")))
		mapped_items.append(rec)
	set_stash_state(mapped_items, 0, mapped_items.size())
	set_inventory_state(next_inventory, next_equipped, next_gold, next_hotbar)
	visible = true
	_apply_interaction_filters()
	_render()


func show_unique_chest(next_entity_id: String, chest_items: Array, next_inventory: Array, next_equipped: Dictionary, next_gold: int, next_hotbar: Array = []) -> void:
	stash_entity_id = next_entity_id
	stash_id = "unique_test_chest"
	stash_title = "Unique Chest"
	container_mode = "unique_chest"
	container_label = "Chest"
	set_stash_state(chest_items, 0, chest_items.size())
	set_inventory_state(next_inventory, next_equipped, next_gold, next_hotbar)
	visible = true
	_apply_interaction_filters()
	_render()


func set_stash_state(next_items: Array, next_stash_gold: int, next_capacity: int) -> void:
	stash_items = _dup_array(next_items)
	stash_gold = max(0, next_stash_gold)
	stash_capacity = max(0, next_capacity)
	if _panel != null:
		_render()


func set_inventory_state(next_inventory: Array, next_equipped: Dictionary, next_gold: int, next_hotbar: Array = []) -> void:
	inventory = _dup_array(next_inventory)
	equipped = next_equipped.duplicate(true)
	hotbar = _dup_array(next_hotbar)
	gold = max(0, next_gold)
	if _panel != null:
		_render()


func hide_display() -> void:
	visible = false
	_apply_interaction_filters()


func set_interactive(enabled: bool) -> void:
	_interactive = enabled
	_apply_interaction_filters()
	_render()


func show_status(text: String, error: bool = false) -> void:
	if _status_label == null:
		return
	_status_label.text = text
	_status_label.add_theme_color_override("font_color", Color("#ff9f7a") if error else Color("#9ee6a8"))


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"stash_id": stash_id,
		"container_mode": container_mode,
		"stash_entity_id": stash_entity_id,
		"gold": gold,
		"stash_gold": stash_gold,
		"stash_capacity": stash_capacity,
		"stash_item_count": stash_items.size(),
		"filtered_stash_item_count": _visible_stash_items().size(),
		"stash_search_text": _search_text,
		"stash_sort_mode": _sort_mode,
		"stash_rows": _debug_stash_rows(),
		"withdraw_buttons": _debug_withdraw_buttons(),
		"deposit_gold_enabled": _deposit_gold_button != null and not _deposit_gold_button.disabled,
		"withdraw_gold_enabled": _withdraw_gold_button != null and not _withdraw_gold_button.disabled,
		"gold_amount_visible": _gold_amount_bar != null and _gold_amount_bar.visible,
		"gold_amount_mode": _gold_amount_mode,
		"gold_amount_text": _gold_amount_input.text if _gold_amount_input != null else "",
		"gold_amount_ok_enabled": _gold_amount_ok_button != null and not _gold_amount_ok_button.disabled,
		"status": _status_label.text if _status_label != null else "",
		"window": _panel.get_debug_state() if _panel != null else {},
	}


func bot_drag_bag_to_stash(item_def_id: String = "", rolled: Variant = null, bag_index: int = 0) -> void:
	var matches := _matching_inventory_items(item_def_id, rolled)
	if bag_index < 0 or bag_index >= matches.size():
		return
	_emit_deposit(matches[bag_index])


func bot_drag_stash_to_bag(stash_item_id: String = "", item_def_id: String = "", rolled: Variant = null, stash_index: int = 0) -> void:
	var matches := _matching_stash_items(stash_item_id, item_def_id, rolled)
	if stash_index < 0 or stash_index >= matches.size():
		return
	_emit_withdraw(matches[stash_index])


func bot_click_deposit_gold(amount: int = 1) -> void:
	_show_gold_amount_entry("deposit")
	if _gold_amount_input != null:
		_gold_amount_input.text = str(amount)
	_confirm_gold_amount()


func bot_click_withdraw_gold(amount: int = 1) -> void:
	_show_gold_amount_entry("withdraw")
	if _gold_amount_input != null:
		_gold_amount_input.text = str(amount)
	_confirm_gold_amount()


func bot_open_deposit_gold() -> void:
	_show_gold_amount_entry("deposit")


func bot_open_withdraw_gold() -> void:
	_show_gold_amount_entry("withdraw")


func bot_set_gold_amount_text(text: String) -> void:
	if _gold_amount_input != null:
		_gold_amount_input.text = text
		_update_gold_amount_ok_state()


func bot_confirm_gold_amount() -> void:
	_confirm_gold_amount()


func bot_set_search_text(text: String) -> void:
	_search_text = text.strip_edges()
	if _search_field != null and _search_field.text != _search_text:
		_search_field.text = _search_text
	if _panel != null:
		_render()


func bot_select_sort_mode(mode: String) -> void:
	_set_sort_mode(mode)


func bot_click_close() -> void:
	if _panel != null and _panel.close_button() != null:
		_panel.close_button().pressed.emit()


func bot_drag_window_by(delta: Vector2) -> void:
	if _panel != null:
		_panel.bot_drag_by(delta)


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_reposition_panel()


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = DraggableWindowScript.new()
	_panel.custom_minimum_size = PANEL_SIZE
	_panel.configure(stash_title, Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 58))
	_panel.add_theme_stylebox_override("panel", _panel_style())
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	_panel.close_requested.connect(hide_display)
	add_child(_panel)
	_reposition_panel()
	_panel.set_layout_key("stash")

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 7)
	root.custom_minimum_size = Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 58)
	_panel.set_content(root)

	var header := HBoxContainer.new()
	header.add_theme_constant_override("separation", 10)
	root.add_child(header)
	_title_label = Label.new()
	_title_label.text = stash_title
	_title_label.visible = false
	_title_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_title_label.add_theme_color_override("font_color", Color("#f4d481"))
	_title_label.add_theme_font_size_override("font_size", TITLE_FONT_SIZE)

	_gold_label = Label.new()
	_gold_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	_gold_label.add_theme_color_override("font_color", Color("#f4c84f"))
	_gold_label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	header.add_child(_gold_label)

	_status_label = Label.new()
	_status_label.text = ""
	_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_status_label.add_theme_color_override("font_color", Color("#b8aa91"))
	_status_label.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	root.add_child(_status_label)

	var filters := HBoxContainer.new()
	filters.add_theme_constant_override("separation", 8)
	root.add_child(filters)
	_search_field = LineEdit.new()
	_search_field.placeholder_text = "Search stash"
	_search_field.clear_button_enabled = true
	_search_field.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_search_field.text_changed.connect(func(text: String) -> void:
		_search_text = text.strip_edges()
		_render()
	)
	filters.add_child(_search_field)

	_sort_option = OptionButton.new()
	_sort_option.custom_minimum_size = Vector2(112, 32)
	_sort_option.add_item("Acquired")
	_sort_option.add_item("Name")
	_sort_option.add_item("Rarity")
	_sort_option.add_item("Slot")
	_sort_option.item_selected.connect(func(index: int) -> void:
		if index >= 0 and index < SORT_MODES.size():
			_set_sort_mode(SORT_MODES[index])
	)
	filters.add_child(_sort_option)

	_section_title_label = _section_label("Stash")
	root.add_child(_section_title_label)
	var stash_scroll := ScrollContainer.new()
	stash_scroll.custom_minimum_size = Vector2(
		COLUMNS * SLOT_SIZE.x + (COLUMNS - 1) * SLOT_GAP + 18,
		STASH_VISIBLE_ROWS * SLOT_SIZE.y + (STASH_VISIBLE_ROWS - 1) * SLOT_GAP
	)
	root.add_child(stash_scroll)
	_stash_grid = GridContainer.new()
	_stash_grid.columns = COLUMNS
	_stash_grid.add_theme_constant_override("h_separation", SLOT_GAP)
	_stash_grid.add_theme_constant_override("v_separation", SLOT_GAP)
	stash_scroll.add_child(_stash_grid)

	var gold_bar := HBoxContainer.new()
	gold_bar.add_theme_constant_override("separation", 8)
	root.add_child(gold_bar)
	_deposit_gold_button = _gold_button("Deposit")
	_deposit_gold_button.pressed.connect(func() -> void: _show_gold_amount_entry("deposit"))
	gold_bar.add_child(_deposit_gold_button)
	_withdraw_gold_button = _gold_button("Withdraw")
	_withdraw_gold_button.pressed.connect(func() -> void: _show_gold_amount_entry("withdraw"))
	gold_bar.add_child(_withdraw_gold_button)

	_gold_amount_bar = HBoxContainer.new()
	_gold_amount_bar.add_theme_constant_override("separation", 6)
	_gold_amount_bar.visible = false
	root.add_child(_gold_amount_bar)
	_gold_amount_label = _section_label("Amount")
	_gold_amount_label.custom_minimum_size = Vector2(82, 30)
	_gold_amount_bar.add_child(_gold_amount_label)
	_gold_amount_input = LineEdit.new()
	_gold_amount_input.placeholder_text = "0"
	_gold_amount_input.custom_minimum_size = Vector2(0, 32)
	_gold_amount_input.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_gold_amount_input.text_changed.connect(func(_text: String) -> void: _update_gold_amount_ok_state())
	_gold_amount_input.text_submitted.connect(func(_text: String) -> void: _confirm_gold_amount())
	_gold_amount_bar.add_child(_gold_amount_input)
	_gold_amount_ok_button = Button.new()
	_gold_amount_ok_button.text = "OK"
	_gold_amount_ok_button.custom_minimum_size = Vector2(58, 32)
	_gold_amount_ok_button.focus_mode = Control.FOCUS_NONE
	_gold_amount_ok_button.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	_gold_amount_ok_button.pressed.connect(_confirm_gold_amount)
	_gold_amount_bar.add_child(_gold_amount_ok_button)

	_render()


func _render() -> void:
	if _panel == null or _stash_grid == null:
		return
	_withdraw_buttons = {}
	_title_label.text = stash_title
	_panel.configure(stash_title, Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 58))
	_gold_label.text = "%d / %d gold" % [gold, stash_gold] if container_mode == "stash" else "%d gold" % gold
	if _section_title_label != null:
		_section_title_label.text = container_label
	if _search_field != null:
		_search_field.placeholder_text = _search_placeholder()
	if _search_field != null and _search_field.text != _search_text:
		_search_field.text = _search_text
	if _sort_option != null:
		var selected: int = max(0, SORT_MODES.find(_sort_mode))
		if _sort_option.selected != selected:
			_sort_option.select(selected)
	_clear_children(_stash_grid)

	var visible_items := _visible_stash_items()
	var stash_slots: int = max(stash_capacity, stash_items.size())
	for i in range(stash_slots):
		var slot := _slot_button("stash")
		var item: Dictionary = visible_items[i] if i < visible_items.size() else {}
		_fill_slot(slot, item, "stash")
		_stash_grid.add_child(slot)

		if _deposit_gold_button != null:
			_deposit_gold_button.visible = container_mode == "stash"
			_deposit_gold_button.disabled = not _interactive or container_mode != "stash" or stash_entity_id == "" or gold <= 0
		if _withdraw_gold_button != null:
			_withdraw_gold_button.visible = container_mode == "stash"
			_withdraw_gold_button.disabled = not _interactive or container_mode != "stash" or stash_entity_id == "" or stash_gold <= 0
		if _gold_amount_bar != null:
			_gold_amount_bar.visible = container_mode == "stash" and _gold_amount_mode != ""
			if container_mode != "stash":
				_gold_amount_mode = ""
			_update_gold_amount_ok_state()


func _section_label(text: String) -> Label:
	var label := Label.new()
	label.text = text
	label.add_theme_color_override("font_color", Color("#d8c7a6"))
	label.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	return label


func _drag_source_for_container() -> String:
	if container_mode == "corpse":
		return DRAG_SOURCE_CORPSE
	if container_mode == "unique_chest":
		return DRAG_SOURCE_UNIQUE_CHEST
	return DRAG_SOURCE_STASH


func _search_placeholder() -> String:
	if container_mode == "corpse":
		return "Search corpse"
	if container_mode == "unique_chest":
		return "Search chest"
	return "Search stash"


func _gold_button(text: String) -> Button:
	var btn := Button.new()
	btn.text = text
	btn.focus_mode = Control.FOCUS_NONE
	btn.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	btn.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	return btn


func _slot_button(kind: String) -> StashSlotButton:
	var btn := StashSlotButton.new()
	btn.panel = self
	btn.slot_kind = kind
	btn.custom_minimum_size = SLOT_SIZE
	btn.focus_mode = Control.FOCUS_NONE
	btn.clip_text = true
	btn.add_theme_stylebox_override("normal", _slot_style(false))
	btn.add_theme_stylebox_override("hover", _slot_style(true))
	btn.add_theme_stylebox_override("pressed", _slot_style(true))
	btn.add_theme_color_override("font_color", Color("#e8dcc8"))
	btn.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	return btn


func _fill_slot(slot: StashSlotButton, item: Dictionary, kind: String) -> void:
	slot.item = item.duplicate(true)
	slot.slot_kind = kind
	if item.is_empty():
		slot.text = ""
		slot.tooltip_text = ""
		slot.disabled = false
		slot.add_theme_stylebox_override("normal", _slot_style(false))
		slot.add_theme_stylebox_override("hover", _slot_style(true))
		slot.add_theme_stylebox_override("pressed", _slot_style(true))
		slot.queue_redraw()
		return
	var enabled := _interactive and stash_entity_id != ""
	if kind == "stash":
		_withdraw_buttons[str(item.get("stash_item_id", ""))] = {"enabled": enabled}
	slot.text = ""
	slot.tooltip_text = "\n".join(_tooltip_lines(item))
	slot.disabled = not enabled
	var rarity := str(item.get("rarity", "common"))
	slot.add_theme_stylebox_override("normal", _item_slot_style(rarity, false, enabled))
	slot.add_theme_stylebox_override("hover", _item_slot_style(rarity, true, enabled))
	slot.add_theme_stylebox_override("pressed", _item_slot_style(rarity, true, enabled))
	slot.queue_redraw()


func _draw_item_icon(slot: Control, item: Dictionary) -> void:
	var def_id := str(item.get("item_def_id", ""))
	var icon: Dictionary = item_presentations.get(def_id, {}).get("icon", {})
	var rect := Rect2(Vector2.ZERO, slot.size)
	var label := str(icon.get("label", _short_label(def_id)))
	ItemIconDrawerScript.draw(slot, rect, icon, label, slot is Button and (slot as Button).disabled, 0.10, ICON_FONT_SIZE)


func _make_item_tooltip(item: Dictionary) -> Control:
	var tooltip := ItemTooltipPanelScript.new()
	tooltip.setup(
		item,
		item_presentations,
		_tooltip_lines(item),
		_requirement_lines(item),
		[],
		-1,
		true,
		_short_label(str(item.get("item_def_id", "")))
	)
	return tooltip


func _make_text_tooltip(text: String) -> Control:
	var tooltip := ItemTooltipPanelScript.new()
	tooltip.setup({}, item_presentations, [text], [], [], -1, true, "")
	return tooltip


func _emit_deposit(item: Dictionary) -> void:
	if stash_entity_id == "" or item.is_empty():
		return
	if not _inventory_item_stashable(item):
		show_status("can't stash item", true)
		return
	intent_requested.emit("stash_deposit_item_intent", {
		"stash_entity_id": stash_entity_id,
		"item_instance_id": str(item.get("item_instance_id", "")),
	})


func _emit_withdraw(item: Dictionary) -> void:
	if stash_entity_id == "" or item.is_empty():
		return
	if container_mode == "corpse":
		intent_requested.emit("corpse_withdraw_item_intent", {
			"corpse_entity_id": stash_entity_id,
			"item_instance_id": str(item.get("item_instance_id", item.get("stash_item_id", ""))),
		})
		return
	if container_mode == "unique_chest":
		intent_requested.emit("unique_chest_take_item_intent", {
			"chest_entity_id": stash_entity_id,
			"chest_item_id": str(item.get("stash_item_id", "")),
		})
		return
	intent_requested.emit("stash_withdraw_item_intent", {
		"stash_entity_id": stash_entity_id,
		"stash_item_id": str(item.get("stash_item_id", "")),
	})


func _handle_drop_on_stash(data: Variant) -> void:
	if typeof(data) != TYPE_DICTIONARY:
		return
	var rec := data as Dictionary
	var source := str(rec.get("source", ""))
	if source != DRAG_SOURCE_INVENTORY_BAG and not source.begins_with("equip:"):
		return
	var item: Dictionary = rec.get("item", {})
	_emit_deposit(item)


func _show_gold_amount_entry(mode: String) -> void:
	if mode != "deposit" and mode != "withdraw":
		return
	if container_mode != "stash" or stash_entity_id == "":
		return
	var max_amount := gold if mode == "deposit" else stash_gold
	if max_amount <= 0:
		show_status("no gold available", true)
		return
	_gold_amount_mode = mode
	if _gold_amount_label != null:
		_gold_amount_label.text = "Deposit" if mode == "deposit" else "Withdraw"
	if _gold_amount_input != null:
		_gold_amount_input.text = str(max_amount)
		_gold_amount_input.select_all()
		_gold_amount_input.grab_focus()
	if _gold_amount_bar != null:
		_gold_amount_bar.visible = true
	_update_gold_amount_ok_state()


func _confirm_gold_amount() -> void:
	if _gold_amount_mode == "":
		return
	var amount := _gold_amount_input_amount()
	if amount <= 0:
		show_status("enter a gold amount", true)
		_update_gold_amount_ok_state()
		return
	if _gold_amount_mode == "deposit":
		_emit_deposit_gold(amount)
	elif _gold_amount_mode == "withdraw":
		_emit_withdraw_gold(amount)
	_gold_amount_mode = ""
	if _gold_amount_bar != null:
		_gold_amount_bar.visible = false


func _gold_amount_input_amount() -> int:
	if _gold_amount_input == null:
		return 0
	var text := _gold_amount_input.text.strip_edges()
	if text == "":
		return 0
	if not text.is_valid_int():
		return 0
	return int(text)


func _update_gold_amount_ok_state() -> void:
	if _gold_amount_ok_button == null:
		return
	var amount := _gold_amount_input_amount()
	var max_amount := 0
	if _gold_amount_mode == "deposit":
		max_amount = gold
	elif _gold_amount_mode == "withdraw":
		max_amount = stash_gold
	_gold_amount_ok_button.disabled = not _interactive or container_mode != "stash" or stash_entity_id == "" or amount <= 0 or amount > max_amount


func _emit_deposit_gold(amount: int) -> void:
	if stash_entity_id == "" or amount <= 0 or gold < amount:
		show_status("not enough gold", true)
		return
	intent_requested.emit("stash_deposit_gold_intent", {
		"stash_entity_id": stash_entity_id,
		"amount": amount,
	})


func _emit_withdraw_gold(amount: int) -> void:
	if stash_entity_id == "" or amount <= 0 or stash_gold < amount:
		show_status("not enough stash gold", true)
		return
	intent_requested.emit("stash_withdraw_gold_intent", {
		"stash_entity_id": stash_entity_id,
		"amount": amount,
	})


func _inventory_item_stashable(item: Dictionary) -> bool:
	var item_id := str(item.get("item_instance_id", ""))
	if item_id == "" or _is_equipped_instance(item_id) or _is_hotbar_assigned(item_id):
		return false
	return true


func _stashable_items() -> Array:
	var out: Array = []
	for item in inventory:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := item as Dictionary
		if _inventory_item_stashable(rec):
			out.append(rec)
	return out


func _matching_inventory_items(item_def_id: String, rolled: Variant) -> Array:
	var out: Array = []
	for item in _stashable_items():
		var rec := item as Dictionary
		if item_def_id != "" and str(rec.get("item_def_id", "")) != item_def_id:
			continue
		if rolled != null and (str(rec.get("item_template_id", "")) != "") != bool(rolled):
			continue
		out.append(rec)
	out.sort_custom(func(a, b) -> bool:
		return str((a as Dictionary).get("item_instance_id", "")) < str((b as Dictionary).get("item_instance_id", ""))
	)
	return out


func _matching_stash_items(stash_item_id: String, item_def_id: String, rolled: Variant) -> Array:
	var out: Array = []
	var source := stash_items if stash_item_id != "" else _visible_stash_items()
	for item in source:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := item as Dictionary
		if stash_item_id != "" and str(rec.get("stash_item_id", "")) != stash_item_id:
			continue
		if item_def_id != "" and str(rec.get("item_def_id", "")) != item_def_id:
			continue
		if rolled != null and (str(rec.get("item_template_id", "")) != "") != bool(rolled):
			continue
		out.append(rec)
	out.sort_custom(func(a, b) -> bool:
		return str((a as Dictionary).get("stash_item_id", "")) < str((b as Dictionary).get("stash_item_id", ""))
	)
	return out


func _visible_stash_items() -> Array:
	var rows: Array = []
	for item in stash_items:
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := (item as Dictionary).duplicate(true)
		if _stash_item_matches_search(rec):
			rows.append(rec)
	_sort_stash_rows(rows)
	return rows


func _stash_item_matches_search(item: Dictionary) -> bool:
	var needle := _search_text.strip_edges().to_lower()
	if needle == "":
		return true
	var haystack := [
		_item_name(item),
		str(item.get("item_def_id", "")),
		str(item.get("item_template_id", "")),
		str(item.get("rarity", "")),
		str(item.get("slot", "")),
	]
	var summary = item.get("summary_lines", [])
	if typeof(summary) == TYPE_ARRAY:
		for line in summary:
			haystack.append(str(line))
	for value in haystack:
		if str(value).to_lower().find(needle) >= 0:
			return true
	return false


func _sort_stash_rows(rows: Array) -> void:
	rows.sort_custom(func(a, b) -> bool:
		var left := a as Dictionary
		var right := b as Dictionary
		match _sort_mode:
			SORT_NAME:
				return _sort_key(_item_name(left), left) < _sort_key(_item_name(right), right)
			SORT_RARITY:
				var lr := _rarity_rank(str(left.get("rarity", "")))
				var rr := _rarity_rank(str(right.get("rarity", "")))
				if lr == rr:
					return _sort_key(_item_name(left), left) < _sort_key(_item_name(right), right)
				return lr > rr
			SORT_SLOT:
				var ls := str(left.get("slot", ""))
				var rs := str(right.get("slot", ""))
				if ls == rs:
					return _sort_key(_item_name(left), left) < _sort_key(_item_name(right), right)
				return ls < rs
			_:
				return int(left.get("stash_item_id", 0)) < int(right.get("stash_item_id", 0))
	)


func _sort_key(primary: String, item: Dictionary) -> String:
	return "%s:%010d" % [primary.to_lower(), int(item.get("stash_item_id", 0))]


func _rarity_rank(rarity: String) -> int:
	match rarity.to_lower():
		"unique":
			return 4
		"rare":
			return 3
		"magic":
			return 2
		"common":
			return 1
		_:
			return 0


func _set_sort_mode(mode: String) -> void:
	if not SORT_MODES.has(mode):
		return
	_sort_mode = mode
	if _sort_option != null:
		var selected := SORT_MODES.find(mode)
		if selected >= 0 and _sort_option.selected != selected:
			_sort_option.select(selected)
	if _panel != null:
		_render()


func _is_equipped_instance(item_instance_id: String) -> bool:
	for slot in equipped.keys():
		var equipped_id = equipped.get(slot, null)
		if equipped_id != null and str(equipped_id) == item_instance_id:
			return true
	return false


func _is_hotbar_assigned(item_instance_id: String) -> bool:
	for slot in hotbar:
		if typeof(slot) == TYPE_DICTIONARY and str((slot as Dictionary).get("item_instance_id", "")) == item_instance_id:
			return true
	return false


func _debug_stash_rows() -> Array:
	var out: Array = []
	for item in _visible_stash_items():
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := item as Dictionary
		out.append(_debug_item_row(rec, "stash"))
	return out


func _debug_item_row(item: Dictionary, kind: String) -> Dictionary:
	return {
		"item_instance_id": str(item.get("item_instance_id", "")),
		"stash_item_id": str(item.get("stash_item_id", "")),
		"item_def_id": str(item.get("item_def_id", "")),
		"item_template_id": str(item.get("item_template_id", "")),
		"display_name": _item_name(item),
		"rarity": str(item.get("rarity", "")),
		"slot": str(item.get("slot", "")),
		"kind": kind,
		"summary_lines": _tooltip_lines(item),
	}


func _debug_withdraw_buttons() -> Dictionary:
	var out := {}
	for item in _visible_stash_items():
		if typeof(item) != TYPE_DICTIONARY:
			continue
		var rec := item as Dictionary
		out[str(rec.get("stash_item_id", ""))] = {"enabled": _interactive and stash_entity_id != ""}
	return out


func _tooltip_lines(row: Dictionary) -> Array:
	var lines: Array = [_item_name(row)]
	var rarity := str(row.get("rarity", ""))
	if rarity != "":
		lines.append("Rarity: %s" % rarity.capitalize())
	var summary = row.get("summary_lines", [])
	if typeof(summary) == TYPE_ARRAY:
		for line in summary:
			var text := str(line)
			if text != "":
				lines.append(text)
	if lines.size() == 1:
		var slot := str(row.get("slot", ""))
		if slot != "":
			lines.append("Slot: %s" % slot)
		lines.append_array(_stat_lines(row.get("rolled_stats", {})))
		var req = row.get("requirements", {})
		if typeof(req) == TYPE_DICTIONARY and int((req as Dictionary).get("level", 0)) > 0:
			lines.append("Requires level %d" % int((req as Dictionary).get("level", 0)))
	return lines


func _requirement_lines(row: Dictionary) -> Array:
	var req = row.get("requirements", {})
	if typeof(req) != TYPE_DICTIONARY:
		return []
	var status_rows = row.get("requirement_status", [])
	var status_by_stat := {}
	if typeof(status_rows) == TYPE_ARRAY:
		for entry in status_rows:
			if typeof(entry) != TYPE_DICTIONARY:
				continue
			var rec := entry as Dictionary
			status_by_stat[str(rec.get("stat", ""))] = rec
	var out: Array = []
	for key in (req as Dictionary).keys():
		var stat := str(key)
		var required := int((req as Dictionary).get(stat, 0))
		if required <= 0:
			continue
		var status: Dictionary = status_by_stat.get(stat, {})
		var current := int(status.get("current", 0))
		var met := bool(status.get("met", true))
		var text := "%s %d" % [StatLabels.display_name(stat), required]
		if status_by_stat.has(stat):
			text = "%s %d/%d" % [StatLabels.display_name(stat), current, required]
		out.append({"text": text, "color": Color("#d8c7a6") if met else Color("#ff8f70")})
	out.sort_custom(func(a, b) -> bool:
		return str((a as Dictionary).get("text", "")) < str((b as Dictionary).get("text", ""))
	)
	return out


func _stat_lines(stats_value: Variant) -> Array:
	if typeof(stats_value) != TYPE_DICTIONARY:
		return []
	var stats := stats_value as Dictionary
	var lines: Array = []
	if int(stats.get("damage_min", 0)) > 0 or int(stats.get("damage_max", 0)) > 0:
		lines.append("Damage %d-%d" % [int(stats.get("damage_min", 0)), int(stats.get("damage_max", 0))])
	for key in ["str", "dex", "vit", "magic", "all_skills", "armor", "block_percent", "attack_speed_percent", "max_hp", "max_mana", "health_regen_per_10_seconds", "mana_regen_per_10_seconds", "skill_damage_percent", "hotbar_slots", "inventory_rows"]:
		var value := int(stats.get(key, 0))
		if value > 0:
			lines.append("%s %s" % [StatLabels.display_name(key), _format_stat_value(key, value)])
	return lines


func _format_stat_value(stat: String, value: int) -> String:
	if stat == "block_percent" or stat == "attack_speed_percent" or stat == "skill_damage_percent":
		return "+%d%%" % value
	return "+%d" % value


func _item_name(item: Dictionary) -> String:
	var name := str(item.get("display_name", ""))
	if name != "":
		return name
	var def_id := str(item.get("item_def_id", ""))
	var def := _item_definition(def_id)
	return str(def.get("name", item.get("item_template_id", def_id if def_id != "" else "item")))


func _short_label(def_id: String) -> String:
	var def: Dictionary = _item_definition(def_id)
	var name := str(def.get("name", def_id))
	var parts := name.split(" ")
	var out := ""
	for part in parts:
		if part.length() > 0:
			out += part.substr(0, 1).to_upper()
	return out.substr(0, 3)


func _item_definition(def_id: String) -> Dictionary:
	if item_rules.has(def_id):
		return item_rules.get(def_id, {})
	return item_templates.get(def_id, {})




func _panel_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.07, 0.06, 0.05, 0.93)
	s.border_color = Color("#6b5420")
	s.border_width_left = 2
	s.border_width_top = 2
	s.border_width_right = 2
	s.border_width_bottom = 2
	s.content_margin_left = 14
	s.content_margin_top = 12
	s.content_margin_right = 14
	s.content_margin_bottom = 12
	return s


func _slot_style(hover: bool) -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color("#3d2e10") if hover else Color("#0a0908")
	s.border_color = Color("#8b6914") if hover else Color("#5c4a1f")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.content_margin_left = 4
	s.content_margin_top = 4
	s.content_margin_right = 4
	s.content_margin_bottom = 4
	return s


func _item_slot_style(rarity: String, hover: bool, enabled: bool) -> StyleBoxFlat:
	var s := _slot_style(hover)
	var base: Color = ITEM_RARITY_BACKGROUNDS.get(rarity.to_lower(), ITEM_RARITY_BACKGROUNDS["common"])
	if not enabled:
		base = base.darkened(0.45)
	s.bg_color = base.lightened(0.12) if hover else base
	s.border_color = base.lightened(0.46) if hover else base.lightened(0.28)
	return s


func _reposition_panel() -> void:
	if _panel == null:
		return
	var margin := 20.0
	var panel_size := _panel.custom_minimum_size
	var viewport_size := get_viewport_rect().size if is_inside_tree() else Vector2(1280, 720)
	var desired := Vector2(margin, margin + 54.0)
	_panel.position = Vector2(
		clampf(desired.x, margin, maxf(margin, viewport_size.x - panel_size.x - margin)),
		clampf(desired.y, margin, maxf(margin, viewport_size.y - panel_size.y - margin))
	)


func _apply_interaction_filters() -> void:
	if _panel == null:
		return
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP if _interactive and visible else Control.MOUSE_FILTER_IGNORE


func _clear_children(node: Node) -> void:
	for child in node.get_children():
		child.queue_free()


func _dup_array(rows: Array) -> Array:
	var out: Array = []
	for row in rows:
		if typeof(row) == TYPE_DICTIONARY:
			out.append((row as Dictionary).duplicate(true))
		else:
			out.append(row)
	return out
