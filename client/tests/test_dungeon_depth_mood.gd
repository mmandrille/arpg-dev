extends SceneTree

const DepthLightingScript := preload("res://scripts/dungeon_depth_lighting.gd")
const DepthMoodLoaderScript := preload("res://scripts/dungeon_depth_mood_loader.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")
const RenderLoaderScript := preload("res://scripts/render_presentation_loader.gd")
const SceneLightingRigScript := preload("res://scripts/scene_lighting_rig.gd")
const TorchLoaderScript := preload("res://scripts/dungeon_torch_presentation_loader.gd")
const TorchLightsScript := preload("res://scripts/dungeon_torch_lights.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_every_existing_palette_has_a_distinct_profile()
	_test_unknown_palette_uses_the_base_profiles()
	_test_rig_applies_selected_fog_and_tier_gate()
	await _test_torch_profile_reaches_live_nodes_and_depth_switch()
	_finish()


func _test_every_existing_palette_has_a_distinct_profile() -> void:
	var factory := GroundWallFactoryScript.new()
	var seen_ids: Array[String] = []
	var fog_signatures: Array[String] = []
	var torch_signatures: Array[String] = []
	for level in [-1, -5, -8]:
		var palette_id := DepthLightingScript.palette_id_for_level(level, factory)
		var mood := DepthMoodLoaderScript.profile_for_palette_id(palette_id)
		if palette_id.is_empty() or mood.is_empty():
			_fail("level %d must resolve an existing mood profile" % level)
			return
		seen_ids.append(palette_id)
		fog_signatures.append(str(mood.get("fog", {})))
		torch_signatures.append(str(mood.get("torch", {})))
	if seen_ids.size() != DepthMoodLoaderScript.profile_ids().size():
		_fail("every catalog palette profile must be reached by the existing level mapping")
		return
	if fog_signatures.duplicate().size() != fog_signatures.size() or torch_signatures.duplicate().size() != torch_signatures.size():
		_fail("each depth identity must have distinct fog and torch settings")
		return
	_pass("all existing biome palette identities have distinct mood profiles")


func _test_unknown_palette_uses_the_base_profiles() -> void:
	var base_context := RenderLoaderScript.context(RenderLoaderScript.CONTEXT_DUNGEON)
	var base_torches := TorchLoaderScript.config()
	if not DepthMoodLoaderScript.profile_for_palette_id("unknown-palette").is_empty():
		_fail("unknown palette must use an empty mood override")
		return
	if DepthMoodLoaderScript.context_with_fog(base_context, "unknown-palette") != base_context:
		_fail("unknown palette must retain base dungeon render context")
		return
	if DepthMoodLoaderScript.torch_config(base_torches, "unknown-palette") != base_torches:
		_fail("unknown palette must retain base torch configuration")
		return
	_pass("unknown palette safely retains existing base presentation")


func _test_rig_applies_selected_fog_and_tier_gate() -> void:
	var factory := GroundWallFactoryScript.new()
	var root := Node3D.new()
	root.name = "DepthMoodLightingTest"
	get_root().add_child(root)
	var rig = SceneLightingRigScript.new()
	rig.attach(root)
	for item in [{"level": -1, "id": "shallow_cave"}, {"level": -5, "id": "sundered_halls"}, {"level": -8, "id": "deep_vault"}]:
		var expected_id := str(item["id"])
		var level := int(item["level"])
		var profile: Dictionary = DepthLightingScript.profile_for_level(level, factory)
		if str(profile.get("palette_id", "")) != expected_id:
			_fail("depth lighting must return the palette identity used by the rig")
			root.queue_free()
			return
		rig.sync(level, factory, "balanced")
		var expected_fog: Dictionary = DepthMoodLoaderScript.profile_for_palette_id(expected_id)["fog"]
		var env: Environment = rig.world_environment.environment
		if not env.fog_enabled or not is_equal_approx(env.fog_density, float(expected_fog["density"])):
			_fail("balanced rig must apply selected depth fog settings")
			root.queue_free()
			return
		rig.sync(level, factory, "performance")
		if env.fog_enabled or not is_equal_approx(env.fog_density, float(expected_fog["density"])):
			_fail("performance tier must gate fog off while retaining selected fog parameters")
			root.queue_free()
			return
		rig.sync(0, factory, "balanced")
		var town_fog: Dictionary = RenderLoaderScript.context(RenderLoaderScript.CONTEXT_TOWN)["fog"]
		if env.fog_enabled != bool(town_fog["enabled"]) or env.fog_light_color != Color(str(town_fog["color"])):
			_fail("returning to town must replace dungeon depth fog with the town context")
			root.queue_free()
			return
	root.queue_free()
	_pass("lighting rig switches depth profiles and preserves tier/town behavior")


func _test_torch_profile_reaches_live_nodes_and_depth_switch() -> void:
	var factory := GroundWallFactoryScript.new()
	var root := Node3D.new()
	root.name = "DepthMoodTorchTest"
	get_root().add_child(root)
	var walls := [
		{"id": "test-wall", "source": "perimeter", "kind": "wall", "position": {"x": 20.0, "y": -0.5}, "size": {"x": 20.0, "y": 1.0}}
	]
	var lights = TorchLightsScript.new(root, null, factory, null)
	lights.sync(-1, walls, true)
	var shallow_state: Dictionary = lights.get_debug_state()
	var shallow_palette_id: String = str(lights._palette_id)
	var torch := _live_torch(lights._root)
	if torch == null:
		_fail("the configured wall fixture must produce a torch")
		root.queue_free()
		return
	var flame := torch.get_node("Flame") as MeshInstance3D
	var material := flame.material_override as StandardMaterial3D
	var shallow_torch: Dictionary = DepthMoodLoaderScript.profile_for_palette_id("shallow_cave")["torch"]
	if material == null or material.albedo_color != Color(str(shallow_torch["flame_color"])):
		_fail("shallow flame material must use its data-driven color")
		root.queue_free()
		return
	if not is_equal_approx(material.emission_energy_multiplier, float(shallow_torch["flame_emission_energy"])):
		_fail("shallow flame emission must use its data-driven energy")
		root.queue_free()
		return
	var torch_light := torch.get_node("TorchLight") as OmniLight3D
	if torch_light == null or torch_light.light_color != Color(str(shallow_torch["light_color"])):
		_fail("shallow torch light must use its data-driven color")
		root.queue_free()
		return
	var shallow_light_color := torch_light.light_color
	lights.sync(-5, walls, true)
	await process_frame
	var halls_state: Dictionary = lights.get_debug_state()
	var halls_palette_id: String = str(lights._palette_id)
	if shallow_palette_id != "shallow_cave" or halls_palette_id != "sundered_halls":
		_fail("torch lighting must resolve depth changes from the existing palette identity")
		root.queue_free()
		return
	if shallow_state["count"] != halls_state["count"] or shallow_state["light_radius"] != halls_state["light_radius"]:
		_fail("depth mood must leave torch placement and gameplay reveal radius unchanged")
		root.queue_free()
		return
	var halls_torch := _live_torch(lights._root)
	var halls_light := halls_torch.get_node_or_null("TorchLight") as OmniLight3D if halls_torch != null else null
	if halls_light == null or halls_light.light_color == shallow_light_color:
		_fail("depth transition must refresh the torch presentation profile")
		root.queue_free()
		return
	await process_frame
	lights.sync(0, walls, true)
	if lights.get_debug_state()["active"]:
		_fail("town transition must clear dungeon torch presentation")
		root.queue_free()
		return
	root.queue_free()
	_pass("torch depth accents reach runtime materials/lights and transition cleanly")


func _live_torch(root: Node3D) -> Node3D:
	for child in root.get_children():
		if child is Node3D and not child.is_queued_for_deletion() and child.has_node("Flame"):
			return child as Node3D
	return null


func _pass(message: String) -> void:
	_pass_count += 1
	print("  ok   %s" % message)


func _fail(message: String) -> void:
	_fail_count += 1
	push_error("FAIL: %s" % message)


func _finish() -> void:
	if _fail_count > 0:
		print("[gdtest] FAIL: test_dungeon_depth_mood (%d failed)" % _fail_count)
		quit(1)
		return
	print("[gdtest] PASS: test_dungeon_depth_mood (%d checks)" % _pass_count)
	quit(0)
