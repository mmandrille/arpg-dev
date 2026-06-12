class_name ClassIcon
extends Control

static var _presentations: Dictionary = {}
static var _loaded_presentations: bool = false

var class_id: String = "barbarian"
var shape: String = "axe"
var fill_color := Color("#c85f3d")
var accent_color := Color("#ffd9a8")


func configure(next_class_id: String) -> void:
	class_id = next_class_id
	_ensure_presentations()
	var presentation: Dictionary = _presentations.get(class_id, {})
	var icon: Dictionary = presentation.get("icon", {})
	shape = str(icon.get("shape", _fallback_shape(class_id)))
	fill_color = Color(str(icon.get("color", _fallback_color(class_id))))
	accent_color = Color(str(icon.get("accent", _fallback_accent(class_id))))
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
	match shape:
		"spark":
			_draw_spark(center, radius)
		"shield":
			_draw_shield(center, radius)
		_:
			_draw_axe(center, radius)
	draw_arc(center, radius, 0.0, TAU, 40, accent_color, 2.0, true)


static func _ensure_presentations() -> void:
	if _loaded_presentations:
		return
	_loaded_presentations = true
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/class_presentations.v0.json")
	if not FileAccess.file_exists(path):
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		_presentations = parsed.get("classes", {})


static func _fallback_shape(next_class_id: String) -> String:
	match next_class_id:
		"sorcerer":
			return "spark"
		"paladin":
			return "shield"
		_:
			return "axe"


static func _fallback_color(next_class_id: String) -> String:
	match next_class_id:
		"sorcerer":
			return "#5e8cff"
		"paladin":
			return "#d9b44a"
		_:
			return "#c85f3d"


static func _fallback_accent(next_class_id: String) -> String:
	match next_class_id:
		"sorcerer":
			return "#dce7ff"
		"paladin":
			return "#fff3ba"
		_:
			return "#ffd9a8"


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
