# Unit tests for attack animation speed scaling (v362).
extends SceneTree

const AttackAnimationScalingScript := preload("res://scripts/attack_animation_scaling.gd")
const CombatFeelPresentationLoaderScript := preload("res://scripts/combat_feel_presentation_loader.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	CombatFeelPresentationLoaderScript.reset_for_tests()
	_assert_approx("baseline speed", AttackAnimationScalingScript.speed_scale_for(1.0), 1.0, 0.001)
	_assert_approx("faster attacks", AttackAnimationScalingScript.speed_scale_for(2.0), 2.0, 0.001)
	_assert_true("slow attacks clamped", AttackAnimationScalingScript.speed_scale_for(0.1) >= 0.75)

	print("[gdtest] PASS: test_attack_animation_scaling (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s" % label)


func _assert_approx(label: String, got: float, want: float, tolerance: float) -> void:
	if absf(got - want) <= tolerance:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s (got %s want %s)" % [label, str(got), str(want)])
