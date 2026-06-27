## Derives deterministic wall-torch mount points from dungeon wall layout.
class_name DungeonTorchPlacement
extends RefCounted

const HeroVisibilityFieldScript := preload("res://scripts/hero_visibility_field.gd")


static func placements_from_walls(walls: Array, cfg: Dictionary, level: int = -1) -> Array:
	if not bool(cfg.get("enabled", true)):
		return []
	var segment_tiles := maxf(1.0, float(cfg.get("wall_segment_tiles", 10.0)))
	var min_per_segment := clampi(int(cfg.get("torches_per_segment_min", 0)), 0, 8)
	var max_per_segment := clampi(int(cfg.get("torches_per_segment_max", 2)), min_per_segment, 8)
	var inset := maxf(0.1, float(cfg.get("wall_inset", 0.55)))
	var end_margin := maxf(0.0, float(cfg.get("wall_end_margin", 0.75)))
	var max_shader := clampi(int(cfg.get("max_shader_torches", 32)), 1, 32)
	var floor_center := _floor_center_from_walls(walls)
	var placements: Array = []
	for raw in walls:
		if typeof(raw) != TYPE_DICTIONARY:
			continue
		var wall := raw as Dictionary
		if not _is_torch_wall(wall):
			continue
		placements.append_array(
			_torches_for_wall(wall, floor_center, segment_tiles, min_per_segment, max_per_segment, inset, end_margin, level)
		)
	if placements.size() <= max_shader:
		return placements

	return _sample_evenly(placements, max_shader)


static func _is_torch_wall(wall: Dictionary) -> bool:
	var kind := str(wall.get("kind", "wall"))
	if kind in ["water", "hole", "rock", "column", "rubble"]:
		return false

	return kind == "wall" or HeroVisibilityFieldScript.wall_blocks_line_of_sight(wall)


static func _torches_for_wall(
	wall: Dictionary,
	floor_center: Vector2,
	segment_tiles: float,
	min_per_segment: int,
	max_per_segment: int,
	inset: float,
	end_margin: float,
	level: int,
) -> Array:
	var pos: Dictionary = wall.get("position", {})
	var size: Dictionary = wall.get("size", {})
	var cx := float(pos.get("x", 0.0))
	var cy := float(pos.get("y", 0.0))
	var sx := float(size.get("x", 1.0))
	var sy := float(size.get("y", 1.0))
	var wall_id := str(wall.get("id", "%.3f_%.3f" % [cx, cy]))
	var horizontal := sx >= sy
	var long_len := sx if horizontal else sy
	var chunk_count := maxi(1, int(ceil(long_len / segment_tiles)))
	var out: Array = []
	for chunk_idx in chunk_count:
		var torch_count := _deterministic_torch_count(wall_id, chunk_idx, level, min_per_segment, max_per_segment)
		if torch_count <= 0:
			continue
		var chunk_start := -long_len * 0.5 + end_margin + float(chunk_idx) * segment_tiles
		var chunk_end := minf(long_len * 0.5 - end_margin, chunk_start + segment_tiles)
		if chunk_end <= chunk_start:
			continue
		for slot_idx in torch_count:
			var t := (float(slot_idx) + 1.0) / float(torch_count + 1)
			var along := lerpf(chunk_start, chunk_end, t)
			var mount := _mount_point(cx, cy, sx, sy, horizontal, along, inset, floor_center)
			out.append(mount)

	return out


static func _mount_point(
	cx: float,
	cy: float,
	sx: float,
	sy: float,
	horizontal: bool,
	along: float,
	inset: float,
	floor_center: Vector2,
) -> Vector2:
	if horizontal:
		var y := cy + (inset if cy <= floor_center.y else -inset)
		return Vector2(cx + along, y)

	var x := cx + (inset if cx <= floor_center.x else -inset)
	return Vector2(x, cy + along)


static func _deterministic_torch_count(wall_id: String, chunk_idx: int, level: int, min_count: int, max_count: int) -> int:
	var span := max_count - min_count + 1
	if span <= 1:
		return min_count
	var seed_key := "%s:%d:%d" % [wall_id, chunk_idx, level]
	var h: int = absi(seed_key.hash())

	return min_count + (h % span)


static func _floor_center_from_walls(walls: Array) -> Vector2:
	var min_x := INF
	var min_y := INF
	var max_x := -INF
	var max_y := -INF
	var found := false
	for raw in walls:
		if typeof(raw) != TYPE_DICTIONARY:
			continue
		var wall := raw as Dictionary
		if not _is_torch_wall(wall):
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
