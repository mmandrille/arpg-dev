class_name BishopLootDebugPanel
extends Control

signal catalog_requested(bishop_entity_id: String)
signal source_catalog_requested(bishop_entity_id: String, depth: int, source_type: String)
signal force_loot_requested(payload: Dictionary)

const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const PANEL_SIZE := Vector2(360, 460)

enum Step { DEPTH, SOURCE, PICK, ITEM_LEVEL }

var bishop_entity_id: String = ""
var _panel: DraggableWindow
var _title_label: Label
var _body_label: Label
var _actions_scroll: ScrollContainer
var _actions_box: VBoxContainer
var _filter_input: LineEdit
var _status_label: Label
var _step: int = Step.DEPTH
var _filter_text: String = ""
var _last_filter_step: int = -1
var _depth_catalog: Dictionary = {}
var _source_catalog: Dictionary = {}
var _selected_depth: int = 0
var _selected_source: String = ""
var _pending_pick: Dictionary = {}


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_build()
	visible = false


func open_for_bishop(next_entity_id: String) -> void:
	bishop_entity_id = next_entity_id
	_reset_flow()
	visible = true
	catalog_requested.emit(bishop_entity_id)


func hide_display() -> void:
	visible = false


func apply_depth_catalog(catalog: Dictionary) -> void:
	_depth_catalog = catalog.duplicate(true)
	_render()


func apply_source_catalog(catalog: Dictionary) -> void:
	_source_catalog = catalog.duplicate(true)
	_step = Step.PICK
	_render()


func show_status(text: String, error: bool = false) -> void:
	if _status_label == null:
		return
	_status_label.text = text
	_status_label.add_theme_color_override("font_color", Color("#ff9f7a") if error else Color("#9ee6a8"))


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"step": _step,
		"selected_depth": _selected_depth,
		"selected_source": _selected_source,
		"pending_pick": _pending_pick.duplicate(true),
		"depth_catalog": _depth_catalog.duplicate(true),
		"source_catalog": _source_catalog.duplicate(true),
		"status": _status_label.text if _status_label != null else "",
		"filter_text": _filter_text,
		"filter_visible": _filter_input.visible if _filter_input != null else false,
	}


func bot_force_pick(payload: Dictionary) -> void:
	if bishop_entity_id == "":
		return
	force_loot_requested.emit(payload.duplicate(true))


func _reset_flow() -> void:
	_step = Step.DEPTH
	_last_filter_step = -1
	_depth_catalog = {}
	_source_catalog = {}
	_selected_depth = 0
	_selected_source = ""
	_pending_pick = {}
	if _status_label != null:
		_status_label.text = ""
	_render()


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_reposition_panel()


func _build() -> void:
	_panel = DraggableWindowScript.new()
	_panel.custom_minimum_size = PANEL_SIZE
	_panel.configure("Force loot drop", Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 52))
	_panel.add_theme_stylebox_override("panel", _panel_style())
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	_panel.close_requested.connect(hide_display)
	add_child(_panel)
	_reposition_panel()
	_panel.set_layout_key("bishop_loot_debug")

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 10)
	root.custom_minimum_size = Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 52)
	_panel.set_content(root)

	_title_label = Label.new()
	_title_label.text = "Force loot drop"
	_title_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_title_label.add_theme_font_size_override("font_size", 26)
	_title_label.add_theme_color_override("font_color", Color("#f4e5d2"))
	root.add_child(_title_label)

	_body_label = Label.new()
	_body_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_body_label.add_theme_font_size_override("font_size", 18)
	_body_label.add_theme_color_override("font_color", Color("#d9c8b5"))
	root.add_child(_body_label)

	_filter_input = LineEdit.new()
	_filter_input.placeholder_text = "Filter options…"
	_filter_input.custom_minimum_size = Vector2(PANEL_SIZE.x - 60, 32)
	_filter_input.visible = false
	_filter_input.text_changed.connect(_on_filter_text_changed)
	root.add_child(_filter_input)

	_actions_scroll = ScrollContainer.new()
	_actions_scroll.custom_minimum_size = Vector2(PANEL_SIZE.x - 60, 280)
	_actions_scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_actions_scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	_actions_scroll.vertical_scroll_mode = ScrollContainer.SCROLL_MODE_SHOW_ALWAYS
	root.add_child(_actions_scroll)

	_actions_box = VBoxContainer.new()
	_actions_box.add_theme_constant_override("separation", 8)
	_actions_box.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_actions_scroll.add_child(_actions_box)

	_status_label = Label.new()
	_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_status_label.add_theme_font_size_override("font_size", 18)
	_status_label.add_theme_color_override("font_color", Color("#9ee6a8"))
	root.add_child(_status_label)
	_render()


