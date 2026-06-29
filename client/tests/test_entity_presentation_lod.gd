# Headless unit tests for EntityPresentationLod (v373).
extends SceneTree

const EntityPresentationLodScript := preload("res://scripts/entity_presentation_lod.gd")
const MainConfigLoaderScript := preload("res://scripts/main_config_loader.gd")

var _pass := 0
var _fail := 0


func _init() -> void:
	MainConfigLoaderScript.ensure_loaded()
	EntityPresentationLodScript.reset_for_tests()
	_test_should_apply_below_threshold()
	_test_should_apply_at_threshold()
	_finish()


func _test_should_apply_below_threshold() -> void:
	var cfg := MainConfigLoaderScript.presentation_lod_rules()
	var min_live := int(cfg.get("min_live_monsters", 24))
	_assert_false("below threshold skips LOD", EntityPresentationLodScript.should_apply(min_live - 1))


func _test_should_apply_at_threshold() -> void:
	var cfg := MainConfigLoaderScript.presentation_lod_rules()
	var min_live := int(cfg.get("min_live_monsters", 24))
	_assert_true("at threshold enables LOD", EntityPresentationLodScript.should_apply(min_live))


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass += 1
	else:
		_fail += 1
		push_error("[gdtest] FAIL %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _finish() -> void:
	if _fail == 0:
		print("[gdtest] PASS: test_entity_presentation_lod (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_entity_presentation_lod (%d failures)" % _fail)
		quit(1)
