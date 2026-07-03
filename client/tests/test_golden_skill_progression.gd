# Cross-language golden for skill-point cadence and Magic Bolt rank scaling.
extends SceneTree

const SkillRankScalingScript := preload("res://scripts/skill_rank_scaling.gd")


func _initialize() -> void:
	SkillRankScalingScript.ensure_loaded()
	var shared := ProjectSettings.globalize_path("res://").path_join("../shared")
	var progression_rules := _read(shared.path_join("rules/character_progression.v0.json"))
	var progression_combat_rules := _read(shared.path_join("rules/combat.v0.json"))
	var skill_golden := _read(shared.path_join("golden/skill_points_and_magic_bolt.json"))
	var skill_rules := _read(shared.path_join("rules/skills.v0.json"))

	if int(skill_golden["progression"]["points_per_level"]) != int(progression_rules["points_per_level"]):
		_fail("skill golden points_per_level mismatch")
		return
	if skill_golden["progression"]["skill_points"] != progression_rules["skill_points"]:
		_fail("skill golden skill point cadence mismatch")
		return
	for c in skill_golden["progression"]["level_cases"]:
		var level := int(c["level"])
		var expected_stats := (level - 1) * int(progression_rules["points_per_level"])
		var expected_skills := _skill_points_for_level(level, progression_rules["skill_points"])
		if expected_stats != int(c["expected_unspent_stat_points"]) or expected_skills != int(c["expected_unspent_skill_points"]):
			_fail("skill golden level %d cadence mismatch" % level)
			return
	if int(skill_golden["attack_speed"]["base_attack_interval_ticks"]) != int(progression_combat_rules["base_attack_interval_ticks"]):
		_fail("skill golden base attack interval mismatch")
		return
	if float(skill_golden["attack_speed"]["min_effective_attack_speed"]) != float(progression_combat_rules["min_effective_attack_speed"]):
		_fail("skill golden min attack speed mismatch")
		return
	if float(skill_golden["attack_speed"]["max_effective_attack_speed"]) != float(progression_combat_rules["max_effective_attack_speed"]):
		_fail("skill golden max attack speed mismatch")
		return
	var skill_id := str(skill_golden["skill"]["skill_id"])
	if not skill_rules["skills"].has(skill_id):
		_fail("skill golden references unknown skill %s" % skill_id)
		return
	var magic_bolt: Dictionary = skill_rules["skills"][skill_id]
	if int(skill_golden["skill"]["max_rank"]) != int(magic_bolt["max_rank"]):
		_fail("skill golden max rank mismatch")
		return
	if skill_golden["skill"].get("requirements", {}) != magic_bolt.get("requirements", {}):
		_fail("skill golden requirements mismatch")
		return
	if int(magic_bolt["requirements"]["stats"]["magic"]) != 5:
		_fail("magic bolt magic requirement mismatch")
		return
	if int(magic_bolt["requirements"].get("level_per_rank", 0)) != 1 or int(magic_bolt["requirements"].get("stats_per_rank", {}).get("magic", 0)) != 3:
		_fail("magic bolt per-rank requirement mismatch")
		return
	var cooldown_multiplier := float(magic_bolt["cooldown"]["multiplier"])
	for c in skill_golden["attack_speed"]["cases"]:
		var effective := float(c["dex_attack_speed"]) * float(c["weapon_attack_speed"]) * (1.0 + float(c["item_attack_speed_percent"]) / 100.0)
		effective = clampf(effective, float(progression_combat_rules["min_effective_attack_speed"]), float(progression_combat_rules["max_effective_attack_speed"]))
		if not is_equal_approx(effective, float(c["expected_effective_attack_speed"])):
			_fail("skill golden attack speed case %s effective mismatch" % str(c["name"]))
			return
		var interval := _attack_interval_ticks(progression_combat_rules, effective)
		var cooldown := maxi(1, int(ceil(float(interval) * cooldown_multiplier)))
		if interval != int(c["expected_attack_interval_ticks"]) or cooldown != int(c["expected_magic_bolt_cooldown_ticks"]):
			_fail("skill golden attack speed case %s interval/cooldown got %d/%d want %d/%d" % [str(c["name"]), interval, cooldown, int(c["expected_attack_interval_ticks"]), int(c["expected_magic_bolt_cooldown_ticks"])])
			return
	for c in skill_golden["skill"]["rank_requirement_cases"]:
		var req_rank := int(c["rank"])
		var expected_requirements := _skill_requirements_for_rank(magic_bolt["requirements"], req_rank)
		if int(c["level"]) != int(expected_requirements["level"]) or not _same_int_dictionary(c["stats"], expected_requirements["stats"]):
			_fail("skill golden rank %d requirement mismatch" % req_rank)
			return
	for c in skill_golden["skill"]["rank_cases"]:
		var rank := int(c["rank"])
		if _skill_mana_cost(magic_bolt, rank) != int(c["mana_cost"]):
			_fail("skill golden rank %d mana mismatch" % rank)
			return
		var expected_damage: Dictionary = c["damage"]
		if _skill_damage_min(magic_bolt, rank) != int(expected_damage["min_percent"]) or _skill_damage_max(magic_bolt, rank) != int(expected_damage["max_percent"]):
			_fail("skill golden rank %d damage mismatch" % rank)
			return

	print("[gdtest] PASS: consumed shared/golden/skill_points_and_magic_bolt.json")
	quit(0)


