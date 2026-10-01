extends SceneTree

const TownPresentationLoaderScript := preload("res://scripts/town_presentation_loader.gd")
const TownAmbientLifeScript := preload("res://scripts/town_ambient_life.gd")
const TownDressingScript := preload("res://scripts/town_dressing.gd")

var _passed := 0
var _failed := 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var dressing := TownPresentationLoaderScript.dressing()
	_test_route_math(dressing)
	await _test_catalog_build(dressing)
	_test_disabled_and_empty(dressing)
	await _test_town_lifecycle()
	if _failed > 0:
		printerr("[gdtest] FAIL: test_town_ambient_life (%d passed, %d failed)" % [_passed, _failed])
		quit(1)
	else:
		print("[gdtest] PASS: test_town_ambient_life (%d checks)" % _passed)
		quit(0)


func _test_route_math(dressing: Dictionary) -> void:
	var life: Dictionary = dressing["ambient_life"]
	var actors: Array = life["actors"]
	var idle: Dictionary = actors[0]
	var idle_position: Dictionary = idle["position"]
	var expected_idle := Vector2(float(idle_position["x"]), float(idle_position["y"]))
	var fixed := TownAmbientLifeScript.route_state(idle, 4.0)
	_check("idle actor stays at its configured point", fixed["position"] == expected_idle and not fixed["moving"])
	var patrol: Dictionary = (actors[1] as Dictionary).duplicate(true)
	patrol["phase_s"] = 0.0
	patrol["pause_s"] = 1.0
	var path: Array = patrol["path"]
	var a := Vector2(float(path[0]["x"]), float(path[0]["y"]))
	var b := Vector2(float(path[1]["x"]), float(path[1]["y"]))
	var travel_s := a.distance_to(b) / float(patrol["speed_mps"])
	var midpoint := TownAmbientLifeScript.route_state(patrol, travel_s * 0.5)
	_check("patrol interpolates along the segment", (midpoint["position"] as Vector2).distance_to(a.lerp(b, 0.5)) < 0.001 and midpoint["moving"])
	var paused := TownAmbientLifeScript.route_state(patrol, travel_s + 0.5)
	_check("patrol pauses at each route point", paused["position"] == b and not paused["moving"])
	var cycle_s := 0.0
	for i in path.size():
		var p0 := Vector2(float(path[i]["x"]), float(path[i]["y"]))
		var p1 := Vector2(float(path[(i + 1) % path.size()]["x"]), float(path[(i + 1) % path.size()]["y"]))
		cycle_s += p0.distance_to(p1) / float(patrol["speed_mps"]) + float(patrol["pause_s"])
	var initial := TownAmbientLifeScript.route_state(patrol, 0.0)
	var after_cycle := TownAmbientLifeScript.route_state(patrol, cycle_s)
	_check("patrol repeats deterministically without position drift", initial["position"].is_equal_approx(after_cycle["position"]) and initial["moving"] == after_cycle["moving"])


func _test_catalog_build(dressing: Dictionary) -> void:
	var layer := TownAmbientLifeScript.build(dressing)
	root.add_child(layer)
	await process_frame
	_check("ambient root has stable name", layer.name == TownAmbientLifeScript.ROOT_NAME)
	_check("catalog builds one node per configured resident", layer.get_child_count() == dressing["ambient_life"]["actors"].size())
	var records: Array = layer.get("_actors")
	_check("actors keep distinct catalog identities", records.size() == 2 and records[0]["node"].name != records[1]["node"].name)
	if records.size() >= 2:
		var resident := records[1]["node"] as Node3D
		var player := resident.find_child("AnimationPlayer", true, false) as AnimationPlayer
		_check("patrol installs the existing logical idle and walk clips", player != null and player.has_animation("idle") and player.has_animation("walk"))
		if player != null and player.has_animation("idle") and player.has_animation("walk"):
			_check("ambient idle and walk clips loop", player.get_animation("idle").loop_mode == Animation.LOOP_LINEAR and player.get_animation("walk").loop_mode == Animation.LOOP_LINEAR)
			_check("patrol starts in walk locomotion", str(player.current_animation) == "walk")
			var before := resident.position
			layer._process(1.0)
			_check("resident follows the configured path over time", resident.position.distance_to(before) > 0.05)
			_check("patrol animation remains on the shared walk clip", str(player.current_animation) == "walk")
		_assert_no_gameplay_nodes(layer)
	layer.free()


func _test_disabled_and_empty(dressing: Dictionary) -> void:
	var disabled := dressing.duplicate(true)
	(disabled["ambient_life"] as Dictionary)["enabled"] = false
	var disabled_layer := TownAmbientLifeScript.build(disabled)
	_check("disabled ambient config creates no residents", disabled_layer.get_child_count() == 0)
	disabled_layer.free()
	var empty := dressing.duplicate(true)
	(empty["ambient_life"] as Dictionary)["actors"] = []
	var empty_layer := TownAmbientLifeScript.build(empty)
	_check("empty roster creates no residents", empty_layer.get_child_count() == 0)
	empty_layer.free()


func _test_town_lifecycle() -> void:
	var ground := Node3D.new()
	root.add_child(ground)
	TownDressingScript.sync(ground, 0)
	await process_frame
	var town_root := ground.find_child(TownDressingScript.ROOT_NAME, false, false) as Node3D
	_check("town level owns the ambient layer", town_root != null and town_root.find_child(TownAmbientLifeScript.ROOT_NAME, false, false) != null)
	var ambient := town_root.find_child(TownAmbientLifeScript.ROOT_NAME, false, false) if town_root != null else null
	TownDressingScript.sync(ground, 0)
	_check("same-level snapshot preserves ambient layer", town_root != null and town_root.find_child(TownAmbientLifeScript.ROOT_NAME, false, false) == ambient)
	TownDressingScript.sync(ground, 1)
	_check("leaving town removes ambient presentation", ground.find_child(TownDressingScript.ROOT_NAME, false, false) == null)
	await process_frame
	ground.free()


func _assert_no_gameplay_nodes(node: Node) -> void:
	_check("ambient graph has no collision or trigger nodes", not (node is CollisionObject3D) and not (node is Area3D))
	_check("ambient nodes do not join gameplay groups", node.get_groups().is_empty())
	for child in node.get_children():
		_assert_no_gameplay_nodes(child)


func _check(label: String, valid: bool) -> void:
	if valid:
		_passed += 1
	else:
		_failed += 1
		printerr("[gdtest] FAIL %s" % label)
