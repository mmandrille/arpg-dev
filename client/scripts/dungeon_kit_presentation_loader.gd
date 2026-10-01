## Static loader for KayKit dungeon piece selection (ADR-0018 P2).
## Data: shared/assets/dungeon_kit_presentation.v0.json.
class_name DungeonKitPresentationLoader
extends RefCounted

const DEFAULT_PATH := "../shared/assets/dungeon_kit_presentation.v0.json"

static var _loaded: bool = false
static var _config: Dictionary = {}
## Test hook: forces the kit on/off regardless of the catalog flag ("" = use catalog).
static var enabled_override: String = ""


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_config = {}
	var path := ProjectSettings.globalize_path("res://").path_join(DEFAULT_PATH)
	if not FileAccess.file_exists(path):
		push_warning("DungeonKitPresentationLoader: data file missing: %s" % path)
		return
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("DungeonKitPresentationLoader: could not open: %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("DungeonKitPresentationLoader: malformed JSON: %s" % path)
		return
	_config = parsed as Dictionary


static func config() -> Dictionary:
	ensure_loaded()
	return _config.duplicate(true)


static func active_for_level(level: int) -> bool:
	ensure_loaded()
	var enabled := bool(_config.get("enabled", false))
	if enabled_override == "on":
		enabled = true
	elif enabled_override == "off":
		enabled = false
	return enabled and level <= -int(_config.get("min_dungeon_depth", 1))


static func wall_asset_ids() -> Dictionary:
	ensure_loaded()
	var wall: Dictionary = _config.get("wall", {})
	return {"full": str(wall.get("full_asset_id", "")), "half": str(wall.get("half_asset_id", ""))}


static func column_asset_id() -> String:
	ensure_loaded()
	return str((_config.get("column", {}) as Dictionary).get("asset_id", ""))


static func floor_config() -> Dictionary:
	ensure_loaded()
	return (_config.get("floor", {}) as Dictionary).duplicate(true)


static func dressing_config() -> Dictionary:
	ensure_loaded()
	return (_config.get("dressing", {}) as Dictionary).duplicate(true)


static func legacy_disabled(feature: String) -> bool:
	ensure_loaded()
	return bool((_config.get("disable_legacy", {}) as Dictionary).get(feature, false))


static func torch_config() -> Dictionary:
	ensure_loaded()
	return (_config.get("torch", {}) as Dictionary).duplicate(true)


static func chest_config() -> Dictionary:
	ensure_loaded()
	return (_config.get("chest", {}) as Dictionary).duplicate(true)


static func stairs_config() -> Dictionary:
	ensure_loaded()
	return (_config.get("stairs", {}) as Dictionary).duplicate(true)


## Props (torches/chests/stairs) follow the global kit flag plus their own catalog flag.
static func prop_enabled(prop: String) -> bool:
	ensure_loaded()
	var enabled := bool(_config.get("enabled", false))
	if enabled_override == "on":
		enabled = true
	elif enabled_override == "off":
		enabled = false
	return enabled and bool((_config.get(prop, {}) as Dictionary).get("enabled", false))
