class_name SkillTreeLayout
extends RefCounted

const DEFAULT_ORIGIN := Vector2(23, 70)
const DEFAULT_SPACING := Vector2(96, 127)
const DEFAULT_BLOCK_SIZE := Vector2(83, 83)

const PASSIVE_DISPLAY_COLUMN_DEFAULT := 5
const SURVIVAL_DISPLAY_COLUMN := 6
const COMBAT_COLUMN_LIMIT := 5

static var _resolved_by_class: Dictionary = {}


static func block_position(skill_id: String, origin: Vector2 = DEFAULT_ORIGIN, spacing: Vector2 = DEFAULT_SPACING) -> Vector2:
	var tree := resolved_tree(skill_id)
	var tier := int(tree.get("tier", 1))
	var column := int(tree.get("column", 1))

	return Vector2(
		origin.x + float(column - 1) * spacing.x,
		origin.y + float(tier - 1) * spacing.y,
	)


static func block_center(skill_id: String, origin: Vector2 = DEFAULT_ORIGIN, spacing: Vector2 = DEFAULT_SPACING, block_size: Vector2 = DEFAULT_BLOCK_SIZE) -> Vector2:
	var pos := block_position(skill_id, origin, spacing)

	return pos + block_size * 0.5


static func resolved_tree(skill_id: String) -> Dictionary:
	var class_id := str(_skill_def(skill_id).get("class", ""))
	if class_id == "":
		return _static_tree(skill_id)

	return _resolve_class(class_id).get(skill_id, _static_tree(skill_id))


static func resolved_column(skill_id: String) -> int:
	return int(resolved_tree(skill_id).get("column", 1))


static func resolved_tier(skill_id: String) -> int:
	return int(resolved_tree(skill_id).get("tier", 1))


static func passive_display_column(class_id: String) -> int:
	_resolve_class(class_id)

	return int(_resolved_by_class.get(class_id, {}).get("_passive_column", PASSIVE_DISPLAY_COLUMN_DEFAULT))


static func prerequisite_edges(visible_skill_ids: Array, skill_progression: Dictionary) -> Array:
	var visible := {}
	for raw_id in visible_skill_ids:
		visible[str(raw_id)] = true
	var edges: Array = []
	for raw_skill_id in visible_skill_ids:
		var skill_id := str(raw_skill_id)
		var requirements: Dictionary = _skill_def(skill_id).get("requirements", {})
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


static func _skill_def(skill_id: String) -> Dictionary:
	return SkillRulesLoader.skill_definition(skill_id)


static func _static_tree(skill_id: String) -> Dictionary:
	var tree: Dictionary = _skill_def(skill_id).get("tree", {})

	return {
		"tier": maxi(1, int(tree.get("tier", 1))),
		"column": maxi(1, int(tree.get("column", 1))),
	}


static func _tree_hint(skill_id: String) -> Dictionary:
	return _skill_def(skill_id).get("tree", {})


static func _tier_hint(skill_id: String) -> int:
	return maxi(1, int(_tree_hint(skill_id).get("tier", 1)))


static func _column_hint(skill_id: String) -> int:
	return maxi(1, int(_tree_hint(skill_id).get("column", 1)))


static func _is_passive(skill_id: String) -> bool:
	var kind := str(_skill_def(skill_id).get("kind", ""))
	if kind == "passive_stat_bonus" or kind == "passive_execute":
		return true

	return _column_hint(skill_id) >= PASSIVE_DISPLAY_COLUMN_DEFAULT and str(_tree_hint(skill_id).get("branch", "")) != "survival"


static func _is_survival(skill_id: String) -> bool:
	return str(_tree_hint(skill_id).get("branch", "")) == "survival"


static func _primary_prereq(skill_id: String) -> String:
	var prereqs: Array = _skill_def(skill_id).get("requirements", {}).get("skills", [])
	if prereqs.is_empty():
		return ""
	var first = prereqs[0]
	if typeof(first) != TYPE_DICTIONARY:
		return ""

	return str((first as Dictionary).get("skill_id", ""))


static func _is_chain_continuer(skill_id: String, parent_id: String, combat_skills: Array) -> bool:
	if _column_hint(skill_id) == _column_hint(parent_id):
		return true
	var tier := _tier_hint(skill_id)
	var sibling_count := 0
	for candidate_id in combat_skills:
		if _primary_prereq(candidate_id) != parent_id:
			continue
		if _tier_hint(candidate_id) != tier:
			continue
		sibling_count += 1

	return sibling_count == 1


static func _class_skills(class_id: String) -> Array:
	var out: Array = []
	for raw_id in SkillRulesLoader.skill_ids():
		var skill_id := str(raw_id)
		if str(_skill_def(skill_id).get("class", "")) == class_id:
			out.append(skill_id)

	return out


static func _combat_skills(class_id: String) -> Array:
	var out: Array = []
	for skill_id in _class_skills(class_id):
		if _is_passive(skill_id) or _is_survival(skill_id):
			continue
		out.append(skill_id)
	out.sort_custom(_sort_skill_ids)

	return out