func _render() -> void:
	if _step != _last_filter_step:
		_clear_filter()
		_last_filter_step = _step
	if _filter_input != null:
		_filter_input.visible = _step == Step.PICK or _step == Step.ITEM_LEVEL
	_clear_actions()
	var option_count := 0
	match _step:
		Step.DEPTH:
			_body_label.text = "Pick dungeon depth (capped at deepest reached)."
			_add_button("Back", _on_back_pressed, _selected_depth > 0 or _selected_source != "", true)
			for depth_row in _depth_catalog.get("depths", []):
				var depth := int(depth_row.get("depth", 0))
				var label := str(depth_row.get("label", "Depth %d" % depth))
				option_count += _add_button(label, _select_depth.bind(depth))
		Step.SOURCE:
			_body_label.text = "Depth %d — pick drop source." % _selected_depth
			_add_button("Back", _on_back_pressed, true, true)
			for source_type in ["monster", "chest", "boss", "boss_chest"]:
				option_count += _add_button(source_type.replace("_", " ").capitalize(), _select_source.bind(source_type))
		Step.PICK:
			_body_label.text = "Depth %d / %s — pick outcome." % [_selected_depth, _selected_source]
			_add_button("Back", _on_back_pressed, true, true)
			option_count += _render_pick_buttons()
		Step.ITEM_LEVEL:
			_body_label.text = "Pick item level (max %d)." % int(_source_catalog.get("max_item_level", 1))
			_add_button("Back", _on_back_pressed, true, true)
			var max_level := maxi(1, int(_source_catalog.get("max_item_level", 1)))
			for level in range(1, max_level + 1):
				option_count += _add_button("Item level %d" % level, _confirm_force.bind(level))
	if option_count == 0 and _filter_text.strip_edges() != "":
		_add_no_matches_label()


func _render_pick_buttons() -> int:
	var option_count := 0
	for attempt in _source_catalog.get("attempts", []):
		var attempt_id := str(attempt.get("attempt_id", "primary"))
		var success_weight := int(attempt.get("success_weight", 0))
		var no_drop_weight := int(attempt.get("no_drop_weight", 0))
		option_count += _add_button("%s: drop (%d%%)" % [attempt_id, success_weight], _pick_treasure_branch.bind(attempt_id))
		if no_drop_weight > 0:
			option_count += _add_button("%s: no drop (%d%%)" % [attempt_id, no_drop_weight], _show_no_drop_hint)
		for entry in attempt.get("entries", []):
			var entry_index := int(entry.get("entry_index", 0))
			var label := str(entry.get("label", "entry"))
			var weight := int(entry.get("weight", 0))
			option_count += _add_button(
				"%s / %s (%d)" % [attempt_id, label, weight],
				_pick_treasure_entry.bind(attempt_id, entry_index, entry.duplicate(true))
			)
	var resource_loot: Dictionary = _source_catalog.get("resource_loot", {})
	if not resource_loot.is_empty():
		var chance := int(resource_loot.get("chance_percent", 0))
		option_count += _add_button(
			"Resource loot: drop (%d%%)" % chance,
			func() -> void: show_status("Pick a resource item below.", false)
		)
		for entry in resource_loot.get("pool", []):
			var item_def_id := str(entry.get("item_def_id", ""))
			var label := str(entry.get("label", item_def_id))
			option_count += _add_button("Resource: %s" % label, _pick_resource_entry.bind(item_def_id, entry.duplicate(true)))
	for wallet_item in _source_catalog.get("wallet_items", []):
		var item_def_id := str(wallet_item.get("item_def_id", ""))
		var label := str(wallet_item.get("label", item_def_id))
		option_count += _add_button("Wallet: %s" % label, _pick_wallet_item.bind(item_def_id))
	return option_count


func _show_no_drop_hint() -> void:
	show_status("No drop selected — pick a drop branch to spawn loot.", true)


func _pick_treasure_branch(attempt_id: String) -> void:
	show_status("Now pick an entry under %s." % attempt_id, false)


