## World-space LOS shadow polygons and occluder normalization for fog compositor.
class_name HeroVisibilityField
extends RefCounted

const FALLBACK_WORLD_TO_SCREEN := 32.0


static func wall_blocks_line_of_sight(wall: Dictionary) -> bool:
	if wall.has("blocks_line_of_sight"):
		return bool(wall.get("blocks_line_of_sight", false))

	var kind := str(wall.get("kind", "wall"))

	return kind == "wall" or kind == "wood"


static func normalized_occluder(occluder: Dictionary) -> Dictionary:
	var pos: Dictionary = occluder.get("position", {})
	var size: Dictionary = occluder.get("size", {})
	var w := float(size.get("x", 0.0))
	var h := float(size.get("y", 0.0))
	if w <= 0.0 or h <= 0.0:
		return {}

	return {
		"x": float(pos.get("x", 0.0)),
		"y": float(pos.get("y", 0.0)),
		"w": w,
		"h": h,
	}


static func normalize_wall_layout(walls: Array) -> Array:
	var out: Array = []
	for wall in walls:
		if typeof(wall) != TYPE_DICTIONARY:
			continue
		if not wall_blocks_line_of_sight(wall as Dictionary):
			continue
		var normalized := normalized_occluder(wall as Dictionary)
		if not normalized.is_empty():
			out.append(normalized)

	return out


static func normalize_occluder_layout(occluders: Array) -> Array:
	var out: Array = []
	for occluder in occluders:
		if typeof(occluder) != TYPE_DICTIONARY:
			continue
		var normalized := normalized_occluder(occluder as Dictionary)
		if not normalized.is_empty():
			out.append(normalized)

	return out


static func combined_occluders(wall_layout: Array, extra_occluders: Array) -> Array:
	var occluders := wall_layout.duplicate()
	for occluder in extra_occluders:
		occluders.append(occluder)

	return occluders


static func build_shadow_polygons(
	camera: Camera3D,
	hero_world: Vector2,
	shadow_reach: float,
	viewport_size: Vector2,
	occluders: Array,
	shadow_cfg: Dictionary,
	fallback_center_px: Vector2,
) -> Dictionary:
	var polygons: Array = []
	var debug: Array = []
	var occluder_count := 0
	var edge_epsilon := float(shadow_cfg.get("edge_epsilon", 0.08))
	var start_offset := float(shadow_cfg.get("start_offset", 0.16))
	var wall_height := float(shadow_cfg.get("wall_height", 1.0))

	for occluder in occluders:
		var poly := shadow_polygon_for_wall(
			camera,
			occluder as Dictionary,
			hero_world,
			shadow_reach,
			viewport_size,
			edge_epsilon,
			start_offset,
			wall_height,
			fallback_center_px,
		)
		if poly.size() < 4:
			continue
		occluder_count += 1
		polygons.append(poly)
		debug.append(debug_shadow(poly))

	return {
		"polygons": polygons,
		"debug": debug,
		"occluder_count": occluder_count,
	}


static func expanded_polygon(points: Array, scale: float) -> Array:
	if points.is_empty():
		return []
	var center := Vector2.ZERO
	for point in points:
		center += point as Vector2
	center /= float(points.size())
	var out: Array = []
	for point in points:
		var p := point as Vector2
		out.append(center + (p - center) * scale)

	return out


