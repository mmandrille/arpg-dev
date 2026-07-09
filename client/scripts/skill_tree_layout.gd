class_name SkillTreeLayout
extends RefCounted

const DEFAULT_ORIGIN := Vector2(23, 70)
const DEFAULT_SPACING := Vector2(96, 127)
const DEFAULT_BLOCK_SIZE := Vector2(83, 83)

const PASSIVE_DISPLAY_COLUMN_DEFAULT := 5
const SURVIVAL_DISPLAY_COLUMN := 6
const COMBAT_COLUMN_LIMIT := 8
const TREE_RIGHT_PADDING := 16.0
const DECUPLE_CLASS_IDS := ["barbarian", "paladin", "sorcerer", "ranger", "rogue"]

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


static func required_tree_width(
	origin: Vector2 = DEFAULT_ORIGIN,
	spacing: Vector2 = DEFAULT_SPACING,
	block_size: Vector2 = DEFAULT_BLOCK_SIZE,
	right_padding: float = TREE_RIGHT_PADDING,
) -> float:
	SkillRulesLoader.ensure_loaded()
	var max_right := 0.0
	for class_id in DECUPLE_CLASS_IDS:
		_resolve_class(class_id)
		for skill_id in _class_skills(class_id):
			var pos := block_position(str(skill_id), origin, spacing)
			max_right = maxf(max_right, pos.x + block_size.x)

	return max_right + right_padding


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


static func _is_ancestor_of(ancestor_id: String, skill_id: String) -> bool:
	var current := skill_id
	while current != "":
		if current == ancestor_id:
			return true
		current = _primary_prereq(current)

	return false


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


static func _tier1_roots(combat_skills: Array) -> Array:
	var roots: Array = []
	for skill_id in combat_skills:
		if _tier_hint(skill_id) == 1 and _primary_prereq(skill_id) == "":
			roots.append(skill_id)
	roots.sort_custom(_sort_skill_ids)

	return roots


static func _branch_skill_ids(root_id: String, combat_skills: Array) -> Array:
	var out: Array = [root_id]
	for skill_id in combat_skills:
		if skill_id == root_id:
			continue
		if _is_ancestor_of(root_id, skill_id):
			out.append(skill_id)

	return out


static func _assign_sibling_columns(parent_id: String, tier: int, branch_skills: Array, combat_skills: Array, resolved: Dictionary) -> void:
	var siblings: Array = []
	for skill_id in branch_skills:
		if _tier_hint(skill_id) != tier:
			continue
		if _primary_prereq(skill_id) != parent_id:
			continue
		siblings.append(skill_id)
	if siblings.is_empty():
		return
	siblings.sort_custom(_sort_skill_ids)
	var parent_col := int(resolved.get(parent_id, {}).get("column", 1))
	var fan_offset := 0
	var chain_used := false
	for sibling_id in siblings:
		var col: int
		if not chain_used and _is_chain_continuer(str(sibling_id), parent_id, combat_skills):
			col = parent_col
			chain_used = true
		else:
			fan_offset += 1
			col = parent_col + fan_offset
		resolved[sibling_id] = {"tier": tier, "column": col}


static func _max_branch_column(branch_skills: Array, resolved: Dictionary) -> int:
	var max_col := 1
	for skill_id in branch_skills:
		if not resolved.has(skill_id):
			continue
		max_col = maxi(max_col, int(resolved[skill_id].get("column", 1)))

	return max_col


static func _assign_branch(root_id: String, start_col: int, combat_skills: Array, resolved: Dictionary) -> int:
	var branch_skills := _branch_skill_ids(root_id, combat_skills)
	resolved[root_id] = {"tier": 1, "column": start_col}
	var max_tier := 1
	for skill_id in branch_skills:
		max_tier = maxi(max_tier, _tier_hint(skill_id))
	for tier in range(2, max_tier + 1):
		var parents := {}
		for skill_id in branch_skills:
			if _tier_hint(skill_id) != tier:
				continue
			var parent_id := _primary_prereq(skill_id)
			if parent_id != "":
				parents[parent_id] = true
		var parent_ids: Array = parents.keys()
		parent_ids.sort_custom(func(a, b) -> bool:
			var tier_a := _tier_hint(str(a))
			var tier_b := _tier_hint(str(b))
			if tier_a != tier_b:
				return tier_a < tier_b
			return str(a) < str(b)
		)
		for parent_id in parent_ids:
			if not resolved.has(str(parent_id)):
				continue
			_assign_sibling_columns(str(parent_id), tier, branch_skills, combat_skills, resolved)

	return _max_branch_column(branch_skills, resolved)


static func _assign_unanchored_combat_skills(combat_skills: Array, resolved: Dictionary) -> void:
	for skill_id in combat_skills:
		if resolved.has(skill_id):
			continue
		if _primary_prereq(skill_id) != "":
			continue
		resolved[skill_id] = _static_tree(skill_id)

	var max_tier := 1
	for skill_id in combat_skills:
		max_tier = maxi(max_tier, _tier_hint(skill_id))

	for tier in range(2, max_tier + 1):
		var parents := {}
		for skill_id in combat_skills:
			if _tier_hint(skill_id) != tier:
				continue
			if resolved.has(skill_id):
				continue
			var parent_id := _primary_prereq(skill_id)
			if parent_id != "" and resolved.has(parent_id):
				parents[parent_id] = true
		var parent_ids: Array = parents.keys()
		parent_ids.sort_custom(func(a, b) -> bool:
			var tier_a := _tier_hint(str(a))
			var tier_b := _tier_hint(str(b))
			if tier_a != tier_b:
				return tier_a < tier_b
			return str(a) < str(b)
		)
		for parent_id in parent_ids:
			_assign_sibling_columns(str(parent_id), tier, combat_skills, combat_skills, resolved)


static func _max_resolved_column(resolved: Dictionary) -> int:
	var max_col := 0
	for skill_id in resolved.keys():
		if str(skill_id).begins_with("_"):
			continue
		max_col = maxi(max_col, int(resolved[skill_id].get("column", 0)))

	return max_col


static func _resolve_class(class_id: String) -> Dictionary:
	if _resolved_by_class.has(class_id):
		return _resolved_by_class[class_id]

	var resolved := {}
	var combat_skills := _combat_skills(class_id)
	var tier1_roots := _tier1_roots(combat_skills)
	if not tier1_roots.is_empty():
		_assign_branch(str(tier1_roots[0]), 1, combat_skills, resolved)

	_assign_unanchored_combat_skills(combat_skills, resolved)

	var next_root_col := _max_resolved_column(resolved) + 1
	for root_index in range(1, tier1_roots.size()):
		var branch_max_col := _assign_branch(str(tier1_roots[root_index]), next_root_col, combat_skills, resolved)
		next_root_col = branch_max_col + 1

	var max_combat_column := 0
	for skill_id in resolved.keys():
		max_combat_column = maxi(max_combat_column, int(resolved[skill_id].get("column", 0)))
	var passive_column := PASSIVE_DISPLAY_COLUMN_DEFAULT
	if max_combat_column >= PASSIVE_DISPLAY_COLUMN_DEFAULT:
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
