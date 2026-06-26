# ClientSettings stores local-only display preferences.
extends RefCounted
class_name ClientSettings

const TextCatalogScript := preload("res://scripts/text_catalog.gd")

const DEFAULT_SIZE := Vector2i(2560, 1440)
const CREATE_GAME_SESSION_TYPE_COOP := "coop"
const CREATE_GAME_SESSION_TYPE_SOLO := "solo"
const DEFAULT_CREATE_GAME_SESSION_TYPE := CREATE_GAME_SESSION_TYPE_COOP
const DEFAULT_LANGUAGE := "en"
const DEFAULT_LOOT_FILTER_MODE := "All"
const MONSTER_HEALTH_BAR_CONTEXTUAL := "contextual"
const MONSTER_HEALTH_BAR_ALWAYS := "always"
const DEFAULT_MONSTER_HEALTH_BAR_MODE := MONSTER_HEALTH_BAR_CONTEXTUAL
const CAMERA_MODE_ISOMETRIC := "isometric"
const CAMERA_MODE_CHEST_VIEW := "chest_view"
const DEFAULT_CAMERA_MODE := CAMERA_MODE_ISOMETRIC
const GRAPHICS_QUALITY_BALANCED := "balanced"
const GRAPHICS_QUALITY_PERFORMANCE := "performance"
const DEFAULT_GRAPHICS_QUALITY := GRAPHICS_QUALITY_BALANCED
const WINDOW_MODE_WINDOWED := "windowed"
const WINDOW_MODE_FULLSCREEN := "fullscreen"
const WINDOW_MODE_WINDOWED_FULLSCREEN := "windowed_fullscreen"
const DEFAULT_WINDOW_MODE := WINDOW_MODE_WINDOWED
const PERFORMANCE_WINDOW_SIZE := Vector2i(1920, 1080)
const SUPPORTED_CAMERA_MODES := [
	CAMERA_MODE_ISOMETRIC,
	CAMERA_MODE_CHEST_VIEW,
]
const DEFAULT_MASTER_VOLUME := 0.8
const DEFAULT_MUSIC_VOLUME := 0.0
const DEFAULT_SFX_VOLUME := 0.8
const DEFAULT_MAP_OPACITY := 0.68
const SUPPORTED_LANGUAGES := ["en", "es"]
const SUPPORTED_LOOT_FILTER_MODES := ["All", "Magic+", "Rare+", "Unique"]
const SUPPORTED_MONSTER_HEALTH_BAR_MODES := [
	MONSTER_HEALTH_BAR_CONTEXTUAL,
	MONSTER_HEALTH_BAR_ALWAYS,
]
const SUPPORTED_SIZES := [
	Vector2i(1280, 720),
	Vector2i(1600, 900),
	Vector2i(1920, 1080),
	Vector2i(2560, 1440),
]
const SUPPORTED_CREATE_GAME_SESSION_TYPES := [
	CREATE_GAME_SESSION_TYPE_COOP,
	CREATE_GAME_SESSION_TYPE_SOLO,
]
const SUPPORTED_GRAPHICS_QUALITIES := [
	GRAPHICS_QUALITY_BALANCED,
	GRAPHICS_QUALITY_PERFORMANCE,
]
const SUPPORTED_WINDOW_MODES := [
	WINDOW_MODE_WINDOWED,
	WINDOW_MODE_FULLSCREEN,
	WINDOW_MODE_WINDOWED_FULLSCREEN,
]

var path: String = "user://settings.json"
var window_size: Vector2i = DEFAULT_SIZE
var window_mode: String = DEFAULT_WINDOW_MODE
var floating_combat_text: bool = true
var status_text: bool = true
var create_game_session_type: String = DEFAULT_CREATE_GAME_SESSION_TYPE
var language: String = DEFAULT_LANGUAGE
var loot_filter_mode: String = DEFAULT_LOOT_FILTER_MODE
var monster_health_bar_mode: String = DEFAULT_MONSTER_HEALTH_BAR_MODE
var master_volume: float = DEFAULT_MASTER_VOLUME
var music_volume: float = DEFAULT_MUSIC_VOLUME
var sfx_volume: float = DEFAULT_SFX_VOLUME
var map_opacity: float = DEFAULT_MAP_OPACITY
var camera_mode: String = DEFAULT_CAMERA_MODE
var graphics_quality: String = DEFAULT_GRAPHICS_QUALITY


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
			return TextCatalogScript.get_text("settings.session_type.solo", "Solo")
		_:
			return TextCatalogScript.get_text("settings.session_type.coop", "Co-op")


