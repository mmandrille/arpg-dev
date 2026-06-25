class_name DiscoveryMinimapState
extends RefCounted

const CELL_SIZE := 1.0
const DEFAULT_LIGHT_RADIUS := 8.0
const EDGE_PADDING := 0.08
const SERVICE_DEFS := {
	"town_vendor": "vendor",
	"town_mystery_seller": "mystery",
	"town_stash": "stash",
	"town_bishop": "bishop",
	"town_market_board": "market",
	"town_blacksmith": "blacksmith",
	"town_mercenary_board": "mercenary",
	"town_unique_chest": "unique_chest",
}

var _explored_by_level: Dictionary = {}


func reset() -> void:
	_explored_by_level.clear()


func update(level: int, player_position: Vector3, light_radius: float, walls: Array, entities: Dictionary, player_facing: Vector2 = Vector2.ZERO) -> Dictionary:
	var radius := maxf(DEFAULT_LIGHT_RADIUS, light_radius)
	var player := Vector2(player_position.x, player_position.z)
	var cells := _cells_for_level(level)
	_reveal_cells(cells, player, radius)
	var normalized_walls := _normalized_walls(walls)
	var known_walls := _known_walls(cells, normalized_walls)
	var objective := _objective_pin(entities, player, radius)
	var quest_path := _quest_path_marker(objective)
	var markers := _poi_markers(cells, entities, player, radius, objective)
	var marker_counts := _marker_kind_counts(markers)
	var facing := player_facing
	if facing.length_squared() <= 0.0001:
		facing = Vector2(0.0, 1.0)
	return {
		"level": level,
		"player_x": player.x,
		"player_y": player.y,
		"player_facing_x": facing.x,
		"player_facing_y": facing.y,
		"player_facing_angle": atan2(facing.y, facing.x),
		"light_radius": radius,
		"map_world_radius": maxf(radius * 1.75, 16.0),
		"explored_cells": _serialized_cells(cells),
		"explored_count": cells.size(),
		"walls": known_walls,
		"wall_count": known_walls.size(),
		"objective": objective,
		"quest_path": quest_path,
		"markers": markers,
		"marker_count": markers.size(),
		"marker_counts": marker_counts,
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


func _quest_path_marker(objective: Dictionary) -> Dictionary:
	if not bool(objective.get("has_pin", false)):
		return _empty_quest_path()
	var target := Vector2(float(objective.get("pin_x", 0.5)), float(objective.get("pin_y", 0.5)))
	var start := Vector2(0.5, 0.5)
	var delta := target - start
	if delta.length() <= 0.001:
		return _empty_quest_path()
	var direction := delta.normalized()
	return {
		"has_marker": true,
		"start_x": start.x,
		"start_y": start.y,
		"end_x": target.x,
		"end_y": target.y,
		"direction_x": direction.x,
		"direction_y": direction.y,
		"angle_radians": atan2(direction.y, direction.x),
	}


func _poi_markers(cells: Dictionary, entities: Dictionary, player: Vector2, radius: float, objective: Dictionary) -> Array:
	var out: Array = []
	for rec_raw in entities.values():
		var rec: Dictionary = rec_raw
		if str(rec.get("type", "")) != "interactable":
			continue
		var def_id := str(rec.get("interactable_def_id", ""))
		var marker := _interactable_marker(def_id)
		if marker.is_empty():
			continue
		var node := rec.get("node", null) as Node3D
		if node == null:
			continue
		var world := Vector2(node.position.x, node.position.z)
		var cell := _world_cell(world)
		if not cells.has(_cell_key(cell.x, cell.y)):
			continue
		var normalized := _normalized_marker_position(world, player, radius)
		marker["x"] = normalized.x
		marker["y"] = normalized.y
		out.append(marker)
	if bool(objective.get("has_pin", false)):
		out.append({
			"kind": "objective",
			"label": "objective",
			"x": float(objective.get("pin_x", 0.5)),
			"y": float(objective.get("pin_y", 0.5)),
		})
	out.sort_custom(func(a, b):
		var ak := str((a as Dictionary).get("kind", ""))
		var bk := str((b as Dictionary).get("kind", ""))
		if ak == bk:
			return str((a as Dictionary).get("label", "")) < str((b as Dictionary).get("label", ""))
		return ak < bk
	)
	return out


func _interactable_marker(def_id: String) -> Dictionary:
	if def_id == "stairs_down":
		return {"kind": "stairs", "label": "down"}
	if def_id == "stairs_up":
		return {"kind": "stairs", "label": "up"}
	if def_id == "teleporter":
		return {"kind": "waypoint", "label": "waypoint"}
	if SERVICE_DEFS.has(def_id):
		return {"kind": "service", "label": str(SERVICE_DEFS[def_id])}
	return {}


func _normalized_marker_position(world: Vector2, player: Vector2, radius: float) -> Vector2:
	var map_radius := maxf(radius * 1.75, 16.0)
	var delta := world - player
	return Vector2(
		clampf(0.5 + delta.x / (map_radius * 2.0), EDGE_PADDING, 1.0 - EDGE_PADDING),
		clampf(0.5 + delta.y / (map_radius * 2.0), EDGE_PADDING, 1.0 - EDGE_PADDING)
	)


func _marker_kind_counts(markers: Array) -> Dictionary:
	var counts: Dictionary = {}
	for raw in markers:
		var marker := raw as Dictionary
		var kind := str(marker.get("kind", ""))
		if kind == "":
			continue
		counts[kind] = int(counts.get(kind, 0)) + 1
	return counts


func _world_cell(point: Vector2) -> Vector2i:
	return Vector2i(int(floor(point.x / CELL_SIZE)), int(floor(point.y / CELL_SIZE)))


func _cell_key(x: int, y: int) -> String:
	return "%d:%d" % [x, y]


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


static func sync_to_panel(
	panel,
	level: int,
	player_anchor: Node3D,
	character_progression: Dictionary,
	wall_layout: Array,
	entities: Dictionary,
	player_facing: Vector2,
) -> void:
	if panel == null:
		return
	var light_radius := float((character_progression.get("derived_stats", {}) as Dictionary).get("light_radius", 0.0))
	var player_pos := player_anchor.position if player_anchor != null else Vector3.ZERO
	panel.sync(level, player_pos, light_radius, wall_layout, entities, player_facing)
