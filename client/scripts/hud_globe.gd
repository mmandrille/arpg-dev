class_name HudGlobe
extends Control
## Round liquid-style resource globe (v517). Drawn with `_draw` only; redraws on change, never per frame.
## Fill height = ratio. The "cur / max" text is always drawn so state never relies on color alone.

const ARC_STEPS := 48

var fill_color: Color = Color("#c0392b"):
	set(value):
		if value == fill_color:
			return
		fill_color = value
		_request_redraw()
var _ratio: float = 1.0
var _label: String = ""
var _tween: Tween
var redraw_requests: int = 0


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE


func set_ratio(value: float) -> void:
	var next := clampf(value, 0.0, 1.0)
	if is_equal_approx(next, _ratio):
		return
	_ratio = next
	_request_redraw()


func set_label(text: String) -> void:
	if text == _label:
		return
	_label = text
	_request_redraw()


## Briefly show `flash` then ease back to `settle` (same timings the old bars used).
func flash(flash_color: Color, settle: Color) -> void:
	if _tween != null and _tween.is_valid():
		_tween.kill()
	fill_color = flash_color
	_tween = create_tween()
	_tween.tween_interval(0.05)
	_tween.tween_property(self, "fill_color", settle, 0.45)


func is_flashing() -> bool:
	return _tween != null and _tween.is_valid() and _tween.is_running()


func get_debug_state() -> Dictionary:
	return {"ratio": _ratio, "label": _label, "fill_color": fill_color, "size": size}


## Liquid body: the part of the circle at or below the level line. Empty for ratio <= 0.
static func liquid_polygon(center: Vector2, radius: float, ratio: float) -> PackedVector2Array:
	var points := PackedVector2Array()
	var r := clampf(ratio, 0.0, 1.0)
	if r <= 0.0 or radius <= 0.0:
		return points
	var level_offset := radius - 2.0 * radius * r # y of the surface relative to the center
	var theta := acos(clampf(level_offset / radius, -1.0, 1.0))
	for i in range(ARC_STEPS + 1):
		var phi := -theta + (2.0 * theta) * float(i) / float(ARC_STEPS)
		points.append(center + Vector2(sin(phi), cos(phi)) * radius)
	return points


func _request_redraw() -> void:
	redraw_requests += 1
	queue_redraw()


func _draw() -> void:
	var center := size * 0.5
	var radius := minf(size.x, size.y) * 0.5
	if radius <= 2.0:
		return
	var inner := radius - HudStyle.globe_ring_width()
	draw_circle(center, radius, HudStyle.globe_ring())
	draw_circle(center, inner, HudStyle.globe_trough())
	var liquid := liquid_polygon(center, inner - 1.0, _ratio)
	if liquid.size() >= 3:
		draw_colored_polygon(liquid, fill_color)
		if _ratio < 1.0:
			var level := center.y + (inner - 1.0) - 2.0 * (inner - 1.0) * _ratio
			var half := sqrt(maxf(0.0, pow(inner - 1.0, 2.0) - pow(level - center.y, 2.0)))
			draw_line(Vector2(center.x - half, level), Vector2(center.x + half, level), HudStyle.globe_surface(), 2.0)
	draw_arc(center, inner * 0.78, deg_to_rad(205.0), deg_to_rad(255.0), 12, HudStyle.globe_highlight(), 3.0, true)
	draw_arc(center, radius - 1.0, 0.0, TAU, 64, HudStyle.globe_ring_edge(), 1.5, true)
	if _label != "":
		var font := ThemeDB.fallback_font
		var font_size := clampi(int(radius * 0.30), 12, 24)
		var text_size := font.get_string_size(_label, HORIZONTAL_ALIGNMENT_CENTER, -1, font_size)
		var origin := Vector2(center.x - text_size.x * 0.5, center.y + font_size * 0.35)
		draw_string_outline(font, origin, _label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, 4, HudStyle.globe_text_shadow())
		draw_string(font, origin, _label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, HudStyle.globe_text())
