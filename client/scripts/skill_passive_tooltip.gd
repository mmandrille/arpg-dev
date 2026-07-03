class_name SkillPassiveTooltip
extends RefCounted

const SkillNextRankTooltipScript := preload("res://scripts/skill_next_rank_tooltip.gd")

static func passive_stat_lines(def: Dictionary, rank: int) -> Array:
	var passive_stats: Dictionary = def.get("passive_stats", {}).get("stats", {})
	var lines: Array = []
	for stat in passive_stats.keys():
		var value: Dictionary = passive_stats.get(stat, {})
		var amount := _ranked_stat_value(int(value.get("base", 0)), int(value.get("per_rank", 0)), rank)
		if amount > 0:
			lines.append("%s: +%s" % [passive_stat_label(str(stat)), passive_stat_value_text(str(stat), amount)])
	return lines


static func passive_next_rank_lines(def: Dictionary, current_rank: int, next_rank: int) -> Array:
	var passive_stats: Dictionary = def.get("passive_stats", {}).get("stats", {})
	var lines: Array = []
	for stat in passive_stats.keys():
		var value: Dictionary = passive_stats.get(stat, {})
		var now := _ranked_stat_value(int(value.get("base", 0)), int(value.get("per_rank", 0)), current_rank)
		var next := _ranked_stat_value(int(value.get("base", 0)), int(value.get("per_rank", 0)), next_rank)
		if now != next:
			var line := SkillNextRankTooltipScript.percent_delta(passive_stat_label(str(stat)), now, next)
			if line != "":
				lines.append(line)
	return lines


static func passive_stat_value_text(stat: String, amount: int) -> String:
	match stat:
		"attack_speed_percent", "block_percent", "hit_chance", "crit_chance", "evade_chance", "skill_damage_percent", "skill_cooldown_reduction_percent", "magic_find_percent", "damage_percent", "armor_percent", "max_hp_percent", "max_mana_percent", "health_regen_percent", "mana_regen_percent", "light_radius_percent":
			return "%d%%" % amount
		"health_regen_per_10_seconds", "mana_regen_per_10_seconds":
			return "%.1f/s" % (float(amount) / 10.0)
		_:
			return str(amount)


static func passive_stat_label(stat: String) -> String:
	match stat:
		"damage_min":
			return "Min damage"
		"damage_max":
			return "Max damage"
		"damage_percent":
			return "Damage"
		"max_hp":
			return "Max HP"
		"max_hp_percent":
			return "Max HP"
		"max_mana":
			return "Max mana"
		"max_mana_percent":
			return "Max mana"
		"attack_speed_percent":
			return "Attack speed"
		"armor_percent":
			return "Armor"
		"block_percent":
			return "Block"
		"hit_chance":
			return "Hit chance"
		"crit_chance":
			return "Crit chance"
		"evade_chance":
			return "Evade"
		"health_regen_per_10_seconds":
			return "Health regen"
		"health_regen_percent":
			return "Health regen"
		"mana_regen_per_10_seconds":
			return "Mana regen"
		"mana_regen_percent":
			return "Mana regen"
		"skill_damage_percent":
			return "Skill damage"
		"skill_cooldown_reduction_percent":
			return "Cooldown reduction"
		"skill_mana_cost_reduction":
			return "Mana cost reduction"
		"magic_find_percent":
			return "Magic Find"
		"light_radius":
			return "Light radius"
		"light_radius_percent":
			return "Light radius"
		_:
			return stat.replace("_", " ").capitalize()


static func _ranked_stat_value(base: int, per_rank: int, rank: int) -> int:
	return SkillRankScaling.rank_scaled_int(base, per_rank, rank, SkillRankScaling.progression_rank_curve())
