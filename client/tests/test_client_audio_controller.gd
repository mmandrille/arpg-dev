# Unit tests for generated client audio cues.
extends SceneTree

const ClientAudioControllerScript := preload("res://scripts/client_audio_controller.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_volume_application_clamps()
	_test_semantic_cues_update_debug_state()
	_test_ambience_zone_state()
	_test_boss_music_state()
	_test_boss_phase_cue_classification()
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
	controller.play_skill("magic_bolt")
	_assert_eq("projectile skill cue", controller.last_cue, "skill_projectile")
	controller.play_skill("rage")
	_assert_eq("buff skill cue", controller.last_cue, "skill_buff")
	_assert_eq("skill id debug", controller.last_skill_id, "rage")
	controller.play_damage(true)
	_assert_eq("player damage cue", controller.last_cue, "player_damage")
	controller.play_damage(false)
	_assert_eq("monster damage cue", controller.last_cue, "monster_damage")
	controller.play_kill(false)
	_assert_eq("monster kill cue", controller.last_cue, "monster_kill")
	_assert_eq("cue count", controller.cue_count, 9)
	controller.free()


func _test_boss_music_state() -> void:
	var controller = ClientAudioControllerScript.new()
	get_root().add_child(controller)
	controller.apply_volumes(1.0, 1.0, 0.0)
	controller.set_ambient_level(-1)
	_assert_eq("dungeon ambience zone", controller.ambient_zone, "dungeon")
	_assert_true("dungeon ambience active", controller.ambient_active)
	_assert_true("dungeon ambience silent", not controller._music_player.playing)
	controller.play_boss_phase()
	_assert_eq("boss phase cue", controller.last_cue, "boss_phase")
	_assert_true("boss music active after phase", controller.boss_music_active)
	_assert_true("ambience paused during boss", not controller.ambient_active)
	_assert_true("boss music silent", not controller._music_player.playing)
	controller.play_kill(true)
	_assert_eq("boss kill cue", controller.last_cue, "boss_kill")
	controller.stop_boss_music()
	_assert_true("boss music stopped", not controller.boss_music_active)
	_assert_eq("boss layer reset", controller.boss_music_layer, "none")
	_assert_float("boss intensity reset", controller.boss_music_intensity, 0.0)
	_assert_true("ambience resumes after boss", controller.ambient_active)
	controller.free()


func _test_ambience_zone_state() -> void:
	var controller = ClientAudioControllerScript.new()
	get_root().add_child(controller)
	controller.apply_volumes(0.0, 0.0, 0.0)
	controller.set_ambient_level(0)
	_assert_eq("town ambience zone", controller.ambient_zone, "town")
	_assert_true("town ambience active while muted", controller.ambient_active)
	controller.set_ambient_level(-2)
	_assert_eq("dungeon ambience zone switch", controller.ambient_zone, "dungeon")
	_assert_true("dungeon ambience active while muted", controller.ambient_active)
	var state := controller.get_debug_state()
	_assert_eq("debug ambience zone", str(state.get("ambient_zone", "")), "dungeon")
	_assert_true("debug ambience active", bool(state.get("ambient_active", false)))
	controller.free()


func _test_boss_phase_cue_classification() -> void:
	var controller = ClientAudioControllerScript.new()
	get_root().add_child(controller)
	controller.apply_volumes(0.0, 0.0, 0.0)
	controller.play_boss_phase("charged_melee", "telegraph")
	_assert_eq("boss telegraph cue", controller.last_cue, "boss_telegraph")
	_assert_eq("boss telegraph layer", controller.boss_music_layer, "windup")
	_assert_float("boss telegraph intensity", controller.boss_music_intensity, 0.64)
	controller.play_boss_phase("ground_slam", "active")
	_assert_eq("boss active cue", controller.last_cue, "boss_active")
	_assert_eq("boss active layer", controller.boss_music_layer, "danger")
	_assert_float("boss active intensity", controller.boss_music_intensity, 1.0)
	controller.play_boss_phase("ground_slam", "recovery")
	_assert_eq("boss recovery cue", controller.last_cue, "boss_recovery")
	_assert_eq("boss recovery layer", controller.boss_music_layer, "release")
	_assert_float("boss recovery intensity", controller.boss_music_intensity, 0.34)
	controller.play_boss_phase("summon_wolves", "active")
	_assert_eq("boss summon cue", controller.last_cue, "boss_summon")
	_assert_eq("boss summon layer", controller.boss_music_layer, "summon")
	controller.play_boss_phase("stone_lance", "telegraph")
	_assert_eq("boss ranged cue", controller.last_cue, "boss_ranged")
	_assert_eq("boss ranged layer", controller.boss_music_layer, "ranged")
	_assert_eq("boss pattern debug", controller.last_boss_pattern_id, "stone_lance")
	_assert_eq("boss phase debug", controller.last_boss_phase_kind, "telegraph")
	_assert_true("boss music active while muted", controller.boss_music_active)
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
