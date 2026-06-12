# Unit test for catalog-backed text lookup.
# Run via: godot --headless --path client --script res://tests/test_text_catalog.gd
extends SceneTree

const TextCatalogScript := preload("res://scripts/text_catalog.gd")
const SkillRulesLoaderScript := preload("res://scripts/skill_rules_loader.gd")
const StatLabelsScript := preload("res://scripts/stat_labels.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	TextCatalogScript.reset_for_tests()
	SkillRulesLoaderScript.reset_for_tests()

	_assert_eq("menu label from English catalog", TextCatalogScript.get_text("menu.create_game", "fallback"), "Create Game")
	_assert_eq("missing key uses supplied fallback", TextCatalogScript.get_text("missing.key", "Fallback"), "Fallback")
	_assert_eq("missing key without fallback returns key", TextCatalogScript.get_text("missing.key"), "missing.key")
	_assert_eq("stat label from English catalog", StatLabelsScript.display_name("armor"), "Armor")
	_assert_eq("skill name from text key", SkillRulesLoaderScript.skill_display_name("magic_bolt"), "Magic Bolt")
	_assert_eq("skill summary from text key", SkillRulesLoaderScript.skill_summary("rage"), "STR/VIT size buff")

	print("[gdtest] PASS: test_text_catalog (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])
