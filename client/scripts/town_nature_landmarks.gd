## Deterministic, presentation-only nature outside the town palisade.
## Every placement comes from the schema-backed town catalog; no collision is created.
class_name TownNatureLandmarks
extends RefCounted

const GroundDetail := preload("res://scripts/town_ground_detail.gd")
const TownLoader := preload("res://scripts/town_presentation_loader.gd")
const Library := preload("res://scripts/kit_piece_library.gd")

const ROOT_NAME := "TownNatureLandmarks"


static func placements(dressing: Dictionary) -> Array:
	var nature: Dictionary = dressing.get("nature", {})
	var placed: Array = []
	if not bool(nature.get("enabled", false)):
		return placed
	var town_center := TownLoader.center()
	var gate := TownLoader.gate_position()
	var frame := GroundDetail.frame(dressing)
	var paved := GroundDetail.region_capsules(dressing, frame)
	var avoid: Array = GroundDetail.anchors(dressing).values()
	for raw in dressing.get("props", []):
		var pos: Dictionary = (raw as Dictionary).get("position", {})
		avoid.append(Vector2(float(pos.get("x", 0.0)), float(pos.get("y", 0.0))))
	for raw_group in nature.get("groups", []):
		var group := raw_group as Dictionary
		var group_center := _position(group.get("center", {}))
		var group_radius := float(group.get("radius_m", 0.0))
		var group_id := str(group.get("id", ""))
		for raw_landmark in group.get("landmarks", []):
			var landmark := raw_landmark as Dictionary
			var p := _position(landmark.get("position", {}))
			var asset_id := str(landmark.get("asset_id", ""))
			var scale := float(landmark.get("scale", 1.0))
			var radius := _footprint_radius(asset_id, scale, float(landmark.get("footprint_radius_m", 0.0)))
			if _placeable(p, radius, group_center, group_radius, town_center, gate, nature, avoid, paved, placed):
				placed.append({"group": group_id, "asset_id": asset_id, "position": p, "scale": scale, "yaw_degrees": float(landmark.get("yaw_degrees", 0.0)), "radius_m": radius})
		var cell := float(group.get("cell_m", 0.0))
		var variants: Array = group.get("variants", [])
		if cell <= 0.0 or variants.is_empty():
			continue
		var steps := ceili(group_radius / cell)
		var salt := int(group.get("seed_salt", 0))
		var occupancy := int(group.get("occupancy_percent", 0))
		for iy in range(-steps, steps + 1):
			for ix in range(-steps, steps + 1):
				var h := absi(hash(Vector3i(ix, iy, salt)))
				if h % 100 >= occupancy:
					continue
				var h2 := absi(hash(Vector3i(ix, iy, salt + 1)))
				var offset := Vector2(float((h / 100) % 1000) / 1000.0 - 0.5, float((h / 100000) % 1000) / 1000.0 - 0.5)
				var p := group_center + Vector2(ix, iy) * cell + offset * cell * 0.5
				var variant := variants[_weighted(variants, h2)] as Dictionary
				var fraction := float((h2 / 10000) % 1000) / 999.0
				var scale := lerpf(float(variant.get("scale_min", 1.0)), float(variant.get("scale_max", 1.0)), fraction)
				var asset_id := str(variant.get("asset_id", ""))
				var radius := _footprint_radius(asset_id, scale, float(variant.get("footprint_radius_m", 0.0)) * scale)
				if _placeable(p, radius, group_center, group_radius, town_center, gate, nature, avoid, paved, placed):
					placed.append({"group": group_id, "asset_id": asset_id, "position": p, "scale": scale, "yaw_degrees": float((h2 / 10000000) % 360), "radius_m": radius})
	return placed


static func _footprint_radius(asset_id: String, scale: float, configured: float) -> float:
	var box := Library.bounds(asset_id)
	var half_diagonal := Vector2(box.size.x, box.size.z).length() * 0.5 * scale
	return maxf(configured, half_diagonal)


static func _position(raw: Dictionary) -> Vector2:
	return Vector2(float(raw.get("x", 0.0)), float(raw.get("y", 0.0)))


static func _weighted(variants: Array, roll: int) -> int:
	var total := 0
	for raw in variants:
		total += int((raw as Dictionary).get("weight", 0))
	if total <= 0:
		return 0
	roll %= total
	for index in variants.size():
		roll -= int((variants[index] as Dictionary).get("weight", 0))
		if roll < 0:
			return index
	return 0


static func _placeable(p: Vector2, radius: float, group_center: Vector2, group_radius: float, town_center: Vector2, gate: Vector2, nature: Dictionary, avoid: Array, paved: Array, placed: Array) -> bool:
	if p.distance_to(group_center) + radius > group_radius:
		return false
	if p.distance_to(town_center) < TownLoader.radius_m() + float(nature.get("fence_clearance_m", 0.0)) + radius:
		return false
	var approach: Dictionary = nature.get("gate_approach", {})
	if GroundDetail.capsule_contains(p, gate, _position(approach.get("end_position", {})), float(approach.get("half_width_m", 0.0)) + radius):
		return false
	for a in avoid:
		if p.distance_to(a) < float(nature.get("anchor_clearance_m", 0.0)) + radius:
			return false
	for cap in paved:
		if GroundDetail.capsule_contains(p, cap["a"], cap["b"], float(cap["half"]) + radius):
			return false
	for other in placed:
		if p.distance_to(other["position"]) < radius + float(other["radius_m"]):
			return false
	return true


static func build(dressing: Dictionary) -> Node3D:
	var root := Node3D.new()
	root.name = ROOT_NAME
	var by_asset := {}
	for placement in placements(dressing):
		var id := str(placement["asset_id"])
		if not by_asset.has(id):
			by_asset[id] = []
		(by_asset[id] as Array).append(placement)
	for id in by_asset:
		var mesh := Library.mesh(str(id))
		if mesh == null:
			continue
		var box := mesh.get_aabb()
		var entries: Array = by_asset[id]
		var multimesh := MultiMesh.new()
		multimesh.transform_format = MultiMesh.TRANSFORM_3D
		multimesh.mesh = mesh
		multimesh.instance_count = entries.size()
		for index in entries.size():
			var entry: Dictionary = entries[index]
			multimesh.set_instance_transform(index, GroundDetail.rock_transform(entry["position"], float(entry["yaw_degrees"]), float(entry["scale"]), box, 0.0))
		var node := MultiMeshInstance3D.new()
		node.name = "Nature_%s" % str(id)
		node.multimesh = multimesh
		root.add_child(node)
	return root
