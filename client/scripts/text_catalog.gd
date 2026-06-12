## TextCatalog - client-side localized text lookup with English fallback.
class_name TextCatalog
extends RefCounted

const ENGLISH_CATALOG_REL := "../shared/i18n/en.json"
const LOCALE_CATALOG_DIR_REL := "../shared/i18n"

static var _english_strings: Dictionary = {}
static var _locale_strings: Dictionary = {}
static var _current_locale: String = "en"
static var _loaded: bool = false


static func reset_for_tests() -> void:
	_english_strings = {}
	_locale_strings = {}
	_current_locale = "en"
	_loaded = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_english_strings = {}
	_locale_strings = {}
	var path := ProjectSettings.globalize_path("res://").path_join(ENGLISH_CATALOG_REL)
	var parsed := _read_json_file(path)
	var strings = parsed.get("strings", {})
	if typeof(strings) == TYPE_DICTIONARY:
		_english_strings = strings
	if _current_locale != "en":
		_load_locale(_current_locale)


static func current_locale() -> String:
	return _current_locale


static func set_locale(locale: String) -> void:
	var normalized := normalize_locale(locale)
	if normalized == _current_locale and _loaded:
		return
	_current_locale = normalized
	_locale_strings = {}
	if _loaded and _current_locale != "en":
		_load_locale(_current_locale)


static func normalize_locale(locale: String) -> String:
	var normalized := locale.strip_edges().to_lower()
	if normalized == "es":
		return "es"
	return "en"


static func get_text(key: String, fallback: String = "") -> String:
	ensure_loaded()
	if key != "" and _locale_strings.has(key):
		return str(_locale_strings.get(key))
	if key != "" and _english_strings.has(key):
		return str(_english_strings.get(key))
	if fallback != "":
		return fallback
	return key


static func _load_locale(locale: String) -> void:
	var path := ProjectSettings.globalize_path("res://").path_join(LOCALE_CATALOG_DIR_REL).path_join("%s.json" % locale)
	var parsed := _read_json_file(path)
	var strings = parsed.get("strings", {})
	if typeof(strings) == TYPE_DICTIONARY:
		_locale_strings = strings


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
