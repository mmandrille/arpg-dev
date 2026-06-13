class_name SkillIcon
extends Control

var skill_id: String = ""
var label_text: String = ""
var shape: String = "bolt"
var fill_color := Color("#62b7ff")
var accent_color := Color("#e8f7ff")


func configure(next_skill_id: String, presentation: Dictionary) -> void:
	skill_id = next_skill_id
	var icon: Dictionary = presentation.get("icon", {})
	label_text = str(icon.get("label", next_skill_id.substr(0, 1).to_upper()))
	shape = str(icon.get("shape", "bolt"))
	fill_color = Color(str(icon.get("color", "#62b7ff")))
	accent_color = Color(str(icon.get("accent", "#e8f7ff")))
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
	draw_circle(center, radius, Color(0.015, 0.014, 0.012, 0.92))
	match shape:
		"burst":
			_draw_burst(center, radius)
		"dash":
			_draw_dash(center, radius)
		"heart":
			_draw_heart(center, radius)
		"ice_spike":
			_draw_ice_spike(center, radius)
		"orb_projectile":
			_draw_orb_projectile(center, radius)
		"poison_dagger":
			_draw_poison_dagger(center, radius)
		"shield":
			_draw_shield(center, radius)
		"slash":
			_draw_slash(center, radius)
		_:
			_draw_bolt(center, radius)
	draw_arc(center, radius, 0.0, TAU, 40, accent_color, 2.0, true)


func _draw_bolt(center: Vector2, radius: float) -> void:
	var points := PackedVector2Array([
		center + Vector2(-radius * 0.10, -radius * 0.82),
		center + Vector2(radius * 0.34, -radius * 0.18),
		center + Vector2(radius * 0.08, -radius * 0.18),
		center + Vector2(radius * 0.22, radius * 0.82),
		center + Vector2(-radius * 0.38, radius * 0.04),
		center + Vector2(-radius * 0.10, radius * 0.04),
	])
	draw_colored_polygon(points, fill_color)
	draw_polyline(points, accent_color, 2.0, true)


func _draw_burst(center: Vector2, radius: float) -> void:
	var points := PackedVector2Array()
	for i in range(12):
		var r := radius * (0.95 if i % 2 == 0 else 0.48)
		var angle := (-PI * 0.5) + (TAU * float(i) / 12.0)
		points.append(center + Vector2(cos(angle), sin(angle)) * r)
	draw_colored_polygon(points, fill_color)
	draw_polyline(points, accent_color, 2.0, true)
	draw_circle(center, radius * 0.34, accent_color)


func _draw_dash(center: Vector2, radius: float) -> void:
	var body := PackedVector2Array([
		center + Vector2(radius * 0.46, -radius * 0.70),
		center + Vector2(radius * 0.78, -radius * 0.22),
		center + Vector2(radius * 0.28, -radius * 0.12),
		center + Vector2(radius * 0.54, radius * 0.62),
		center + Vector2(-radius * 0.02, radius * 0.20),
		center + Vector2(-radius * 0.18, -radius * 0.34),
	])
	draw_colored_polygon(body, fill_color)
	draw_polyline(body, accent_color, 2.0, true)
	for i in range(3):
		var y := -radius * 0.42 + float(i) * radius * 0.34
		draw_line(
			center + Vector2(-radius * 0.78, y),
			center + Vector2(-radius * 0.28, y - radius * 0.10),
			accent_color,
			2.0,
			true
		)
	draw_circle(center + Vector2(radius * 0.12, -radius * 0.42), radius * 0.11, accent_color)


func _draw_heart(center: Vector2, radius: float) -> void:
	var points := PackedVector2Array()
	for i in range(34):
		var t := TAU * float(i) / 34.0
		var x := 16.0 * pow(sin(t), 3.0)
		var y := -(13.0 * cos(t) - 5.0 * cos(2.0 * t) - 2.0 * cos(3.0 * t) - cos(4.0 * t))
		points.append(center + Vector2(x, y) * (radius / 18.0))
	draw_colored_polygon(points, fill_color)
	draw_polyline(points, accent_color, 2.0, true)


