# Unit test for blocking gameplay keyboard input while text fields have focus.
# Run via: godot --headless --path client --script res://tests/test_text_input_focus_guard.gd
extends SceneTree

const TextInputFocusGuardScript := preload("res://scripts/text_input_focus_guard.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var edit := LineEdit.new()
	root.add_child(edit)
	await process_frame
	edit.grab_focus()
	await process_frame
	_assert_true("line edit focus blocks gameplay keys", TextInputFocusGuardScript.has_text_input_focus(root.get_viewport()))
	edit.free()
	print("[gdtest] PASS: test_text_input_focus_guard (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
