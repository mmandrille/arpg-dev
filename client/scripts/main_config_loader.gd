## MainConfigLoader — lazy singleton for shared/rules/main_config.v0.json gameplay values.
class_name MainConfigLoader
extends RefCounted

static var gameplay: Dictionary = {}
static var presentation_lod: Dictionary = {}
static var loot_labels: Dictionary = {}
static var client_perf: Dictionary = {}
static var _loaded: bool = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/main_config.v0.json")
	var parsed = _read_json(path)
	if typeof(parsed) == TYPE_DICTIONARY:
		var root := parsed as Dictionary
		gameplay = root.get("gameplay", {})
		presentation_lod = root.get("presentation_lod", {})
		loot_labels = root.get("loot_labels", {})
		client_perf = root.get("client_perf", {})


static func presentation_lod_rules() -> Dictionary:
	ensure_loaded()
	return presentation_lod


static func presentation_lod_distance_threshold() -> float:
	return float(presentation_lod_rules().get("distance_threshold", 14.0))


static func loot_label_rules() -> Dictionary:
	ensure_loaded()
	return loot_labels


static func reconciliation_backpressure_threshold() -> float:
	ensure_loaded()
	return float(client_perf.get("reconciliation_backpressure_threshold", 1.5))


static func windup_marker_max_concurrent() -> int:
	ensure_loaded()
	return int(client_perf.get("windup_marker_max_concurrent", 12))


static func projectile_visible_cap() -> int:
	ensure_loaded()
	return int(client_perf.get("projectile_visible_cap", 16))


static func reset_for_tests() -> void:
	_loaded = false
	gameplay = {}
	presentation_lod = {}
	loot_labels = {}
	client_perf = {}


static func base_movement_speed() -> float:
	ensure_loaded()
	return float(gameplay.get("base_movement_speed", 1.0))


static func movement_acceleration_seconds() -> float:
	ensure_loaded()
	return float(gameplay.get("movement_acceleration_seconds", 2.0))


static func movement_min_speed_factor() -> float:
	ensure_loaded()
	return float(gameplay.get("movement_min_speed_factor", 0.2))


static func movement_direction_grace_seconds() -> float:
	ensure_loaded()
	return float(gameplay.get("movement_direction_grace_seconds", 0.2))


static func _read_json(path: String):
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("[main-config] cannot open %s" % path)
		return null
	return JSON.parse_string(file.get_as_text())
