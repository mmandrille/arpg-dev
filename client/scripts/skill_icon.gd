class_name SkillIcon
extends Control

const SkillRankIntensityScript := preload("res://scripts/skill_rank_intensity.gd")

var skill_id: String = ""
var label_text: String = ""
var shape: String = "bolt"
var fill_color := Color("#62b7ff")
var accent_color := Color("#e8f7ff")


var rank: int = 0
var accent_width: float = 2.0
var glow_ring_count: int = 0


func configure(next_skill_id: String, presentation: Dictionary, next_rank: int = 0) -> void:
	skill_id = next_skill_id
	rank = maxi(0, next_rank)
	var icon: Dictionary = presentation.get("icon", {})
	label_text = str(icon.get("label", next_skill_id.substr(0, 1).to_upper()))
	shape = str(icon.get("shape", "bolt"))
	fill_color = Color(str(icon.get("color", "#62b7ff")))
	accent_color = Color(str(icon.get("accent", "#e8f7ff")))
	var intensity: Dictionary = SkillRankIntensityScript.resolve(presentation, rank)
	accent_width = float(intensity.get("accent_width", 2.0))
	glow_ring_count = int(intensity.get("glow_ring_count", 0))
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
	if _draw_passive_shape(center, radius):
		pass
	else:
		match shape:
			"arrow":
				_draw_arrow(center, radius)
			"burst":
				_draw_burst(center, radius)
			"dash":
				_draw_dash(center, radius)
			"heart":
				_draw_heart(center, radius)
			"ice_spike":
				_draw_ice_spike(center, radius)
			"flame":
				_draw_flame(center, radius)
			"leap":
				_draw_leap(center, radius)
			"orb_projectile":
				_draw_orb_projectile(center, radius)
			"pin":
				_draw_pin(center, radius)
			"poison_dagger":
				_draw_poison_dagger(center, radius)
			"quake":
				_draw_quake(center, radius)
			"shield":
				_draw_shield(center, radius)
			"slash":
				_draw_slash(center, radius)
			"volley":
				_draw_volley(center, radius)
			_:
				_draw_bolt(center, radius)
	for i in range(glow_ring_count):
		var ring_radius := radius * (0.58 - float(i) * 0.08)
		var ring_alpha := clampf(0.22 - float(i) * 0.04, 0.06, 0.22)
		draw_arc(center, ring_radius, 0.0, TAU, 32, Color(accent_color.r, accent_color.g, accent_color.b, ring_alpha), accent_width * 0.7, true)
	draw_arc(center, radius, 0.0, TAU, 40, accent_color, accent_width, true)


