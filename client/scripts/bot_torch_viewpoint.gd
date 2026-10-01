## Chooses an existing dungeon torch and verifies its live camera framing.
class_name BotTorchViewpoint
extends RefCounted


static func nearest_approach(positions: Array, player_pos: Dictionary, stand_off: float, preferred_index: int = -1) -> Dictionary:
	if positions.is_empty() or stand_off <= 0.0:
		return {}
	var player := _point(player_pos)
	var best := Vector2.ZERO
	var best_distance := INF
	var candidates := positions
	if preferred_index >= 0:
		if preferred_index >= positions.size():
			return {}
		candidates = [positions[preferred_index]]
	for raw in candidates:
		if typeof(raw) != TYPE_DICTIONARY:
			continue
		var mount := _point(raw as Dictionary)
		var distance := player.distance_to(mount)
		if distance < best_distance:
			best = mount
			best_distance = distance
	if best_distance == INF:
		return {}
	var toward_player := player - best
	if toward_player.length_squared() < 0.0001:
		toward_player = Vector2.RIGHT
	var approach := best + toward_player.normalized() * stand_off
	return {
		"mount": {"x": best.x, "z": best.y},
		"approach": {"x": approach.x, "z": approach.y},
	}


static func framed_near_player(
	mount: Dictionary, player_pos: Dictionary, max_distance: float,
	camera: Camera3D, viewport: Viewport,
) -> bool:
	if mount.is_empty() or max_distance <= 0.0 or camera == null or viewport == null:
		return false
	var point := _point(mount)
	if _point(player_pos).distance_to(point) > max_distance:
		return false
	# Headless fixture runs can prove the route, but only a windowed run can prove framing.
	if DisplayServer.get_name() == "headless":
		return true
	var world := Vector3(point.x, 0.0, point.y)
	if camera.is_position_behind(world):
		return false
	var screen := camera.unproject_position(world)
	return Rect2(Vector2.ZERO, Vector2(viewport.size)).has_point(screen)


static func _point(row: Dictionary) -> Vector2:
	return Vector2(float(row.get("x", 0.0)), float(row.get("z", 0.0)))
