class_name GearSocketsLoader
extends RefCounted

static var _loaded: bool = false
static var _default_sockets: Dictionary = {}
static var _class_overrides: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_default_sockets = {}
	_class_overrides = {}
	var root := ProjectSettings.globalize_path("res://")
	var data := _read_json(root.path_join("../shared/assets/gear_sockets.v0.json"))
	var default_entry: Dictionary = data.get("default", {}) if typeof(data.get("default", {})) == TYPE_DICTIONARY else {}
	var sockets = default_entry.get("sockets", {})
	if typeof(sockets) == TYPE_DICTIONARY:
		_default_sockets = sockets
	var classes = data.get("classes", {})
	if typeof(classes) == TYPE_DICTIONARY:
		_class_overrides = classes


static func sockets_for_class(class_id: String) -> Dictionary:
	ensure_loaded()
	var merged := _default_sockets.duplicate(true)
	var class_entry: Dictionary = _class_overrides.get(class_id, {})
	if typeof(class_entry) != TYPE_DICTIONARY:
		return merged
	var overrides = class_entry.get("sockets", {})
	if typeof(overrides) != TYPE_DICTIONARY:
		return merged
	for socket_name in overrides.keys():
		var override_entry: Dictionary = overrides[socket_name]
		if typeof(override_entry) != TYPE_DICTIONARY:
			continue
		var base: Dictionary = merged.get(socket_name, {})
		if typeof(base) != TYPE_DICTIONARY:
			base = {}
		merged[socket_name] = _merge_socket_entry(base, override_entry)
	return merged


static func _merge_socket_entry(base: Dictionary, override_entry: Dictionary) -> Dictionary:
	var out := base.duplicate(true)
	for key in override_entry.keys():
		if key == "offset" or key == "rotation_degrees":
			var existing: Dictionary = out.get(key, {}) if typeof(out.get(key, {})) == TYPE_DICTIONARY else {}
			var patch: Dictionary = override_entry.get(key, {}) if typeof(override_entry.get(key, {})) == TYPE_DICTIONARY else {}
			out[key] = existing.merged(patch)
		else:
			out[key] = override_entry[key]
	return out


static func _read_json(path: String) -> Dictionary:
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_error("GearSocketsLoader: cannot read %s" % path)
		return {}
	var parsed = JSON.parse_string(file.get_as_text())
	return parsed if typeof(parsed) == TYPE_DICTIONARY else {}
