extends SceneTree

## v493 town ground detail: edge/rim/paths/scatter planning derived from
## town_presentation.v0.json -> dressing. No pinned coordinates or counts.

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	var dressing := TownPresentationLoader.dressing()
	var frame := TownGroundDetail.frame(dressing)
	_test_loader_keeps_ground_detail_keys()
	_test_frame(frame)
	_test_layers_partition_and_definitions(dressing, frame)
	_test_paths_connect_plaza_to_targets(dressing, frame)
	_test_scatter_rules(dressing, frame)
	_test_scatter_is_deterministic_and_data_driven(dressing, frame)
	_test_rock_transform_seats_on_the_surface()
	_finish()


func _test_loader_keeps_ground_detail_keys() -> void:
	var dressing := TownPresentationLoader.dressing()
	for key in ["anchors", "service_paths", "edge", "scatter"]:
		_assert_true("dressing keeps %s (v491 merge-drop regression)" % key, dressing.has(key))
	_assert_true("plaza keeps rim", (dressing.get("plaza", {}) as Dictionary).has("rim"))


func _test_frame(frame: Dictionary) -> void:
	_assert_true("tile size is positive", float(frame["tile"]) > 0.0)
	_assert_true("fence radius comes from the town catalog", is_equal_approx(float(frame["fence_radius"]), TownPresentationLoader.radius_m()))


func _key(p: Vector2, frame: Dictionary) -> Vector2i:
	return TownGroundDetail.cell_index(p, frame)


func _set_of(cells: Array, frame: Dictionary) -> Dictionary:
	var out := {}
	for c in cells:
		out[_key(c, frame)] = true
	return out


func _chebyshev_steps(width_m: float, tile: float) -> int:
	return floori(width_m / tile + 0.001)


func _near(set: Dictionary, idx: Vector2i, steps: int) -> bool:
	for dy in range(-steps, steps + 1):
		for dx in range(-steps, steps + 1):
			if (dx != 0 or dy != 0) and set.has(idx + Vector2i(dx, dy)):
				return true
	return false


func _test_layers_partition_and_definitions(dressing: Dictionary, frame: Dictionary) -> void:
	var layers := TownGroundDetail.layers(dressing, frame)
	var caps := TownGroundDetail.region_capsules(dressing, frame)
	var tile := float(frame["tile"])
	var core := _set_of(layers["core"], frame)
	var rim := _set_of(layers["rim"], frame)
	var edge := _set_of(layers["edge"], frame)
	_assert_true("core is not empty", core.size() > 0)
	_assert_true("rim is not empty", rim.size() > 0)
	_assert_true("edge is not empty", edge.size() > 0)
	var overlap := 0
	for k in core:
		if rim.has(k) or edge.has(k):
			overlap += 1
	for k in rim:
		if edge.has(k):
			overlap += 1
	_assert_true("layers are disjoint", overlap == 0)
	var paved := {}
	for k in core:
		paved[k] = true
	for k in rim:
		paved[k] = true
	var bad_paved := 0
	for c in layers["core"] + layers["rim"]:
		if not TownGroundDetail.region_contains(caps, c):
			bad_paved += 1
	_assert_true("core and rim cells are inside the paved region", bad_paved == 0)
	var edge_steps := _chebyshev_steps(float((dressing["edge"] as Dictionary)["width_m"]), tile)
	var rim_steps := _chebyshev_steps(float(((dressing["plaza"] as Dictionary)["rim"] as Dictionary)["width_m"]), tile)
	var bad_edge := 0
	for c in layers["edge"]:
		var idx := _key(c, frame)
		if TownGroundDetail.region_contains(caps, c) or paved.has(idx) or not _near(paved, idx, edge_steps):
			bad_edge += 1
	_assert_true("edge cells are unpaved and within the edge width of paved", bad_edge == 0)
	var not_paved := {}
	var bad_rim := 0
	for c in layers["rim"]:
		var idx := _key(c, frame)
		for dy in range(-rim_steps, rim_steps + 1):
			for dx in range(-rim_steps, rim_steps + 1):
				var n := idx + Vector2i(dx, dy)
				if (dx != 0 or dy != 0) and not paved.has(n):
					not_paved[idx] = true
		if not not_paved.has(idx):
			bad_rim += 1
	_assert_true("rim cells touch unpaved ground within the rim width", bad_rim == 0)
	var bad_core := 0
	for c in layers["core"]:
		var idx := _key(c, frame)
		for dy in range(-rim_steps, rim_steps + 1):
			for dx in range(-rim_steps, rim_steps + 1):
				if (dx != 0 or dy != 0) and not paved.has(idx + Vector2i(dx, dy)):
					bad_core += 1
	_assert_true("core cells are deeper than the rim width", bad_core == 0)
	_assert_true("the plaza centre is core or rim", paved.has(_key(frame["center"], frame)))
	var reaches_gate := false
	for c in layers["core"] + layers["rim"]:
		if (c as Vector2).distance_to(frame["gate"]) <= float(frame["tile"]):
			reaches_gate = true
	_assert_true("the road still reaches the gate", reaches_gate)


