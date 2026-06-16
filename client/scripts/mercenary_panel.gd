class_name MercenaryPanel
extends Control

const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const PANEL_SIZE := Vector2(340, 340)
const TITLE_FONT_SIZE := 28
const BODY_FONT_SIZE := 18
const MERCENARY_MONSTER_DEF_ID := "mercenary_guard"

var board_entity_id: String = ""
var service_id: String = "mercenary"
var offer_id: String = "fixed:mercenary_guard"
var monster_def_id: String = MERCENARY_MONSTER_DEF_ID
var price: int = 0
var affordable: bool = false
var gold: int = 0
var hired_entity_id: String = ""

var _panel: DraggableWindow
var _offer_label: Label
var _status_label: Label
var _roster_label: Label
var _companions: Array = []


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_build()
	visible = false


func show_board(next_entity_id: String, next_service_id: String, next_offer_id: String, next_monster_def_id: String, next_price: int, next_affordable: bool, next_gold: int) -> void:
	board_entity_id = next_entity_id
	service_id = next_service_id if next_service_id != "" else "mercenary"
	offer_id = next_offer_id if next_offer_id != "" else "fixed:mercenary_guard"
	monster_def_id = next_monster_def_id if next_monster_def_id != "" else MERCENARY_MONSTER_DEF_ID
	price = max(0, next_price)
	gold = max(0, next_gold)
	affordable = next_affordable
	hired_entity_id = ""
	if _status_label != null:
		_status_label.text = ""
	visible = true
	_render()


func set_gold(next_gold: int) -> void:
	gold = max(0, next_gold)
	affordable = gold >= price
	if _status_label != null and hired_entity_id == "":
		_status_label.text = ""
	_render()


func apply_hired_event(ev: Dictionary) -> void:
	hired_entity_id = str(ev.get("target_entity_id", hired_entity_id))
	monster_def_id = str(ev.get("monster_def_id", monster_def_id))
	price = max(0, int(ev.get("price", price)))
	gold = max(0, int(ev.get("total_gold", gold)))
	affordable = gold >= price
	if _status_label != null:
		_status_label.text = "Hired %s" % _display_name(monster_def_id)
		_status_label.add_theme_color_override("font_color", Color("#9ee6a8"))
	_render()


func set_companions(next_companions: Array) -> void:
	_companions = []
	for companion in next_companions:
		if typeof(companion) != TYPE_DICTIONARY:
			continue
		var rec := (companion as Dictionary).duplicate(true)
		if str(rec.get("monster_def_id", "")).find("mercenary") < 0:
			continue
		_companions.append(rec)
	_companions.sort_custom(func(a: Dictionary, b: Dictionary) -> bool:
		return str(a.get("id", "")) < str(b.get("id", ""))
	)
	_render()


func hide_display() -> void:
	visible = false


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"board_entity_id": board_entity_id,
		"service_id": service_id,
		"offer_id": offer_id,
		"monster_def_id": monster_def_id,
		"price": price,
		"gold": gold,
		"affordable": affordable,
		"hired_entity_id": hired_entity_id,
		"hired_count": _companions.size(),
		"hired_rows": _companions.duplicate(true),
		"status": _status_label.text if _status_label != null else "",
		"window": _panel.get_debug_state() if _panel != null else {},
	}


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_reposition_panel()


func _build() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	_panel = DraggableWindowScript.new()
	_panel.custom_minimum_size = PANEL_SIZE
	_panel.configure("Mercenaries", Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 52))
	_panel.add_theme_stylebox_override("panel", _panel_style())
	_panel.mouse_filter = Control.MOUSE_FILTER_STOP
	_panel.close_requested.connect(hide_display)
	add_child(_panel)
	_reposition_panel()
	_panel.set_layout_key("mercenaries")

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 10)
	root.custom_minimum_size = Vector2(PANEL_SIZE.x - 28, PANEL_SIZE.y - 52)
	_panel.set_content(root)

	var title := Label.new()
	title.text = "Mercenaries"
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.add_theme_font_size_override("font_size", TITLE_FONT_SIZE)
	title.add_theme_color_override("font_color", Color("#f3e7d6"))
	root.add_child(title)

	_offer_label = Label.new()
	_offer_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_offer_label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	_offer_label.add_theme_color_override("font_color", Color("#d9c8b5"))
	root.add_child(_offer_label)

	_status_label = Label.new()
	_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_status_label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	_status_label.add_theme_color_override("font_color", Color("#d9c8b5"))
	root.add_child(_status_label)

	_roster_label = Label.new()
	_roster_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_roster_label.add_theme_font_size_override("font_size", BODY_FONT_SIZE)
	_roster_label.add_theme_color_override("font_color", Color("#d9c8b5"))
	root.add_child(_roster_label)
	_render()


func _render() -> void:
	if _offer_label == null:
		return
	_offer_label.text = "%s\nOffer: %s\nCost: %d gold\nGold: %d" % [
		_display_name(monster_def_id),
		offer_id,
		price,
		gold,
	]
	if _status_label != null and _status_label.text == "":
		_status_label.text = "Ready to hire" if affordable else "Not enough gold"
		_status_label.add_theme_color_override("font_color", Color("#9ee6a8") if affordable else Color("#ff9f7a"))
	if _roster_label != null:
		if _companions.is_empty():
			_roster_label.text = "Hired roster: none"
		else:
			var lines: Array = ["Hired roster:"]
			for companion in _companions:
				var rec := companion as Dictionary
				lines.append("%s  HP %d/%d  ID %s" % [
					_display_name(str(rec.get("monster_def_id", ""))),
					int(rec.get("hp", 0)),
					int(rec.get("max_hp", 0)),
					str(rec.get("id", "")),
				])
			_roster_label.text = "\n".join(lines)


func _reposition_panel() -> void:
	if _panel == null:
		return
	var viewport_size := get_viewport_rect().size
	_panel.position = Vector2(maxf(16.0, viewport_size.x - PANEL_SIZE.x - 22.0), 106.0)


func _display_name(id: String) -> String:
	if id == "mercenary_guard":
		return "Mercenary Guard"
	return id.capitalize()


func _panel_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color("#17202a")
	style.border_color = Color("#617d91")
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
