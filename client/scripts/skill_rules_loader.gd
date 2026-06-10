## SkillRulesLoader - static singleton for shared skill rule data.
##
## Loads skills.v0.json and skill_presentations.v0.json once and keeps gameplay
## mechanics separate from client presentation metadata.
class_name SkillRulesLoader
extends RefCounted

static var skill_rules: Dictionary = {}
static var skill_presentations: Dictionary = {}
static var _loaded: bool = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_load_skill_rules()
	_load_skill_presentations()


static func skill_definition(skill_id: String) -> Dictionary:
	ensure_loaded()
	return skill_rules.get(skill_id, {})


static func skill_presentation(skill_id: String) -> Dictionary:
	ensure_loaded()
	return skill_presentations.get(skill_id, {})


static func skill_ids() -> Array:
	ensure_loaded()
	var ids := skill_rules.keys()
	ids.sort()
	return ids


static func first_skill_id() -> String:
	var ids := skill_ids()
	if ids.is_empty():
		return ""
	return str(ids[0])


static func _load_skill_rules() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/skills.v0.json")
	if not FileAccess.file_exists(path):
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		skill_rules = parsed.get("skills", {})


static func _load_skill_presentations() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/skill_presentations.v0.json")
	if not FileAccess.file_exists(path):
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		skill_presentations = parsed.get("skills", {})
