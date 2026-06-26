# Unit tests for window display mode settings.
# Run via: godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_window_display_mode_settings.gd
extends SceneTree

const ClientSettingsScript := preload("res://scripts/client_settings.gd")
const SettingsPanelScript := preload("res://scripts/settings_panel.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_normalize_unknown_window_mode()
	_test_window_mode_save_load()
	_test_window_mode_save_shape()
	_test_settings_panel_window_mode_button_sync()
	print("[gdtest] PASS: test_window_display_mode_settings (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_normalize_unknown_window_mode() -> void:
	_assert_eq(
		"unknown window mode normalizes to windowed",
		ClientSettingsScript.normalize_window_mode("bogus"),
		ClientSettingsScript.WINDOW_MODE_WINDOWED,
	)
	_assert_eq(
		"fullscreen normalizes",
		ClientSettingsScript.normalize_window_mode("FULLSCREEN"),
		ClientSettingsScript.WINDOW_MODE_FULLSCREEN,
	)
	_assert_eq(
		"windowed fullscreen normalizes",
		ClientSettingsScript.normalize_window_mode("windowed_fullscreen"),
		ClientSettingsScript.WINDOW_MODE_WINDOWED_FULLSCREEN,
	)


func _test_window_mode_save_load() -> void:
	var path := "user://test_window_mode_save_load.json"
	var s := ClientSettingsScript.new(path)
	s.window_mode = ClientSettingsScript.WINDOW_MODE_FULLSCREEN
	s.save()
	var s2 := ClientSettingsScript.new(path)
	s2.load()
	_assert_eq("saved fullscreen persists after reload", s2.window_mode, ClientSettingsScript.WINDOW_MODE_FULLSCREEN)


func _test_window_mode_save_shape() -> void:
	var path := "user://test_window_mode_save_shape.json"
	var s := ClientSettingsScript.new(path)
	s.window_mode = ClientSettingsScript.WINDOW_MODE_WINDOWED_FULLSCREEN
	s.save()
	var parsed = JSON.parse_string(FileAccess.get_file_as_string(path))
	_assert_eq("save shape includes window_mode", str(parsed.get("window_mode", "")), ClientSettingsScript.WINDOW_MODE_WINDOWED_FULLSCREEN)


func _test_settings_panel_window_mode_button_sync() -> void:
	var windowed_button := Button.new()
	var fullscreen_button := Button.new()
	var borderless_button := Button.new()
	var panel := SettingsPanelScript.new()
	panel._window_mode_buttons[ClientSettingsScript.WINDOW_MODE_WINDOWED] = windowed_button
	panel._window_mode_buttons[ClientSettingsScript.WINDOW_MODE_FULLSCREEN] = fullscreen_button
	panel._window_mode_buttons[ClientSettingsScript.WINDOW_MODE_WINDOWED_FULLSCREEN] = borderless_button

	panel.set_window_mode(ClientSettingsScript.WINDOW_MODE_WINDOWED)
	_assert_true("windowed button disabled when mode is windowed", windowed_button.disabled)
	_assert_true("fullscreen button enabled when mode is windowed", not fullscreen_button.disabled)

	panel.set_window_mode(ClientSettingsScript.WINDOW_MODE_FULLSCREEN)
	_assert_true("fullscreen button disabled when mode is fullscreen", fullscreen_button.disabled)
	_assert_true("windowed button enabled when mode is fullscreen", not windowed_button.disabled)

	panel.set_window_mode("unknown_mode_should_normalize")
	_assert_true("normalized unknown mode disables windowed button", windowed_button.disabled)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
