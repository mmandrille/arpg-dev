## Town look (ADR-0018, v491): a paved KayKit plaza around the town centre, a road to the gate, and
## kit props grouped behind the services; the ground detail (rim, edge, paths, scatter; v493) is
## built by TownGroundDetail. Presentation only: the server never sees these, and
## tools/test_town_dressing.py keeps props clear of gameplay positions.
## Data: shared/assets/town_presentation.v0.json -> center, gate_position, dressing.
##
## Everything sits under one root in town world space. The live ground plane is translated
## (GroundWallFactory.TOWN_GROUND_CENTER), so sync() offsets the root by the ground's position, and it
## removes the root on every non-town level.
class_name TownDressing
extends RefCounted

const LoaderScript := preload("res://scripts/town_presentation_loader.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")
const GroundDetailScript := preload("res://scripts/town_ground_detail.gd")
const NatureLandmarksScript := preload("res://scripts/town_nature_landmarks.gd")

const ROOT_NAME := "TownDressing"
const PLAZA_NAME := GroundDetailScript.PLAZA_NAME
const PROP_PREFIX := "TownProp_"


static func sync_with_trace(ground_node: Node3D, level: int, trace_enabled: bool) -> void:
	var started := Time.get_ticks_usec()
	sync(ground_node, level)
	if trace_enabled:
		print("[client-startup] town_dressing_ms=%.3f" % (float(Time.get_ticks_usec() - started) / 1000.0))


## Attach (town) or remove (any other level) the dressing under the ground node.
static func sync(ground_node: Node3D, level: int) -> void:
	if ground_node == null:
		return
	var existing := ground_node.find_child(ROOT_NAME, false, false)
	if level == 0 and existing != null:
		# The dressing is static for the session. A snapshot can set the same level
		# again; preserve the already-built root and keep its world alignment.
		(existing as Node3D).position = -ground_node.position
		return
	if existing != null:
		ground_node.remove_child(existing)
		existing.queue_free()
	if level != 0:
		return
	var root := build()
	root.position = -ground_node.position
	ground_node.add_child(root)


## The dressing in town world space (x, z = town x, y).
static func build() -> Node3D:
	var root := Node3D.new()
	root.name = ROOT_NAME
	var cfg := LoaderScript.dressing()
	if bool(cfg.get("enabled", false)):
		root.add_child(GroundDetailScript.build(cfg))
		root.add_child(NatureLandmarksScript.build(cfg))
		var props: Array = cfg.get("props", [])
		for i in props.size():
			var node := _make_prop(props[i], i)
			if node != null:
				root.add_child(node)
	return root


## Tile cell centres: grid at the kit tile size, aligned on the centre, inside the plaza disc or the
## road (a path_width_m band from the centre to the gate).
static func plaza_cells(plaza: Dictionary, tile_size: float) -> Array:
	var cells: Array = []
	if tile_size <= 0.0:
		return cells
	var center := LoaderScript.center()
	var gate := LoaderScript.gate_position()
	var radius := float(plaza.get("radius_m", 0.0))
	var half_width := float(plaza.get("path_width_m", 0.0)) * 0.5
	var reach := maxf(radius, center.distance_to(gate)) + tile_size
	var steps := int(ceil(reach / tile_size))
	for iy in range(-steps, steps + 1):
		for ix in range(-steps, steps + 1):
			var cell := center + Vector2(ix, iy) * tile_size
			if in_plaza(cell, center, gate, radius, half_width):
				cells.append(cell)
	return cells


static func in_plaza(p: Vector2, center: Vector2, gate: Vector2, radius: float, half_width: float) -> bool:
	return p.distance_to(center) <= radius or GroundDetailScript.capsule_contains(p, center, gate, half_width)


## Prop node name: PROP_PREFIX + index + asset id (unique even when an asset repeats).
static func prop_node_name(index: int, asset_id: String) -> String:
	return "%s%02d_%s" % [PROP_PREFIX, index, asset_id]


static func _make_prop(prop, index: int) -> Node3D:
	if typeof(prop) != TYPE_DICTIONARY:
		return null
	var entry := prop as Dictionary
	var asset_id := str(entry.get("asset_id", ""))
	var node := LibraryScript.instantiate(asset_id)
	if node == null:
		return null
	var pos: Dictionary = entry.get("position", {})
	node.name = prop_node_name(index, asset_id)
	node.position = Vector3(float(pos.get("x", 0.0)), 0.0, float(pos.get("y", 0.0)))
	node.rotation.y = deg_to_rad(float(entry.get("yaw_degrees", 0.0)))
	node.scale = Vector3.ONE * float(entry.get("scale", 1.0))
	return node
