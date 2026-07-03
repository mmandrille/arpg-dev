extends SceneTree

const SkillMechanicTooltipScript := preload("res://scripts/skill_mechanic_tooltip.gd")
const SkillRulesLoaderScript := preload("res://scripts/skill_rules_loader.gd")
const SkillRankScalingScript := preload("res://scripts/skill_rank_scaling.gd")
const SkillsPanelScript := preload("res://scripts/skills_panel.gd")


func _init() -> void:
	SkillRulesLoaderScript.ensure_loaded()
	_test_eviscerate_mechanics()
	_test_eviscerate_synergy_in_rich_tooltip()
	print("[gdtest] PASS: test_skill_mechanic_tooltip")
	quit()


func _test_eviscerate_mechanics() -> void:
	var def := SkillRulesLoaderScript.skill_definition("eviscerate")
	var lines := SkillMechanicTooltipScript.mechanic_lines(def, 1)
	_assert_true("cone line present", _lines_contain(lines, "Melee"))
	_assert_true("poison line present", _lines_contain(lines, "Poison: 20%"))
	_assert_true("mark line present", _lines_contain(lines, "Mark: +30%"))
	var next_lines := SkillMechanicTooltipScript.mechanic_next_rank_lines(def, 1, 2)
	var rank2_poison := SkillRankScalingScript.rank_scaled_int(20, 10, 2, SkillRankScalingScript.progression_rank_curve())
	_assert_true("poison scales next rank", _lines_contain(next_lines, "next rank: +%d%% poison" % (rank2_poison - 20)))


func _test_eviscerate_synergy_in_rich_tooltip() -> void:
	var progression := {
		"skills": [
			{"skill_id": "poison_stab", "rank": 1, "max_rank": 10},
			{"skill_id": "eviscerate", "rank": 1, "max_rank": 10},
		],
	}
	var body := SkillsPanelScript.tooltip_plain_body("eviscerate", 1, progression, {"level": 6, "base_stats": {"dex": 14}})
	_assert_true("plain tooltip includes synergies", body.contains("Synergies:"))
	_assert_true("plain tooltip includes poison stab bonus", body.contains("Poison Stab"))
	var panel := SkillsPanelScript.new()
	panel.skill_progression = progression
	panel.character_progression = {"level": 6, "base_stats": {"dex": 14}}
	var rich := panel._tooltip_rich_text_for("eviscerate", 1)
	_assert_true("rich tooltip includes synergies", rich.contains("Synergies:"))
	_assert_true("rich tooltip includes poison stab bonus", rich.contains("Poison Stab"))
	panel.free()


func _lines_contain(lines: Array, needle: String) -> bool:
	for line in lines:
		if str(line).contains(needle):
			return true
	return false


func _assert_true(label: String, value: bool) -> void:
	if not value:
		push_error(label)
		quit(1)
