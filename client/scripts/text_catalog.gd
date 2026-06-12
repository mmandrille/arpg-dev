## TextCatalog - client-side localized text lookup with English fallback.
class_name TextCatalog
extends RefCounted

const ENGLISH_CATALOG_REL := "../shared/i18n/en.json"

static var _english_strings: Dictionary = {}
static var _loaded: bool = false


static func reset_for_tests() -> void:
	_english_strings = {}
	_loaded = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_english_strings = {}
	var path := ProjectSettings.globalize_path("res://").path_join(ENGLISH_CATALOG_REL)
	var parsed := _read_json_file(path)
	var strings = parsed.get("strings", {})
	if typeof(strings) == TYPE_DICTIONARY:
		_english_strings = strings


static func get_text(key: String, fallback: String = "") -> String:
	ensure_loaded()
	if key != "" and _english_strings.has(key):
		return str(_english_strings.get(key))
	if fallback != "":
		return fallback
	return key


static func _read_json_file(path: String) -> Dictionary:
	if not FileAccess.file_exists(path):
		push_warning("text catalog not found: %s" % path)
		return {}
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		push_warning("text catalog could not be opened: %s" % path)
		return {}
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		return parsed
	push_warning("text catalog is not a JSON object: %s" % path)
	return {}
