class_name StashPanel
extends Control

signal intent_requested(intent_type: String, payload: Dictionary)

const StatLabels := preload("res://scripts/stat_labels.gd")
const PANEL_SIZE := Vector2(390, 520)
const COLUMNS := 5
const STASH_VISIBLE_ROWS := 6
const SLOT_SIZE := Vector2(50, 50)
const SLOT_GAP := 6
const TITLE_FONT_SIZE := 31
const BODY_FONT_SIZE := 21
const DETAIL_FONT_SIZE := 18
const ICON_FONT_SIZE := 18
const DRAG_SOURCE_INVENTORY_BAG := "bag"
const DRAG_SOURCE_STASH := "stash"
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

var _panel: PanelContainer
var _title_label: Label
var _gold_label: Label
var _status_label: Label
var _stash_grid: GridContainer
var _withdraw_buttons: Dictionary = {}
var _deposit_gold_button: Button
var _withdraw_gold_button: Button
var _interactive: bool = true


class StashSlotButton:
	extends Button

	var panel: StashPanel
	var item: Dictionary = {}
	var slot_kind: String = "stash"

	func _draw() -> void:
		if item.is_empty():
			return
		panel._draw_item_icon(self, item)

	func _get_drag_data(_at_position: Vector2) -> Variant:
		if not panel._interactive or item.is_empty() or slot_kind != "stash":
			return null
		var data := {
			"source": DRAG_SOURCE_STASH,
			"stash_entity_id": panel.stash_entity_id,
			"stash_item_id": str(item.get("stash_item_id", "")),
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
		return source == DRAG_SOURCE_INVENTORY_BAG \
			and panel.stash_entity_id != "" \
			and not dragged.is_empty() \
			and str(dragged.get("item_instance_id", "")) != ""

	func _drop_data(_at_position: Vector2, data: Variant) -> void:
		panel._handle_drop_on_stash(data)


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
	set_stash_state(next_items, next_stash_gold, next_capacity)
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
		"stash_entity_id": stash_entity_id,
		"gold": gold,
		"stash_gold": stash_gold,
		"stash_capacity": stash_capacity,
		"stash_item_count": stash_items.size(),
		"stash_rows": _debug_stash_rows(),
		"withdraw_buttons": _debug_withdraw_buttons(),
		"deposit_gold_enabled": _deposit_gold_button != null and not _deposit_gold_button.disabled,
		"withdraw_gold_enabled": _withdraw_gold_button != null and not _withdraw_gold_button.disabled,
		"status": _status_label.text if _status_label != null else "",
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
	_emit_deposit_gold(amount)


func bot_click_withdraw_gold(amount: int = 1) -> void:
	_emit_withdraw_gold(amount)


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_reposition_panel()


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = PanelContainer.new()
	_panel.custom_minimum_size = PANEL_SIZE
	_panel.set_anchors_preset(Control.PRESET_TOP_LEFT)
	_panel.add_theme_stylebox_override("panel", _panel_style())
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	add_child(_panel)
	_reposition_panel()

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 7)
	root.custom_minimum_size = Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 24)
	_panel.add_child(root)

	var header := HBoxContainer.new()
	header.add_theme_constant_override("separation", 10)
	root.add_child(header)
	_title_label = Label.new()
	_title_label.text = stash_title
	_title_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_title_label.add_theme_color_override("font_color", Color("#f4d481"))
	_title_label.add_theme_font_size_override("font_size", TITLE_FONT_SIZE)
	header.add_child(_title_label)

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

	root.add_child(_section_label("Stash"))
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
	_deposit_gold_button = _gold_button("Deposit 1")
	_deposit_gold_button.pressed.connect(func() -> void: _emit_deposit_gold(1))
	gold_bar.add_child(_deposit_gold_button)
	_withdraw_gold_button = _gold_button("Withdraw 1")
	_withdraw_gold_button.pressed.connect(func() -> void: _emit_withdraw_gold(1))
	gold_bar.add_child(_withdraw_gold_button)

	_render()


func _render() -> void:
	if _panel == null or _stash_grid == null:
		return
	_withdraw_buttons = {}
	_title_label.text = stash_title
	_gold_label.text = "%d / %d gold" % [gold, stash_gold]
	_clear_children(_stash_grid)

	var stash_slots = max(stash_capacity, stash_items.size())
	for i in range(stash_slots):
		var slot := _slot_button("stash")
		var item: Dictionary = stash_items[i] if i < stash_items.size() else {}
		_fill_slot(slot, item, "stash")
		_stash_grid.add_child(slot)

	if _deposit_gold_button != null:
		_deposit_gold_button.disabled = not _interactive or stash_entity_id == "" or gold <= 0
	if _withdraw_gold_button != null:
		_withdraw_gold_button.disabled = not _interactive or stash_entity_id == "" or stash_gold <= 0


