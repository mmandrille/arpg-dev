extends SceneTree

const SkillRulesLoaderScript := preload("res://scripts/skill_rules_loader.gd")
const SkillTreeLayoutScript := preload("res://scripts/skill_tree_layout.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	SkillRulesLoaderScript.ensure_loaded()
	_test_sorcerer_branch_bands()
	_test_ranger_prerequisite_columns()
	_test_rogue_fan_column()
	_test_no_false_stacks()
	print("[gdtest] PASS: test_skill_tree_layout (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_sorcerer_branch_bands() -> void:
	_assert_eq("magic bolt starts branch band", _column("magic_bolt"), 1)
	_assert_eq("teleport starts after revive band", _column("teleport"), 5)
	_assert_eq("ice shard under magic bolt band", _column("ice_shard"), 1)
	_assert_eq("lightning fans in magic bolt row", _column("lightning"), 2)
	_assert_eq("revive sits in second row column", _column("revive"), 4)
	_assert_ne("revive does not stack on lightning", _column("revive"), _column("lightning"))
	_assert_eq("fireball chains ice shard", _column("fireball"), _column("ice_shard"))
	_assert_eq("arcane barrage chains lightning", _column("arcane_barrage"), _column("lightning"))
	_assert_eq("energy ward chains teleport", _column("energy_ward"), _column("teleport"))


func _test_ranger_prerequisite_columns() -> void:
	_assert_ne("pinning shot not under disengage column", _column("pinning_shot"), _column("disengage"))
	_assert_ne("snipe not under companion column", _column("snipe"), _column("black_wolf_companion"))
	_assert_eq("volley chains under piercing shot", _column("volley"), _column("piercing_shot"))
	_assert_eq("hunters volley chains under volley", _column("hunters_volley"), _column("volley"))
	_assert_eq("rain fans beside hunters volley", _column("rain_of_arrows"), _column("volley") + 1)
	_assert_eq("explosive shot chains under snipe", _column("explosive_shot"), _column("snipe"))


func _test_rogue_fan_column() -> void:
	_assert_ne("fan of blades not under dash", _column("fan_of_blades"), _column("dash"))
	_assert_true("fan of blades descends from poison stab", _is_ancestor_of("poison_stab", "fan_of_blades"))


func _test_no_false_stacks() -> void:
	for class_id in ["barbarian", "paladin", "sorcerer", "ranger", "rogue"]:
		_assert_cross_branch_false_stacks_for_class(class_id)


func _tier1_roots_for_class(class_id: String) -> Array:
	var roots: Array = []
	for skill_id in _combat_skill_ids(class_id):
		var def := SkillRulesLoaderScript.skill_definition(skill_id)
		var tier := int(def.get("tree", {}).get("tier", 1))
		var prereqs: Array = def.get("requirements", {}).get("skills", [])
		if tier == 1 and prereqs.is_empty():
			roots.append(skill_id)
	roots.sort()

	return roots


func _branch_root(skill_id: String, roots: Array) -> String:
	for root_id in roots:
		if skill_id == str(root_id) or _is_ancestor_of(str(root_id), skill_id):
			return str(root_id)

	return ""


func _assert_cross_branch_false_stacks_for_class(class_id: String) -> void:
	var roots := _tier1_roots_for_class(class_id)
	for skill_id in _combat_skill_ids(class_id):
		var tier := SkillTreeLayoutScript.resolved_tier(skill_id)
		var column := SkillTreeLayoutScript.resolved_column(skill_id)
		var skill_root := _branch_root(skill_id, roots)
		for other_id in _combat_skill_ids(class_id):
			if other_id == skill_id:
				continue
			if SkillTreeLayoutScript.resolved_tier(other_id) >= tier:
				continue
			if SkillTreeLayoutScript.resolved_column(other_id) != column:
				continue
			var other_root := _branch_root(str(other_id), roots)
			_assert_true(
				"%s and %s share column %d across branches" % [other_id, skill_id, column],
				skill_root == "" or other_root == "" or skill_root == other_root,
			)


func _is_ancestor_of(ancestor_id: String, skill_id: String) -> bool:
	var current := skill_id
	while current != "":
		if current == ancestor_id:
			return true
		current = _primary_prereq(current)

	return false


func _combat_skill_ids(class_id: String) -> Array:
	var out: Array = []
	for raw_id in SkillRulesLoaderScript.skill_ids():
		var skill_id := str(raw_id)
		var def := SkillRulesLoaderScript.skill_definition(skill_id)
		if str(def.get("class", "")) != class_id:
			continue
		var kind := str(def.get("kind", ""))
		if kind == "passive_stat_bonus" or kind == "passive_execute":
			continue
		if str(def.get("tree", {}).get("branch", "")) == "survival":
			continue
		out.append(skill_id)

	return out


func _primary_prereq(skill_id: String) -> String:
	var prereqs: Array = SkillRulesLoaderScript.skill_definition(skill_id).get("requirements", {}).get("skills", [])
	if prereqs.is_empty():
		return ""
	var first = prereqs[0]
	if typeof(first) != TYPE_DICTIONARY:
		return ""

	return str((first as Dictionary).get("skill_id", ""))


func _column(skill_id: String) -> int:
	return SkillTreeLayoutScript.resolved_column(skill_id)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_ne(label: String, got, expected) -> void:
	if got != expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected not %s" % [label, str(expected)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
