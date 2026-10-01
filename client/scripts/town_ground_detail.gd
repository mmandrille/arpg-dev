## Town ground detail (ADR-0018, v493): stone rim, dirt edge band, paths to services and a
## deterministic scatter of dirt patches and stone chunks, all planned from
## town_presentation.v0.json -> dressing. Presentation only: the server never sees it.
##
## Planning functions are pure (data in, cells/placements out) so they are unit-testable without a
## scene. Placement uses hash() only (same scheme as DungeonKitFloor.pick) so every client renders
## the same town. The builders at the bottom turn a plan into MultiMeshes.
class_name TownGroundDetail
extends RefCounted

const LoaderScript := preload("res://scripts/town_presentation_loader.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")
const KitFloorScript := preload("res://scripts/dungeon_kit_floor.gd")

const PLAZA_NAME := "TownPlaza"
const RIM_NAME := "TownPlazaRim"
const EDGE_NAME := "TownPlazaEdge"
const SCATTER_NAME := "TownScatter"
const ROOT_NAME := "TownGround"

const LAYER_CORE := "core"
const LAYER_RIM := "rim"
const LAYER_EDGE := "edge"

## Tile-variant hash salts per layer (core keeps 0 so the v491 plaza look is unchanged).
const SALT_CORE := 0
const SALT_RIM := 1
const SALT_EDGE := 2
const SCATTER_SALT := 4931


## Everything the planner needs from outside the dressing dictionary.
static func frame(dressing: Dictionary) -> Dictionary:
	return {
		"center": LoaderScript.center(),
		"gate": LoaderScript.gate_position(),
		"fence_radius": LoaderScript.radius_m(),
		"tile": tile_size(dressing),
	}


## Grid pitch: the first plaza variant's footprint times plaza.tile_scale (same rule as v491).
static func tile_size(dressing: Dictionary) -> float:
	var plaza: Dictionary = dressing.get("plaza", {})
	var variants: Array = plaza.get("tile_variants", [])
	if variants.is_empty():
		return 0.0
	var box := LibraryScript.bounds(str((variants[0] as Dictionary).get("asset_id", "")))
	return maxf(box.size.x, box.size.z) * float(plaza.get("tile_scale", 1.0))


static func capsule_contains(p: Vector2, a: Vector2, b: Vector2, half: float) -> bool:
	var seg := b - a
	var t := clampf((p - a).dot(seg) / maxf(seg.length_squared(), 0.0001), 0.0, 1.0)
	return p.distance_to(a + seg * t) <= half


static func anchors(dressing: Dictionary) -> Dictionary:
	var out := {}
	for raw in dressing.get("anchors", []):
		if typeof(raw) != TYPE_DICTIONARY:
			continue
		var entry := raw as Dictionary
		var pos: Dictionary = entry.get("position", {})
		out[str(entry.get("id", ""))] = Vector2(float(pos.get("x", 0.0)), float(pos.get("y", 0.0)))
	return out


## One capsule from the town centre to each configured service anchor.
static func path_segments(dressing: Dictionary, center: Vector2) -> Array:
	var cfg: Dictionary = dressing.get("service_paths", {})
	var half := float(cfg.get("path_width_m", 0.0)) * 0.5
	var known := anchors(dressing)
	var out: Array = []
	if half <= 0.0:
		return out
	for id in cfg.get("targets", []):
		if known.has(str(id)):
			out.append({"id": str(id), "a": center, "b": known[str(id)], "half": half})
	return out


## Paved region as capsules: the plaza disc (a == b), the gate road, and the service paths.
static func region_capsules(dressing: Dictionary, frame_data: Dictionary) -> Array:
	var plaza: Dictionary = dressing.get("plaza", {})
	var center: Vector2 = frame_data["center"]
	var caps: Array = [
		{"a": center, "b": center, "half": float(plaza.get("radius_m", 0.0))},
		{"a": center, "b": frame_data["gate"], "half": float(plaza.get("path_width_m", 0.0)) * 0.5},
	]
	caps.append_array(path_segments(dressing, center))
	return caps


static func region_contains(caps: Array, p: Vector2) -> bool:
	for cap in caps:
		if capsule_contains(p, cap["a"], cap["b"], float(cap["half"])):
			return true
	return false


static func cell_index(p: Vector2, frame_data: Dictionary) -> Vector2i:
	var rel: Vector2 = (p - (frame_data["center"] as Vector2)) / float(frame_data["tile"])
	return Vector2i(roundi(rel.x), roundi(rel.y))