static func normalize_language(language_id: String) -> String:
	var normalized := language_id.strip_edges().to_lower()
	if normalized in SUPPORTED_LANGUAGES:
		return normalized
	return DEFAULT_LANGUAGE


static func language_label(language_id: String) -> String:
	var normalized := normalize_language(language_id)
	return TextCatalogScript.get_text("settings.language.%s" % normalized, normalized)


static func normalize_loot_filter_mode(mode: String) -> String:
	var normalized := mode.strip_edges().to_lower()
	for supported in SUPPORTED_LOOT_FILTER_MODES:
		if supported.to_lower() == normalized:
			return supported
	return DEFAULT_LOOT_FILTER_MODE


static func normalize_monster_health_bar_mode(mode: String) -> String:
	var normalized := mode.strip_edges().to_lower()
	if normalized == "default":
		return MONSTER_HEALTH_BAR_CONTEXTUAL
	for supported in SUPPORTED_MONSTER_HEALTH_BAR_MODES:
		if supported == normalized:
			return supported
	return DEFAULT_MONSTER_HEALTH_BAR_MODE


static func normalize_camera_mode(mode: String) -> String:
	var normalized := mode.strip_edges().to_lower()
	if normalized in SUPPORTED_CAMERA_MODES:
		return normalized
	return DEFAULT_CAMERA_MODE


static func normalize_graphics_quality(quality: String) -> String:
	var normalized := quality.strip_edges().to_lower()
	if normalized in SUPPORTED_GRAPHICS_QUALITIES:
		return normalized
	return DEFAULT_GRAPHICS_QUALITY


static func graphics_quality_label(quality: String) -> String:
	match normalize_graphics_quality(quality):
		GRAPHICS_QUALITY_PERFORMANCE:
			return TextCatalogScript.get_text("settings.graphics_quality.performance", "Performance")
		_:
			return TextCatalogScript.get_text("settings.graphics_quality.balanced", "Balanced")


static func normalize_window_mode(mode: String) -> String:
	var normalized := mode.strip_edges().to_lower()
	if normalized in SUPPORTED_WINDOW_MODES:
		return normalized
	return DEFAULT_WINDOW_MODE


static func window_mode_label(mode: String) -> String:
	match normalize_window_mode(mode):
		WINDOW_MODE_FULLSCREEN:
			return TextCatalogScript.get_text("settings.window_mode.fullscreen", "Fullscreen")
		WINDOW_MODE_WINDOWED_FULLSCREEN:
			return TextCatalogScript.get_text("settings.window_mode.windowed_fullscreen", "Windowed Fullscreen")
		_:
			return TextCatalogScript.get_text("settings.window_mode.windowed", "Windowed")


static func normalize_volume(value, fallback: float) -> float:
	if typeof(value) != TYPE_FLOAT and typeof(value) != TYPE_INT:
		return fallback
	return clampf(float(value), 0.0, 1.0)


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


static func language_from_data(data) -> String:
	if typeof(data) != TYPE_DICTIONARY:
		return DEFAULT_LANGUAGE
	return normalize_language(str((data as Dictionary).get("language", DEFAULT_LANGUAGE)))


static func loot_filter_mode_from_data(data) -> String:
	if typeof(data) != TYPE_DICTIONARY:
		return DEFAULT_LOOT_FILTER_MODE
	return normalize_loot_filter_mode(str((data as Dictionary).get("loot_filter_mode", DEFAULT_LOOT_FILTER_MODE)))


static func monster_health_bar_mode_from_data(data) -> String:
	if typeof(data) != TYPE_DICTIONARY:
		return DEFAULT_MONSTER_HEALTH_BAR_MODE
	return normalize_monster_health_bar_mode(str((data as Dictionary).get("monster_health_bar_mode", DEFAULT_MONSTER_HEALTH_BAR_MODE)))


