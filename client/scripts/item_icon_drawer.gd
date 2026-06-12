class_name ItemIconDrawer
extends RefCounted


static func draw(canvas: Control, rect: Rect2, icon: Dictionary, fallback_label: String = "", dimmed: bool = false, label_y_factor: float = 0.36, font_size: int = 16) -> void:
	var shape := str(icon.get("shape", "box"))
	var color := Color(str(icon.get("color", "#d8d0bd")))
	var accent := Color(str(icon.get("accent", "#6b5420")))
	if dimmed:
		color = color.darkened(0.35)
		accent = accent.darkened(0.35)
	var center := rect.get_center()
	var min_side := minf(rect.size.x, rect.size.y)

	match shape:
		"blade":
			_draw_blade(canvas, center, min_side, color, accent)
		"bow":
			_draw_bow(canvas, center, min_side, color, accent)
		"shield":
			_draw_shield(canvas, center, min_side, color, accent)
		"helm":
			_draw_helm(canvas, center, min_side, color, accent)
		"chest":
			_draw_chest(canvas, center, min_side, color, accent)
		"gloves":
			_draw_gloves(canvas, center, min_side, color, accent)
		"belt":
			_draw_belt(canvas, center, min_side, color, accent)
		"boots":
			_draw_boots(canvas, center, min_side, color, accent)
		"ring":
			_draw_ring(canvas, center, min_side, color, accent)
		"amulet":
			_draw_amulet(canvas, center, min_side, color, accent)
		"coin":
			_draw_coin(canvas, center, min_side, color, accent)
		"leaf":
			_draw_leaf(canvas, center, min_side, color, accent)
		"potion":
			_draw_potion(canvas, center, min_side, color, accent)
		_:
			canvas.draw_rect(Rect2(center - Vector2(min_side * 0.20, min_side * 0.20), Vector2(min_side * 0.40, min_side * 0.40)), color, true)

	var label := str(icon.get("label", fallback_label))
	if label == "":
		return
	var font := canvas.get_theme_default_font()
	var text_size := font.get_string_size(label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size)
	canvas.draw_string(font, center + Vector2(-text_size.x * 0.5, min_side * label_y_factor), label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, Color("#f4ead8"))


static func _draw_blade(canvas: Control, center: Vector2, min_side: float, color: Color, accent: Color) -> void:
	var a := center + Vector2(-min_side * 0.23, min_side * 0.22)
	var b := center + Vector2(min_side * 0.24, -min_side * 0.24)
	canvas.draw_line(a, b, color, maxf(4.0, min_side * 0.09), true)
	canvas.draw_line(a + Vector2(-min_side * 0.07, min_side * 0.07), a + Vector2(min_side * 0.10, -min_side * 0.10), accent, maxf(3.0, min_side * 0.065), true)


static func _draw_bow(canvas: Control, center: Vector2, min_side: float, color: Color, accent: Color) -> void:
	canvas.draw_arc(center, min_side * 0.29, -1.25, 1.25, 22, color, maxf(4.0, min_side * 0.07), true)
	canvas.draw_line(center + Vector2(min_side * 0.18, -min_side * 0.28), center + Vector2(min_side * 0.18, min_side * 0.28), accent, maxf(2.0, min_side * 0.035), true)


static func _draw_shield(canvas: Control, center: Vector2, min_side: float, color: Color, accent: Color) -> void:
	var r := min_side * 0.34
	var pts := PackedVector2Array([
		center + Vector2(-r * 0.72, -r * 0.72),
		center + Vector2(r * 0.72, -r * 0.72),
		center + Vector2(r * 0.58, r * 0.28),
		center + Vector2(0, r * 0.92),
		center + Vector2(-r * 0.58, r * 0.28),
	])
	canvas.draw_colored_polygon(pts, color)
	canvas.draw_polyline(pts, accent, maxf(2.0, min_side * 0.035), true)


static func _draw_helm(canvas: Control, center: Vector2, min_side: float, color: Color, accent: Color) -> void:
	var r := min_side * 0.27
	canvas.draw_arc(center + Vector2(0, min_side * 0.04), r, PI, TAU, 24, color, maxf(8.0, min_side * 0.12), true)
	canvas.draw_rect(Rect2(center + Vector2(-r, min_side * 0.02), Vector2(r * 2.0, min_side * 0.12)), accent, true)


static func _draw_chest(canvas: Control, center: Vector2, min_side: float, color: Color, accent: Color) -> void:
	var body := Rect2(center + Vector2(-min_side * 0.20, -min_side * 0.22), Vector2(min_side * 0.40, min_side * 0.44))
	canvas.draw_rect(body, color, true)
	canvas.draw_line(body.position + Vector2(body.size.x * 0.5, 0), body.position + Vector2(body.size.x * 0.5, body.size.y), accent, maxf(2.0, min_side * 0.035), true)
	canvas.draw_rect(Rect2(center + Vector2(-min_side * 0.34, -min_side * 0.18), Vector2(min_side * 0.14, min_side * 0.16)), accent, true)
	canvas.draw_rect(Rect2(center + Vector2(min_side * 0.20, -min_side * 0.18), Vector2(min_side * 0.14, min_side * 0.16)), accent, true)


