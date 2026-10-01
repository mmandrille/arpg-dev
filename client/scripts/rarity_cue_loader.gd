class_name RarityCueLoader
extends RefCounted

const ItemRulesLoaderScript := preload("res://scripts/item_rules_loader.gd")
static var _loaded: bool = false
static var _catalog: Dictionary = {}


static func invalidate() -> void:
	_loaded = false
	_catalog = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/rarity_cues.v0.json")
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_error("RarityCueLoader: cannot read rarity cue catalog")
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		_catalog = parsed


static func catalog() -> Dictionary:
	ensure_loaded()
	return _catalog


static func cue_for_rarity(rarity: String) -> Dictionary:
	ensure_loaded()
	var entries: Dictionary = _catalog.get("rarities", {})
	return entries.get(rarity.strip_edges().to_lower(), {})


static func cue_for_item(item: Dictionary) -> Dictionary:
	if item.is_empty() or str(item.get("kind", "")) == "mystery" \
			or bool(item.get("concealed", false)) \
			or str(item.get("offer_id", "")).begins_with("mystery:"):
		return {}
	var def_id := str(item.get("item_def_id", ""))
	if def_id == "":
		def_id = str(item.get("item_template_id", ""))
	if def_id == "":
		return {}
	ItemRulesLoaderScript.ensure_loaded()
	var definition := ItemRulesLoaderScript.item_definition(def_id)
	if str(definition.get("category", "")).to_lower() != "equipment" \
			or not bool(definition.get("equippable", false)):
		return {}
	return cue_for_rarity(str(item.get("rarity", "")))


static func revealed_label(base_name: String, cue: Dictionary) -> String:
	if cue.is_empty() or base_name == "":
		return base_name
	var name := str(cue.get("name", ""))
	var lowered := base_name.to_lower()
	if lowered == name.to_lower() or lowered.begins_with(name.to_lower() + " "):
		return base_name
	return "%s · %s" % [name, base_name]
