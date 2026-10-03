## Server-owned dungeon props rendered as batched KayKit MultiMeshes (v494 look, v536 ownership).
## The server places props as blocking wall-layout entries (kind "prop"); this file only maps each one to a
## model, scale and yaw from the presentation catalog and builds the batches once per wall layout.
class_name DungeonRoomDressing
extends RefCounted

const LibraryScript := preload("res://scripts/kit_piece_library.gd")
const ROOT_NAME := "DungeonRoomDressing"
const PROP_KIND := "prop"


static func _hash32(value: String) -> int:
	var result := 2166136261
	for byte in value.to_utf8_buffer():
		result = ((result ^ int(byte)) * 16777619) & 0xffffffff
	return result


## Placements for the prop walls in a layout. Positions are the server's; an unknown prop_id is skipped.
## Yaw is a presentation choice taken from the catalog by a stable hash of the wall id.
static func placements_from_walls(walls: Array, cfg: Dictionary) -> Array:
	var placements: Array = []
	if not bool(cfg.get("enabled", false)):
		return placements
	var catalog: Dictionary = cfg.get("props", {})
	for raw in walls:
		if typeof(raw) != TYPE_DICTIONARY or str((raw as Dictionary).get("kind", "")) != PROP_KIND:
			continue
		var wall := raw as Dictionary
		var prop: Dictionary = catalog.get(str(wall.get("prop_id", "")), {})
		if prop.is_empty():
			continue
		var pos: Dictionary = wall.get("position", {})
		var yaws: Array = prop.get("yaw_degrees", [0])
		placements.append({
			"point": Vector2(float(pos.get("x", 0.0)), float(pos.get("y", 0.0))),
			"asset_id": str(prop["asset_id"]),
			"scale": float(prop["scale"]),
			"yaw_degrees": float(yaws[_hash32(str(wall.get("id", ""))) % yaws.size()]),
		})
	return placements


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
