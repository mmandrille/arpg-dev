class_name InventoryPanel
extends Control

signal intent_requested(intent_type: String, payload: Dictionary)

const ItemTooltipPanelScript := preload("res://scripts/item_tooltip_panel.gd")
const ItemIconDrawerScript := preload("res://scripts/item_icon_drawer.gd")
const PotionIconLabelScript := preload("res://scripts/potion_icon_label.gd")
const PaperDollBackdropScript := preload("res://scripts/paper_doll_backdrop.gd")
const PaperDollLayoutScript := preload("res://scripts/paper_doll_layout.gd")
const UniqueEffectTooltipScript := preload("res://scripts/unique_effect_tooltip.gd")
const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const WeaponSetTabsScript := preload("res://scripts/weapon_set_tabs.gd")
const InventoryTransferRouterScript := preload("res://scripts/inventory_transfer_router.gd")
const SetCollectionPanelScript := preload("res://scripts/set_collection_panel.gd")
const InventoryRenderGuardScript := preload("res://scripts/inventory_render_guard.gd")
const InventoryPanelStylesScript := preload("res://scripts/inventory_panel_styles.gd")
const MaterialWalletPanelScript := preload("res://scripts/material_wallet_panel.gd")
const ItemRequirementViewsScript := preload("res://scripts/item_requirement_views.gd")
const SLOT_KIND_BAG := "bag"
const SLOT_KIND_EQUIP_PREFIX := "equip:"
const DRAG_SOURCE_SHOP_OFFER := "shop_offer"
const DRAG_SOURCE_STASH := "stash"
const DRAG_SOURCE_CORPSE := "corpse"
const DRAG_SOURCE_UNIQUE_CHEST := "unique_chest"
const BAG_COLUMNS := 5
const BASE_INVENTORY_ROWS := 3
const HOTKEY_LABELS := InventoryTooltipContent.HOTKEY_LABELS
const TITLE_FONT_SIZE := 33
const BODY_FONT_SIZE := 23
const SLOT_FONT_SIZE := 23
const ICON_FONT_SIZE := 22
const EQUIPMENT_SLOT_SIZE := Vector2(96, 58)
const EQUIPMENT_SLOTS := ["head", "amulet", "chest", "gloves", "belt", "boots", "ring_left", "ring_right", "main_hand", "off_hand"]
const EQUIPMENT_LABELS := {
	"head": "Head",
	"amulet": "Amulet",
	"chest": "Chest",
	"gloves": "Gloves",
	"belt": "Belt",
	"boots": "Boots",
	"ring_left": "Ring L",
	"ring_right": "Ring R",
	"main_hand": "Main",
	"off_hand": "Off"
}

var inventory: Array = []
var equipped: Dictionary = {}
var active_weapon_set: int = 0
var viewed_weapon_set: int = 0
var weapon_sets: Array = []
var hotbar: Array = []
var hotbar_capacity: int = 2
var inventory_rows: int = BASE_INVENTORY_ROWS
var inventory_capacity: int = BASE_INVENTORY_ROWS * BAG_COLUMNS
var gold: int = 0
var character_progression: Dictionary = {}
var item_rules: Dictionary:
	get: return ItemRulesLoader.item_rules
var item_templates: Dictionary:
	get: return ItemRulesLoader.item_templates
var item_presentations: Dictionary:
	get: return ItemRulesLoader.item_presentations
var _panel: DraggableWindow
var _equipment_slots: Dictionary = {}
var _weapon_set_tabs: Array = []
var _bag_grid: GridContainer
var _gold_label: Label
var _resources_button: Button
var _set_collection_button: Button
var _resource_wallet: Dictionary = {}
var _resource_bag_items: Array = []
var _wallet_window: Control
var _set_collection_panel: SetCollectionPanel
var _paper_doll_preview: Control
var _drag_data: Dictionary = {}
var _interactive: bool = true
var _gesture_hint: Label
var _gesture_tween: Tween
var _rendered_bag_slot_count: int = 0
var _shop_sell_entity_id: String = ""
var _market_context: String = ""
var _market_hidden_item_ids: Array = []
var _blacksmith_hidden_item_ids: Array = []
var _blacksmith_context_enabled: bool = false