func _draw_ice_spike(center: Vector2, radius: float) -> void:
	var spike := PackedVector2Array([
		center + Vector2(radius * 0.10, -radius * 0.92),
		center + Vector2(radius * 0.48, radius * 0.62),
		center + Vector2(-radius * 0.08, radius * 0.82),
		center + Vector2(-radius * 0.36, radius * 0.18),
	])
	draw_colored_polygon(spike, fill_color)
	draw_polyline(spike, accent_color, 2.0, true)
	draw_line(center + Vector2(radius * 0.06, -radius * 0.68), center + Vector2(radius * 0.06, radius * 0.54), accent_color, 2.0, true)
	draw_line(center + Vector2(-radius * 0.12, -radius * 0.12), center + Vector2(radius * 0.34, radius * 0.44), accent_color, 1.6, true)


func _draw_orb_projectile(center: Vector2, radius: float) -> void:
	for i in range(3):
		var offset := radius * (0.46 + float(i) * 0.18)
		draw_line(
			center + Vector2(-offset, -radius * (0.20 - float(i) * 0.13)),
			center + Vector2(-offset - radius * 0.34, -radius * (0.30 - float(i) * 0.10)),
			accent_color,
			1.8,
			true
		)
	draw_circle(center + Vector2(radius * 0.12, 0.0), radius * 0.46, fill_color)
	draw_arc(center + Vector2(radius * 0.12, 0.0), radius * 0.46, 0.0, TAU, 32, accent_color, 2.0, true)
	draw_circle(center + Vector2(radius * 0.28, -radius * 0.18), radius * 0.14, accent_color)


func _draw_poison_dagger(center: Vector2, radius: float) -> void:
	var blade := PackedVector2Array([
		center + Vector2(radius * 0.52, -radius * 0.78),
		center + Vector2(radius * 0.22, radius * 0.18),
		center + Vector2(-radius * 0.02, radius * 0.02),
	])
	draw_colored_polygon(blade, fill_color)
	draw_polyline(blade, accent_color, 2.0, true)
	draw_line(center + Vector2(radius * 0.08, radius * 0.12), center + Vector2(-radius * 0.48, radius * 0.66), accent_color, 4.0, true)
	draw_line(center + Vector2(-radius * 0.26, radius * 0.22), center + Vector2(radius * 0.10, radius * 0.58), accent_color, 2.0, true)
	draw_circle(center + Vector2(-radius * 0.42, -radius * 0.18), radius * 0.13, fill_color)
	draw_circle(center + Vector2(-radius * 0.62, radius * 0.10), radius * 0.09, accent_color)
	draw_circle(center + Vector2(-radius * 0.28, -radius * 0.38), radius * 0.08, accent_color)


func _draw_shield(center: Vector2, radius: float) -> void:
	var points := PackedVector2Array([
		center + Vector2(0.0, -radius * 0.86),
		center + Vector2(radius * 0.62, -radius * 0.46),
		center + Vector2(radius * 0.48, radius * 0.34),
		center + Vector2(0.0, radius * 0.84),
		center + Vector2(-radius * 0.48, radius * 0.34),
		center + Vector2(-radius * 0.62, -radius * 0.46),
	])
	draw_colored_polygon(points, fill_color)
	draw_polyline(points, accent_color, 2.0, true)
	draw_line(center + Vector2(0.0, -radius * 0.64), center + Vector2(0.0, radius * 0.52), accent_color, 2.0, true)
	draw_arc(center, radius * 0.34, PI * 0.12, PI * 0.88, 16, accent_color, 2.0, true)


func _draw_slash(center: Vector2, radius: float) -> void:
	var upper := center + Vector2(radius * 0.58, -radius * 0.68)
	var lower := center + Vector2(-radius * 0.58, radius * 0.68)
	var blade := PackedVector2Array([
		upper + Vector2(radius * 0.12, radius * 0.02),
		center + Vector2(radius * 0.12, radius * 0.12),
		lower + Vector2(-radius * 0.06, radius * 0.14),
		center + Vector2(-radius * 0.16, -radius * 0.06),
		upper + Vector2(radius * 0.02, -radius * 0.18),
	])
	draw_colored_polygon(blade, fill_color)
	draw_polyline(blade, accent_color, 2.0, true)
	draw_line(center + Vector2(-radius * 0.42, -radius * 0.44), center + Vector2(radius * 0.44, radius * 0.42), accent_color, 2.0, true)
