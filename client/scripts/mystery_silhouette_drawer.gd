extends RefCounted

static func key(row: Dictionary) -> String:
	var slot := str(row.get("slot", "")).strip_edges().to_lower()
	match slot:
		"ring", "amulet", "head", "chest", "gloves", "boots", "main_hand", "off_hand":
			return slot
		"hands":
			return "gloves"
		"feet":
			return "boots"
	var category := str(row.get("category", "")).strip_edges().to_lower()
	return category if category != "" else "item"


static func label(row: Dictionary) -> String:
	return key(row).replace("_", " ").capitalize()


static func draw(slot: Control, center: Vector2, min_side: float, row: Dictionary, color: Color, accent: Color) -> void:
	var r := min_side * 0.26
	match key(row):
		"ring":
			slot.draw_arc(center + Vector2(0.0, -min_side * 0.05), r, 0.0, TAU, 32, color, 3.0, true)
			slot.draw_circle(center + Vector2(r * 0.45, -r * 0.55), r * 0.18, accent)
		"amulet":
			slot.draw_line(center + Vector2(-r * 0.75, -r * 0.65), center + Vector2(0.0, r * 0.15), color, 2.5)
			slot.draw_line(center + Vector2(r * 0.75, -r * 0.65), center + Vector2(0.0, r * 0.15), color, 2.5)
			slot.draw_circle(center + Vector2(0.0, r * 0.42), r * 0.28, accent)
		"off_hand":
			var shield := PackedVector2Array([center + Vector2(-r * 0.65, -r * 0.75), center + Vector2(r * 0.65, -r * 0.75), center + Vector2(r * 0.48, r * 0.28), center + Vector2(0.0, r * 0.85), center + Vector2(-r * 0.48, r * 0.28), center + Vector2(-r * 0.65, -r * 0.75)])
			slot.draw_polyline(shield, color, 2.5, true)
			slot.draw_line(center + Vector2(0.0, -r * 0.60), center + Vector2(0.0, r * 0.45), accent, 1.5)
		"head":
			slot.draw_arc(center + Vector2(0.0, r * 0.10), r * 0.74, PI, TAU, 24, color, 3.0, true)
			slot.draw_line(center + Vector2(-r * 0.85, r * 0.05), center + Vector2(r * 0.85, r * 0.05), accent, 2.0)
		"chest", "equipment":
			slot.draw_rect(Rect2(center - Vector2(r * 0.55, r * 0.70), Vector2(r * 1.10, r * 1.35)), color, false, 2.5)
			slot.draw_line(center + Vector2(-r * 0.25, -r * 0.60), center + Vector2(0.0, -r * 0.30), accent, 2.0)
			slot.draw_line(center + Vector2(r * 0.25, -r * 0.60), center + Vector2(0.0, -r * 0.30), accent, 2.0)
		"gloves":
			slot.draw_line(center + Vector2(-r * 0.58, r * 0.45), center + Vector2(-r * 0.58, -r * 0.45), color, 4.0)
			slot.draw_line(center + Vector2(r * 0.58, r * 0.45), center + Vector2(r * 0.58, -r * 0.45), color, 4.0)
			slot.draw_line(center + Vector2(-r * 0.82, -r * 0.05), center + Vector2(r * 0.82, -r * 0.05), accent, 2.0)
		"boots":
			slot.draw_line(center + Vector2(-r * 0.70, -r * 0.25), center + Vector2(-r * 0.35, r * 0.55), color, 4.0)
			slot.draw_line(center + Vector2(r * 0.40, -r * 0.25), center + Vector2(r * 0.75, r * 0.55), color, 4.0)
			slot.draw_line(center + Vector2(-r * 0.48, r * 0.55), center + Vector2(r * 0.80, r * 0.55), accent, 2.0)
		"main_hand":
			slot.draw_line(center + Vector2(-r * 0.58, r * 0.55), center + Vector2(r * 0.55, -r * 0.58), color, 3.0)
			slot.draw_line(center + Vector2(-r * 0.72, r * 0.18), center + Vector2(-r * 0.22, r * 0.68), accent, 2.0)
		_:
			slot.draw_arc(center + Vector2(0.0, -min_side * 0.02), min_side * 0.24, -0.9, 4.0, 24, color, 3.0, true)
			slot.draw_circle(center + Vector2(0.0, min_side * 0.20), min_side * 0.035, accent)
