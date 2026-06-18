class_name DiscoveryMinimapState
extends RefCounted

const CELL_SIZE := 1.0
const DEFAULT_LIGHT_RADIUS := 8.0
const EDGE_PADDING := 0.08

var _explored_by_level: Dictionary = {}


func reset() -> void:
	_explored_by_level.clear()


func update(level: int, player_position: Vector3, light_radius: float, walls: Array, entities: Dictionary) -> Dictionary:
	var radius := maxf(DEFAULT_LIGHT_RADIUS, light_radius)
	var player := Vector2(player_position.x, player_position.z)
	var cells := _cells_for_level(level)
	_reveal_cells(cells, player, radius)
	var normalized_walls := _normalized_walls(walls)
	var known_walls := _known_walls(cells, normalized_walls)
	return {
		"level": level,
		"player_x": player.x,
		"player_y": player.y,
		"light_radius": radius,
		"map_world_radius": maxf(radius * 1.75, 16.0),
		"explored_cells": _serialized_cells(cells),
		"explored_count": cells.size(),
		"walls": known_walls,
		"wall_count": known_walls.size(),
		"objective": _objective_pin(entities, player, radius),
	}


func _cells_for_level(level: int) -> Dictionary:
	if not _explored_by_level.has(level):
		_explored_by_level[level] = {}
	return _explored_by_level[level] as Dictionary


func _reveal_cells(cells: Dictionary, player: Vector2, radius: float) -> void:
	var min_x := int(floor((player.x - radius) / CELL_SIZE))
	var max_x := int(ceil((player.x + radius) / CELL_SIZE))
	var min_y := int(floor((player.y - radius) / CELL_SIZE))
	var max_y := int(ceil((player.y + radius) / CELL_SIZE))
	for x in range(min_x, max_x + 1):
		for y in range(min_y, max_y + 1):
			var center := Vector2((float(x) + 0.5) * CELL_SIZE, (float(y) + 0.5) * CELL_SIZE)
			if center.distance_to(player) <= radius:
				cells[_cell_key(x, y)] = {"x": x, "y": y}


func _serialized_cells(cells: Dictionary) -> Array:
	var out: Array = []
	for key in cells.keys():
		out.append((cells[key] as Dictionary).duplicate(true))
	out.sort_custom(func(a, b):
		var ay := int((a as Dictionary).get("y", 0))
		var by := int((b as Dictionary).get("y", 0))
		if ay == by:
			return int((a as Dictionary).get("x", 0)) < int((b as Dictionary).get("x", 0))
		return ay < by
	)
	return out


func _normalized_walls(walls: Array) -> Array:
	var out: Array = []
	for raw in walls:
		if typeof(raw) != TYPE_DICTIONARY:
			continue
		var wall := raw as Dictionary
		var pos: Dictionary = wall.get("position", {})
		var size: Dictionary = wall.get("size", {})
		var w := float(size.get("x", 0.0))
		var h := float(size.get("y", 0.0))
		if w <= 0.0 or h <= 0.0:
			continue
		out.append({
			"x": float(pos.get("x", 0.0)),
			"y": float(pos.get("y", 0.0)),
			"w": w,
			"h": h,
		})
	return out


func _known_walls(cells: Dictionary, walls: Array) -> Array:
	var out: Array = []
	for raw in walls:
		var wall := raw as Dictionary
		var center_cell := _world_cell(Vector2(float(wall.get("x", 0.0)), float(wall.get("y", 0.0))))
		if cells.has(_cell_key(center_cell.x, center_cell.y)):
			out.append(wall.duplicate(true))
	return out


func _objective_pin(entities: Dictionary, player: Vector2, radius: float) -> Dictionary:
	for rec_raw in entities.values():
		var rec: Dictionary = rec_raw
		if not bool(rec.get("elite_objective", false)):
			continue
		if str(rec.get("state", "")) == "open":
			return {"has_pin": false, "status": "complete", "pin_x": 0.5, "pin_y": 0.5}
		var node := rec.get("node", null) as Node3D
		if node == null:
			return {"has_pin": false, "status": "hidden", "pin_x": 0.5, "pin_y": 0.5}
		var delta := Vector2(node.position.x, node.position.z) - player
		var map_radius := maxf(radius * 1.75, 16.0)
		return {
			"has_pin": true,
			"status": "active",
			"pin_x": clampf(0.5 + delta.x / (map_radius * 2.0), EDGE_PADDING, 1.0 - EDGE_PADDING),
			"pin_y": clampf(0.5 + delta.y / (map_radius * 2.0), EDGE_PADDING, 1.0 - EDGE_PADDING),
		}
	return {"has_pin": false, "status": "hidden", "pin_x": 0.5, "pin_y": 0.5}


func _world_cell(point: Vector2) -> Vector2i:
	return Vector2i(int(floor(point.x / CELL_SIZE)), int(floor(point.y / CELL_SIZE)))


func _cell_key(x: int, y: int) -> String:
	return "%d:%d" % [x, y]
