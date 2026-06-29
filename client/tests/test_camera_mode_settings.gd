# Unit tests for CameraPresentationsLoader, ClientSettings camera mode, and PlayerCameraController.
# Run via: godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_camera_mode_settings.gd
extends SceneTree

const CameraPresentationsLoaderScript := preload("res://scripts/camera_presentations_loader.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")
const PlayerCameraContextScript := preload("res://scripts/player_camera_context.gd")
const PlayerCameraControllerScript := preload("res://scripts/player_camera_controller.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_loader_defaults_isometric()
	_test_loader_perspective_modes()
	_test_loader_unknown_mode_falls_back_to_isometric()
	_test_settings_normalize_unknown()
	_test_settings_cycle_modes()
	_test_settings_save_load()
	_test_settings_graphics_quality_from_data_and_save()
	_test_controller_isometric_projection()
	_test_controller_late_settings_mode_sync()
	_test_settings_panel_camera_mode_button_sync()
	_test_controller_isometric_mouse_capture_policy()
	_test_loader_isometric_follow_damping_config()
	print("[gdtest] PASS: test_camera_mode_settings (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_loader_defaults_isometric() -> void:
	CameraPresentationsLoaderScript.reset_for_tests()
	CameraPresentationsLoaderScript.ensure_loaded()
	var cfg := CameraPresentationsLoaderScript.mode("isometric")
	_assert_true("isometric returns non-empty dict", not cfg.is_empty())
	_assert_eq("isometric projection is orthogonal", str(cfg.get("projection", "")), "orthogonal")
	_assert_true("isometric reticle disabled", cfg.get("reticle_enabled", true) == false)
	_assert_true("isometric zoom_default > 0", float(cfg.get("zoom_default", 0.0)) > 0.0)


func _test_loader_perspective_modes() -> void:
	CameraPresentationsLoaderScript.reset_for_tests()
	CameraPresentationsLoaderScript.ensure_loaded()
	var cv := CameraPresentationsLoaderScript.mode("chest_view")
	_assert_eq("chest_view projection", cv.get("projection", ""), "perspective")
	_assert_true("chest_view reticle enabled", cv.get("reticle_enabled", false) == true)


func _test_loader_unknown_mode_falls_back_to_isometric() -> void:
	CameraPresentationsLoaderScript.reset_for_tests()
	CameraPresentationsLoaderScript.ensure_loaded()
	var cfg := CameraPresentationsLoaderScript.mode("nonexistent_mode")
	_assert_true("unknown mode returns non-empty dict (isometric fallback)", not cfg.is_empty())
	_assert_eq("unknown mode fallback projection is orthogonal", str(cfg.get("projection", "")), "orthogonal")


func _test_settings_normalize_unknown() -> void:
	var normalized := ClientSettingsScript.normalize_camera_mode("totally_unknown_mode")
	_assert_eq("unknown mode normalizes to isometric", normalized, ClientSettingsScript.CAMERA_MODE_ISOMETRIC)


func _test_settings_cycle_modes() -> void:
	var s := ClientSettingsScript.new("user://test_cycle_cam.json")
	s.camera_mode = ClientSettingsScript.CAMERA_MODE_ISOMETRIC
	var next1 := s.cycle_camera_mode()
	_assert_eq("isometric cycles to chest_view", next1, ClientSettingsScript.CAMERA_MODE_CHEST_VIEW)
	var next2 := s.cycle_camera_mode()
	_assert_eq("chest_view cycles back to isometric", next2, ClientSettingsScript.CAMERA_MODE_ISOMETRIC)


func _test_settings_save_load() -> void:
	var path := "user://test_cam_save_load.json"
	var s := ClientSettingsScript.new(path)
	s.camera_mode = ClientSettingsScript.CAMERA_MODE_CHEST_VIEW
	s.save()
	var s2 := ClientSettingsScript.new(path)
	s2.load()
	_assert_eq("saved chest_view persists after reload", s2.camera_mode, ClientSettingsScript.CAMERA_MODE_CHEST_VIEW)


func _test_settings_graphics_quality_from_data_and_save() -> void:
	var normalized := ClientSettingsScript.normalize_graphics_quality("PERFORMANCE")
	_assert_eq("unknown graphics quality normalizes to balanced", ClientSettingsScript.normalize_graphics_quality("bogus"), ClientSettingsScript.GRAPHICS_QUALITY_BALANCED)
	_assert_eq("performance quality normalizes", normalized, ClientSettingsScript.GRAPHICS_QUALITY_PERFORMANCE)
	var path := "user://test_graphics_quality_save_load.json"
	var s := ClientSettingsScript.new(path)
	s.graphics_quality = ClientSettingsScript.GRAPHICS_QUALITY_PERFORMANCE
	s.save()
	var s2 := ClientSettingsScript.new(path)
	s2.load()
	_assert_eq("saved performance quality persists", s2.graphics_quality, ClientSettingsScript.GRAPHICS_QUALITY_PERFORMANCE)
	_assert_eq("performance effective window size", s2.effective_window_size(), ClientSettingsScript.PERFORMANCE_WINDOW_SIZE)


func _test_controller_isometric_projection() -> void:
	# Controller creates its own Camera3D — no scene-tree parent required for this check.
	CameraPresentationsLoaderScript.reset_for_tests()
	var ctrl := PlayerCameraControllerScript.new()
	var root := Node3D.new()
	get_root().add_child(root)
	var ctx := PlayerCameraContextScript.new()
	ctx.player_anchor = root
	ctrl.setup(ctx, root)
	ctrl.apply_mode("isometric")
	var cam := ctrl.get_gameplay_camera()
	_assert_true("controller isometric: camera is non-null", cam != null)
	if cam != null:
		_assert_eq("controller isometric: projection is orthogonal", cam.projection, Camera3D.PROJECTION_ORTHOGONAL)
	root.queue_free()
	CameraPresentationsLoaderScript.reset_for_tests()


func _test_controller_late_settings_mode_sync() -> void:
	CameraPresentationsLoaderScript.reset_for_tests()
	var settings := ClientSettingsScript.new("user://test_late_cam_sync.json")
	settings.camera_mode = ClientSettingsScript.CAMERA_MODE_CHEST_VIEW
	var ctrl := PlayerCameraControllerScript.new()
	var root := Node3D.new()
	get_root().add_child(root)
	var anchor := Node3D.new()
	root.add_child(anchor)
	var visual := Node3D.new()
	anchor.add_child(visual)
	var ctx := PlayerCameraContextScript.make(anchor, visual, null, Callable())
	ctrl.setup(ctx, root)
	_assert_eq("null settings default rig is isometric", ctrl.get_gameplay_camera().projection, Camera3D.PROJECTION_ORTHOGONAL)
	ctrl.apply_mode(settings.camera_mode)
	_assert_eq("late chest_view apply uses perspective projection", ctrl.get_gameplay_camera().projection, Camera3D.PROJECTION_PERSPECTIVE)
	root.queue_free()
	CameraPresentationsLoaderScript.reset_for_tests()


func _test_settings_panel_camera_mode_button_sync() -> void:
	# Verify SettingsPanel.set_camera_mode() button sync for both modes.
	# Headless state-mutation test: create button dict, call set_camera_mode(),
	# verify the correct button is disabled (state-based contract).
	const SettingsPanelScript := preload("res://scripts/settings_panel.gd")

	var iso_button := Button.new()
	var chest_button := Button.new()

	var panel := SettingsPanelScript.new()
	panel._camera_mode_buttons[ClientSettingsScript.CAMERA_MODE_ISOMETRIC] = iso_button
	panel._camera_mode_buttons[ClientSettingsScript.CAMERA_MODE_CHEST_VIEW] = chest_button

	panel.set_camera_mode(ClientSettingsScript.CAMERA_MODE_ISOMETRIC)
	_assert_true("isometric button disabled when mode is isometric", iso_button.disabled == true)
	_assert_true("chest_view button enabled when mode is isometric", chest_button.disabled == false)

	panel.set_camera_mode(ClientSettingsScript.CAMERA_MODE_CHEST_VIEW)
	_assert_true("chest_view button disabled when mode is chest_view", chest_button.disabled == true)
	_assert_true("isometric button enabled when mode is chest_view", iso_button.disabled == false)

	# Unknown mode normalizes to isometric.
	panel.set_camera_mode("unknown_mode_should_normalize")
	_assert_true("normalized unknown mode disables isometric button", iso_button.disabled == true)
	_assert_true("normalized unknown mode enables chest_view button", chest_button.disabled == false)


func _test_controller_isometric_mouse_capture_policy() -> void:
	# Isometric uses ORTHOGONAL; chest_view uses PERSPECTIVE.
	# main.gd:_update_mouse_capture() gates mouse capture on this distinction.
	CameraPresentationsLoaderScript.reset_for_tests()
	var ctrl := PlayerCameraControllerScript.new()
	var root := Node3D.new()
	get_root().add_child(root)
	var ctx := PlayerCameraContextScript.new()
	ctx.player_anchor = root
	ctrl.setup(ctx, root)
	ctrl.apply_mode(ClientSettingsScript.CAMERA_MODE_ISOMETRIC)
	var cam := ctrl.get_gameplay_camera()
	_assert_true("isometric camera is non-null", cam != null)
	if cam != null:
		_assert_eq("isometric camera projection is ORTHOGONAL", cam.projection, Camera3D.PROJECTION_ORTHOGONAL)

	ctrl.apply_mode(ClientSettingsScript.CAMERA_MODE_CHEST_VIEW)
	cam = ctrl.get_gameplay_camera()
	_assert_true("chest_view camera is non-null", cam != null)
	if cam != null:
		_assert_eq("chest_view camera projection is PERSPECTIVE", cam.projection, Camera3D.PROJECTION_PERSPECTIVE)

	root.queue_free()
	CameraPresentationsLoaderScript.reset_for_tests()


func _test_loader_isometric_follow_damping_config() -> void:
	CameraPresentationsLoaderScript.reset_for_tests()
	CameraPresentationsLoaderScript.ensure_loaded()
	var cfg := CameraPresentationsLoaderScript.mode("isometric")
	_assert_true("isometric follow_damping_seconds configured", float(cfg.get("follow_damping_seconds", 0.0)) > 0.0)
	_assert_eq("chest_view follow_damping_seconds disabled", float(CameraPresentationsLoaderScript.mode("chest_view").get("follow_damping_seconds", -1.0)), 0.0)


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