class InventorySlotButton:
	extends Button

	var panel: InventoryPanel
	var slot_kind: String = ""
	var item: Dictionary = {}

	func _draw() -> void:
		if item.is_empty():
			return
		panel._draw_item_icon(self, item)

	func _gui_input(event: InputEvent) -> void:
		if not panel._interactive:
			return
		if event is InputEventMouseButton \
				and event.button_index == MOUSE_BUTTON_LEFT \
				and event.pressed \
				and slot_kind == SLOT_KIND_BAG \
				and not item.is_empty():
			if event.shift_pressed:
				panel._handle_shift_click(item)
				accept_event()
			elif event.double_click:
				panel._handle_double_click(item)

	func _get_drag_data(_at_position: Vector2) -> Variant:
		if not panel._interactive or item.is_empty() or bool(item.get("_blocked_by_two_handed", false)):
			return null
		panel._drag_data = {"source": slot_kind, "item": item}
		var preview := Label.new()
		preview.text = str(item.get("item_def_id", "item"))
		preview.add_theme_color_override("font_color", Color("#e8dcc8"))
		set_drag_preview(preview)
		return panel._drag_data

	func _can_drop_data(_at_position: Vector2, data: Variant) -> bool:
		if typeof(data) != TYPE_DICTIONARY:
			return false
		var source := str(data.get("source", ""))
		var dragged: Dictionary = data.get("item", {})
		if dragged.is_empty():
			return false
		if InventoryTransferRouterScript.is_equipment_slot(slot_kind):
			if source == "blacksmith_stage" and data.get("blacksmith_panel", null) != null:
				return panel._item_can_equip_to(dragged, panel._slot_from_kind(slot_kind))
			if source == "blacksmith_resource_stage" and data.get("blacksmith_panel", null) != null:
				return false
			return (source == SLOT_KIND_BAG or source == DRAG_SOURCE_STASH or source == DRAG_SOURCE_CORPSE) and panel._item_can_equip_to(dragged, panel._slot_from_kind(slot_kind))
		if slot_kind == SLOT_KIND_BAG:
			if source == DRAG_SOURCE_SHOP_OFFER and str(data.get("offer_id", "")) != "" and str(data.get("shop_entity_id", "")) != "":
				return true
			if source == DRAG_SOURCE_STASH \
					and str(data.get("stash_entity_id", "")) != "" \
					and str(data.get("stash_item_id", "")) != "":
				return true
			if source == DRAG_SOURCE_CORPSE \
					and str(data.get("corpse_entity_id", "")) != "" \
					and str(data.get("item_instance_id", "")) != "":
				return true
			if source == DRAG_SOURCE_UNIQUE_CHEST \
					and str(data.get("stash_entity_id", "")) != "" \
					and str(data.get("stash_item_id", "")) != "":
				return true
			if source == "blacksmith_stage" and data.get("blacksmith_panel", null) != null:
				return true
			if source == "blacksmith_resource_stage" and data.get("blacksmith_panel", null) != null:
				return true
			if source == "blacksmith_merge" and data.get("merge_panel", null) != null:
				return true
			if source == InventoryTransferRouterScript.DRAG_SOURCE_RESOURCE_BAG \
					and InventoryTransferRouterScript.resource_bag_item_id(data as Dictionary) != "":
				return true
			return InventoryTransferRouterScript.is_equipment_slot(source)
		return false

	func _drop_data(_at_position: Vector2, data: Variant) -> void:
		panel._handle_drop_on_slot(slot_kind, data)

	func _make_custom_tooltip(for_text: String) -> Object:
		if panel == null:
			return null
		if item.is_empty():
			return panel._make_text_tooltip(for_text)
		if bool(item.get("_blocked_by_two_handed", false)):
			return panel._make_text_tooltip("%s occupies both hands" % str(item.get("display_name", item.get("item_def_id", "Two-handed item"))))
		return panel._make_item_tooltip(item)


func _ready() -> void:
	ItemRulesLoader.ensure_loaded()
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_ensure_built()
	_render()
	visible = false


func toggle() -> void:
	if not _interactive:
		return
	visible = not visible
	_apply_interaction_filters()


func set_interactive(enabled: bool) -> void:
	_interactive = enabled
	_apply_interaction_filters()


func set_shop_sell_context(shop_entity_id: String) -> void:
	_shop_sell_entity_id = shop_entity_id

func clear_shop_sell_context() -> void:
	_shop_sell_entity_id = ""

func set_market_context(context: String) -> void:
	_market_context = context
	if _bag_grid != null:
		_render()

func clear_market_context() -> void:
	_market_context = ""
	_market_hidden_item_ids = []
	if _bag_grid != null:
		_render()


func set_market_hidden_item_ids(item_instance_ids: Array) -> void:
	_market_hidden_item_ids = []
	for item_id in item_instance_ids:
		var id := str(item_id)
		if id != "" and not _market_hidden_item_ids.has(id):
			_market_hidden_item_ids.append(id)
	if _bag_grid != null:
		_render()


func set_blacksmith_context(enabled: bool) -> void:
	_blacksmith_context_enabled = enabled
	if not enabled:
		_blacksmith_hidden_item_ids = []
	if _bag_grid != null:
		_render()


func set_blacksmith_hidden_item_ids(item_instance_ids: Array) -> void:
	_blacksmith_hidden_item_ids = []
	for item_id in item_instance_ids:
		var id := str(item_id)
		if id != "" and not _blacksmith_hidden_item_ids.has(id):
			_blacksmith_hidden_item_ids.append(id)
	if _bag_grid != null:
		_render()

func set_resource_wallet(next_wallet: Dictionary, next_bag_items: Array = []) -> void:
	_resource_wallet = next_wallet.duplicate(true)
	_resource_bag_items = _dup_items(next_bag_items)
	_ensure_built()
	_render_resources_button()
	_sync_wallet_window()


func open_wallet_window() -> void:
	if _wallet_rows().is_empty() and _resource_bag_items.is_empty():
		return
	_ensure_built()
	if _wallet_window == null:
		_wallet_window = MaterialWalletPanelScript.new()
		_wallet_window.intent_requested.connect(_on_wallet_intent_requested)
		add_child(_wallet_window)
	_wallet_window.show_wallet(_resource_wallet, _resource_bag_items)


func _on_wallet_intent_requested(intent_type: String, payload: Dictionary) -> void:
	intent_requested.emit(intent_type, payload)


func set_character_progression(next_progression: Dictionary) -> void:
	character_progression = next_progression.duplicate(true)
	_sync_paper_doll_backdrop()


func ensure_display_visible() -> void:
	visible = true
	_apply_interaction_filters()


func hide_display() -> void:
	visible = false


func bot_click_close() -> void:
	if _panel != null and _panel.close_button() != null:
		_panel.close_button().pressed.emit()


func bot_drag_window_by(delta: Vector2) -> void:
	if _panel != null:
		_panel.bot_drag_by(delta)


