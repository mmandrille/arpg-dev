class_name SkillSynergyTooltip
extends RefCounted

const SYNERGY_COLOR := Color("#c9b0ff")


static func lines_for_skill(skill_id: String, skill_progression: Dictionary) -> Array:
	var statuses := _synergy_status_for_skill(skill_id, skill_progression)
	if statuses.is_empty():
		statuses = _synergy_status_from_rules(skill_id, skill_progression)
	if statuses.is_empty():
		return []
	var lines: Array = []
	for status in statuses:
		if typeof(status) != TYPE_DICTIONARY:
			continue
		var rec := status as Dictionary
		var display := str(rec.get("display", "")).strip_edges()
		if display == "":
			display = _fallback_display(rec)
		if display == "":
			continue
		lines.append({
			"text": display,
			"color": SYNERGY_COLOR,
		})
	return lines


static func _synergy_status_for_skill(skill_id: String, skill_progression: Dictionary) -> Array:
	var skills = skill_progression.get("skills", [])
	if typeof(skills) != TYPE_ARRAY:
		return []
	for row in skills:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if str(rec.get("skill_id", "")) != skill_id:
			continue
		var statuses = rec.get("synergy_status", [])
		if typeof(statuses) == TYPE_ARRAY:
			return statuses
		return []
	return []


static func _synergy_status_from_rules(skill_id: String, skill_progression: Dictionary) -> Array:
	SkillRulesLoader.ensure_loaded()
	var def := SkillRulesLoader.skill_definition(skill_id)
	var synergies = def.get("synergies", [])
	if typeof(synergies) != TYPE_ARRAY or synergies.is_empty():
		return []
	var out: Array = []
	for synergy in synergies:
		if typeof(synergy) != TYPE_DICTIONARY:
			continue
		var rec := synergy as Dictionary
		var source_id := str(rec.get("source_skill_id", ""))
		if source_id == "":
			continue
		var source_rank := _allocated_rank(source_id, skill_progression)
		var per_rank := float(rec.get("percent_per_source_rank", 0))
		var bonus := int(round(float(source_rank) * per_rank))
		var source_name := SkillRulesLoader.skill_display_name(source_id)
		var display := ""
		if source_rank > 0:
			display = "+%d%% from %s (rank %d)" % [bonus, source_name, source_rank]
		elif per_rank > 0:
			display = "+%.0f%% per %s rank" % [per_rank, source_name]
		out.append({
			"source_skill_id": source_id,
			"source_name": source_name,
			"source_rank": source_rank,
			"modifier": str(rec.get("modifier", "")),
			"bonus_percent": bonus,
			"display": display,
		})
	return out


static func _allocated_rank(skill_id: String, skill_progression: Dictionary) -> int:
	var skills = skill_progression.get("skills", [])
	if typeof(skills) != TYPE_ARRAY:
		return 0
	for row in skills:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		if str(rec.get("skill_id", "")) != skill_id:
			continue
		return maxi(int(rec.get("rank", 0)), 0)
	return 0


static func _fallback_display(rec: Dictionary) -> String:
	var bonus := int(rec.get("bonus_percent", 0))
	var source_name := str(rec.get("source_name", rec.get("source_skill_id", "")))
	var source_rank := int(rec.get("source_rank", 0))
	if source_name == "":
		return ""
	return "+%d%% from %s (rank %d)" % [bonus, source_name, source_rank]