func _draw_passive_shape(center: Vector2, radius: float) -> bool:
	match shape:
		"passive_arcane_focus":
			_draw_orb_projectile(center, radius)
			draw_arc(center, radius * 0.72, PI * 0.12, PI * 1.22, 28, accent_color, 1.8, true)
		"passive_mana_weaving":
			for i in range(3):
				var y := -radius * 0.42 + float(i) * radius * 0.36
				_draw_wave(center + Vector2(0.0, y), radius * 0.72)
		"passive_spell_dynamo":
			_draw_burst(center, radius * 0.86)
			draw_circle(center, radius * 0.26, Color(0.015, 0.014, 0.012, 0.92))
			draw_arc(center, radius * 0.26, 0.0, TAU, 24, accent_color, 1.8, true)
		"passive_iron_hide":
			_draw_shield(center, radius)
			for y in [-0.28, 0.02, 0.32]:
				draw_line(center + Vector2(-radius * 0.34, radius * y), center + Vector2(radius * 0.34, radius * y), accent_color, 1.8, true)
		"passive_battle_tempo":
			for i in range(3):
				var x := -radius * 0.44 + float(i) * radius * 0.36
				_draw_chevron(center + Vector2(x, 0.0), radius * 0.26)
			draw_line(center + Vector2(-radius * 0.68, radius * 0.58), center + Vector2(radius * 0.68, radius * 0.58), accent_color, 2.0, true)
		"passive_crushing_force":
			_draw_quake(center, radius)
			draw_rect(Rect2(center + Vector2(-radius * 0.42, -radius * 0.60), Vector2(radius * 0.84, radius * 0.32)), fill_color)
			draw_rect(Rect2(center + Vector2(-radius * 0.28, -radius * 0.30), Vector2(radius * 0.56, radius * 0.34)), fill_color)
		"passive_vigilant_guard":
			_draw_shield(center, radius)
			_draw_eye(center + Vector2(0.0, -radius * 0.12), radius * 0.44)
		"passive_faithful_bulwark":
			for i in range(3):
				draw_rect(Rect2(center + Vector2(-radius * 0.66 + float(i) * radius * 0.44, -radius * 0.62), Vector2(radius * 0.34, radius * 0.40)), fill_color)
			draw_rect(Rect2(center + Vector2(-radius * 0.70, -radius * 0.16), Vector2(radius * 1.40, radius * 0.72)), fill_color)
			draw_line(center + Vector2(-radius * 0.70, -radius * 0.16), center + Vector2(radius * 0.70, -radius * 0.16), accent_color, 2.0, true)
		"passive_consecrated_vitality":
			_draw_heart(center, radius)
			draw_line(center + Vector2(0.0, -radius * 0.34), center + Vector2(0.0, radius * 0.34), accent_color, 3.0, true)
			draw_line(center + Vector2(-radius * 0.28, 0.0), center + Vector2(radius * 0.28, 0.0), accent_color, 3.0, true)
		"passive_quick_hands":
			_draw_dash(center, radius)
			for i in range(3):
				draw_line(center + Vector2(-radius * 0.80, -radius * 0.44 + float(i) * radius * 0.34), center + Vector2(-radius * 0.38, -radius * 0.54 + float(i) * radius * 0.34), accent_color, 2.2, true)
		"passive_killer_instinct":
			_draw_target(center, radius * 0.72)
			_draw_slash(center, radius * 0.92)
		"passive_evasive_footwork":
			_draw_footprint(center + Vector2(-radius * 0.24, radius * 0.18), radius * 0.34)
			_draw_footprint(center + Vector2(radius * 0.26, -radius * 0.20), radius * 0.34)
		"passive_trail_sense":
			_draw_pin(center, radius)
			draw_line(center + Vector2(0.0, -radius * 0.72), center + Vector2(0.0, radius * 0.72), accent_color, 1.6, true)
			draw_line(center + Vector2(-radius * 0.72, 0.0), center + Vector2(radius * 0.72, 0.0), accent_color, 1.6, true)
		"passive_precision_draw":
			_draw_arrow(center, radius)
			draw_arc(center + Vector2(-radius * 0.20, radius * 0.02), radius * 0.56, -PI * 0.45, PI * 0.45, 20, accent_color, 2.0, true)
		"passive_deadeye":
			_draw_target(center, radius * 0.76)
			draw_line(center + Vector2(-radius * 0.82, radius * 0.58), center + Vector2(radius * 0.62, -radius * 0.48), fill_color, radius * 0.12, true)
			draw_line(center + Vector2(-radius * 0.82, radius * 0.58), center + Vector2(radius * 0.62, -radius * 0.48), accent_color, radius * 0.04, true)
		_:
			return false
	return true


func _draw_wave(center: Vector2, width: float) -> void:
	var points := PackedVector2Array()
	for i in range(9):
		var t := float(i) / 8.0
		var x := -width * 0.5 + width * t
		var y := sin(t * TAU) * width * 0.12
		points.append(center + Vector2(x, y))
	draw_polyline(points, fill_color, 4.0, false)
	draw_polyline(points, accent_color, 1.6, false)


