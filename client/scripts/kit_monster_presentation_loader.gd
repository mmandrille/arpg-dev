## Static loader for KayKit monster scenes (ADR-0018 P4a).
## Data: shared/assets/kit_monster_presentation.v0.json.
class_name KitMonsterPresentationLoader
extends RefCounted

const DEFAULT_PATH := "../shared/assets/kit_monster_presentation.v0.json"

static var _loaded: bool = false
static var _config: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_config = {}
	var path := ProjectSettings.globalize_path("res://").path_join(DEFAULT_PATH)
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("KitMonsterPresentationLoader: cannot open %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		_config = parsed as Dictionary
	else:
		push_warning("KitMonsterPresentationLoader: malformed JSON: %s" % path)


static func monster(visual_key: String) -> Dictionary:
	ensure_loaded()
	return ((_config.get("monsters", {}) as Dictionary).get(visual_key, {}) as Dictionary).duplicate(true)


static func clip_profile(profile_id: String) -> Dictionary:
	ensure_loaded()
	return ((_config.get("clip_profiles", {}) as Dictionary).get(profile_id, {}) as Dictionary).duplicate(true)


static func visual_keys() -> Array:
	ensure_loaded()
	var keys := (_config.get("monsters", {}) as Dictionary).keys()
	keys.sort()
	return keys


## v511 variant looks (rarity/depth/family); empty when the catalog has no `variants` section.
static func variants() -> Dictionary:
	ensure_loaded()
	return ((_config.get("variants", {}) as Dictionary)).duplicate(true)
