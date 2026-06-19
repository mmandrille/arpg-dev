class_name BlacksmithPanel
extends Control

signal upgrade_requested(stash_item_id: String)
signal upgrade_inventory_requested(item_instance_id: String)

const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const BlacksmithUpgradePreviewScript := preload("res://scripts/blacksmith_upgrade_preview.gd")
const BlacksmithUpgradeHistoryScript := preload("res://scripts/blacksmith_upgrade_history.gd")
const BlacksmithRecipesScript := preload("res://scripts/blacksmith_recipes.gd")
const ItemIconDrawerScript := preload("res://scripts/item_icon_drawer.gd")
const PANEL_SIZE := Vector2(320, 260)
const STAGE_SLOT_SIZE := Vector2(84, 84)
const BODY_FONT_SIZE := 18
const DETAIL_FONT_SIZE := 15
const ICON_FONT_SIZE := 28

var blacksmith_entity_id: String = ""
var inventory_items: Array = []
var gold: int = 0
var stash_gold: int = 0
var base_cost: int = 100
var growth_cost: int = 50
var max_level: int = 3
var success_chance_percent: int = 100
var pity_failure_threshold: int = 0
var resource_item_def_id: String = ""
var resource_count: int = 0
var resource_wallet: Dictionary = {}
var item_presentations: Dictionary:
	get: return ItemRulesLoader.item_presentations
var staged_item: Dictionary = {}
var _selected_recipe_id: String = BlacksmithRecipesScript.RECIPE_ITEM_UPGRADE
var _panel: DraggableWindow
var _status_label: Label
var _gold_label: Label
var _recipe_selector: OptionButton
var _history_view: Control
var _rows: VBoxContainer

class BlacksmithStageSlot:
	extends Button

	var panel: BlacksmithPanel
	var item: Dictionary = {}

	func _draw() -> void:
		if item.is_empty():
			return
		panel._draw_item_icon(self, item)

	func _gui_input(event: InputEvent) -> void:
		if event is InputEventMouseButton \
				and event.button_index == MOUSE_BUTTON_LEFT \
				and event.pressed \
				and event.double_click \
				and not item.is_empty():
			panel.unstage_item()
			accept_event()

	func _get_drag_data(_at_position: Vector2) -> Variant:
		if item.is_empty():
			return null
		var data := {
			"source": "blacksmith_stage",
			"item": item.duplicate(true),
			"blacksmith_panel": panel,
		}
		set_drag_preview(panel._drag_preview(item))
		return data

	func _can_drop_data(_at_position: Vector2, data: Variant) -> bool:
		if typeof(data) != TYPE_DICTIONARY or typeof(data.get("item", {})) != TYPE_DICTIONARY:
			return false
		var source := str(data.get("source", ""))
		return source == "bag" or source.begins_with("equip:")

	func _drop_data(_at_position: Vector2, data: Variant) -> void:
		panel.stage_inventory_item(data.get("item", {}))
func _ready() -> void:
	ItemRulesLoader.ensure_loaded()
	_build()
	hide_display()

func show_blacksmith(entity_id: String, next_stash_items: Array, next_gold: int, next_stash_gold: int, config: Dictionary, status: String = "", next_resource_wallet: Dictionary = {}) -> void:
	if _panel == null:
		_build()
	blacksmith_entity_id = entity_id
	inventory_items = _dup_array(next_stash_items)
	gold = next_gold
	stash_gold = next_stash_gold
	base_cost = int(config.get("item_upgrade_cost_gold", base_cost))
	growth_cost = int(config.get("item_upgrade_cost_growth_per_level", growth_cost))
	max_level = int(config.get("item_upgrade_max_level", max_level))
	success_chance_percent = int(config.get("item_upgrade_success_chance_percent", success_chance_percent))
	pity_failure_threshold = int(config.get("item_upgrade_pity_failure_threshold", 0))
	resource_item_def_id = str(config.get("item_upgrade_resource_item_def_id", ""))
	resource_count = int(config.get("item_upgrade_resource_count", 0))
	resource_wallet = next_resource_wallet.duplicate(true)
	_status_label.text = status
	_rebuild()
	visible = true
	_panel.visible = true
	_panel.clamp_to_viewport()

func hide_display() -> void:
	unstage_item(false)
	visible = false
	if _panel != null:
		_panel.visible = false

