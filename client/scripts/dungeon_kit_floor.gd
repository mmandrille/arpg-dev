## KayKit floor tiles for a dungeon wall layout (ADR-0018 P2).
## A deterministic square grid (tile size from the first variant's mesh bounds) over the
## perimeter-bounded floor, skipping cells whose center lies inside any wall rectangle (walls,
## columns, holes and water all arrive as wall rectangles), drawn as one
## MultiMeshInstance3D per variant.
class_name DungeonKitFloor
extends RefCounted

const LoaderScript := preload("res://scripts/dungeon_kit_presentation_loader.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")


static func _rect(wall: Dictionary) -> Rect2:
	var pos: Dictionary = wall.get("position", {})
	var size: Dictionary = wall.get("size", {})
	var s := Vector2(float(size.get("x", 1.0)), float(size.get("y", 1.0)))
	return Rect2(Vector2(float(pos.get("x", 0.0)), float(pos.get("y", 0.0))) - s * 0.5, s)


static func floor_bounds(walls: Array) -> Rect2:
	var bounds := Rect2()
	var any := false
	for pass_perimeter in [true, false]:
		for raw in walls:
			if typeof(raw) != TYPE_DICTIONARY:
				continue
			var wall := raw as Dictionary
			if pass_perimeter and str(wall.get("source", "")) != "perimeter":
				continue
			var rect := _rect(wall)
			bounds = rect if not any else bounds.merge(rect)
			any = true
		if any:
			break
	return bounds


## Cell centers (x, z) that receive a floor tile.
static func plan_cells(walls: Array, tile_size: float) -> Array:
	var cells: Array = []
	if tile_size <= 0.0:
		return cells
	var bounds := floor_bounds(walls)
	if bounds.size.x <= 0.0 or bounds.size.y <= 0.0:
		return cells
	var blockers: Array = []
	for raw in walls:
		if typeof(raw) == TYPE_DICTIONARY:
			blockers.append(_rect(raw as Dictionary))
	var ix0 := floori(bounds.position.x / tile_size)
	var ix1 := ceili(bounds.end.x / tile_size)
	var iz0 := floori(bounds.position.y / tile_size)
	var iz1 := ceili(bounds.end.y / tile_size)
	for iz in range(iz0, iz1):
		for ix in range(ix0, ix1):
			var center := Vector2((float(ix) + 0.5) * tile_size, (float(iz) + 0.5) * tile_size)
			if not bounds.has_point(center):
				continue
			var blocked := false
			for rect in blockers:
				if (rect as Rect2).has_point(center):
					blocked = true
					break
			if not blocked:
				cells.append(center)
	return cells


## Deterministic weighted variant index + quarter-turn for a cell.
static func pick(cell: Vector2, level: int, weights: Array) -> Vector2i:
	var total := 0
	for w in weights:
		total += int(w)
	if total <= 0:
		return Vector2i(0, 0)
	var h := absi(hash(Vector3i(roundi(cell.x * 2.0), roundi(cell.y * 2.0), level)))
	var roll := h % total
	var index := 0
	for i in weights.size():
		roll -= int(weights[i])
		if roll < 0:
			index = i
			break
	return Vector2i(index, (h / total) % 4)


static func build(walls: Array, level: int) -> Node3D:
	var cfg := LoaderScript.floor_config()
	if not bool(cfg.get("enabled", false)):
		return null
	var variants: Array = cfg.get("variants", [])
	if variants.is_empty():
		return null
	var ids: Array = []
	var weights: Array = []
	for v in variants:
		ids.append(str((v as Dictionary).get("asset_id", "")))
		weights.append(int((v as Dictionary).get("weight", 1)))
	var base_box := LibraryScript.bounds(str(ids[0]))
	var tile_size := maxf(base_box.size.x, base_box.size.z)
	var surface_y := float(cfg.get("surface_y", 0.0))
	var per_variant: Array = []
	for i in ids.size():
		per_variant.append([])
	for cell in plan_cells(walls, tile_size):
		var choice := pick(cell, level, weights)
		(per_variant[choice.x] as Array).append(Vector3(cell.x, float(choice.y), cell.y))
	var root := Node3D.new()
	root.name = "KitFloor"
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
			var p: Vector3 = placements[j]
			var basis := Basis(Vector3.UP, deg_to_rad(90.0 * p.y))
			var center := box.get_center()
			var origin := Vector3(p.x, surface_y - box.end.y, p.z) - basis * Vector3(center.x, 0.0, center.z)
			mm.set_instance_transform(j, Transform3D(basis, origin))
		var instance := MultiMeshInstance3D.new()
		instance.name = "KitFloor_%s" % str(ids[i])
		instance.multimesh = mm
		root.add_child(instance)
	return root