func show_gesture_hint(text: String) -> void:
	if _gesture_hint == null:
		return
	if _gesture_tween != null and is_instance_valid(_gesture_tween):
		_gesture_tween.kill()
	_gesture_hint.text = text
	_gesture_hint.modulate.a = 1.0
	_gesture_hint.visible = true
	_gesture_tween = create_tween()
	_gesture_tween.tween_interval(0.9)
	_gesture_tween.tween_property(_gesture_hint, "modulate:a", 0.0, 0.35)
	_gesture_tween.tween_callback(func() -> void:
		if _gesture_hint != null:
			_gesture_hint.visible = false)


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_reposition_panel()


func set_inventory_state(next_inventory: Array, next_equipped: Dictionary, next_inventory_rows: int = BASE_INVENTORY_ROWS, next_inventory_capacity: int = BASE_INVENTORY_ROWS * BAG_COLUMNS, next_gold: int = 0, next_hotbar: Array = [], next_hotbar_capacity: int = 2, next_active_weapon_set: int = 0, next_weapon_sets: Array = []) -> void:
	ItemRulesLoader.ensure_loaded()
	inventory = []
	for item in next_inventory:
		inventory.append((item as Dictionary).duplicate(true))
	equipped = next_equipped.duplicate(true)
	var previous_active_weapon_set := active_weapon_set
	active_weapon_set = clamp(next_active_weapon_set, 0, 1)
	weapon_sets = []
	for set_data in next_weapon_sets:
		weapon_sets.append((set_data as Dictionary).duplicate(true))
	if weapon_sets.is_empty():
		weapon_sets = WeaponSetTabsScript.fallback_sets(equipped)
	if active_weapon_set != previous_active_weapon_set:
		viewed_weapon_set = active_weapon_set
	viewed_weapon_set = clamp(viewed_weapon_set, 0, 1)
	hotbar = []
	for slot in next_hotbar:
		hotbar.append((slot as Dictionary).duplicate(true))
	hotbar_capacity = clamp(next_hotbar_capacity, 2, HOTKEY_LABELS.size())
	inventory_rows = max(0, next_inventory_rows)
	inventory_capacity = max(0, next_inventory_capacity)
	gold = max(0, next_gold)
	_ensure_built()
	_rendered_bag_slot_count = _target_bag_slot_count()
	if _bag_grid != null and InventoryRenderGuardScript.should_render(self):
		_render()
		InventoryRenderGuardScript.mark_rendered(self)


func get_debug_state() -> Dictionary:
	_ensure_built()
	return {
		"visible": visible,
		"bag_count": inventory.size(),
		"visible_bag_count": _bag_items().size(),
		"market_hidden_item_ids": _market_hidden_item_ids.duplicate(),
		"equipped": equipped.duplicate(true),
		"active_weapon_set": active_weapon_set,
		"viewed_weapon_set": viewed_weapon_set,
		"weapon_sets": weapon_sets.duplicate(true),
		"equipped_main_hand": equipped.get("main_hand", null),
		"main_hand_item": _equipped_item("main_hand"),
		"weapon_item": _equipped_item("main_hand"),
		"item_presentations": _debug_presentations(),
		"inventory_rows": inventory_rows,
		"inventory_capacity": inventory_capacity,
		"gold": gold,
		"hotbar_assigned_item_ids": _debug_hotbar_assigned_item_ids(),
		"hotbar_assigned_inventory_count": _hotbar_assigned_inventory_count(),
		"bag_columns": _bag_grid.columns if _bag_grid != null else BAG_COLUMNS,
		"available_slot_count": inventory_capacity,
		"rendered_slot_count": _rendered_bag_slot_count,
		"paper_doll_slot_ids": EQUIPMENT_SLOTS.duplicate(),
		"paper_doll_slots": _debug_paper_doll_slots(),
		"paper_doll_preview": {
			"exists": _paper_doll_preview != null,
			"name": _paper_doll_preview.name if _paper_doll_preview != null else "",
			"visible": _paper_doll_preview.visible if _paper_doll_preview != null else false,
		},
		"set_collection": _set_collection_panel.get_debug_state() if _set_collection_panel != null else {},
		"resources_button_text": _resources_button.text if _resources_button != null else "",
		"wallet_visible": not _wallet_rows().is_empty() or not _resource_bag_items.is_empty(),
		"wallet_text": "  ".join(_wallet_rows()),
		"wallet_tooltip": "\n".join(_wallet_detail_lines()),
		"wallet_rows": _wallet_rows(),
		"wallet_details": _wallet_detail_lines(),
		"wallet_window": _wallet_window.get_debug_state() if _wallet_window != null else {"visible": false, "row_count": 0, "rows": [], "text": ""},
		"requirement_row_count": _requirement_row_count(),
		"equip_preview_row_count": _equip_preview_row_count(),
		"empty_slot_style": "gray_block",
		"window": _panel.get_debug_state() if _panel != null else {},
	}


func get_weapon_slot_screen_center() -> Vector2:
	return get_equipment_slot_screen_center("main_hand")


func get_equipment_slot_screen_center(slot: String) -> Vector2:
	return _slot_screen_center(_equipment_slots.get(slot, null))


func get_bag_area_screen_center() -> Vector2:
	if _bag_grid == null:
		return Vector2.ZERO
	for child in _bag_grid.get_children():
		if child is InventorySlotButton and child.slot_kind == SLOT_KIND_BAG and child.item.is_empty():
			return _slot_screen_center(child)
	for child in _bag_grid.get_children():
		if child is InventorySlotButton and child.slot_kind == SLOT_KIND_BAG:
			return _slot_screen_center(child)
	return Vector2.ZERO


