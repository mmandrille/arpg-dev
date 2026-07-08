# Unit test for the compact character HUD bar.
extends SceneTree

const CharacterBarScript := preload("res://scripts/character_bar.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	await _test_character_slot_debug_state()
	print("[gdtest] PASS: test_character_bar (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_character_slot_debug_state() -> void:
	var bar = CharacterBarScript.new()
	get_root().add_child(bar)
	await process_frame
	bar.set_progression({"unspent_stat_points": 2})
	var state := bar.get_debug_state()
	_assert_eq("character slot label", str(state.get("slot_text", "")), "C")
	_assert_eq("character tooltip", str(state.get("tooltip_text", "")), "Character")
	_assert_true("stat badge visible with points", bool(state.get("upgrade_badge_visible", false)))
	bar.start_attack_recovery(1.0)
	state = bar.get_debug_state()
	_assert_true("attack recovery overlay visible", bool(state.get("cooldown_overlay_visible", false)))
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
