# Unit test for the compact character HUD bar.
extends SceneTree

const CharacterBarScript := preload("res://scripts/character_bar.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	await _test_resource_wallet_debug_state()
	await _test_material_wallet_window_state()
	print("[gdtest] PASS: test_character_bar (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_resource_wallet_debug_state() -> void:
	var bar = CharacterBarScript.new()
	get_root().add_child(bar)
	await process_frame
	bar.set_resource_wallet({"respec_badge": 2, "empty": 0})
	var state := bar.get_debug_state()
	_assert_true("wallet visible", bool(state.get("wallet_visible", false)))
	_assert_eq("wallet text", str(state.get("wallet_text", "")), "Respec 2")
	_assert_true("wallet tooltip name", str(state.get("wallet_tooltip", "")).contains("Respec Badge x2"))
	_assert_true("wallet tooltip category", str(state.get("wallet_tooltip", "")).contains("Category: Currency"))
	_assert_true("wallet tooltip storage", str(state.get("wallet_tooltip", "")).contains("Stored account-wide"))
	bar.set_resource_wallet({"upgrade_shard": 0})
	state = bar.get_debug_state()
	_assert_false("empty wallet hidden", bool(state.get("wallet_visible", true)))
	_assert_eq("empty wallet details", (state.get("wallet_details", []) as Array).size(), 0)
	bar.free()


func _test_material_wallet_window_state() -> void:
	var bar = CharacterBarScript.new()
	get_root().add_child(bar)
	await process_frame
	bar.open_wallet_window()
	var state := bar.get_debug_state()
	_assert_false("empty wallet window hidden", bool((state.get("wallet_window", {}) as Dictionary).get("visible", true)))
	bar.set_resource_wallet({"respec_badge": 2})
	bar.open_wallet_window()
	state = bar.get_debug_state()
	var window: Dictionary = state.get("wallet_window", {})
	_assert_true("wallet window visible", bool(window.get("visible", false)))
	_assert_eq("wallet window row count", int(window.get("row_count", 0)), 1)
	_assert_true("wallet window row text", str(window.get("text", "")).contains("Respec Badge x2"))
	_assert_true("wallet window storage text", str(window.get("text", "")).contains("Stored account-wide"))
	bar.set_resource_wallet({"respec_badge": 4})
	state = bar.get_debug_state()
	window = state.get("wallet_window", {})
	_assert_true("wallet window refreshed", str(window.get("text", "")).contains("Respec Badge x4"))
	bar.set_resource_wallet({"respec_badge": 0})
	state = bar.get_debug_state()
	window = state.get("wallet_window", {})
	_assert_false("empty update closes wallet window", bool(window.get("visible", true)))
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
