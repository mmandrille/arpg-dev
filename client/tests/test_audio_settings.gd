# Unit tests for audio settings persistence and panel synchronization.
extends SceneTree

const ClientSettingsScript := preload("res://scripts/client_settings.gd")
const ClientAudioBridgeScript := preload("res://scripts/client_audio_bridge.gd")
const SettingsPanelScript := preload("res://scripts/settings_panel.gd")
const TextCatalogScript := preload("res://scripts/text_catalog.gd")

var _pass_count: int = 0
var _fail_count: int = 0
var _sync_call_count: int = 0


func _initialize() -> void:
	_test_audio_settings_defaults_and_clamps()
	_test_audio_settings_save_shape()
	_test_settings_panel_audio_slider_sync()
	_test_audio_slider_change_does_not_force_panel_resync()
	_test_map_opacity_settings()
	print("[gdtest] PASS: test_audio_settings (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_audio_settings_defaults_and_clamps() -> void:
	_assert_float("master default", ClientSettingsScript.master_volume_from_data({}), ClientSettingsScript.DEFAULT_MASTER_VOLUME)
	_assert_float("music default", ClientSettingsScript.music_volume_from_data({}), ClientSettingsScript.DEFAULT_MUSIC_VOLUME)
	_assert_float("sfx default", ClientSettingsScript.sfx_volume_from_data({}), ClientSettingsScript.DEFAULT_SFX_VOLUME)
	_assert_float("map opacity default", ClientSettingsScript.map_opacity_from_data({}), ClientSettingsScript.DEFAULT_MAP_OPACITY)
	_assert_float("master clamps high", ClientSettingsScript.master_volume_from_data({"master_volume": 2.0}), 1.0)
	_assert_float("music clamps low", ClientSettingsScript.music_volume_from_data({"music_volume": -1.0}), 0.0)
	_assert_float("sfx rejects invalid", ClientSettingsScript.sfx_volume_from_data({"sfx_volume": "loud"}), ClientSettingsScript.DEFAULT_SFX_VOLUME)
	_assert_float("map opacity rejects invalid", ClientSettingsScript.map_opacity_from_data({"map_opacity": "clear"}), ClientSettingsScript.DEFAULT_MAP_OPACITY)


func _test_audio_settings_save_shape() -> void:
	var path := "user://test_audio_settings.json"
	var absolute_path := ProjectSettings.globalize_path(path)
	if FileAccess.file_exists(path):
		DirAccess.remove_absolute(absolute_path)
	var settings := ClientSettingsScript.new(path)
	settings.set_audio_volumes(0.25, 0.35, 0.45, false)
	settings.set_map_opacity(0.55, false)
	settings.save()
	var parsed = JSON.parse_string(FileAccess.get_file_as_string(path)) as Dictionary
	_assert_float("saved master", float(parsed.get("master_volume", 0.0)), 0.25)
	_assert_float("saved music", float(parsed.get("music_volume", 0.0)), 0.35)
	_assert_float("saved sfx", float(parsed.get("sfx_volume", 0.0)), 0.45)
	_assert_float("saved map opacity", float(parsed.get("map_opacity", 0.0)), 0.55)
	var reloaded := ClientSettingsScript.new(path)
	reloaded.load()
	_assert_float("reloaded master", reloaded.master_volume, 0.25)
	_assert_float("reloaded music", reloaded.music_volume, 0.35)
	_assert_float("reloaded sfx", reloaded.sfx_volume, 0.45)
	_assert_float("reloaded map opacity", reloaded.map_opacity, 0.55)
	DirAccess.remove_absolute(absolute_path)


func _test_settings_panel_audio_slider_sync() -> void:
	TextCatalogScript.reset_for_tests()
	TextCatalogScript.set_locale("es")
	var panel = SettingsPanelScript.new()
	get_root().add_child(panel)
	panel._build()
	panel.show_settings("1920x1080", true, true, "solo", "es", "always", 0.2, 0.3, 0.4, 0.65)
	_assert_float("master slider sync", float(panel._master_volume_slider.value), 0.2)
	_assert_float("music slider sync", float(panel._music_volume_slider.value), 0.3)
	_assert_float("sfx slider sync", float(panel._sfx_volume_slider.value), 0.4)
	_assert_float("map transparency slider sync", float(panel._map_opacity_slider.value), 0.35)
	_assert_eq("master translated", panel._master_volume_label.text, "Volumen general")
	_assert_eq("map transparency translated", panel._map_opacity_label.text, "Transparencia del mapa")
	panel.set_audio_volumes(1.5, -1.0, 0.55)
	_assert_float("master slider clamps high", float(panel._master_volume_slider.value), 1.0)
	_assert_float("music slider clamps low", float(panel._music_volume_slider.value), 0.0)
	_assert_float("sfx slider updates", float(panel._sfx_volume_slider.value), 0.55)
	TextCatalogScript.set_locale("en")
	panel.free()


func _test_audio_slider_change_does_not_force_panel_resync() -> void:
	_sync_call_count = 0
	var settings := ClientSettingsScript.new("user://unused-audio-slider-drag-test.json")
	var panel = SettingsPanelScript.new()
	get_root().add_child(panel)
	panel._build()
	panel.show_settings("1920x1080", true, true, "solo", "en", "contextual", 0.2, 0.3, 0.4, 0.65)
	ClientAudioBridgeScript.connect_volume_signals(panel, null, settings, Callable(self, "_count_sync_call"))
	panel.master_volume_changed.emit(0.75)
	_assert_float("master drag updates settings", settings.master_volume, 0.75)
	_assert_eq("master drag does not resync panel", _sync_call_count, 0)
	panel.music_volume_changed.emit(0.55)
	_assert_float("music drag updates settings", settings.music_volume, 0.55)
	_assert_eq("music drag does not resync panel", _sync_call_count, 0)
	panel.sfx_volume_changed.emit(0.35)
	_assert_float("sfx drag updates settings", settings.sfx_volume, 0.35)
	_assert_eq("sfx drag does not resync panel", _sync_call_count, 0)
	panel.free()


func _count_sync_call() -> void:
	_sync_call_count += 1


func _test_map_opacity_settings() -> void:
	var settings := ClientSettingsScript.new("user://unused-map-opacity-test.json")
	settings.set_map_opacity(1.4, false)
	_assert_float("map opacity clamps high", settings.map_opacity, 1.0)
	settings.set_map_opacity(-0.4, false)
	_assert_float("map opacity clamps low", settings.map_opacity, 0.0)


func _assert_float(label: String, got: float, expected: float) -> void:
	if absf(got - expected) <= 0.001:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])
