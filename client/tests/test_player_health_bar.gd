# Contract test for the player vitals HUD (v517): debug-state keys/values stay stable across the globe re-seat.
extends SceneTree

const PlayerHealthBarScript := preload("res://scripts/player_health_bar.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	await _test_debug_state_contract()
	print("[gdtest] PASS: test_player_health_bar (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_debug_state_contract() -> void:
	var bar = PlayerHealthBarScript.new()
	get_root().add_child(bar)
	await process_frame
	var state: Dictionary = bar.get_debug_state()
	for key in ["character_name", "level", "identity_text", "hp", "max_hp", "mana", "max_mana", "attack_recovery_remaining", "attack_recovery_total", "attack_recovery_fraction"]:
		_assert_true("debug key %s present" % key, state.has(key))
	_assert_eq("default identity", str(state.get("identity_text", "")), "Hero  Lv 1")
	bar.set_identity("  Astra ", 4)
	bar.update_hp(9, 12)
	bar.update_mana(7, 14)
	state = bar.get_debug_state()
	_assert_eq("identity text", str(state.get("identity_text", "")), "Astra  Lv 4")
	_assert_eq("hp", int(state.get("hp", -1)), 9)
	_assert_eq("max hp", int(state.get("max_hp", -1)), 12)
	_assert_eq("mana", int(state.get("mana", -1)), 7)
	_assert_eq("max mana", int(state.get("max_mana", -1)), 14)
	_assert_eq("no recovery at rest", float(state.get("attack_recovery_fraction", -1.0)), 0.0)
	bar.start_attack_recovery(1.0)
	state = bar.get_debug_state()
	_assert_eq("recovery total", float(state.get("attack_recovery_total", -1.0)), 1.0)
	_assert_eq("recovery fraction", float(state.get("attack_recovery_fraction", -1.0)), 1.0)
	bar.set_identity("", 0)
	_assert_eq("blank name falls back to Hero, level floors at 1", str(bar.get_debug_state().get("identity_text", "")), "Hero  Lv 1")
	bar.free()


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