static func camera_mode_from_data(data) -> String:
	if typeof(data) != TYPE_DICTIONARY:
		return DEFAULT_CAMERA_MODE
	return normalize_camera_mode(str((data as Dictionary).get("camera_mode", DEFAULT_CAMERA_MODE)))


static func graphics_quality_from_data(data) -> String:
	if typeof(data) != TYPE_DICTIONARY:
		return DEFAULT_GRAPHICS_QUALITY
	return normalize_graphics_quality(str((data as Dictionary).get("graphics_quality", DEFAULT_GRAPHICS_QUALITY)))


static func master_volume_from_data(data) -> float:
	if typeof(data) != TYPE_DICTIONARY:
		return DEFAULT_MASTER_VOLUME
	return normalize_volume((data as Dictionary).get("master_volume", DEFAULT_MASTER_VOLUME), DEFAULT_MASTER_VOLUME)


static func music_volume_from_data(data) -> float:
	if typeof(data) != TYPE_DICTIONARY:
		return DEFAULT_MUSIC_VOLUME
	return normalize_volume((data as Dictionary).get("music_volume", DEFAULT_MUSIC_VOLUME), DEFAULT_MUSIC_VOLUME)


static func sfx_volume_from_data(data) -> float:
	if typeof(data) != TYPE_DICTIONARY:
		return DEFAULT_SFX_VOLUME
	return normalize_volume((data as Dictionary).get("sfx_volume", DEFAULT_SFX_VOLUME), DEFAULT_SFX_VOLUME)


static func map_opacity_from_data(data) -> float:
	if typeof(data) != TYPE_DICTIONARY:
		return DEFAULT_MAP_OPACITY
	return normalize_volume((data as Dictionary).get("map_opacity", DEFAULT_MAP_OPACITY), DEFAULT_MAP_OPACITY)


static func window_mode_from_data(data) -> String:
	if typeof(data) != TYPE_DICTIONARY:
		return DEFAULT_WINDOW_MODE
	return normalize_window_mode(str((data as Dictionary).get("window_mode", DEFAULT_WINDOW_MODE)))


func load() -> void:
	if not FileAccess.file_exists(path):
		window_size = DEFAULT_SIZE
		window_mode = DEFAULT_WINDOW_MODE
		floating_combat_text = true
		status_text = true
		create_game_session_type = DEFAULT_CREATE_GAME_SESSION_TYPE
		language = DEFAULT_LANGUAGE
		loot_filter_mode = DEFAULT_LOOT_FILTER_MODE
		monster_health_bar_mode = DEFAULT_MONSTER_HEALTH_BAR_MODE
		camera_mode = DEFAULT_CAMERA_MODE
		graphics_quality = DEFAULT_GRAPHICS_QUALITY
		master_volume = DEFAULT_MASTER_VOLUME
		music_volume = DEFAULT_MUSIC_VOLUME
		sfx_volume = DEFAULT_SFX_VOLUME
		map_opacity = DEFAULT_MAP_OPACITY
		return
	var text := FileAccess.get_file_as_string(path)
	var parsed = JSON.parse_string(text)
	window_size = size_from_data(parsed)
	window_mode = window_mode_from_data(parsed)
	floating_combat_text = floating_combat_text_from_data(parsed)
	status_text = status_text_from_data(parsed)
	create_game_session_type = create_game_session_type_from_data(parsed)
	language = language_from_data(parsed)
	loot_filter_mode = loot_filter_mode_from_data(parsed)
	monster_health_bar_mode = monster_health_bar_mode_from_data(parsed)
	camera_mode = camera_mode_from_data(parsed)
	graphics_quality = graphics_quality_from_data(parsed)
	master_volume = master_volume_from_data(parsed)
	music_volume = music_volume_from_data(parsed)
	sfx_volume = sfx_volume_from_data(parsed)
	map_opacity = map_opacity_from_data(parsed)


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
		"window_mode": window_mode,
		"floating_combat_text": floating_combat_text,
		"status_text": status_text,
		"create_game_session_type": create_game_session_type,
		"language": language,
		"loot_filter_mode": loot_filter_mode,
		"monster_health_bar_mode": monster_health_bar_mode,
		"camera_mode": camera_mode,
		"graphics_quality": graphics_quality,
		"master_volume": master_volume,
		"music_volume": music_volume,
		"sfx_volume": sfx_volume,
		"map_opacity": map_opacity,
	}))


