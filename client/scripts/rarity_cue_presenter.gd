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
	var background := Color(str(config.get("background", "#101010")))
	var foreground := Color(str(config.get("foreground", "#ffffff")))
	canvas.draw_rect(badge, background, true)
	var center := badge.get_center()
	var radius := badge.size.x * 0.42
	var points := _shape_points(str(cue.get("shape", "")), center, radius)
	if points.size() >= 3:
		canvas.draw_colored_polygon(points, foreground)


static func slot_badge_rect(rect: Rect2) -> Rect2:
	var config: Dictionary = Loader.catalog().get("slot", {})
	var side := maxf(float(config.get("minimum_size_px", 15)), minf(rect.size.x, rect.size.y) * float(config.get("badge_fraction", 0.28)))
	var margin := float(config.get("margin_px", 2))
	return Rect2(rect.position + Vector2(margin, margin), Vector2(side, side))


static func _shape_points(shape: String, center: Vector2, radius: float) -> PackedVector2Array:
	var points := PackedVector2Array()
	match shape:
		"square":
			points = PackedVector2Array([
				center + Vector2(-radius, -radius), center + Vector2(radius, -radius),
				center + Vector2(radius, radius), center + Vector2(-radius, radius),
			])
		"circle":
			for i in range(24):
				var angle := -PI * 0.5 + float(i) * TAU / 24.0
				points.append(center + Vector2(cos(angle), sin(angle)) * radius)
		"triangle":
			points = PackedVector2Array([center + Vector2(0, -radius), center + Vector2(radius, radius), center + Vector2(-radius, radius)])
		"diamond":
			points = PackedVector2Array([center + Vector2(0, -radius), center + Vector2(radius, 0), center + Vector2(0, radius), center + Vector2(-radius, 0)])
		"star":
			for i in range(10):
				var angle := -PI * 0.5 + float(i) * PI / 5.0
				var point_radius := radius if i % 2 == 0 else radius * 0.46
				points.append(center + Vector2(cos(angle), sin(angle)) * point_radius)
	return points
