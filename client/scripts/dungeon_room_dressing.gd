## Deterministic, nonblocking KayKit props for generated dungeon floors (v494).
## The planner consumes only stable wall/interactable records and catalog data. MultiMeshes are
## built once per floor layout, never in the frame loop.
class_name DungeonRoomDressing
extends RefCounted

const FloorScript := preload("res://scripts/dungeon_kit_floor.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")
const ROOT_NAME := "DungeonRoomDressing"


static func _hash32(value: String) -> int:
	var result := 2166136261
	for byte in value.to_utf8_buffer():
		result = ((result ^ int(byte)) * 16777619) & 0xffffffff
	return result


static func _cell_hash(key: String, cell: Vector2, salt: String) -> int:
	return _hash32("%s|%d|%d|%s" % [key, roundi(cell.x * 1000.0), roundi(cell.y * 1000.0), salt])


static func _active_band(cfg: Dictionary, level: int) -> Dictionary:
	var depth := absi(level)
	for raw in cfg.get("density_bands", []):
		var band := raw as Dictionary
		var maximum = band.get("max_depth", null)
		if depth >= int(band.get("min_depth", 1)) and (maximum == null or depth <= int(maximum)):
			return band
	return {}


static func _anchor_points(records: Array) -> Array:
	var result: Array = []
	for raw in records:
		if typeof(raw) != TYPE_DICTIONARY:
			continue
		var rec := raw as Dictionary
		if str(rec.get("type", "")) != "interactable":
			continue
		var point := Vector2.ZERO
		if rec.get("position", null) is Dictionary:
			var pos: Dictionary = rec["position"]
			point = Vector2(float(pos.get("x", 0.0)), float(pos.get("y", 0.0)))
		elif rec.get("node", null) is Node3D:
			var node := rec["node"] as Node3D
			point = Vector2(node.global_position.x, node.global_position.z)
		else:
			continue
		result.append({"id": str(rec.get("id", rec.get("interactable_def_id", ""))), "def_id": str(rec.get("interactable_def_id", "")), "point": point})
	result.sort_custom(func(a: Dictionary, b: Dictionary) -> bool:
		if a["id"] != b["id"]:
			return str(a["id"]) < str(b["id"])
		var a_point: Vector2 = a["point"]
		var b_point: Vector2 = b["point"]
		return a_point.x < b_point.x if a_point.x != b_point.x else a_point.y < b_point.y
	)
	return result


static func _distance_to_rect(point: Vector2, rect: Rect2) -> float:
	var dx := maxf(maxf(rect.position.x - point.x, 0.0), point.x - rect.end.x)
	var dy := maxf(maxf(rect.position.y - point.y, 0.0), point.y - rect.end.y)
	return Vector2(dx, dy).length()


static func _distance_to_segment(point: Vector2, a: Vector2, b: Vector2) -> float:
	var segment := b - a
	var along := clampf((point - a).dot(segment) / maxf(segment.length_squared(), 0.0001), 0.0, 1.0)
	return point.distance_to(a + segment * along)


static func _max_prop_radius(props: Array) -> float:
	var radius := 0.0
	for raw in props:
		var prop := raw as Dictionary
		var box := LibraryScript.bounds(str(prop.get("asset_id", "")))
		var scale := float(prop.get("scale", 1.0))
		radius = maxf(radius, maxf(box.size.x, box.size.z) * scale * 0.5)
	return radius


static func _safe_cell(cell: Vector2, walls: Array, anchors: Array, cfg: Dictionary, radius: float) -> bool:
	var wall_margin := float(cfg.get("wall_clearance_m", 0.0)) + radius
	for raw in walls:
		if typeof(raw) == TYPE_DICTIONARY and _distance_to_rect(cell, FloorScript._rect(raw as Dictionary)) < wall_margin:
			return false
	var anchor_margin := float(cfg.get("anchor_clearance_m", 0.0)) + radius
	for anchor in anchors:
		if cell.distance_to(anchor["point"]) < anchor_margin:
			return false
	# Keep a continuous visible lane from the up stair (or first anchor) to every destination.
	# This is intentionally conservative: a direct line may cross a wall, which only omits props.
	if anchors.size() > 1:
		var origin: Vector2 = anchors[0]["point"]
		for anchor in anchors:
			if str(anchor["def_id"]) == "stairs_up":
				origin = anchor["point"]
				break
		for anchor in anchors:
			if _distance_to_segment(cell, origin, anchor["point"]) < float(cfg.get("route_clearance_m", 0.0)) + radius:
				return false
	return true


