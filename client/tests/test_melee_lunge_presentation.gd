# Unit tests for local melee lunge presentation (v301).
# Run via: godot --headless --path client --script res://tests/test_melee_lunge_presentation.gd
extends SceneTree

const AnimationControllerScript := preload("res://scripts/animation_controller.gd")
const MeleeLungePresentationScript := preload("res://scripts/melee_lunge_presentation.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	await _test_melee_attack_lunges_and_recovers()
	await _test_ranged_attack_does_not_lunge()
	await _test_non_attack_clip_does_not_lunge()

	print("[gdtest] PASS: test_melee_lunge_presentation (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_melee_attack_lunges_and_recovers() -> void:
	var fixture := _make_character_fixture()
	await process_frame
	var controller = AnimationControllerScript.new(fixture["animation_player"])
	controller.play_one_shot("attack", "melee")
	var lunge: Dictionary = controller.get_debug_state().get("melee_lunge", {})
	_assert_true("melee lunge active", bool(lunge.get("active", false)))
	_assert_eq("melee lunge count", int(lunge.get("count", 0)), 1)
	_assert_true("melee lunge offset", float(lunge.get("offset_z", 0.0)) > 0.05)
	await create_timer(MeleeLungePresentationScript.RECOVERY_SECONDS + 0.05).timeout
	lunge = controller.get_debug_state().get("melee_lunge", {})
	_assert_false("melee lunge recovers", bool(lunge.get("active", true)))
	_assert_true("melee lunge offset settled", float(lunge.get("offset_length", 1.0)) <= 0.02)
	(fixture["visual"] as Node).free()


func _test_ranged_attack_does_not_lunge() -> void:
	var fixture := _make_character_fixture()
	await process_frame
	var controller = AnimationControllerScript.new(fixture["animation_player"])
	controller.play_one_shot("attack", "ranged")
	var lunge: Dictionary = controller.get_debug_state().get("melee_lunge", {})
	_assert_false("ranged lunge inactive", bool(lunge.get("active", true)))
	_assert_eq("ranged lunge count", int(lunge.get("count", 0)), 0)
	(fixture["visual"] as Node).free()


func _test_non_attack_clip_does_not_lunge() -> void:
	var fixture := _make_character_fixture()
	await process_frame
	var controller = AnimationControllerScript.new(fixture["animation_player"])
	controller.play_one_shot("hit", "melee")
	var lunge: Dictionary = controller.get_debug_state().get("melee_lunge", {})
	_assert_false("hit lunge inactive", bool(lunge.get("active", true)))
	_assert_eq("hit lunge count", int(lunge.get("count", 0)), 0)
	(fixture["visual"] as Node).free()


func _make_character_fixture() -> Dictionary:
	var visual := Node3D.new()
	visual.name = "CharacterVisual"
	var model := Node3D.new()
	model.name = "ModelRoot"
	visual.add_child(model)
	var ap := AnimationPlayer.new()
	visual.add_child(ap)
	var lib := AnimationLibrary.new()
	for clip in ["idle", "walk", "attack", "attack_off_hand", "hit"]:
		var anim := Animation.new()
		anim.length = 0.5
		lib.add_animation(clip, anim)
	ap.add_animation_library("", lib)
	get_root().add_child(visual)
	return {"visual": visual, "model": model, "animation_player": ap}


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _assert_eq(label: String, got: Variant, want: Variant) -> void:
	if got == want:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s (got %s want %s)" % [label, str(got), str(want)])