static func cell_center(idx: Vector2i, frame_data: Dictionary) -> Vector2:
	return (frame_data["center"] as Vector2) + Vector2(idx) * float(frame_data["tile"])


## True when any other cell within `steps` (Chebyshev) is in `set` == `want`.
static func _has_neighbour(paved: Dictionary, idx: Vector2i, steps: int, want_paved: bool) -> bool:
	for dy in range(-steps, steps + 1):
		for dx in range(-steps, steps + 1):
			if (dx != 0 or dy != 0) and paved.has(idx + Vector2i(dx, dy)) == want_paved:
				return true
	return false


## Grid layers. core: paved. rim: paved cells within plaza.rim.width_m of unpaved ground. edge:
## unpaved cells within edge.width_m of paved ground. Distances are Chebyshev between cell centres,
## so width_m == tile size selects the 8-neighbour ring.
static func layers(dressing: Dictionary, frame_data: Dictionary) -> Dictionary:
	var out := {LAYER_CORE: [], LAYER_RIM: [], LAYER_EDGE: []}
	var tile := float(frame_data["tile"])
	if tile <= 0.0:
		return out
	var caps := region_capsules(dressing, frame_data)
	var center: Vector2 = frame_data["center"]
	var rim_cfg: Dictionary = (dressing.get("plaza", {}) as Dictionary).get("rim", {})
	var edge_cfg: Dictionary = dressing.get("edge", {})
	var rim_steps := floori(float(rim_cfg.get("width_m", 0.0)) / tile + 0.001)
	var edge_steps := 0
	if bool(edge_cfg.get("enabled", false)):
		edge_steps = floori(float(edge_cfg.get("width_m", 0.0)) / tile + 0.001)
	var reach := 0.0
	for cap in caps:
		reach = maxf(reach, maxf(center.distance_to(cap["a"]), center.distance_to(cap["b"])) + float(cap["half"]))
	var span := ceili(reach / tile) + edge_steps + 1
	var paved := {}
	for iy in range(-span, span + 1):
		for ix in range(-span, span + 1):
			var idx := Vector2i(ix, iy)
			if region_contains(caps, cell_center(idx, frame_data)):
				paved[idx] = true
	for idx in paved:
		var layer := LAYER_RIM if rim_steps > 0 and _has_neighbour(paved, idx, rim_steps, false) else LAYER_CORE
		(out[layer] as Array).append(cell_center(idx, frame_data))
	if edge_steps > 0:
		for iy in range(-span, span + 1):
			for ix in range(-span, span + 1):
				var idx := Vector2i(ix, iy)
				if not paved.has(idx) and _has_neighbour(paved, idx, edge_steps, true):
					(out[LAYER_EDGE] as Array).append(cell_center(idx, frame_data))
	return out


static func _weighted(variants: Array, roll: int) -> int:
	var total := 0
	for v in variants:
		total += int((v as Dictionary).get("weight", 1))
	if total <= 0:
		return 0
	roll = roll % total
	for i in variants.size():
		roll -= int((variants[i] as Dictionary).get("weight", 1))
		if roll < 0:
			return i
	return 0