func show_status(message: String, warning: bool = false) -> void:
	if _status_label == null:
		return
	_status_label.text = message
	_status_label.add_theme_color_override("font_color", Color("#ffcf5a") if warning else Color("#9fd7ff"))

func update_after_upgrade(item: Dictionary, next_gold: int, next_stash_gold: int, charged_cost: int, success: bool = true, next_resource_wallet: Dictionary = {}) -> void:
	staged_item = item.duplicate(true)
	gold = next_gold
	stash_gold = next_stash_gold
	resource_wallet = next_resource_wallet.duplicate(true)
	show_status(("Upgraded for %d gold" if success else "Upgrade failed for %d gold") % charged_cost, not success)
	_history_view.record_attempt(_selected_recipe_label(), _item_title(item), success, charged_cost)
	_rebuild()

func bot_click_upgrade(stash_item_id: String = "", item_def_id: String = "", stash_index: int = 0) -> void:
	var item := _matching_item(stash_item_id, item_def_id, stash_index)
	if item.is_empty():
		show_status("No matching inventory item", true)
		return
	stage_inventory_item(item)
	_emit_upgrade(staged_item)

func bot_stage_item(stash_item_id: String = "", item_def_id: String = "", stash_index: int = 0) -> void:
	var item := _matching_item(stash_item_id, item_def_id, stash_index)
	if item.is_empty():
		show_status("No matching inventory item", true)
		return
	stage_inventory_item(item)


func bot_select_recipe(recipe_id: String) -> void:
	select_recipe(recipe_id)


func select_recipe(recipe_id: String) -> void:
	if not _recipe_ids().has(recipe_id):
		return
	_selected_recipe_id = recipe_id
	_sync_recipe_selector()
	_rebuild()


func selected_recipe_id() -> String:
	return _selected_recipe_id


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"blacksmith_entity_id": blacksmith_entity_id,
		"gold": gold,
		"stash_gold": stash_gold,
		"wallet_gold": _wallet_gold(),
		"success_chance_percent": success_chance_percent,
		"pity_failure_count": BlacksmithUpgradePreviewScript.pity_failure_count(staged_item),
		"pity_threshold": pity_failure_threshold,
		"pity_guaranteed": _pity_guaranteed(staged_item),
		"resource_item_def_id": resource_item_def_id,
		"resource_required_count": resource_count,
		"resource_wallet_count": _resource_wallet_count(),
		"recipe_selector_visible": _recipe_selector.visible if _recipe_selector != null else false,
		"selected_recipe_id": selected_recipe_id(),
		"selected_recipe_label": _selected_recipe_label(),
		"recipe_options": _recipe_options(),
		"item_count": inventory_items.size(),
		"staged_item": staged_item.duplicate(true),
		"staged_item_id": str(staged_item.get("item_instance_id", staged_item.get("stash_item_id", ""))),
		"stage_slot_size": {"x": STAGE_SLOT_SIZE.x, "y": STAGE_SLOT_SIZE.y},
		"stage_slot_centered": true,
		"stage_icon_visible": not staged_item.is_empty(),
		"preview_lines": _upgrade_preview_lines(staged_item) if not staged_item.is_empty() else [],
		"upgrade_history": _history_view.get_debug_state() if _history_view != null else {},
		"instruction_visible": false,
		"rows": _debug_rows(),
		"status": _status_label.text if _status_label != null else "",
		"window": _panel.get_debug_state() if _panel != null else {},
	}

func stage_inventory_item(item: Dictionary) -> void:
	if item.is_empty():
		return
	staged_item = item.duplicate(true)
	show_status("Ready to upgrade %s" % _item_title(staged_item))
	_rebuild()

func unstage_item(show_message: bool = true) -> void:
	if staged_item.is_empty():
		return
	var title := _item_title(staged_item)
	staged_item = {}
	if show_message:
		show_status("%s returned to inventory" % title)
	_rebuild()

