class_name ItemFamilyIconPreview
extends Control

const ItemIconDrawerScript := preload("res://scripts/item_icon_drawer.gd")

var icon: Dictionary = {}


func configure(next_icon: Dictionary) -> void:
	icon = next_icon
	queue_redraw()


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE


func _draw() -> void:
	if size.x <= 0 or size.y <= 0:
		return
	var center := size * 0.5
	var min_side := minf(size.x, size.y)
	draw_circle(center, min_side * 0.42, Color(0.015, 0.014, 0.012, 0.92))
	ItemIconDrawerScript.draw(
		self,
		Rect2(Vector2.ZERO, size),
		icon,
		str(icon.get("label", "")),
		false,
		0.36,
		12
	)
