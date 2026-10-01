class_name RarityCuePresenter
extends RefCounted

const Loader := preload("res://scripts/rarity_cue_loader.gd")
static var capture_baseline: bool = false


static func draw_slot(canvas: Control, rect: Rect2, item: Dictionary) -> void:
	if capture_baseline:
		return
	var cue := Loader.cue_for_item(item)
	if cue.is_empty():
		return
	var badge := slot_badge_rect(rect)
	var config: Dictionary = Loader.catalog().get("slot", {})
	var side := badge.size.x
	var background := Color(str(config.get("background", "#101010")))
	var foreground := Color(str(config.get("foreground", "#ffffff")))
	var stroke := float(config.get("stroke_px", 1.5))
	canvas.draw_rect(badge, background, true)
	var center := badge.get_center()
	var radius := side * 0.43
	match str(cue.get("shape", "")):
		"square":
			canvas.draw_rect(Rect2(center - Vector2.ONE * radius, Vector2.ONE * radius * 2.0), foreground, false, stroke)
		"circle":
			canvas.draw_arc(center, radius, 0.0, TAU, 20, foreground, stroke, true)
		"triangle":
			_draw_outline(canvas, PackedVector2Array([center + Vector2(0, -radius), center + Vector2(radius, radius), center + Vector2(-radius, radius)]), foreground, stroke)
		"diamond":
			_draw_outline(canvas, PackedVector2Array([center + Vector2(0, -radius), center + Vector2(radius, 0), center + Vector2(0, radius), center + Vector2(-radius, 0)]), foreground, stroke)
		"star":
			var points := PackedVector2Array()
			for i in range(10):
				var angle := -PI * 0.5 + float(i) * PI / 5.0
				var r := radius if i % 2 == 0 else radius * 0.48
				points.append(center + Vector2(cos(angle), sin(angle)) * r)
			_draw_outline(canvas, points, foreground, stroke)
	var font := canvas.get_theme_default_font()
	var short := str(cue.get("short", ""))
	var font_size := int(config.get("font_size_px", 10))
	var text_size := font.get_string_size(short, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size)
	canvas.draw_string(font, center + Vector2(-text_size.x * 0.5, text_size.y * 0.34), short, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, foreground)


static func slot_badge_rect(rect: Rect2) -> Rect2:
	var config: Dictionary = Loader.catalog().get("slot", {})
	var side := maxf(float(config.get("minimum_size_px", 15)), minf(rect.size.x, rect.size.y) * float(config.get("badge_fraction", 0.28)))
	var margin := float(config.get("margin_px", 2))
	return Rect2(rect.position + Vector2(margin, margin), Vector2(side, side))


static func _draw_outline(canvas: Control, points: PackedVector2Array, color: Color, stroke: float) -> void:
	var closed := points.duplicate()
	closed.append(points[0])
	canvas.draw_polyline(closed, color, stroke, true)


static func add_world_marker(parent: Node3D, item: Dictionary, scale: float) -> void:
	if capture_baseline:
		return
	var cue := Loader.cue_for_item(item)
	if cue.is_empty():
		return
	var config: Dictionary = Loader.catalog().get("world", {})
	var marker := Label3D.new()
	marker.name = "RarityCue"
	marker.text = str(cue.get("world_symbol", ""))
	marker.position = Vector3(float(config.get("offset_x", 0.44)) * scale, float(config.get("height", 0.58)) * maxf(scale, 0.8), 0.0)
	marker.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	marker.no_depth_test = true
	marker.fixed_size = true
	marker.pixel_size = float(config.get("pixel_size", 0.0022))
	marker.font_size = int(config.get("font_size", 32))
	marker.modulate = Color(str(config.get("foreground", "#ffffff")))
	marker.outline_modulate = Color(str(config.get("outline", "#000000")))
	marker.outline_size = int(config.get("outline_size", 8))
	parent.add_child(marker)
