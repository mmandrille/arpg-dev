## AimReticleOverlay — draws a simple crosshair at viewport center.
##
## Added to gameplay_ui_layer. Visible only when the current camera mode has
## reticle_enabled=true (perspective modes) and gameplay is active.
## Call set_visible(true/false) from main.gd whenever mode or state changes.
class_name AimReticleOverlay
extends Control

const _LINE_LENGTH := 10.0
const _LINE_THICKNESS := 1.5
const _GAP := 4.0
const _COLOR_IDLE := Color(1.0, 1.0, 1.0, 0.82)
const _COLOR_LOCKED := Color(0.83, 0.63, 0.09, 0.95)
const _OUTLINE_COLOR := Color(0.04, 0.05, 0.07, 0.75)
const _OUTLINE_THICKNESS := 3.0

var _locked := false


func set_locked(on: bool) -> void:
	if _locked == on:
		return
	_locked = on
	queue_redraw()


func is_locked() -> bool:
	return _locked


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	z_index = 100
	_sync_viewport_size()
	if get_viewport() != null:
		get_viewport().size_changed.connect(_sync_viewport_size)
	queue_redraw()


func _notification(what: int) -> void:
	if what == NOTIFICATION_VISIBILITY_CHANGED and is_visible_in_tree():
		queue_redraw()


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	queue_redraw()


func _viewport_center() -> Vector2:
	var viewport_size := get_viewport_rect().size
	if viewport_size.x > 1.0 and viewport_size.y > 1.0:
		return viewport_size * 0.5
	if size.x > 1.0 and size.y > 1.0:
		return size * 0.5
	return Vector2(320.0, 180.0)


func _draw() -> void:
	var center := _viewport_center()
	var color := _COLOR_LOCKED if _locked else _COLOR_IDLE
	_draw_crosshair_arm(center, Vector2(-1.0, 0.0), color)
	_draw_crosshair_arm(center, Vector2(1.0, 0.0), color)
	_draw_crosshair_arm(center, Vector2(0.0, -1.0), color)
	_draw_crosshair_arm(center, Vector2(0.0, 1.0), color)


func _draw_crosshair_arm(center: Vector2, axis: Vector2, color: Color) -> void:
	var gap := axis * _GAP
	var arm := axis * _LINE_LENGTH
	var inner := center + gap
	var outer := center + gap + arm
	draw_line(inner, outer, _OUTLINE_COLOR, _OUTLINE_THICKNESS)
	draw_line(inner, outer, color, _LINE_THICKNESS)
