## CameraPresentationsLoader — static singleton for camera mode tuning data.
##
## Loads camera_presentations.v0.json once (lazy, guarded by _loaded) and caches
## it as a static var. Falls back to the "isometric" mode for unknown keys so
## callers never receive an empty dictionary.
##
## Usage (any script — no autoload registration needed):
##   CameraPresentationsLoader.ensure_loaded()
##   var cfg := CameraPresentationsLoader.mode("chest_view")
class_name CameraPresentationsLoader
extends RefCounted

static var _loaded: bool = false
static var _modes: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/camera_presentations.v0.json")
	if not FileAccess.file_exists(path):
		push_warning("CameraPresentationsLoader: data file missing: %s" % path)
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		push_warning("CameraPresentationsLoader: could not open: %s" % path)
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("CameraPresentationsLoader: malformed JSON: %s" % path)
		return
	var modes = parsed.get("modes", {})
	if typeof(modes) == TYPE_DICTIONARY:
		_modes = modes


static func mode(name: String) -> Dictionary:
	ensure_loaded()
	if _modes.has(name):
		return _modes[name]
	return _modes.get("isometric", {})


static func reset_for_tests() -> void:
	_loaded = false
	_modes = {}