func get_bag_item_screen_center(item_instance_id: String = "") -> Vector2:
	if _bag_grid == null:
		return Vector2.ZERO
	for child in _bag_grid.get_children():
		if child is InventorySlotButton and child.slot_kind == SLOT_KIND_BAG:
			if item_instance_id == "" or str(child.item.get("item_instance_id", "")) == item_instance_id:
				return _slot_screen_center(child)
	if item_instance_id != "":
		var bag_index := 0
		for item in inventory:
			if _is_equipped_instance(str(item.get("item_instance_id", ""))):
				if str(item.get("item_instance_id", "")) == item_instance_id:
					return _bag_cell_screen_center(bag_index)
				continue
			if str(item.get("item_instance_id", "")) == item_instance_id:
				return _bag_cell_screen_center(bag_index)
			bag_index += 1
	return Vector2.ZERO


func get_drop_outside_screen_point() -> Vector2:
	if _panel == null or not is_inside_tree():
		return Vector2(120, get_viewport_rect().size.y - 80)
	var panel_rect := _panel.get_global_rect()
	return Vector2(panel_rect.position.x - 48, panel_rect.position.y + panel_rect.size.y * 0.5)


func _slot_screen_center(slot: Control) -> Vector2:
	if slot == null or not is_inside_tree():
		return Vector2.ZERO
	return slot.get_global_rect().get_center()


func _bag_cell_screen_center(cell_index: int) -> Vector2:
	if _bag_grid == null or not is_inside_tree():
		return Vector2.ZERO
	var cell_w := 48.0
	var cell_h := 48.0
	var sep := float(_bag_grid.get_theme_constant("h_separation"))
	if sep <= 0.0:
		sep = 6.0
	var grid_rect := _bag_grid.get_global_rect()
	var col := cell_index % _bag_grid.columns
	var row := int(cell_index / _bag_grid.columns)
	return grid_rect.position + Vector2(
		col * (cell_w + sep) + cell_w * 0.5,
		row * (cell_h + sep) + cell_h * 0.5,
	)


func _notification(what: int) -> void:
	if what == NOTIFICATION_DRAG_END and _interactive and not _drag_data.is_empty():
		if not get_viewport().gui_is_drag_successful():
			var item: Dictionary = _drag_data.get("item", {})
			if not item.is_empty():
				intent_requested.emit("drop_intent", {"item_instance_id": str(item.get("item_instance_id", ""))})
		_drag_data = {}


func _build() -> void:
	if _panel != null:
		return
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = DraggableWindowScript.new()
	_panel.custom_minimum_size = Vector2(750, 460)
	_panel.configure("Inventory", Vector2(720, 396))
	_reposition_panel()
	_panel.set_layout_key("inventory")
	_panel.add_theme_stylebox_override("panel", InventoryPanelStylesScript.panel_style())
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	_panel.close_requested.connect(hide_display)
	add_child(_panel)

	var body := VBoxContainer.new()
	body.add_theme_constant_override("separation", 2)
	body.custom_minimum_size = Vector2(720, 396)
	_panel.set_content(body)

	_gesture_hint = Label.new()
	_gesture_hint.visible = false
	_gesture_hint.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_gesture_hint.add_theme_color_override("font_color", Color("#c9a227"))
	_gesture_hint.add_theme_font_size_override("font_size", 23)
	body.add_child(_gesture_hint)

	var root := HBoxContainer.new()
	root.add_theme_constant_override("separation", 18)
	root.custom_minimum_size = Vector2(720, 370)
	body.add_child(root)

	var left := VBoxContainer.new()
	left.custom_minimum_size = Vector2(350, 0)
	root.add_child(left)
	left.add_child(_title("Equipment"))
	var paper := Control.new()
	paper.custom_minimum_size = Vector2(340, 360)
	left.add_child(paper)
	_paper_doll_preview = PaperDollBackdropScript.new()
	paper.add_child(_paper_doll_preview)
	for slot in EQUIPMENT_SLOTS:
		var btn := _slot_button(_slot_kind_for_equipment(str(slot)), EQUIPMENT_SLOT_SIZE)
		btn.position = PaperDollLayoutScript.position_for(str(slot))
		btn.size = btn.custom_minimum_size
		_equipment_slots[str(slot)] = btn
		paper.add_child(btn)
	_sync_paper_doll_backdrop()
	_weapon_set_tabs = WeaponSetTabsScript.build_tabs(paper, Callable(self, "_set_viewed_weapon_set"))

	var right := VBoxContainer.new()
	right.custom_minimum_size = Vector2(350, 0)
	root.add_child(right)
	right.add_child(_caption("Bag"))
	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(340, 382)
	right.add_child(scroll)
	_bag_grid = GridContainer.new()
	_bag_grid.columns = BAG_COLUMNS
	_bag_grid.add_theme_constant_override("h_separation", 6)
	_bag_grid.add_theme_constant_override("v_separation", 6)
	scroll.add_child(_bag_grid)
	var footer := HBoxContainer.new()
	footer.add_theme_constant_override("separation", 8)
	right.add_child(footer)
	_gold_label = Label.new()
	_gold_label.text = "Gold: 0"
	_gold_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	_gold_label.add_theme_color_override("font_color", Color("#f4c84f"))
	_gold_label.add_theme_font_size_override("font_size", 26)
	footer.add_child(_gold_label)
	_resources_button = Button.new()
	_resources_button.text = "Resources"
	_resources_button.tooltip_text = "Account materials"
	_resources_button.pressed.connect(open_wallet_window)
	footer.add_child(_resources_button)
	_set_collection_button = Button.new()
	_set_collection_button.text = "Sets"
	_set_collection_button.pressed.connect(_toggle_set_collection_panel)
	footer.add_child(_set_collection_button)
	_set_collection_panel = SetCollectionPanelScript.new()
	body.add_child(_set_collection_panel)
	_render()