static func shadow_polygon_for_wall(
	camera: Camera3D,
	wall: Dictionary,
	hero_world: Vector2,
	shadow_reach: float,
	viewport_size: Vector2,
	edge_epsilon: float,
	start_offset: float,
	wall_height: float,
	fallback_center_px: Vector2,
) -> Array:
	var center := Vector2(float(wall.get("x", 0.0)), float(wall.get("y", 0.0)))
	var size := Vector2(float(wall.get("w", 0.0)), float(wall.get("h", 0.0)))
	if size.x <= 0.0 or size.y <= 0.0:
		return []
	var wall_reach := size.length() * 0.5
	if hero_world.distance_to(center) > shadow_reach + wall_reach:
		return []
	if point_inside_wall(hero_world, center, size):
		return []
	var tangent := tangent_corners(hero_world, wall_corners(center, size))
	if tangent.size() < 2:
		return []
	var corner_a: Vector2 = tangent[0]
	var corner_b: Vector2 = tangent[1]
	var dir_a := (corner_a - hero_world).normalized()
	var dir_b := (corner_b - hero_world).normalized()
	if dir_a.length() <= 0.0 or dir_b.length() <= 0.0:
		return []
	var edge_offset := clampf(start_offset, edge_epsilon, minf(size.x, size.y) * 0.5)
	var extend := maxf(
		shadow_reach * 4.0,
		hero_world.distance_to(center) + viewport_size.length() / FALLBACK_WORLD_TO_SCREEN + wall_reach
	)
	var start_a := corner_a + dir_a * edge_offset
	var start_b := corner_b + dir_b * edge_offset
	var end_a := hero_world + dir_a * extend
	var end_b := hero_world + dir_b * extend

	return [
		project_world_point(camera, start_a, wall_height, hero_world, fallback_center_px),
		project_world_point(camera, end_a, 0.0, hero_world, fallback_center_px),
		project_world_point(camera, end_b, 0.0, hero_world, fallback_center_px),
		project_world_point(camera, start_b, wall_height, hero_world, fallback_center_px),
	]


static func project_world_point(
	camera: Camera3D,
	world_point: Vector2,
	height: float,
	hero_world: Vector2,
	fallback_center_px: Vector2,
) -> Vector2:
	if camera != null:
		return camera.unproject_position(Vector3(world_point.x, height, world_point.y))

	return fallback_center_px + (world_point - hero_world) * FALLBACK_WORLD_TO_SCREEN


static func tangent_corners(hero_world: Vector2, corners: Array) -> Array:
	var entries: Array = []
	for corner in corners:
		var point := corner as Vector2
		entries.append({"angle": atan2(point.y - hero_world.y, point.x - hero_world.x), "corner": point})
	entries.sort_custom(func(a, b): return float(a.get("angle", 0.0)) < float(b.get("angle", 0.0)))
	var max_gap := -1.0
	var gap_index := 0
	for i in range(entries.size()):
		var next_i := (i + 1) % entries.size()
		var a := float(entries[i].get("angle", 0.0))
		var b := float(entries[next_i].get("angle", 0.0))
		var gap := b - a
		if next_i == 0:
			gap += TAU
		if gap > max_gap:
			max_gap = gap
			gap_index = i
	var first: Vector2 = entries[(gap_index + 1) % entries.size()].get("corner", Vector2.ZERO)
	var second: Vector2 = entries[gap_index].get("corner", Vector2.ZERO)
	if first.distance_to(second) <= 0.001:
		return []

	return [first, second]


static func wall_corners(center: Vector2, size: Vector2) -> Array:
	var half := size * 0.5

	return [
		Vector2(center.x - half.x, center.y - half.y),
		Vector2(center.x + half.x, center.y - half.y),
		Vector2(center.x + half.x, center.y + half.y),
		Vector2(center.x - half.x, center.y + half.y),
	]


static func point_inside_wall(point: Vector2, center: Vector2, size: Vector2) -> bool:
	var half := size * 0.5

	return (
		point.x >= center.x - half.x
		and point.x <= center.x + half.x
		and point.y >= center.y - half.y
		and point.y <= center.y + half.y
	)


static func debug_shadow(points: Array) -> Dictionary:
	var min_p := points[0] as Vector2
	var max_p := points[0] as Vector2
	var serialized: Array = []
	for point in points:
		var p := point as Vector2
		min_p.x = minf(min_p.x, p.x)
		min_p.y = minf(min_p.y, p.y)
		max_p.x = maxf(max_p.x, p.x)
		max_p.y = maxf(max_p.y, p.y)
		serialized.append({"x": p.x, "y": p.y})

	return {
		"points": serialized,
		"bounds": {
			"min_x": min_p.x,
			"min_y": min_p.y,
			"max_x": max_p.x,
			"max_y": max_p.y,
		},
	}
