extends SceneTree

const DungeonDepthLightingScript := preload("res://scripts/dungeon_depth_lighting.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")

var _pass_count := 0
var _fail_count := 0

func _initialize() -> void:
	_test_town_profile_is_warm()
	_test_shallow_and_deep_profiles_differ()
	_test_apply_updates_scene_lights()
	_test_fog_suppression_scales_energy()
	_finish()

func _test_town_profile_is_warm() -> void:
	var factory := GroundWallFactoryScript.new()
	var profile: Dictionary = DungeonDepthLightingScript.profile_for_level(0, factory)
	if float(profile.get("directional_energy", 0.0)) <= 0.0:
		_fail("town directional energy must be positive")
		return
	if str(profile.get("directional_color", "")).is_empty():
		_fail("town directional color must be set")
		return
	_pass("town profile resolves")

func _test_shallow_and_deep_profiles_differ() -> void:
	var factory := GroundWallFactoryScript.new()
	var shallow: Dictionary = DungeonDepthLightingScript.profile_for_level(-1, factory)
	var deep: Dictionary = DungeonDepthLightingScript.profile_for_level(-4, factory)
	if shallow == deep:
		_fail("shallow and deep lighting profiles must differ")
		return
	if str(shallow.get("directional_color", "")) == str(deep.get("directional_color", "")):
		_fail("shallow and deep directional colors must differ")
		return
	_pass("depth lighting profiles differ")

func _test_apply_updates_scene_lights() -> void:
	var factory := GroundWallFactoryScript.new()
	var root := Node3D.new()
	var directional := DirectionalLight3D.new()
	var world_environment := WorldEnvironment.new()
	root.add_child(directional)
	root.add_child(world_environment)

	var deep_profile: Dictionary = DungeonDepthLightingScript.apply_for_level(
		-4,
		directional,
		world_environment,
		factory,
	)
	if not is_equal_approx(directional.light_energy, float(deep_profile.get("directional_energy", -1.0))):
		_fail("directional light energy did not apply")
		return
	if world_environment.environment == null:
		_fail("world environment was not configured")
		return
	if not is_equal_approx(
		world_environment.environment.ambient_light_energy,
		float(deep_profile.get("ambient_energy", -1.0)),
	):
		_fail("ambient light energy did not apply")
		return
	_pass("scene lights apply profile")

func _test_fog_suppression_scales_energy() -> void:
	var factory := GroundWallFactoryScript.new()
	var base: Dictionary = DungeonDepthLightingScript.profile_for_level(-2, factory)
	var suppressed: Dictionary = DungeonDepthLightingScript.apply_fog_suppression(
		base,
		{"directional_scale": 0.5, "ambient_scale": 0.25},
	)
	if float(suppressed.get("directional_energy", 0.0)) >= float(base.get("directional_energy", 0.0)):
		_fail("fog suppression must lower directional energy")
		return
	if float(suppressed.get("ambient_energy", 0.0)) >= float(base.get("ambient_energy", 0.0)):
		_fail("fog suppression must lower ambient energy")
		return
	_pass("fog suppression scales profile")

func _pass(label: String) -> void:
	_pass_count += 1

func _fail(label: String) -> void:
	_fail_count += 1
	push_error(label)

func _finish() -> void:
	if _fail_count == 0:
		print("[gdtest] PASS: test_dungeon_depth_lighting (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(0)
	else:
		print("[gdtest] FAIL: test_dungeon_depth_lighting (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)