static func _sort_skill_ids(a: String, b: String) -> bool:
	var tier_a := _tier_hint(a)
	var tier_b := _tier_hint(b)
	if tier_a != tier_b:
		return tier_a < tier_b
	var column_a := _column_hint(a)
	var column_b := _column_hint(b)
	if column_a != column_b:
		return column_a < column_b

	return a < b


static func _occupancy_key(column: int, tier: int) -> String:
	return "%d:%d" % [column, tier]


static func _is_occupied(occupancy: Dictionary, column: int, tier: int) -> bool:
	return occupancy.has(_occupancy_key(column, tier))


static func _is_ancestor_of(ancestor_id: String, skill_id: String) -> bool:
	var current := skill_id
	while current != "":
		if current == ancestor_id:
			return true
		current = _primary_prereq(current)

	return false


static func _would_false_stack(occupancy: Dictionary, column: int, tier: int, skill_id: String) -> bool:
	for key in occupancy.keys():
		var parts: PackedStringArray = str(key).split(":")
		if parts.size() != 2:
			continue
		var occupied_column := int(parts[0])
		var occupied_tier := int(parts[1])
		if occupied_column != column or occupied_tier >= tier:
			continue
		var occupant_id := str(occupancy[key])
		if not _is_ancestor_of(occupant_id, skill_id):
			return true

	return false


static func _claim(occupancy: Dictionary, column: int, tier: int, skill_id: String) -> void:
	occupancy[_occupancy_key(column, tier)] = skill_id


static func _parent_column(parent_id: String, resolved: Dictionary) -> int:
	if resolved.has(parent_id):
		return int(resolved[parent_id].get("column", 1))

	return _column_hint(parent_id)


static func _fan_column(skill_id: String, parent_id: String, tier: int, combat_skills: Array, resolved: Dictionary) -> int:
	var siblings: Array = []
	for candidate_id in combat_skills:
		if _tier_hint(candidate_id) != tier:
			continue
		if _primary_prereq(candidate_id) != parent_id:
			continue
		siblings.append(candidate_id)
	siblings.sort_custom(_sort_skill_ids)
	var parent_column := _parent_column(parent_id, resolved)
	var fan_index := 0
	for sibling_id in siblings:
		if _is_chain_continuer(str(sibling_id), parent_id, combat_skills):
			continue
		fan_index += 1
		if str(sibling_id) == skill_id:
			return parent_column + fan_index

	return parent_column + 1


static func _lowest_free_column(start: int, tier: int, occupancy: Dictionary, skill_id: String) -> int:
	var column := maxi(1, start)
	while column <= COMBAT_COLUMN_LIMIT:
		if _is_occupied(occupancy, column, tier):
			column += 1
			continue
		if _would_false_stack(occupancy, column, tier, skill_id):
			column += 1
			continue

		return column

	return column


static func _resolve_class(class_id: String) -> Dictionary:
	if _resolved_by_class.has(class_id):
		return _resolved_by_class[class_id]

	var resolved := {}
	var occupancy := {}
	var combat_skills := _combat_skills(class_id)
	var max_tier := 1
	for skill_id in combat_skills:
		max_tier = maxi(max_tier, _tier_hint(skill_id))

	for tier in range(1, max_tier + 1):
		for skill_id in combat_skills:
			if _tier_hint(skill_id) != tier:
				continue
			var parent_id := _primary_prereq(skill_id)
			var column := _column_hint(skill_id)
			if parent_id != "":
				if _is_chain_continuer(skill_id, parent_id, combat_skills):
					column = _parent_column(parent_id, resolved)
				else:
					column = _fan_column(skill_id, parent_id, tier, combat_skills, resolved)
			column = _lowest_free_column(column, tier, occupancy, skill_id)
			resolved[skill_id] = {"tier": tier, "column": column}
			_claim(occupancy, column, tier, skill_id)

	var max_combat_column := 0
	for skill_id in resolved.keys():
		max_combat_column = maxi(max_combat_column, int(resolved[skill_id].get("column", 0)))
	var passive_column := PASSIVE_DISPLAY_COLUMN_DEFAULT
	if max_combat_column >= COMBAT_COLUMN_LIMIT:
		passive_column = maxi(PASSIVE_DISPLAY_COLUMN_DEFAULT + 2, max_combat_column + 2)

	for skill_id in _class_skills(class_id):
		if _is_passive(skill_id):
			resolved[skill_id] = {"tier": _tier_hint(skill_id), "column": passive_column}
		elif _is_survival(skill_id):
			resolved[skill_id] = {"tier": _tier_hint(skill_id), "column": SURVIVAL_DISPLAY_COLUMN}

	resolved["_passive_column"] = passive_column
	_resolved_by_class[class_id] = resolved

	return resolved


static func _skill_rank(skill_progression: Dictionary, skill_id: String) -> int:
	var rows: Array = skill_progression.get("skills", [])
	for row in rows:
		if typeof(row) == TYPE_DICTIONARY and str((row as Dictionary).get("skill_id", "")) == skill_id:
			return int((row as Dictionary).get("rank", 0))

	return 0