static func _weighted_prop(props: Array, roll: int) -> Dictionary:
	var total := 0
	for raw in props:
		total += int((raw as Dictionary).get("weight", 0))
	if total <= 0:
		return {}
	var index := roll % total
	for raw in props:
		var prop := raw as Dictionary
		index -= int(prop.get("weight", 0))
		if index < 0:
			return prop
	return {}


## Returns {placements, safe_candidates, reason}. No placement is a valid explicit result.
static func plan(walls: Array, floor_key: String, level: int, records: Array, cfg: Dictionary) -> Dictionary:
	var result := {"placements": [], "safe_candidates": 0, "reason": "disabled"}
	if level >= 0 or not bool(cfg.get("enabled", false)) or floor_key == "":
		return result
	var props: Array = cfg.get("props", [])
	var band := _active_band(cfg, level)
	if props.is_empty() or band.is_empty():
		return result
	var anchors := _anchor_points(records)
	var radius := _max_prop_radius(props)
	var ranked: Array = []
	for cell in FloorScript.plan_cells(walls, float(cfg.get("grid_pitch_m", 1.0))):
		if not _safe_cell(cell, walls, anchors, cfg, radius):
			continue
		ranked.append({"point": cell, "score": _cell_hash(floor_key, cell, "rank")})
	result["safe_candidates"] = ranked.size()
	if ranked.is_empty():
		result["reason"] = "no_safe_candidate"
		return result
	ranked.sort_custom(func(a: Dictionary, b: Dictionary) -> bool:
		if a["score"] != b["score"]:
			return int(a["score"]) < int(b["score"])
		var a_point: Vector2 = a["point"]
		var b_point: Vector2 = b["point"]
		return a_point.x < b_point.x if a_point.x != b_point.x else a_point.y < b_point.y
	)
	var density := int(band.get("per_thousand_cells", 1))
	var limit: int = mini(int(cfg.get("max_instances", 32)), int(band.get("max_instances", 32)))
	var target: int = mini(limit, maxi(1, roundi(float(ranked.size() * density) / 1000.0)))
	var spacing := float(cfg.get("prop_spacing_m", 0.0)) + radius * 2.0
	var placements: Array = []
	for entry in ranked:
		if placements.size() >= target:
			break
		var point: Vector2 = entry["point"]
		var nearby := false
		for placed in placements:
			if point.distance_to(placed["point"]) < spacing:
				nearby = true
				break
		if nearby:
			continue
		var prop := _weighted_prop(props, _cell_hash(floor_key, point, "asset"))
		if prop.is_empty():
			continue
		var yaws: Array = prop.get("yaw_degrees", [0])
		placements.append({"point": point, "asset_id": str(prop["asset_id"]), "scale": float(prop["scale"]), "yaw_degrees": float(yaws[_cell_hash(floor_key, point, "yaw") % yaws.size()])})
	result["placements"] = placements
	result["reason"] = "placed" if not placements.is_empty() else "spacing_excluded"
	return result


static func build(placements: Array, surface_y: float) -> Node3D:
	var root := Node3D.new()
	root.name = ROOT_NAME
	var by_asset := {}
	for placement in placements:
		var id := str(placement["asset_id"])
		if not by_asset.has(id):
			by_asset[id] = []
		(by_asset[id] as Array).append(placement)
	for id in by_asset.keys():
		var mesh := LibraryScript.mesh(str(id))
		if mesh == null:
			continue
		var box := mesh.get_aabb()
		var entries: Array = by_asset[id]
		var mm := MultiMesh.new()
		mm.transform_format = MultiMesh.TRANSFORM_3D
		mm.mesh = mesh
		mm.instance_count = entries.size()
		for i in entries.size():
			var entry: Dictionary = entries[i]
			var point: Vector2 = entry["point"]
			var scale := float(entry["scale"])
			var basis := Basis(Vector3.UP, deg_to_rad(float(entry["yaw_degrees"]))).scaled(Vector3.ONE * scale)
			var center := box.get_center()
			var origin := Vector3(point.x, surface_y - box.position.y * scale, point.y) - basis * Vector3(center.x, 0.0, center.z)
			mm.set_instance_transform(i, Transform3D(basis, origin))
		var node := MultiMeshInstance3D.new()
		node.name = "Dressing_%s" % str(id)
		node.multimesh = mm
		root.add_child(node)
	root.set_meta("placement_count", placements.size())
	return root
