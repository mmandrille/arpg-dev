class_name DiscoveryMinimap
extends PanelContainer

const DiscoveryMinimapStateScript := preload("res://scripts/discovery_minimap_state.gd")
const MODE_HIDDEN := "hidden"
const MODE_COMPACT := "compact"
const MODE_FULLSCREEN := "fullscreen"
const COMPACT_MAP_SIZE := Vector2(208, 208)
const COMPACT_PANEL_SIZE := Vector2(216, 216)
const FULLSCREEN_MAP_SIZE := Vector2(568, 568)
const FULLSCREEN_PANEL_SIZE := Vector2(584, 584)
const DEFAULT_PANEL_OPACITY := 0.68
const FLOOR_COLOR := Color(0.32, 0.34, 0.32, 0.58)
const WALL_COLOR := Color(0.18, 0.15, 0.11, 0.82)
const GRID_COLOR := Color(1.0, 1.0, 1.0, 0.045)
const PLAYER_COLOR := Color("#8fe8a7")
const PIN_OUTER_COLOR := Color("#f0dfbb")
const PIN_INNER_COLOR := Color("#ff7f50")
const SERVICE_MARKER_COLOR := Color("#87c7ff")
const STAIRS_MARKER_COLOR := Color("#f2cc66")
const WAYPOINT_MARKER_COLOR := Color("#b58cff")
const QUEST_PATH_COLOR := Color("#f7e47d")

var _map: Control
var _state: Dictionary = _empty_state()
var _state_tracker: DiscoveryMinimapState = DiscoveryMinimapStateScript.new()
var _display_mode: String = MODE_HIDDEN
var _session_key: String = ""
var _panel_opacity: float = DEFAULT_PANEL_OPACITY


func _ready() -> void:
	_build()


func toggle() -> void:
	set_toggle_visible(_display_mode == MODE_HIDDEN)


func cycle_display_mode() -> void:
	if _display_mode == MODE_HIDDEN:
		set_display_mode(MODE_COMPACT)
	elif _display_mode == MODE_COMPACT:
		set_display_mode(MODE_FULLSCREEN)
	else:
		set_display_mode(MODE_HIDDEN)


func set_toggle_visible(enabled: bool) -> void:
	set_display_mode(MODE_COMPACT if enabled else MODE_HIDDEN)


func set_display_mode(mode: String) -> void:
	if mode != MODE_COMPACT and mode != MODE_FULLSCREEN:
		mode = MODE_HIDDEN
	_display_mode = mode
	_apply_layout()


func sync_session(session_key: String) -> void:
	session_key = session_key.strip_edges()
	if session_key == "":
		return
	if _session_key == session_key:
		return
	_session_key = session_key
	_state_tracker.reset()
	_state = _empty_state()
	_apply_layout()


func reset_session() -> void:
	_session_key = ""
	_state_tracker.reset()
	_state = _empty_state()
	_apply_layout()


func set_state(state: Dictionary) -> void:
	_state = _normalized_state(state)
	if _map == null:
		_build()
	_apply_layout()
	_map.queue_redraw()


func sync(level: int, player_position: Vector3, light_radius: float, walls: Array, entities: Dictionary) -> void:
	set_state(_state_tracker.update(level, player_position, light_radius, walls, entities))


func set_panel_opacity(value: float) -> void:
	_panel_opacity = clampf(value, 0.0, 1.0)
	add_theme_stylebox_override("panel", _panel_style())
	if _map != null:
		_map.queue_redraw()


