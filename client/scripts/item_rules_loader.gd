## ItemRulesLoader — static singleton for shared item rule data.
##
## Loads items.v0.json, item_templates.v0.json, shops.v0.json, unique_effects.v0.json, and item_presentations.v0.json
## once (lazy, guarded by _loaded) and caches the results as static vars.
## All panels and main.gd access these dictionaries directly instead of each
## loading their own copy.
##
## Usage (any script — no autoload registration needed):
##   ItemRulesLoader.ensure_loaded()          # call once in _ready()
##   var def := ItemRulesLoader.item_definition(def_id)
##   var icon := ItemRulesLoader.item_presentation(def_id).get("icon", {})
class_name ItemRulesLoader
extends RefCounted

static var item_rules: Dictionary = {}
static var item_templates: Dictionary = {}
static var shop_rules: Dictionary = {}
static var unique_effects: Dictionary = {}
static var item_presentations: Dictionary = {}
static var item_presentation_families: Dictionary = {}
static var _loaded: bool = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_load_item_rules()
	_load_item_templates()
	_load_shop_rules()
	_load_unique_effects()
	_load_item_presentations()


static func item_definition(def_id: String) -> Dictionary:
	if item_rules.has(def_id):
		return item_rules.get(def_id, {})
	return item_templates.get(def_id, {})


static func item_presentation(def_id: String) -> Dictionary:
	return item_presentations.get(def_id, {})


static func unique_effect_definition(effect_id: String) -> Dictionary:
	return unique_effects.get(effect_id, {})


static func _load_item_rules() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/items.v0.json")
	if not FileAccess.file_exists(path):
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		item_rules = parsed.get("items", {})


static func _load_item_templates() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/item_templates.v0.json")
	if not FileAccess.file_exists(path):
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		item_templates = parsed.get("templates", {})


static func _load_shop_rules() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/shops.v0.json")
	if not FileAccess.file_exists(path):
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		shop_rules = parsed


static func _load_unique_effects() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/unique_effects.v0.json")
	if not FileAccess.file_exists(path):
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		unique_effects = parsed.get("effects", {})


static func _load_item_presentations() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/item_presentations.v0.json")
	if not FileAccess.file_exists(path):
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		item_presentation_families = parsed.get("families", {})
		item_presentations = _resolved_item_presentations(parsed.get("items", {}), item_presentation_families)


static func _resolved_item_presentations(items: Dictionary, families: Dictionary) -> Dictionary:
	var resolved := {}
	for def_id in items.keys():
		var entry: Dictionary = items.get(def_id, {})
		var family_id := str(entry.get("family", ""))
		var family: Dictionary = families.get(family_id, {})
		var presentation := family.duplicate(true)
		for key in ["icon", "ground", "3d_model"]:
			if entry.has(key):
				var value = entry.get(key)
				presentation[key] = (value as Dictionary).duplicate(true) if typeof(value) == TYPE_DICTIONARY else value
		presentation["family"] = family_id
		resolved[str(def_id)] = presentation
	return resolved
