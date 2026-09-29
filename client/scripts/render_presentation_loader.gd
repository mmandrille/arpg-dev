## Static loader for the client render baseline (ADR-0018 D7).
## Data: shared/assets/render_presentation.v0.json (key light, per-context environment, quality tiers).
class_name RenderPresentationLoader
extends RefCounted

const DEFAULT_PATH := "../shared/assets/render_presentation.v0.json"
const CONTEXT_TOWN := "town"
const CONTEXT_TOWN_NIGHT := "town_night"
const CONTEXT_DUNGEON := "dungeon"

static var _loaded: bool = false
static var _config: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_config = {}
	var path := ProjectSettings.globalize_path("res://").path_join(DEFAULT_PATH)
	if not FileAccess.file_exists(path):
		push_warning("RenderPresentationLoader: data file missing: %s" % path)
		return
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("RenderPresentationLoader: could not open: %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("RenderPresentationLoader: malformed JSON: %s" % path)
		return
	_config = parsed as Dictionary


static func config() -> Dictionary:
	ensure_loaded()
	return _config.duplicate(true)


static func key_light() -> Dictionary:
	ensure_loaded()
	return (_config.get("key_light", {}) as Dictionary).duplicate(true)


## Lighting context for a level, mirroring DungeonDepthLighting.profile_for_level.
static func context_for_level(level: int, town_fog_active: bool = false) -> String:
	if level < 0:
		return CONTEXT_DUNGEON
	return CONTEXT_TOWN_NIGHT if town_fog_active else CONTEXT_TOWN


static func context(context_id: String) -> Dictionary:
	ensure_loaded()
	var contexts: Dictionary = _config.get("contexts", {})
	return (contexts.get(context_id, {}) as Dictionary).duplicate(true)


## Unknown tiers fall back to "balanced"; a missing catalog yields an empty tier (all effects off).
static func quality_tier(quality: String) -> Dictionary:
	ensure_loaded()
	var tiers: Dictionary = _config.get("quality_tiers", {})
	if tiers.has(quality):
		return (tiers[quality] as Dictionary).duplicate(true)
	return (tiers.get("balanced", {}) as Dictionary).duplicate(true)
