# Unit test for codex panel navigation.
extends SceneTree

const CodexPanelScript := preload("res://scripts/codex_panel.gd")
const CodexLoaderScript := preload("res://scripts/codex_loader.gd")

var _failures: int = 0


func _initialize() -> void:
	CodexLoaderScript.ensure_loaded()
	var panel: CodexPanel = CodexPanelScript.new()
	get_root().add_child(panel)
	panel.ensure_built()
	panel.show_codex()
	var state := panel.get_debug_state()
	_assert_true("panel visible", bool(state.get("visible", false)))
	_assert_true("chapter selected", str(state.get("chapter_id", "")) != "")
	panel.select_page("class:sorcerer")
	state = panel.get_debug_state()
	_assert_eq("sorcerer page", str(state.get("page_id", "")), "class:sorcerer")
	_assert_true("sorcerer title", str(state.get("page_title", "")).find("Sorcerer") >= 0)
	panel.hide_panel()
	_assert_true("panel hidden", not bool(panel.get_debug_state().get("visible", true)))
	panel.free()
	if _failures > 0:
		quit(1)
	print("[gdtest] PASS: test_codex_panel")
	quit(0)


func _assert_true(label: String, value: bool) -> void:
	if not value:
		_failures += 1
		printerr("[gdtest] FAIL %s" % label)


func _assert_eq(label: String, got, want) -> void:
	if got != want:
		_failures += 1
		printerr("[gdtest] FAIL %s got=%s want=%s" % [label, str(got), str(want)])
