## SkillRulesLoader - static singleton for shared skill rule data.
##
## Loads shared skill rules and skill presentations through the content-library
## manifest once and keeps gameplay mechanics separate from presentation metadata.
class_name SkillRulesLoader
extends RefCounted

const CONTENT_MANIFEST_REL := "../shared/content/content_libraries.v0.json"
const TextCatalogScript := preload("res://scripts/text_catalog.gd")

static var skill_rules: Dictionary = {}
static var skill_presentations: Dictionary = {}
static var _loaded: bool = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	skill_rules = {}
	skill_presentations = {}
	var manifest_path := ProjectSettings.globalize_path("res://").path_join(CONTENT_MANIFEST_REL)
	var manifest := _read_json_file(manifest_path, "content manifest")
	if manifest.is_empty():
		return
	_load_skill_rules(manifest_path, manifest)
	_load_skill_presentations(manifest_path, manifest)


static func skill_definition(skill_id: String) -> Dictionary:
	ensure_loaded()
	return skill_rules.get(skill_id, {})


static func skill_presentation(skill_id: String) -> Dictionary:
	ensure_loaded()
	return skill_presentations.get(skill_id, {})


static func skill_display_name(skill_id: String) -> String:
	var def := skill_definition(skill_id)
	return TextCatalogScript.get_text(str(def.get("name_key", "")), str(def.get("name", skill_id)))


static func skill_summary(skill_id: String) -> String:
	var presentation := skill_presentation(skill_id)
	return TextCatalogScript.get_text(str(presentation.get("summary_key", "")), str(presentation.get("summary", "Skill")))


static func skill_ids() -> Array:
	ensure_loaded()
	var ids := skill_rules.keys()
	ids.sort()
	return ids


static func skill_ids_by_tree() -> Array:
	ensure_loaded()
	var ids := skill_rules.keys()
	ids.sort_custom(func(a, b) -> bool:
		var a_id := str(a)
		var b_id := str(b)
		var a_tree: Dictionary = skill_definition(a_id).get("tree", {})
		var b_tree: Dictionary = skill_definition(b_id).get("tree", {})
		var a_tier := int(a_tree.get("tier", 999))
		var b_tier := int(b_tree.get("tier", 999))
		if a_tier != b_tier:
			return a_tier < b_tier
		var a_column := int(a_tree.get("column", 999))
		var b_column := int(b_tree.get("column", 999))
		if a_column != b_column:
			return a_column < b_column
		return a_id < b_id
	)
	return ids


static func first_skill_id() -> String:
	var ids := skill_ids_by_tree()
	if ids.is_empty():
		return ""
	return str(ids[0])


static func reset_for_tests() -> void:
	skill_rules = {}
	skill_presentations = {}
	_loaded = false


static func _load_skill_rules(manifest_path: String, manifest: Dictionary) -> void:
	var entries := _manifest_entries(manifest, ["rules", "skills"], "rules.skills")
	_merge_manifest_collection(manifest_path, entries, "skills", skill_rules)


static func _load_skill_presentations(manifest_path: String, manifest: Dictionary) -> void:
	var entries := _manifest_entries(manifest, ["assets", "skills", "presentations"], "assets.skills.presentations")
	_merge_manifest_collection(manifest_path, entries, "skills", skill_presentations)


static func _manifest_entries(manifest: Dictionary, keys: Array, label: String) -> Array:
	var node: Variant = manifest
	for key in keys:
		if typeof(node) != TYPE_DICTIONARY:
			push_warning("skill content manifest %s is not an object" % label)
			return []
		var node_dict: Dictionary = node
		if not node_dict.has(key):
			push_warning("skill content manifest missing %s" % label)
			return []
		node = node_dict.get(key)
	if typeof(node) != TYPE_ARRAY:
		push_warning("skill content manifest %s is not an array" % label)
		return []
	return node


static func _merge_manifest_collection(manifest_path: String, entries: Array, collection_key: String, target: Dictionary) -> void:
	for raw_entry in entries:
		if typeof(raw_entry) != TYPE_DICTIONARY:
			push_warning("skill content manifest entry is not an object")
			continue
		var entry: Dictionary = raw_entry
		var rel_path := str(entry.get("path", ""))
		if rel_path == "":
			push_warning("skill content manifest entry is missing path")
			continue
		var path := _resolve_manifest_path(manifest_path, rel_path)
		if path == "":
			continue
		var parsed := _read_json_file(path, "skill content file")
		if parsed.is_empty():
			continue
		var collection = parsed.get(collection_key, {})
		if typeof(collection) != TYPE_DICTIONARY:
			push_warning("skill content file %s missing %s object" % [rel_path, collection_key])
			continue
		for raw_id in collection.keys():
			var content_id := str(raw_id)
			if target.has(content_id):
				push_warning("skill content manifest duplicate id %s in %s" % [content_id, rel_path])
				continue
			target[content_id] = collection.get(raw_id)


static func _resolve_manifest_path(manifest_path: String, rel_path: String) -> String:
	if rel_path.begins_with("/") or rel_path.contains("://"):
		push_warning("skill content manifest path must be relative: %s" % rel_path)
		return ""
	return manifest_path.get_base_dir().path_join(rel_path).simplify_path()


static func _read_json_file(path: String, label: String) -> Dictionary:
	if not FileAccess.file_exists(path):
		push_warning("%s not found: %s" % [label, path])
		return {}
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		push_warning("%s could not be opened: %s" % [label, path])
		return {}
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		return parsed
	push_warning("%s is not a JSON object: %s" % [label, path])
	return {}
