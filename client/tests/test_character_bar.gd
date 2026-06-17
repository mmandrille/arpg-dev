# Unit test for the compact character HUD bar.
extends SceneTree

const CharacterBarScript := preload("res://scripts/character_bar.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_test_resource_wallet_debug_state()
	print("[gdtest] PASS: test_character_bar (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_resource_wallet_debug_state() -> void:
	var bar = CharacterBarScript.new()
	get_root().add_child(bar)
	await process_frame
	bar.set_resource_wallet({"upgrade_shard": 2, "empty": 0})
	var state := bar.get_debug_state()
	_assert_true("wallet visible", bool(state.get("wallet_visible", false)))
	_assert_eq("wallet text", str(state.get("wallet_text", "")), "Shard 2")
	bar.set_resource_wallet({"upgrade_shard": 0})
	state = bar.get_debug_state()
	_assert_false("empty wallet hidden", bool(state.get("wallet_visible", true)))
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


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)
