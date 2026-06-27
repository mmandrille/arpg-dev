## Derives deterministic wall-torch mount points from perimeter wall layout.
class_name DungeonTorchPlacement
extends RefCounted

const PERIMETER_SOURCE := "perimeter"


static func placements_from_walls(walls: Array, cfg: Dictionary) -> Array:
	if not bool(cfg.get("enabled", true)):
		return []
	var spacing := maxf(1.0, float(cfg.get("spacing", 10.0)))
	var inset := maxf(0.25, float(cfg.get("wall_inset", 1.2)))
	var margin := maxf(0.0, float(cfg.get("edge_margin", 2.0)))
	var max_count := clampi(int(cfg.get("max_count", 8)), 1, 8)
	var floor_center := _floor_center_from_perimeter(walls)
	var candidates: Array = []
	for raw in walls:
		if typeof(raw) != TYPE_DICTIONARY:
			continue
		var wall := raw as Dictionary
		if str(wall.get("source", "")) != PERIMETER_SOURCE:
			continue
		if str(wall.get("kind", "wall")) != "wall":
			continue
		var pos: Dictionary = wall.get("position", {})
		var size: Dictionary = wall.get("size", {})
		var cx := float(pos.get("x", 0.0))
		var cy := float(pos.get("y", 0.0))
		var sx := float(size.get("x", 1.0))
		var sy := float(size.get("y", 1.0))
		if sx >= sy:
			var y := cy + (inset if cy < floor_center.y else -inset)
			var start_x := cx - sx * 0.5 + margin
			var end_x := cx + sx * 0.5 - margin
			var x := start_x
			while x <= end_x + 0.001:
				candidates.append(Vector2(x, y))
				x += spacing
		else:
			var x := cx + (inset if cx < floor_center.x else -inset)
			var start_y := cy - sy * 0.5 + margin
			var end_y := cy + sy * 0.5 - margin
			var y := start_y
			while y <= end_y + 0.001:
				candidates.append(Vector2(x, y))
				y += spacing

	if candidates.is_empty():
		return []

	return _sample_evenly(candidates, max_count)


static func _floor_center_from_perimeter(walls: Array) -> Vector2:
	var min_x := INF
	var min_y := INF
	var max_x := -INF
	var max_y := -INF
	var found := false
	for raw in walls:
		if typeof(raw) != TYPE_DICTIONARY:
			continue
		var wall := raw as Dictionary
		if str(wall.get("source", "")) != PERIMETER_SOURCE:
			continue
		var pos: Dictionary = wall.get("position", {})
		var size: Dictionary = wall.get("size", {})
		var cx := float(pos.get("x", 0.0))
		var cy := float(pos.get("y", 0.0))
		var sx := float(size.get("x", 1.0))
		var sy := float(size.get("y", 1.0))
		min_x = minf(min_x, cx - sx * 0.5)
		max_x = maxf(max_x, cx + sx * 0.5)
		min_y = minf(min_y, cy - sy * 0.5)
		max_y = maxf(max_y, cy + sy * 0.5)
		found = true
	if not found:
		return Vector2.ZERO

	return Vector2((min_x + max_x) * 0.5, (min_y + max_y) * 0.5)


static func _sample_evenly(points: Array, max_count: int) -> Array:
	if points.size() <= max_count:
		return points.duplicate()
	var out: Array = []
	var step := float(points.size()) / float(max_count)
	for i in range(max_count):
		out.append(points[int(floor(i * step))])

	return out
