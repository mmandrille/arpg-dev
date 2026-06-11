class_name ClassIcon
extends Control

var class_id: String = "barbarian"
var fill_color := Color("#c85f3d")
var accent_color := Color("#ffd9a8")


func configure(next_class_id: String) -> void:
	class_id = next_class_id
	match class_id:
		"sorcerer":
			fill_color = Color("#5e8cff")
			accent_color = Color("#dce7ff")
		"paladin":
			fill_color = Color("#d9b44a")
			accent_color = Color("#fff3ba")
		_:
			fill_color = Color("#c85f3d")
			accent_color = Color("#ffd9a8")
	queue_redraw()


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE


func _draw() -> void:
	var rect := Rect2(Vector2.ZERO, size)
	if rect.size.x <= 0 or rect.size.y <= 0:
		return
	var min_side := minf(rect.size.x, rect.size.y)
	var center := rect.size * 0.5
	var radius := min_side * 0.42
	draw_circle(center, radius, Color(0.012, 0.012, 0.014, 0.94))
	match class_id:
		"sorcerer":
			_draw_spark(center, radius)
		"paladin":
			_draw_shield(center, radius)
		_:
			_draw_axe(center, radius)
	draw_arc(center, radius, 0.0, TAU, 40, accent_color, 2.0, true)


func _draw_axe(center: Vector2, radius: float) -> void:
	draw_line(center + Vector2(-radius * 0.42, radius * 0.58), center + Vector2(radius * 0.36, -radius * 0.55), accent_color, 4.0)
	var blade := PackedVector2Array([
		center + Vector2(radius * 0.08, -radius * 0.74),
		center + Vector2(radius * 0.62, -radius * 0.58),
		center + Vector2(radius * 0.48, -radius * 0.04),
		center + Vector2(radius * 0.08, -radius * 0.22),
	])
	draw_colored_polygon(blade, fill_color)
	draw_polyline(blade, accent_color, 2.0, true)


func _draw_spark(center: Vector2, radius: float) -> void:
	var points := PackedVector2Array()
	for i in range(8):
		var r := radius * (0.9 if i % 2 == 0 else 0.32)
		var angle := (-PI * 0.5) + (TAU * float(i) / 8.0)
		points.append(center + Vector2(cos(angle), sin(angle)) * r)
	draw_colored_polygon(points, fill_color)
	draw_polyline(points, accent_color, 2.0, true)
	draw_circle(center, radius * 0.18, accent_color)


func _draw_shield(center: Vector2, radius: float) -> void:
	var shield := PackedVector2Array([
		center + Vector2(-radius * 0.58, -radius * 0.48),
		center + Vector2(radius * 0.58, -radius * 0.48),
		center + Vector2(radius * 0.46, radius * 0.20),
		center + Vector2(0, radius * 0.76),
		center + Vector2(-radius * 0.46, radius * 0.20),
	])
	draw_colored_polygon(shield, fill_color)
	draw_polyline(shield, accent_color, 2.0, true)
	draw_line(center + Vector2(0, -radius * 0.34), center + Vector2(0, radius * 0.48), accent_color, 2.0)
	draw_line(center + Vector2(-radius * 0.30, -radius * 0.04), center + Vector2(radius * 0.30, -radius * 0.04), accent_color, 2.0)