func get_debug_state() -> Dictionary:
	var objective: Dictionary = _state.get("objective", {})
	var quest_path: Dictionary = _state.get("quest_path", {})
	var map_size := _current_map_size()
	var marker_counts: Dictionary = _state.get("marker_counts", {})
	return {
		"visible": visible,
		"toggle_visible": _display_mode != MODE_HIDDEN,
		"display_mode": _display_mode,
		"full_screen": _display_mode == MODE_FULLSCREEN,
		"map_size_x": map_size.x,
		"map_size_y": map_size.y,
		"panel_opacity": _panel_opacity,
		"session_key": _session_key,
		"level": int(_state.get("level", 0)),
		"explored_count": int(_state.get("explored_count", 0)),
		"wall_count": int(_state.get("wall_count", 0)),
		"marker_count": int(_state.get("marker_count", 0)),
		"service_marker_count": int(marker_counts.get("service", 0)),
		"stairs_marker_count": int(marker_counts.get("stairs", 0)),
		"waypoint_marker_count": int(marker_counts.get("waypoint", 0)),
		"objective_marker_count": int(marker_counts.get("objective", 0)),
		"player_marker_x": 0.5,
		"player_marker_y": 0.5,
		"has_pin": bool(objective.get("has_pin", false)),
		"status": str(objective.get("status", "hidden")),
		"pin_status": str(objective.get("status", "hidden")),
		"pin_x": float(objective.get("pin_x", 0.5)),
		"pin_y": float(objective.get("pin_y", 0.5)),
		"has_quest_path": bool(quest_path.get("has_marker", false)),
		"quest_path_start_x": float(quest_path.get("start_x", 0.5)),
		"quest_path_start_y": float(quest_path.get("start_y", 0.5)),
		"quest_path_end_x": float(quest_path.get("end_x", 0.5)),
		"quest_path_end_y": float(quest_path.get("end_y", 0.5)),
		"quest_path_angle_radians": float(quest_path.get("angle_radians", 0.0)),
	}


func _build() -> void:
	if _map != null:
		return
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_theme_stylebox_override("panel", _panel_style())
	_map = Control.new()
	_map.draw.connect(_draw_map)
	add_child(_map)
	set_state(_state)


func _apply_layout() -> void:
	var panel_size := _current_panel_size()
	var map_size := _current_map_size()
	visible = _display_mode != MODE_HIDDEN
	custom_minimum_size = panel_size
	if _map != null:
		_map.custom_minimum_size = map_size
		_map.size = map_size
	if _display_mode == MODE_FULLSCREEN:
		set_anchors_preset(Control.PRESET_CENTER)
		offset_left = -panel_size.x * 0.5
		offset_top = -panel_size.y * 0.5
		offset_right = panel_size.x * 0.5
		offset_bottom = panel_size.y * 0.5
		z_index = 80
	else:
		set_anchors_preset(Control.PRESET_TOP_RIGHT)
		offset_left = -234
		offset_top = 18
		offset_right = -18
		offset_bottom = 234
		z_index = 0
	queue_redraw()


func _draw_map() -> void:
	var map_size := _current_map_size()
	var rect := Rect2(Vector2.ZERO, map_size)
	_map.draw_rect(rect, Color(0.024, 0.027, 0.024, _panel_opacity), true)
	for t in [0.25, 0.5, 0.75]:
		var x := map_size.x * float(t)
		var y := map_size.y * float(t)
		_map.draw_line(Vector2(x, 0), Vector2(x, map_size.y), GRID_COLOR, 1.0)
		_map.draw_line(Vector2(0, y), Vector2(map_size.x, y), GRID_COLOR, 1.0)
	_draw_explored_cells()
	_draw_walls()
	_draw_quest_path()
	_draw_poi_markers()
	_map.draw_rect(rect, Color(0.55, 0.48, 0.35, 0.52), false, 1.0)
	_map.draw_circle(map_size * 0.5, 5.0 if _display_mode != MODE_FULLSCREEN else 7.0, PLAYER_COLOR)
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
	var pin := Vector2(float(objective.get("pin_x", 0.5)), float(objective.get("pin_y", 0.5))) * _current_map_size()
	var radius := 8.0 if _display_mode != MODE_FULLSCREEN else 11.0
	_map.draw_circle(pin, radius, PIN_OUTER_COLOR)
	_map.draw_circle(pin, radius * 0.5, PIN_INNER_COLOR)


func _draw_poi_markers() -> void:
	var markers: Array = _state.get("markers", [])
	for raw in markers:
		var marker := raw as Dictionary
		var kind := str(marker.get("kind", ""))
		if kind == "objective":
			continue
		var center := Vector2(float(marker.get("x", 0.5)), float(marker.get("y", 0.5))) * _current_map_size()
		var radius := 5.0 if _display_mode != MODE_FULLSCREEN else 7.0
		match kind:
			"stairs":
				_draw_triangle_marker(center, radius, STAIRS_MARKER_COLOR)
			"waypoint":
				_draw_diamond_marker(center, radius, WAYPOINT_MARKER_COLOR)
			"service":
				_map.draw_rect(Rect2(center - Vector2(radius, radius), Vector2(radius * 2.0, radius * 2.0)), SERVICE_MARKER_COLOR, true)
			_:
				_map.draw_circle(center, radius, Color.WHITE)


