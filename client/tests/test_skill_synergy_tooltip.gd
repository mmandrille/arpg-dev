extends SceneTree

const SkillSynergyTooltipScript := preload("res://scripts/skill_synergy_tooltip.gd")


func _init() -> void:
	SkillRulesLoader.ensure_loaded()
	_test_server_synergy_status()
	_test_rules_fallback()
	print("[gdtest] PASS: test_skill_synergy_tooltip")
	quit()


func _test_server_synergy_status() -> void:
	var progression := {
		"skills": [
			{"skill_id": "piercing_shot", "rank": 3, "max_rank": 10, "can_spend": false},
			{
				"skill_id": "snipe",
				"rank": 1,
				"max_rank": 10,
				"can_spend": true,
				"synergy_status": [{
					"source_skill_id": "piercing_shot",
					"source_name": "Piercing Shot",
					"source_rank": 3,
					"modifier": "damage_percent",
					"bonus_percent": 30,
					"display": "+30% from Piercing Shot (rank 3)",
				}],
			},
		],
	}
	var lines := SkillSynergyTooltipScript.lines_for_skill("snipe", progression)
	_assert_eq(lines.size(), 1, "one synergy line")
	_assert_eq(str((lines[0] as Dictionary).get("text", "")), "+30% from Piercing Shot (rank 3)", "server display")


func _test_rules_fallback() -> void:
	var lines := SkillSynergyTooltipScript.lines_for_skill("snipe", {
		"skills": [
			{"skill_id": "piercing_shot", "rank": 2, "max_rank": 10, "can_spend": false},
			{"skill_id": "snipe", "rank": 1, "max_rank": 10, "can_spend": true},
		],
	})
	_assert_eq(lines.size(), 1, "fallback synergy line")
	var text := str((lines[0] as Dictionary).get("text", ""))
	_assert_true("fallback contains +20%", text.contains("+20%"))
	_assert_true("fallback names piercing shot", text.to_lower().contains("piercing"))


func _assert_eq(got: Variant, want: Variant, label: String) -> void:
	if got != want:
		push_error("%s: got %s want %s" % [label, str(got), str(want)])
		quit(1)


func _assert_true(label: String, value: bool) -> void:
	if not value:
		push_error(label)
		quit(1)
