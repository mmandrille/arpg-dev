class_name RangerAffinityDamage
extends RefCounted


static func bonus_percent(rule: Dictionary, effective_stat: int, affinities: int, rank: int) -> int:
	if rank <= 0 or affinities <= 0:
		return 0
	var denominator := int(rule.get("stat_points_per_extra_percent", 0))
	var maximum := int(rule.get("max_active_affinities", 0))
	if denominator <= 0 or maximum <= 0:
		return 0
	var count := mini(affinities, maximum)
	var stat := maxi(0, effective_stat)
	var bonus := count * (int(rule.get("base_percent_per_affinity", 0)) + floori(float(stat) / float(denominator)))
	return maxi(0, mini(int(rule.get("max_bonus_percent", 0)), bonus))


static func scale_damage(raw: int, bonus_percent_value: int) -> int:
	return floori(float(raw) * float(100 + bonus_percent_value) / 100.0)
