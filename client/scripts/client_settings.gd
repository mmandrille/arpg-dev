# ClientSettings stores local-only display preferences.
extends RefCounted
class_name ClientSettings

const DEFAULT_SIZE := Vector2i(1920, 1080)
const CREATE_GAME_SESSION_TYPE_COOP := "coop"
const CREATE_GAME_SESSION_TYPE_SOLO := "solo"
const DEFAULT_CREATE_GAME_SESSION_TYPE := CREATE_GAME_SESSION_TYPE_COOP
const SUPPORTED_SIZES := [
	Vector2i(1280, 720),
	Vector2i(1600, 900),
	Vector2i(1920, 1080),
]
const SUPPORTED_CREATE_GAME_SESSION_TYPES := [
	CREATE_GAME_SESSION_TYPE_COOP,
	CREATE_GAME_SESSION_TYPE_SOLO,
]

var path: String = "user://settings.json"
var window_size: Vector2i = DEFAULT_SIZE
var floating_combat_text: bool = true
var status_text: bool = true
var create_game_session_type: String = DEFAULT_CREATE_GAME_SESSION_TYPE


func _init(settings_path: String = "user://settings.json") -> void:
	path = settings_path


static func supported_size_labels() -> Array:
	var labels: Array = []
	for size in SUPPORTED_SIZES:
		labels.append(size_label(size))
	return labels


static func size_label(size: Vector2i) -> String:
	return "%dx%d" % [size.x, size.y]


static func parse_size_label(label: String) -> Vector2i:
	var parts := label.strip_edges().split("x")
	if parts.size() != 2:
		return DEFAULT_SIZE
	return normalize_size(Vector2i(int(parts[0]), int(parts[1])))


static func normalize_size(size: Vector2i) -> Vector2i:
	for supported in SUPPORTED_SIZES:
		if supported == size:
			return supported
	return DEFAULT_SIZE


static func normalize_create_game_session_type(session_type: String) -> String:
	var normalized := session_type.strip_edges().to_lower()
	if normalized in SUPPORTED_CREATE_GAME_SESSION_TYPES:
		return normalized
	return DEFAULT_CREATE_GAME_SESSION_TYPE


static func create_game_session_type_label(session_type: String) -> String:
	match normalize_create_game_session_type(session_type):
		CREATE_GAME_SESSION_TYPE_SOLO:
			return "Solo"
		_:
			return "Co-op"


static func size_from_data(data) -> Vector2i:
	if typeof(data) != TYPE_DICTIONARY:
		return DEFAULT_SIZE
	var window = data.get("window_size", {})
	if typeof(window) != TYPE_DICTIONARY:
		return DEFAULT_SIZE
	return normalize_size(Vector2i(int(window.get("width", DEFAULT_SIZE.x)), int(window.get("height", DEFAULT_SIZE.y))))


static func floating_combat_text_from_data(data) -> bool:
	if typeof(data) != TYPE_DICTIONARY or not (data as Dictionary).has("floating_combat_text"):
		return true
	return bool((data as Dictionary).get("floating_combat_text", true))


static func status_text_from_data(data) -> bool:
	if typeof(data) != TYPE_DICTIONARY:
		return true
	var settings := data as Dictionary
	if settings.has("status_text"):
		return bool(settings.get("status_text", true))
	return bool(settings.get("top_right_status_text", true))


static func create_game_session_type_from_data(data) -> String:
	if typeof(data) != TYPE_DICTIONARY:
		return DEFAULT_CREATE_GAME_SESSION_TYPE
	return normalize_create_game_session_type(str((data as Dictionary).get("create_game_session_type", DEFAULT_CREATE_GAME_SESSION_TYPE)))


func load() -> void:
	if not FileAccess.file_exists(path):
		window_size = DEFAULT_SIZE
		floating_combat_text = true
		status_text = true
		create_game_session_type = DEFAULT_CREATE_GAME_SESSION_TYPE
		return
	var text := FileAccess.get_file_as_string(path)
	var parsed = JSON.parse_string(text)
	window_size = size_from_data(parsed)
	floating_combat_text = floating_combat_text_from_data(parsed)
	status_text = status_text_from_data(parsed)
	create_game_session_type = create_game_session_type_from_data(parsed)


func save() -> void:
	var file := FileAccess.open(path, FileAccess.WRITE)
	if file == null:
		push_warning("could not save settings: %s" % path)
		return
	file.store_string(JSON.stringify({
		"window_size": {
			"width": window_size.x,
			"height": window_size.y,
		},
		"floating_combat_text": floating_combat_text,
		"status_text": status_text,
		"create_game_session_type": create_game_session_type,
	}))


func apply() -> void:
	var target_size := window_size
	var screen := DisplayServer.window_get_current_screen()
	var scale := DisplayServer.screen_get_scale(screen)
	if scale > 1.0:
		target_size = Vector2i(roundi(float(window_size.x) * scale), roundi(float(window_size.y) * scale))
	DisplayServer.window_set_size(target_size)
	var screen_pos := DisplayServer.screen_get_position(screen)
	var screen_size := DisplayServer.screen_get_size(screen)
	var centered := screen_pos + Vector2i(
		maxi(0, int((screen_size.x - target_size.x) / 2)),
		maxi(0, int((screen_size.y - target_size.y) / 2))
	)
	DisplayServer.window_set_position(centered)


func set_window_size(size: Vector2i, persist: bool = true, apply_now: bool = true) -> void:
	window_size = normalize_size(size)
	if apply_now:
		apply()
	if persist:
		save()


func set_window_size_label(label: String, persist: bool = true, apply_now: bool = true) -> void:
	set_window_size(parse_size_label(label), persist, apply_now)


func set_floating_combat_text(enabled: bool, persist: bool = true) -> void:
	floating_combat_text = enabled
	if persist:
		save()


func set_status_text(enabled: bool, persist: bool = true) -> void:
	status_text = enabled
	if persist:
		save()


func set_create_game_session_type(session_type: String, persist: bool = true) -> void:
	create_game_session_type = normalize_create_game_session_type(session_type)
	if persist:
		save()
