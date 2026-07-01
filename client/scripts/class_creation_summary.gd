## ClassCreationSummary — shared rules for hero-creation class feature lists.
class_name ClassCreationSummary
extends RefCounted

const PROGRESSION_REL := "../shared/rules/character_progression.v0.json"
const SkillRulesLoaderScript := preload("res://scripts/skill_rules_loader.gd")
const TextCatalogScript := preload("res://scripts/text_catalog.gd")

static var _classes: Dictionary = {}
static var _loaded: bool = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_classes = {}
	var path := ProjectSettings.globalize_path("res://").path_join(PROGRESSION_REL)
	var data := _read_json(path)
	if data.is_empty():
		return
	_classes = data.get("classes", {}) as Dictionary
	SkillRulesLoaderScript.ensure_loaded()


static func feature_summary(class_id: String) -> Dictionary:
	ensure_loaded()
	var class_def: Dictionary = _classes.get(class_id, {}) as Dictionary
	var stats: Dictionary = class_def.get("base_stats", {}) as Dictionary
	var actives_by_tier: Dictionary = {}
	var passives: Array = []
	for skill_id in SkillRulesLoaderScript.skill_ids():
		var def: Dictionary = SkillRulesLoaderScript.skill_definition(str(skill_id))
		if str(def.get("class", "")) != class_id:
			continue
		var kind := str(def.get("kind", ""))
		var display := SkillRulesLoaderScript.skill_display_name(str(skill_id))
		if kind == "passive_stat_bonus":
			passives.append(display)
			continue
		var tree: Dictionary = def.get("tree", {}) as Dictionary
		var tier := int(tree.get("tier", 1))
		if not actives_by_tier.has(tier):
			actives_by_tier[tier] = []
		(actives_by_tier[tier] as Array).append(display)
	for tier in actives_by_tier.keys():
		(actives_by_tier[tier] as Array).sort()
	passives.sort()
	var tier_keys := actives_by_tier.keys()
	tier_keys.sort()
	var active_tiers: Array = []
	for tier in tier_keys:
		active_tiers.append({
			"tier": int(tier),
			"skills": (actives_by_tier[tier] as Array).duplicate(),
		})
	var movement := float(class_def.get("base_movement_speed", 0.0))
	var light := float(class_def.get("light_radius", 0.0))
	return {
		"class_id": class_id,
		"class_name": str(class_def.get("name", class_id.capitalize())),
		"stats": stats.duplicate(true),
		"movement_speed": movement,
		"light_radius": light,
		"active_tiers": active_tiers,
		"passives": passives,
		"feature_lines": _feature_lines(class_def, stats, movement, light, active_tiers, passives),
	}


static func _feature_lines(
	class_def: Dictionary,
	stats: Dictionary,
	movement: float,
	light: float,
	active_tiers: Array,
	passives: Array,
) -> Array:
	var lines: Array = []
	lines.append(str(class_def.get("name", "Hero")))
	lines.append("%s %d  %s %d  %s %d  %s %d" % [
		TextCatalogScript.get_text("stat.str", "STR"), int(stats.get("str", 0)),
		TextCatalogScript.get_text("stat.dex", "DEX"), int(stats.get("dex", 0)),
		TextCatalogScript.get_text("stat.vit", "VIT"), int(stats.get("vit", 0)),
		TextCatalogScript.get_text("stat.magic", "MAGIC").to_upper(), int(stats.get("magic", 0)),
	])
	if movement > 0.0 or light > 0.0:
		lines.append("%s %.0f%%  %s %.0f" % [
			TextCatalogScript.get_text("character.movement", "Move"),
			movement * 100.0,
			TextCatalogScript.get_text("character.light_radius", "Light"),
			light,
		])
	for tier_entry in active_tiers:
		if typeof(tier_entry) != TYPE_DICTIONARY:
			continue
		var rec: Dictionary = tier_entry
		var tier := int(rec.get("tier", 0))
		var skills: Array = rec.get("skills", []) as Array
		if skills.is_empty():
			continue
		lines.append("%s %d: %s" % [
			TextCatalogScript.get_text("character.skill_tier", "Tier"),
			tier,
			", ".join(skills),
		])
	if not passives.is_empty():
		lines.append("%s: %s" % [
			TextCatalogScript.get_text("character.passives", "Passives"),
			", ".join(passives),
		])

	return lines


static func _read_json(path: String) -> Dictionary:
	if not FileAccess.file_exists(path):
		push_warning("class creation summary missing: %s" % path)
		return {}
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		return {}
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("class creation summary malformed: %s" % path)
		return {}
	return parsed