func _build() -> void:
	if _panel != null:
		return
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	set_anchors_preset(Control.PRESET_FULL_RECT)

	_panel = DraggableWindowScript.new()
	_panel.configure("Blacksmith", PANEL_SIZE)
	_panel.custom_minimum_size = Vector2(PANEL_SIZE.x, PANEL_SIZE.y + DraggableWindowScript.TITLEBAR_HEIGHT)
	_panel.size = _panel.custom_minimum_size
	_panel.set_layout_key("blacksmith_panel")
	_panel.position = Vector2(300, 88)
	_panel.close_requested.connect(hide_display)
	add_child(_panel)

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 8)
	root.alignment = BoxContainer.ALIGNMENT_CENTER
	_panel.set_content(root)

	_status_label = Label.new()
	_status_label.text = ""
	_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_status_label.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	_status_label.add_theme_color_override("font_color", Color("#9fd7ff"))
	root.add_child(_status_label)

	_gold_label = Label.new()
	_gold_label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	_gold_label.add_theme_color_override("font_color", Color("#f4d481"))
	root.add_child(_gold_label)

	root.add_child(_recipe_selector_row())

	_rows = VBoxContainer.new()
	_rows.add_theme_constant_override("separation", 6)
	_rows.alignment = BoxContainer.ALIGNMENT_CENTER
	_rows.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
	root.add_child(_rows)
	_history_view = BlacksmithUpgradeHistoryScript.new()
	root.add_child(_history_view)

func _rebuild() -> void:
	_gold_label.text = "Gold: %d  Stash: %d" % [gold, stash_gold]
	_sync_recipe_selector()
	_clear_rows()
	_rows.add_child(_stage_slot())
	_rows.add_child(_preview_block())
	var button_center := CenterContainer.new()
	var button := Button.new()
	button.text = "Upgrade"
	button.custom_minimum_size = Vector2(150, 40)
	button.disabled = staged_item.is_empty() or not _upgrade_enabled(staged_item)
	button.pressed.connect(func() -> void: _emit_upgrade(staged_item))
	button_center.add_child(button)
	_rows.add_child(button_center)

func _stage_slot() -> Control:
	var center := CenterContainer.new()
	var btn := BlacksmithStageSlot.new()
	btn.panel = self
	btn.item = staged_item.duplicate(true)
	btn.custom_minimum_size = STAGE_SLOT_SIZE
	btn.text = "" if not staged_item.is_empty() else "Empty"
	btn.clip_text = true
	btn.tooltip_text = _item_detail(staged_item) if not staged_item.is_empty() else "Drop inventory item"
	btn.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	btn.add_theme_color_override("font_color", _rarity_color(str(staged_item.get("rarity", ""))) if not staged_item.is_empty() else Color("#8f826b"))
	btn.add_theme_stylebox_override("normal", _stage_slot_style(false))
	btn.add_theme_stylebox_override("hover", _stage_slot_style(true))
	btn.add_theme_stylebox_override("pressed", _stage_slot_style(true))
	center.add_child(btn)
	return center

func _preview_block() -> Control:
	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 4)
	box.custom_minimum_size = Vector2(280, 0)
	if staged_item.is_empty():
		box.add_child(_empty_label("No item selected"))
		return box
	var level := _item_level(staged_item)
	var cost := _next_cost(level)
	var cost_row := HBoxContainer.new()
	cost_row.add_theme_constant_override("separation", 8)
	var level_label := Label.new()
	level_label.text = "Level %d/%d" % [level, max_level]
	level_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	level_label.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	level_label.add_theme_color_override("font_color", Color("#b8b8b8"))
	cost_row.add_child(level_label)
	var cost_label := Label.new()
	cost_label.text = "%d gold" % cost
	cost_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	cost_label.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	cost_label.add_theme_color_override("font_color", Color("#d8c8a8"))
	cost_row.add_child(cost_label)
	box.add_child(cost_row)
	if resource_count > 0:
		var resource_label := _empty_label("%s: %d/%d" % [_resource_display_name(), _resource_wallet_count(), resource_count])
		resource_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		resource_label.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
		resource_label.add_theme_color_override("font_color", Color("#d8c8a8") if _has_upgrade_resource() else Color("#ff9f7a"))
		box.add_child(resource_label)
	for line in _upgrade_preview_lines(staged_item):
		var label := _empty_label(line)
		label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		box.add_child(label)
	return box

func _emit_upgrade(item: Dictionary) -> void:
	var level := _item_level(item)
	var cost := _next_cost(level)
	if not _recipe_accepts_item(item):
		show_status(BlacksmithRecipesScript.rejection_message(_selected_recipe_id), true)
		return
	if level >= max_level:
		show_status("Item is already at max level", true)
		return
	if _wallet_gold() < cost:
		show_status("Need %d gold" % cost, true)
		return
	if not _has_upgrade_resource():
		show_status("Need %d %s" % [resource_count, _resource_display_name()], true)
		return
	var item_instance_id := str(item.get("item_instance_id", ""))
	if item_instance_id != "":
		upgrade_inventory_requested.emit(item_instance_id)
		return
	var stash_item_id := str(item.get("stash_item_id", ""))
	if stash_item_id == "":
		show_status("Missing item id", true)
		return
	upgrade_requested.emit(stash_item_id)

