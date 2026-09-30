## Town look (ADR-0018, v491): a paved KayKit plaza around the town centre, a road to the gate, and
## kit props grouped behind the services. Presentation only (the server
## never sees these; tools/test_town_dressing.py keeps props clear of gameplay positions). Data:
## shared/assets/town_presentation.v0.json -> center, gate_position, dressing.
##
## Everything sits under one root in town world space. The live ground plane is translated
## (GroundWallFactory.TOWN_GROUND_CENTER), so sync() offsets the root by the ground's position, and it
## removes the root on every non-town level.
class_name TownDressing
extends RefCounted

const LoaderScript := preload("res://scripts/town_presentation_loader.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")
const KitFloorScript := preload("res://scripts/dungeon_kit_floor.gd")

const ROOT_NAME := "TownDressing"
const PLAZA_NAME := "TownPlaza"
const PROP_PREFIX := "TownProp_"


## Attach (town) or remove (any other level) the dressing under the ground node.
static func sync(ground_node: Node3D, level: int) -> void:
	if ground_node == null:
		return
	var existing := ground_node.find_child(ROOT_NAME, false, false)
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
		var plaza := build_plaza(cfg.get("plaza", {}))
		if plaza != null:
			root.add_child(plaza)
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
	if p.distance_to(center) <= radius:
		return true
	var seg := gate - center
	var t := clampf((p - center).dot(seg) / maxf(seg.length_squared(), 0.0001), 0.0, 1.0)
	return p.distance_to(center + seg * t) <= half_width


static func build_plaza(plaza: Dictionary) -> Node3D:
	# Intact tiles only: the dungeon floor's broken variants have holes meant for a dark dungeon base.
	var variants: Array = plaza.get("tile_variants", [])
	if variants.is_empty():
		return null
	var ids: Array = []
	var weights: Array = []
	for v in variants:
		ids.append(str((v as Dictionary).get("asset_id", "")))
		weights.append(int((v as Dictionary).get("weight", 1)))
	var scale := float(plaza.get("tile_scale", 1.0))
	var base_box := LibraryScript.bounds(str(ids[0]))
	var tile_size := maxf(base_box.size.x, base_box.size.z) * scale
	var surface_y := float(plaza.get("surface_y", 0.0))
	var per_variant: Array = []
	for i in ids.size():
		per_variant.append([])
	for cell in plaza_cells(plaza, tile_size):
		var choice := KitFloorScript.pick(cell, 0, weights)
		(per_variant[choice.x] as Array).append(Vector3(cell.x, float(choice.y), cell.y))
	var root := Node3D.new()
	root.name = PLAZA_NAME
	for i in ids.size():
		var mesh := LibraryScript.mesh(str(ids[i]))
		var placements: Array = per_variant[i]
		if mesh == null or placements.is_empty():
			continue
		var box := mesh.get_aabb()
		var mm := MultiMesh.new()
		mm.transform_format = MultiMesh.TRANSFORM_3D
		mm.mesh = mesh
		mm.instance_count = placements.size()
		for j in placements.size():
			mm.set_instance_transform(j, KitFloorScript.tile_transform(placements[j], box, base_box, surface_y, scale))
		var instance := MultiMeshInstance3D.new()
		instance.name = "%s_%s" % [PLAZA_NAME, str(ids[i])]
		instance.multimesh = mm
		root.add_child(instance)
	return root


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
