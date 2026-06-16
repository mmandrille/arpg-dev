# Unit tests for generated client audio cues.
extends SceneTree

const ClientAudioControllerScript := preload("res://scripts/client_audio_controller.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_volume_application_clamps()
	_test_semantic_cues_update_debug_state()
	_test_boss_music_state()
	print("[gdtest] PASS: test_client_audio_controller (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_volume_application_clamps() -> void:
	var controller = ClientAudioControllerScript.new()
	get_root().add_child(controller)
	controller.apply_volumes(1.5, -1.0, 0.25)
	_assert_float("master clamps high", controller.master_volume, 1.0)
	_assert_float("music clamps low", controller.music_volume, 0.0)
	_assert_float("sfx applies", controller.sfx_volume, 0.25)
	controller.free()


func _test_semantic_cues_update_debug_state() -> void:
	var controller = ClientAudioControllerScript.new()
	get_root().add_child(controller)
	controller.apply_volumes(0.0, 0.0, 0.0)
	controller.play_movement()
	_assert_eq("movement cue", controller.last_cue, "movement")
	controller.play_attack()
	_assert_eq("attack cue", controller.last_cue, "attack")
	controller.play_skill("heal")
	_assert_eq("heal skill cue", controller.last_cue, "heal")
	controller.play_skill("charge")
	_assert_eq("movement skill cue", controller.last_cue, "movement_skill")
	controller.play_damage(true)
	_assert_eq("player damage cue", controller.last_cue, "player_damage")
	controller.play_damage(false)
	_assert_eq("monster damage cue", controller.last_cue, "monster_damage")
	controller.play_kill(false)
	_assert_eq("monster kill cue", controller.last_cue, "monster_kill")
	_assert_eq("cue count", controller.cue_count, 7)
	controller.free()


func _test_boss_music_state() -> void:
	var controller = ClientAudioControllerScript.new()
	get_root().add_child(controller)
	controller.apply_volumes(0.0, 0.0, 0.0)
	controller.play_boss_phase()
	_assert_eq("boss phase cue", controller.last_cue, "boss_phase")
	_assert_true("boss music active after phase", controller.boss_music_active)
	controller.play_kill(true)
	_assert_eq("boss kill cue", controller.last_cue, "boss_kill")
	controller.stop_boss_music()
	_assert_true("boss music stopped", not controller.boss_music_active)
	controller.free()


func _assert_float(label: String, got: float, expected: float) -> void:
	if absf(got - expected) <= 0.001:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
