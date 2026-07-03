class_name SkillTreeLayout
extends RefCounted

const DEFAULT_ORIGIN := Vector2(23, 70)
const DEFAULT_SPACING := Vector2(96, 127)
const DEFAULT_BLOCK_SIZE := Vector2(83, 83)


static func block_position(skill_id: String, origin: Vector2 = DEFAULT_ORIGIN, spacing: Vector2 = DEFAULT_SPACING) -> Vector2:
	var tree: Dictionary = SkillRulesLoader.skill_definition(skill_id).get("tree", {})
	var tier := maxi(1, int(tree.get("tier", 1)))
	var column := maxi(1, int(tree.get("column", 1)))

	return Vector2(
		origin.x + float(column - 1) * spacing.x,
		origin.y + float(tier - 1) * spacing.y,
	)


static func block_center(skill_id: String, origin: Vector2 = DEFAULT_ORIGIN, spacing: Vector2 = DEFAULT_SPACING, block_size: Vector2 = DEFAULT_BLOCK_SIZE) -> Vector2:
	var pos := block_position(skill_id, origin, spacing)

	return pos + block_size * 0.5


static func prerequisite_edges(visible_skill_ids: Array, skill_progression: Dictionary) -> Array:
	var visible := {}
	for raw_id in visible_skill_ids:
		visible[str(raw_id)] = true
	var edges: Array = []
	for raw_skill_id in visible_skill_ids:
		var skill_id := str(raw_skill_id)
		var requirements: Dictionary = SkillRulesLoader.skill_definition(skill_id).get("requirements", {})
		var prereqs: Array = requirements.get("skills", [])
		for raw_prereq in prereqs:
			if typeof(raw_prereq) != TYPE_DICTIONARY:
				continue
			var prereq := raw_prereq as Dictionary
			var from_id := str(prereq.get("skill_id", ""))
			var required_rank := int(prereq.get("rank", 0))
			if from_id == "" or required_rank <= 0 or not visible.has(from_id):
				continue
			var current_rank := _skill_rank(skill_progression, from_id)
			edges.append({
				"from": from_id,
				"to": skill_id,
				"met": current_rank >= required_rank,
			})

	return edges


static func _skill_rank(skill_progression: Dictionary, skill_id: String) -> int:
	var rows: Array = skill_progression.get("skills", [])
	for row in rows:
		if typeof(row) == TYPE_DICTIONARY and str((row as Dictionary).get("skill_id", "")) == skill_id:
			return int((row as Dictionary).get("rank", 0))

	return 0