func _ensure_built() -> void:
	if _panel != null:
		return
	_build()


func _render() -> void:
	if _bag_grid == null:
		return
	for slot in EQUIPMENT_SLOTS:
		_fill_slot(_equipment_slots.get(slot, null), _equipment_slot_display_item(str(slot)))
	WeaponSetTabsScript.render_tabs(_weapon_set_tabs, active_weapon_set, viewed_weapon_set)
	for child in _bag_grid.get_children():
		child.queue_free()
	var bag_items := _bag_items()
	_rendered_bag_slot_count = _target_bag_slot_count()
	for i in range(_rendered_bag_slot_count):
		var slot := _slot_button(SLOT_KIND_BAG, Vector2(58, 58))
		var item: Dictionary = bag_items[i] if i < bag_items.size() else {}
		_fill_slot(slot, item)
		_bag_grid.add_child(slot)
	if _gold_label != null:
		_gold_label.text = "Gold: %d" % gold
	_render_resources_button()
	if _set_collection_panel != null:
		_set_collection_panel.set_items(inventory, equipped, weapon_sets)
	_position_gesture_hint()

func _toggle_set_collection_panel() -> void:
	if _set_collection_panel == null:
		return
	_set_collection_panel.visible = not _set_collection_panel.visible
	_set_collection_panel.set_items(inventory, equipped, weapon_sets)


func _set_viewed_weapon_set(index: int) -> void:
	viewed_weapon_set = clamp(index, 0, 1)
	_render()


func _weapon_set_payload_for_slot(slot: String) -> Dictionary:
	return {"weapon_set": viewed_weapon_set} if slot == "main_hand" or slot == "off_hand" else {}

func _bag_items() -> Array:
	var items: Array = []
	for item in inventory:
		if _is_equipped_instance(str(item.get("item_instance_id", ""))):
			continue
		if _market_context == "offer" and _market_hidden_item_ids.has(str(item.get("item_instance_id", ""))):
			continue
		if _blacksmith_context_enabled and _blacksmith_hidden_item_ids.has(str(item.get("item_instance_id", ""))):
			continue
		items.append(item)
	return items


func _target_bag_slot_count() -> int:
	var bag_item_count := int(_bag_items().size())
	var required_slots: int = inventory_capacity if inventory_capacity > bag_item_count else bag_item_count
	if required_slots <= 0:
		return 0
	return int(ceil(float(required_slots) / float(BAG_COLUMNS))) * BAG_COLUMNS


func _apply_interaction_filters() -> void:
	if _panel == null:
		return
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP if _interactive else Control.MOUSE_FILTER_IGNORE


func _position_gesture_hint() -> void:
	if _gesture_hint == null or _panel == null:
		return
	_gesture_hint.set_anchors_preset(Control.PRESET_TOP_WIDE)
	_gesture_hint.offset_top = 4
	_gesture_hint.offset_bottom = 22


func _fill_slot(slot: InventorySlotButton, item: Dictionary) -> void:
	if slot == null:
		return
	slot.item = item.duplicate(true)
	if item.is_empty():
		if InventoryTransferRouterScript.is_equipment_slot(slot.slot_kind):
			var slot_name := _slot_from_kind(slot.slot_kind)
			slot.text = str(EQUIPMENT_LABELS.get(slot_name, slot_name))
			slot.tooltip_text = "Empty %s" % str(EQUIPMENT_LABELS.get(slot_name, slot_name))
			slot.add_theme_stylebox_override("normal", InventoryPanelStylesScript.empty_slot_style(false))
			slot.add_theme_stylebox_override("hover", InventoryPanelStylesScript.empty_slot_style(true))
			slot.add_theme_stylebox_override("pressed", InventoryPanelStylesScript.empty_slot_style(true))
		else:
			slot.text = ""
			slot.tooltip_text = "Empty"
			slot.add_theme_stylebox_override("normal", InventoryPanelStylesScript.slot_style(false))
			slot.add_theme_stylebox_override("hover", InventoryPanelStylesScript.slot_style(true))
			slot.add_theme_stylebox_override("pressed", InventoryPanelStylesScript.slot_style(true))
		slot.queue_redraw()
		return
	slot.text = ""
	slot.tooltip_text = InventoryTooltipContent.tooltip_text(item, _tooltip_context())
	var rarity := str(item.get("rarity", "common"))
	var invalid_requirements := _item_shows_requirement_warning(item)
	if bool(item.get("_blocked_by_two_handed", false)):
		slot.tooltip_text = "%s\nOccupies both hands" % InventoryTooltipContent.tooltip_text(item, _tooltip_context())
		slot.add_theme_stylebox_override("normal", InventoryPanelStylesScript.blocked_slot_style(false))
		slot.add_theme_stylebox_override("hover", InventoryPanelStylesScript.blocked_slot_style(true))
		slot.add_theme_stylebox_override("pressed", InventoryPanelStylesScript.blocked_slot_style(true))
	else:
		slot.add_theme_stylebox_override("normal", InventoryPanelStylesScript.item_slot_style(rarity, false, invalid_requirements))
		slot.add_theme_stylebox_override("hover", InventoryPanelStylesScript.item_slot_style(rarity, true, invalid_requirements))
		slot.add_theme_stylebox_override("pressed", InventoryPanelStylesScript.item_slot_style(rarity, true, invalid_requirements))
	slot.queue_redraw()


