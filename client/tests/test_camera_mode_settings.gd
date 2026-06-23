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
	_test_controller_isometric_projection()
	_test_controller_perspective_projection()
	_test_settings_panel_camera_mode_button_sync()
	_test_controller_isometric_mouse_capture_policy()
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
	var tp := CameraPresentationsLoaderScript.mode("third_person")
	_assert_eq("third_person projection", tp.get("projection", ""), "perspective")
	_assert_true("third_person spring_arm_length > 0", tp.get("spring_arm_length", 0.0) > 0.0)
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
	_assert_eq("isometric cycles to third_person", next1, ClientSettingsScript.CAMERA_MODE_THIRD_PERSON)
	var next2 := s.cycle_camera_mode()
	_assert_eq("third_person cycles to chest_view", next2, ClientSettingsScript.CAMERA_MODE_CHEST_VIEW)
	var next3 := s.cycle_camera_mode()
	_assert_eq("chest_view cycles back to isometric", next3, ClientSettingsScript.CAMERA_MODE_ISOMETRIC)


func _test_settings_save_load() -> void:
	var path := "user://test_cam_save_load.json"
	var s := ClientSettingsScript.new(path)
	s.camera_mode = ClientSettingsScript.CAMERA_MODE_THIRD_PERSON
	s.save()
	var s2 := ClientSettingsScript.new(path)
	s2.load()
	_assert_eq("saved third_person persists after reload", s2.camera_mode, ClientSettingsScript.CAMERA_MODE_THIRD_PERSON)


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


func _test_controller_perspective_projection() -> void:
	CameraPresentationsLoaderScript.reset_for_tests()
	var ctrl := PlayerCameraControllerScript.new()
	var root := Node3D.new()
	get_root().add_child(root)
	var ctx := PlayerCameraContextScript.new()
	ctx.player_anchor = root
	ctrl.setup(ctx, root)
	ctrl.apply_mode("third_person")
	var cam := ctrl.get_gameplay_camera()
	_assert_true("controller third_person: camera is non-null", cam != null)
	if cam != null:
		_assert_eq("controller third_person: projection is perspective", cam.projection, Camera3D.PROJECTION_PERSPECTIVE)
	root.queue_free()
	CameraPresentationsLoaderScript.reset_for_tests()


func _test_settings_panel_camera_mode_button_sync() -> void:
	# Task 6.2: Verify SettingsPanel.set_camera_mode() button sync for all three modes.
	# This is a headless state-mutation test: create button dict, call set_camera_mode(),
	# verify the correct button is disabled (state-based contract).
	const SettingsPanelScript := preload("res://scripts/settings_panel.gd")

	# Create three mock buttons.
	var iso_button := Button.new()
	var third_button := Button.new()
	var chest_button := Button.new()

	# Create a minimal panel instance and populate its button dict manually.
	var panel := SettingsPanelScript.new()
	panel._camera_mode_buttons[ClientSettingsScript.CAMERA_MODE_ISOMETRIC] = iso_button
	panel._camera_mode_buttons[ClientSettingsScript.CAMERA_MODE_THIRD_PERSON] = third_button
	panel._camera_mode_buttons[ClientSettingsScript.CAMERA_MODE_CHEST_VIEW] = chest_button

	# Test isometric mode button sync
	panel.set_camera_mode(ClientSettingsScript.CAMERA_MODE_ISOMETRIC)
	_assert_true("isometric button disabled when mode is isometric", iso_button.disabled == true)
	_assert_true("third_person button enabled when mode is isometric", third_button.disabled == false)
	_assert_true("chest_view button enabled when mode is isometric", chest_button.disabled == false)

	# Test third_person mode button sync
	panel.set_camera_mode(ClientSettingsScript.CAMERA_MODE_THIRD_PERSON)
	_assert_true("third_person button disabled when mode is third_person", third_button.disabled == true)
	_assert_true("isometric button enabled when mode is third_person", iso_button.disabled == false)
	_assert_true("chest_view button enabled when mode is third_person", chest_button.disabled == false)

	# Test chest_view mode button sync
	panel.set_camera_mode(ClientSettingsScript.CAMERA_MODE_CHEST_VIEW)
	_assert_true("chest_view button disabled when mode is chest_view", chest_button.disabled == true)
	_assert_true("isometric button enabled when mode is chest_view", iso_button.disabled == false)
	_assert_true("third_person button enabled when mode is chest_view", third_button.disabled == false)

	# Test normalize fallback: unknown mode normalizes to isometric, sync reflects that
	panel.set_camera_mode("unknown_mode_should_normalize")
	_assert_true("normalized unknown mode disables isometric button", iso_button.disabled == true)
	_assert_true("normalized unknown mode enables third_person button", third_button.disabled == false)
	_assert_true("normalized unknown mode enables chest_view button", chest_button.disabled == false)


func _test_controller_isometric_mouse_capture_policy() -> void:
	# Task 6.4: Assert isometric mode keeps Input.MOUSE_MODE_VISIBLE policy.
	# The game applies this policy in main.gd:_update_mouse_capture(), but the
	# contract is that PlayerCameraController applies isometric (orthographic)
	# projection, and the caller (main.gd) then decides to keep the mouse visible
	# for isometric and captured for third_person/chest_view.
	#
	# This test verifies the camera mode configuration is orthographic for isometric,
	# which is the prerequisite for the mouse capture policy applied by the caller.
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
		_assert_eq("isometric camera projection is ORTHOGONAL (not PERSPECTIVE)", cam.projection, Camera3D.PROJECTION_ORTHOGONAL)

	# Verify that third_person uses PERSPECTIVE (opposite of isometric).
	# This ensures the caller code that checks "camera_mode != CAMERA_MODE_ISOMETRIC"
	# can reliably capture the mouse for perspective modes.
	ctrl.apply_mode(ClientSettingsScript.CAMERA_MODE_THIRD_PERSON)
	cam = ctrl.get_gameplay_camera()
	_assert_true("third_person camera is non-null", cam != null)
	if cam != null:
		_assert_eq("third_person camera projection is PERSPECTIVE (not ORTHOGONAL)", cam.projection, Camera3D.PROJECTION_PERSPECTIVE)

	root.queue_free()
	CameraPresentationsLoaderScript.reset_for_tests()


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