func _draw_chevron(center: Vector2, radius: float) -> void:
	var points := PackedVector2Array([
		center + Vector2(-radius, -radius),
		center + Vector2(radius * 0.20, 0.0),
		center + Vector2(-radius, radius),
	])
	draw_polyline(points, fill_color, 5.0, false)
	draw_polyline(points, accent_color, 1.8, false)


func _draw_eye(center: Vector2, radius: float) -> void:
	draw_arc(center, radius, PI * 0.05, PI * 0.95, 22, accent_color, 2.0, true)
	draw_arc(center, radius, PI * 1.05, PI * 1.95, 22, accent_color, 2.0, true)
	draw_circle(center, radius * 0.26, accent_color)


func _draw_target(center: Vector2, radius: float) -> void:
	for scale in [1.0, 0.58, 0.24]:
		draw_arc(center, radius * scale, 0.0, TAU, 32, accent_color if scale < 1.0 else fill_color, 2.0, true)
	draw_line(center + Vector2(-radius, 0.0), center + Vector2(radius, 0.0), accent_color, 1.6, true)
	draw_line(center + Vector2(0.0, -radius), center + Vector2(0.0, radius), accent_color, 1.6, true)


func _draw_footprint(center: Vector2, radius: float) -> void:
	var sole := PackedVector2Array()
	for i in range(18):
		var t := TAU * float(i) / 18.0
		sole.append(center + Vector2(cos(t) * radius * 0.30, sin(t) * radius * 0.46))
	draw_colored_polygon(sole, fill_color)
	draw_arc(center, radius * 0.46, 0.0, TAU, 18, accent_color, 1.8, true)
	for i in range(3):
		draw_circle(center + Vector2(-radius * 0.30 + float(i) * radius * 0.30, -radius * 0.62), radius * 0.10, accent_color)


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


func _draw_arrow(center: Vector2, radius: float) -> void:
	var shaft_start := center + Vector2(-radius * 0.62, radius * 0.34)
	var shaft_end := center + Vector2(radius * 0.34, -radius * 0.30)
	draw_line(shaft_start, shaft_end, fill_color, radius * 0.24, true)
	draw_line(shaft_start, shaft_end, accent_color, radius * 0.08, true)
	var head := PackedVector2Array([
		center + Vector2(radius * 0.78, -radius * 0.60),
		center + Vector2(radius * 0.30, -radius * 0.10),
		center + Vector2(radius * 0.10, -radius * 0.68),
	])
	draw_colored_polygon(head, fill_color)
	draw_polyline(head, accent_color, 2.0, true)
	for side in [-1.0, 1.0]:
		draw_line(
			center + Vector2(-radius * 0.46, radius * 0.24 + radius * 0.12 * side),
			center + Vector2(-radius * 0.72, radius * 0.02 + radius * 0.18 * side),
			accent_color,
			2.0,
			true
		)


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


func _draw_flame(center: Vector2, radius: float) -> void:
	var flame := PackedVector2Array([
		center + Vector2(-radius * 0.12, radius * 0.86),
		center + Vector2(-radius * 0.58, radius * 0.30),
		center + Vector2(-radius * 0.34, -radius * 0.22),
		center + Vector2(-radius * 0.08, -radius * 0.84),
		center + Vector2(radius * 0.16, -radius * 0.26),
		center + Vector2(radius * 0.48, -radius * 0.56),
		center + Vector2(radius * 0.38, radius * 0.18),
		center + Vector2(radius * 0.14, radius * 0.86),
	])
	draw_colored_polygon(flame, fill_color)
	draw_polyline(flame, accent_color, 2.0, true)
	var core := PackedVector2Array([
		center + Vector2(-radius * 0.04, radius * 0.60),
		center + Vector2(-radius * 0.22, radius * 0.18),
		center + Vector2(radius * 0.08, -radius * 0.28),
		center + Vector2(radius * 0.20, radius * 0.26),
	])
	draw_colored_polygon(core, accent_color)