func _draw_quest_path() -> void:
	var quest_path: Dictionary = _state.get("quest_path", {})
	if not bool(quest_path.get("has_marker", false)):
		return
	var start := Vector2(float(quest_path.get("start_x", 0.5)), float(quest_path.get("start_y", 0.5))) * _current_map_size()
	var end := Vector2(float(quest_path.get("end_x", 0.5)), float(quest_path.get("end_y", 0.5))) * _current_map_size()
	var delta := end - start
	if delta.length() <= 1.0:
		return
	var dir := delta.normalized()
	var start_gap := 12.0 if _display_mode != MODE_FULLSCREEN else 16.0
	var end_gap := 12.0 if _display_mode != MODE_FULLSCREEN else 16.0
	var line_start := start + dir * start_gap
	var line_end := end - dir * end_gap
	_map.draw_line(line_start, line_end, QUEST_PATH_COLOR, 2.5 if _display_mode != MODE_FULLSCREEN else 3.5)
	_draw_arrow_marker(line_end, dir, QUEST_PATH_COLOR)


func _draw_arrow_marker(center: Vector2, dir: Vector2, color: Color) -> void:
	var radius := 6.0 if _display_mode != MODE_FULLSCREEN else 8.5
	var side := Vector2(-dir.y, dir.x)
	_map.draw_polygon([
		center + dir * radius,
		center - dir * radius * 0.8 + side * radius * 0.65,
		center - dir * radius * 0.8 - side * radius * 0.65,
	], [color])


func _draw_triangle_marker(center: Vector2, radius: float, color: Color) -> void:
	_map.draw_polygon([
		center + Vector2(0, -radius),
		center + Vector2(radius, radius),
		center + Vector2(-radius, radius),
	], [color])


func _draw_diamond_marker(center: Vector2, radius: float, color: Color) -> void:
	_map.draw_polygon([
		center + Vector2(0, -radius),
		center + Vector2(radius, 0),
		center + Vector2(0, radius),
		center + Vector2(-radius, 0),
	], [color])


func _world_to_map(world: Vector2) -> Vector2:
	var player := Vector2(float(_state.get("player_x", 0.0)), float(_state.get("player_y", 0.0)))
	return _current_map_size() * 0.5 + (world - player) * _world_to_map_scale()


func _world_to_map_scale() -> float:
	var radius := maxf(1.0, float(_state.get("map_world_radius", 16.0)))
	if _display_mode == MODE_FULLSCREEN:
		radius *= 2.35
	return _current_map_size().x / (radius * 2.0)


func _cell_pixel_size() -> float:
	return maxf(2.0, _world_to_map_scale())


func _map_rect() -> Rect2:
	return Rect2(Vector2.ZERO, _current_map_size())


func _current_map_size() -> Vector2:
	return FULLSCREEN_MAP_SIZE if _display_mode == MODE_FULLSCREEN else COMPACT_MAP_SIZE


func _current_panel_size() -> Vector2:
	return FULLSCREEN_PANEL_SIZE if _display_mode == MODE_FULLSCREEN else COMPACT_PANEL_SIZE


func _normalized_state(state: Dictionary) -> Dictionary:
	var normalized := state.duplicate(true)
	normalized["explored_cells"] = state.get("explored_cells", [])
	normalized["walls"] = state.get("walls", [])
	normalized["objective"] = state.get("objective", {"has_pin": false, "status": "hidden", "pin_x": 0.5, "pin_y": 0.5})
	normalized["quest_path"] = state.get("quest_path", _empty_quest_path())
	normalized["markers"] = state.get("markers", [])
	normalized["marker_counts"] = state.get("marker_counts", {})
	normalized["explored_count"] = int(state.get("explored_count", (normalized["explored_cells"] as Array).size()))
	normalized["wall_count"] = int(state.get("wall_count", (normalized["walls"] as Array).size()))
	normalized["marker_count"] = int(state.get("marker_count", (normalized["markers"] as Array).size()))
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
		"markers": [],
		"marker_count": 0,
		"marker_counts": {},
		"objective": {"has_pin": false, "status": "hidden", "pin_x": 0.5, "pin_y": 0.5},
		"quest_path": _empty_quest_path(),
	}


static func _empty_quest_path() -> Dictionary:
	return {
		"has_marker": false,
		"start_x": 0.5,
		"start_y": 0.5,
		"end_x": 0.5,
		"end_y": 0.5,
		"direction_x": 0.0,
		"direction_y": 0.0,
		"angle_radians": 0.0,
	}


func _panel_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color(0.035, 0.032, 0.026, _panel_opacity)
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