func effective_window_size() -> Vector2i:
	if graphics_quality == GRAPHICS_QUALITY_PERFORMANCE:
		return PERFORMANCE_WINDOW_SIZE
	return window_size


func apply() -> void:
	var target_size := _fit_size_to_screen(effective_window_size())
	DisplayServer.window_set_min_size(SUPPORTED_SIZES[0])
	match normalize_window_mode(window_mode):
		WINDOW_MODE_FULLSCREEN:
			DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_FULLSCREEN)
		WINDOW_MODE_WINDOWED_FULLSCREEN:
			DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_EXCLUSIVE_FULLSCREEN)
		_:
			DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_WINDOWED)
			DisplayServer.window_set_size(target_size)
			_center_window(target_size)


static func _fit_size_to_screen(size: Vector2i) -> Vector2i:
	var screen := DisplayServer.window_get_current_screen()
	var screen_size := DisplayServer.screen_get_size(screen)
	var margin := Vector2i(96, 96)
	var available := Vector2i(maxi(640, screen_size.x - margin.x), maxi(360, screen_size.y - margin.y))
	if size.x <= available.x and size.y <= available.y:
		return size
	var fit_scale := minf(float(available.x) / float(size.x), float(available.y) / float(size.y))
	return Vector2i(
		maxi(SUPPORTED_SIZES[0].x, roundi(float(size.x) * fit_scale)),
		maxi(SUPPORTED_SIZES[0].y, roundi(float(size.y) * fit_scale))
	)


static func _center_window(target_size: Vector2i) -> void:
	var screen := DisplayServer.window_get_current_screen()
	var screen_pos := DisplayServer.screen_get_position(screen)
	var screen_size := DisplayServer.screen_get_size(screen)
	var centered := screen_pos + Vector2i(
		maxi(0, int((screen_size.x - target_size.x) / 2)),
		maxi(0, int((screen_size.y - target_size.y) / 2))
	)
	DisplayServer.window_set_position(centered)


func set_window_mode(mode: String, persist: bool = true, apply_now: bool = true) -> void:
	window_mode = normalize_window_mode(mode)
	if apply_now:
		apply()
	if persist:
		save()


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


func set_language(language_id: String, persist: bool = true) -> void:
	language = normalize_language(language_id)
	TextCatalogScript.set_locale(language)
	if persist:
		save()


func set_loot_filter_mode(mode: String, persist: bool = true) -> void:
	loot_filter_mode = normalize_loot_filter_mode(mode)
	if persist:
		save()


func set_monster_health_bar_mode(mode: String, persist: bool = true) -> void:
	monster_health_bar_mode = normalize_monster_health_bar_mode(mode)
	if persist:
		save()


func set_camera_mode(mode: String, persist: bool = true) -> void:
	camera_mode = normalize_camera_mode(mode)
	if persist:
		save()


func set_graphics_quality(quality: String, persist: bool = true, apply_now: bool = true) -> void:
	graphics_quality = normalize_graphics_quality(quality)
	if apply_now:
		apply()
	if persist:
		save()


func cycle_camera_mode() -> String:
	var idx := SUPPORTED_CAMERA_MODES.find(camera_mode)
	camera_mode = SUPPORTED_CAMERA_MODES[(idx + 1) % SUPPORTED_CAMERA_MODES.size()]
	save()
	return camera_mode


func set_audio_volumes(master: float, music: float, sfx: float, persist: bool = true) -> void:
	master_volume = clampf(master, 0.0, 1.0)
	music_volume = clampf(music, 0.0, 1.0)
	sfx_volume = clampf(sfx, 0.0, 1.0)
	if persist:
		save()


func set_master_volume(value: float, persist: bool = true) -> void:
	master_volume = clampf(value, 0.0, 1.0)
	if persist:
		save()


func set_music_volume(value: float, persist: bool = true) -> void:
	music_volume = clampf(value, 0.0, 1.0)
	if persist:
		save()


func set_sfx_volume(value: float, persist: bool = true) -> void:
	sfx_volume = clampf(value, 0.0, 1.0)
	if persist:
		save()


func set_map_opacity(value: float, persist: bool = true) -> void:
	map_opacity = clampf(value, 0.0, 1.0)
	if persist:
		save()
