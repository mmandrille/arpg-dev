extends SceneTree
## ADR-0018 D7 render baseline: catalog -> Environment / key light / viewport AA.
## Expectations are derived from the loaded catalog, never duplicated tuning literals.

const LoaderScript := preload("res://scripts/render_presentation_loader.gd")
const PresentationScript := preload("res://scripts/render_environment_presentation.gd")
const SceneLightingRigScript := preload("res://scripts/scene_lighting_rig.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_catalog_has_every_context_and_tier()
	_test_context_for_level()
	_test_balanced_applies_context_values()
	_test_performance_tier_gates_expensive_features()
	_test_rig_configures_key_light_and_environment()
	_finish()


func _test_catalog_has_every_context_and_tier() -> void:
	for context_id in [LoaderScript.CONTEXT_TOWN, LoaderScript.CONTEXT_TOWN_NIGHT, LoaderScript.CONTEXT_DUNGEON]:
		if LoaderScript.context(context_id).is_empty():
			_fail("missing render context %s" % context_id)
			return
	for quality in [ClientSettingsScript.GRAPHICS_QUALITY_BALANCED, ClientSettingsScript.GRAPHICS_QUALITY_PERFORMANCE]:
		if LoaderScript.quality_tier(quality).is_empty():
			_fail("missing quality tier %s" % quality)
			return
	if LoaderScript.quality_tier("unknown-tier") != LoaderScript.quality_tier(ClientSettingsScript.GRAPHICS_QUALITY_BALANCED):
		_fail("unknown quality tier must fall back to balanced")
		return
	_pass("catalog contexts and tiers resolve")


func _test_context_for_level() -> void:
	if LoaderScript.context_for_level(-3) != LoaderScript.CONTEXT_DUNGEON:
		_fail("negative levels must use the dungeon context")
		return
	if LoaderScript.context_for_level(0) != LoaderScript.CONTEXT_TOWN:
		_fail("town level without fog must use the town context")
		return
	if LoaderScript.context_for_level(0, true) != LoaderScript.CONTEXT_TOWN_NIGHT:
		_fail("town level with fog must use the town_night context")
		return
	_pass("context_for_level mirrors depth lighting")


func _apply(context_id: String, quality: String) -> Dictionary:
	var light := DirectionalLight3D.new()
	var world_environment := WorldEnvironment.new()
	PresentationScript.apply(
		LoaderScript.context(context_id), LoaderScript.quality_tier(quality), LoaderScript.key_light(), light, world_environment
	)
	return {"light": light, "env": world_environment.environment, "world_environment": world_environment}


func _test_balanced_applies_context_values() -> void:
	var quality := ClientSettingsScript.GRAPHICS_QUALITY_BALANCED
	var ctx := LoaderScript.context(LoaderScript.CONTEXT_DUNGEON)
	var tier := LoaderScript.quality_tier(quality)
	var applied := _apply(LoaderScript.CONTEXT_DUNGEON, quality)
	var env: Environment = applied["env"]
	if env == null:
		_fail("apply must create an Environment when missing")
		return
	var tonemap: Dictionary = ctx["tonemap"]
	if env.tonemap_mode != PresentationScript.TONEMAP_MODES[str(tonemap["mode"])]:
		_fail("tonemap mode did not follow catalog")
		return
	if not is_equal_approx(env.tonemap_exposure, float(tonemap["exposure"])):
		_fail("tonemap exposure did not follow catalog")
		return
	var ssao: Dictionary = ctx["ssao"]
	if env.ssao_enabled != (bool(tier["ssao"]) and bool(ssao["enabled"])):
		_fail("SSAO enable must be context AND tier")
		return
	var fog: Dictionary = ctx["fog"]
	if env.fog_enabled != (bool(tier["fog"]) and bool(fog["enabled"])):
		_fail("fog enable must be context AND tier")
		return
	if fog.has("density") and not is_equal_approx(env.fog_density, float(fog["density"])):
		_fail("fog density did not follow catalog")
		return
	var light: DirectionalLight3D = applied["light"]
	if light.shadow_enabled != (bool(LoaderScript.key_light().get("shadow_enabled", false)) and bool(tier["key_light_shadows"])):
		_fail("key-light shadows must be catalog AND tier")
		return
	(applied["light"] as Node).free()
	(applied["world_environment"] as Node).free()
	_pass("balanced tier applies dungeon context values")


func _test_performance_tier_gates_expensive_features() -> void:
	var quality := ClientSettingsScript.GRAPHICS_QUALITY_PERFORMANCE
	var tier := LoaderScript.quality_tier(quality)
	var applied := _apply(LoaderScript.CONTEXT_DUNGEON, quality)
	var env: Environment = applied["env"]
	var light: DirectionalLight3D = applied["light"]
	if not bool(tier["ssao"]) and env.ssao_enabled:
		_fail("performance tier must be able to disable SSAO")
		return
	if not bool(tier["fog"]) and env.fog_enabled:
		_fail("performance tier must be able to disable fog")
		return
	if not bool(tier["key_light_shadows"]) and light.shadow_enabled:
		_fail("performance tier must be able to disable key-light shadows")
		return
	light.free()
	(applied["world_environment"] as Node).free()
	_pass("performance tier gates SSAO / fog / shadows")


func _test_rig_configures_key_light_and_environment() -> void:
	var root := Node3D.new()
	var rig = SceneLightingRigScript.new()
	rig.attach(root)
	var profile: Dictionary = rig.sync(-2, GroundWallFactoryScript.new(), ClientSettingsScript.DEFAULT_GRAPHICS_QUALITY)
	var rot: Dictionary = LoaderScript.key_light()["rotation_degrees"]
	var expected := Vector3(float(rot["x"]), float(rot["y"]), float(rot["z"]))
	if not rig.directional.rotation_degrees.is_equal_approx(expected):
		_fail("rig key light rotation must follow catalog")
		root.free()
		return
	if rig.world_environment.environment == null:
		_fail("rig must configure an Environment")
		root.free()
		return
	if not is_equal_approx(rig.directional.light_energy, float(profile.get("directional_energy", -1.0))):
		_fail("rig must keep DungeonDepthLighting energy")
		root.free()
		return
	root.free()
	_pass("scene lighting rig wires depth lighting + render baseline")


func _pass(msg: String) -> void:
	_pass_count += 1
	print("  ok   %s" % msg)


func _fail(msg: String) -> void:
	_fail_count += 1
	push_error("FAIL: %s" % msg)
	print("  FAIL %s" % msg)


func _finish() -> void:
	if _fail_count > 0:
		print("[gdtest] FAIL: test_render_environment_presentation (%d failed)" % _fail_count)
		quit(1)
		return
	print("[gdtest] PASS: test_render_environment_presentation (%d checks)" % _pass_count)
	quit(0)
