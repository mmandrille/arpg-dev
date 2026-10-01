## Resolves the trusted depth mood catalog by an existing dungeon palette identity.
class_name DungeonDepthMoodLoader
extends RefCounted

const DEFAULT_PATH := "../shared/assets/dungeon_depth_mood.v0.json"

static var _loaded := false
static var _config: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	var path := ProjectSettings.globalize_path("res://").path_join(DEFAULT_PATH)
	if not FileAccess.file_exists(path):
		push_warning("DungeonDepthMoodLoader: data file missing: %s" % path)
		return
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("DungeonDepthMoodLoader: could not open: %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("DungeonDepthMoodLoader: malformed JSON: %s" % path)
		return
	_config = parsed as Dictionary


static func profile_for_palette_id(palette_id: String) -> Dictionary:
	ensure_loaded()
	var profiles: Dictionary = _config.get("profiles", {})
	return (profiles.get(palette_id, {}) as Dictionary).duplicate(true)


static func profile_ids() -> Array:
	ensure_loaded()
	return (_config.get("profiles", {}) as Dictionary).keys()


static func context_with_fog(context_cfg: Dictionary, palette_id: String) -> Dictionary:
	var out := context_cfg.duplicate(true)
	var mood := profile_for_palette_id(palette_id)
	var fog: Dictionary = mood.get("fog", {})
	if fog.is_empty():
		return out
	var merged_fog: Dictionary = out.get("fog", {}).duplicate(true)
	merged_fog.merge(fog, true)
	out["fog"] = merged_fog
	return out


static func torch_config(base_config: Dictionary, palette_id: String) -> Dictionary:
	var out := base_config.duplicate(true)
	var mood := profile_for_palette_id(palette_id)
	var torch: Dictionary = mood.get("torch", {})
	out.merge(torch, true)
	return out
