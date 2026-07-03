class_name SkillMechanicTooltip
extends RefCounted

const SkillNextRankTooltipScript := preload("res://scripts/skill_next_rank_tooltip.gd")
const TICK_DURATION_S := 0.1


static func mechanic_lines(def: Dictionary, rank: int) -> Array:
	var effective_rank := maxi(1, rank)
	var lines: Array = []
	var damage: Dictionary = def.get("damage", {})
	var damage_type := str(damage.get("type", ""))
	if damage_type != "":
		lines.append_array(_damage_lines(damage, damage_type, effective_rank))
	var cone: Dictionary = def.get("cone", {})
	if not cone.is_empty():
		lines.append_array(_cone_lines(cone, damage_type == ""))
	var projectile: Dictionary = def.get("projectile", {})
	if not projectile.is_empty() and damage_type == "":
		lines.append_array(_projectile_lines(projectile))
	lines.append_array(_poison_lines(def.get("poison", {}), effective_rank))
	lines.append_array(_bleed_lines(def.get("bleed", {}), effective_rank))
	lines.append_array(_mark_lines(def.get("mark", {}), effective_rank))
	var effects: Array = def.get("effects", [])
	for effect in effects:
		if typeof(effect) == TYPE_DICTIONARY:
			lines.append_array(_effect_detail_lines(effect as Dictionary, effective_rank))
	return lines


static func mechanic_next_rank_lines(def: Dictionary, current_rank: int, next_rank: int) -> Array:
	var lines: Array = []
	var damage: Dictionary = def.get("damage", {})
	var damage_type := str(damage.get("type", ""))
	if damage_type == "rank_linear_range" or damage_type == "weapon_multiplier_range":
		var min_now := _ranked_value(int(damage.get("min_base", 0)), int(damage.get("min_per_rank", 0)), current_rank)
		var max_now := _ranked_value(int(damage.get("max_base", 0)), int(damage.get("max_per_rank", 0)), current_rank)
		var min_next := _ranked_value(int(damage.get("min_base", 0)), int(damage.get("min_per_rank", 0)), next_rank)
		var max_next := _ranked_value(int(damage.get("max_base", 0)), int(damage.get("max_per_rank", 0)), next_rank)
		if damage_type == "weapon_multiplier_range":
			_append_line(lines, SkillNextRankTooltipScript.percent_delta("weapon damage min", min_now, min_next))
			_append_line(lines, SkillNextRankTooltipScript.percent_delta("weapon damage max", max_now, max_next))
		else:
			_append_line(lines, SkillNextRankTooltipScript.value_delta("damage min", min_now, min_next))
			_append_line(lines, SkillNextRankTooltipScript.value_delta("damage max", max_now, max_next))
	lines.append_array(_poison_next_rank_lines(def.get("poison", {}), current_rank, next_rank))
	lines.append_array(_bleed_next_rank_lines(def.get("bleed", {}), current_rank, next_rank))
	lines.append_array(_mark_next_rank_lines(def.get("mark", {}), current_rank, next_rank))
	var effects: Array = def.get("effects", [])
	for effect in effects:
		if typeof(effect) != TYPE_DICTIONARY:
			continue
		var rec := effect as Dictionary
		if not rec.has("percent_base") and not rec.has("percent_per_rank"):
			continue
		var label := _effect_label(rec)
		var now := _ranked_value(int(rec.get("percent_base", 0)), int(rec.get("percent_per_rank", 0)), current_rank)
		var nxt := _ranked_value(int(rec.get("percent_base", 0)), int(rec.get("percent_per_rank", 0)), next_rank)
		_append_line(lines, SkillNextRankTooltipScript.percent_delta(label, now, nxt))
	return lines


static func _damage_lines(damage: Dictionary, damage_type: String, rank: int) -> Array:
	var min_val := _ranked_value(int(damage.get("min_base", 0)), int(damage.get("min_per_rank", 0)), rank)
	var max_val := _ranked_value(int(damage.get("max_base", 0)), int(damage.get("max_per_rank", 0)), rank)
	if min_val <= 0 and max_val <= 0:
		return []
	if damage_type == "weapon_multiplier_range":
		return ["Weapon damage: %d%%-%d%%" % [min_val, max_val]]
	return ["Damage: %d-%d" % [min_val, max_val]]


static func _cone_lines(cone: Dictionary, include_weapon_note: bool) -> Array:
	var range_val := float(cone.get("range", 0))
	var angle := float(cone.get("angle_degrees", 0))
	if range_val <= 0 and angle <= 0:
		return []
	var parts: Array[String] = []
	if angle >= 360:
		parts.append("radial")
	else:
		parts.append("%.0f° cone" % angle)
	if range_val > 0:
		parts.append("range %.1f" % range_val)
	var text := "Melee %s" % ", ".join(parts)
	if include_weapon_note and str(cone.get("damage_source", "")) == "weapon":
		text += " (weapon damage)"
	return [text]


static func _projectile_lines(projectile: Dictionary) -> Array:
	var range_val := float(projectile.get("range", 0))
	if range_val <= 0:
		return []
	return ["Projectile range: %.0f" % range_val]