static func _draw_gloves(canvas: Control, center: Vector2, min_side: float, color: Color, accent: Color) -> void:
	canvas.draw_rect(Rect2(center + Vector2(-min_side * 0.32, -min_side * 0.10), Vector2(min_side * 0.22, min_side * 0.26)), color, true)
	canvas.draw_rect(Rect2(center + Vector2(min_side * 0.10, -min_side * 0.10), Vector2(min_side * 0.22, min_side * 0.26)), color, true)
	canvas.draw_line(center + Vector2(-min_side * 0.30, -min_side * 0.15), center + Vector2(-min_side * 0.10, -min_side * 0.03), accent, maxf(2.0, min_side * 0.035), true)
	canvas.draw_line(center + Vector2(min_side * 0.30, -min_side * 0.15), center + Vector2(min_side * 0.10, -min_side * 0.03), accent, maxf(2.0, min_side * 0.035), true)


static func _draw_belt(canvas: Control, center: Vector2, min_side: float, color: Color, accent: Color) -> void:
	canvas.draw_rect(Rect2(center + Vector2(-min_side * 0.34, -min_side * 0.08), Vector2(min_side * 0.68, min_side * 0.16)), color, true)
	canvas.draw_rect(Rect2(center + Vector2(-min_side * 0.10, -min_side * 0.13), Vector2(min_side * 0.20, min_side * 0.26)), accent, false, maxf(2.0, min_side * 0.035))


static func _draw_boots(canvas: Control, center: Vector2, min_side: float, color: Color, accent: Color) -> void:
	canvas.draw_rect(Rect2(center + Vector2(-min_side * 0.30, -min_side * 0.20), Vector2(min_side * 0.18, min_side * 0.40)), color, true)
	canvas.draw_rect(Rect2(center + Vector2(min_side * 0.12, -min_side * 0.20), Vector2(min_side * 0.18, min_side * 0.40)), color, true)
	canvas.draw_rect(Rect2(center + Vector2(-min_side * 0.34, min_side * 0.14), Vector2(min_side * 0.24, min_side * 0.08)), accent, true)
	canvas.draw_rect(Rect2(center + Vector2(min_side * 0.10, min_side * 0.14), Vector2(min_side * 0.24, min_side * 0.08)), accent, true)


static func _draw_ring(canvas: Control, center: Vector2, min_side: float, color: Color, accent: Color) -> void:
	canvas.draw_arc(center, min_side * 0.22, 0.0, TAU, 32, color, maxf(5.0, min_side * 0.08), true)
	canvas.draw_circle(center + Vector2(0, -min_side * 0.22), min_side * 0.07, accent)


static func _draw_amulet(canvas: Control, center: Vector2, min_side: float, color: Color, accent: Color) -> void:
	canvas.draw_arc(center + Vector2(0, -min_side * 0.05), min_side * 0.25, 0.15, PI - 0.15, 24, color, maxf(3.0, min_side * 0.045), true)
	var gem := PackedVector2Array([
		center + Vector2(0, min_side * 0.02),
		center + Vector2(min_side * 0.13, min_side * 0.16),
		center + Vector2(0, min_side * 0.31),
		center + Vector2(-min_side * 0.13, min_side * 0.16),
	])
	canvas.draw_colored_polygon(gem, accent)


static func _draw_coin(canvas: Control, center: Vector2, min_side: float, color: Color, accent: Color) -> void:
	canvas.draw_circle(center, min_side * 0.24, color)
	canvas.draw_arc(center, min_side * 0.17, 0.0, TAU, 22, accent, maxf(2.0, min_side * 0.035), true)


static func _draw_leaf(canvas: Control, center: Vector2, min_side: float, color: Color, accent: Color) -> void:
	var pts := PackedVector2Array([
		center + Vector2(0, -min_side * 0.30),
		center + Vector2(min_side * 0.24, -min_side * 0.02),
		center + Vector2(0, min_side * 0.28),
		center + Vector2(-min_side * 0.24, -min_side * 0.02),
	])
	canvas.draw_colored_polygon(pts, color)
	canvas.draw_line(center + Vector2(0, -min_side * 0.22), center + Vector2(0, min_side * 0.22), accent, maxf(2.0, min_side * 0.035), true)


static func _draw_potion(canvas: Control, center: Vector2, min_side: float, color: Color, accent: Color) -> void:
	canvas.draw_rect(Rect2(center + Vector2(-min_side * 0.13, -min_side * 0.05), Vector2(min_side * 0.26, min_side * 0.28)), color, true)
	canvas.draw_rect(Rect2(center + Vector2(-min_side * 0.08, -min_side * 0.22), Vector2(min_side * 0.16, min_side * 0.16)), accent, true)