func _slot_button(kind: String, size: Vector2) -> InventorySlotButton:
	var btn := InventorySlotButton.new()
	btn.panel = self
	btn.slot_kind = kind
	btn.custom_minimum_size = size
	btn.focus_mode = Control.FOCUS_NONE
	btn.clip_text = true
	btn.add_theme_stylebox_override("normal", InventoryPanelStylesScript.slot_style(false))
	btn.add_theme_stylebox_override("hover", InventoryPanelStylesScript.slot_style(true))
	btn.add_theme_stylebox_override("pressed", InventoryPanelStylesScript.slot_style(true))
	btn.add_theme_color_override("font_color", Color("#e8dcc8"))
	btn.add_theme_font_size_override("font_size", SLOT_FONT_SIZE)
	return btn


func _draw_item_icon(slot: Control, item: Dictionary) -> void:
	var def_id := str(item.get("item_def_id", ""))
	var icon: Dictionary = item_presentations.get(def_id, {}).get("icon", {})
	var rect := Rect2(Vector2.ZERO, slot.size)
	var label := PotionIconLabelScript.icon_label(item, str(icon.get("label", _short_label(def_id))))
	var blocked := bool(item.get("_blocked_by_two_handed", false))
	var invalid_requirements := _item_shows_requirement_warning(item)
	ItemIconDrawerScript.draw(slot, rect, icon, label, blocked or invalid_requirements, 0.38, ICON_FONT_SIZE, item, true)
	if blocked:
		slot.draw_rect(rect.grow(-3.0), Color(0.05, 0.05, 0.05, 0.46), true)
		return
	if invalid_requirements:
		slot.draw_rect(rect.grow(-4.0), Color(1.0, 0.35, 0.35, 0.30), true)
	_draw_hotbar_badge(slot, item)


func _draw_hotbar_badge(slot: Control, item: Dictionary) -> void:
	var assigned_slots := InventoryTooltipContent.hotbar_slots_for_item(hotbar, str(item.get("item_instance_id", "")))
	if assigned_slots.is_empty():
		return
	var label := "H%s" % InventoryTooltipContent.hotbar_label_for_slot(int(assigned_slots[0]))
	if assigned_slots.size() > 1:
		label = "H+"
	var badge_rect := Rect2(Vector2(slot.size.x - 28.0, 3.0), Vector2(24.0, 16.0))
	slot.draw_rect(badge_rect, Color("#15110a"), true)
	slot.draw_rect(badge_rect, Color("#d6a94d"), false, 1.0)
	var font := slot.get_theme_default_font()
	var font_size := 10
	var text_size := font.get_string_size(label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size)
	slot.draw_string(font, badge_rect.position + Vector2((badge_rect.size.x - text_size.x) * 0.5, 12.0), label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, Color("#f4ead8"))


func _title(text: String) -> Label:
	var label := _caption(text)
	label.add_theme_font_size_override("font_size", TITLE_FONT_SIZE)
	return label


func _caption(text: String) -> Label:
	var label := Label.new()
	label.text = text
	label.add_theme_color_override("font_color", Color("#c9a227"))
	label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	return label


func _tooltip_context() -> InventoryTooltipContent.Context:
	return InventoryTooltipContent.Context.new(hotbar)


func _tooltip_lines(item: Dictionary) -> Array:
	return InventoryTooltipContent.tooltip_lines(item, _tooltip_context())


func _make_item_tooltip(item: Dictionary) -> Control:
	var tooltip := ItemTooltipPanelScript.new()
	tooltip.setup(
		item,
		item_presentations,
		InventoryTooltipContent.tooltip_lines(item, _tooltip_context()),
		InventoryTooltipContent.requirement_lines(item),
		InventoryTooltipContent.comparison_entries(item) + UniqueEffectTooltipScript.rich_lines_for_item(item),
		InventoryItemPricing.item_gold_value(item),
		true,
		_short_label(str(item.get("item_def_id", ""))),
		preload("res://scripts/class_affinity_tooltip.gd").equipment_bonus_lines_for_item(item, str(character_progression.get("character_class", "")))
	)
	return tooltip


func _make_text_tooltip(text: String) -> Control:
	var tooltip := ItemTooltipPanelScript.new()
	tooltip.setup({}, item_presentations, [text], [], [], -1, true, "")
	return tooltip


func _reposition_panel() -> void:
	if _panel == null:
		return
	var margin := 20.0
	var panel_size := _panel.custom_minimum_size
	var viewport_size := get_viewport_rect().size if is_inside_tree() else Vector2(1280, 720)
	var bottom_margin := maxf(margin, minf(140.0, viewport_size.y * 0.16))
	_panel.position = Vector2(
		maxf(margin, viewport_size.x - panel_size.x - margin),
		maxf(margin, viewport_size.y - panel_size.y - bottom_margin)
	)


func _handle_double_click(item: Dictionary) -> void:
	var slot := _preferred_equip_slot(item)
	var decision := InventoryTransferRouterScript.double_click_route(
		item,
		_shop_sell_entity_id,
		_market_context,
		_blacksmith_context_enabled,
		_is_equipped_instance(str(item.get("item_instance_id", ""))),
		slot,
		_weapon_set_payload_for_slot(slot),
		_is_consumable(item)
	)
	_apply_transfer_decision(decision, {})


func _handle_shift_click(item: Dictionary) -> void:
	if not _is_consumable(item):
		return
	var slot_index := _first_empty_hotbar_slot()
	if slot_index < 0:
		show_gesture_hint("belt full")
		return
	_apply_transfer_decision(InventoryTransferRouterScript.shift_click_route(item, true, slot_index), {})


