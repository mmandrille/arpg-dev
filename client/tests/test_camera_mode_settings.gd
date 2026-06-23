# Unit tests for CameraPresentationsLoader: defaults and unknown-mode fallback.
# Run via: godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_camera_mode_settings.gd
extends SceneTree

const CameraPresentationsLoaderScript := preload("res://scripts/camera_presentations_loader.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_loader_defaults_isometric()
	_test_loader_unknown_mode_falls_back_to_isometric()
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
