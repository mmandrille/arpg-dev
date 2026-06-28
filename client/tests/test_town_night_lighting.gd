extends SceneTree

const DungeonDepthLightingScript := preload("res://scripts/dungeon_depth_lighting.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")
const FogPresentationLoaderScript := preload("res://scripts/fog_presentation_loader.gd")

var _pass_count := 0
var _fail_count := 0

func _initialize() -> void:
	_test_town_night_profile_when_fog_active()
	_test_town_fog_suppression_lowers_ambient()
	_finish()

func _test_town_night_profile_when_fog_active() -> void:
	var factory := GroundWallFactoryScript.new()
	var day: Dictionary = DungeonDepthLightingScript.profile_for_level(0, factory, false)
	var night: Dictionary = DungeonDepthLightingScript.profile_for_level(0, factory, true)
	if day == night:
		_fail("town night profile must differ from daylight profile")
		return
	if float(night.get("ambient_energy", 1.0)) >= float(day.get("ambient_energy", 0.0)):
		_fail("town night ambient must be darker than daylight")
		return
	_pass("town night profile differs from daylight")

func _test_town_fog_suppression_lowers_ambient() -> void:
	FogPresentationLoaderScript.ensure_loaded()
	var factory := GroundWallFactoryScript.new()
	var profile: Dictionary = DungeonDepthLightingScript.apply_for_level(
		0,
		null,
		null,
		factory,
		true,
		FogPresentationLoaderScript.ambient_suppression(),
		true,
	)
	if float(profile.get("ambient_energy", 1.0)) > 0.08:
		_fail("fog-suppressed town ambient must be low")
		return
	_pass("town fog suppression lowers ambient")

func _pass(label: String) -> void:
	_pass_count += 1

func _fail(label: String) -> void:
	_fail_count += 1
	push_error(label)

func _finish() -> void:
	if _fail_count == 0:
		print("[gdtest] PASS: test_town_night_lighting (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(0)
	else:
		print("[gdtest] FAIL: test_town_night_lighting (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)
