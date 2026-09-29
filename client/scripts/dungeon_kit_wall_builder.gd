## Builds KayKit wall/pillar visuals for one server wall rectangle (ADR-0018 P2).
## Per-wall tiling (not a global auto-tiler) keeps each wall's occlusion/pick identity.
## Visual space: origin at the rectangle center on the floor (y = 0); the caller parents it.
class_name DungeonKitWallBuilder
extends RefCounted

const LoaderScript := preload("res://scripts/dungeon_kit_presentation_loader.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")

const EPSILON := 0.001
## A remainder shorter than this fraction of a half piece is absorbed by stretching the previous
## piece instead of spawning a sliver.
const SLIVER_FRACTION := 0.25


## Splits a run of `length` into full/half pieces: [{piece, start, length}], covering [0, length].
static func plan_segments(length: float, full_len: float, half_len: float) -> Array:
	var segments: Array = []
	if length <= EPSILON or full_len <= EPSILON or half_len <= EPSILON:
		return segments
	var pos := 0.0
	while length - pos >= full_len - EPSILON:
		segments.append({"piece": "full", "start": pos, "length": full_len})
		pos += full_len
	if length - pos >= half_len - EPSILON:
		segments.append({"piece": "half", "start": pos, "length": half_len})
		pos += half_len
	var rem := length - pos
	if rem > EPSILON:
		if not segments.is_empty() and rem < half_len * SLIVER_FRACTION:
			var last: Dictionary = segments[segments.size() - 1]
			last["length"] = float(last["length"]) + rem
		else:
			segments.append({"piece": "half", "start": pos, "length": rem})
	return segments


static func build_wall_visual(wall: Dictionary, wall_height: float) -> Node3D:
	var ids := LoaderScript.wall_asset_ids()
	var full_box := LibraryScript.bounds(str(ids["full"]))
	var half_box := LibraryScript.bounds(str(ids["half"]))
	if full_box.size.x <= EPSILON or half_box.size.x <= EPSILON:
		return null
	var size: Dictionary = wall.get("size", {})
	var sx := float(size.get("x", 1.0))
	var sz := float(size.get("y", 1.0))
	var along_x := sx >= sz
	var length := sx if along_x else sz
	var thickness := sz if along_x else sx
	var depth := maxf(EPSILON, full_box.size.z)
	var rows: int = maxi(1, roundi(thickness / depth))
	var row_depth := thickness / float(rows)
	var root := Node3D.new()
	root.name = "KitWall"
	var segments := plan_segments(length, full_box.size.x, half_box.size.x)
	for r in rows:
		var across := -thickness * 0.5 + row_depth * (float(r) + 0.5)
		for seg in segments:
			var is_full := str(seg["piece"]) == "full"
			var asset_id := str(ids["full"] if is_full else ids["half"])
			var box := full_box if is_full else half_box
			var piece := LibraryScript.instantiate(asset_id)
			if piece == null:
				continue
			var along := -length * 0.5 + float(seg["start"]) + float(seg["length"]) * 0.5
			var scale := Vector3(
				float(seg["length"]) / box.size.x,
				wall_height / maxf(EPSILON, box.size.y),
				row_depth / maxf(EPSILON, box.size.z),
			)
			_place(piece, box, scale, along, across, along_x)
			piece.name = "Piece_%d_%d" % [r, root.get_child_count()]
			root.add_child(piece)
	return root


static func build_column_visual(wall: Dictionary, wall_height: float) -> Node3D:
	var asset_id := LoaderScript.column_asset_id()
	var box := LibraryScript.bounds(asset_id)
	var piece := LibraryScript.instantiate(asset_id)
	if piece == null or box.size.x <= EPSILON or box.size.z <= EPSILON:
		return null
	var size: Dictionary = wall.get("size", {})
	var scale := Vector3(
		float(size.get("x", 1.0)) / box.size.x,
		wall_height / maxf(EPSILON, box.size.y),
		float(size.get("y", 1.0)) / box.size.z,
	)
	_place(piece, box, scale, 0.0, 0.0, true)
	var root := Node3D.new()
	root.name = "KitColumn"
	root.add_child(piece)
	return root


## Scales the piece and centers its bounds on (along, across) with its base on y = 0.
static func _place(piece: Node3D, box: AABB, scale: Vector3, along: float, across: float, along_x: bool) -> void:
	var center := box.get_center()
	var base_offset := Vector3(-center.x * scale.x, -box.position.y * scale.y, -center.z * scale.z)
	piece.scale = scale
	if along_x:
		piece.position = Vector3(along, 0.0, across) + base_offset
	else:
		piece.rotation_degrees = Vector3(0.0, 90.0, 0.0)
		# After a +90° yaw, local +x maps to world -z and local +z to world +x.
		piece.position = Vector3(across + base_offset.z, base_offset.y, along - base_offset.x)
