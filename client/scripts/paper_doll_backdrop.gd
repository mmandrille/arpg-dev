class_name PaperDollBackdrop
extends Control
## Class-tinted backdrop for the inventory paper-doll (v516 D5): a rounded body card,
## a class badge, and connector lines from each equipment slot to the body axis.
## Self-contained: knows nothing about the inventory panel. Callers pass the active class
## id and slot rectangles in this node's local coordinates.

const ClassIconScript := preload("res://scripts/class_icon.gd")
const InventoryPanelStylesScript := preload("res://scripts/inventory_panel_styles.gd")

const EMBLEM_ALPHA := 0.9
const EMBLEM_SIZE := 52.0

var class_id: String = ""
var accent: Color = Color.WHITE
## Card spans the whole slot field.
var body_rect: Rect2 = Rect2(6, 2, 328, 336)
## Class badge sits in the empty top-left cell of the PaperDollLayout grid.
var emblem_center: Vector2 = Vector2(56, 37)
var slot_rects: Dictionary = {}
var _card: Panel
var _emblem: Control


func _init() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	name = "character_paper_doll"
	_card = Panel.new()
	_card.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(_card)
	_emblem = ClassIconScript.new()
	_emblem.modulate = Color(1, 1, 1, EMBLEM_ALPHA)
	_emblem.custom_minimum_size = Vector2(EMBLEM_SIZE, EMBLEM_SIZE)
	_emblem.size = _emblem.custom_minimum_size
	add_child(_emblem)
	configure("")


## `slots` maps slot id -> Rect2 in this node's local space (optional; no connectors without it).
func configure(next_class_id: String, slots: Dictionary = {}, next_body_rect: Rect2 = Rect2()) -> void:
	class_id = next_class_id.strip_edges()
	slot_rects = slots.duplicate()
	if next_body_rect.size.x > 0.0 and next_body_rect.size.y > 0.0:
		body_rect = next_body_rect
	accent = InventoryPanelStylesScript.paper_doll_neutral_accent()
	_emblem.visible = class_id != ""
	if class_id != "":
		_emblem.call("configure", class_id)
		accent = _emblem.fill_color
	_card.position = body_rect.position
	_card.size = body_rect.size
	_card.add_theme_stylebox_override("panel", InventoryPanelStylesScript.paper_doll_card_style(accent))
	_emblem.position = emblem_center - _emblem.size * 0.5
	queue_redraw()


func connector_count() -> int:
	return slot_rects.size()


func get_backdrop_state() -> Dictionary:
	return {
		"class_id": class_id,
		"accent": accent.to_html(false),
		"emblem_visible": _emblem.visible,
		"connector_count": connector_count(),
	}


func _draw() -> void:
	var axis_x := body_rect.get_center().x
	var spine := Rect2(axis_x - 48.0, body_rect.position.y, 96.0, body_rect.size.y)
	var color := InventoryPanelStylesScript.paper_doll_connector_color(accent)
	for slot_id in slot_rects:
		var rect: Rect2 = slot_rects[slot_id]
		var from := rect.get_center()
		# Side slots join the central slot column with a thin line at the same height.
		if rect.intersects(spine):
			continue
		var start := Vector2(rect.end.x if from.x < axis_x else rect.position.x, from.y)
		var edge_x := spine.position.x if from.x < axis_x else spine.end.x
		draw_line(start, Vector2(edge_x, from.y), color, 2.0, true)