func _matching_item(stash_item_id: String, item_def_id: String, stash_index: int) -> Dictionary:
	var matches: Array = []
	for value in inventory_items:
		if typeof(value) != TYPE_DICTIONARY:
			continue
		var item := value as Dictionary
		if stash_item_id != "" and str(item.get("item_instance_id", item.get("stash_item_id", ""))) != stash_item_id:
			continue
		if item_def_id != "" and str(item.get("item_def_id", "")) != item_def_id:
			continue
		matches.append(item)
	if matches.is_empty() or stash_index < 0 or stash_index >= matches.size():
		return {}
	return (matches[stash_index] as Dictionary).duplicate(true)


func _debug_rows() -> Array:
	var rows: Array = []
	if not staged_item.is_empty():
		rows.append(_debug_row(staged_item))
	for value in inventory_items:
		if typeof(value) != TYPE_DICTIONARY:
			continue
		var item := value as Dictionary
		if not staged_item.is_empty() and str(item.get("item_instance_id", "")) == str(staged_item.get("item_instance_id", "")):
			continue
		rows.append(_debug_row(item))
	return rows


func _debug_row(item: Dictionary) -> Dictionary:
	var level := _item_level(item)
	return {
		"item_instance_id": str(item.get("item_instance_id", "")),
		"stash_item_id": str(item.get("stash_item_id", "")),
		"item_def_id": str(item.get("item_def_id", "")),
		"display_name": _item_title(item),
		"rarity": str(item.get("rarity", "")),
		"item_level": level,
		"next_cost_gold": _next_cost(level),
		"upgrade_enabled": _upgrade_enabled(item),
	}


func _upgrade_enabled(item: Dictionary) -> bool:
	var level := _item_level(item)
	return _is_upgrade_candidate(item) and _recipe_accepts_item(item) and level < max_level and _wallet_gold() >= _next_cost(level) and _has_upgrade_resource()


func _is_upgrade_candidate(item: Dictionary) -> bool:
	if item.is_empty():
		return false
	if resource_item_def_id != "" and str(item.get("item_def_id", "")) == resource_item_def_id:
		return false
	return str(item.get("item_template_id", "")) != "" or str(item.get("slot", "")) != "" or str(item.get("category", "")) == "equipment"


func _has_upgrade_resource() -> bool:
	return resource_count <= 0 or _resource_wallet_count() >= resource_count


func _resource_wallet_count() -> int:
	if resource_count <= 0 or resource_item_def_id == "":
		return 0
	return max(0, int(resource_wallet.get(resource_item_def_id, 0)))


func _resource_display_name() -> String:
	if resource_item_def_id == "":
		return "resource"
	var def := ItemRulesLoader.item_definition(resource_item_def_id)
	return str(def.get("name", resource_item_def_id.replace("_", " ").capitalize()))

func _pity_failure_count(item: Dictionary) -> int:
	return BlacksmithUpgradePreviewScript.pity_failure_count(item)

func _pity_guaranteed(item: Dictionary) -> bool:
	return BlacksmithUpgradePreviewScript.pity_guaranteed(item, pity_failure_threshold)

func _wallet_gold() -> int:
	return gold + stash_gold

func _item_level(item: Dictionary) -> int:
	return BlacksmithUpgradePreviewScript.item_level(item)

func _next_cost(level: int) -> int:
	return BlacksmithUpgradePreviewScript.next_cost(level, base_cost, growth_cost)

func _upgrade_preview_lines(item: Dictionary) -> Array:
	var lines: Array = ["Recipe: %s" % _selected_recipe_label()]
	lines.append(_selected_recipe_eligibility())
	if not _recipe_accepts_item(item):
		lines.append("Recipe cannot modify this item")
		return lines
	lines.append_array(BlacksmithUpgradePreviewScript.preview_lines(item, {
		"base_cost": base_cost,
		"growth_cost": growth_cost,
		"max_level": max_level,
		"success_chance_percent": success_chance_percent,
		"pity_failure_threshold": pity_failure_threshold,
		"resource_count": resource_count,
		"resource_wallet_count": _resource_wallet_count(),
		"resource_name": _resource_display_name(),
		"wallet_gold": _wallet_gold(),
	}))
	return lines