func _handle_drop_on_slot(slot_kind: String, data: Variant) -> void:
	if typeof(data) != TYPE_DICTIONARY:
		return
	var item: Dictionary = data.get("item", {})
	if item.is_empty():
		return
	var can_equip_to_slot := false
	var weapon_payload := {}
	if InventoryTransferRouterScript.is_equipment_slot(slot_kind):
		var slot := _slot_from_kind(slot_kind)
		can_equip_to_slot = _item_can_equip_to(item, slot)
		weapon_payload = _weapon_set_payload_for_slot(slot)
	elif slot_kind == SLOT_KIND_BAG and InventoryTransferRouterScript.is_equipment_slot(str(data.get("source", ""))):
		weapon_payload = _weapon_set_payload_for_slot(_slot_from_kind(str(data.get("source", ""))))
	_apply_transfer_decision(InventoryTransferRouterScript.drop_route(slot_kind, data, can_equip_to_slot, weapon_payload), data)


func _apply_transfer_decision(decision: Dictionary, data: Dictionary) -> void:
	match str(decision.get("kind", "")):
		InventoryTransferRouterScript.KIND_INTENT:
			intent_requested.emit(str(decision.get("intent_type", "")), decision.get("payload", {}))
		InventoryTransferRouterScript.KIND_BLACKSMITH_UNSTAGE:
			var blacksmith = data.get("blacksmith_panel", null)
			var unstage_kind := str(decision.get("unstage", "item"))
			if blacksmith != null:
				if unstage_kind == "resource" and blacksmith.has_method("unstage_resource"):
					blacksmith.call("unstage_resource")
				elif blacksmith.has_method("unstage_item"):
					blacksmith.call("unstage_item")
		InventoryTransferRouterScript.KIND_BLACKSMITH_MERGE_UNPLACE:
			var merge_panel = data.get("merge_panel", null)
			if merge_panel != null and merge_panel.has_method("clear_merge_slot"):
				merge_panel.call("clear_merge_slot", int(data.get("slot_index", -1)))


func _item_shows_requirement_warning(item: Dictionary) -> bool:
	return ItemRequirementViewsScript.shows_invalid_requirement_warning(
		item,
		bool(InventoryTooltipContent.item_definition_for_item(item).get("equippable", false))
	)


func _item_can_equip_to(item: Dictionary, slot: String) -> bool:
	var def: Dictionary = InventoryTooltipContent.item_definition_for_item(item)
	if not bool(def.get("equippable", false)):
		return false
	var item_slot := str(def.get("slot", ""))
	if item_slot == "ring":
		return slot == "ring_left" or slot == "ring_right"
	if slot == "off_hand" and item_slot == "main_hand":
		return _can_rogue_offhand_weapon(def)
	return item_slot == slot


func _preferred_equip_slot(item: Dictionary) -> String:
	var def: Dictionary = InventoryTooltipContent.item_definition_for_item(item)
	if not bool(def.get("equippable", false)):
		return ""
	var item_slot := str(def.get("slot", ""))
	if item_slot == "ring":
		if equipped.get("ring_left", null) == null:
			return "ring_left"
		return "ring_right"
	if item_slot == "main_hand" and _can_rogue_offhand_weapon(def) and equipped.get("main_hand", null) != null and equipped.get("off_hand", null) == null:
		return "off_hand"
	return item_slot


func _can_rogue_offhand_weapon(def: Dictionary) -> bool:
	return str(character_progression.get("character_class", "")) == "rogue" \
		and str(def.get("handedness", "")) == "one_handed" \
		and str(def.get("attack_mode", "melee")) == "melee"


func _is_consumable(item: Dictionary) -> bool:
	var def_id := str(item.get("item_def_id", ""))
	var def: Dictionary = InventoryTooltipContent.item_definition(def_id)
	return str(def.get("category", "")) == "consumable"


func _equipped_item(slot: String) -> Dictionary:
	var item_id = WeaponSetTabsScript.hand_equipped_id(weapon_sets, equipped, viewed_weapon_set, slot)
	if item_id == null:
		return {}
	for item in inventory:
		if str(item.get("item_instance_id", "")) == str(item_id):
			return item
	return {}


func _equipment_slot_display_item(slot: String) -> Dictionary:
	var item := _equipped_item(slot)
	if not item.is_empty():
		return item
	if slot == "off_hand":
		var main_hand := _equipped_item("main_hand")
		if _item_occupies_off_hand(main_hand):
			var blocked := main_hand.duplicate(true)
			blocked["_blocked_by_two_handed"] = true
			blocked["_blocked_slot"] = "off_hand"
			return blocked
	return {}


func _item_occupies_off_hand(item: Dictionary) -> bool:
	if item.is_empty():
		return false
	var def := InventoryTooltipContent.item_definition_for_item(item)
	if str(def.get("handedness", "")) == "two_handed":
		return true
	var occupies = def.get("occupies_hands", [])
	if typeof(occupies) == TYPE_ARRAY and (occupies as Array).has("off_hand"):
		return true
	var item_occupies = item.get("occupies_hands", [])
	return typeof(item_occupies) == TYPE_ARRAY and (item_occupies as Array).has("off_hand")


func _is_equipped_instance(item_instance_id: String) -> bool:
	return WeaponSetTabsScript.is_equipped_instance(equipped, weapon_sets, EQUIPMENT_SLOTS, item_instance_id)


func _slot_kind_for_equipment(slot: String) -> String:
	return SLOT_KIND_EQUIP_PREFIX + slot


func _slot_from_kind(kind: String) -> String:
	return InventoryTransferRouterScript.slot_from_kind(kind)


func _first_empty_hotbar_slot() -> int:
	for slot_index in range(hotbar_capacity):
		if not _hotbar_slot_has_item(slot_index):
			return slot_index
	return -1


