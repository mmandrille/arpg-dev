class_name DiscoveryMinimap
extends PanelContainer

const DiscoveryMinimapStateScript := preload("res://scripts/discovery_minimap_state.gd")
const MAP_SIZE := Vector2(208, 208)
const PANEL_SIZE := Vector2(216, 216)
const PANEL_OPACITY := 0.68
const FLOOR_COLOR := Color(0.32, 0.34, 0.32, 0.58)
const WALL_COLOR := Color(0.18, 0.15, 0.11, 0.82)
const GRID_COLOR := Color(1.0, 1.0, 1.0, 0.045)
const PLAYER_COLOR := Color("#8fe8a7")
const PIN_OUTER_COLOR := Color("#f0dfbb")
const PIN_INNER_COLOR := Color("#ff7f50")

var _map: Control
var _state: Dictionary = _empty_state()
var _state_tracker: DiscoveryMinimapState = DiscoveryMinimapStateScript.new()
var _toggle_visible: bool = false


func _ready() -> void:
	_build()


func toggle() -> void:
	set_toggle_visible(not _toggle_visible)


func set_toggle_visible(enabled: bool) -> void:
	_toggle_visible = enabled
	_sync_visibility()


func set_state(state: Dictionary) -> void:
	_state = _normalized_state(state)
	if _map == null:
		_build()
	_sync_visibility()
	_map.queue_redraw()


func sync(level: int, player_position: Vector3, light_radius: float, walls: Array, entities: Dictionary) -> void:
	set_state(_state_tracker.update(level, player_position, light_radius, walls, entities))


func get_debug_state() -> Dictionary:
	var objective: Dictionary = _state.get("objective", {})
	return {
		"visible": visible,
		"toggle_visible": _toggle_visible,
		"map_size_x": MAP_SIZE.x,
		"map_size_y": MAP_SIZE.y,
		"panel_opacity": PANEL_OPACITY,
		"level": int(_state.get("level", 0)),
		"explored_count": int(_state.get("explored_count", 0)),
		"wall_count": int(_state.get("wall_count", 0)),
		"player_marker_x": 0.5,
		"player_marker_y": 0.5,
		"has_pin": bool(objective.get("has_pin", false)),
		"status": str(objective.get("status", "hidden")),
		"pin_status": str(objective.get("status", "hidden")),
		"pin_x": float(objective.get("pin_x", 0.5)),
		"pin_y": float(objective.get("pin_y", 0.5)),
	}


func _build() -> void:
	if _map != null:
		return
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	set_anchors_preset(Control.PRESET_TOP_RIGHT)
	offset_left = -234
	offset_top = 18
	offset_right = -18
	offset_bottom = 234
	custom_minimum_size = PANEL_SIZE
	add_theme_stylebox_override("panel", _panel_style())
	_map = Control.new()
	_map.custom_minimum_size = MAP_SIZE
	_map.draw.connect(_draw_map)
	add_child(_map)
	set_state(_state)


func _sync_visibility() -> void:
	visible = _toggle_visible


func _draw_map() -> void:
	var rect := Rect2(Vector2.ZERO, MAP_SIZE)
	_map.draw_rect(rect, Color(0.024, 0.027, 0.024, PANEL_OPACITY), true)
	for t in [0.25, 0.5, 0.75]:
		var x := MAP_SIZE.x * float(t)
		var y := MAP_SIZE.y * float(t)
		_map.draw_line(Vector2(x, 0), Vector2(x, MAP_SIZE.y), GRID_COLOR, 1.0)
		_map.draw_line(Vector2(0, y), Vector2(MAP_SIZE.x, y), GRID_COLOR, 1.0)
	_draw_explored_cells()
	_draw_walls()
	_map.draw_rect(rect, Color(0.55, 0.48, 0.35, 0.52), false, 1.0)
	_map.draw_circle(MAP_SIZE * 0.5, 5.0, PLAYER_COLOR)
	_draw_objective_pin()


func _draw_explored_cells() -> void:
	var cells: Array = _state.get("explored_cells", [])
	var cell_px := _cell_pixel_size()
	for raw in cells:
		var cell := raw as Dictionary
		var point := _world_to_map(Vector2(float(cell.get("x", 0)) + 0.5, float(cell.get("y", 0)) + 0.5))
		if not _map_rect().has_point(point):
			continue
		_map.draw_rect(Rect2(point - Vector2(cell_px, cell_px) * 0.5, Vector2(cell_px, cell_px)), FLOOR_COLOR, true)


func _draw_walls() -> void:
	var walls: Array = _state.get("walls", [])
	for raw in walls:
		var wall := raw as Dictionary
		var center := _world_to_map(Vector2(float(wall.get("x", 0.0)), float(wall.get("y", 0.0))))
		var scale := _world_to_map_scale()
		var size := Vector2(float(wall.get("w", 0.0)), float(wall.get("h", 0.0))) * scale
		var wall_rect := Rect2(center - size * 0.5, size)
		if _map_rect().intersects(wall_rect):
			_map.draw_rect(wall_rect, WALL_COLOR, true)


func _draw_objective_pin() -> void:
	var objective: Dictionary = _state.get("objective", {})
	if not bool(objective.get("has_pin", false)):
		return
	var pin := Vector2(float(objective.get("pin_x", 0.5)), float(objective.get("pin_y", 0.5))) * MAP_SIZE
	_map.draw_circle(pin, 8.0, PIN_OUTER_COLOR)
	_map.draw_circle(pin, 4.0, PIN_INNER_COLOR)


func _world_to_map(world: Vector2) -> Vector2:
	var player := Vector2(float(_state.get("player_x", 0.0)), float(_state.get("player_y", 0.0)))
	return MAP_SIZE * 0.5 + (world - player) * _world_to_map_scale()


func _world_to_map_scale() -> float:
	var radius := maxf(1.0, float(_state.get("map_world_radius", 16.0)))
	return MAP_SIZE.x / (radius * 2.0)


func _cell_pixel_size() -> float:
	return maxf(2.0, _world_to_map_scale())


func _map_rect() -> Rect2:
	return Rect2(Vector2.ZERO, MAP_SIZE)


func _normalized_state(state: Dictionary) -> Dictionary:
	var normalized := state.duplicate(true)
	normalized["explored_cells"] = state.get("explored_cells", [])
	normalized["walls"] = state.get("walls", [])
	normalized["objective"] = state.get("objective", {"has_pin": false, "status": "hidden", "pin_x": 0.5, "pin_y": 0.5})
	normalized["explored_count"] = int(state.get("explored_count", (normalized["explored_cells"] as Array).size()))
	normalized["wall_count"] = int(state.get("wall_count", (normalized["walls"] as Array).size()))
	return normalized


static func _empty_state() -> Dictionary:
	return {
		"level": 0,
		"player_x": 0.0,
		"player_y": 0.0,
		"light_radius": 0.0,
		"map_world_radius": 16.0,
		"explored_cells": [],
		"explored_count": 0,
		"walls": [],
		"wall_count": 0,
		"objective": {"has_pin": false, "status": "hidden", "pin_x": 0.5, "pin_y": 0.5},
	}


func _panel_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color(0.035, 0.032, 0.026, PANEL_OPACITY)
	style.border_color = Color(0.46, 0.39, 0.27, 0.72)
	style.border_width_left = 1
	style.border_width_top = 1
	style.border_width_right = 1
	style.border_width_bottom = 1
	style.content_margin_left = 4
	style.content_margin_top = 4
	style.content_margin_right = 4
	style.content_margin_bottom = 4
	return style
