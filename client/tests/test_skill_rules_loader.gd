# Unit test for content-manifest skill loading.
# Run via: godot --headless --path client --script res://tests/test_skill_rules_loader.gd
extends SceneTree

const SkillRulesLoaderScript := preload("res://scripts/skill_rules_loader.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	SkillRulesLoaderScript.reset_for_tests()
	SkillRulesLoaderScript.ensure_loaded()

	var ids := SkillRulesLoaderScript.skill_ids()
	_assert_true("manifest-loaded skills include holy shield", ids.has("holy_shield"))
	_assert_true("manifest-loaded skills include cleave", ids.has("cleave"))
	_assert_true("manifest-loaded skills include earthbreaker", ids.has("earthbreaker"))
	_assert_true("manifest-loaded skills include ice shard", ids.has("ice_shard"))
	_assert_true("manifest-loaded skills include ligthing", ids.has("ligthing"))
	_assert_true("manifest-loaded skills include arcane barrage", ids.has("arcane_barrage"))
	_assert_true("manifest-loaded skills include poison stab", ids.has("poison_stab"))
	_assert_true("manifest-loaded skills include dash", ids.has("dash"))
	_assert_true("manifest-loaded skills include shadow flurry", ids.has("shadow_flurry"))
	_assert_true("manifest-loaded skills include piercing shot", ids.has("piercing_shot"))
	_assert_true("manifest-loaded skills include pinning shot", ids.has("pinning_shot"))
	_assert_true("manifest-loaded skills include volley", ids.has("volley"))
	_assert_true("manifest-loaded skills include split arrow", ids.has("split_arrow"))
	_assert_eq("alphabetical first id stable", str(ids[0]), "arcane_barrage")
	_assert_eq("tree-order first skill stable", SkillRulesLoaderScript.first_skill_id(), "cleave")

	var skill := SkillRulesLoaderScript.skill_definition("magic_bolt")
	_assert_eq("skill name from manifest-listed rules", str(skill.get("name", "")), "Magic Bolt")
	_assert_eq("skill name key from manifest-listed rules", str(skill.get("name_key", "")), "skill.magic_bolt.name")
	_assert_eq("localized skill display name", SkillRulesLoaderScript.skill_display_name("magic_bolt"), "Magic Bolt")
	_assert_eq("skill projectile visual", str(skill.get("projectile", {}).get("visual", "")), "magic_bolt_projectile")
	_assert_eq("rage kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("rage").get("kind", "")), "self_buff")
	_assert_eq("cleave kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("cleave").get("kind", "")), "cone_attack")
	_assert_eq("earthbreaker kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("earthbreaker").get("kind", "")), "cone_attack")
	_assert_eq("ice shard kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("ice_shard").get("kind", "")), "cold_projectile_attack")
	_assert_eq("ligthing kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("ligthing").get("kind", "")), "chain_projectile_attack")
	_assert_eq("arcane barrage kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("arcane_barrage").get("kind", "")), "projectile_attack")
	_assert_eq("arcane barrage prerequisite", str(((SkillRulesLoaderScript.skill_definition("arcane_barrage").get("requirements", {}).get("skills", []) as Array)[0] as Dictionary).get("skill_id", "")), "ligthing")
	_assert_eq("heal kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("heal").get("kind", "")), "area_heal")
	_assert_eq("holy shield kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("holy_shield").get("kind", "")), "area_stat_buff")
	_assert_eq("sanctuary kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("sanctuary").get("kind", "")), "area_stat_buff")
	_assert_eq("sanctuary prerequisite", str(((SkillRulesLoaderScript.skill_definition("sanctuary").get("requirements", {}).get("skills", []) as Array)[0] as Dictionary).get("skill_id", "")), "holy_shield")
	_assert_eq("poison stab kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("poison_stab").get("kind", "")), "cone_attack")
	_assert_eq("dash kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("dash").get("kind", "")), "cone_attack")
	_assert_eq("shadow flurry kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("shadow_flurry").get("kind", "")), "cone_attack")
	_assert_eq("executioner kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("executioner").get("kind", "")), "passive_execute")
	_assert_eq("piercing shot kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("piercing_shot").get("kind", "")), "projectile_attack")
	_assert_eq("pinning shot kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("pinning_shot").get("kind", "")), "projectile_attack")
	_assert_eq("piercing shot max hits", int(SkillRulesLoaderScript.skill_definition("piercing_shot").get("pierce", {}).get("max_hits", 0)), 4)
	_assert_eq("pinning shot root effect", str(SkillRulesLoaderScript.skill_definition("pinning_shot").get("root", {}).get("effect_id", "")), "pinning_root")
	_assert_eq("volley kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("volley").get("kind", "")), "projectile_attack")
	_assert_eq("volley arrow count", int(SkillRulesLoaderScript.skill_definition("volley").get("volley", {}).get("arrow_count", 0)), 5)
	_assert_eq("volley prerequisite", str(((SkillRulesLoaderScript.skill_definition("volley").get("requirements", {}).get("skills", []) as Array)[0] as Dictionary).get("skill_id", "")), "piercing_shot")
	_assert_eq("split arrow kind from manifest-listed rules", str(SkillRulesLoaderScript.skill_definition("split_arrow").get("kind", "")), "projectile_attack")
	_assert_eq("split arrow prerequisite", str(((SkillRulesLoaderScript.skill_definition("split_arrow").get("requirements", {}).get("skills", []) as Array)[0] as Dictionary).get("skill_id", "")), "volley")

	var presentation := SkillRulesLoaderScript.skill_presentation("magic_bolt")
	_assert_eq("presentation summary key from manifest-listed assets", str(presentation.get("summary_key", "")), "skill.magic_bolt.summary")
	_assert_eq("localized skill summary", SkillRulesLoaderScript.skill_summary("magic_bolt"), "Projectile spell")
	_assert_eq("presentation label from manifest-listed assets", str(presentation.get("icon", {}).get("label", "")), "M")
	_assert_eq("presentation projectile visual", str(presentation.get("projectile_visual", "")), "magic_bolt_projectile")
	_assert_eq("rage presentation label", str(SkillRulesLoaderScript.skill_presentation("rage").get("icon", {}).get("label", "")), "R")
	_assert_eq("cleave presentation label", str(SkillRulesLoaderScript.skill_presentation("cleave").get("icon", {}).get("label", "")), "C")
	_assert_eq("earthbreaker presentation label", str(SkillRulesLoaderScript.skill_presentation("earthbreaker").get("icon", {}).get("label", "")), "E")
	_assert_eq("rage presentation shape", str(SkillRulesLoaderScript.skill_presentation("rage").get("icon", {}).get("shape", "")), "flame")
	_assert_eq("cleave presentation shape", str(SkillRulesLoaderScript.skill_presentation("cleave").get("icon", {}).get("shape", "")), "slash")
	_assert_eq("leap presentation shape", str(SkillRulesLoaderScript.skill_presentation("leap").get("icon", {}).get("shape", "")), "leap")
	_assert_eq("earthbreaker presentation shape", str(SkillRulesLoaderScript.skill_presentation("earthbreaker").get("icon", {}).get("shape", "")), "quake")
	_assert_eq("ice shard presentation label", str(SkillRulesLoaderScript.skill_presentation("ice_shard").get("icon", {}).get("label", "")), "I")
	_assert_eq("ice shard presentation shape", str(SkillRulesLoaderScript.skill_presentation("ice_shard").get("icon", {}).get("shape", "")), "ice_spike")
	_assert_eq("ligthing presentation label", str(SkillRulesLoaderScript.skill_presentation("ligthing").get("icon", {}).get("label", "")), "L")
	_assert_eq("arcane barrage presentation label", str(SkillRulesLoaderScript.skill_presentation("arcane_barrage").get("icon", {}).get("label", "")), "A")
	_assert_eq("arcane barrage projectile presentation", str(SkillRulesLoaderScript.skill_presentation("arcane_barrage").get("projectile_visual", "")), "arcane_barrage_projectile")
	_assert_eq("heal presentation label", str(SkillRulesLoaderScript.skill_presentation("heal").get("icon", {}).get("label", "")), "H")
	_assert_eq("charge presentation shape", str(SkillRulesLoaderScript.skill_presentation("charge").get("icon", {}).get("shape", "")), "shield_charge")
	_assert_eq("holy shield presentation label", str(SkillRulesLoaderScript.skill_presentation("holy_shield").get("icon", {}).get("label", "")), "S")
	_assert_eq("holy shield presentation shape", str(SkillRulesLoaderScript.skill_presentation("holy_shield").get("icon", {}).get("shape", "")), "shield")
	_assert_eq("sanctuary presentation label", str(SkillRulesLoaderScript.skill_presentation("sanctuary").get("icon", {}).get("label", "")), "T")
	_assert_eq("sanctuary presentation shape", str(SkillRulesLoaderScript.skill_presentation("sanctuary").get("icon", {}).get("shape", "")), "shield")
	_assert_eq("poison stab presentation label", str(SkillRulesLoaderScript.skill_presentation("poison_stab").get("icon", {}).get("label", "")), "P")
	_assert_eq("poison stab presentation shape", str(SkillRulesLoaderScript.skill_presentation("poison_stab").get("icon", {}).get("shape", "")), "poison_dagger")
	_assert_eq("dash presentation label", str(SkillRulesLoaderScript.skill_presentation("dash").get("icon", {}).get("label", "")), "D")
	_assert_eq("dash presentation shape", str(SkillRulesLoaderScript.skill_presentation("dash").get("icon", {}).get("shape", "")), "dash")
	_assert_eq("shadow flurry presentation label", str(SkillRulesLoaderScript.skill_presentation("shadow_flurry").get("icon", {}).get("label", "")), "F")
	_assert_eq("executioner presentation label", str(SkillRulesLoaderScript.skill_presentation("executioner").get("icon", {}).get("label", "")), "X")
	_assert_eq("piercing shot presentation label", str(SkillRulesLoaderScript.skill_presentation("piercing_shot").get("icon", {}).get("label", "")), "P")
	_assert_eq("piercing shot presentation shape", str(SkillRulesLoaderScript.skill_presentation("piercing_shot").get("icon", {}).get("shape", "")), "arrow")
	_assert_eq("pinning shot presentation label", str(SkillRulesLoaderScript.skill_presentation("pinning_shot").get("icon", {}).get("label", "")), "N")
	_assert_eq("pinning shot presentation shape", str(SkillRulesLoaderScript.skill_presentation("pinning_shot").get("icon", {}).get("shape", "")), "pin")
	_assert_eq("piercing shot projectile presentation", str(SkillRulesLoaderScript.skill_presentation("piercing_shot").get("projectile_visual", "")), "piercing_shot_projectile")
	_assert_eq("pinning shot projectile presentation", str(SkillRulesLoaderScript.skill_presentation("pinning_shot").get("projectile_visual", "")), "pinning_shot_projectile")
	_assert_eq("volley presentation label", str(SkillRulesLoaderScript.skill_presentation("volley").get("icon", {}).get("label", "")), "V")
	_assert_eq("volley presentation shape", str(SkillRulesLoaderScript.skill_presentation("volley").get("icon", {}).get("shape", "")), "volley")
	_assert_eq("volley projectile presentation", str(SkillRulesLoaderScript.skill_presentation("volley").get("projectile_visual", "")), "volley_arrow_projectile")
	_assert_eq("split arrow presentation label", str(SkillRulesLoaderScript.skill_presentation("split_arrow").get("icon", {}).get("label", "")), "S")
	_assert_eq("split arrow projectile presentation", str(SkillRulesLoaderScript.skill_presentation("split_arrow").get("projectile_visual", "")), "split_arrow_projectile")

	print("[gdtest] PASS: test_skill_rules_loader (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