func _hotbar_slot_has_item(slot_index: int) -> bool:
	for slot in hotbar:
		if typeof(slot) != TYPE_DICTIONARY:
			continue
		var rec := slot as Dictionary
		if int(rec.get("slot_index", -1)) != slot_index:
			continue
		var assigned_id = rec.get("item_instance_id", null)
		return assigned_id != null and str(assigned_id) != ""
	return false


func _hotbar_assigned_inventory_count() -> int:
	var total := 0
	for item in _bag_items():
		if not InventoryTooltipContent.hotbar_slots_for_item(hotbar, str((item as Dictionary).get("item_instance_id", ""))).is_empty():
			total += 1
	return total


func _requirement_row_count() -> int:
	var total := 0
	for item in inventory:
		if typeof(item) == TYPE_DICTIONARY:
			total += InventoryTooltipContent.requirement_lines(item as Dictionary).size()
	for slot in EQUIPMENT_SLOTS:
		var item := _equipped_item(str(slot))
		if not item.is_empty():
			total += InventoryTooltipContent.requirement_lines(item).size()
	return total


func _equip_preview_row_count() -> int:
	var total := 0
	for item in inventory:
		if typeof(item) == TYPE_DICTIONARY:
			total += InventoryTooltipContent.equip_preview_count(item as Dictionary)
	for slot in EQUIPMENT_SLOTS:
		var item := _equipped_item(str(slot))
		if not item.is_empty():
			total += InventoryTooltipContent.equip_preview_count(item)
	return total


func _short_label(def_id: String) -> String:
	var def: Dictionary = item_rules.get(def_id, {})
	var name := str(def.get("name", def_id))
	var parts := name.split(" ")
	var out := ""
	for part in parts:
		if part.length() > 0:
			out += part.substr(0, 1).to_upper()
	return out.substr(0, 3)


func _debug_presentations() -> Dictionary:
	var out := {}
	for item in inventory:
		var def_id := str(item.get("item_def_id", ""))
		if def_id != "":
			out[def_id] = item_presentations.has(def_id)
	return out


func _debug_hotbar_assigned_item_ids() -> Array:
	var ids: Array = []
	for item in _bag_items():
		var item_id := str((item as Dictionary).get("item_instance_id", ""))
		if item_id != "" and not InventoryTooltipContent.hotbar_slots_for_item(hotbar, item_id).is_empty():
			ids.append(item_id)
	return ids


func _sync_paper_doll_backdrop() -> void:
	if _paper_doll_preview != null:
		_paper_doll_preview.configure(str(character_progression.get("character_class", "")), PaperDollLayoutScript.slot_rects(EQUIPMENT_SLOT_SIZE))


func _debug_paper_doll_slots() -> Dictionary:
	var out := {}
	for slot in EQUIPMENT_SLOTS:
		var btn: InventorySlotButton = _equipment_slots.get(str(slot), null)
		var display_item := _equipment_slot_display_item(str(slot))
		out[str(slot)] = {
			"exists": btn != null,
			"position": {
				"x": btn.position.x if btn != null else 0.0,
				"y": btn.position.y if btn != null else 0.0,
			},
			"empty": display_item.is_empty(),
			"blocked_by_two_handed": bool(display_item.get("_blocked_by_two_handed", false)),
			"item_def_id": str(display_item.get("item_def_id", "")),
			"label": str(EQUIPMENT_LABELS.get(str(slot), str(slot))),
		}
	return out


func _render_resources_button() -> void:
	if _resources_button == null:
		return
	var rows := _wallet_rows()
	if rows.is_empty():
		_resources_button.tooltip_text = "No account materials stored"
		return
	_resources_button.tooltip_text = "\n".join(_wallet_detail_lines())


func _sync_wallet_window() -> void:
	if _wallet_window != null:
		_wallet_window.set_wallet(_resource_wallet, _resource_bag_items)


func _wallet_rows() -> Array:
	var rows: Array = []
	for key in _wallet_resource_keys():
		var amount := int(_resource_wallet.get(key, 0))
		rows.append("%s %d" % [_resource_label(str(key)), amount])
	return rows


func _wallet_detail_lines() -> Array:
	var lines: Array = []
	for key in _wallet_resource_keys():
		var resource_id := str(key)
		var amount := int(_resource_wallet.get(resource_id, 0))
		lines.append("%s x%d" % [_resource_name(resource_id), amount])
		var category := _resource_category(resource_id)
		if category != "":
			lines.append("Category: %s" % category)
		lines.append("Stored account-wide")
	return lines


func _wallet_resource_keys() -> Array:
	var out: Array = []
	var keys: Array = _resource_wallet.keys()
	keys.sort()
	for key in keys:
		if str(key) == "upgrade_shard":
			continue
		var amount := int(_resource_wallet.get(key, 0))
		if amount <= 0:
			continue
		out.append(key)
	return out


func _resource_label(resource_id: String) -> String:
	if resource_id == "upgrade_shard":
		return "Shard"
	if resource_id == "respec_badge":
		return "Respec"
	if resource_id == "stat_badge":
		return "Stat"
	if resource_id == "skill_badge":
		return "Skill"
	if resource_id == "resurrection_badge":
		return "Resurrect"
	return _resource_name(resource_id)


func _resource_name(resource_id: String) -> String:
	var def := ItemRulesLoader.item_definition(resource_id)
	if def.has("name"):
		return str(def.get("name", ""))
	return resource_id.replace("_", " ").capitalize()


func _resource_category(resource_id: String) -> String:
	var def := ItemRulesLoader.item_definition(resource_id)
	return str(def.get("category", "")).replace("_", " ").capitalize()


func _dup_items(values: Array) -> Array:
	var out: Array = []
	for value in values:
		if typeof(value) == TYPE_DICTIONARY:
			out.append((value as Dictionary).duplicate(true))
	return out
