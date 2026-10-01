## Loads the isolated, client-only monster death presentation catalog.
class_name MonsterDeathPresentationLoader
extends RefCounted

const DEFAULT_PATH := "../shared/assets/monster_death_presentation.v0.json"

static var _loaded := false
static var _rim_flash: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_rim_flash = {}
	var path := ProjectSettings.globalize_path("res://").path_join(DEFAULT_PATH)
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("MonsterDeathPresentationLoader: cannot open %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	_rim_flash = parse_rim_flash(parsed)
	if _rim_flash.is_empty():
		push_warning("MonsterDeathPresentationLoader: invalid catalog; death rim flash disabled: %s" % path)


static func rim_flash() -> Dictionary:
	ensure_loaded()
	return _rim_flash.duplicate(true)


## Returns an empty dictionary for malformed or out-of-range input. Empty config disables the
## optional effect while leaving the existing death pose and corpse treatment intact.
static func parse_rim_flash(parsed: Variant) -> Dictionary:
	if typeof(parsed) != TYPE_DICTIONARY:
		return {}
	var catalog := parsed as Dictionary
	var version = catalog.get("version", null)
	if typeof(version) not in [TYPE_INT, TYPE_FLOAT] or float(version) != 0.0:
		return {}
	var raw = catalog.get("rim_flash", null)
	if typeof(raw) != TYPE_DICTIONARY:
		return {}
	var values := raw as Dictionary
	if typeof(values.get("enabled", null)) != TYPE_BOOL:
		return {}
	var tint_value = values.get("tint", null)
	if typeof(tint_value) != TYPE_STRING or not _is_hex_color(str(tint_value)):
		return {}
	if not _number_in_range(values.get("peak_strength", null), 0.0, 1.0, false):
		return {}
	if not _number_in_range(values.get("rise_seconds", null), 0.0, 0.5, false):
		return {}
	if not _number_in_range(values.get("hold_seconds", null), 0.0, 0.5, true):
		return {}
	if not _number_in_range(values.get("release_seconds", null), 0.0, 1.5, false):
		return {}
	return {
		"enabled": bool(values["enabled"]),
		"tint": str(values["tint"]),
		"peak_strength": float(values["peak_strength"]),
		"rise_seconds": float(values["rise_seconds"]),
		"hold_seconds": float(values["hold_seconds"]),
		"release_seconds": float(values["release_seconds"]),
	}


static func reset_for_tests() -> void:
	_loaded = false
	_rim_flash = {}


static func _is_hex_color(value: String) -> bool:
	if value.length() != 7 or not value.begins_with("#"):
		return false
	var parsed := Color.from_string(value, Color.TRANSPARENT)
	return parsed.to_html(false).to_lower() == value.substr(1).to_lower()


static func _number_in_range(value: Variant, low: float, high: float, allow_zero: bool) -> bool:
	if typeof(value) not in [TYPE_INT, TYPE_FLOAT]:
		return false
	var number := float(value)
	return number >= low and number <= high and (allow_zero or number > low)