func _section_label(text: String) -> Label:
	var label := Label.new()
	label.text = text
	label.add_theme_color_override("font_color", Color("#d8c7a6"))
	label.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	return label


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
	var shape := str(icon.get("shape", "box"))
	var color := Color(str(icon.get("color", "#d8d0bd")))
	var accent := Color(str(icon.get("accent", "#6b5420")))
	if slot is Button and (slot as Button).disabled:
		color = color.darkened(0.35)
		accent = accent.darkened(0.35)
	var rect := Rect2(Vector2.ZERO, slot.size)
	var center := rect.get_center()
	var min_side = min(rect.size.x, rect.size.y)
	var label := str(icon.get("label", _short_label(def_id)))

	match shape:
		"blade":
			var a := center + Vector2(-min_side * 0.22, min_side * 0.22)
			var b := center + Vector2(min_side * 0.22, -min_side * 0.22)
			slot.draw_line(a, b, color, 5.0, true)
			slot.draw_line(a + Vector2(-4, 4), a + Vector2(5, -5), accent, 4.0, true)
		"bow":
			slot.draw_arc(center, min_side * 0.28, -1.25, 1.25, 18, color, 4.0, true)
			slot.draw_line(center + Vector2(min_side * 0.18, -min_side * 0.26), center + Vector2(min_side * 0.18, min_side * 0.26), accent, 2.0, true)
		"badge", "coin":
			slot.draw_circle(center, min_side * 0.24, color)
			slot.draw_arc(center, min_side * 0.17, 0.0, TAU, 18, accent, 2.0, true)
		"leaf":
			var pts := PackedVector2Array([
				center + Vector2(0, -min_side * 0.30),
				center + Vector2(min_side * 0.24, -min_side * 0.02),
				center + Vector2(0, min_side * 0.28),
				center + Vector2(-min_side * 0.24, -min_side * 0.02),
			])
			slot.draw_colored_polygon(pts, color)
			slot.draw_line(center + Vector2(0, -min_side * 0.22), center + Vector2(0, min_side * 0.22), accent, 2.0, true)
		"potion":
			slot.draw_rect(Rect2(center + Vector2(-min_side * 0.13, -min_side * 0.05), Vector2(min_side * 0.26, min_side * 0.28)), color, true)
			slot.draw_rect(Rect2(center + Vector2(-min_side * 0.08, -min_side * 0.22), Vector2(min_side * 0.16, min_side * 0.16)), accent, true)
		_:
			slot.draw_rect(Rect2(center - Vector2(min_side * 0.20, min_side * 0.20), Vector2(min_side * 0.40, min_side * 0.40)), color, true)

	var font := slot.get_theme_default_font()
	var text_size := font.get_string_size(label, HORIZONTAL_ALIGNMENT_LEFT, -1, ICON_FONT_SIZE)
	slot.draw_string(font, center + Vector2(-text_size.x * 0.5, min_side * 0.10), label, HORIZONTAL_ALIGNMENT_LEFT, -1, ICON_FONT_SIZE, Color("#f4ead8"))


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
	intent_requested.emit("stash_withdraw_item_intent", {
		"stash_entity_id": stash_entity_id,
		"stash_item_id": str(item.get("stash_item_id", "")),
	})


func _handle_drop_on_stash(data: Variant) -> void:
	if typeof(data) != TYPE_DICTIONARY:
		return
	var rec := data as Dictionary
	if str(rec.get("source", "")) != DRAG_SOURCE_INVENTORY_BAG:
		return
	var item: Dictionary = rec.get("item", {})
	_emit_deposit(item)


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
	for item in stash_items:
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
	for item in stash_items:
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
	for item in stash_items:
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


func _stat_lines(stats_value: Variant) -> Array:
	if typeof(stats_value) != TYPE_DICTIONARY:
		return []
	var stats := stats_value as Dictionary
	var lines: Array = []
	if int(stats.get("damage_min", 0)) > 0 or int(stats.get("damage_max", 0)) > 0:
		lines.append("Damage %d-%d" % [int(stats.get("damage_min", 0)), int(stats.get("damage_max", 0))])
	for key in ["armor", "block_percent", "max_hp", "hotbar_slots", "inventory_rows"]:
		var value := int(stats.get(key, 0))
		if value > 0:
			lines.append("%s +%d" % [StatLabels.display_name(key), value])
	return lines


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
	var viewport_size := get_viewport_rect().size
	_panel.offset_left = margin
	_panel.offset_top = margin + 54.0
	_panel.offset_right = _panel.offset_left + panel_size.x
	_panel.offset_bottom = _panel.offset_top + panel_size.y
	if viewport_size.y > 0.0 and _panel.offset_bottom > viewport_size.y - margin:
		var overflow := _panel.offset_bottom - (viewport_size.y - margin)
		_panel.offset_top -= overflow
		_panel.offset_bottom -= overflow
	if _panel.offset_left < margin:
		_panel.offset_left = margin
		_panel.offset_right = margin + panel_size.x


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