func _read(path: String) -> Dictionary:
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		_fail("cannot open %s" % path)
		return {}
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		_fail("invalid JSON in %s" % path)
		return {}
	return parsed


func _fail(message: String) -> void:
	push_error("[gdtest] FAIL: %s" % message)
	print("[gdtest] FAIL: %s" % message)
	quit(1)


func _skill_points_for_level(level: int, cadence: Dictionary) -> int:
	var first := int(cadence["first_grant_level"])
	var second := int(cadence.get("second_grant_level", 0))
	var every := int(cadence["grant_every_levels"])
	var points := int(cadence["points_per_grant"])
	if level < first or points <= 0:
		return 0
	var grants := 0
	for grant_level in range(1, level + 1):
		if grant_level == first:
			grants += 1
		elif second > 0 and grant_level == second:
			grants += 1
		elif every > 0:
			var min_level := int(cadence.get("grant_every_min_level", 0))
			if min_level <= 0:
				min_level = every
			if grant_level >= min_level and grant_level % every == 0:
				grants += 1
	return grants * points


func _attack_interval_ticks(combat: Dictionary, effective_attack_speed: float) -> int:
	var min_speed := float(combat["min_effective_attack_speed"])
	var max_speed := float(combat["max_effective_attack_speed"])
	var speed := clampf(effective_attack_speed, min_speed, max_speed)
	if speed <= 0.0:
		speed = 1.0
	return maxi(1, int(ceil(float(combat["base_attack_interval_ticks"]) / speed)))


func _skill_mana_cost(skill: Dictionary, rank: int) -> int:
	var mana: Dictionary = skill["cost"]["mana"]
	return SkillRankScalingScript.rank_scaled_int(
		int(mana["base"]),
		int(mana["per_rank"]),
		rank,
		SkillRankScalingScript.progression_mana_curve(),
	)


func _skill_requirements_for_rank(requirements: Dictionary, rank: int) -> Dictionary:
	var rank_offset := maxi(0, rank - 1)
	var stats: Dictionary = {}
	var base_stats: Dictionary = requirements.get("stats", {})
	var stats_per_rank: Dictionary = requirements.get("stats_per_rank", {})
	for stat in ["str", "dex", "vit", "magic"]:
		var required := int(base_stats.get(stat, 0)) + int(stats_per_rank.get(stat, 0)) * rank_offset
		if required > 0:
			stats[stat] = required
	return {
		"level": int(requirements.get("level", 0)) + int(requirements.get("level_per_rank", 0)) * rank_offset,
		"stats": stats,
	}


func _same_int_dictionary(a: Dictionary, b: Dictionary) -> bool:
	if a.keys().size() != b.keys().size():
		return false
	for key in b.keys():
		if not a.has(key) or int(a[key]) != int(b[key]):
			return false
	return true


func _skill_damage_min(skill: Dictionary, rank: int) -> int:
	var damage: Dictionary = skill["damage"]
	return SkillRankScalingScript.rank_scaled_int(
		int(damage["min_base"]),
		int(damage["min_per_rank"]),
		rank,
		SkillRankScalingScript.progression_rank_curve(),
	)


func _skill_damage_max(skill: Dictionary, rank: int) -> int:
	var damage: Dictionary = skill["damage"]
	return SkillRankScalingScript.rank_scaled_int(
		int(damage["max_base"]),
		int(damage["max_per_rank"]),
		rank,
		SkillRankScalingScript.progression_rank_curve(),
	)