func _test_paths_connect_plaza_to_targets(dressing: Dictionary, frame: Dictionary) -> void:
	var layers := TownGroundDetail.layers(dressing, frame)
	var paved := _set_of(layers["core"] + layers["rim"], frame)
	var anchors := TownGroundDetail.anchors(dressing)
	var targets: Array = (dressing["service_paths"] as Dictionary)["targets"]
	_assert_true("there are service path targets", targets.size() > 0)
	for id in targets:
		var target: Vector2 = anchors[str(id)]
		# 4-connected flood fill from the centre cell over paved cells.
		var seen := {}
		var stack: Array = [_key(frame["center"], frame)]
		while not stack.is_empty():
			var cur: Vector2i = stack.pop_back()
			if seen.has(cur) or not paved.has(cur):
				continue
			seen[cur] = true
			for d in [Vector2i(1, 0), Vector2i(-1, 0), Vector2i(0, 1), Vector2i(0, -1)]:
				stack.append(cur + d)
		var reached := false
		var half := float((dressing["service_paths"] as Dictionary)["path_width_m"]) * 0.5
		for idx in seen:
			if TownGroundDetail.cell_center(idx, frame).distance_to(target) <= half + float(frame["tile"]):
				reached = true
		_assert_true("path tiles connect the plaza to %s" % id, reached)


func _test_scatter_rules(dressing: Dictionary, frame: Dictionary) -> void:
	var cfg: Dictionary = dressing["scatter"]
	var placements := TownGroundDetail.scatter(dressing, frame)
	var layers := TownGroundDetail.layers(dressing, frame)
	var occupied := _set_of(layers["core"] + layers["rim"] + layers["edge"], frame)
	var center: Vector2 = frame["center"]
	var tile := float(frame["tile"])
	var steps := ceili(float(cfg["radius_m"]) / float(cfg["cell_m"]))
	var max_cells := (2 * steps + 1) * (2 * steps + 1)
	# Sanity floor derived from the grid the planner walks (5% of the expected occupied cells), so it
	# survives data retuning without pinning a count. Both kinds must appear.
	var floor_count := maxi(2, floori(float(max_cells) * float(cfg["occupancy_percent"]) / 100.0 * 0.05))
	var kinds := {}
	for p in placements:
		kinds[str(p["kind"])] = true
	_assert_true("scatter produces a healthy number of placements", placements.size() >= floor_count)
	_assert_true("scatter contains both patches and rocks", kinds.has("patch") and kinds.has("rock"))
	_assert_true("scatter count is bounded by the grid it walks", placements.size() <= max_cells)
	var avoid: Array = []
	for pos in TownGroundDetail.anchors(dressing).values():
		avoid.append(pos)
	for prop in dressing.get("props", []):
		avoid.append(Vector2(float(prop["position"]["x"]), float(prop["position"]["y"])))
	var bad_radius := 0
	var bad_fence := 0
	var bad_avoid := 0
	var bad_tiles := 0
	for p in placements:
		var pos: Vector2 = p["position"]
		var r := float(p["radius_m"])
		if pos.distance_to(center) + r > float(cfg["radius_m"]) + 0.001:
			bad_radius += 1
		if absf(pos.distance_to(center) - float(frame["fence_radius"])) < float(cfg["fence_clearance_m"]) + r - 0.001:
			bad_fence += 1
		for a in avoid:
			if pos.distance_to(a) < float(cfg["min_clearance_m"]) + r - 0.001:
				bad_avoid += 1
		var idx := _key(pos, frame)
		var reach := ceili(r / tile) + 1
		for dy in range(-reach, reach + 1):
			for dx in range(-reach, reach + 1):
				var n := idx + Vector2i(dx, dy)
				if not occupied.has(n):
					continue
				var c := TownGroundDetail.cell_center(n, frame)
				var ddx := maxf(absf(pos.x - c.x) - tile * 0.5, 0.0)
				var ddy := maxf(absf(pos.y - c.y) - tile * 0.5, 0.0)
				if Vector2(ddx, ddy).length() < r - 0.001:
					bad_tiles += 1
	_assert_true("placements stay inside the scatter radius", bad_radius == 0)
	_assert_true("placements keep clear of the fence ring", bad_fence == 0)
	_assert_true("placements keep clear of anchors and props", bad_avoid == 0)
	_assert_true("placements do not touch paved or edge tiles", bad_tiles == 0)


func _test_scatter_is_deterministic_and_data_driven(dressing: Dictionary, frame: Dictionary) -> void:
	var a := TownGroundDetail.scatter(dressing, frame)
	var b := TownGroundDetail.scatter(dressing, frame)
	_assert_true("scatter is deterministic", a == b)
	var denser := dressing.duplicate(true)
	var denser_scatter: Dictionary = denser["scatter"]
	denser_scatter["occupancy_percent"] = 100
	var c := TownGroundDetail.scatter(denser, frame)
	_assert_true("occupancy_percent drives the count", c.size() > a.size())
	var off := dressing.duplicate(true)
	var off_scatter: Dictionary = off["scatter"]
	off_scatter["enabled"] = false
	_assert_true("scatter.enabled=false yields nothing", TownGroundDetail.scatter(off, frame).is_empty())
	var no_edge := dressing.duplicate(true)
	var no_edge_cfg: Dictionary = no_edge["edge"]
	no_edge_cfg["enabled"] = false
	_assert_true("edge.enabled=false yields no edge cells", (TownGroundDetail.layers(no_edge, frame)["edge"] as Array).is_empty())


func _test_rock_transform_seats_on_the_surface() -> void:
	var box := AABB(Vector3(-0.5, -0.1, -0.4), Vector3(1.0, 0.6, 0.8))
	var t := TownGroundDetail.rock_transform(Vector2(3.0, 4.0), 37.0, 0.5, box, 0.02)
	var bottom_y := t.origin.y + box.position.y * 0.5
	_assert_true("rock bottom sits on the surface", is_equal_approx(bottom_y, 0.02))
	var centre := t * box.get_center()
	_assert_true("rock centre lands on the requested xz", is_equal_approx(centre.x, 3.0) and is_equal_approx(centre.z, 4.0))


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL %s" % label)


func _finish() -> void:
	if _fail_count > 0:
		printerr("[gdtest] FAIL: test_town_ground_detail (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)
		return
	print("[gdtest] PASS: test_town_ground_detail (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit()