static func _poison_lines(poison: Dictionary, rank: int) -> Array:
	if poison.is_empty():
		return []
	var lines: Array = []
	var damage_pct := _ranked_value(int(poison.get("damage_percent_base", 0)), int(poison.get("damage_percent_per_rank", 0)), rank)
	if damage_pct > 0:
		lines.append("Poison: %d%% max HP%s" % [damage_pct, _duration_suffix(int(poison.get("duration_ticks", 0)))])
	var mark_bonus := int(poison.get("mark_damage_bonus_percent", 0))
	var mark_ticks := int(poison.get("mark_duration_ticks", 0))
	if mark_bonus > 0 and mark_ticks > 0:
		lines.append("Mark: +%d%% damage%s" % [mark_bonus, _duration_suffix(mark_ticks)])
	return lines


static func _bleed_lines(bleed: Dictionary, rank: int) -> Array:
	if bleed.is_empty():
		return []
	var lines: Array = []
	var damage_pct := _ranked_value(int(bleed.get("damage_percent_base", 0)), int(bleed.get("damage_percent_per_rank", 0)), rank)
	if damage_pct > 0:
		lines.append("Bleed: %d%% max HP%s" % [damage_pct, _duration_suffix(int(bleed.get("duration_ticks", 0)))])
	return lines


static func _mark_lines(mark: Dictionary, rank: int) -> Array:
	if mark.is_empty():
		return []
	var bonus := _ranked_value(int(mark.get("damage_bonus_percent", 0)), int(mark.get("damage_bonus_percent_per_rank", 0)), rank)
	var duration := int(mark.get("duration_ticks", 0))
	if bonus <= 0:
		return []
	return ["Mark: +%d%% damage%s" % [bonus, _duration_suffix(duration)]]


static func _effect_detail_lines(effect: Dictionary, rank: int) -> Array:
	var lines: Array = []
	var label := _effect_label(effect)
	if effect.has("percent_base") or effect.has("percent_per_rank"):
		var percent := _ranked_value(int(effect.get("percent_base", 0)), int(effect.get("percent_per_rank", 0)), rank)
		if percent > 0:
			lines.append("%s: %d%%" % [label, percent])
	var duration := int(effect.get("duration_ticks", 0))
	if duration > 0 and lines.is_empty():
		lines.append("%s%s" % [label, _duration_suffix(duration)])
	elif duration > 0 and not lines.is_empty():
		lines[0] = "%s%s" % [str(lines[0]), _duration_suffix(duration)]
	var radius := float(effect.get("radius", 0))
	var area_range := float(effect.get("range", 0))
	if radius > 0 or area_range > 0:
		var area_parts: Array[String] = []
		if area_range > 0:
			area_parts.append("range %.0f" % area_range)
		if radius > 0:
			area_parts.append("radius %.0f" % radius)
		lines.append("Area: %s" % ", ".join(area_parts))
	return lines


static func _poison_next_rank_lines(poison: Dictionary, current_rank: int, next_rank: int) -> Array:
	if poison.is_empty():
		return []
	return _percent_next_line(
		"Poison",
		int(poison.get("damage_percent_base", 0)),
		int(poison.get("damage_percent_per_rank", 0)),
		current_rank,
		next_rank,
		true
	)


static func _bleed_next_rank_lines(bleed: Dictionary, current_rank: int, next_rank: int) -> Array:
	if bleed.is_empty():
		return []
	return _percent_next_line(
		"Bleed",
		int(bleed.get("damage_percent_base", 0)),
		int(bleed.get("damage_percent_per_rank", 0)),
		current_rank,
		next_rank,
		true
	)


static func _mark_next_rank_lines(mark: Dictionary, current_rank: int, next_rank: int) -> Array:
	if mark.is_empty():
		return []
	return _percent_next_line(
		"Mark bonus",
		int(mark.get("damage_bonus_percent", 0)),
		int(mark.get("damage_bonus_percent_per_rank", 0)),
		current_rank,
		next_rank,
		true
	)


static func _percent_next_line(label: String, base: int, per_rank: int, current_rank: int, next_rank: int, _as_percent: bool) -> Array:
	var now := _ranked_value(base, per_rank, current_rank)
	var nxt := _ranked_value(base, per_rank, next_rank)
	var line := SkillNextRankTooltipScript.percent_delta(label, now, nxt)
	if line == "":
		return []
	return [line]


static func _append_line(lines: Array, line: String) -> void:
	if line != "":
		lines.append(line)


static func _effect_label(effect: Dictionary) -> String:
	match str(effect.get("type", "")):
		"area_percent_heal":
			return "Heal"
		"stat_percent_buff", "area_stat_percent_buff":
			return "Buff"
		"area_immunity_buff":
			return "Immunity"
		"revive":
			return "Revive"
		_:
			return "Effect"


static func _duration_suffix(ticks: int) -> String:
	if ticks <= 0:
		return ""
	var seconds := float(ticks) * TICK_DURATION_S
	if is_equal_approx(seconds, roundf(seconds)):
		return " for %ds" % int(roundf(seconds))
	return " for %.1fs" % seconds


static func _ranked_value(base: int, per_rank: int, rank: int) -> int:
	return SkillRankScaling.rank_scaled_int(base, per_rank, rank, SkillRankScaling.progression_rank_curve())
