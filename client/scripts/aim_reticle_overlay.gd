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
const _COLOR := Color(1.0, 1.0, 1.0, 0.5)


func _ready() -> void:
	set_anchors_preset(Control.PRESET_FULL_RECT)
	mouse_filter = Control.MOUSE_FILTER_IGNORE


func _draw() -> void:
	var center := size * 0.5
	# Horizontal arms
	draw_line(
		Vector2(center.x - _LINE_LENGTH - _GAP, center.y),
		Vector2(center.x - _GAP, center.y),
		_COLOR, _LINE_THICKNESS
	)
	draw_line(
		Vector2(center.x + _GAP, center.y),
		Vector2(center.x + _LINE_LENGTH + _GAP, center.y),
		_COLOR, _LINE_THICKNESS
	)
	# Vertical arms
	draw_line(
		Vector2(center.x, center.y - _LINE_LENGTH - _GAP),
		Vector2(center.x, center.y - _GAP),
		_COLOR, _LINE_THICKNESS
	)
	draw_line(
		Vector2(center.x, center.y + _GAP),
		Vector2(center.x, center.y + _LINE_LENGTH + _GAP),
		_COLOR, _LINE_THICKNESS
	)
