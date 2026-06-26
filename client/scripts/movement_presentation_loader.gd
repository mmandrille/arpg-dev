## MovementPresentationLoader — static singleton for movement presentation tuning.
class_name MovementPresentationLoader
extends RefCounted

static var _loaded: bool = false
static var _tick_smoothing: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/movement_presentation.v0.json")
	if not FileAccess.file_exists(path):
		push_warning("MovementPresentationLoader: data file missing: %s" % path)
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		push_warning("MovementPresentationLoader: could not open: %s" % path)
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("MovementPresentationLoader: malformed JSON: %s" % path)
		return
	var tick_smoothing = parsed.get("tick_smoothing", {})
	if typeof(tick_smoothing) == TYPE_DICTIONARY:
		_tick_smoothing = tick_smoothing


static func tick_smoothing() -> Dictionary:
	ensure_loaded()
	return _tick_smoothing


static func reset_for_tests() -> void:
	_loaded = false
	_tick_smoothing = {}
