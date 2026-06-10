## ItemRulesLoader — static singleton for shared item rule data.
##
## Loads items.v0.json, item_templates.v0.json, and item_presentations.v0.json
## once (lazy, guarded by _loaded) and caches the results as static vars.
## All panels and main.gd access these dictionaries directly instead of each
## loading their own copy.
##
## Usage (any script — no autoload registration needed):
##   ItemRulesLoader.ensure_loaded()          # call once in _ready()
##   var def := ItemRulesLoader.item_definition(def_id)
##   var icon := ItemRulesLoader.item_presentations.get(def_id, {}).get("icon", {})
class_name ItemRulesLoader
extends RefCounted

static var item_rules: Dictionary = {}
static var item_templates: Dictionary = {}
static var item_presentations: Dictionary = {}
static var _loaded: bool = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_load_item_rules()
	_load_item_templates()
	_load_item_presentations()


static func item_definition(def_id: String) -> Dictionary:
	if item_rules.has(def_id):
		return item_rules.get(def_id, {})
	return item_templates.get(def_id, {})


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


static func _load_item_presentations() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/item_presentations.v0.json")
	if not FileAccess.file_exists(path):
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		item_presentations = parsed.get("items", {})
