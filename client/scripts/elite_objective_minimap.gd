class_name EliteObjectiveMinimap
extends PanelContainer

const MAP_SIZE := Vector2(104, 104)
const PLAYER_POS := Vector2(0.5, 0.5)

var _map: Control
var _state: Dictionary = {"visible": false, "has_pin": false, "status": "hidden", "pin_x": 0.5, "pin_y": 0.5}


func _ready() -> void:
	_build()


func set_state(state: Dictionary) -> void:
	_state = _normalized_state(state)
	if _map == null:
		_build()
	visible = bool(_state.get("visible", false))
	_map.queue_redraw()


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"has_pin": bool(_state.get("has_pin", false)),
		"status": str(_state.get("status", "hidden")),
		"pin_x": float(_state.get("pin_x", 0.5)),
		"pin_y": float(_state.get("pin_y", 0.5)),
	}


func _build() -> void:
	if _map != null:
		return
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	set_anchors_preset(Control.PRESET_TOP_RIGHT)
	offset_left = -128
	offset_top = 18
	offset_right = -18
	offset_bottom = 128
	custom_minimum_size = Vector2(110, 110)
	add_theme_stylebox_override("panel", _panel_style())
	_map = Control.new()
	_map.custom_minimum_size = MAP_SIZE
	_map.draw.connect(_draw_map)
	add_child(_map)
	set_state(_state)


func _draw_map() -> void:
	var rect := Rect2(Vector2.ZERO, MAP_SIZE)
	_map.draw_rect(rect, Color(0.024, 0.027, 0.024, 0.92), true)
	_map.draw_rect(rect, Color("#4b422f"), false, 1.0)
	for t in [0.25, 0.5, 0.75]:
		var x: float = MAP_SIZE.x * float(t)
		var y: float = MAP_SIZE.y * float(t)
		_map.draw_line(Vector2(x, 0), Vector2(x, MAP_SIZE.y), Color(1, 1, 1, 0.08), 1.0)
		_map.draw_line(Vector2(0, y), Vector2(MAP_SIZE.x, y), Color(1, 1, 1, 0.08), 1.0)
	_map.draw_circle(_map_point(PLAYER_POS), 4.0, Color("#8fe8a7"))
	if bool(_state.get("has_pin", false)):
		var pin := _map_point(Vector2(float(_state.get("pin_x", 0.5)), float(_state.get("pin_y", 0.5))))
		_map.draw_circle(pin, 7.0, Color("#f0dfbb"))
		_map.draw_circle(pin, 3.5, Color("#ff7f50"))


func _normalized_state(state: Dictionary) -> Dictionary:
	return {
		"visible": bool(state.get("visible", false)),
		"has_pin": bool(state.get("has_pin", false)),
		"status": str(state.get("status", "hidden")),
		"pin_x": clampf(float(state.get("pin_x", 0.5)), 0.0, 1.0),
		"pin_y": clampf(float(state.get("pin_y", 0.5)), 0.0, 1.0),
	}


func _map_point(normalized: Vector2) -> Vector2:
	return Vector2(normalized.x * MAP_SIZE.x, normalized.y * MAP_SIZE.y)


func _panel_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color(0.035, 0.032, 0.026, 0.86)
	style.border_color = Color("#3d3322")
	style.border_width_left = 1
	style.border_width_top = 1
	style.border_width_right = 1
	style.border_width_bottom = 1
	style.content_margin_left = 3
	style.content_margin_top = 3
	style.content_margin_right = 3
	style.content_margin_bottom = 3
	return style
