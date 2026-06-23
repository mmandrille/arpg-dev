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
	_test_loader_unknown_mode_falls_back_to_isometric()
	_test_settings_normalize_unknown()
	_test_settings_cycle_modes()
	_test_settings_save_load()
	_test_controller_isometric_projection()
	_test_controller_perspective_projection()
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
