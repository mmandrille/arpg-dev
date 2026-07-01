# Unit test for compiled codex loader.
extends SceneTree

const CodexLoaderScript := preload("res://scripts/codex_loader.gd")

var _failures: int = 0


func _initialize() -> void:
	CodexLoaderScript.reset_for_tests()
	CodexLoaderScript.ensure_loaded()
	_assert_true("six chapters loaded", CodexLoaderScript.chapters.size() >= 6)
	_assert_true("concepts chapter present", "concepts" in CodexLoaderScript.chapter_ids())
	var barbarian := CodexLoaderScript.page("class:barbarian")
	_assert_eq("barbarian title", str(barbarian.get("title", "")), "Barbarian")
	var skill := CodexLoaderScript.page("skill:magic_bolt")
	_assert_true("magic bolt page", str(skill.get("title", "")).find("Magic Bolt") >= 0)
	if _failures > 0:
		quit(1)
	print("[gdtest] PASS: test_codex_loader")
	quit(0)


func _assert_true(label: String, value: bool) -> void:
	if not value:
		_failures += 1
		printerr("[gdtest] FAIL %s" % label)


func _assert_eq(label: String, got, want) -> void:
	if got != want:
		_failures += 1
		printerr("[gdtest] FAIL %s got=%s want=%s" % [label, str(got), str(want)])