## Scatter plan: a jittered grid of scatter.cell_m cells inside scatter.radius_m. Each cell rolls
## occupancy, kind (patch/rock), variant, scale and yaw from hash(); a placement is dropped when its
## footprint would touch the fence ring, an anchor or v491 prop, or a paved/edge tile.
static func scatter(dressing: Dictionary, frame_data: Dictionary, planned_layers: Dictionary = {}) -> Array:
	var cfg: Dictionary = dressing.get("scatter", {})
	var out: Array = []
	var tile := float(frame_data["tile"])
	var radius := float(cfg.get("radius_m", 0.0))
	var cell_m := float(cfg.get("cell_m", 0.0))
	if not bool(cfg.get("enabled", false)) or tile <= 0.0 or radius <= 0.0 or cell_m <= 0.0:
		return out
	var patches: Array = cfg.get("patches", [])
	var rocks: Array = cfg.get("rocks", [])
	if patches.is_empty() and rocks.is_empty():
		return out
	var center: Vector2 = frame_data["center"]
	var fence_radius := float(frame_data["fence_radius"])
	var jitter_fraction := float(cfg.get("jitter_fraction", 0.0))
	var clearance := float(cfg.get("min_clearance_m", 0.0))
	var fence_clear := float(cfg.get("fence_clearance_m", 0.0))
	var occupancy := int(cfg.get("occupancy_percent", 0))
	var patch_share := int(cfg.get("patch_share_percent", 0))
	var avoid: Array = anchors(dressing).values()
	for prop in dressing.get("props", []):
		var pp: Dictionary = (prop as Dictionary).get("position", {})
		avoid.append(Vector2(float(pp.get("x", 0.0)), float(pp.get("y", 0.0))))
	var lay := planned_layers if not planned_layers.is_empty() else layers(dressing, frame_data)
	var occupied := {}
	for name in [LAYER_CORE, LAYER_RIM, LAYER_EDGE]:
		for c in lay[name]:
			occupied[cell_index(c, frame_data)] = true
	var steps := ceili(radius / cell_m)
	for iy in range(-steps, steps + 1):
		for ix in range(-steps, steps + 1):
			var h := absi(hash(Vector3i(ix, iy, SCATTER_SALT)))
			if h % 100 >= occupancy:
				continue
			var h2 := absi(hash(Vector3i(ix, iy, SCATTER_SALT + 1)))
			var jitter := Vector2(float((h / 100) % 1000) / 1000.0 - 0.5, float((h / 100000) % 1000) / 1000.0 - 0.5)
			var pos := center + Vector2(ix, iy) * cell_m + jitter * cell_m * jitter_fraction
			var is_patch := (h2 % 100) < patch_share
			if patches.is_empty():
				is_patch = false
			elif rocks.is_empty():
				is_patch = true
			var group: Array = patches if is_patch else rocks
			var variant := group[_weighted(group, (h2 / 100) % 100000)] as Dictionary
			var t := float((h2 / 10000) % 1000) / 999.0
			var scale := lerpf(float(variant.get("scale_min", 1.0)), float(variant.get("scale_max", 1.0)), t)
			var r := float(variant.get("radius_m", 0.0)) * scale
			if not _placeable(pos, r, center, radius, fence_radius, fence_clear, clearance, avoid, occupied, frame_data):
				continue
			out.append({
				"kind": "patch" if is_patch else "rock",
				"asset_id": str(variant.get("asset_id", "")),
				"position": pos,
				"yaw_quarter": (h2 / 10000000) % 4,
				"yaw_degrees": float((h2 / 7) % 360),
				"scale": scale,
				"radius_m": r,
			})
	return out


static func _placeable(pos: Vector2, r: float, center: Vector2, radius: float, fence_radius: float, fence_clear: float, clearance: float, avoid: Array, occupied: Dictionary, frame_data: Dictionary) -> bool:
	var d := pos.distance_to(center)
	if d + r > radius:
		return false
	if absf(d - fence_radius) < fence_clear + r:
		return false
	for a in avoid:
		if pos.distance_to(a) < clearance + r:
			return false
	var tile := float(frame_data["tile"])
	var idx := cell_index(pos, frame_data)
	var reach := ceili(r / tile) + 1
	for dy in range(-reach, reach + 1):
		for dx in range(-reach, reach + 1):
			var n := idx + Vector2i(dx, dy)
			if not occupied.has(n):
				continue
			var c := cell_center(n, frame_data)
			var gap := Vector2(maxf(absf(pos.x - c.x) - tile * 0.5, 0.0), maxf(absf(pos.y - c.y) - tile * 0.5, 0.0))
			if gap.length() < r:
				return false
	return true


## Free-yaw placement that rests the piece's bottom on surface_y and centres it on `pos` (x, z).
static func rock_transform(pos: Vector2, yaw_degrees: float, scale: float, box: AABB, surface_y: float) -> Transform3D:
	var basis := Basis(Vector3.UP, deg_to_rad(yaw_degrees)).scaled(Vector3.ONE * scale)
	var c := box.get_center()
	var origin := Vector3(pos.x, surface_y - box.position.y * scale, pos.y) - basis * Vector3(c.x, 0.0, c.z)
	return Transform3D(basis, origin)


