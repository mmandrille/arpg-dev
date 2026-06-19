# Unit test for badge resources in the material wallet.
extends SceneTree

const CharacterBarScript := preload("res://scripts/character_bar.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var bar = CharacterBarScript.new()
	get_root().add_child(bar)
	await process_frame
	bar.set_resource_wallet({
		"upgrade_shard": 2,
		"respec_badge": 1,
		"skill_badge": 3,
	})
	var state := bar.get_debug_state()
	_assert_true("wallet visible", bool(state.get("wallet_visible", false)))
	var compact := str(state.get("wallet_text", ""))
	_assert_true("compact upgrade badge", compact.contains("Upgrade 2"))
	_assert_true("compact respec badge", compact.contains("Respec 1"))
	_assert_true("compact skill badge", compact.contains("Skill 3"))
	var tooltip := str(state.get("wallet_tooltip", ""))
	_assert_true("tooltip upgrade badge", tooltip.contains("Upgrade Badge x2"))
	_assert_true("tooltip respec badge", tooltip.contains("Respec Badge x1"))
	_assert_true("tooltip skill badge", tooltip.contains("Skill Badge x3"))
	bar.open_wallet_window()
	state = bar.get_debug_state()
	var window: Dictionary = state.get("wallet_window", {})
	_assert_true("wallet window visible", bool(window.get("visible", false)))
	_assert_eq("wallet window rows", int(window.get("row_count", 0)), 3)
	var text := str(window.get("text", ""))
	_assert_true("window upgrade badge", text.contains("Upgrade Badge x2"))
	_assert_true("window respec badge", text.contains("Respec Badge x1"))
	_assert_true("window skill badge", text.contains("Skill Badge x3"))
	bar.free()
	print("[gdtest] PASS: test_material_wallet_badges (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


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
