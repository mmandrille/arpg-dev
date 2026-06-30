# Unit tests for combat feel presentation catalog loader (v359).
extends SceneTree

const CombatFeelPresentationLoaderScript := preload("res://scripts/combat_feel_presentation_loader.gd")
const CombatFeelConfigScript := preload("res://scripts/combat_feel_config.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	CombatFeelPresentationLoaderScript.reset_for_tests()
	CombatFeelConfigScript.reset_for_tests()
	_test_catalog_loads_positive_values()
	_test_config_facade_matches_loader()

	print("[gdtest] PASS: test_combat_feel_presentation_loader (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_catalog_loads_positive_values() -> void:
	CombatFeelPresentationLoaderScript.ensure_loaded()
	_assert_true("input buffer positive", CombatFeelPresentationLoaderScript.input_buffer_seconds() > 0.0)
	_assert_true("movement smoothing has catch_up_speed", CombatFeelPresentationLoaderScript.movement_smoothing().has("catch_up_speed"))
	_assert_false("enemy impact disabled by default", CombatFeelPresentationLoaderScript.enemy_impact_feedback_enabled())


func _test_config_facade_matches_loader() -> void:
	CombatFeelConfigScript.ensure_loaded()
	_assert_approx("attack buffer", CombatFeelConfigScript.attack_buffer_seconds(), CombatFeelPresentationLoaderScript.input_buffer_seconds(), 0.0001)
	_assert_approx("lunge distance", CombatFeelConfigScript.melee_lunge_distance(), float(CombatFeelPresentationLoaderScript.melee_lunge().get("distance", 0.0)), 0.0001)


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _assert_approx(label: String, got: float, want: float, tolerance: float) -> void:
	if absf(got - want) <= tolerance:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s (got %s want %s)" % [label, str(got), str(want)])