## The whole ground detail under one node, in town world space (x, z = town x, y).
static func build(dressing: Dictionary) -> Node3D:
	var root := Node3D.new()
	root.name = ROOT_NAME
	var frame_data := frame(dressing)
	if float(frame_data["tile"]) <= 0.0:
		return root
	var plaza: Dictionary = dressing.get("plaza", {})
	var scale := float(plaza.get("tile_scale", 1.0))
	var base_id := str(((plaza.get("tile_variants", []) as Array)[0] as Dictionary).get("asset_id", ""))
	var base_box := LibraryScript.bounds(base_id)
	var lay := layers(dressing, frame_data)
	var rim_cfg: Dictionary = plaza.get("rim", {})
	var edge_cfg: Dictionary = dressing.get("edge", {})
	var plaza_y := float(plaza.get("surface_y", 0.0))
	_add(root, _tile_layer(PLAZA_NAME, lay[LAYER_CORE], plaza.get("tile_variants", []), base_box, scale, plaza_y, SALT_CORE))
	_add(root, _tile_layer(RIM_NAME, lay[LAYER_RIM], rim_cfg.get("tile_variants", []), base_box, scale, plaza_y, SALT_RIM))
	_add(root, _tile_layer(EDGE_NAME, lay[LAYER_EDGE], edge_cfg.get("tile_variants", []), base_box, scale, float(edge_cfg.get("surface_y", plaza_y)), SALT_EDGE))
	_add(root, _scatter_layer(dressing, frame_data, base_box, lay))
	return root


static func _add(root: Node3D, layer: Node3D) -> void:
	if layer != null:
		root.add_child(layer)


static func _multimesh_instance(node_name: String, mesh: Mesh, transforms: Array) -> MultiMeshInstance3D:
	var mm := MultiMesh.new()
	mm.transform_format = MultiMesh.TRANSFORM_3D
	mm.mesh = mesh
	mm.instance_count = transforms.size()
	for i in transforms.size():
		mm.set_instance_transform(i, transforms[i])
	var instance := MultiMeshInstance3D.new()
	instance.name = node_name
	instance.multimesh = mm
	return instance


## One MultiMesh per variant; the variant and quarter-turn come from DungeonKitFloor.pick(cell, salt).
static func _tile_layer(layer_name: String, cells: Array, variants: Array, base_box: AABB, scale: float, surface_y: float, salt: int) -> Node3D:
	if cells.is_empty() or variants.is_empty():
		return null
	var ids: Array = []
	var weights: Array = []
	var per_variant: Array = []
	for v in variants:
		ids.append(str((v as Dictionary).get("asset_id", "")))
		weights.append(int((v as Dictionary).get("weight", 1)))
		per_variant.append([])
	for cell in cells:
		var choice := KitFloorScript.pick(cell, salt, weights)
		(per_variant[choice.x] as Array).append(Vector3(cell.x, float(choice.y), cell.y))
	var layer := Node3D.new()
	layer.name = layer_name
	for i in ids.size():
		var mesh := LibraryScript.mesh(str(ids[i]))
		var placements: Array = per_variant[i]
		if mesh == null or placements.is_empty():
			continue
		var box := mesh.get_aabb()
		var transforms: Array = []
		for cell3 in placements:
			transforms.append(KitFloorScript.tile_transform(cell3, box, base_box, surface_y, scale))
		layer.add_child(_multimesh_instance("%s_%s" % [layer_name, str(ids[i])], mesh, transforms))
	return layer


## Patches are seated like tiles (quarter turns, top on the plain tile's slab); rocks rest on the surface.
static func _scatter_layer(dressing: Dictionary, frame_data: Dictionary, base_box: AABB, planned_layers: Dictionary) -> Node3D:
	var placements := scatter(dressing, frame_data, planned_layers)
	if placements.is_empty():
		return null
	var surface_y := float((dressing.get("scatter", {}) as Dictionary).get("surface_y", 0.0))
	var by_asset := {}
	for p in placements:
		var id := str(p["asset_id"])
		if not by_asset.has(id):
			by_asset[id] = []
		(by_asset[id] as Array).append(p)
	var layer := Node3D.new()
	layer.name = SCATTER_NAME
	for id in by_asset:
		var mesh := LibraryScript.mesh(str(id))
		if mesh == null:
			continue
		var box := mesh.get_aabb()
		var transforms: Array = []
		for p in by_asset[id]:
			var pos: Vector2 = p["position"]
			if str(p["kind"]) == "patch":
				transforms.append(KitFloorScript.tile_transform(Vector3(pos.x, float(p["yaw_quarter"]), pos.y), box, base_box, surface_y, float(p["scale"])))
			else:
				transforms.append(rock_transform(pos, float(p["yaw_degrees"]), float(p["scale"]), box, surface_y))
		layer.add_child(_multimesh_instance("%s_%s" % [SCATTER_NAME, str(id)], mesh, transforms))
	return layer
