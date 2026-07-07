class_name QuestStewardPanel
extends Control

signal intent_requested(intent_type: String, payload: Dictionary)

const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const PANEL_SIZE := Vector2(360, 320)

var quest_giver_entity_id: String = ""
var _panel: DraggableWindow
var _summary_label: Label
var _offers_box: VBoxContainer
var _status_label: Label
var _interactive: bool = true


func _ready() -> void:
	_build()
	visible = false


func show_offers(entity_id: String, offers: Array, trophy_name: String = "") -> void:
	quest_giver_entity_id = entity_id
	if _panel == null:
		_build()
	visible = true
	if _status_label != null:
		_status_label.text = ""
	if _summary_label != null:
		var label := trophy_name if trophy_name != "" else "Quest trophy"
		_summary_label.text = "Choose a reward family for %s:" % label
	_render_offers(offers)
	if _panel != null:
		_panel.clamp_to_viewport()


func hide_display() -> void:
	visible = false


func show_status(text: String, is_error: bool = false) -> void:
	if _status_label == null:
		return
	_status_label.text = text
	_status_label.add_theme_color_override("font_color", Color("#ff8a8a") if is_error else Color("#9be7a8"))


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"quest_giver_entity_id": quest_giver_entity_id,
		"offer_count": _offers_box.get_child_count() if _offers_box != null else 0,
		"status": _status_label.text if _status_label != null else "",
	}


func _build() -> void:
	if _panel != null:
		return
	_panel = DraggableWindowScript.new()
	_panel.configure("Quest Steward", PANEL_SIZE)
	_panel.custom_minimum_size = Vector2(PANEL_SIZE.x, PANEL_SIZE.y + DraggableWindowScript.TITLEBAR_HEIGHT)
	_panel.size = _panel.custom_minimum_size
	_panel.set_layout_key("quest_steward_panel")
	_panel.position = Vector2(96, 120)
	_panel.close_requested.connect(hide_display)
	add_child(_panel)
	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 8)
	_panel.set_content(root)
	_summary_label = Label.new()
	_summary_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_summary_label.add_theme_font_size_override("font_size", 15)
	_summary_label.add_theme_color_override("font_color", Color("#f0dfbb"))
	root.add_child(_summary_label)
	_offers_box = VBoxContainer.new()
	_offers_box.add_theme_constant_override("separation", 6)
	root.add_child(_offers_box)
	_status_label = Label.new()
	_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_status_label.add_theme_font_size_override("font_size", 13)
	root.add_child(_status_label)


func _render_offers(offers: Array) -> void:
	if _offers_box == null:
		return
	for child in _offers_box.get_children():
		_offers_box.remove_child(child)
		child.queue_free()
	for offer in offers:
		if not offer is Dictionary:
			continue
		var row: Dictionary = offer
		var button := Button.new()
		button.text = str(row.get("label", row.get("family_id", "Reward")))
		button.disabled = not _interactive
		var offer_id := str(row.get("offer_id", ""))
		button.pressed.connect(func() -> void:
			if offer_id == "" or quest_giver_entity_id == "":
				return
			intent_requested.emit("quest_steward_pick_intent", {
				"quest_giver_entity_id": quest_giver_entity_id,
				"offer_id": offer_id,
			})
		)
		_offers_box.add_child(button)


func bot_click_offer(offer_index: int = 0) -> void:
	if _offers_box == null:
		return
	if offer_index < 0 or offer_index >= _offers_box.get_child_count():
		return
	var button = _offers_box.get_child(offer_index)
	if button is Button:
		(button as Button).emit_signal("pressed")
