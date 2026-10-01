extends SceneTree

var _passed := 0
var _failed := 0


func _initialize() -> void:
	var dressing := TownPresentationLoader.dressing()
	_test_placement(dressing)
	_test_build(dressing)
	_test_ground(dressing)
	if _failed > 0:
		printerr("[gdtest] FAIL: test_town_nature_landmarks (%d passed, %d failed)" % [_passed, _failed])
		quit(1)
	else:
		print("[gdtest] PASS: test_town_nature_landmarks (%d checks)" % _passed)
		quit(0)


func _test_placement(dressing: Dictionary) -> void:
	var first := TownNatureLandmarks.placements(dressing)
	var second := TownNatureLandmarks.placements(dressing)
	_check("same data yields identical placements", first == second)
	var nature: Dictionary = dressing["nature"]
	var groups := {}
	var known := {}
	for raw in nature["groups"]:
		var group := raw as Dictionary
		groups[str(group["id"])] = 0
		for raw_asset in group["landmarks"]:
			known[str(raw_asset["asset_id"])] = true
		for raw_asset in group["variants"]:
			known[str(raw_asset["asset_id"])] = true
	for item in first:
		var p: Vector2 = item["position"]
		var radius := float(item["radius_m"])
		groups[str(item["group"])] += 1
		_check("asset belongs to catalog", known.has(str(item["asset_id"])))
		_check("nature outside palisade", p.distance_to(TownPresentationLoader.center()) >= TownPresentationLoader.radius_m() + float(nature["fence_clearance_m"]) + radius)
		for anchor in TownGroundDetail.anchors(dressing).values():
			_check("nature clears gameplay anchor", p.distance_to(anchor) >= float(nature["anchor_clearance_m"]) + radius)
		var approach: Dictionary = nature["gate_approach"]
		var end: Dictionary = approach["end_position"]
		_check("nature clears gate approach", not TownGroundDetail.capsule_contains(p, TownPresentationLoader.gate_position(), Vector2(float(end["x"]), float(end["y"])), float(approach["half_width_m"]) + radius))
	for id in groups:
		_check("group %s has placements" % id, int(groups[id]) > 0)
	var configured_groups: Array = nature["groups"]
	_check("two groups start with distinct tree silhouettes", str(((configured_groups[0] as Dictionary)["landmarks"] as Array)[0]["asset_id"]) != str(((configured_groups[1] as Dictionary)["landmarks"] as Array)[0]["asset_id"]))
	_check("variant weights affect selection", TownNatureLandmarks._weighted([{"weight": 1}, {"weight": 9}], 5) != TownNatureLandmarks._weighted([{"weight": 9}, {"weight": 1}], 5))
	var off := dressing.duplicate(true)
	(off["nature"] as Dictionary)["enabled"] = false
	_check("nature enable flag removes all placements", TownNatureLandmarks.placements(off).is_empty())
	var sparse := dressing.duplicate(true)
	for raw in (sparse["nature"] as Dictionary)["groups"]:
		(raw as Dictionary)["occupancy_percent"] = 0
	_check("zero occupancy keeps only explicit landmarks", TownNatureLandmarks.placements(sparse).size() <= first.size())
	var scaled := dressing.duplicate(true)
	var scaled_groups: Array = (scaled["nature"] as Dictionary)["groups"]
	var first_group := scaled_groups[0] as Dictionary
	var grass: Dictionary = (first_group["variants"] as Array)[1].duplicate(true)
	first_group["variants"] = [grass]
	first_group["occupancy_percent"] = 100
	(scaled_groups[1] as Dictionary)["occupancy_percent"] = 0
	grass["scale_min"] = 0.5
	grass["scale_max"] = 0.5
	var small := TownNatureLandmarks.placements(scaled)
	grass["scale_min"] = 1.5
	grass["scale_max"] = 1.5
	var large := TownNatureLandmarks.placements(scaled)
	var small_count := 0
	var large_count := 0
	for entry in small:
		if str(entry["asset_id"]) == str(grass["asset_id"]):
			small_count += 1
			_check("scatter follows lower scale fixture", is_equal_approx(float(entry["scale"]), 0.5))
	for entry in large:
		if str(entry["asset_id"]) == str(grass["asset_id"]):
			large_count += 1
			_check("scatter follows upper scale fixture", is_equal_approx(float(entry["scale"]), 1.5))
	_check("scale fixtures produce visible scatter", small_count > 0 and large_count > 0)


func _test_build(dressing: Dictionary) -> void:
	var root := TownNatureLandmarks.build(dressing)
	_check("nature root named", root.name == TownNatureLandmarks.ROOT_NAME)
	var instances := 0
	for child in root.get_children():
		_check("nature batched by asset", child is MultiMeshInstance3D)
		instances += (child as MultiMeshInstance3D).multimesh.instance_count
	_check("every placement has one instance", instances == TownNatureLandmarks.placements(dressing).size())
	root.free()


func _test_ground(dressing: Dictionary) -> void:
	var factory := GroundWallFactory.new()
	var ground := factory.make_ground_node(0)
	var expected: Dictionary = (dressing["terrain"] as Dictionary)["ground_size_m"]
	_check("town plane size follows catalog", (ground.mesh as PlaneMesh).size.is_equal_approx(Vector2(float(expected["x"]), float(expected["y"]))))
	_check("town terrain uses blend shader", ground.material_override is ShaderMaterial)
	if ground.material_override is ShaderMaterial:
		var mat := ground.material_override as ShaderMaterial
		_check("blend width follows catalog", is_equal_approx(float(mat.get_shader_parameter("blend_width_m")), float((dressing["terrain"] as Dictionary)["blend_width_m"])))
	var no_edge := TownGroundDetail.layers(dressing, TownGroundDetail.frame(dressing))[TownGroundDetail.LAYER_EDGE] as Array
	_check("legacy taupe edge disabled", no_edge.is_empty())
	ground.free()


func _check(label: String, valid: bool) -> void:
	if valid:
		_passed += 1
	else:
		_failed += 1
		printerr("[gdtest] FAIL %s" % label)