func _recipe_selector_row() -> Control:
	var row := HBoxContainer.new()
	row.add_theme_constant_override("separation", 8)
	var label := Label.new()
	label.text = "Recipe"
	label.add_theme_font_size_override("font_size", DETAIL_FONT_SIZE)
	label.add_theme_color_override("font_color", Color("#d8c8a8"))
	row.add_child(label)
	_recipe_selector = OptionButton.new()
	_recipe_selector.custom_minimum_size = Vector2(190, 30)
	_recipe_selector.focus_mode = Control.FOCUS_NONE
	_recipe_selector.item_selected.connect(_on_recipe_selected)
	row.add_child(_recipe_selector)
	_sync_recipe_selector()
	return row


func _sync_recipe_selector() -> void:
	if _recipe_selector == null:
		return
	_recipe_selector.clear()
	for option in _recipe_options():
		_recipe_selector.add_item(str((option as Dictionary).get("label", "")))
	_recipe_selector.select(max(0, _recipe_ids().find(_selected_recipe_id)))


func _recipe_options() -> Array:
	return BlacksmithRecipesScript.options(resource_item_def_id, resource_count, success_chance_percent, max_level)


func _selected_recipe_label() -> String:
	return BlacksmithRecipesScript.label(_selected_recipe_id)


func _selected_recipe_eligibility() -> String:
	return BlacksmithRecipesScript.eligibility(_selected_recipe_id)


func _recipe_ids() -> Array:
	return BlacksmithRecipesScript.ids()


func _on_recipe_selected(index: int) -> void:
	var options := _recipe_options()
	if index >= 0 and index < options.size():
		_selected_recipe_id = str((options[index] as Dictionary).get("id", BlacksmithRecipesScript.RECIPE_ITEM_UPGRADE))
		_rebuild()


func _recipe_accepts_item(item: Dictionary) -> bool:
	return BlacksmithRecipesScript.accepts_item(_selected_recipe_id, item)


func _item_title(item: Dictionary) -> String:
	var display := str(item.get("display_name", ""))
	if display != "":
		return display
	return str(item.get("item_def_id", "Unknown item")).replace("_", " ").capitalize()


func _item_detail(item: Dictionary) -> String:
	var lines: Array = item.get("summary_lines", [])
	if not lines.is_empty():
		return str(lines[0])
	return "Level %d/%d" % [_item_level(item), max_level]


func _draw_item_icon(slot: Control, item: Dictionary) -> void:
	var def_id := str(item.get("item_def_id", ""))
	var icon: Dictionary = item_presentations.get(def_id, {}).get("icon", {})
	var rect := Rect2(Vector2.ZERO, slot.size)
	var label := str(icon.get("label", _short_label(def_id)))
	ItemIconDrawerScript.draw(slot, rect, icon, label, false, 0.18, ICON_FONT_SIZE)


func _drag_preview(item: Dictionary) -> Control:
	var preview := Control.new()
	preview.custom_minimum_size = STAGE_SLOT_SIZE
	preview.size = STAGE_SLOT_SIZE
	preview.draw.connect(func() -> void:
		_draw_item_icon(preview, item)
	)
	return preview


func _short_label(def_id: String) -> String:
	var parts := def_id.replace("_", " ").split(" ")
	var out := ""
	for part in parts:
		if part.length() > 0:
			out += part.substr(0, 1).to_upper()
	return out.substr(0, 3)


func _rarity_color(rarity: String) -> Color:
	match rarity:
		"magic":
			return Color("#93c5fd")
		"rare":
			return Color("#f4d481")
		"unique":
			return Color("#ffb26b")
		_:
			return Color("#e8dcc8")


func _clear_rows() -> void:
	for child in _rows.get_children():
		child.queue_free()


func _empty_label(text: String) -> Label:
	var empty := Label.new()
	empty.text = text
	empty.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	empty.add_theme_color_override("font_color", Color("#e8dcc8"))
	return empty

func _stage_slot_style(hover: bool) -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color("#241d15") if hover else Color("#15110d")
	s.border_color = Color("#c59035") if hover else Color("#6f5524")
	s.set_border_width_all(2)
	s.set_content_margin_all(6)
	return s


func _dup_array(values: Array) -> Array:
	var out: Array = []
	for value in values:
		out.append((value as Dictionary).duplicate(true) if typeof(value) == TYPE_DICTIONARY else value)
	return out