func _pick_treasure_entry(attempt_id: String, entry_index: int, entry: Dictionary) -> void:
	_pending_pick = {
		"drop_kind": "treasure_entry",
		"attempt_id": attempt_id,
		"entry_index": entry_index,
		"supports_item_level": bool(entry.get("supports_item_level", false)),
	}
	if bool(entry.get("supports_item_level", false)):
		_step = Step.ITEM_LEVEL
		_render()
		return
	_confirm_force(1)


func _pick_resource_entry(item_def_id: String, entry: Dictionary) -> void:
	_pending_pick = {
		"drop_kind": "resource_pool",
		"item_def_id": item_def_id,
		"supports_item_level": bool(entry.get("supports_item_level", false)),
	}
	if bool(entry.get("supports_item_level", false)):
		_step = Step.ITEM_LEVEL
		_render()
		return
	_confirm_force(1)


func _pick_wallet_item(item_def_id: String) -> void:
	_pending_pick = {
		"drop_kind": "wallet_item",
		"item_def_id": item_def_id,
		"supports_item_level": false,
	}
	_confirm_force(1)


func _confirm_force(item_level: int) -> void:
	if bishop_entity_id == "":
		return
	var payload := {
		"bishop_entity_id": bishop_entity_id,
		"depth": _selected_depth,
		"source_type": _selected_source,
		"drop_kind": str(_pending_pick.get("drop_kind", "")),
		"item_level": item_level,
	}
	if _pending_pick.has("attempt_id"):
		payload["attempt_id"] = str(_pending_pick.get("attempt_id", ""))
	if _pending_pick.has("entry_index"):
		payload["entry_index"] = int(_pending_pick.get("entry_index", 0))
	if _pending_pick.has("item_def_id"):
		payload["item_def_id"] = str(_pending_pick.get("item_def_id", ""))
	force_loot_requested.emit(payload)


func _select_depth(depth: int) -> void:
	_selected_depth = depth
	_step = Step.SOURCE
	_render()


func _select_source(source_type: String) -> void:
	_selected_source = source_type
	source_catalog_requested.emit(bishop_entity_id, _selected_depth, source_type)


func _on_back_pressed() -> void:
	match _step:
		Step.SOURCE:
			_step = Step.DEPTH
		Step.PICK:
			_step = Step.SOURCE
			_source_catalog = {}
		Step.ITEM_LEVEL:
			_step = Step.PICK
		_:
			_step = Step.DEPTH
	_render()


func _on_filter_text_changed(text: String) -> void:
	_filter_text = text
	_render()


func _clear_filter() -> void:
	_filter_text = ""
	if _filter_input != null:
		_filter_input.text = ""


func _matches_filter(text: String) -> bool:
	var needle := _filter_text.strip_edges().to_lower()
	if needle == "":
		return true
	return text.to_lower().contains(needle)


func _clear_actions() -> void:
	if _actions_box == null:
		return
	for child in _actions_box.get_children():
		child.queue_free()


func _add_no_matches_label() -> void:
	var label := Label.new()
	label.text = "No matching options"
	label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	label.add_theme_font_size_override("font_size", 16)
	label.add_theme_color_override("font_color", Color("#b8a898"))
	_actions_box.add_child(label)


func _add_button(text: String, callback: Callable, enabled: bool = true, always_show: bool = false) -> int:
	if not always_show and not _matches_filter(text):
		return 0
	var button := Button.new()
	button.text = text
	button.custom_minimum_size = Vector2(PANEL_SIZE.x - 72, 34)
	button.disabled = not enabled
	if callback is Callable:
		button.pressed.connect(callback)
	_actions_box.add_child(button)
	return 1


func _reposition_panel() -> void:
	if _panel == null:
		return
	var viewport_size := get_viewport_rect().size
	_panel.position = Vector2(maxf(16.0, viewport_size.x - PANEL_SIZE.x - 22.0), 92.0)


func _panel_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color("#1a1818")
	style.border_color = Color("#6d2424")
	style.border_width_left = 2
	style.border_width_top = 2
	style.border_width_right = 2
	style.border_width_bottom = 2
	style.corner_radius_top_left = 8
	style.corner_radius_top_right = 8
	style.corner_radius_bottom_left = 8
	style.corner_radius_bottom_right = 8
	style.content_margin_left = 12
	style.content_margin_top = 12
	style.content_margin_right = 12
	style.content_margin_bottom = 12
	return style
