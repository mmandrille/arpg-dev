extends SceneTree

const RangerAffinityDamageScript := preload("res://scripts/ranger_affinity_damage.gd")
const SkillPassiveTooltipScript := preload("res://scripts/skill_passive_tooltip.gd")


func _initialize() -> void:
	var shared := ProjectSettings.globalize_path("res://").path_join("../shared")
	var golden := _read(shared.path_join("golden/ranger_affinity_damage.json"))
	var skills := _read(shared.path_join("rules/skills.v0.json"))
	var rule: Dictionary = skills["skills"]["deadeye"]["passive_stats"]["ranger_affinity_damage"]
	if rule != golden["rule"]:
		_fail("Deadeye rule differs from shared golden")
		return
	for case in golden["cases"]:
		var cfg: Dictionary = case.get("rule_override", rule)
		var percent := RangerAffinityDamageScript.bonus_percent(cfg, int(case["effective_stat"]), int(case["affinities"]), int(case["rank"]))
		if percent != int(case["expected_percent"]):
			_fail("%s bonus %d != %d" % [case["name"], percent, int(case["expected_percent"])])
			return
		if RangerAffinityDamageScript.scale_damage(int(case["raw_min"]), percent) != int(case["expected_min"]) or RangerAffinityDamageScript.scale_damage(int(case["raw_max"]), percent) != int(case["expected_max"]):
			_fail("%s scaled range mismatch" % case["name"])
			return
	var lines := SkillPassiveTooltipScript.ranger_affinity_lines(skills["skills"]["deadeye"], {"derived_stats": {"ranged_damage_bonus_percent": 6}})
	if lines.size() != 2 or not str(lines[0]).contains("Dexterity") or not str(lines[1]).contains("+6%"):
		_fail("Deadeye tooltip omits rule or authoritative current bonus")
		return
	print("[gdtest] PASS: test_ranger_affinity_damage")
	quit(0)


func _read(path: String) -> Dictionary:
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		return {}
	return JSON.parse_string(file.get_as_text()) as Dictionary


func _fail(message: String) -> void:
	push_error(message)
	quit(1)