func _draw_leap(center: Vector2, radius: float) -> void:
	var takeoff := center + Vector2(-radius * 0.66, radius * 0.48)
	var landing := center + Vector2(radius * 0.66, radius * 0.48)
	draw_arc(center + Vector2(0.0, radius * 0.42), radius * 0.74, PI * 1.10, PI * 1.90, 26, accent_color, 3.0, true)
	draw_circle(takeoff, radius * 0.16, fill_color)
	draw_circle(landing, radius * 0.20, fill_color)
	var body := PackedVector2Array([
		center + Vector2(-radius * 0.16, -radius * 0.68),
		center + Vector2(radius * 0.30, -radius * 0.36),
		center + Vector2(radius * 0.08, radius * 0.02),
		center + Vector2(-radius * 0.34, -radius * 0.18),
	])
	draw_colored_polygon(body, fill_color)
	draw_polyline(body, accent_color, 2.0, true)
	draw_line(center + Vector2(-radius * 0.24, radius * 0.00), center + Vector2(-radius * 0.58, radius * 0.42), accent_color, 2.4, true)
	draw_line(center + Vector2(radius * 0.02, radius * 0.04), center + Vector2(radius * 0.46, radius * 0.42), accent_color, 2.4, true)


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


func _draw_pin(center: Vector2, radius: float) -> void:
	draw_circle(center + Vector2(0.0, -radius * 0.40), radius * 0.30, fill_color)
	draw_arc(center + Vector2(0.0, -radius * 0.40), radius * 0.30, 0.0, TAU, 24, accent_color, 2.0, true)
	var spike := PackedVector2Array([
		center + Vector2(-radius * 0.18, -radius * 0.18),
		center + Vector2(radius * 0.18, -radius * 0.18),
		center + Vector2(0.0, radius * 0.82),
	])
	draw_colored_polygon(spike, fill_color)
	draw_polyline(spike, accent_color, 2.0, true)
	draw_line(center + Vector2(-radius * 0.42, radius * 0.68), center + Vector2(radius * 0.42, radius * 0.68), accent_color, 2.0, true)


func _draw_volley(center: Vector2, radius: float) -> void:
	for i in range(3):
		var offset := float(i - 1) * radius * 0.28
		var start := center + Vector2(-radius * 0.58, radius * 0.38 + offset)
		var end := center + Vector2(radius * 0.46, -radius * 0.24 + offset * 0.30)
		draw_line(start, end, fill_color, radius * 0.14, true)
		draw_line(start, end, accent_color, radius * 0.05, true)
		var head := PackedVector2Array([
			end + Vector2(radius * 0.28, -radius * 0.18),
			end + Vector2(-radius * 0.02, radius * 0.04),
			end + Vector2(-radius * 0.12, -radius * 0.28),
		])
		draw_colored_polygon(head, fill_color)
		draw_polyline(head, accent_color, 1.6, true)


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


func _draw_quake(center: Vector2, radius: float) -> void:
	var hammer := PackedVector2Array([
		center + Vector2(-radius * 0.20, -radius * 0.78),
		center + Vector2(radius * 0.58, -radius * 0.30),
		center + Vector2(radius * 0.36, radius * 0.06),
		center + Vector2(-radius * 0.42, -radius * 0.42),
	])
	draw_colored_polygon(hammer, fill_color)
	draw_polyline(hammer, accent_color, 2.0, true)
	draw_line(center + Vector2(radius * 0.14, radius * 0.00), center + Vector2(-radius * 0.42, radius * 0.66), accent_color, 4.0, true)
	var cracks := [
		[Vector2(-0.74, 0.62), Vector2(-0.42, 0.38), Vector2(-0.18, 0.68)],
		[Vector2(0.00, 0.62), Vector2(0.22, 0.34), Vector2(0.50, 0.66)],
		[Vector2(0.34, 0.20), Vector2(0.60, 0.06), Vector2(0.78, 0.24)],
	]
	for crack in cracks:
		var points := PackedVector2Array()
		for p in crack:
			points.append(center + p * radius)
		draw_polyline(points, accent_color, 2.0, false)


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
